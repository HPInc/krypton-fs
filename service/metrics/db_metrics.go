// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	// Database request processing latency is partitioned by the Postgres method. It uses
	// custom buckets based on the expected request duration.
	MetricDatabaseLatency = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       "fs_db_latency_milliseconds",
			Help:       "A latency histogram for database operations issued by FS",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{"method"},
	)

	// Total number of errors committing database transactions.
	MetricDatabaseCommitErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "fs_db_commit_errors",
			Help: "Total number of errors committing transactions",
		})

	// Total number of errors rolling back database transactions.
	MetricDatabaseRollbackErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "fs_db_rollback_errors",
			Help: "Total number of errors rolling back transactions",
		})

	// Total number of successful scavenger runs.
	MetricScavengerRuns = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "fs_db_scavenger_runs",
			Help: "Total number of successful scavenger runs",
		})

	// Total number of abandoned scavenger runs.
	MetricAbandonedScavengerRun = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "fs_db_abandoned_scavenger_runs",
			Help: "Total number of abandoned scavenger runs",
		})

	// Total number of errors scavenging expired files.
	MetricScavengeExpiredFileFailures = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "fs_db_scavenge_expired_failures",
			Help: "Total number of errors scavenging expired files",
		})

	// Total number of expired files that were scavenged.
	MetricScavengeExpiredFiles = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "fs_db_scavenge_expired_files",
			Help: "Total number of expired files that were scavenged",
		})

	// Total number of errors scavenging tombstoned files.
	MetricScavengeTombstonedFileFailures = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "fs_db_scavenge_tombstoned_failures",
			Help: "Total number of errors scavenging tombstoned files",
		})

	// Total number of tombstoned files that were scavenged.
	MetricScavengeTombstonedFiles = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "fs_db_scavenge_tombstoned_files",
			Help: "Total number of tombstoned files that were scavenged",
		})

	// Total number of times the requested file was not found in the database.
	MetricDatabaseFileNotFoundErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "fs_db_file_not_found_errors",
			Help: "Total number of times file was not found in database",
		})

	// Total number of failed database get file operations.
	MetricDatabaseGetFileFailures = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "fs_db_get_file_failures",
			Help: "Total number of failed get file database operations",
		})

	// Total number of failed database update file operations.
	MetricDatabaseUpdateFileFailures = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "fs_db_update_file_failures",
			Help: "Total number of failed update file database operations",
		})

	// Total number of failed database delete file operations.
	MetricDatabaseDeleteFileFailures = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "fs_db_delete_file_failures",
			Help: "Total number of failed delete file database operations",
		})

	// Total number of get file database operations.
	MetricDatabaseFilesRetrieved = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "fs_db_files_retrieved",
			Help: "Total number of get file database operations",
		})

	// Total number of update file database operations.
	MetricDatabaseFilesUpdated = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "fs_db_files_updated",
			Help: "Total number of update file database operations",
		})

	// Total number of delete file database operations.
	MetricDatabaseFilesDeleted = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "fs_db_files_deleted",
			Help: "Total number of delete file database operations",
		})
)
