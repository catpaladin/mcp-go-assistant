package ratelimit

import (
	"os"
	"sync"
	"testing"
	"time"
)

// TestDefaultConfig tests default configuration
func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Enabled != true {
		t.Errorf("Expected Enabled=true, got %v", cfg.Enabled)
	}

	if cfg.Limit != 100 {
		t.Errorf("Expected Limit=100, got %d", cfg.Limit)
	}

	if cfg.Window != 1*time.Minute {
		t.Errorf("Expected Window=1m, got %v", cfg.Window)
	}

	if cfg.Mode != ModePerTool {
		t.Errorf("Expected Mode=%s, got %s", ModePerTool, cfg.Mode)
	}

	if cfg.Algorithm != AlgorithmTokenBucket {
		t.Errorf("Expected Algorithm=%s, got %s", AlgorithmTokenBucket, cfg.Algorithm)
	}

	if cfg.StoreType != StoreMemory {
		t.Errorf("Expected StoreType=%s, got %s", StoreMemory, cfg.StoreType)
	}
}

// TestConfigValidation tests configuration validation
func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name:    "valid config",
			cfg:     DefaultConfig(),
			wantErr: false,
		},
		{
			name: "invalid limit - zero",
			cfg: &Config{
				Limit:  0,
				Window: 1 * time.Minute,
				Mode:   ModePerTool,
			},
			wantErr: true,
		},
		{
			name: "invalid limit - negative",
			cfg: &Config{
				Limit:  -1,
				Window: 1 * time.Minute,
				Mode:   ModePerTool,
			},
			wantErr: true,
		},
		{
			name: "invalid window - zero",
			cfg: &Config{
				Limit:  100,
				Window: 0,
				Mode:   ModePerTool,
			},
			wantErr: true,
		},
		{
			name: "invalid mode",
			cfg: &Config{
				Limit:  100,
				Window: 1 * time.Minute,
				Mode:   Mode("invalid"),
			},
			wantErr: true,
		},
		{
			name: "invalid algorithm",
			cfg: &Config{
				Limit:     100,
				Window:    1 * time.Minute,
				Mode:      ModePerTool,
				Algorithm: Algorithm("invalid"),
			},
			wantErr: true,
		},
		{
			name: "invalid store type",
			cfg: &Config{
				Limit:     100,
				Window:    1 * time.Minute,
				Mode:      ModePerTool,
				Algorithm: AlgorithmTokenBucket,
				StoreType: StoreType("invalid"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestNewToolConfig tests tool-specific configuration
func TestNewToolConfig(t *testing.T) {
	cfg := NewToolConfig(true)

	if cfg.Enabled != true {
		t.Errorf("Expected Enabled=true, got %v", cfg.Enabled)
	}

	if cfg.Limit != 50 {
		t.Errorf("Expected Limit=50, got %d", cfg.Limit)
	}

	if cfg.Window != 1*time.Minute {
		t.Errorf("Expected Window=1m, got %v", cfg.Window)
	}
}

// TestToolConfigValidation tests tool configuration validation
func TestToolConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *ToolConfig
		wantErr bool
	}{
		{
			name: "valid enabled config",
			cfg: &ToolConfig{
				Enabled: true,
				Limit:   50,
				Window:  1 * time.Minute,
			},
			wantErr: false,
		},
		{
			name: "disabled config",
			cfg: &ToolConfig{
				Enabled: false,
				Limit:   0,
				Window:  0,
			},
			wantErr: false,
		},
		{
			name: "invalid enabled config - zero limit",
			cfg: &ToolConfig{
				Enabled: true,
				Limit:   0,
				Window:  1 * time.Minute,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestMemoryStore tests in-memory store
func TestMemoryStore(t *testing.T) {
	store := NewMemoryStore()
	defer store.Stop()

	key := "test-key"
	window := 1 * time.Minute

	// Test increment - first time
	count, err := store.Increment(key, window)
	if err != nil {
		t.Fatalf("Increment() error = %v", err)
	}
	if count != 1 {
		t.Errorf("Expected count=1, got %d", count)
	}

	// Test increment - second time
	count, err = store.Increment(key, window)
	if err != nil {
		t.Fatalf("Increment() error = %v", err)
	}
	if count != 2 {
		t.Errorf("Expected count=2, got %d", count)
	}

	// Test get
	count, err = store.Get(key)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if count != 2 {
		t.Errorf("Expected count=2, got %d", count)
	}

	// Test reset
	err = store.Reset(key)
	if err != nil {
		t.Fatalf("Reset() error = %v", err)
	}

	count, err = store.Get(key)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if count != 0 {
		t.Errorf("Expected count=0 after reset, got %d", count)
	}
}

// TestMemoryStoreWindowReset tests window reset
func TestMemoryStoreWindowReset(t *testing.T) {
	store := NewMemoryStore()
	defer store.Stop()

	key := "test-key"
	window := 100 * time.Millisecond

	// Increment first time
	count, _ := store.Increment(key, window)
	if count != 1 {
		t.Errorf("Expected count=1, got %d", count)
	}

	// Wait for window to expire
	time.Sleep(150 * time.Millisecond)

	// Should reset to 1 (new window)
	count, err := store.Increment(key, window)
	if err != nil {
		t.Fatalf("Increment() error = %v", err)
	}
	if count != 1 {
		t.Errorf("Expected count=1 after window reset, got %d", count)
	}
}

// TestNoOpStore tests no-op store
func TestNoOpStore(t *testing.T) {
	store := NewNoOpStore()

	key := "test-key"
	window := 1 * time.Minute

	// Test increment - always returns 1
	count, err := store.Increment(key, window)
	if err != nil {
		t.Fatalf("Increment() error = %v", err)
	}
	if count != 1 {
		t.Errorf("Expected count=1, got %d", count)
	}

	// Test get - always returns 0
	count, err = store.Get(key)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if count != 0 {
		t.Errorf("Expected count=0, got %d", count)
	}

	// Test reset - no error
	err = store.Reset(key)
	if err != nil {
		t.Fatalf("Reset() error = %v", err)
	}
}

// TestMemoryStoreConcurrency tests concurrent access to memory store
func TestMemoryStoreConcurrency(t *testing.T) {
	store := NewMemoryStore()
	defer store.Stop()

	key := "test-key"
	window := 1 * time.Minute
	numGoroutines := 100
	incrementsPerGoroutine := 10

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < incrementsPerGoroutine; j++ {
				_, err := store.Increment(key, window)
				if err != nil {
					errors <- err
				}
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("Increment() error = %v", err)
	}

	// Verify final count
	count, err := store.Get(key)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	expectedCount := numGoroutines * incrementsPerGoroutine
	if count != expectedCount {
		t.Errorf("Expected count=%d, got %d", expectedCount, count)
	}
}

// TestRateLimitError tests rate limit error
func TestRateLimitError(t *testing.T) {
	err := &RateLimitError{
		Key:        "test-key",
		Limit:      100,
		Window:     1 * time.Minute,
		RetryAfter: 30 * time.Second,
	}

	errorStr := err.Error()
	expectedStr := "rate limit exceeded for key test-key: 100 requests per 1m0s exceeded, retry after 30s"

	if errorStr != expectedStr {
		t.Errorf("Expected error string: %s, got: %s", expectedStr, errorStr)
	}
}

// TestLimiterGenerateKey tests key generation
func TestLimiterGenerateKey(t *testing.T) {
	tests := []struct {
		name      string
		mode      Mode
		toolName  string
		clientID  string
		prefix    string
		wantStart string
	}{
		{
			name:      "per-tool mode",
			mode:      ModePerTool,
			toolName:  "godoc",
			clientID:  "client123",
			prefix:    "mcp",
			wantStart: "mcp:tool:godoc:client123",
		},
		{
			name:      "global mode",
			mode:      ModeGlobal,
			toolName:  "",
			clientID:  "client123",
			prefix:    "mcp",
			wantStart: "mcp:global:client123",
		},
		{
			name:      "ip-based mode",
			mode:      ModeIPBased,
			toolName:  "",
			clientID:  "192.168.1.1",
			prefix:    "mcp",
			wantStart: "mcp:ip:192.168.1.1",
		},
		{
			name:      "custom mode",
			mode:      ModeCustom,
			toolName:  "",
			clientID:  "custom-key-123",
			prefix:    "mcp",
			wantStart: "mcp:custom-key-123",
		},
		{
			name:      "no prefix",
			mode:      ModePerTool,
			toolName:  "godoc",
			clientID:  "client123",
			prefix:    "",
			wantStart: "tool:godoc:client123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Mode:      tt.mode,
				KeyPrefix: tt.prefix,
			}

			store := NewNoOpStore()
			limiter := &Limiter{
				config: cfg,
				store:  store,
			}

			key := limiter.GenerateKey(tt.toolName, tt.clientID)
			if key != tt.wantStart {
				t.Errorf("GenerateKey() = %v, want %v", key, tt.wantStart)
			}
		})
	}
}

// TestExtractToolName tests tool name extraction
func TestExtractToolName(t *testing.T) {
	cfg := &Config{
		KeyPrefix: "mcp",
		Mode:      ModePerTool,
	}

	limiter := &Limiter{
		config: cfg,
	}

	tests := []struct {
		key      string
		expected string
	}{
		{"mcp:tool:godoc:client123", "godoc"},
		{"mcp:tool:code-review:client456", "code-review"},
		{"mcp:global:client789", ""},
		{"other:tool:test", ""},
		{"invalid-key", ""},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			result := limiter.extractToolName(tt.key)
			if result != tt.expected {
				t.Errorf("extractToolName(%q) = %v, want %v", tt.key, result, tt.expected)
			}
		})
	}
}

// TestLoadFromEnv tests loading config from environment
func TestLoadFromEnv(t *testing.T) {
	// Save original environment values
	origEnabled := setEnv("MCP_RATELIMIT_ENABLED", "false")
	defer unsetEnv("MCP_RATELIMIT_ENABLED", origEnabled)

	origLimit := setEnv("MCP_RATELIMIT_LIMIT", "200")
	defer unsetEnv("MCP_RATELIMIT_LIMIT", origLimit)

	origWindow := setEnv("MCP_RATELIMIT_WINDOW", "2m")
	defer unsetEnv("MCP_RATELIMIT_WINDOW", origWindow)

	origMode := setEnv("MCP_RATELIMIT_MODE", "global")
	defer unsetEnv("MCP_RATELIMIT_MODE", origMode)

	cfg := LoadFromEnv()

	if cfg.Enabled != false {
		t.Errorf("Expected Enabled=false, got %v", cfg.Enabled)
	}

	if cfg.Limit != 200 {
		t.Errorf("Expected Limit=200, got %d", cfg.Limit)
	}

	if cfg.Window != 2*time.Minute {
		t.Errorf("Expected Window=2m, got %v", cfg.Window)
	}

	if cfg.Mode != ModeGlobal {
		t.Errorf("Expected Mode=%s, got %s", ModeGlobal, cfg.Mode)
	}
}

// Helper functions for environment variable testing

func setEnv(key, value string) string {
	orig := ""
	if val, exists := os.LookupEnv(key); exists {
		orig = val
	}
	_ = os.Setenv(key, value)
	return orig
}

func unsetEnv(key, orig string) {
	if orig != "" {
		_ = os.Setenv(key, orig)
	} else {
		_ = os.Unsetenv(key)
	}
}
