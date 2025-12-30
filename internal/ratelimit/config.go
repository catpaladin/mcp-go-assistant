package ratelimit

import (
	"fmt"
	"os"
	"time"
)

// Config holds rate limiting configuration
type Config struct {
	// Enabled enables or disables rate limiting
	Enabled bool `mapstructure:"enabled"`
	// Limit is the maximum number of requests per window
	Limit int `mapstructure:"limit"`
	// Window is the time window for rate limiting
	Window time.Duration `mapstructure:"window"`
	// Mode determines how rate limiting is applied (per-tool, global, ip-based)
	Mode Mode `mapstructure:"mode"`
	// Algorithm determines the rate limiting algorithm (token-bucket, sliding-window)
	Algorithm Algorithm `mapstructure:"algorithm"`
	// StoreType determines the storage backend (memory, noop)
	StoreType StoreType `mapstructure:"store_type"`
	// KeyPrefix is the prefix for all rate limit keys
	KeyPrefix string `mapstructure:"key_prefix"`
}

// ToolConfig holds tool-specific rate limiting configuration
type ToolConfig struct {
	// Enabled enables or disables rate limiting for this tool
	Enabled bool `mapstructure:"enabled"`
	// Limit is the maximum number of requests per window for this tool
	Limit int `mapstructure:"limit"`
	// Window is the time window for this tool
	Window time.Duration `mapstructure:"window"`
}

// Validate validates the rate limit configuration
func (c *Config) Validate() error {
	if c.Limit <= 0 {
		return fmt.Errorf("rate limit must be positive: %d", c.Limit)
	}

	if c.Window <= 0 {
		return fmt.Errorf("rate limit window must be positive: %v", c.Window)
	}

	switch c.Mode {
	case ModePerTool, ModeGlobal, ModeIPBased, ModeCustom:
		// Valid modes
	default:
		return fmt.Errorf("invalid rate limit mode: %s (valid: per-tool, global, ip-based, custom)", c.Mode)
	}

	switch c.Algorithm {
	case AlgorithmTokenBucket, AlgorithmSlidingWindow:
		// Valid algorithms
	default:
		return fmt.Errorf("invalid rate limit algorithm: %s (valid: token-bucket, sliding-window)", c.Algorithm)
	}

	switch c.StoreType {
	case StoreMemory, StoreNoOp:
		// Valid store types
	default:
		return fmt.Errorf("invalid store type: %s (valid: memory, noop)", c.StoreType)
	}

	return nil
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Enabled:   true,
		Limit:     100,
		Window:    1 * time.Minute,
		Mode:      ModePerTool,
		Algorithm: AlgorithmTokenBucket,
		StoreType: StoreMemory,
		KeyPrefix: "mcp",
	}
}

// LoadFromEnv loads configuration from environment variables
// Returns a config with defaults overridden by environment variables
func LoadFromEnv() *Config {
	cfg := DefaultConfig()

	// MCP_RATELIMIT_ENABLED
	if val := os.Getenv("MCP_RATELIMIT_ENABLED"); val != "" {
		cfg.Enabled = val == "true" || val == "1"
	}

	// MCP_RATELIMIT_LIMIT
	if val := os.Getenv("MCP_RATELIMIT_LIMIT"); val != "" {
		var limit int
		if _, err := fmt.Sscanf(val, "%d", &limit); err == nil && limit > 0 {
			cfg.Limit = limit
		}
	}

	// MCP_RATELIMIT_WINDOW
	if val := os.Getenv("MCP_RATELIMIT_WINDOW"); val != "" {
		if window, err := time.ParseDuration(val); err == nil && window > 0 {
			cfg.Window = window
		}
	}

	// MCP_RATELIMIT_MODE
	if val := os.Getenv("MCP_RATELIMIT_MODE"); val != "" {
		mode := Mode(val)
		switch mode {
		case ModePerTool, ModeGlobal, ModeIPBased, ModeCustom:
			cfg.Mode = mode
		}
	}

	// MCP_RATELIMIT_ALGORITHM
	if val := os.Getenv("MCP_RATELIMIT_ALGORITHM"); val != "" {
		algorithm := Algorithm(val)
		switch algorithm {
		case AlgorithmTokenBucket, AlgorithmSlidingWindow:
			cfg.Algorithm = algorithm
		}
	}

	// MCP_RATELIMIT_STORE_TYPE
	if val := os.Getenv("MCP_RATELIMIT_STORE_TYPE"); val != "" {
		storeType := StoreType(val)
		switch storeType {
		case StoreMemory, StoreNoOp:
			cfg.StoreType = storeType
		}
	}

	// MCP_RATELIMIT_KEY_PREFIX
	if val := os.Getenv("MCP_RATELIMIT_KEY_PREFIX"); val != "" {
		cfg.KeyPrefix = val
	}

	return cfg
}

// NewToolConfig creates a new tool-specific configuration with defaults
func NewToolConfig(enabled bool) *ToolConfig {
	return &ToolConfig{
		Enabled: enabled,
		Limit:   50,
		Window:  1 * time.Minute,
	}
}

// Validate validates the tool-specific configuration
func (c *ToolConfig) Validate() error {
	if !c.Enabled {
		return nil
	}

	if c.Limit <= 0 {
		return fmt.Errorf("tool rate limit must be positive: %d", c.Limit)
	}

	if c.Window <= 0 {
		return fmt.Errorf("tool rate limit window must be positive: %v", c.Window)
	}

	return nil
}
