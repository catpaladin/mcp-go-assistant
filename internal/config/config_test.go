package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Server.Name != "mcp-go-assistant" {
		t.Errorf("expected server name 'mcp-go-assistant', got '%s'", cfg.Server.Name)
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("expected server port 8080, got %d", cfg.Server.Port)
	}

	if cfg.Logging.Level != "info" {
		t.Errorf("expected log level 'info', got '%s'", cfg.Logging.Level)
	}

	if cfg.Metrics.Enabled != true {
		t.Errorf("expected metrics enabled, got %v", cfg.Metrics.Enabled)
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  DefaultConfig(),
			wantErr: false,
		},
		{
			name: "invalid port - negative",
			config: func() *Config {
				cfg := DefaultConfig()
				cfg.Server.Port = -1
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "invalid port - zero",
			config: func() *Config {
				cfg := DefaultConfig()
				cfg.Server.Port = 0
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "invalid port - too large",
			config: func() *Config {
				cfg := DefaultConfig()
				cfg.Server.Port = 70000
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "invalid log level",
			config: func() *Config {
				cfg := DefaultConfig()
				cfg.Logging.Level = "invalid"
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "invalid log format",
			config: func() *Config {
				cfg := DefaultConfig()
				cfg.Logging.Format = "invalid"
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "zero default timeout",
			config: func() *Config {
				cfg := DefaultConfig()
				cfg.Timeouts.Default = 0
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "zero shutdown timeout",
			config: func() *Config {
				cfg := DefaultConfig()
				cfg.Timeouts.Shutdown = 0
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "valid custom timeouts",
			config: func() *Config {
				cfg := DefaultConfig()
				cfg.Timeouts.Default = 10 * time.Second
				cfg.Timeouts.Shutdown = 20 * time.Second
				return cfg
			}(),
			wantErr: false,
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

func TestValidateLogLevel(t *testing.T) {
	tests := []struct {
		name    string
		level   string
		wantErr bool
	}{
		{"trace", "trace", false},
		{"debug", "debug", false},
		{"info", "info", false},
		{"warn", "warn", false},
		{"error", "error", false},
		{"fatal", "fatal", false},
		{"panic", "panic", false},
		{"invalid", "invalid", true},
		{"", "", true},
		{"INFO", "INFO", false}, // Case insensitive
		{"DEBUG", "DEBUG", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateLogLevel(tt.level)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateLogLevel(%q) error = %v, wantErr %v", tt.level, err, tt.wantErr)
			}
		})
	}
}

func TestValidateLogFormat(t *testing.T) {
	tests := []struct {
		name    string
		format  string
		wantErr bool
	}{
		{"json", "json", false},
		{"console", "console", false},
		{"JSON", "JSON", false}, // Case insensitive
		{"CONSOLE", "CONSOLE", false},
		{"invalid", "invalid", true},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateLogFormat(tt.format)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateLogFormat(%q) error = %v, wantErr %v", tt.format, err, tt.wantErr)
			}
		})
	}
}

func TestCircuitBreakerConfig_ToCircuitBreakerConfig(t *testing.T) {
	cfg := CircuitBreakerConfig{
		MaxFailures:         10,
		Timeout:             30 * time.Second,
		MaxHalfOpenRequests: 5,
	}

	cbCfg := cfg.ToCircuitBreakerConfig("test-cb")

	if cbCfg.Name != "test-cb" {
		t.Errorf("expected name 'test-cb', got '%s'", cbCfg.Name)
	}

	if cbCfg.MaxFailures != 10 {
		t.Errorf("expected MaxFailures 10, got %d", cbCfg.MaxFailures)
	}

	if cbCfg.Timeout != 30*time.Second {
		t.Errorf("expected Timeout 30s, got %v", cbCfg.Timeout)
	}

	if cbCfg.MaxHalfOpenRequests != 5 {
		t.Errorf("expected MaxHalfOpenRequests 5, got %d", cbCfg.MaxHalfOpenRequests)
	}
}

func TestRetryConfig_ToRetryConfig(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:  5,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.5,
		Jitter:       true,
		Strategy:     "exponential",
	}

	retryCfg := cfg.ToRetryConfig()

	if retryCfg.MaxAttempts != 5 {
		t.Errorf("expected MaxAttempts 5, got %d", retryCfg.MaxAttempts)
	}

	if retryCfg.InitialDelay != 1*time.Second {
		t.Errorf("expected InitialDelay 1s, got %v", retryCfg.InitialDelay)
	}

	if retryCfg.MaxDelay != 30*time.Second {
		t.Errorf("expected MaxDelay 30s, got %v", retryCfg.MaxDelay)
	}

	if retryCfg.Multiplier != 2.5 {
		t.Errorf("expected Multiplier 2.5, got %v", retryCfg.Multiplier)
	}

	if retryCfg.Jitter != true {
		t.Errorf("expected Jitter true, got %v", retryCfg.Jitter)
	}

	if retryCfg.Strategy != "exponential" {
		t.Errorf("expected Strategy 'exponential', got '%s'", retryCfg.Strategy)
	}
}

func TestRetryToolConfig_ToRetryConfig(t *testing.T) {
	cfg := RetryToolConfig{
		MaxAttempts:  3,
		InitialDelay: 500 * time.Millisecond,
		MaxDelay:     10 * time.Second,
		Strategy:     "exponential",
	}

	retryCfg := cfg.ToRetryConfig()

	if retryCfg.MaxAttempts != 3 {
		t.Errorf("expected MaxAttempts 3, got %d", retryCfg.MaxAttempts)
	}

	if retryCfg.Multiplier != 2.0 {
		t.Errorf("expected default Multiplier 2.0, got %v", retryCfg.Multiplier)
	}

	if retryCfg.Jitter != true {
		t.Errorf("expected default Jitter true, got %v", retryCfg.Jitter)
	}
}

func TestLoad_WithValidConfigFile(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
server:
  name: "test-server"
  port: 9090

logging:
  level: "debug"
  format: "console"

metrics:
  enabled: false
  path: "/custom-metrics"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	// Set environment variable to point to config file
	t.Setenv("MCP_CONFIG", configPath)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.Server.Name != "test-server" {
		t.Errorf("expected server name 'test-server', got '%s'", cfg.Server.Name)
	}

	if cfg.Server.Port != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.Server.Port)
	}

	if cfg.Logging.Level != "debug" {
		t.Errorf("expected log level 'debug', got '%s'", cfg.Logging.Level)
	}

	if cfg.Logging.Format != "console" {
		t.Errorf("expected log format 'console', got '%s'", cfg.Logging.Format)
	}

	if cfg.Metrics.Enabled != false {
		t.Errorf("expected metrics disabled, got %v", cfg.Metrics.Enabled)
	}

	if cfg.Metrics.Path != "/custom-metrics" {
		t.Errorf("expected metrics path '/custom-metrics', got '%s'", cfg.Metrics.Path)
	}
}

func TestLoad_WithEnvVars(t *testing.T) {
	// Unset any existing config file
	t.Setenv("MCP_CONFIG", "")

	// Set environment variables
	t.Setenv("MCP_PORT", "9999")
	t.Setenv("MCP_LOG_LEVEL", "trace")
	t.Setenv("MCP_METRICS_ENABLED", "false")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.Server.Port != 9999 {
		t.Errorf("expected port 9999 from env, got %d", cfg.Server.Port)
	}

	if cfg.Logging.Level != "trace" {
		t.Errorf("expected log level 'trace' from env, got '%s'", cfg.Logging.Level)
	}

	if cfg.Metrics.Enabled != false {
		t.Errorf("expected metrics disabled from env, got %v", cfg.Metrics.Enabled)
	}
}

func TestLoad_WithMissingConfigFile(t *testing.T) {
	// Unset any existing config file to use defaults
	t.Setenv("MCP_CONFIG", "")

	// Should not error, should use defaults
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() should not error with missing config file, got: %v", err)
	}

	// Verify defaults are used
	if cfg.Server.Port != 8080 {
		t.Errorf("expected default port 8080, got %d", cfg.Server.Port)
	}
}

func TestLoad_WithInvalidConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write invalid YAML
	if err := os.WriteFile(configPath, []byte("invalid: yaml: content:\n  - broken"), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	t.Setenv("MCP_CONFIG", configPath)

	_, err := Load()
	if err == nil {
		t.Error("expected error for invalid config file, got nil")
	}
}

func TestConfig_Defaults(t *testing.T) {
	cfg := DefaultConfig()

	// Check server defaults
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("expected default host '0.0.0.0', got '%s'", cfg.Server.Host)
	}

	if cfg.Server.ReadTimeout != 30*time.Second {
		t.Errorf("expected default read timeout 30s, got %v", cfg.Server.ReadTimeout)
	}

	if cfg.Server.WriteTimeout != 30*time.Second {
		t.Errorf("expected default write timeout 30s, got %v", cfg.Server.WriteTimeout)
	}

	// Check tool defaults
	if cfg.Tools.GoDocTimeout != 30*time.Second {
		t.Errorf("expected default godoc timeout 30s, got %v", cfg.Tools.GoDocTimeout)
	}

	if cfg.Tools.CodeReviewTimeout != 60*time.Second {
		t.Errorf("expected default code review timeout 60s, got %v", cfg.Tools.CodeReviewTimeout)
	}

	if cfg.Tools.TestGenTimeout != 45*time.Second {
		t.Errorf("expected default test gen timeout 45s, got %v", cfg.Tools.TestGenTimeout)
	}

	// Check validation defaults
	if cfg.Validations.MaxInputSize != 1024*1024 {
		t.Errorf("expected default max input size 1MB, got %d", cfg.Validations.MaxInputSize)
	}

	// Check error handling defaults
	if cfg.ErrorHandling.Verbosity != "detailed" {
		t.Errorf("expected default verbosity 'detailed', got '%s'", cfg.ErrorHandling.Verbosity)
	}

	if cfg.ErrorHandling.TrackMetrics != true {
		t.Errorf("expected default track_metrics true, got %v", cfg.ErrorHandling.TrackMetrics)
	}

	// Check rate limit defaults
	if cfg.RateLimit.Enabled != true {
		t.Errorf("expected default rate limit enabled, got %v", cfg.RateLimit.Enabled)
	}

	if cfg.RateLimit.Limit != 100 {
		t.Errorf("expected default rate limit 100, got %d", cfg.RateLimit.Limit)
	}

	// Check retry defaults
	if cfg.Retry.Enabled != true {
		t.Errorf("expected default retry enabled, got %v", cfg.Retry.Enabled)
	}

	if cfg.Retry.MaxAttempts != 3 {
		t.Errorf("expected default max attempts 3, got %d", cfg.Retry.MaxAttempts)
	}
}

func TestConfig_ToolSpecificConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Check tool-specific rate limit configs
	if cfg.RateLimit.Tools["godoc"].Limit != 50 {
		t.Errorf("expected godoc rate limit 50, got %d", cfg.RateLimit.Tools["godoc"].Limit)
	}

	if cfg.RateLimit.Tools["code-review"].Limit != 30 {
		t.Errorf("expected code-review rate limit 30, got %d", cfg.RateLimit.Tools["code-review"].Limit)
	}

	if cfg.RateLimit.Tools["test-gen"].Limit != 30 {
		t.Errorf("expected test-gen rate limit 30, got %d", cfg.RateLimit.Tools["test-gen"].Limit)
	}

	// Check tool-specific retry configs
	if cfg.Retry.Tools["godoc"].MaxAttempts != 3 {
		t.Errorf("expected godoc max attempts 3, got %d", cfg.Retry.Tools["godoc"].MaxAttempts)
	}

	if cfg.Retry.Tools["code-review"].MaxAttempts != 2 {
		t.Errorf("expected code-review max attempts 2, got %d", cfg.Retry.Tools["code-review"].MaxAttempts)
	}

	if cfg.Retry.Tools["test-gen"].MaxAttempts != 2 {
		t.Errorf("expected test-gen max attempts 2, got %d", cfg.Retry.Tools["test-gen"].MaxAttempts)
	}
}
