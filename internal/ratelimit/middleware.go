package ratelimit

import (
	"context"
	"fmt"

	"mcp-go-assistant/internal/logging"
)

// Middleware wraps tool handlers with rate limiting
type Middleware struct {
	// limiter is the rate limiter instance
	limiter *Limiter
	// logger is the logger
	logger *logging.Logger
}

// NewMiddleware creates a new rate limiting middleware
func NewMiddleware(limiter *Limiter, log *logging.Logger) *Middleware {
	return &Middleware{
		limiter: limiter,
		logger:  log,
	}
}

// Handler wraps a tool handler with rate limiting
type Handler func(ctx context.Context, request interface{}) (interface{}, error)

// Wrap wraps a handler with rate limiting middleware
func (m *Middleware) Wrap(toolName string, handler Handler) Handler {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		// Extract client identifier from context
		clientID := m.extractClientID(ctx)

		// Generate rate limit key
		key := m.limiter.GenerateKey(toolName, clientID)

		// Check rate limit
		allowed, err := m.limiter.Allow(key)
		if err != nil {
			// Log error but allow request on store error to fail open
			m.logger.ErrorEvent().
				Str("tool", toolName).
				Str("key", key).
				Err(err).
				Msg("rate limit check failed, allowing request (fail open)")
			return handler(ctx, request)
		}

		if !allowed {
			// Get stats for error response
			stats, _ := m.limiter.Stats(key)

			// Return rate limit error
			return nil, &RateLimitError{
				Key:        key,
				Limit:      stats.Limit,
				Window:     stats.Window,
				RetryAfter: stats.Window,
			}
		}

		// Allow request to proceed
		return handler(ctx, request)
	}
}

// CheckRateLimit checks if a request is allowed for a tool and client
func (m *Middleware) CheckRateLimit(toolName, clientID string) error {
	key := m.limiter.GenerateKey(toolName, clientID)

	allowed, err := m.limiter.Allow(key)
	if err != nil {
		m.logger.ErrorEvent().
			Str("tool", toolName).
			Str("key", key).
			Err(err).
			Msg("rate limit check failed, allowing request (fail open)")
		return nil
	}

	if !allowed {
		stats, _ := m.limiter.Stats(key)
		return &RateLimitError{
			Key:        key,
			Limit:      stats.Limit,
			Window:     stats.Window,
			RetryAfter: stats.Window,
		}
	}

	return nil
}

// GetRateLimitInfo returns rate limit information for a tool and client
func (m *Middleware) GetRateLimitInfo(toolName, clientID string) (Stats, error) {
	key := m.limiter.GenerateKey(toolName, clientID)
	return m.limiter.Stats(key)
}

// extractClientID extracts a client identifier from the context
// In a real MCP implementation, this would come from the request metadata
func (m *Middleware) extractClientID(ctx context.Context) string {
	// TODO: Extract client ID from MCP request context
	// For now, return a default identifier
	return "default"
}

// ExtractClientID is a helper function to extract client ID from MCP request
// This is a placeholder - the actual implementation would depend on the MCP SDK
func ExtractClientID(request interface{}) string {
	// Check for common fields in MCP requests
	// This is a simplified version that should be adapted based on actual MCP request structure
	if request == nil {
		return "unknown"
	}

	// Try to get client ID from request
	// In production, this would be extracted from the MCP session or request metadata
	return "default"
}

// ExtractIPAddress extracts an IP address from the request (for IP-based rate limiting)
func ExtractIPAddress(request interface{}) string {
	// TODO: Extract IP address from MCP request context
	// For now, return a default value
	return "0.0.0.0"
}

// GetClientIdentifier extracts the appropriate client identifier based on rate limiting mode
func GetClientIdentifier(mode Mode, request interface{}) string {
	switch mode {
	case ModeIPBased:
		return ExtractIPAddress(request)
	case ModeGlobal, ModePerTool, ModeCustom:
		return ExtractClientID(request)
	default:
		return "default"
	}
}

// GenerateRateLimitKey generates a rate limit key for a tool and client
func (m *Middleware) GenerateRateLimitKey(toolName string, request interface{}) string {
	clientID := GetClientIdentifier(m.limiter.config.Mode, request)
	return m.limiter.GenerateKey(toolName, clientID)
}

// ResetRateLimit resets the rate limit counter for a tool and client
func (m *Middleware) ResetRateLimit(toolName, clientID string) error {
	key := m.limiter.GenerateKey(toolName, clientID)
	return m.limiter.Reset(key)
}

// IsRateLimitError checks if an error is a rate limit error
func IsRateLimitError(err error) bool {
	_, ok := err.(*RateLimitError)
	return ok
}

// GetRetryAfter extracts the retry-after duration from a rate limit error
func GetRetryAfter(err error) interface{} {
	if rlErr, ok := err.(*RateLimitError); ok {
		return rlErr.RetryAfter
	}
	return nil
}

// GetRateLimitHeaders returns headers that should be included in rate limit responses
func GetRateLimitHeaders(stats Stats) map[string]string {
	return map[string]string{
		"X-RateLimit-Limit":     fmt.Sprintf("%d", stats.Limit),
		"X-RateLimit-Remaining": fmt.Sprintf("%d", stats.Remaining),
		"X-RateLimit-Reset":     stats.ResetTime.Format("Mon, 02 Jan 2006 15:04:05 MST"),
	}
}
