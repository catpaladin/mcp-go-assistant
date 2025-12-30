package ratelimit

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

// Metrics holds rate limiting specific metrics
type Metrics struct {
	// allowedTotal counts allowed requests
	allowedTotal *prometheus.CounterVec
	// rejectedTotal counts rejected requests
	rejectedTotal *prometheus.CounterVec
	// currentCount tracks current request count per key
	currentCount *prometheus.GaugeVec
	// limitExceeded tracks rate limit exceeded events
	limitExceeded *prometheus.CounterVec

	mu sync.Mutex
}

// NewMetrics creates new rate limiting metrics
func NewMetrics() *Metrics {
	m := &Metrics{}

	// Initialize metrics
	m.allowedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcp_ratelimit_allowed_total",
			Help: "Total number of requests allowed by rate limiter",
		},
		[]string{"tool", "mode"},
	)

	m.rejectedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcp_ratelimit_rejected_total",
			Help: "Total number of requests rejected by rate limiter",
		},
		[]string{"tool", "mode"},
	)

	m.currentCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mcp_ratelimit_current",
			Help: "Current number of requests in the rate limit window",
		},
		[]string{"tool", "mode"},
	)

	m.limitExceeded = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcp_ratelimit_limit_exceeded_total",
			Help: "Total number of times rate limit was exceeded",
		},
		[]string{"tool", "mode"},
	)

	// Register metrics with default registry
	prometheus.MustRegister(
		m.allowedTotal,
		m.rejectedTotal,
		m.currentCount,
		m.limitExceeded,
	)

	return m
}

// RecordAllowed records an allowed request
func (m *Metrics) RecordAllowed(toolName, mode string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.allowedTotal.WithLabelValues(toolName, mode).Inc()
}

// RecordRejected records a rejected request
func (m *Metrics) RecordRejected(toolName, mode string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.rejectedTotal.WithLabelValues(toolName, mode).Inc()
	m.limitExceeded.WithLabelValues(toolName, mode).Inc()
}

// SetCurrent sets the current count for a key
func (m *Metrics) SetCurrent(toolName, mode string, count int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.currentCount.WithLabelValues(toolName, mode).Set(float64(count))
}

// Stop cleans up metrics resources
func (m *Metrics) Stop() {
	// Metrics are automatically unregistered when the application stops
}

// Global rate limit metrics instance
var globalMetrics *Metrics

// InitMetrics initializes the global rate limit metrics
func InitMetrics() {
	if globalMetrics == nil {
		globalMetrics = NewMetrics()
	}
}

// GetMetrics returns the global rate limit metrics instance
func GetMetrics() *Metrics {
	return globalMetrics
}

// RecordAllowed records an allowed request using the global metrics
func RecordAllowed(toolName, mode string) {
	if globalMetrics != nil {
		globalMetrics.RecordAllowed(toolName, mode)
	}
}

// RecordRejected records a rejected request using the global metrics
func RecordRejected(toolName, mode string) {
	if globalMetrics != nil {
		globalMetrics.RecordRejected(toolName, mode)
	}
}

// SetCurrent sets the current count using the global metrics
func SetCurrent(toolName, mode string, count int) {
	if globalMetrics != nil {
		globalMetrics.SetCurrent(toolName, mode, count)
	}
}
