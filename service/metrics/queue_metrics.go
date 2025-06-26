package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	// Total number of file upload notifications processed successfully.
	MetricUploadNotificationsProcessed = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "fs_queue_upload_notifications",
			Help: "Total number of file upload notifications processed successfully",
		})

	// Total number of errors processing file upload notifications.
	MetricUploadNotificationProcessingErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "fs_queue_upload_notification_processing_errors",
			Help: "Total number of errors processing file upload notifications",
		})

	// Total number of errors parsing file upload notifications.
	MetricUploadNotificationParsingErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "fs_queue_upload_notification_parsing_errors",
			Help: "Total number of errors parsing file upload notifications",
		})
)
