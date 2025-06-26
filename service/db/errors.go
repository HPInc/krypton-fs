package db

import (
	"context"
	"errors"

	"github.com/HPInc/krypton-fs/service/metrics"
	"github.com/jackc/pgerrcode"
	pgxv5 "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

var (
	ErrNoBuckets      = errors.New("no buckets have configured for the service")
	ErrDuplicateEntry = errors.New("a duplicate entry was found in the database")
	ErrNotFound       = errors.New("the requested entry was not found in the database")
	ErrNotAllowed     = errors.New("the requested operation is not allowed")
	ErrInvalidRequest = errors.New("the request contained one or more invalid parameters")
	ErrInternalError  = errors.New("an internal error occured while performing the database operation")
)

func isDuplicateKeyError(err error) bool {
	pgErr, ok := err.(*pgconn.PgError)
	if ok {
		if pgErr.Code == pgerrcode.UniqueViolation {
			return true
		}
	}
	return false
}

func commit(tx pgxv5.Tx, ctx context.Context) {
	err := tx.Commit(ctx)
	if err != nil {
		fsLogger.Error("Failed to commit transaction!",
			zap.Error(err),
		)
		metrics.MetricDatabaseCommitErrors.Inc()
	}
}

func rollback(tx pgxv5.Tx, ctx context.Context) {
	err := tx.Rollback(ctx)
	if err != nil {
		fsLogger.Error("Failed to rollback transaction!",
			zap.Error(err),
		)
		metrics.MetricDatabaseRollbackErrors.Inc()
	}
}
