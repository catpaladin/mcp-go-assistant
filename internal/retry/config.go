package retry

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds retry configuration
type Config struct {
	MaxAttempts  uint          `mapstructure:"max_attempts"`
	InitialDelay time.Duration `mapstructure:"initial_delay"`
	MaxDelay     time.Duration `mapstructure:"max_delay"`
	Multiplier   float64       `mapstructure:"multiplier"`
	Jitter       bool          `mapstructure:"jitter"`
	Strategy     string        `mapstructure:"strategy"`
}

// Validate validates the retry configuration
func (c *Config) Validate() error {
	if c.MaxAttempts < 1 {
		return fmt.Errorf("max_attempts must be at least 1")
	}
	if c.InitialDelay < 0 {
		return fmt.Errorf("initial_delay cannot be negative")
	}
	if c.MaxDelay < 0 {
		return fmt.Errorf("max_delay cannot be negative")
	}
	if c.MaxDelay > 0 && c.InitialDelay > c.MaxDelay {
		return fmt.Errorf("initial_delay cannot be greater than max_delay")
	}
	if c.Multiplier <= 0 {
		return fmt.Errorf("multiplier must be positive")
	}
	if c.Strategy == "" {
		c.Strategy = "exponential"
	}
	validStrategies := map[string]bool{
		"exponential": true,
		"linear":      true,
		"constant":    true,
		"none":        true,
	}
	if !validStrategies[c.Strategy] {
		return fmt.Errorf("invalid strategy: %s (must be exponential, linear, constant, or none)", c.Strategy)
	}
	return nil
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
		Jitter:       true,
		Strategy:     "exponential",
	}
}

// LoadFromEnv loads configuration from environment variables
func LoadFromEnv() *Config {
	cfg := DefaultConfig()

	// Read environment variables
	if val := os.Getenv("MCP_RETRY_MAX_ATTEMPTS"); val != "" {
		if attempts, err := strconv.ParseUint(val, 10, 32); err == nil {
			cfg.MaxAttempts = uint(attempts)
		}
	}
	if val := os.Getenv("MCP_RETRY_INITIAL_DELAY"); val != "" {
		if delay, err := time.ParseDuration(val); err == nil {
			cfg.InitialDelay = delay
		}
	}
	if val := os.Getenv("MCP_RETRY_MAX_DELAY"); val != "" {
		if delay, err := time.ParseDuration(val); err == nil {
			cfg.MaxDelay = delay
		}
	}
	if val := os.Getenv("MCP_RETRY_MULTIPLIER"); val != "" {
		if mult, err := strconv.ParseFloat(val, 64); err == nil {
			cfg.Multiplier = mult
		}
	}
	if val := os.Getenv("MCP_RETRY_JITTER"); val != "" {
		if jitter, err := strconv.ParseBool(val); err == nil {
			cfg.Jitter = jitter
		}
	}
	if val := os.Getenv("MCP_RETRY_STRATEGY"); val != "" {
		cfg.Strategy = val
	}

	return cfg
}

// Clone returns a copy of the configuration
func (c *Config) Clone() *Config {
	return &Config{
		MaxAttempts:  c.MaxAttempts,
		InitialDelay: c.InitialDelay,
		MaxDelay:     c.MaxDelay,
		Multiplier:   c.Multiplier,
		Jitter:       c.Jitter,
		Strategy:     c.Strategy,
	}
}
