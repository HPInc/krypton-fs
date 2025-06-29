// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package rest

import (
	"net/http"

	"github.com/HPInc/krypton-fs/service/db"
	"github.com/HPInc/krypton-fs/service/metrics"
	"go.uber.org/zap"
)

func DeleteFileHandler(w http.ResponseWriter, r *http.Request) {
	requestID := r.Header.Get(headerRequestID)

	// Retrieve the specified file identifier.
	fileID, err := getPathVariable(r, paramFileID, true)
	if err != nil {
		fsLogger.Error("The required file id path variable was not specified in the request",
			zap.String("Request ID:", requestID),
		)
		sendBadRequestErrorResponse(w)
		metrics.MetricDeleteFileBadRequests.Inc()
		return
	}

	// Delete the specified file from the database.
	err = db.DeleteFile(requestID, fileID)
	if err != nil {
		if err == db.ErrNotFound {
			fsLogger.Error("No file with the requested file ID was found in the database",
				zap.String("Request ID:", requestID),
			)
			sendNotFoundErrorResponse(w)
			metrics.MetricDeleteFileNotFoundErrors.Inc()
			return
		}

		fsLogger.Error("Failed to delete the specified file from the database!",
			zap.String("Request ID:", requestID),
			zap.Error(err),
		)
		sendInternalServerErrorResponse(w)
		metrics.MetricDeleteFileInternalErrors.Inc()
		return
	}

	err = sendJsonResponse(w, http.StatusNoContent, nil)
	if err != nil {
		metrics.MetricDeleteFileInternalErrors.Inc()
	}

	metrics.MetricDeleteFileResponses.Inc()
}
