package metrics

import (
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds all application metrics
type Metrics struct {
	// Request metrics
	requestsTotal   *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
	requestErrors   *prometheus.CounterVec
	requestActive   *prometheus.GaugeVec

	// Tool-specific metrics
	toolCallsTotal *prometheus.CounterVec
	toolDuration   *prometheus.HistogramVec
	toolErrors     *prometheus.CounterVec

	// Validation metrics
	validationAttempts *prometheus.CounterVec
	validationFailures *prometheus.CounterVec

	// Error metrics
	errorsTotal *prometheus.CounterVec

	// System metrics
	uptime    prometheus.Gauge
	startTime time.Time

	mu sync.Mutex
}

// New creates a new metrics collector
func New() *Metrics {
	return newWithRegistry(prometheus.DefaultRegisterer)
}

// NewWithRegistry creates a new metrics collector with a custom registry (for testing)
func NewWithRegistry(reg prometheus.Registerer) *Metrics {
	return newWithRegistry(reg)
}

// newWithRegistry creates metrics with a custom registry for testing
func newWithRegistry(reg prometheus.Registerer) *Metrics {
	m := &Metrics{
		startTime: time.Now(),
	}

	// Initialize request metrics
	m.requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcp_requests_total",
			Help: "Total number of requests",
		},
		[]string{"method", "tool", "status"},
	)

	m.requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mcp_request_duration_seconds",
			Help:    "Request duration in seconds",
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"method", "tool"},
	)

	m.requestErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcp_request_errors_total",
			Help: "Total number of request errors",
		},
		[]string{"method", "tool", "error_type"},
	)

	m.requestActive = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mcp_requests_active",
			Help: "Number of active requests",
		},
		[]string{"tool"},
	)

	// Initialize tool-specific metrics
	m.toolCallsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcp_tool_calls_total",
			Help: "Total number of tool calls",
		},
		[]string{"tool", "status"},
	)

	m.toolDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mcp_tool_duration_seconds",
			Help:    "Tool execution duration in seconds",
			Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1, 5, 10, 30, 60},
		},
		[]string{"tool"},
	)

	m.toolErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcp_tool_errors_total",
			Help: "Total number of tool errors",
		},
		[]string{"tool", "error_type"},
	)

	// Initialize validation metrics
	m.validationAttempts = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcp_validation_attempts_total",
			Help: "Total number of validation attempts",
		},
		[]string{"rule", "tool"},
	)

	m.validationFailures = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcp_validation_failures_total",
			Help: "Total number of validation failures",
		},
		[]string{"rule", "tool"},
	)

	// Initialize error metrics
	m.errorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcp_errors_total",
			Help: "Total number of errors by category and code",
		},
		[]string{"category", "code", "tool"},
	)

	// Initialize system metrics
	m.uptime = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "mcp_uptime_seconds",
			Help: "Server uptime in seconds",
		},
	)

	// Register metrics with provided registry
	reg.MustRegister(
		m.requestsTotal,
		m.requestDuration,
		m.requestErrors,
		m.requestActive,
		m.toolCallsTotal,
		m.toolDuration,
		m.toolErrors,
		m.validationAttempts,
		m.validationFailures,
		m.errorsTotal,
		m.uptime,
	)

	// Start uptime updater
	go m.updateUptime()

	return m
}

// IncrementActiveRequest increments the active request counter
func (m *Metrics) IncrementActiveRequest(tool string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.requestActive.WithLabelValues(tool).Inc()
}

// DecrementActiveRequest decrements the active request counter
func (m *Metrics) DecrementActiveRequest(tool string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.requestActive.WithLabelValues(tool).Dec()
}

// updateUptime periodically updates the uptime metric
func (m *Metrics) updateUptime() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		m.uptime.Set(time.Since(m.startTime).Seconds())
	}
}

// RecordRequest records a request
func (m *Metrics) RecordRequest(method, tool, status string, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.requestsTotal.WithLabelValues(method, tool, status).Inc()
	m.requestDuration.WithLabelValues(method, tool).Observe(duration.Seconds())
}

// RecordRequestError records a request error
func (m *Metrics) RecordRequestError(method, tool, errorType string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.requestErrors.WithLabelValues(method, tool, errorType).Inc()
}

// RecordToolCall records a tool invocation
func (m *Metrics) RecordToolCall(tool, status string, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.toolCallsTotal.WithLabelValues(tool, status).Inc()
	m.toolDuration.WithLabelValues(tool).Observe(duration.Seconds())
}

// RecordToolError records a tool error
func (m *Metrics) RecordToolError(tool, errorType string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.toolErrors.WithLabelValues(tool, errorType).Inc()
}

// Handler returns an HTTP handler for serving metrics
func (m *Metrics) Handler() http.Handler {
	return promhttp.Handler()
}

// RecordValidationAttempt records a validation attempt
func (m *Metrics) RecordValidationAttempt(rule, tool string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.validationAttempts.WithLabelValues(rule, tool).Inc()
}

// RecordValidationFailure records a validation failure
func (m *Metrics) RecordValidationFailure(rule, tool string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.validationFailures.WithLabelValues(rule, tool).Inc()
}

// RecordError records an error by category and code
func (m *Metrics) RecordError(category, code, tool string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if category == "" {
		category = "unknown"
	}
	if code == "" {
		code = "unknown"
	}
	if tool == "" {
		tool = "none"
	}

	m.errorsTotal.WithLabelValues(category, code, tool).Inc()
}

// Stop stops the metrics collector
func (m *Metrics) Stop() {
	// Cleanup if needed
}
