package db

import (
	"context"
	"time"

	"github.com/HPInc/krypton-fs/service/cache"
	"github.com/HPInc/krypton-fs/service/common"
	"github.com/HPInc/krypton-fs/service/metrics"
	"go.uber.org/zap"
)

var (
	scavengerCtx        context.Context
	scavengerCancelFunc context.CancelFunc
	scavengerDone       chan bool
)

// roundToMidnight truncates time to midnight
func roundToMidnight() time.Time {
	timeNow := time.Now()
	return time.Date(timeNow.Year(), timeNow.Month(), timeNow.Day(), 23, 59, 59,
		0, timeNow.Location())
}

// Start the periodic scavenger goroutine.
func startScavenger() {
	var leaderLockAcquired bool

	// Initialize the scavenger timer to fire at midnight.
	nextRun := time.Until(roundToMidnight())
	scavengerTimer := time.NewTimer(nextRun)

	// Create the context used to listen for cancellation of scavenger runs or
	// service shutdown signals. Also create a channel for the scavenger
	// goroutine to signal when it is done shutting down.
	scavengerCtx, scavengerCancelFunc = context.WithCancel(context.Background())
	scavengerDone = make(chan bool, 1)

	for {
		fsLogger.Info("Scavenger has been configured for its next execution.",
			zap.Duration("Duration:", nextRun),
		)

		select {
		case <-scavengerTimer.C:
			leaderLockAcquired = false

			// Give ourselves a few retry attempts to acquire the leader lock.
			for i := maxDbAcquireLeaderLockRetries; i > 0; i-- {
				if !cache.AcquireLeaderLock() {
					// We failed to acquire the leader lock. Wait for as long as
					// the leader lock's lifetime before retrying.
					time.Sleep(cache.LeaderLockLifetime)
				} else {
					// Run the periodic scavenger.
					leaderLockAcquired = true
					RunScavenger()
					metrics.MetricScavengerRuns.Inc()

					// Release the leader lock. The scavenger run has completed.
					cache.ReleaseLeaderLock()
				}
			}

			if !leaderLockAcquired {
				fsLogger.Error("All attempts to acquire leader lock have failed. Bailing on this scavenger run ...")
				metrics.MetricAbandonedScavengerRun.Inc()
			}

			// Reset the scavenger timer to run the next day.
			nextRun = time.Until(roundToMidnight())
			_ = scavengerTimer.Reset(nextRun)
			continue

		case <-scavengerCtx.Done():
			fsLogger.Info("Scavenger has received shutdown signal and is stopping!")
			scavengerTimer.Stop()
			scavengerDone <- true
			return
		}
	}
}

// Stop the scavenger goroutine and wait for it to be done.
func stopScavenger() {
	if scavengerCancelFunc != nil {
		scavengerCancelFunc()
		<-scavengerDone
	}
}

func RunScavenger() {
	startTime := time.Now()
	defer common.TimeIt(fsLogger, startTime, "DB Scavenger")

	fsLogger.Info("Executing the database scavenger ...",
		zap.Time("Start time:", startTime),
	)

	// Scavenge files older than the configured threshold (expired files).
	if err := scavengeExpiredFiles(); err != nil {
		fsLogger.Info("Scavenger run failed", zap.Error(err))
	}

	fsLogger.Info("The database scavenger run has completed.")
}

// ////////////////////////  Phase 1 scavenge  /////////////////////////////////
// In this phase, files that have been created before the configured time
// threshold are tombstoned. An entry for the file is created in the
// tombstoned_files table and its entry is deleted from the files table.
// /////////////////////////////////////////////////////////////////////////////
func scavengeExpiredFiles() error {
	startTime := time.Now()
	defer common.TimeIt(fsLogger, startTime, "scavengeExpiredFiles")

	// Check if the scavenger's context is no longer valid. This means we've been
	// instructed to stop and likely the service is shutting down or encountered
	// an error.
	if scavengerCtx.Err() != nil {
		fsLogger.Info("Aborting expired files scavenger run. Context has been cancelled.")
		return scavengerCtx.Err()
	}

	// Get a list of candidate files from the files table.
	err := deleteExpiredFiles()
	if err != nil {
		fsLogger.Error("Failed to query for scavengeable expired files!",
			zap.Error(err),
		)
		metrics.MetricScavengeExpiredFileFailures.Inc()
		return err
	}

	return nil
}
