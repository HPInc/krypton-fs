// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package rest

import (
	"net/http"
	"time"

	"github.com/HPInc/krypton-fs/service/common"
	"github.com/HPInc/krypton-fs/service/db"
	"github.com/HPInc/krypton-fs/service/metrics"
	"github.com/HPInc/krypton-fs/service/storage"
	"go.uber.org/zap"
)

// GetSignedUrlHandler gets a presigned GET/PUT/HEAD URL to perform an operation
// on the existing file store resource assuming that the resource does exist,
// ie. was previously created by POST using presigned URL.
func GetSignedUrlHandler(w http.ResponseWriter, r *http.Request) {
	requestID := r.Header.Get(headerRequestID)

	// Retrieve the specified file identifier.
	fileID, err := getPathVariable(r, "id", true)
	if err != nil {
		fsLogger.Error("The required file id path variable was not specified in the request",
			zap.String("Request ID:", requestID),
		)
		sendBadRequestErrorResponse(w)
		metrics.MetricGetSignedUrlBadRequests.Inc()
		return
	}

	// Extract parameters from the request.
	err = r.ParseForm()
	if err != nil {
		fsLogger.Error("Failed to parse the request form!",
			zap.String("Request ID:", requestID),
			zap.Error(err),
		)
		sendBadRequestErrorResponse(w)
		metrics.MetricGetSignedUrlBadRequests.Inc()
		return
	}

	method := r.Form.Get(paramMethod)
	if method == "" {
		fsLogger.Error("The required HTTP method was not specified in the request",
			zap.String("Request ID:", requestID),
		)
		sendBadRequestErrorResponse(w)
		metrics.MetricGetSignedUrlBadRequests.Inc()
		return
	}

	// Retrieve information about the file corresponding to this ID.
	foundFile, err := db.GetFile(requestID, fileID)
	if err != nil {
		if err == db.ErrNotFound {
			fsLogger.Error("No file with the requested file ID was found in the database",
				zap.String("Request ID:", requestID),
			)
			sendNotFoundErrorResponse(w)
			metrics.MetricGetSignedUrlFileNotFoundErrors.Inc()
			return
		}

		fsLogger.Error("Failed to read information about file from the database!",
			zap.String("Request ID:", requestID),
			zap.Error(err),
		)
		sendInternalServerErrorResponse(w)
		metrics.MetricGetSignedUrlInternalErrors.Inc()
		return
	}
	// if file status is quarantined, return 403
	if foundFile.Status == db.FileStatusQuarantined {
		err = sendJsonResponse(w, http.StatusForbidden, nil)
		if err != nil {
			metrics.MetricGetSignedUrlForbiddenErrors.Inc()
		}
		return
	}

	response := common.SignedUrlResponse{
		RequestID:    requestID,
		ResponseTime: time.Now(),
		FileName:     foundFile.Name,
	}

	// Generate a signed URL for the file - the signed URL generated corresponds
	// to the requested HTTP method.
	response.SignedUrl, err = storage.Provider.GetSignedUrl(
		foundFile.BucketName,
		storage.GetObjectName(foundFile.TenantID, foundFile.DeviceID, foundFile.FileID),
		method,
		foundFile.Checksum,
		foundFile.Size)
	if err != nil {
		fsLogger.Error("Failed to generate a signed URL for the file!",
			zap.String("Request ID:", requestID),
			zap.Error(err),
		)
		sendInternalServerErrorResponse(w)
		metrics.MetricGetSignedUrlInternalErrors.Inc()
		return
	}

	// JSON encode and return information about the file.
	err = sendJsonResponse(w, http.StatusOK, response)
	if err != nil {
		metrics.MetricGetSignedUrlInternalErrors.Inc()
	}

	metrics.MetricGetSignedUrlResponses.Inc()
}
