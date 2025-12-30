package retry

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

// BackoffStrategy defines the interface for backoff strategies
type BackoffStrategy interface {
	// NextDelay returns the delay duration for the given attempt
	NextDelay(attempt uint) time.Duration
}

// ExponentialBackoff implements exponential backoff with optional jitter
type ExponentialBackoff struct {
	initialDelay time.Duration
	maxDelay     time.Duration
	multiplier   float64
	jitter       bool
}

// NewExponentialBackoff creates a new exponential backoff strategy
func NewExponentialBackoff(initialDelay, maxDelay time.Duration, multiplier float64, jitter bool) *ExponentialBackoff {
	return &ExponentialBackoff{
		initialDelay: initialDelay,
		maxDelay:     maxDelay,
		multiplier:   multiplier,
		jitter:       jitter,
	}
}

// NextDelay calculates the delay using exponential backoff
// Formula: initialDelay * (multiplier ^ attempt), capped at maxDelay
// With jitter, adds randomization up to ±25% of the calculated delay
func (b *ExponentialBackoff) NextDelay(attempt uint) time.Duration {
	// Calculate delay: initialDelay * (multiplier ^ attempt)
	delay := float64(b.initialDelay) * math.Pow(b.multiplier, float64(attempt))
	duration := time.Duration(delay)

	// Cap at maxDelay
	if b.maxDelay > 0 && duration > b.maxDelay {
		duration = b.maxDelay
	}

	// Add jitter if enabled
	if b.jitter {
		// Add randomization up to ±25% of the delay
		jitterAmount := float64(duration) * 0.25
		jitterOffset := (rand.Float64() - 0.5) * 2 * jitterAmount // -25% to +25%
		duration = time.Duration(float64(duration) + jitterOffset)
		// Ensure non-negative
		if duration < 0 {
			duration = 0
		}
	}

	return duration
}

// LinearBackoff implements linear backoff
type LinearBackoff struct {
	initialDelay time.Duration
	maxDelay     time.Duration
}

// NewLinearBackoff creates a new linear backoff strategy
func NewLinearBackoff(initialDelay, maxDelay time.Duration) *LinearBackoff {
	return &LinearBackoff{
		initialDelay: initialDelay,
		maxDelay:     maxDelay,
	}
}

// NextDelay calculates the delay using linear backoff
// Formula: initialDelay * attempt, capped at maxDelay
func (b *LinearBackoff) NextDelay(attempt uint) time.Duration {
	duration := b.initialDelay * time.Duration(attempt)

	// Cap at maxDelay
	if b.maxDelay > 0 && duration > b.maxDelay {
		duration = b.maxDelay
	}

	return duration
}

// ConstantBackoff implements constant (fixed) delay
type ConstantBackoff struct {
	delay time.Duration
}

// NewConstantBackoff creates a new constant backoff strategy
func NewConstantBackoff(delay time.Duration) *ConstantBackoff {
	return &ConstantBackoff{
		delay: delay,
	}
}

// NextDelay returns the constant delay
func (b *ConstantBackoff) NextDelay(_ uint) time.Duration {
	return b.delay
}

// NoBackoff implements immediate retry with no delay
type NoBackoff struct{}

// NewNoBackoff creates a new no-backoff strategy
func NewNoBackoff() *NoBackoff {
	return &NoBackoff{}
}

// NextDelay returns zero delay (immediate retry)
func (b *NoBackoff) NextDelay(_ uint) time.Duration {
	return 0
}

// NewStrategy creates a backoff strategy by name
func NewStrategy(name string, initialDelay, maxDelay time.Duration, multiplier float64, jitter bool) (BackoffStrategy, error) {
	switch name {
	case "exponential":
		return NewExponentialBackoff(initialDelay, maxDelay, multiplier, jitter), nil
	case "linear":
		return NewLinearBackoff(initialDelay, maxDelay), nil
	case "constant":
		return NewConstantBackoff(initialDelay), nil
	case "none":
		return NewNoBackoff(), nil
	default:
		return nil, fmt.Errorf("unknown backoff strategy: %s", name)
	}
}
