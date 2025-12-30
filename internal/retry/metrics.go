package retry

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// RetryMetrics holds metrics for retry operations
type RetryMetrics struct {
	// retriesTotal is a counter for total retry attempts by tool and result
	retriesTotal *prometheus.CounterVec
	// retryAttempts is a histogram for distribution of retry attempts
	retryAttempts *prometheus.HistogramVec
	// retryDelaySeconds is a histogram for distribution of retry delays
	retryDelaySeconds *prometheus.HistogramVec

	mu sync.Mutex
}

var (
	// globalMetrics is the singleton metrics instance
	globalMetrics *RetryMetrics
	// metricsOnce ensures metrics are initialized only once
	metricsOnce sync.Once
)

// initMetrics initializes the metrics collector
func initMetrics() {
	globalMetrics = &RetryMetrics{}

	// Initialize retries total counter
	globalMetrics.retriesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcp_retries_total",
			Help: "Total number of retry attempts per tool and result",
		},
		[]string{"tool", "result"}, // result: success, failed, exhausted
	)

	// Initialize retry attempts histogram
	globalMetrics.retryAttempts = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mcp_retry_attempts",
			Help:    "Distribution of retry attempts",
			Buckets: []float64{1, 2, 3, 4, 5, 10, 20},
		},
		[]string{"tool"},
	)

	// Initialize retry delay seconds histogram
	globalMetrics.retryDelaySeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mcp_retry_delay_seconds",
			Help:    "Distribution of retry delays in seconds",
			Buckets: prometheus.ExponentialBuckets(0.1, 2, 10), // 0.1s, 0.2s, 0.4s, ... 51.2s
		},
		[]string{"tool"},
	)

	// Register metrics with default registry
	prometheus.MustRegister(
		globalMetrics.retriesTotal,
		globalMetrics.retryAttempts,
		globalMetrics.retryDelaySeconds,
	)
}

// getMetrics returns the metrics instance, initializing it if necessary
func getMetrics() *RetryMetrics {
	metricsOnce.Do(initMetrics)
	return globalMetrics
}

// RecordRetryAttempt records a retry attempt
func RecordRetryAttempt(tool string, attempt uint) {
	metrics := getMetrics()
	metrics.mu.Lock()
	defer metrics.mu.Unlock()

	metrics.retryAttempts.WithLabelValues(tool).Observe(float64(attempt))
}

// RecordRetryDelay records the delay before a retry
func RecordRetryDelay(tool string, delay time.Duration) {
	metrics := getMetrics()
	metrics.mu.Lock()
	defer metrics.mu.Unlock()

	metrics.retryDelaySeconds.WithLabelValues(tool).Observe(delay.Seconds())
}

// RecordRetrySuccess records a successful retry
func RecordRetrySuccess(tool string) {
	metrics := getMetrics()
	metrics.mu.Lock()
	defer metrics.mu.Unlock()

	metrics.retriesTotal.WithLabelValues(tool, "success").Inc()
}

// RecordRetryFailed records a failed retry
func RecordRetryFailed(tool string) {
	metrics := getMetrics()
	metrics.mu.Lock()
	defer metrics.mu.Unlock()

	metrics.retriesTotal.WithLabelValues(tool, "failed").Inc()
}

// RecordRetryExhausted records that all retry attempts were exhausted
func RecordRetryExhausted(tool string) {
	metrics := getMetrics()
	metrics.mu.Lock()
	defer metrics.mu.Unlock()

	metrics.retriesTotal.WithLabelValues(tool, "exhausted").Inc()
}
