// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package db

import (
	"errors"
	"strconv"
	"time"

	"github.com/HPInc/krypton-fs/service/cache"
	"github.com/HPInc/krypton-fs/service/metrics"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

// mark status as uploaded. update size with incoming size
// return nil on success
// return err on error
func updateFileStatus(id, status string, size int64) error {
	// Check the parameters
	fileID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		fsLogger.Error("Failed to parse the specified file ID",
			zap.Error(err),
		)
		return err
	}

	start := time.Now()

	ctx, cancelFunc := context.WithTimeout(context.Background(), dbOperationTimeout)
	defer cancelFunc()
	defer metrics.ReportLatencyMetric(metrics.MetricDatabaseLatency, start,
		operationDbUpdateFile)

	tx, err := gDbPool.Begin(ctx)
	if err != nil {
		fsLogger.Error("Failed to acquire transaction to update file status!",
			zap.Error(err),
		)
		return err
	}

	var retFileID uint64
	response := tx.QueryRow(ctx, queryUpdateFileStatus, fileID, size, status)
	err = response.Scan(&retFileID)
	if err != nil {
		rollback(tx, ctx)
		if errors.Is(err, pgx.ErrNoRows) {
			fsLogger.Error("File with the specified file ID was not found in the database!",
				zap.Uint64("File ID: ", fileID),
				zap.Error(err),
			)
			metrics.MetricDatabaseFileNotFoundErrors.Inc()
			return ErrNotFound
		}

		fsLogger.Error("Failed to update the file in the database!",
			zap.Error(err),
		)
		metrics.MetricDatabaseUpdateFileFailures.Inc()
		return ErrInternalError
	}

	commit(tx, ctx)
	metrics.MetricDatabaseFilesUpdated.Inc()

	// Remove the cache entry on a separate goroutine. The next subsequent
	// read of this file will refresh the cache entry.
	go cache.RemoveFile("", fileID)

	return nil
}

// mark status = uploaded
func MarkFileUploaded(id string, size int64) error {
	return updateFileStatus(id, FileStatusUploaded, size)
}

// make status = quarantined
func MarkFileQuarantined(id string, size int64) error {
	return updateFileStatus(id, FileStatusQuarantined, size)
}
