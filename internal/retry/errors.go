package retry

import (
	"fmt"
	"strings"
	"time"
)

// RetryError wraps an error with retry information
type RetryError struct {
	OriginalError error
	Attempts      uint
	LastDelay     time.Duration
	TotalDelay    time.Duration
}

// Error returns a formatted error message
func (e *RetryError) Error() string {
	if e.OriginalError != nil {
		return fmt.Sprintf("retry failed after %d attempts (total delay: %v): %v",
			e.Attempts, e.TotalDelay, e.OriginalError)
	}
	return fmt.Sprintf("retry failed after %d attempts (total delay: %v)",
		e.Attempts, e.TotalDelay)
}

// Unwrap returns the original error
func (e *RetryError) Unwrap() error {
	return e.OriginalError
}

// Predefined retry errors
var (
	// ErrMaxAttemptsReached is returned when all retry attempts are exhausted
	ErrMaxAttemptsReached = fmt.Errorf("max retry attempts reached")

	// ErrContextCancelled is returned when the context is cancelled during retry
	ErrContextCancelled = fmt.Errorf("context cancelled during retry")
)

// IsRetryError checks if an error is a RetryError
func IsRetryError(err error) bool {
	_, ok := err.(*RetryError)
	return ok
}

// IsMaxAttemptsError checks if an error indicates max attempts were reached
func IsMaxAttemptsError(err error) bool {
	if retryErr, ok := err.(*RetryError); ok {
		return retryErr.Attempts > 0
	}
	return false
}

// IsContextCancelledError checks if an error indicates context cancellation
func IsContextCancelledError(err error) bool {
	errStr := fmt.Sprintf("%v", err)
	// Check for direct match or wrapped error format (e.g., "context cancelled during retry: context canceled")
	return errStr == ErrContextCancelled.Error() ||
		strings.Contains(errStr, ErrContextCancelled.Error())
}
