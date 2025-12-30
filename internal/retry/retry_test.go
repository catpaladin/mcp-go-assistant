package retry

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestRetryError tests the RetryError implementation
func TestRetryError(t *testing.T) {
	originalErr := errors.New("test error")
	retryErr := &RetryError{
		OriginalError: originalErr,
		Attempts:      3,
		LastDelay:     2 * time.Second,
		TotalDelay:    6 * time.Second,
	}

	// Test Error method
	errStr := retryErr.Error()
	expected := "retry failed after 3 attempts (total delay: 6s): test error"
	if errStr != expected {
		t.Errorf("Error() = %v, want %v", errStr, expected)
	}

	// Test Unwrap method
	if retryErr.Unwrap() != originalErr {
		t.Errorf("Unwrap() = %v, want %v", retryErr.Unwrap(), originalErr)
	}
}

// TestRetryErrorNoOriginalError tests RetryError without original error
func TestRetryErrorNoOriginalError(t *testing.T) {
	retryErr := &RetryError{
		Attempts:   3,
		LastDelay:  2 * time.Second,
		TotalDelay: 6 * time.Second,
	}

	errStr := retryErr.Error()
	expected := "retry failed after 3 attempts (total delay: 6s)"
	if errStr != expected {
		t.Errorf("Error() = %v, want %v", errStr, expected)
	}

	if retryErr.Unwrap() != nil {
		t.Errorf("Unwrap() = %v, want nil", retryErr.Unwrap())
	}
}

// TestIsRetryError tests the IsRetryError helper
func TestIsRetryError(t *testing.T) {
	retryErr := &RetryError{Attempts: 1}
	normalErr := errors.New("normal error")

	if !IsRetryError(retryErr) {
		t.Error("IsRetryError(retryErr) = false, want true")
	}

	if IsRetryError(normalErr) {
		t.Error("IsRetryError(normalErr) = true, want false")
	}
}

// TestIsMaxAttemptsError tests the IsMaxAttemptsError helper
func TestIsMaxAttemptsError(t *testing.T) {
	retryErr := &RetryError{Attempts: 3}
	normalErr := errors.New("normal error")

	if !IsMaxAttemptsError(retryErr) {
		t.Error("IsMaxAttemptsError(retryErr) = false, want true")
	}

	if IsMaxAttemptsError(normalErr) {
		t.Error("IsMaxAttemptsError(normalErr) = true, want false")
	}
}

// TestIsContextCancelledError tests the IsContextCancelledError helper
func TestIsContextCancelledError(t *testing.T) {
	normalErr := errors.New("normal error")

	if !IsContextCancelledError(fmt.Errorf("%w", ErrContextCancelled)) {
		t.Error("IsContextCancelledError(wrapped) = false, want true")
	}

	if IsContextCancelledError(normalErr) {
		t.Error("IsContextCancelledError(normalErr) = true, want false")
	}
}

// TestConfigValidation tests configuration validation
func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:   "valid config",
			config: DefaultConfig(),
		},
		{
			name: "max attempts too low",
			config: &Config{
				MaxAttempts: 0,
			},
			wantErr: true,
		},
		{
			name: "negative initial delay",
			config: &Config{
				MaxAttempts:  3,
				InitialDelay: -1 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "negative max delay",
			config: &Config{
				MaxAttempts: 3,
				MaxDelay:    -1 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "initial delay greater than max delay",
			config: &Config{
				MaxAttempts:  3,
				InitialDelay: 10 * time.Second,
				MaxDelay:     5 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "negative multiplier",
			config: &Config{
				MaxAttempts: 3,
				Multiplier:  -1.0,
			},
			wantErr: true,
		},
		{
			name: "zero multiplier",
			config: &Config{
				MaxAttempts: 3,
				Multiplier:  0,
			},
			wantErr: true,
		},
		{
			name: "invalid strategy",
			config: &Config{
				MaxAttempts: 3,
				Strategy:    "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestConfigClone tests configuration cloning
func TestConfigClone(t *testing.T) {
	original := DefaultConfig()
	clone := original.Clone()

	if clone.MaxAttempts != original.MaxAttempts {
		t.Errorf("Clone() MaxAttempts = %v, want %v", clone.MaxAttempts, original.MaxAttempts)
	}
	if clone.InitialDelay != original.InitialDelay {
		t.Errorf("Clone() InitialDelay = %v, want %v", clone.InitialDelay, original.InitialDelay)
	}

	// Modify clone and verify original is unchanged
	clone.MaxAttempts = 100
	if original.MaxAttempts == 100 {
		t.Error("Modifying clone affected original")
	}
}

// TestExponentialBackoff tests exponential backoff strategy
func TestExponentialBackoff(t *testing.T) {
	tests := []struct {
		name      string
		attempt   uint
		wantDelay time.Duration
	}{
		{
			name:      "attempt 0",
			attempt:   0,
			wantDelay: 1 * time.Second,
		},
		{
			name:      "attempt 1",
			attempt:   1,
			wantDelay: 2 * time.Second,
		},
		{
			name:      "attempt 2",
			attempt:   2,
			wantDelay: 4 * time.Second,
		},
		{
			name:      "attempt 3",
			attempt:   3,
			wantDelay: 8 * time.Second,
		},
	}

	strategy := NewExponentialBackoff(1*time.Second, 30*time.Second, 2.0, false)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delay := strategy.NextDelay(tt.attempt)
			if delay != tt.wantDelay {
				t.Errorf("NextDelay() = %v, want %v", delay, tt.wantDelay)
			}
		})
	}
}

// TestExponentialBackoffWithMaxDelay tests exponential backoff with max delay
func TestExponentialBackoffWithMaxDelay(t *testing.T) {
	strategy := NewExponentialBackoff(1*time.Second, 5*time.Second, 2.0, false)

	// Should cap at maxDelay
	delay := strategy.NextDelay(10)
	if delay != 5*time.Second {
		t.Errorf("NextDelay() with high attempt = %v, want %v", delay, 5*time.Second)
	}
}

// TestExponentialBackoffWithJitter tests exponential backoff with jitter
func TestExponentialBackoffWithJitter(t *testing.T) {
	strategy := NewExponentialBackoff(1*time.Second, 30*time.Second, 2.0, true)

	// Run multiple times and ensure we get different values
	delays := make(map[time.Duration]bool)
	for i := 0; i < 100; i++ {
		delay := strategy.NextDelay(2)
		delays[delay] = true
	}

	// With jitter, we should get different values
	if len(delays) < 10 {
		t.Errorf("With jitter, expected multiple different delay values, got %d unique values", len(delays))
	}
}

// TestLinearBackoff tests linear backoff strategy
func TestLinearBackoff(t *testing.T) {
	tests := []struct {
		name      string
		attempt   uint
		wantDelay time.Duration
	}{
		{
			name:      "attempt 0",
			attempt:   0,
			wantDelay: 0,
		},
		{
			name:      "attempt 1",
			attempt:   1,
			wantDelay: 1 * time.Second,
		},
		{
			name:      "attempt 2",
			attempt:   2,
			wantDelay: 2 * time.Second,
		},
		{
			name:      "attempt 3",
			attempt:   3,
			wantDelay: 3 * time.Second,
		},
	}

	strategy := NewLinearBackoff(1*time.Second, 30*time.Second)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delay := strategy.NextDelay(tt.attempt)
			if delay != tt.wantDelay {
				t.Errorf("NextDelay() = %v, want %v", delay, tt.wantDelay)
			}
		})
	}
}

// TestLinearBackoffWithMaxDelay tests linear backoff with max delay
func TestLinearBackoffWithMaxDelay(t *testing.T) {
	strategy := NewLinearBackoff(1*time.Second, 5*time.Second)

	// Should cap at maxDelay
	delay := strategy.NextDelay(10)
	if delay != 5*time.Second {
		t.Errorf("NextDelay() with high attempt = %v, want %v", delay, 5*time.Second)
	}
}

// TestConstantBackoff tests constant backoff strategy
func TestConstantBackoff(t *testing.T) {
	strategy := NewConstantBackoff(2 * time.Second)

	for i := uint(0); i < 5; i++ {
		delay := strategy.NextDelay(i)
		if delay != 2*time.Second {
			t.Errorf("NextDelay() = %v, want %v", delay, 2*time.Second)
		}
	}
}

// TestNoBackoff tests no-backoff strategy
func TestNoBackoff(t *testing.T) {
	strategy := NewNoBackoff()

	for i := uint(0); i < 5; i++ {
		delay := strategy.NextDelay(i)
		if delay != 0 {
			t.Errorf("NextDelay() = %v, want 0", delay)
		}
	}
}

// TestNewStrategy tests strategy factory
func TestNewStrategy(t *testing.T) {
	tests := []struct {
		name     string
		strategy string
		wantType string
		wantErr  bool
	}{
		{
			name:     "exponential",
			strategy: "exponential",
			wantType: "*retry.ExponentialBackoff",
		},
		{
			name:     "linear",
			strategy: "linear",
			wantType: "*retry.LinearBackoff",
		},
		{
			name:     "constant",
			strategy: "constant",
			wantType: "*retry.ConstantBackoff",
		},
		{
			name:     "none",
			strategy: "none",
			wantType: "*retry.NoBackoff",
		},
		{
			name:     "invalid",
			strategy: "invalid",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy, err := NewStrategy(tt.strategy, 1*time.Second, 30*time.Second, 2.0, false)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewStrategy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if fmt.Sprintf("%T", strategy) != tt.wantType {
					t.Errorf("NewStrategy() type = %T, want %s", strategy, tt.wantType)
				}
			}
		})
	}
}

// TestRetrySuccessOnFirstAttempt tests retry succeeding on first attempt
func TestRetrySuccessOnFirstAttempt(t *testing.T) {
	config := DefaultConfig()
	retryer := NewRetryer(config)

	callCount := 0
	fn := func(attempt uint) error {
		callCount++
		return nil
	}

	ctx := context.Background()
	err := retryer.Do(ctx, fn)

	if err != nil {
		t.Errorf("Do() error = %v, want nil", err)
	}
	if callCount != 1 {
		t.Errorf("Do() called fn %d times, want 1", callCount)
	}
}

// TestRetrySuccessOnSecondAttempt tests retry succeeding on second attempt
func TestRetrySuccessOnSecondAttempt(t *testing.T) {
	config := DefaultConfig()
	config.MaxAttempts = 3
	retryer := NewRetryer(config)

	callCount := 0
	fn := func(attempt uint) error {
		callCount++
		if callCount < 2 {
			return errors.New("test error")
		}
		return nil
	}

	ctx := context.Background()
	err := retryer.Do(ctx, fn)

	if err != nil {
		t.Errorf("Do() error = %v, want nil", err)
	}
	if callCount != 2 {
		t.Errorf("Do() called fn %d times, want 2", callCount)
	}
}

// TestRetryMaxAttempts tests retry exhausting all attempts
func TestRetryMaxAttempts(t *testing.T) {
	config := DefaultConfig()
	config.MaxAttempts = 3
	retryer := NewRetryer(config)

	callCount := 0
	fn := func(attempt uint) error {
		callCount++
		return errors.New("always fails")
	}

	ctx := context.Background()
	err := retryer.Do(ctx, fn)

	if err == nil {
		t.Error("Do() error = nil, want error")
	}

	retryErr, ok := err.(*RetryError)
	if !ok {
		t.Fatalf("Do() error type = %T, want *RetryError", err)
	}

	if retryErr.Attempts != 3 {
		t.Errorf("Do() retry attempts = %d, want 3", retryErr.Attempts)
	}

	if callCount != 3 {
		t.Errorf("Do() called fn %d times, want 3", callCount)
	}
}

// TestRetryContextCancellation tests retry with context cancellation
func TestRetryContextCancellation(t *testing.T) {
	config := DefaultConfig()
	config.MaxAttempts = 10
	retryer := NewRetryer(config)

	ctx, cancel := context.WithCancel(context.Background())

	callCount := 0
	fn := func(attempt uint) error {
		callCount++
		if callCount == 3 {
			cancel()
		}
		return errors.New("always fails")
	}

	err := retryer.Do(ctx, fn)

	if err == nil {
		t.Error("Do() error = nil, want error")
	}

	if !IsContextCancelledError(err) {
		t.Errorf("Do() error = %v, want context cancelled error", err)
	}

	if callCount != 3 {
		t.Errorf("Do() called fn %d times, want 3", callCount)
	}
}

// TestRetryWithRetryIf tests retry with custom retry condition
func TestRetryWithRetryIf(t *testing.T) {
	config := DefaultConfig()
	config.MaxAttempts = 5
	retryer := NewRetryer(config).(*Retry)

	// Only retry on specific error
	retryer.WithRetryIf(func(err error) bool {
		return err.Error() == "retryable error"
	})

	callCount := 0
	fn := func(attempt uint) error {
		callCount++
		if callCount == 1 {
			return errors.New("retryable error")
		}
		return errors.New("non-retryable error")
	}

	ctx := context.Background()
	err := retryer.Do(ctx, fn)

	if err == nil {
		t.Error("Do() error = nil, want error")
	}

	if callCount != 2 {
		t.Errorf("Do() called fn %d times, want 2", callCount)
	}
}

// TestRetryWithOnRetry tests retry with onRetry callback
func TestRetryWithOnRetry(t *testing.T) {
	config := DefaultConfig()
	config.MaxAttempts = 3
	retryer := NewRetryer(config).(*Retry)

	callbackCalls := 0
	retryer.WithOnRetry(func(attempt uint, err error, delay time.Duration) {
		callbackCalls++
		if delay == 0 {
			t.Error("onRetry called with zero delay")
		}
	})

	fn := func(attempt uint) error {
		return errors.New("always fails")
	}

	ctx := context.Background()
	err := retryer.Do(ctx, fn)

	if err == nil {
		t.Error("Do() error = nil, want error")
	}

	// Should be called 2 times (after attempt 0 and attempt 1)
	// Not called after the last attempt (attempt 2)
	if callbackCalls != 2 {
		t.Errorf("onRetry called %d times, want 2", callbackCalls)
	}
}

// TestRetryWithDataSuccess tests retry with data returning successfully
func TestRetryWithDataSuccess(t *testing.T) {
	config := DefaultConfig()
	retryer := NewRetryer(config)

	callCount := 0
	fn := func(attempt uint) (interface{}, error) {
		callCount++
		return "success", nil
	}

	ctx := context.Background()
	result, err := retryer.DoWithData(ctx, fn)

	if err != nil {
		t.Errorf("DoWithData() error = %v, want nil", err)
	}

	if result != "success" {
		t.Errorf("DoWithData() result = %v, want 'success'", result)
	}

	if callCount != 1 {
		t.Errorf("DoWithData() called fn %d times, want 1", callCount)
	}
}

// TestRetryWithDataRetry tests retry with data and retries
func TestRetryWithDataRetry(t *testing.T) {
	config := DefaultConfig()
	config.MaxAttempts = 3
	retryer := NewRetryer(config)

	callCount := 0
	fn := func(attempt uint) (interface{}, error) {
		callCount++
		if callCount < 2 {
			return nil, errors.New("test error")
		}
		return "success", nil
	}

	ctx := context.Background()
	result, err := retryer.DoWithData(ctx, fn)

	if err != nil {
		t.Errorf("DoWithData() error = %v, want nil", err)
	}

	if result != "success" {
		t.Errorf("DoWithData() result = %v, want 'success'", result)
	}

	if callCount != 2 {
		t.Errorf("DoWithData() called fn %d times, want 2", callCount)
	}
}

// TestRetryChainMethods tests retry chain methods
func TestRetryChainMethods(t *testing.T) {
	config := DefaultConfig()
	retryer := NewRetryer(config).(*Retry)

	// Test chain methods
	retryer.WithMaxAttempts(5).
		WithDelay(500 * time.Millisecond).
		WithMaxDelay(10 * time.Second).
		WithMultiplier(3.0).
		WithJitter(false)

	if retryer.maxAttempts != 5 {
		t.Errorf("WithMaxAttempts() = %d, want 5", retryer.maxAttempts)
	}

	if retryer.initialDelay != 500*time.Millisecond {
		t.Errorf("WithDelay() = %v, want 500ms", retryer.initialDelay)
	}

	if retryer.maxDelay != 10*time.Second {
		t.Errorf("WithMaxDelay() = %v, want 10s", retryer.maxDelay)
	}

	if retryer.multiplier != 3.0 {
		t.Errorf("WithMultiplier() = %v, want 3.0", retryer.multiplier)
	}

	if retryer.jitter != false {
		t.Errorf("WithJitter() = %v, want false", retryer.jitter)
	}
}

// TestConcurrentRetries tests concurrent retry operations
func TestConcurrentRetries(t *testing.T) {
	config := DefaultConfig()
	config.MaxAttempts = 3
	retryer := NewRetryer(config)

	var wg sync.WaitGroup
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			fn := func(attempt uint) error {
				if attempt < 2 {
					return errors.New("test error")
				}
				return nil
			}

			ctx := context.Background()
			err := retryer.Do(ctx, fn)
			if err != nil {
				t.Errorf("Goroutine %d: Do() error = %v", id, err)
			}
		}(i)
	}

	wg.Wait()
}

// TestRetryAlreadyCancelledContext tests retry with already cancelled context
func TestRetryAlreadyCancelledContext(t *testing.T) {
	config := DefaultConfig()
	retryer := NewRetryer(config)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	callCount := 0
	fn := func(attempt uint) error {
		callCount++
		return errors.New("test error")
	}

	err := retryer.Do(ctx, fn)

	if err == nil {
		t.Error("Do() error = nil, want error")
	}

	if !IsContextCancelledError(err) {
		t.Errorf("Do() error = %v, want context cancelled error", err)
	}

	if callCount != 0 {
		t.Errorf("Do() called fn %d times, want 0 (should not call with cancelled context)", callCount)
	}
}

// TestNewRetryerNilConfig tests creating retryer with nil config
func TestNewRetryerNilConfig(t *testing.T) {
	retryer := NewRetryer(nil)
	if retryer == nil {
		t.Error("NewRetryer(nil) = nil, want non-nil retryer")
	}
}
