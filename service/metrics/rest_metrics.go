// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	// REST request processing latency is partitioned by the REST method. It uses
	// custom buckets based on the expected request duration.
	MetricRestLatency = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       "fs_rest_latency_milliseconds",
			Help:       "A latency histogram for REST requests served by FS",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{"method"},
	)

	// Number of REST requests received by FS.
	MetricRequestCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        "fs_rest_requests",
		Help:        "Number of requests received by FS",
		ConstLabels: prometheus.Labels{"version": "1"},
	})

	// Number of internal errors encountered when processing create file requests.
	MetricCreateFileInternalErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "fs_rest_create_file_internal_errors",
			Help: "Total number of internal errors encountered processing create file requests",
		})

	// Number of bad create file requests encountered.
	MetricCreateFileBadRequests = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "fs_rest_create_file_bad_requests",
			Help: "Total number of bad create file requests",
		})

	// Number of successful create file requests served.
	MetricCreateFileResponses = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "fs_rest_create_file_requests",
			Help: "Total number of successful create file requests served by FS",
		})

	// Number of internal errors encountered when processing get file requests.
	MetricGetFileInternalErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "fs_rest_get_file_internal_errors",
			Help: "Total number of internal errors encountered processing get file requests",
		})

	// Number of bad get file requests encountered.
	MetricGetFileBadRequests = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "fs_rest_get_file_bad_requests",
			Help: "Total number of bad get file requests",
		})

	// Number of get file requests where the requested file was not found.
	MetricGetFileNotFoundErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "fs_rest_get_file_not_found_errors",
			Help: "Total number of get file requests where file was not found",
		})

	// Number of get file requests where an invalid access was requested
	MetricGetFileInvalidAccessErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "fs_rest_get_file_invalid_access_errors",
			Help: "Total number of get file requests where requested file did not belong to tenant/device",
		})

	// Number of successful get file requests served.
	MetricGetFileResponses = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "fs_rest_get_file_requests",
			Help: "Total number of successful get file requests served by FS",
		})

	// Number of internal errors encountered when processing list file requests.
	MetricListFilesInternalErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "fs_rest_list_files_internal_errors",
			Help: "Total number of internal errors encountered processing list files requests",
		})

	// Number of bad list file requests encountered.
	MetricListFilesBadRequests = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "fs_rest_list_files_bad_requests",
			Help: "Total number of bad list file requests",
		})

	// Number of get file requests where the requested file was not found.
	MetricListFilesNotFoundErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "fs_rest_list_files_not_found_errors",
			Help: "Total number of list file requests where file was not found",
		})

	// Number of successful list file requests served.
	MetricListFilesResponses = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "fs_rest_list_file_requests",
			Help: "Total number of successful list file requests served by FS",
		})

	// Number of internal errors encountered when processing delete file requests.
	MetricDeleteFileInternalErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "fs_rest_delete_file_internal_errors",
			Help: "Total number of internal errors encountered processing delete file requests",
		})

	// Number of bad delete file requests encountered.
	MetricDeleteFileBadRequests = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "fs_rest_delete_file_bad_requests",
			Help: "Total number of bad delete file requests",
		})

	// Number of delete file requests where the requested file was not found.
	MetricDeleteFileNotFoundErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "fs_rest_delete_file_not_found_errors",
			Help: "Total number of delete file requests where file was not found",
		})

	// Number of successful delete file requests served.
	MetricDeleteFileResponses = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "fs_rest_delete_file_requests",
			Help: "Total number of successful delete file requests served by FS",
		})

	// Number of internal errors encountered when processing get signed URL requests.
	MetricGetSignedUrlInternalErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "fs_rest_signed_url_internal_errors",
			Help: "Total number of internal errors encountered processing get signed url requests",
		})

	// Number of bad get signed URL requests encountered.
	MetricGetSignedUrlBadRequests = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "fs_rest_signed_url_bad_requests",
			Help: "Total number of bad get signed url requests",
		})

	// Number of update file requests where the requested file was not found.
	MetricGetSignedUrlFileNotFoundErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "fs_rest_signed_url_file_not_found_errors",
			Help: "Total number of get signed url requests where file was not found",
		})

	// Number of signed url requests where the requested file was quarantined
	MetricGetSignedUrlForbiddenErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "fs_rest_signed_url_forbidden_errors",
			Help: "Total number of get signed url requests where file was quarantined",
		})

	// Number of successful get signed URL requests served.
	MetricGetSignedUrlResponses = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "fs_rest_signed_url_requests",
			Help: "Total number of successful get signed url requests served by FS",
		})

	// Number of unauthorized requests encountered for create_file.
	MetricCreateFileUnauthorizedRequests = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "fs_rest_create_file_unauthorized_requests",
			Help: "Total number of create file unauthorized requests",
		})

	// Number of unsupported media type requests encountered for create_file.
	MetricCreateFileUnSupportedMediaTypeRequests = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "fs_rest_create_file_unsupported_mediatype_requests",
			Help: "Total number of create file unsupported media type requests",
		})

	// Number of unauthorized requests encountered for get_file.
	MetricGetFileUnauthorizedRequests = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "fs_rest_get_file_unauthorized_requests",
			Help: "Total number of get file unauthorized requests",
		})
)
