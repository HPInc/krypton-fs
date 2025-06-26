package rest

import (
	"net/http"
	"time"

	"github.com/HPInc/krypton-fs/service/common"
	"github.com/HPInc/krypton-fs/service/db"
	"github.com/HPInc/krypton-fs/service/metrics"
	"go.uber.org/zap"
)

// Lists files matching the requested filter. Scoped to a single tenant and
// device at a time.
func ListFilesHandler(w http.ResponseWriter, r *http.Request) {
	requestID := r.Header.Get(headerRequestID)

	// Extract the tenant ID and the device ID from the request.
	tenantID := r.FormValue(paramTenantID)
	if tenantID == "" {
		fsLogger.Error("No tenant was specified in the request!",
			zap.String("Request ID:", requestID),
		)
		sendBadRequestErrorResponse(w)
		metrics.MetricListFilesBadRequests.Inc()
		return
	}

	deviceID := r.FormValue(paramDeviceID)
	if deviceID == "" {
		fsLogger.Error("No device was specified in the request!",
			zap.String("Request ID:", requestID),
		)
		sendBadRequestErrorResponse(w)
		metrics.MetricListFilesBadRequests.Inc()
		return
	}

	// Get a list of files matching the requested filter.
	foundFiles, count, err := db.ListFilesForDevice(tenantID, deviceID)
	if err != nil {
		fsLogger.Error("Failed to list files matching the requested filter in the database!",
			zap.String("Request ID:", requestID),
			zap.Error(err),
		)
		sendInternalServerErrorResponse(w)
		metrics.MetricListFilesInternalErrors.Inc()
		return
	}

	response := common.ListFilesResponse{
		RequestID:    requestID,
		ResponseTime: time.Now(),
		Count:        count,
		Files:        nil,
	}
	for _, item := range foundFiles {
		response.Files = append(response.Files, common.FileInformation{
			FileID:    item.FileID,
			TenantID:  item.TenantID,
			DeviceID:  item.DeviceID,
			Name:      item.Name,
			Checksum:  item.Checksum,
			Size:      item.Size,
			Status:    item.Status,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		})
	}

	err = sendJsonResponse(w, http.StatusOK, response)
	if err != nil {
		metrics.MetricListFilesInternalErrors.Inc()
	}

	metrics.MetricListFilesResponses.Inc()
}
