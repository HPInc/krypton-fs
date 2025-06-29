// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package db

import (
	"time"

	"github.com/HPInc/krypton-fs/service/metrics"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

func (b *Bucket) ArchiveBucket() error {
	start := time.Now()

	ctx, cancelFunc := context.WithTimeout(context.Background(), dbOperationTimeout)
	defer cancelFunc()
	metrics.ReportLatencyMetric(metrics.MetricDatabaseLatency, start,
		operationDbUpdateBucket)

	tx, err := gDbPool.Begin(ctx)
	if err != nil {
		fsLogger.Error("Failed to acquire transaction to add bucket!",
			zap.Error(err),
		)
		return err
	}

	_, err = tx.Exec(ctx, queryArchiveBucket, b.BucketName)
	if err != nil {
		rollback(tx, ctx)
		if isDuplicateKeyError(err) {
			fsLogger.Debug("Bucket already exists in the database!",
				zap.String("Bucket name:", b.BucketName),
				zap.Error(err),
			)
			return nil
		}

		fsLogger.Error("Failed to mark the requested bucket as archived in the database!",
			zap.String("Bucket name:", b.BucketName),
			zap.Error(err),
		)
		return err
	}

	commit(tx, ctx)
	return nil
}
