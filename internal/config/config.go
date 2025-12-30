package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"mcp-go-assistant/internal/circuitbreaker"
	"mcp-go-assistant/internal/ratelimit"
	"mcp-go-assistant/internal/retry"

	"github.com/spf13/viper"
)

// ValidationConfig contains validation-related settings
type ValidationConfig struct {
	MaxInputSize  int      `mapstructure:"max_input_size"`
	AllowedChars  string   `mapstructure:"allowed_chars"`
	EnabledRules  []string `mapstructure:"enabled_rules"`
	DisabledRules []string `mapstructure:"disabled_rules"`
}

// Config holds all application configuration
type Config struct {
	Server        ServerConfig        `mapstructure:"server"`
	Logging       LoggingConfig       `mapstructure:"logging"`
	Metrics       MetricsConfig       `mapstructure:"metrics"`
	Tools         ToolsConfig         `mapstructure:"tools"`
	Timeouts      TimeoutConfig       `mapstructure:"timeouts"`
	Validations   ValidationConfig    `mapstructure:"validations"`
	RateLimit     RateLimitConfig     `mapstructure:"rate_limit"`
	ErrorHandling ErrorHandlingConfig `mapstructure:"error_handling"`
	Retry         RetryConfig         `mapstructure:"retry"`
}

// ServerConfig contains server-related settings
type ServerConfig struct {
	Name         string        `mapstructure:"name"`
	Version      string        `mapstructure:"version"`
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// LoggingConfig contains logging-related settings
type LoggingConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"` // json, console
	OutputPath string `mapstructure:"output_path"`
	NoColor    bool   `mapstructure:"no_color"`
}

// ErrorHandlingConfig contains error-handling-related settings
type ErrorHandlingConfig struct {
	Verbosity        string            `mapstructure:"verbosity"`         // minimal, detailed, debug
	IncludeStack     bool              `mapstructure:"include_stack"`     // Include stack traces in error responses
	ResponseFormat   string            `mapstructure:"response_format"`   // json, text
	ExposeDetails    bool              `mapstructure:"expose_details"`    // Expose error details to clients
	LogAllErrors     bool              `mapstructure:"log_all_errors"`    // Log all errors, including client errors
	TrackMetrics     bool              `mapstructure:"track_metrics"`     // Track error metrics
	CategoryMappings map[string]string `mapstructure:"category_mappings"` // Custom category mappings
}

// MetricsConfig contains metrics-related settings
type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Path    string `mapstructure:"path"`
}

// ToolsConfig contains tool-specific settings
type ToolsConfig struct {
	GoDocTimeout             time.Duration        `mapstructure:"godoc_timeout"`
	CodeReviewTimeout        time.Duration        `mapstructure:"code_review_timeout"`
	TestGenTimeout           time.Duration        `mapstructure:"test_gen_timeout"`
	GoDocCircuitBreaker      CircuitBreakerConfig `mapstructure:"godoc_circuit_breaker"`
	CodeReviewCircuitBreaker CircuitBreakerConfig `mapstructure:"code_review_circuit_breaker"`
	TestGenCircuitBreaker    CircuitBreakerConfig `mapstructure:"test_gen_circuit_breaker"`
}

// CircuitBreakerConfig contains circuit breaker configuration
type CircuitBreakerConfig struct {
	MaxFailures         int           `mapstructure:"max_failures"`
	Timeout             time.Duration `mapstructure:"timeout"`
	MaxHalfOpenRequests int           `mapstructure:"max_half_open_requests"`
}

// ToCircuitBreakerConfig converts to circuitbreaker.Config
func (c *CircuitBreakerConfig) ToCircuitBreakerConfig(name string) *circuitbreaker.Config {
	return &circuitbreaker.Config{
		Name:                name,
		MaxFailures:         c.MaxFailures,
		Timeout:             c.Timeout,
		MaxHalfOpenRequests: c.MaxHalfOpenRequests,
	}
}

// TimeoutConfig contains general timeout settings
type TimeoutConfig struct {
	Default     time.Duration `mapstructure:"default"`
	Shutdown    time.Duration `mapstructure:"shutdown"`
	Request     time.Duration `mapstructure:"request"`
	GracePeriod time.Duration `mapstructure:"grace_period"`
}

// RateLimitConfig contains rate limiting configuration
type RateLimitConfig struct {
	Enabled   bool                           `mapstructure:"enabled"`
	Limit     int                            `mapstructure:"limit"`
	Window    time.Duration                  `mapstructure:"window"`
	Mode      string                         `mapstructure:"mode"`
	Algorithm string                         `mapstructure:"algorithm"`
	StoreType string                         `mapstructure:"store_type"`
	Tools     map[string]RateLimitToolConfig `mapstructure:"tools"`
}

// RateLimitToolConfig contains tool-specific rate limiting configuration
type RateLimitToolConfig struct {
	Enabled bool          `mapstructure:"enabled"`
	Limit   int           `mapstructure:"limit"`
	Window  time.Duration `mapstructure:"window"`
}

// RetryConfig contains retry configuration
type RetryConfig struct {
	Enabled      bool                       `mapstructure:"enabled"`
	MaxAttempts  uint                       `mapstructure:"max_attempts"`
	InitialDelay time.Duration              `mapstructure:"initial_delay"`
	MaxDelay     time.Duration              `mapstructure:"max_delay"`
	Multiplier   float64                    `mapstructure:"multiplier"`
	Jitter       bool                       `mapstructure:"jitter"`
	Strategy     string                     `mapstructure:"strategy"`
	Tools        map[string]RetryToolConfig `mapstructure:"tools"`
}

// RetryToolConfig contains tool-specific retry configuration
type RetryToolConfig struct {
	Enabled      bool          `mapstructure:"enabled"`
	MaxAttempts  uint          `mapstructure:"max_attempts"`
	InitialDelay time.Duration `mapstructure:"initial_delay"`
	MaxDelay     time.Duration `mapstructure:"max_delay"`
	Strategy     string        `mapstructure:"strategy"`
}

// ToRetryConfig converts to retry.Config
func (c *RetryConfig) ToRetryConfig() *retry.Config {
	cfg := &retry.Config{
		MaxAttempts:  c.MaxAttempts,
		InitialDelay: c.InitialDelay,
		MaxDelay:     c.MaxDelay,
		Multiplier:   c.Multiplier,
		Jitter:       c.Jitter,
		Strategy:     c.Strategy,
	}
	return cfg
}

// ToRetryToolConfig converts tool config to retry.Config
func (c *RetryToolConfig) ToRetryConfig() *retry.Config {
	cfg := &retry.Config{
		MaxAttempts:  c.MaxAttempts,
		InitialDelay: c.InitialDelay,
		MaxDelay:     c.MaxDelay,
		Multiplier:   2.0,  // Default multiplier for tool configs
		Jitter:       true, // Default jitter for tool configs
		Strategy:     c.Strategy,
	}
	return cfg
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Name:         "mcp-go-assistant",
			Version:      "1.2.0",
			Host:         "0.0.0.0",
			Port:         8080,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		},
		Logging: LoggingConfig{
			Level:      "info",
			Format:     "json",
			OutputPath: "stdout",
			NoColor:    false,
		},
		ErrorHandling: ErrorHandlingConfig{
			Verbosity:        "detailed",
			IncludeStack:     false,
			ResponseFormat:   "json",
			ExposeDetails:    false,
			LogAllErrors:     true,
			TrackMetrics:     true,
			CategoryMappings: map[string]string{},
		},
		Metrics: MetricsConfig{
			Enabled: true,
			Path:    "/metrics",
		},
		Tools: ToolsConfig{
			GoDocTimeout:      30 * time.Second,
			CodeReviewTimeout: 60 * time.Second,
			TestGenTimeout:    45 * time.Second,
			GoDocCircuitBreaker: CircuitBreakerConfig{
				MaxFailures:         5,
				Timeout:             30 * time.Second,
				MaxHalfOpenRequests: 3,
			},
			CodeReviewCircuitBreaker: CircuitBreakerConfig{
				MaxFailures:         5,
				Timeout:             30 * time.Second,
				MaxHalfOpenRequests: 3,
			},
			TestGenCircuitBreaker: CircuitBreakerConfig{
				MaxFailures:         5,
				Timeout:             30 * time.Second,
				MaxHalfOpenRequests: 3,
			},
		},
		Timeouts: TimeoutConfig{
			Default:     30 * time.Second,
			Shutdown:    30 * time.Second,
			Request:     60 * time.Second,
			GracePeriod: 10 * time.Second,
		},
		Validations: ValidationConfig{
			MaxInputSize: 1024 * 1024, // 1MB
			AllowedChars: `[a-zA-Z0-9_ ./\-]`,
			EnabledRules: []string{
				"not_empty",
				"package_path",
				"file_path",
				"code_safety",
				"symbol_name",
			},
			DisabledRules: []string{},
		},
		RateLimit: RateLimitConfig{
			Enabled:   true,
			Limit:     100,
			Window:    1 * time.Minute,
			Mode:      string(ratelimit.ModePerTool),
			Algorithm: string(ratelimit.AlgorithmTokenBucket),
			StoreType: string(ratelimit.StoreMemory),
			Tools: map[string]RateLimitToolConfig{
				"godoc": {
					Enabled: true,
					Limit:   50,
					Window:  1 * time.Minute,
				},
				"code-review": {
					Enabled: true,
					Limit:   30,
					Window:  1 * time.Minute,
				},
				"test-gen": {
					Enabled: true,
					Limit:   30,
					Window:  1 * time.Minute,
				},
			},
		},
		Retry: RetryConfig{
			Enabled:      true,
			MaxAttempts:  3,
			InitialDelay: 1 * time.Second,
			MaxDelay:     30 * time.Second,
			Multiplier:   2.0,
			Jitter:       true,
			Strategy:     "exponential",
			Tools: map[string]RetryToolConfig{
				"godoc": {
					Enabled:      true,
					MaxAttempts:  3,
					InitialDelay: 1 * time.Second,
					MaxDelay:     15 * time.Second,
					Strategy:     "exponential",
				},
				"code-review": {
					Enabled:      true,
					MaxAttempts:  2,
					InitialDelay: 500 * time.Millisecond,
					MaxDelay:     10 * time.Second,
					Strategy:     "exponential",
				},
				"test-gen": {
					Enabled:      true,
					MaxAttempts:  2,
					InitialDelay: 500 * time.Millisecond,
					MaxDelay:     10 * time.Second,
					Strategy:     "exponential",
				},
			},
		},
	}
}

// Load loads configuration from file and environment variables
// Config file path can be provided via MCP_CONFIG environment variable
// Environment variables override config file settings
func Load() (*Config, error) {
	v := viper.New()
	cfg := DefaultConfig()

	// Set defaults
	setDefaults(v)

	// Read config file if specified
	configPath := os.Getenv("MCP_CONFIG")
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		// Try to find config in current directory
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("/etc/mcp-go-assistant")
	}

	// Read config file (ignore error if not found)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Environment variable bindings
	bindEnvVars(v)

	// Unmarshal config
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	if err := validateLogLevel(c.Logging.Level); err != nil {
		return err
	}

	if err := validateLogFormat(c.Logging.Format); err != nil {
		return err
	}

	if c.Timeouts.Default <= 0 {
		return fmt.Errorf("default timeout must be positive")
	}

	if c.Timeouts.Shutdown <= 0 {
		return fmt.Errorf("shutdown timeout must be positive")
	}

	return nil
}

// validateLogLevel checks if the log level is valid
func validateLogLevel(level string) error {
	validLevels := map[string]bool{
		"trace": true,
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
		"fatal": true,
		"panic": true,
	}

	if !validLevels[strings.ToLower(level)] {
		return fmt.Errorf("invalid log level: %s (valid levels: trace, debug, info, warn, error, fatal, panic)", level)
	}

	return nil
}

// validateLogFormat checks if the log format is valid
func validateLogFormat(format string) error {
	validFormats := map[string]bool{
		"json":    true,
		"console": true,
	}

	if !validFormats[strings.ToLower(format)] {
		return fmt.Errorf("invalid log format: %s (valid formats: json, console)", format)
	}

	return nil
}

// setDefaults sets default values in viper
func setDefaults(v *viper.Viper) {
	cfg := DefaultConfig()

	v.SetDefault("server.name", cfg.Server.Name)
	v.SetDefault("server.version", cfg.Server.Version)
	v.SetDefault("server.host", cfg.Server.Host)
	v.SetDefault("server.port", cfg.Server.Port)
	v.SetDefault("server.read_timeout", cfg.Server.ReadTimeout)
	v.SetDefault("server.write_timeout", cfg.Server.WriteTimeout)

	v.SetDefault("logging.level", cfg.Logging.Level)
	v.SetDefault("logging.format", cfg.Logging.Format)
	v.SetDefault("logging.output_path", cfg.Logging.OutputPath)
	v.SetDefault("logging.no_color", cfg.Logging.NoColor)

	v.SetDefault("metrics.enabled", cfg.Metrics.Enabled)
	v.SetDefault("metrics.path", cfg.Metrics.Path)

	v.SetDefault("tools.godoc_timeout", cfg.Tools.GoDocTimeout)
	v.SetDefault("tools.code_review_timeout", cfg.Tools.CodeReviewTimeout)
	v.SetDefault("tools.test_gen_timeout", cfg.Tools.TestGenTimeout)

	// Circuit breaker defaults
	v.SetDefault("tools.godoc_circuit_breaker.max_failures", cfg.Tools.GoDocCircuitBreaker.MaxFailures)
	v.SetDefault("tools.godoc_circuit_breaker.timeout", cfg.Tools.GoDocCircuitBreaker.Timeout)
	v.SetDefault("tools.godoc_circuit_breaker.max_half_open_requests", cfg.Tools.GoDocCircuitBreaker.MaxHalfOpenRequests)
	v.SetDefault("tools.code_review_circuit_breaker.max_failures", cfg.Tools.CodeReviewCircuitBreaker.MaxFailures)
	v.SetDefault("tools.code_review_circuit_breaker.timeout", cfg.Tools.CodeReviewCircuitBreaker.Timeout)
	v.SetDefault("tools.code_review_circuit_breaker.max_half_open_requests", cfg.Tools.CodeReviewCircuitBreaker.MaxHalfOpenRequests)
	v.SetDefault("tools.test_gen_circuit_breaker.max_failures", cfg.Tools.TestGenCircuitBreaker.MaxFailures)
	v.SetDefault("tools.test_gen_circuit_breaker.timeout", cfg.Tools.TestGenCircuitBreaker.Timeout)
	v.SetDefault("tools.test_gen_circuit_breaker.max_half_open_requests", cfg.Tools.TestGenCircuitBreaker.MaxHalfOpenRequests)

	v.SetDefault("timeouts.default", cfg.Timeouts.Default)
	v.SetDefault("timeouts.shutdown", cfg.Timeouts.Shutdown)
	v.SetDefault("timeouts.request", cfg.Timeouts.Request)
	v.SetDefault("timeouts.grace_period", cfg.Timeouts.GracePeriod)

	// Validations
	v.SetDefault("validations.max_input_size", cfg.Validations.MaxInputSize)
	v.SetDefault("validations.allowed_chars", cfg.Validations.AllowedChars)
	v.SetDefault("validations.enabled_rules", cfg.Validations.EnabledRules)
	v.SetDefault("validations.disabled_rules", cfg.Validations.DisabledRules)

	// Rate limiting
	v.SetDefault("rate_limit.enabled", cfg.RateLimit.Enabled)
	v.SetDefault("rate_limit.limit", cfg.RateLimit.Limit)
	v.SetDefault("rate_limit.window", cfg.RateLimit.Window)
	v.SetDefault("rate_limit.mode", cfg.RateLimit.Mode)
	v.SetDefault("rate_limit.algorithm", cfg.RateLimit.Algorithm)
	v.SetDefault("rate_limit.store_type", cfg.RateLimit.StoreType)

	// Retry
	v.SetDefault("retry.enabled", cfg.Retry.Enabled)
	v.SetDefault("retry.max_attempts", cfg.Retry.MaxAttempts)
	v.SetDefault("retry.initial_delay", cfg.Retry.InitialDelay)
	v.SetDefault("retry.max_delay", cfg.Retry.MaxDelay)
	v.SetDefault("retry.multiplier", cfg.Retry.Multiplier)
	v.SetDefault("retry.jitter", cfg.Retry.Jitter)
	v.SetDefault("retry.strategy", cfg.Retry.Strategy)
}

// bindEnvVars binds environment variables to config keys
func bindEnvVars(v *viper.Viper) {
	// Server
	v.BindEnv("server.host", "MCP_HOST")
	v.BindEnv("server.port", "MCP_PORT")
	v.BindEnv("server.read_timeout", "MCP_READ_TIMEOUT")
	v.BindEnv("server.write_timeout", "MCP_WRITE_TIMEOUT")

	// Logging
	v.BindEnv("logging.level", "MCP_LOG_LEVEL")
	v.BindEnv("logging.format", "MCP_LOG_FORMAT")
	v.BindEnv("logging.output_path", "MCP_LOG_OUTPUT")
	v.BindEnv("logging.no_color", "MCP_LOG_NO_COLOR")

	// Metrics
	v.BindEnv("metrics.enabled", "MCP_METRICS_ENABLED")
	v.BindEnv("metrics.path", "MCP_METRICS_PATH")

	// Tools
	v.BindEnv("tools.godoc_timeout", "MCP_GODOC_TIMEOUT")
	v.BindEnv("tools.code_review_timeout", "MCP_CODE_REVIEW_TIMEOUT")
	v.BindEnv("tools.test_gen_timeout", "MCP_TEST_GEN_TIMEOUT")

	// Circuit breaker settings
	v.BindEnv("tools.godoc_circuit_breaker.max_failures", "MCP_GODOC_CB_MAX_FAILURES")
	v.BindEnv("tools.godoc_circuit_breaker.timeout", "MCP_GODOC_CB_TIMEOUT")
	v.BindEnv("tools.godoc_circuit_breaker.max_half_open_requests", "MCP_GODOC_CB_MAX_HALF_OPEN")
	v.BindEnv("tools.code_review_circuit_breaker.max_failures", "MCP_CODE_REVIEW_CB_MAX_FAILURES")
	v.BindEnv("tools.code_review_circuit_breaker.timeout", "MCP_CODE_REVIEW_CB_TIMEOUT")
	v.BindEnv("tools.code_review_circuit_breaker.max_half_open_requests", "MCP_CODE_REVIEW_CB_MAX_HALF_OPEN")
	v.BindEnv("tools.test_gen_circuit_breaker.max_failures", "MCP_TEST_GEN_CB_MAX_FAILURES")
	v.BindEnv("tools.test_gen_circuit_breaker.timeout", "MCP_TEST_GEN_CB_TIMEOUT")
	v.BindEnv("tools.test_gen_circuit_breaker.max_half_open_requests", "MCP_TEST_GEN_CB_MAX_HALF_OPEN")

	// Timeouts
	v.BindEnv("timeouts.default", "MCP_TIMEOUT_DEFAULT")
	v.BindEnv("timeouts.shutdown", "MCP_TIMEOUT_SHUTDOWN")
	v.BindEnv("timeouts.request", "MCP_TIMEOUT_REQUEST")
	v.BindEnv("timeouts.grace_period", "MCP_TIMEOUT_GRACE_PERIOD")

	// Validations
	v.BindEnv("validations.max_input_size", "MCP_VALIDATION_MAX_INPUT_SIZE")
	v.BindEnv("validations.allowed_chars", "MCP_VALIDATION_ALLOWED_CHARS")

	// Rate limiting
	v.BindEnv("rate_limit.enabled", "MCP_RATELIMIT_ENABLED")
	v.BindEnv("rate_limit.limit", "MCP_RATELIMIT_LIMIT")
	v.BindEnv("rate_limit.window", "MCP_RATELIMIT_WINDOW")
	v.BindEnv("rate_limit.mode", "MCP_RATELIMIT_MODE")
	v.BindEnv("rate_limit.algorithm", "MCP_RATELIMIT_ALGORITHM")
	v.BindEnv("rate_limit.store_type", "MCP_RATELIMIT_STORE_TYPE")

	// Error handling
	v.BindEnv("error_handling.verbosity", "MCP_ERROR_VERBOSITY")
	v.BindEnv("error_handling.include_stack", "MCP_ERROR_INCLUDE_STACK")
	v.BindEnv("error_handling.response_format", "MCP_ERROR_RESPONSE_FORMAT")
	v.BindEnv("error_handling.expose_details", "MCP_ERROR_EXPOSE_DETAILS")
	v.BindEnv("error_handling.log_all_errors", "MCP_ERROR_LOG_ALL")
	v.BindEnv("error_handling.track_metrics", "MCP_ERROR_TRACK_METRICS")

	// Retry
	v.BindEnv("retry.enabled", "MCP_RETRY_ENABLED")
	v.BindEnv("retry.max_attempts", "MCP_RETRY_MAX_ATTEMPTS")
	v.BindEnv("retry.initial_delay", "MCP_RETRY_INITIAL_DELAY")
	v.BindEnv("retry.max_delay", "MCP_RETRY_MAX_DELAY")
	v.BindEnv("retry.multiplier", "MCP_RETRY_MULTIPLIER")
	v.BindEnv("retry.jitter", "MCP_RETRY_JITTER")
	v.BindEnv("retry.strategy", "MCP_RETRY_STRATEGY")
}
