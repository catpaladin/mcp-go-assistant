package retry

import (
	"context"
	"fmt"
	"time"

	"mcp-go-assistant/internal/logging"
)

// RetryWrapper wraps retry operations with metrics and logging
type RetryWrapper struct {
	retryer Retryer
	logger  *Logger
	tool    string
}

// NewRetryWrapper creates a new retry wrapper for a tool
func NewRetryWrapper(tool string, retryer Retryer, logger *logging.Logger) *RetryWrapper {
	return &RetryWrapper{
		retryer: retryer,
		logger:  NewLogger(logger),
		tool:    tool,
	}
}

// Do executes a function with retry logic, metrics, and logging
func (w *RetryWrapper) Do(ctx context.Context, fn RetryableFunc) error {
	startTime := time.Now()
	var lastAttempt uint

	// Configure onRetry callback to log and record metrics
	if r, ok := w.retryer.(*Retry); ok {
		r.WithOnRetry(func(attempt uint, err error, delay time.Duration) {
			lastAttempt = attempt

			// Log retry attempt
			w.logger.LogRetryAttempt(w.tool, attempt, err, delay)

			// Record metrics
			RecordRetryAttempt(w.tool, attempt)
			RecordRetryDelay(w.tool, delay)
		})
	}

	// Execute with retry
	err := w.retryer.Do(ctx, fn)
	totalDuration := time.Since(startTime)

	if err == nil {
		// Success
		if lastAttempt > 0 {
			w.logger.LogRetrySuccess(w.tool, lastAttempt+1, totalDuration)
			RecordRetrySuccess(w.tool)
		}
		return nil
	}

	// Check error type
	if retryErr, ok := err.(*RetryError); ok {
		// All attempts exhausted
		w.logger.LogRetryExhausted(w.tool, retryErr.Attempts, retryErr.TotalDelay, retryErr.OriginalError)
		RecordRetryExhausted(w.tool)
		return err
	}

	if IsContextCancelledError(err) {
		// Context cancelled
		w.logger.LogRetryCancelled(w.tool, lastAttempt+1)
		return err
	}

	// Other error
	RecordRetryFailed(w.tool)
	return err
}

// DoWithData executes a function with retry logic, metrics, and logging, returning data
func (w *RetryWrapper) DoWithData(ctx context.Context, fn RetryableFuncWithData) (interface{}, error) {
	startTime := time.Now()
	var lastAttempt uint

	// Configure onRetry callback to log and record metrics
	if r, ok := w.retryer.(*Retry); ok {
		r.WithOnRetry(func(attempt uint, err error, delay time.Duration) {
			lastAttempt = attempt

			// Log retry attempt
			w.logger.LogRetryAttempt(w.tool, attempt, err, delay)

			// Record metrics
			RecordRetryAttempt(w.tool, attempt)
			RecordRetryDelay(w.tool, delay)
		})
	}

	// Execute with retry
	result, err := w.retryer.DoWithData(ctx, fn)
	totalDuration := time.Since(startTime)

	if err == nil {
		// Success
		if lastAttempt > 0 {
			w.logger.LogRetrySuccess(w.tool, lastAttempt+1, totalDuration)
			RecordRetrySuccess(w.tool)
		}
		return result, nil
	}

	// Check error type
	if retryErr, ok := err.(*RetryError); ok {
		// All attempts exhausted
		w.logger.LogRetryExhausted(w.tool, retryErr.Attempts, retryErr.TotalDelay, retryErr.OriginalError)
		RecordRetryExhausted(w.tool)
		return result, err
	}

	if IsContextCancelledError(err) {
		// Context cancelled
		w.logger.LogRetryCancelled(w.tool, lastAttempt+1)
		return result, err
	}

	// Other error
	RecordRetryFailed(w.tool)
	return result, err
}

// GetRetryer returns underlying retryer
func (w *RetryWrapper) GetRetryer() Retryer {
	return w.retryer
}

// GetTool returns tool name
func (w *RetryWrapper) GetTool() string {
	return w.tool
}

// SetRetryIf sets the retry condition function
func (w *RetryWrapper) SetRetryIf(fn RetryIfFunc) {
	if r, ok := w.retryer.(*Retry); ok {
		r.WithRetryIf(fn)
		w.retryer = r
	}
}

// RetryableErrors returns a default retry condition for retryable errors
// This function can be customized based on tool-specific error patterns
func RetryableErrors(tool string) RetryIfFunc {
	return func(err error) bool {
		// Don't retry on context cancellation
		if IsContextCancelledError(err) {
			return false
		}

		// Tool-specific error patterns
		errStr := fmt.Sprintf("%v", err)

		// Network errors, timeouts, and temporary errors are retryable
		retryablePatterns := map[string]bool{
			"timeout":                          true,
			"deadline exceeded":                true,
			"temporary":                        true,
			"connection":                       true,
			"network":                          true,
			"temporary failure":                true,
			"try again":                        true,
			"resource temporarily unavailable": true,
			"connection reset by peer":         true,
			"broken pipe":                      true,
		}

		for pattern, retryable := range retryablePatterns {
			if retryable && contains(errStr, pattern) {
				return true
			}
		}

		// GoDoc specific: retry on go doc command errors
		if tool == "go-doc" {
			goDocRetryablePatterns := []string{
				"command not found",
				"exit status",
				"no such file",
			}
			for _, pattern := range goDocRetryablePatterns {
				if contains(errStr, pattern) {
					return true
				}
			}
		}

		// Default: retry on unknown errors
		return true
	}
}

// contains checks if a substring is in the text (case-insensitive)
func contains(text, substr string) bool {
	if substr == "" {
		return false
	}
	// Simple case-insensitive substring check
	textLen := len(text)
	subLen := len(substr)
	if subLen > textLen {
		return false
	}
	for i := 0; i <= textLen-subLen; i++ {
		match := true
		for j := 0; j < subLen; j++ {
			textChar := toLower(text[i+j])
			subChar := toLower(substr[j])
			if textChar != subChar {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// toLower converts a byte to lowercase
func toLower(b byte) byte {
	if b >= 'A' && b <= 'Z' {
		return b + 32
	}
	return b
}
