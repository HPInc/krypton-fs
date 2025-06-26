package db

import (
	"time"

	"github.com/HPInc/krypton-fs/service/cache"
	"github.com/HPInc/krypton-fs/service/common"
	"github.com/HPInc/krypton-fs/service/metrics"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

// CreateFile new record allocating new ID
func CreateFile(requestID string, request *common.CreateFileRequest) (*File, error) {
	var newFile File
	start := time.Now()

	ctx, cancelFunc := context.WithTimeout(context.Background(), dbOperationTimeout)
	defer cancelFunc()
	defer metrics.ReportLatencyMetric(metrics.MetricDatabaseLatency, start,
		operationDbCreateFile)

	tx, err := gDbPool.Begin(ctx)
	if err != nil {
		fsLogger.Error("Failed to acquire transaction to create file!",
			zap.Error(err),
		)
		return nil, err
	}

	response := tx.QueryRow(ctx, queryInsertNewFile, request.TenantID, request.DeviceID, request.Name,
		request.Checksum, request.Size, FileStatusNew, selectBucket())
	err = response.Scan(&newFile.FileID, &newFile.TenantID, &newFile.DeviceID, &newFile.Name,
		&newFile.Checksum, &newFile.Size, &newFile.Status, &newFile.CreatedAt, &newFile.UpdatedAt,
		&newFile.BucketName)
	if err != nil {
		rollback(tx, ctx)
		if isDuplicateKeyError(err) {
			fsLogger.Error("Failed to create a new file. Duplicate exists!",
				zap.String("Request ID:", requestID),
				zap.Error(err),
			)
			return nil, ErrDuplicateEntry
		}

		fsLogger.Error("Failed to create a new file.",
			zap.String("Request ID:", requestID),
			zap.Error(err),
		)
		return nil, err
	}
	commit(tx, ctx)

	// Add the file to the cache on a separate goroutine.
	go cache.AddFile(requestID, newFile.FileID, newFile)

	return &newFile, nil
}
