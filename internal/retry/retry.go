package retry

import (
	"context"
	"fmt"
	"time"
)

// RetryableFunc is a function that can be retried
// attempt is the current attempt number (starting at 0 for the first call)
type RetryableFunc func(attempt uint) error

// RetryableFuncWithData is a function that can be retried and returns data
// attempt is the current attempt number (starting at 0 for the first call)
type RetryableFuncWithData func(attempt uint) (interface{}, error)

// OnRetryFunc is called after each failed attempt
// attempt is the attempt number that just failed
// err is the error from the failed attempt
// delay is the delay before the next attempt
type OnRetryFunc func(attempt uint, err error, delay time.Duration)

// RetryIfFunc determines whether an error should be retried
// err is the error from the failed attempt
// returns true if the error should be retried, false otherwise
type RetryIfFunc func(err error) bool

// Retryer defines the interface for retry operations
type Retryer interface {
	// Do executes the function with retry logic
	Do(ctx context.Context, fn RetryableFunc) error

	// DoWithData executes the function with retry logic and returns data
	DoWithData(ctx context.Context, fn RetryableFuncWithData) (interface{}, error)
}

// Retry implements the Retryer interface
type Retry struct {
	maxAttempts  uint
	initialDelay time.Duration
	maxDelay     time.Duration
	multiplier   float64
	jitter       bool
	onRetry      OnRetryFunc
	retryIf      RetryIfFunc
	strategy     BackoffStrategy
}

// NewRetryer creates a new Retryer with the given configuration
func NewRetryer(config *Config) Retryer {
	// Validate config
	if config == nil {
		config = DefaultConfig()
	}
	if err := config.Validate(); err != nil {
		panic(fmt.Sprintf("invalid retry config: %v", err))
	}

	// Create backoff strategy
	strategy, err := NewStrategy(
		config.Strategy,
		config.InitialDelay,
		config.MaxDelay,
		config.Multiplier,
		config.Jitter,
	)
	if err != nil {
		panic(fmt.Sprintf("failed to create backoff strategy: %v", err))
	}

	return &Retry{
		maxAttempts:  config.MaxAttempts,
		initialDelay: config.InitialDelay,
		maxDelay:     config.MaxDelay,
		multiplier:   config.Multiplier,
		jitter:       config.Jitter,
		onRetry:      nil,
		retryIf:      nil,
		strategy:     strategy,
	}
}

// Do executes the function with retry logic
func (r *Retry) Do(ctx context.Context, fn RetryableFunc) error {
	var lastErr error
	var totalDelay time.Duration
	var lastDelay time.Duration

	for attempt := uint(0); attempt < r.maxAttempts; attempt++ {
		// Check if context is cancelled before executing
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("%w: %v", ErrContextCancelled, err)
		}

		// Execute the function
		lastErr = fn(attempt)

		// If no error, return success
		if lastErr == nil {
			return nil
		}

		// Check if we should retry based on the error
		if r.retryIf != nil && !r.retryIf(lastErr) {
			return lastErr
		}

		// If this was the last attempt, return the error
		if attempt == r.maxAttempts-1 {
			break
		}

		// Calculate delay for next attempt
		nextDelay := r.strategy.NextDelay(attempt)
		lastDelay = nextDelay
		totalDelay += nextDelay

		// Call onRetry callback if set
		if r.onRetry != nil {
			r.onRetry(attempt, lastErr, nextDelay)
		}

		// Wait for delay or context cancellation
		if nextDelay > 0 {
			select {
			case <-time.After(nextDelay):
				// Delay elapsed, continue to next attempt
			case <-ctx.Done():
				return fmt.Errorf("%w: %v", ErrContextCancelled, ctx.Err())
			}
		}
	}

	// All attempts exhausted
	return &RetryError{
		OriginalError: lastErr,
		Attempts:      r.maxAttempts,
		LastDelay:     lastDelay,
		TotalDelay:    totalDelay,
	}
}

// DoWithData executes the function with retry logic and returns data
func (r *Retry) DoWithData(ctx context.Context, fn RetryableFuncWithData) (interface{}, error) {
	var result interface{}
	var lastErr error
	var totalDelay time.Duration
	var lastDelay time.Duration

	for attempt := uint(0); attempt < r.maxAttempts; attempt++ {
		// Check if context is cancelled before executing
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrContextCancelled, err)
		}

		// Execute the function
		result, lastErr = fn(attempt)

		// If no error, return success
		if lastErr == nil {
			return result, nil
		}

		// Check if we should retry based on the error
		if r.retryIf != nil && !r.retryIf(lastErr) {
			return nil, lastErr
		}

		// If this was the last attempt, return the error
		if attempt == r.maxAttempts-1 {
			break
		}

		// Calculate delay for next attempt
		nextDelay := r.strategy.NextDelay(attempt)
		lastDelay = nextDelay
		totalDelay += nextDelay

		// Call onRetry callback if set
		if r.onRetry != nil {
			r.onRetry(attempt, lastErr, nextDelay)
		}

		// Wait for delay or context cancellation
		if nextDelay > 0 {
			select {
			case <-time.After(nextDelay):
				// Delay elapsed, continue to next attempt
			case <-ctx.Done():
				return nil, fmt.Errorf("%w: %v", ErrContextCancelled, ctx.Err())
			}
		}
	}

	// All attempts exhausted
	return result, &RetryError{
		OriginalError: lastErr,
		Attempts:      r.maxAttempts,
		LastDelay:     lastDelay,
		TotalDelay:    totalDelay,
	}
}

// WithMaxAttempts sets the maximum number of retry attempts
func (r *Retry) WithMaxAttempts(attempts uint) *Retry {
	r.maxAttempts = attempts
	return r
}

// WithDelay sets the initial delay
func (r *Retry) WithDelay(delay time.Duration) *Retry {
	r.initialDelay = delay
	return r
}

// WithMaxDelay sets the maximum delay
func (r *Retry) WithMaxDelay(delay time.Duration) *Retry {
	r.maxDelay = delay
	return r
}

// WithMultiplier sets the multiplier for exponential backoff
func (r *Retry) WithMultiplier(multiplier float64) *Retry {
	r.multiplier = multiplier
	return r
}

// WithJitter enables or disables jitter
func (r *Retry) WithJitter(jitter bool) *Retry {
	r.jitter = jitter
	return r
}

// WithOnRetry sets the callback function to call on each retry
func (r *Retry) WithOnRetry(fn OnRetryFunc) *Retry {
	r.onRetry = fn
	return r
}

// WithRetryIf sets the function to determine if an error should be retried
func (r *Retry) WithRetryIf(fn RetryIfFunc) *Retry {
	r.retryIf = fn
	return r
}

// WithStrategy sets the backoff strategy
func (r *Retry) WithStrategy(strategy BackoffStrategy) *Retry {
	r.strategy = strategy
	return r
}
