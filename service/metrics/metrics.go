package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// ReportLatencyMetric reports the latency of the specified operation to the
// specified summary vector metric. The label is used to partition the resulting
// histogram.
func ReportLatencyMetric(metric *prometheus.SummaryVec,
	startTime time.Time, label string) {
	duration := time.Since(startTime)
	metric.WithLabelValues(label).Observe(float64(duration.Milliseconds()))
}

func RegisterPrometheusMetrics() {
	prometheus.MustRegister(MetricRestLatency)
	prometheus.MustRegister(MetricDatabaseLatency)
	prometheus.MustRegister(MetricCacheLatency)
}
