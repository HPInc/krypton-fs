// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	// Cache request processing latency is partitioned by the Redis method. It uses
	// custom buckets based on the expected request duration.
	MetricCacheLatency = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       "fs_cache_latency_milliseconds",
			Help:       "A latency histogram for cache operations issued by the FS",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{"method"},
	)

	// Total number of failed cache set device operations.
	MetricCacheSetFileFailures = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "fs_cache_set_file_failures",
			Help: "Total number of failed cache set device operations",
		})

	// Total number of failed cache get device operations.
	MetricCacheGetFileFailures = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "fs_cache_get_file_failures",
			Help: "Total number of failed cache get device operations",
		})

	// Total number of cache hits for get device operations.
	MetricCacheGetFileCacheHits = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "fs_cache_get_file_cache_hits",
			Help: "Total number of cache hits for get device operations",
		})

	// Total number of cache misses for get device operations.
	MetricCacheGetFileCacheMisses = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "fs_cache_get_file_cache_misses",
			Help: "Total number of cache misses for get device operations",
		})

	// Total number of failed cache delete device operations.
	MetricCacheDelFileFailures = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "fs_cache_del_file_failures",
			Help: "Total number of failed cache delete device operations",
		})
)
