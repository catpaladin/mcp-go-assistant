package circuitbreaker

import (
	"fmt"
	"time"
)

// Config holds circuit breaker configuration
type Config struct {
	// MaxFailures is the threshold for opening the circuit (default 5)
	MaxFailures int
	// Timeout is the duration before attempting reset (default 30s)
	Timeout time.Duration
	// MaxHalfOpenRequests is the number of requests allowed in half-open state (default 3)
	MaxHalfOpenRequests int
	// Name is the circuit breaker name
	Name string
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.MaxFailures <= 0 {
		return fmt.Errorf("max_failures must be positive, got %d", c.MaxFailures)
	}

	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive, got %v", c.Timeout)
	}

	if c.MaxHalfOpenRequests <= 0 {
		return fmt.Errorf("max_half_open_requests must be positive, got %d", c.MaxHalfOpenRequests)
	}

	if c.Name == "" {
		return fmt.Errorf("name cannot be empty")
	}

	return nil
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig(name string) *Config {
	return &Config{
		MaxFailures:         5,
		Timeout:             30 * time.Second,
		MaxHalfOpenRequests: 3,
		Name:                name,
	}
}
