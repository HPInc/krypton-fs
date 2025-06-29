// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package db

import (
	"time"

	"github.com/HPInc/krypton-fs/service/metrics"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

func (b *Bucket) ListBuckets() (*[]Bucket, error) {
	var foundBuckets []Bucket

	start := time.Now()

	ctx, cancelFunc := context.WithTimeout(context.Background(), dbOperationTimeout)
	defer cancelFunc()
	defer metrics.ReportLatencyMetric(metrics.MetricDatabaseLatency, start,
		operationDbListBuckets)

	response, err := gDbPool.Query(ctx, queryGetEnabledBuckets)
	if err != nil {
		fsLogger.Error("Failed to get a list of buckets from the database!",
			zap.Error(err),
		)
		return nil, err
	}
	defer response.Close()

	for response.Next() {
		var b Bucket
		err = response.Scan(&b.BucketName, &b.IsArchived, &b.CreatedAt, &b.UpdatedAt)
		if err != nil {
			fsLogger.Error("Failed to get a list of buckets from the database!",
				zap.Error(err),
			)
			return nil, err
		}
		foundBuckets = append(foundBuckets, b)
	}

	if response.Err() != nil {
		fsLogger.Error("Failed reading list of buckets from the database!",
			zap.Error(response.Err()),
		)
		return nil, response.Err()
	}
	return &foundBuckets, nil
}
