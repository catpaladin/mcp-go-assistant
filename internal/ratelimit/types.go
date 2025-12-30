package ratelimit

import (
	"fmt"
	"time"

	"mcp-go-assistant/internal/types"
)

// Stats represents rate limiting statistics for a key
type Stats struct {
	// Limit is the maximum requests allowed in the window
	Limit int `json:"limit"`
	// Window is the time duration for the rate limit
	Window time.Duration `json:"window"`
	// Current is the current number of requests in the window
	Current int `json:"current"`
	// Remaining is the number of requests remaining in the window
	Remaining int `json:"remaining"`
	// ResetTime is the time when the window resets
	ResetTime time.Time `json:"reset_time"`
	// Allowed indicates if the request was allowed
	Allowed bool `json:"allowed"`
}

// RateLimitError represents a rate limit exceeded error
type RateLimitError struct {
	// Key is the rate limit key that triggered the error
	Key string `json:"key"`
	// Limit is the rate limit that was exceeded
	Limit int `json:"limit"`
	// Window is the time window
	Window time.Duration `json:"window"`
	// RetryAfter is the duration to wait before retrying
	RetryAfter time.Duration `json:"retry_after"`
}

// Error implements the error interface
func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limit exceeded for key %s: %d requests per %v exceeded, retry after %v",
		e.Key, e.Limit, e.Window, e.RetryAfter)
}

// ToMCPError converts a RateLimitError to an MCPError
func (e *RateLimitError) ToMCPError() types.MCPError {
	details := make(map[string]interface{})
	if e.Key != "" {
		details["key"] = e.Key
	}
	if e.Limit > 0 {
		details["limit"] = e.Limit
	}
	if e.Window > 0 {
		details["window"] = e.Window.String()
	}
	if e.RetryAfter > 0 {
		details["retry_after"] = e.RetryAfter.String()
	}

	return types.NewRateLimitError(e.Error(), details)
}

// Mode represents the rate limiting mode
type Mode string

const (
	// ModePerTool rate limits per tool
	ModePerTool Mode = "per-tool"
	// ModeGlobal rate limits globally
	ModeGlobal Mode = "global"
	// ModeIPBased rate limits by IP address
	ModeIPBased Mode = "ip-based"
	// ModeCustom rate limits using custom keys
	ModeCustom Mode = "custom"
)

// Algorithm represents the rate limiting algorithm
type Algorithm string

const (
	// AlgorithmTokenBucket uses the token bucket algorithm
	AlgorithmTokenBucket Algorithm = "token-bucket"
	// AlgorithmSlidingWindow uses the sliding window algorithm
	AlgorithmSlidingWindow Algorithm = "sliding-window"
)

// StoreType represents the storage backend type
type StoreType string

const (
	// StoreMemory uses in-memory storage
	StoreMemory StoreType = "memory"
	// StoreNoOp disables rate limiting (always allows)
	StoreNoOp StoreType = "noop"
)

// Bucket represents a rate limit bucket
type Bucket struct {
	// Count is the current count in the bucket
	Count int `json:"count"`
	// LastUpdate is the last time the bucket was updated
	LastUpdate time.Time `json:"last_update"`
	// WindowStart is the start of the current window
	WindowStart time.Time `json:"window_start"`
}
