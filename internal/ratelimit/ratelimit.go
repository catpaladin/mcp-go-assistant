package ratelimit

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"mcp-go-assistant/internal/logging"
	"mcp-go-assistant/internal/metrics"
)

// Initialize metrics on package import
func init() {
	InitMetrics()
}

// RateLimiter defines the interface for rate limiting
type RateLimiter interface {
	// Allow checks if a request is allowed for the given key
	Allow(key string) (bool, error)
	// Reset resets the counter for a key
	Reset(key string) error
	// Stats returns statistics for a key
	Stats(key string) (Stats, error)
}

// Limiter implements the rate limiting logic
type Limiter struct {
	// config holds the rate limit configuration
	config *Config
	// store is the storage backend
	store Store
	// toolConfigs holds tool-specific configurations
	toolConfigs map[string]*ToolConfig
	// metrics is the metrics collector
	metrics *metrics.Metrics
	// logger is the logger
	logger *logging.Logger
	// mutex provides thread-safe access to tool configs
	mutex sync.RWMutex
}

// NewLimiter creates a new rate limiter with the given configuration
func NewLimiter(cfg *Config, store Store, m *metrics.Metrics, log *logging.Logger) *Limiter {
	if err := cfg.Validate(); err != nil {
		panic(fmt.Sprintf("invalid rate limit config: %v", err))
	}

	return &Limiter{
		config:      cfg,
		store:       store,
		toolConfigs: make(map[string]*ToolConfig),
		metrics:     m,
		logger:      log,
	}
}

// Allow checks if a request is allowed for the given key
func (l *Limiter) Allow(key string) (bool, error) {
	if !l.config.Enabled {
		return true, nil
	}

	// Determine the limit to use
	limit := l.config.Limit
	window := l.config.Window

	// Check if key is for a specific tool
	if l.config.Mode == ModePerTool {
		toolName := l.extractToolName(key)
		if toolName != "" {
			l.mutex.RLock()
			toolCfg := l.toolConfigs[toolName]
			l.mutex.RUnlock()

			if toolCfg != nil && toolCfg.Enabled {
				limit = toolCfg.Limit
				window = toolCfg.Window
			}
		}
	}

	// Increment counter
	count, err := l.store.Increment(key, window)
	if err != nil {
		l.logger.ErrorEvent().
			Str("key", key).
			Err(err).
			Msg("failed to increment rate limit counter")
		return false, fmt.Errorf("rate limit store error: %w", err)
	}

	allowed := count <= limit

	// Record metrics and logging
	toolName := l.extractToolName(key)
	if toolName == "" {
		toolName = "unknown"
	}

	mode := string(l.config.Mode)

	if allowed {
		RecordAllowed(toolName, mode)
		l.logger.DebugEvent().
			Str("key", key).
			Str("tool", toolName).
			Str("mode", mode).
			Int("count", count).
			Int("limit", limit).
			Msg("rate limit check passed")
	} else {
		RecordRejected(toolName, mode)
		l.logger.WarnEvent().
			Str("key", key).
			Str("tool", toolName).
			Str("mode", mode).
			Int("count", count).
			Int("limit", limit).
			Dur("window", window).
			Msg("rate limit exceeded")
	}

	// Update current count gauge
	SetCurrent(toolName, mode, count)

	return allowed, nil
}

// Reset resets the counter for a key
func (l *Limiter) Reset(key string) error {
	if err := l.store.Reset(key); err != nil {
		return fmt.Errorf("failed to reset rate limit: %w", err)
	}

	l.logger.DebugEvent().
		Str("key", key).
		Msg("rate limit counter reset")

	return nil
}

// Stats returns statistics for a key
func (l *Limiter) Stats(key string) (Stats, error) {
	// Determine the limit to use
	limit := l.config.Limit
	window := l.config.Window

	// Check if key is for a specific tool
	if l.config.Mode == ModePerTool {
		toolName := l.extractToolName(key)
		if toolName != "" {
			l.mutex.RLock()
			toolCfg := l.toolConfigs[toolName]
			l.mutex.RUnlock()

			if toolCfg != nil && toolCfg.Enabled {
				limit = toolCfg.Limit
				window = toolCfg.Window
			}
		}
	}

	count, err := l.store.Get(key)
	if err != nil {
		return Stats{}, fmt.Errorf("failed to get rate limit stats: %w", err)
	}

	return Stats{
		Limit:     limit,
		Window:    window,
		Current:   count,
		Remaining: max(0, limit-count),
		Allowed:   count <= limit,
		ResetTime: time.Now().Add(window),
	}, nil
}

// SetToolConfig sets the configuration for a specific tool
func (l *Limiter) SetToolConfig(toolName string, cfg *ToolConfig) error {
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid tool config: %w", err)
	}

	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.toolConfigs[toolName] = cfg

	l.logger.DebugEvent().
		Str("tool", toolName).
		Int("limit", cfg.Limit).
		Dur("window", cfg.Window).
		Msg("tool rate limit config updated")

	return nil
}

// GetToolConfig returns the configuration for a specific tool
func (l *Limiter) GetToolConfig(toolName string) *ToolConfig {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	if cfg, exists := l.toolConfigs[toolName]; exists {
		return cfg
	}

	// Return default tool config if not set
	return NewToolConfig(l.config.Enabled)
}

// GenerateKey generates a rate limit key based on the mode
func (l *Limiter) GenerateKey(toolName, clientID string) string {
	var key strings.Builder

	if l.config.KeyPrefix != "" {
		key.WriteString(l.config.KeyPrefix)
		key.WriteString(":")
	}

	switch l.config.Mode {
	case ModePerTool:
		key.WriteString("tool:")
		if toolName != "" {
			key.WriteString(toolName)
		}
		if clientID != "" {
			key.WriteString(":")
			key.WriteString(clientID)
		}
	case ModeGlobal:
		key.WriteString("global:")
		if clientID != "" {
			key.WriteString(clientID)
		}
	case ModeIPBased:
		key.WriteString("ip:")
		if clientID != "" {
			key.WriteString(clientID)
		}
	case ModeCustom:
		// Use the provided key directly
		key.WriteString(clientID)
	}

	return key.String()
}

// extractToolName extracts the tool name from a key
func (l *Limiter) extractToolName(key string) string {
	parts := strings.Split(key, ":")
	if len(parts) >= 2 && parts[0] == l.config.KeyPrefix {
		if parts[1] == "tool" && len(parts) >= 3 {
			return parts[2]
		}
	}
	return ""
}

// Close cleans up resources
func (l *Limiter) Close() error {
	if memStore, ok := l.store.(*MemoryStore); ok {
		memStore.Stop()
	}
	return nil
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
