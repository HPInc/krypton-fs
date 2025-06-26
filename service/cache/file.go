package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/HPInc/krypton-fs/service/metrics"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// AddFile - cache information about the specified file. This function is
// typically called from a goroutine and errors adding to the cache are not
// surfaced to the caller.
func AddFile(requestID string, fileID uint64, file interface{}) {
	if !isEnabled {
		return
	}

	// Marshal the file object for caching.
	cacheEntry, err := json.Marshal(file)
	if err != nil {
		fsLogger.Error("Failed to marshal file for caching!",
			zap.String("Request ID: ", requestID),
			zap.Uint64("File ID: ", fileID),
			zap.Error(err),
		)
		return
	}

	// Add the file to the cache.
	ctx, cancelFunc := context.WithTimeout(gCtx, cacheTimeout)
	defer cancelFunc()

	start := time.Now()
	err = cacheClient.Set(ctx, fmt.Sprintf(filePrefix, fileID),
		cacheEntry, ttlFile).Err()
	metrics.ReportLatencyMetric(metrics.MetricCacheLatency, start,
		operationCacheSet)
	if err != nil {
		fsLogger.Error("Failed to add the file to the cache!",
			zap.String("Request ID: ", requestID),
			zap.Uint64("File ID: ", fileID),
			zap.Error(err),
		)
		metrics.MetricCacheSetFileFailures.Inc()
	}
}

// GetFile - retrieve information about a file object from the cache.
func GetFile(requestID string, fileID uint64) ([]byte, error) {
	if !isEnabled {
		return nil, ErrCacheNotFound
	}

	// Get the requested file object from the cache.
	ctx, cancelFunc := context.WithTimeout(gCtx, cacheTimeout)
	defer cancelFunc()

	start := time.Now()
	cacheEntry, err := cacheClient.Get(ctx,
		fmt.Sprintf(filePrefix, fileID)).Result()
	metrics.ReportLatencyMetric(metrics.MetricCacheLatency, start,
		operationCacheGet)
	if err != nil {
		if err == redis.Nil {
			metrics.MetricCacheGetFileCacheMisses.Inc()
			return nil, ErrCacheNotFound
		}

		fsLogger.Error("Error while looking up the file in the cache!",
			zap.String("Request ID: ", requestID),
			zap.Uint64("File ID: ", fileID),
			zap.Error(err),
		)
		metrics.MetricCacheGetFileFailures.Inc()
		return nil, err
	}

	metrics.MetricCacheGetFileCacheHits.Inc()
	return []byte(cacheEntry), nil
}

// RemoveFile - remove cached information about the specified file. This
// function is typically called from within a goroutine and errors removing
// from the cache are not surfaced to the caller.
func RemoveFile(requestID string, fileID uint64) {
	if !isEnabled {
		return
	}

	// Delete the requested file object from the cache.
	ctx, cancelFunc := context.WithTimeout(gCtx, cacheTimeout)
	defer cancelFunc()

	start := time.Now()
	err := cacheClient.Del(ctx, fmt.Sprintf(filePrefix, fileID)).Err()
	metrics.ReportLatencyMetric(metrics.MetricCacheLatency, start,
		operationCacheDel)
	if err != nil {
		fsLogger.Error("Failed to remove the file from the cache!",
			zap.String("Request ID: ", requestID),
			zap.Uint64("File ID: ", fileID),
			zap.Error(err),
		)
		metrics.MetricCacheDelFileFailures.Inc()
	}
}
