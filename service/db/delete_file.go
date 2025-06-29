// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package db

import (
	"errors"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
	"golang.org/x/net/context"

	"github.com/HPInc/krypton-fs/service/cache"
	"github.com/HPInc/krypton-fs/service/metrics"
)

// DeleteFile deletes the specified file from the files table. It creates an
// entry for the file in the tombstoned_files table.
func DeleteFile(requestID string, id string) error {
	// Check the parameters
	fileID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		fsLogger.Error("Failed to parse the specified file ID",
			zap.String("Request ID:", requestID),
			zap.Error(err),
		)
		return err
	}

	start := time.Now()

	ctx, cancelFunc := context.WithTimeout(context.Background(), dbOperationTimeout)
	defer cancelFunc()
	defer metrics.ReportLatencyMetric(metrics.MetricDatabaseLatency, start,
		operationDbDeleteFile)

	tx, err := gDbPool.Begin(ctx)
	if err != nil {
		fsLogger.Error("Failed to acquire transaction to delete file!",
			zap.Error(err),
		)
		return err
	}

	_, err = tx.Exec(ctx, deleteFileByID, fileID)
	if err != nil {
		rollback(tx, ctx)

		fsLogger.Error("Failed to delete the requested file from the database!",
			zap.String("Request ID:", requestID),
			zap.Uint64("File ID: ", fileID),
			zap.Error(err),
		)
		if errors.Is(err, pgx.ErrNoRows) {
			fsLogger.Error("No matching file was found in the database!",
				zap.String("Request ID:", requestID),
				zap.Uint64("File ID: ", fileID),
			)
			metrics.MetricDatabaseFileNotFoundErrors.Inc()
			return ErrNotFound
		}

		metrics.MetricDatabaseDeleteFileFailures.Inc()
		return ErrInternalError
	}
	commit(tx, ctx)
	metrics.MetricDatabaseFilesDeleted.Inc()

	// Remove the device from the cache on a separate goroutine.
	go cache.RemoveFile(requestID, fileID)

	return nil
}

// Delete candidate expired files from the files table.
func deleteExpiredFiles() error {
	start := time.Now()

	ctx, cancelFunc := context.WithTimeout(context.Background(), dbOperationTimeout)
	defer cancelFunc()
	defer metrics.ReportLatencyMetric(metrics.MetricDatabaseLatency, start,
		operationDbDeleteExpiredFiles)

	tx, err := gDbPool.Begin(ctx)
	if err != nil {
		fsLogger.Error("Failed to acquire transaction to add bucket!",
			zap.Error(err),
		)
		return err
	}

	ct, err := tx.Exec(ctx, queryDeleteExpiredFiles,
		time.Now().AddDate(0, 0, scavengeExpiredFilesThreshold))
	if err != nil {
		rollback(tx, ctx)

		fsLogger.Error("Failed to delete the expired files from the database!",
			zap.Error(err),
		)
		if errors.Is(err, pgx.ErrNoRows) {
			fsLogger.Error("No matching files were found in the database!")
			metrics.MetricDatabaseFileNotFoundErrors.Inc()
			return ErrNotFound
		}

		metrics.MetricScavengeExpiredFileFailures.Inc()
		return ErrInternalError
	}
	commit(tx, ctx)

	fsLogger.Info("Deleted expired files from the the database!",
		zap.Int("Number of files deleted:", int(ct.RowsAffected())),
	)
	metrics.MetricScavengeExpiredFiles.Add(float64(ct.RowsAffected()))

	return nil
}
