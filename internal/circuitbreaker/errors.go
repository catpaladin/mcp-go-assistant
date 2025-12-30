package circuitbreaker

import (
	"fmt"

	"mcp-go-assistant/internal/types"
)

// ErrCircuitBreakerOpen is returned when the circuit breaker is open
var ErrCircuitBreakerOpen = fmt.Errorf("circuit breaker is open")

// CircuitBreakerError represents an error with circuit breaker context
type CircuitBreakerError struct {
	// Message is the error message
	Message string
	// Name is the circuit breaker name
	Name string
	// Err is the underlying error
	Err error
}

// Error implements the error interface
func (e *CircuitBreakerError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("circuit breaker '%s': %s: %v", e.Name, e.Message, e.Err)
	}
	return fmt.Sprintf("circuit breaker '%s': %s", e.Name, e.Message)
}

// Unwrap returns the underlying error
func (e *CircuitBreakerError) Unwrap() error {
	return e.Err
}

// NewCircuitBreakerError creates a new circuit breaker error
func NewCircuitBreakerError(name, message string, err error) *CircuitBreakerError {
	return &CircuitBreakerError{
		Name:    name,
		Message: message,
		Err:     err,
	}
}

// ToMCPError converts a CircuitBreakerError to an MCPError
func (e *CircuitBreakerError) ToMCPError() types.MCPError {
	details := make(map[string]interface{})
	if e.Name != "" {
		details["circuit_breaker"] = e.Name
	}

	if e.Err != nil {
		return types.WrapCircuitBreakerError(e.Err, e.Error(), details)
	}

	return types.NewCircuitBreakerError(e.Error(), details)
}

// IsCircuitBreakerError checks if an error is a CircuitBreakerError
func IsCircuitBreakerError(err error) bool {
	_, ok := err.(*CircuitBreakerError)
	return ok
}
