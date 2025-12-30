package circuitbreaker

import (
	"fmt"
	"sync"
	"time"
)

// State represents the current state of the circuit breaker
type State string

const (
	// StateClosed is the closed state (normal operation)
	StateClosed State = "closed"
	// StateOpen is the open state (rejecting requests)
	StateOpen State = "open"
	// StateHalfOpen is the half-open state (testing recovery)
	StateHalfOpen State = "half-open"
)

// CircuitBreaker implements the circuit breaker pattern for fault tolerance
type CircuitBreaker struct {
	// name is the identifier for the circuit breaker
	name string
	// maxFailures is the threshold for opening the circuit
	maxFailures int
	// timeout is the duration before attempting reset
	timeout time.Duration
	// maxHalfOpenRequests is the number of requests allowed in half-open state
	maxHalfOpenRequests int

	// state is the current state
	state State
	// failures is the current failure count
	failures int
	// lastFailureTime is the time of the last failure
	lastFailureTime time.Time
	// halfOpenRequests is the count of requests in half-open state
	halfOpenRequests int

	// mutex provides thread safety
	mutex sync.RWMutex
}

// NewCircuitBreaker creates a new circuit breaker with the given configuration
func NewCircuitBreaker(name string, config *Config) *CircuitBreaker {
	// Validate config
	if config.Name == "" {
		config.Name = name
	}
	if err := config.Validate(); err != nil {
		panic(fmt.Sprintf("invalid circuit breaker config: %v", err))
	}

	cb := &CircuitBreaker{
		name:                config.Name,
		maxFailures:         config.MaxFailures,
		timeout:             config.Timeout,
		maxHalfOpenRequests: config.MaxHalfOpenRequests,
		state:               StateClosed,
		failures:            0,
		halfOpenRequests:    0,
	}

	// Record initial state
	recordState(cb.name, cb.state)

	return cb
}

// Call executes the given function with circuit breaker protection
func (cb *CircuitBreaker) Call(fn func() error) error {
	// Transition to half-open if timeout has elapsed in open state
	cb.mutex.Lock()
	shouldTransition := cb.state == StateOpen && cb.canAttemptReset()
	if shouldTransition {
		cb.transitionTo(StateHalfOpen)
	}
	cb.mutex.Unlock()

	// Check if we should allow the request
	if !cb.allowRequest() {
		recordRequestRejected(cb.name)
		return NewCircuitBreakerError(cb.name, "request rejected", ErrCircuitBreakerOpen)
	}

	recordRequestAllowed(cb.name)

	// Execute the function
	err := fn()

	// Record the result
	if err != nil {
		cb.RecordFailure(err)
		return err
	}

	cb.RecordSuccess()
	return nil
}

// allowRequest determines if a request should be allowed based on current state
func (cb *CircuitBreaker) allowRequest() bool {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		// Check if timeout has elapsed
		return cb.canAttemptReset()
	case StateHalfOpen:
		// Allow limited requests in half-open state
		return cb.halfOpenRequests < cb.maxHalfOpenRequests
	default:
		return false
	}
}

// canAttemptReset checks if the timeout has elapsed since the last failure
func (cb *CircuitBreaker) canAttemptReset() bool {
	return time.Since(cb.lastFailureTime) >= cb.timeout
}

// RecordSuccess records a successful call
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	switch cb.state {
	case StateClosed:
		// Reset failure count on success
		cb.failures = 0
	case StateHalfOpen:
		// Check if we've had enough successful requests to close the circuit
		cb.halfOpenRequests++
		if cb.halfOpenRequests >= cb.maxHalfOpenRequests {
			cb.transitionTo(StateClosed)
		}
	case StateOpen:
		// Should not happen, but reset to closed if it does
		cb.transitionTo(StateClosed)
	}
}

// RecordFailure records a failed call
func (cb *CircuitBreaker) RecordFailure(err error) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.lastFailureTime = time.Now()

	switch cb.state {
	case StateClosed:
		cb.failures++
		// Check if we should open the circuit
		if cb.failures >= cb.maxFailures {
			cb.transitionTo(StateOpen)
		}
	case StateHalfOpen:
		// Any failure in half-open state opens the circuit again
		cb.transitionTo(StateOpen)
	case StateOpen:
		// Update failure time
		cb.lastFailureTime = time.Now()
	}
}

// State returns the current state of the circuit breaker
func (cb *CircuitBreaker) State() State {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state
}

// IsClosed returns true if the circuit breaker is in closed state
func (cb *CircuitBreaker) IsClosed() bool {
	return cb.State() == StateClosed
}

// IsOpen returns true if the circuit breaker is in open state
func (cb *CircuitBreaker) IsOpen() bool {
	return cb.State() == StateOpen
}

// IsHalfOpen returns true if the circuit breaker is in half-open state
func (cb *CircuitBreaker) IsHalfOpen() bool {
	return cb.State() == StateHalfOpen
}

// Failures returns the current failure count
func (cb *CircuitBreaker) Failures() int {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.failures
}

// Name returns the circuit breaker name
func (cb *CircuitBreaker) Name() string {
	return cb.name
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.transitionTo(StateClosed)
	cb.failures = 0
	cb.lastFailureTime = time.Time{}
	cb.halfOpenRequests = 0
}

// transitionTo transitions the circuit breaker to a new state
func (cb *CircuitBreaker) transitionTo(newState State) {
	if cb.state == newState {
		return
	}

	oldState := cb.state
	cb.state = newState

	// Reset counters based on state transition
	switch newState {
	case StateClosed:
		cb.failures = 0
		cb.halfOpenRequests = 0
	case StateHalfOpen:
		cb.halfOpenRequests = 0
	case StateOpen:
		// Keep lastFailureTime
	}

	// Record the transition
	recordTransition(cb.name, oldState, newState)
	recordState(cb.name, newState)
}

// String returns a string representation of the circuit breaker
func (cb *CircuitBreaker) String() string {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	return fmt.Sprintf("CircuitBreaker{name=%s, state=%s, failures=%d, lastFailure=%v}",
		cb.name, cb.state, cb.failures, cb.lastFailureTime.Format(time.RFC3339))
}
