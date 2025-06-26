package db

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/HPInc/krypton-fs/service/cache"
	"github.com/HPInc/krypton-fs/service/metrics"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

// GetFile - retrieve information about the file corresponding to the specified
// file ID.
func GetFile(requestID string, id string) (*File, error) {
	var foundFile File

	// Check the parameters
	fileID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		fsLogger.Error("Failed to parse the specified file ID",
			zap.String("Request ID:", requestID),
			zap.Error(err),
		)
		return nil, err
	}

	// Find the file corresponding to the specified file ID in the files cache.
	cacheEntry, err := cache.GetFile(requestID, fileID)
	if err == nil {
		fsLogger.Debug("GetFile - cache hit!")
		err = json.Unmarshal([]byte(cacheEntry), &foundFile)
		if err != nil {
			fsLogger.Error("Failed to unmarshal file from cache",
				zap.String("Request ID: ", requestID),
				zap.String("File ID: ", id),
			)
		}
	}

	// File was not found in the cache. Check to see if it is available in
	// the database.
	if err != nil {
		start := time.Now()

		ctx, cancelFunc := context.WithTimeout(context.Background(), dbOperationTimeout)
		defer cancelFunc()
		defer metrics.ReportLatencyMetric(metrics.MetricDatabaseLatency, start,
			operationDbGetFile)

		response := gDbPool.QueryRow(ctx, queryFileByID, fileID)
		err = response.Scan(&foundFile.FileID, &foundFile.TenantID, &foundFile.DeviceID,
			&foundFile.Name, &foundFile.Checksum, &foundFile.Size, &foundFile.Status,
			&foundFile.CreatedAt, &foundFile.UpdatedAt, &foundFile.BucketName)
		if err != nil {
			fsLogger.Error("Failed to find the specified file in the database!",
				zap.String("Request ID:", requestID),
				zap.Uint64("File ID: ", fileID),
				zap.Error(err),
			)

			if errors.Is(err, pgx.ErrNoRows) {
				metrics.MetricDatabaseFileNotFoundErrors.Inc()
				return nil, ErrNotFound
			}

			metrics.MetricDatabaseGetFileFailures.Inc()
			return nil, ErrInternalError
		}

		metrics.MetricDatabaseFilesRetrieved.Inc()

		// Add the file to the cache on a separate goroutine.
		go cache.AddFile(requestID, foundFile.FileID, foundFile)
	}

	return &foundFile, nil
}

// ListFilesForDevice - list all files belonging to a specific device within
// the specified tenant.
func ListFilesForDevice(tenantID, deviceID string) ([]File, int64, error) {
	// Read the entity by filter on non-key attributes
	var fas []File
	start := time.Now()

	ctx, cancelFunc := context.WithTimeout(context.Background(), dbOperationTimeout)
	defer cancelFunc()
	defer metrics.ReportLatencyMetric(metrics.MetricDatabaseLatency, start,
		operationDbListFiles)

	response, err := gDbPool.Query(ctx, queryFilesForSpecificDevice, tenantID,
		deviceID)
	if err != nil {
		fsLogger.Error("Failed to get a list of files from the database!",
			zap.Error(err),
		)
		return nil, 0, err
	}
	defer response.Close()

	for response.Next() {
		var foundFile File
		err = response.Scan(&foundFile.FileID, &foundFile.TenantID,
			&foundFile.DeviceID, &foundFile.Name, &foundFile.Checksum,
			&foundFile.Size, &foundFile.Status, &foundFile.CreatedAt,
			&foundFile.UpdatedAt, &foundFile.BucketName)
		if err != nil {
			fsLogger.Error("Failed to get a list of files from the database!",
				zap.Error(err),
			)
			return nil, 0, err
		}
		fas = append(fas, foundFile)
	}

	if response.Err() != nil {
		fsLogger.Error("Failed reading list of buckets from the database!",
			zap.Error(response.Err()),
		)
		return nil, 0, response.Err()
	}

	return fas, int64(len(fas)), nil
}
