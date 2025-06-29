// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package db

import (
	"errors"
	"time"

	"github.com/HPInc/krypton-fs/service/metrics"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

func (b *Bucket) GetBucket(bucketName string) error {
	start := time.Now()

	ctx, cancelFunc := context.WithTimeout(context.Background(), dbOperationTimeout)
	defer cancelFunc()
	defer metrics.ReportLatencyMetric(metrics.MetricDatabaseLatency, start,
		operationDbGetBucket)

	response := gDbPool.QueryRow(ctx, queryGetBucketByName, bucketName)
	err := response.Scan(&b.BucketName, &b.IsArchived, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		fsLogger.Error("Failed to find the specified bucket in the database!",
			zap.String("Bucket name:", bucketName),
			zap.Error(err),
		)

		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return ErrInternalError
	}

	return nil
}
