// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package rest

import (
	"net/http"
	"time"

	"github.com/HPInc/krypton-fs/service/common"
	"github.com/HPInc/krypton-fs/service/db"
	"github.com/HPInc/krypton-fs/service/metrics"
	"go.uber.org/zap"
)

// Retrieves information about the file with the specified file ID from the
// database.
func GetFileHandler(w http.ResponseWriter, r *http.Request) {
	requestID := r.Header.Get(headerRequestID)

	// Retrieve the specified file identifier.
	fileID, err := getPathVariable(r, paramFileID, true)
	if err != nil {
		fsLogger.Error("The required file id path variable was not specified in the request",
			zap.String("Request ID:", requestID),
		)
		sendBadRequestErrorResponse(w)
		metrics.MetricGetFileBadRequests.Inc()
		return
	}

	// validate device token
	info, err := getDeviceInfoFromToken(r)
	if err != nil {
		fsLogger.Info("GetFile token validation error",
			zap.Error(err))
		sendUnauthorizedErrorResponse(w)
		metrics.MetricGetFileUnauthorizedRequests.Inc()
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
			metrics.MetricGetFileNotFoundErrors.Inc()
			return
		}

		fsLogger.Error("Failed to read information about file from the database!",
			zap.String("Request ID:", requestID),
			zap.Error(err),
		)
		sendInternalServerErrorResponse(w)
		metrics.MetricGetFileInternalErrors.Inc()
		return
	}

	// check if found file belongs to tenant and device
	if foundFile.TenantID != info.TenantID ||
		foundFile.DeviceID != info.DeviceID {
		fsLogger.Error("Attempted file read does not match auth!",
			zap.String("Request ID:", requestID),
			zap.String("Token Tenant ID:", info.TenantID),
			zap.String("Token Device ID:", info.DeviceID),
			zap.String("File Tenant ID:", foundFile.TenantID),
			zap.String("File Device ID:", foundFile.DeviceID),
			zap.Uint64("File ID:", foundFile.FileID),
			zap.Error(err))
		sendNotFoundErrorResponse(w)
		metrics.MetricGetFileNotFoundErrors.Inc()
		metrics.MetricGetFileInvalidAccessErrors.Inc()
		return
	}

	response := common.CommonFileResponse{
		RequestID:    requestID,
		ResponseTime: time.Now(),
		File: common.FileInformation{
			FileID:    foundFile.FileID,
			TenantID:  foundFile.TenantID,
			DeviceID:  foundFile.DeviceID,
			Name:      foundFile.Name,
			Checksum:  foundFile.Checksum,
			Size:      foundFile.Size,
			Status:    foundFile.Status,
			CreatedAt: foundFile.CreatedAt,
			UpdatedAt: foundFile.UpdatedAt,
		},
	}

	// JSON encode and return information about the file.
	err = sendJsonResponse(w, http.StatusOK, response)
	if err != nil {
		metrics.MetricGetFileInternalErrors.Inc()
	}

	metrics.MetricGetFileResponses.Inc()
}
