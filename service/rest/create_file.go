// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package rest

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"github.com/HPInc/krypton-fs/service/common"
	"github.com/HPInc/krypton-fs/service/config"
	"github.com/HPInc/krypton-fs/service/db"
	"github.com/HPInc/krypton-fs/service/metrics"
	"github.com/HPInc/krypton-fs/service/storage"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	minFileNameLength = 1
	maxFileNameLength = 127

	minChecksumLength = 3
	maxChecksumLength = 25

	minFileLength = 1
)

// Creates a record for a new file in the database and returns a signed URL for
// the caller to upload the file to storage (S3).
func CreateFileHandler(w http.ResponseWriter, r *http.Request) {
	requestID := r.Header.Get(headerRequestID)

	// Check if the contents of the POST were provided using JSON encoding.
	if r.Header.Get(headerContentType) != contentTypeJson {
		fsLogger.Error("CreateFile POST request does not have JSON encoding!",
			zap.String("Request ID:", requestID),
		)
		sendUnsupportedMediaTypeResponse(w)
		metrics.MetricCreateFileUnSupportedMediaTypeRequests.Inc()
		return
	}

	// validate device token and get values
	deviceInfo, err := getDeviceInfoFromToken(r)
	if err != nil {
		fsLogger.Info("CreateFile token validation error",
			zap.Error(err))
		sendUnauthorizedErrorResponse(w)
		metrics.MetricCreateFileUnauthorizedRequests.Inc()
		return
	}

	// Extract the create file request payload.
	payload, err := getRequestPayload(r)
	if err != nil {
		fsLogger.Error("Failed to read the create file request payload",
			zap.String("Request ID:", requestID),
			zap.Error(err),
		)
		sendBadRequestErrorResponse(w)
		metrics.MetricCreateFileBadRequests.Inc()
		return
	}

	// Unmarshal the create file request.
	var request common.CreateFileRequest
	err = json.Unmarshal(payload, &request)
	if err != nil {
		fsLogger.Error("Failed to unmarshall the create file request",
			zap.String("Request ID:", requestID),
			zap.Error(err),
		)
		sendBadRequestErrorResponse(w)
		metrics.MetricCreateFileBadRequests.Inc()
		return
	}

	if request.Size < minFileLength {
		fsLogger.Error("Invalid file size in create file request",
			zap.String("Request ID", requestID),
			zap.Int64("Size", request.Size),
			zap.Error(err),
		)
		sendBadRequestErrorResponse(w)
		metrics.MetricCreateFileBadRequests.Inc()
		return
	}

	// fill in device and tenant from token
	request.TenantID = deviceInfo.TenantID
	request.DeviceID = deviceInfo.DeviceID

	// Validate the create file request.
	if !isValidCreateFileRequest(requestID, &request) {
		sendBadRequestErrorResponse(w)
		metrics.MetricCreateFileBadRequests.Inc()
		return
	}

	// Create an entry for the file in the database. This process will yield
	// a unique sequence number (ID) for the file.
	createdFile, err := db.CreateFile(requestID, &request)
	if err != nil {
		fsLogger.Error("Failed to create an entry for the file in the database!",
			zap.String("Request ID:", requestID),
			zap.Error(err),
		)
		sendInternalServerErrorResponse(w)
		metrics.MetricCreateFileInternalErrors.Inc()
		return
	}

	response := common.CommonFileResponse{
		File: common.FileInformation{
			FileID:    createdFile.FileID,
			TenantID:  createdFile.TenantID,
			DeviceID:  createdFile.DeviceID,
			Name:      createdFile.Name,
			Checksum:  createdFile.Checksum,
			Size:      createdFile.Size,
			CreatedAt: createdFile.CreatedAt,
			UpdatedAt: createdFile.UpdatedAt,
		},
		RequestID:    requestID,
		ResponseTime: time.Now(),
	}

	// Issue a pre-signed URL corresponding to this file ID. The pre-signed URL
	// can be used by the client to upload the file to storage.
	response.File.SignedUrl, err = storage.Provider.GetSignedUrl(
		createdFile.BucketName,
		storage.GetObjectName(request.TenantID, request.DeviceID, createdFile.FileID),
		config.AccessMethodPut,
		request.Checksum,
		request.Size)
	if err != nil {
		fsLogger.Error("Failed to generate a signed URL for the file!",
			zap.String("Request ID:", requestID),
			zap.Error(err),
		)
		sendInternalServerErrorResponse(w)
		metrics.MetricCreateFileInternalErrors.Inc()
		return
	}

	err = sendJsonResponse(w, http.StatusCreated, response)
	if err != nil {
		metrics.MetricCreateFileInternalErrors.Inc()
	}

	metrics.MetricCreateFileResponses.Inc()
}

func isValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}

// Validate the create file request. Invalid requests are failed with an HTTP bad
// request error.
func isValidCreateFileRequest(requestID string,
	request *common.CreateFileRequest) bool {
	// Validate that the tenant ID specified is a valid UUID.
	if !isValidUUID(request.TenantID) {
		fsLogger.Error("Invalid tenant id",
			zap.String("Request ID", requestID),
			zap.String("Tenant ID", request.TenantID),
		)
		return false
	}

	// Validate that the device ID specified is a valid UUID.
	if !isValidUUID(request.DeviceID) {
		fsLogger.Error("Invalid device id",
			zap.String("Request ID", requestID),
			zap.String("Device ID", request.DeviceID),
		)
		return false
	}

	// Ensure the request provided a valid file name.
	if !isValidFileName(request.Name) {
		fsLogger.Error("Invalid file name",
			zap.String("Request ID", requestID),
			zap.String("File name", request.Name),
		)
		return false
	}

	// Ensure the request specified a non-empty checksum for the file.
	if !isValidChecksum(request.Checksum) {
		fsLogger.Error("Invalid checksum",
			zap.String("Request ID", requestID),
			zap.String("Checksum", request.Checksum),
		)
		return false
	}
	return true
}

// validate file name
func isValidFileName(name string) bool {
	nameLength := len(name)
	if nameLength < minFileNameLength || nameLength > maxFileNameLength {
		return false
	}
	return fileNameRegex.MatchString(name)
}

// validate checksum
func isValidChecksum(checksum string) bool {
	checksumLength := len(checksum)
	if checksumLength < minChecksumLength || checksumLength > maxChecksumLength {
		fsLogger.Error("Invalid checksum length",
			zap.String("Checksum", checksum),
			zap.Int("Length", checksumLength),
		)
		return false
	}
	_, err := base64.StdEncoding.DecodeString(checksum)
	if err != nil {
		fsLogger.Error("Invalid base64 checksum",
			zap.String("Checksum", checksum),
			zap.Error(err),
		)
		return false
	}
	return err == nil
}
