package circuitbreaker

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

// CircuitBreakerMetrics holds metrics for circuit breakers
type CircuitBreakerMetrics struct {
	// circuitBreakerState is a gauge for current state (0=closed, 1=half-open, 2=open)
	circuitBreakerState *prometheus.GaugeVec
	// circuitBreakerTransitionsTotal is a counter for state transitions
	circuitBreakerTransitionsTotal *prometheus.CounterVec
	// circuitBreakerRequestsRejectedTotal is a counter for rejected requests
	circuitBreakerRequestsRejectedTotal *prometheus.CounterVec
	// circuitBreakerRequestsAllowedTotal is a counter for allowed requests
	circuitBreakerRequestsAllowedTotal *prometheus.CounterVec

	mu sync.Mutex
}

var (
	// globalMetrics is the singleton metrics instance
	globalMetrics *CircuitBreakerMetrics
	// metricsOnce ensures metrics are initialized only once
	metricsOnce sync.Once
)

// initMetrics initializes the metrics collector
func initMetrics() {
	globalMetrics = &CircuitBreakerMetrics{}

	// Initialize circuit breaker state gauge
	globalMetrics.circuitBreakerState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "circuit_breaker_state",
			Help: "Current state of the circuit breaker (0=closed, 1=half-open, 2=open)",
		},
		[]string{"name"},
	)

	// Initialize circuit breaker transitions counter
	globalMetrics.circuitBreakerTransitionsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "circuit_breaker_transitions_total",
			Help: "Total number of state transitions",
		},
		[]string{"name", "from_state", "to_state"},
	)

	// Initialize circuit breaker requests rejected counter
	globalMetrics.circuitBreakerRequestsRejectedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "circuit_breaker_requests_rejected_total",
			Help: "Total number of rejected requests",
		},
		[]string{"name"},
	)

	// Initialize circuit breaker requests allowed counter
	globalMetrics.circuitBreakerRequestsAllowedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "circuit_breaker_requests_allowed_total",
			Help: "Total number of allowed requests",
		},
		[]string{"name"},
	)

	// Register metrics with default registry
	prometheus.MustRegister(
		globalMetrics.circuitBreakerState,
		globalMetrics.circuitBreakerTransitionsTotal,
		globalMetrics.circuitBreakerRequestsRejectedTotal,
		globalMetrics.circuitBreakerRequestsAllowedTotal,
	)
}

// getMetrics returns the metrics instance, initializing it if necessary
func getMetrics() *CircuitBreakerMetrics {
	metricsOnce.Do(initMetrics)
	return globalMetrics
}

// recordState records the current state of a circuit breaker
func recordState(name string, state State) {
	metrics := getMetrics()
	metrics.mu.Lock()
	defer metrics.mu.Unlock()

	var stateValue float64
	switch state {
	case StateClosed:
		stateValue = 0
	case StateHalfOpen:
		stateValue = 1
	case StateOpen:
		stateValue = 2
	default:
		stateValue = 0
	}

	metrics.circuitBreakerState.WithLabelValues(name).Set(stateValue)
}

// recordTransition records a state transition
func recordTransition(name string, fromState, toState State) {
	metrics := getMetrics()
	metrics.mu.Lock()
	defer metrics.mu.Unlock()

	metrics.circuitBreakerTransitionsTotal.WithLabelValues(name, string(fromState), string(toState)).Inc()
}

// recordRequestRejected records a rejected request
func recordRequestRejected(name string) {
	metrics := getMetrics()
	metrics.mu.Lock()
	defer metrics.mu.Unlock()

	metrics.circuitBreakerRequestsRejectedTotal.WithLabelValues(name).Inc()
}

// recordRequestAllowed records an allowed request
func recordRequestAllowed(name string) {
	metrics := getMetrics()
	metrics.mu.Lock()
	defer metrics.mu.Unlock()

	metrics.circuitBreakerRequestsAllowedTotal.WithLabelValues(name).Inc()
}
