package circuitbreaker

import (
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig("test")

	if config.Name != "test" {
		t.Errorf("expected name 'test', got '%s'", config.Name)
	}

	if config.MaxFailures != 5 {
		t.Errorf("expected MaxFailures 5, got %d", config.MaxFailures)
	}

	if config.Timeout != 30*time.Second {
		t.Errorf("expected Timeout 30s, got %v", config.Timeout)
	}

	if config.MaxHalfOpenRequests != 3 {
		t.Errorf("expected MaxHalfOpenRequests 3, got %d", config.MaxHalfOpenRequests)
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Name:                "test",
				MaxFailures:         5,
				Timeout:             30 * time.Second,
				MaxHalfOpenRequests: 3,
			},
			wantErr: false,
		},
		{
			name: "empty name",
			config: Config{
				MaxFailures:         5,
				Timeout:             30 * time.Second,
				MaxHalfOpenRequests: 3,
			},
			wantErr: true,
		},
		{
			name: "zero max failures",
			config: Config{
				Name:                "test",
				MaxFailures:         0,
				Timeout:             30 * time.Second,
				MaxHalfOpenRequests: 3,
			},
			wantErr: true,
		},
		{
			name: "zero timeout",
			config: Config{
				Name:                "test",
				MaxFailures:         5,
				Timeout:             0,
				MaxHalfOpenRequests: 3,
			},
			wantErr: true,
		},
		{
			name: "zero max half open requests",
			config: Config{
				Name:                "test",
				MaxFailures:         5,
				Timeout:             30 * time.Second,
				MaxHalfOpenRequests: 0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewCircuitBreaker(t *testing.T) {
	config := DefaultConfig("test")
	cb := NewCircuitBreaker("test", config)

	if cb.Name() != "test" {
		t.Errorf("expected name 'test', got '%s'", cb.Name())
	}

	if !cb.IsClosed() {
		t.Errorf("expected initial state to be closed")
	}

	if cb.Failures() != 0 {
		t.Errorf("expected initial failures to be 0, got %d", cb.Failures())
	}
}

func TestCircuitBreaker_AllowRequest_ClosedState(t *testing.T) {
	cb := NewCircuitBreaker("test", DefaultConfig("test"))

	if !cb.allowRequest() {
		t.Error("expected request to be allowed in closed state")
	}
}

func TestCircuitBreaker_AllowRequest_OpenState(t *testing.T) {
	config := DefaultConfig("test")
	config.Timeout = 100 * time.Millisecond
	cb := NewCircuitBreaker("test", config)

	// Force open by recording failures
	for i := 0; i < config.MaxFailures; i++ {
		cb.RecordFailure(errors.New("test error"))
	}

	if !cb.IsOpen() {
		t.Error("expected circuit to be open after max failures")
	}

	// Request should be denied immediately
	if cb.allowRequest() {
		t.Error("expected request to be denied in open state")
	}

	// Wait for timeout
	time.Sleep(150 * time.Millisecond)

	// Request should now be allowed (transition to half-open)
	if !cb.allowRequest() {
		t.Error("expected request to be allowed after timeout")
	}
}

func TestCircuitBreaker_AllowRequest_HalfOpenState(t *testing.T) {
	config := DefaultConfig("test")
	config.Timeout = 50 * time.Millisecond
	config.MaxHalfOpenRequests = 2
	cb := NewCircuitBreaker("test", config)

	// Force open
	for i := 0; i < config.MaxFailures; i++ {
		cb.RecordFailure(errors.New("test error"))
	}

	// Wait for timeout to transition to half-open
	time.Sleep(75 * time.Millisecond)

	// First request should be allowed
	if !cb.allowRequest() {
		t.Error("expected first request to be allowed in half-open state")
	}

	// Record success to track half-open request count
	cb.RecordSuccess()

	// Second request should still be allowed
	if !cb.allowRequest() {
		t.Error("expected second request to be allowed in half-open state")
	}

	cb.RecordSuccess()

	// Third request should be denied (circuit should be closed now)
	if !cb.allowRequest() {
		t.Error("expected request to be allowed after closing circuit")
	}

	if !cb.IsClosed() {
		t.Error("expected circuit to be closed after successful half-open requests")
	}
}

func TestCircuitBreaker_RecordSuccess(t *testing.T) {
	t.Run("closed state resets failures", func(t *testing.T) {
		cb := NewCircuitBreaker("test", DefaultConfig("test"))

		// Record some failures
		cb.RecordFailure(errors.New("test"))
		cb.RecordFailure(errors.New("test"))

		if cb.Failures() != 2 {
			t.Errorf("expected 2 failures, got %d", cb.Failures())
		}

		// Record success
		cb.RecordSuccess()

		if cb.Failures() != 0 {
			t.Errorf("expected failures to be reset to 0, got %d", cb.Failures())
		}
	})

	t.Run("half-open state transitions to closed after max requests", func(t *testing.T) {
		config := DefaultConfig("test")
		config.Timeout = 50 * time.Millisecond
		config.MaxHalfOpenRequests = 3
		cb := NewCircuitBreaker("test", config)

		// Force open
		for i := 0; i < config.MaxFailures; i++ {
			cb.RecordFailure(errors.New("test"))
		}

		// Wait for timeout
		time.Sleep(75 * time.Millisecond)

		// Record successful requests
		for i := 0; i < config.MaxHalfOpenRequests; i++ {
			cb.RecordSuccess()
		}

		if !cb.IsClosed() {
			t.Error("expected circuit to be closed after successful half-open requests")
		}
	})
}

func TestCircuitBreaker_RecordFailure(t *testing.T) {
	t.Run("closed state opens circuit after max failures", func(t *testing.T) {
		config := DefaultConfig("test")
		config.MaxFailures = 3
		cb := NewCircuitBreaker("test", config)

		// Record failures up to threshold
		for i := 0; i < config.MaxFailures; i++ {
			cb.RecordFailure(errors.New("test"))
		}

		if !cb.IsOpen() {
			t.Error("expected circuit to be open after max failures")
		}
	})

	t.Run("half-open state opens circuit on failure", func(t *testing.T) {
		config := DefaultConfig("test")
		config.Timeout = 50 * time.Millisecond
		config.MaxHalfOpenRequests = 3
		cb := NewCircuitBreaker("test", config)

		// Force open
		for i := 0; i < config.MaxFailures; i++ {
			cb.RecordFailure(errors.New("test"))
		}

		// Wait for timeout to be able to attempt reset
		time.Sleep(75 * time.Millisecond)

		// The circuit should still be open until a request is made
		if !cb.IsOpen() {
			t.Error("expected circuit to still be open before request")
		}

		// Make a request to trigger transition to half-open
		_ = cb.Call(func() error {
			// This request will fail, opening the circuit again
			return errors.New("test")
		})

		// After a failure during the half-open attempt, circuit should be open
		if !cb.IsOpen() {
			t.Error("expected circuit to be open after failure in half-open state")
		}
	})
}

func TestCircuitBreaker_Call(t *testing.T) {
	t.Run("successful call", func(t *testing.T) {
		cb := NewCircuitBreaker("test", DefaultConfig("test"))

		err := cb.Call(func() error {
			return nil
		})

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if cb.Failures() != 0 {
			t.Errorf("expected no failures, got %d", cb.Failures())
		}
	})

	t.Run("failed call", func(t *testing.T) {
		cb := NewCircuitBreaker("test", DefaultConfig("test"))
		testErr := errors.New("test error")

		err := cb.Call(func() error {
			return testErr
		})

		if err != testErr {
			t.Errorf("expected test error, got %v", err)
		}

		if cb.Failures() != 1 {
			t.Errorf("expected 1 failure, got %d", cb.Failures())
		}
	})

	t.Run("open circuit rejects call", func(t *testing.T) {
		config := DefaultConfig("test")
		config.MaxFailures = 2
		cb := NewCircuitBreaker("test", config)

		// Force open
		for i := 0; i < config.MaxFailures; i++ {
			cb.RecordFailure(errors.New("test"))
		}

		if !cb.IsOpen() {
			t.Fatal("expected circuit to be open")
		}

		err := cb.Call(func() error {
			return nil
		})

		var cbErr *CircuitBreakerError
		if !errors.As(err, &cbErr) {
			t.Errorf("expected CircuitBreakerError, got %T", err)
		}

		if cbErr.Err != ErrCircuitBreakerOpen {
			t.Errorf("expected ErrCircuitBreakerOpen, got %v", cbErr.Err)
		}
	})
}

func TestCircuitBreaker_Transitions(t *testing.T) {
	config := DefaultConfig("test")
	config.MaxFailures = 3
	config.Timeout = 100 * time.Millisecond
	config.MaxHalfOpenRequests = 2
	cb := NewCircuitBreaker("test", config)

	// Initial state should be closed
	if !cb.IsClosed() {
		t.Error("expected initial state to be closed")
	}

	// Record failures to open circuit
	for i := 0; i < config.MaxFailures; i++ {
		cb.RecordFailure(errors.New("test"))
	}

	if !cb.IsOpen() {
		t.Error("expected circuit to be open after failures")
	}

	// Wait for timeout
	time.Sleep(150 * time.Millisecond)

	// First successful request transitions from open to half-open to closed
	// (because MaxHalfOpenRequests is 2 and we succeed)
	_ = cb.Call(func() error {
		return nil
	})

	if !cb.IsHalfOpen() {
		t.Error("expected circuit to be half-open after first successful request")
	}

	// Second successful request closes the circuit
	_ = cb.Call(func() error {
		return nil
	})

	if !cb.IsClosed() {
		t.Error("expected circuit to be closed after two successful half-open requests")
	}
}

func TestCircuitBreaker_Reset(t *testing.T) {
	config := DefaultConfig("test")
	config.MaxFailures = 2
	cb := NewCircuitBreaker("test", config)

	// Force open
	for i := 0; i < config.MaxFailures; i++ {
		cb.RecordFailure(errors.New("test"))
	}

	if !cb.IsOpen() {
		t.Fatal("expected circuit to be open")
	}

	// Reset
	cb.Reset()

	if !cb.IsClosed() {
		t.Error("expected circuit to be closed after reset")
	}

	if cb.Failures() != 0 {
		t.Errorf("expected failures to be reset to 0, got %d", cb.Failures())
	}
}

func TestCircuitBreaker_ConcurrentAccess(t *testing.T) {
	config := DefaultConfig("test")
	config.MaxFailures = 100
	config.MaxHalfOpenRequests = 50
	cb := NewCircuitBreaker("test", config)

	var wg sync.WaitGroup
	numGoroutines := 100
	numOperations := 100

	// Concurrent failures
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				cb.RecordFailure(errors.New("test"))
				_ = cb.State()
				_ = cb.Failures()
			}
		}()
	}

	wg.Wait()

	// Should be open after many failures
	if !cb.IsOpen() {
		t.Error("expected circuit to be open after concurrent failures")
	}

	// Reset and test concurrent successes
	cb.Reset()

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				cb.RecordSuccess()
				_ = cb.State()
				_ = cb.Failures()
			}
		}()
	}

	wg.Wait()

	// Should still be closed
	if !cb.IsClosed() {
		t.Error("expected circuit to be closed after concurrent successes")
	}

	if cb.Failures() != 0 {
		t.Errorf("expected no failures, got %d", cb.Failures())
	}
}

func TestCircuitBreaker_String(t *testing.T) {
	cb := NewCircuitBreaker("test", DefaultConfig("test"))

	str := cb.String()
	if str == "" {
		t.Error("expected non-empty string representation")
	}

	// Should contain name
	if !strings.Contains(str, "name=test") {
		t.Errorf("string should contain name, got: %s", str)
	}

	// Should contain state
	if !strings.Contains(str, "state=closed") {
		t.Errorf("string should contain state, got: %s", str)
	}
}

func TestCircuitBreakerError(t *testing.T) {
	err := NewCircuitBreakerError("test", "message", errors.New("underlying"))

	var cbErr *CircuitBreakerError
	if !errors.As(err, &cbErr) {
		t.Error("expected CircuitBreakerError")
	}

	if cbErr.Name != "test" {
		t.Errorf("expected name 'test', got '%s'", cbErr.Name)
	}

	if cbErr.Message != "message" {
		t.Errorf("expected message 'message', got '%s'", cbErr.Message)
	}

	if cbErr.Unwrap() == nil {
		t.Error("expected unwrapped error to be non-nil")
	}

	errorStr := err.Error()
	if errorStr == "" {
		t.Error("expected non-empty error string")
	}
}

func TestCircuitBreakerError_Wrap(t *testing.T) {
	wrappedErr := NewCircuitBreakerError("test", "message", ErrCircuitBreakerOpen)

	if !errors.Is(wrappedErr, ErrCircuitBreakerOpen) {
		t.Error("expected error to wrap ErrCircuitBreakerOpen")
	}
}

func TestCircuitBreaker_StateTransitions(t *testing.T) {
	config := DefaultConfig("test")
	config.MaxFailures = 2
	config.Timeout = 50 * time.Millisecond
	config.MaxHalfOpenRequests = 2
	cb := NewCircuitBreaker("test", config)

	states := []State{
		cb.State(),
	}

	// Record failures
	cb.RecordFailure(errors.New("test"))
	states = append(states, cb.State())

	cb.RecordFailure(errors.New("test"))
	states = append(states, cb.State())

	// Wait for timeout
	time.Sleep(75 * time.Millisecond)

	// Make a call to trigger half-open transition
	_ = cb.Call(func() error {
		return nil
	})
	states = append(states, cb.State())

	// Make another successful call
	_ = cb.Call(func() error {
		return nil
	})
	states = append(states, cb.State())

	expected := []State{StateClosed, StateClosed, StateOpen, StateHalfOpen, StateClosed}

	if len(states) != len(expected) {
		t.Fatalf("expected %d states, got %d", len(expected), len(states))
	}

	for i, actual := range states {
		if actual != expected[i] {
			t.Errorf("step %d: expected state %s, got %s", i, expected[i], actual)
		}
	}
}
