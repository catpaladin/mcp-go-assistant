# Rate Limiting Framework

The rate limiting framework provides protection against abuse and DoS attacks for the MCP Go Assistant.

## Overview

The rate limiting framework supports:
- **Multiple rate limiting modes**: Per-tool, global, IP-based, and custom key-based
- **Configurable algorithms**: Token bucket and sliding window (token bucket implemented)
- **Flexible storage**: In-memory store with automatic cleanup, or no-op store for testing
- **Tool-specific limits**: Different rate limits for each tool (GoDoc, CodeReview, TestGen)
- **Comprehensive metrics**: Prometheus metrics for monitoring
- **Structured logging**: Detailed logs for rate limit events

## Components

### Types (`types.go`)
- `Stats`: Rate limiting statistics for a key
- `RateLimitError`: Error when rate limit is exceeded
- `Mode`: Rate limiting mode (per-tool, global, ip-based, custom)
- `Algorithm`: Rate limiting algorithm (token-bucket, sliding-window)
- `StoreType`: Storage backend type (memory, noop)

### Configuration (`config.go`)
- `Config`: Main rate limiting configuration
- `ToolConfig`: Tool-specific configuration
- Environment variable support: `MCP_RATELIMIT_*`

### Storage (`store.go`)
- `Store`: Interface for storage backends
- `MemoryStore`: Thread-safe in-memory storage with automatic cleanup
- `NoOpStore`: No-op store for testing/disabled state

### Rate Limiter (`ratelimit.go`)
- `RateLimiter`: Interface for rate limiting
- `Limiter`: Main rate limiting implementation
- Support for per-tool and global rate limiting
- Metrics and logging integration

### Middleware (`middleware.go`)
- `Middleware`: Wrapper for MCP tool handlers
- Helper functions for rate limit checking
- Rate limit error handling

### Metrics (`metrics.go`)
- `Metrics`: Prometheus metrics for rate limiting
- Global metrics instance
- Metrics: `mcp_ratelimit_allowed_total`, `mcp_ratelimit_rejected_total`, `mcp_ratelimit_current`

## Configuration

### Default Configuration
```yaml
rate_limit:
  enabled: true
  limit: 100
  window: 1m
  mode: "per-tool"
  algorithm: "token-bucket"
  store_type: "memory"
  tools:
    godoc:
      enabled: true
      limit: 50
      window: 1m
    code-review:
      enabled: true
      limit: 30
      window: 1m
    test-gen:
      enabled: true
      limit: 30
      window: 1m
```

### Environment Variables
- `MCP_RATELIMIT_ENABLED`: Enable/disable rate limiting
- `MCP_RATELIMIT_LIMIT`: Default request limit per window
- `MCP_RATELIMIT_WINDOW`: Time window duration
- `MCP_RATELIMIT_MODE`: Rate limiting mode
- `MCP_RATELIMIT_ALGORITHM`: Rate limiting algorithm
- `MCP_RATELIMIT_STORE_TYPE`: Storage backend type

## Usage

### Basic Usage
```go
import "mcp-go-assistant/internal/ratelimit"

// Create configuration
cfg := ratelimit.DefaultConfig()

// Create store
store := ratelimit.NewMemoryStore()
defer store.Stop()

// Create rate limiter
limiter := ratelimit.NewLimiter(cfg, store, metrics, logger)

// Check if request is allowed
allowed, err := limiter.Allow("mcp:tool:godoc:client123")
if err != nil {
    // Handle error
}
if !allowed {
    // Rate limit exceeded
}
```

### Using Middleware
```go
// Create middleware
middleware := ratelimit.NewMiddleware(limiter, logger)

// Check rate limit before tool execution
if err := middleware.CheckRateLimit("godoc", "client123"); err != nil {
    return nil, err
}

// Execute tool
result, err := executeTool(...)
```

### Tool-Specific Configuration
```go
// Set tool-specific rate limit
toolCfg := &ratelimit.ToolConfig{
    Enabled: true,
    Limit:   50,
    Window:  1 * time.Minute,
}
err := limiter.SetToolConfig("godoc", toolCfg)
```

## Rate Limiting Modes

### Per-Tool Mode
Rate limits are applied per tool and per client:
```
Key format: mcp:tool:{toolName}:{clientID}
Example: mcp:tool:godoc:client123
```

### Global Mode
Rate limits are applied globally per client:
```
Key format: mcp:global:{clientID}
Example: mcp:global:client123
```

### IP-Based Mode
Rate limits are applied per IP address:
```
Key format: mcp:ip:{ipAddress}
Example: mcp:ip:192.168.1.1
```

### Custom Mode
Rate limits use custom keys provided by the application:
```
Key format: mcp:{customKey}
Example: mcp:user123-session456
```

## Metrics

The following Prometheus metrics are exposed:

- `mcp_ratelimit_allowed_total`: Total allowed requests by tool and mode
- `mcp_ratelimit_rejected_total`: Total rejected requests by tool and mode
- `mcp_ratelimit_current`: Current request count in the window
- `mcp_ratelimit_limit_exceeded_total`: Times limit was exceeded

Example queries:
```promql
# Rate of allowed requests
rate(mcp_ratelimit_allowed_total[5m])

# Rate of rejected requests
rate(mcp_ratelimit_rejected_total[5m])

# Current request count
mcp_ratelimit_current

# Rejection rate
rate(mcp_ratelimit_rejected_total[5m]) / rate(mcp_ratelimit_allowed_total[5m])
```

## Testing

Run tests for the rate limiting package:
```bash
go test ./internal/ratelimit/...
```

Run tests with coverage:
```bash
go test -cover ./internal/ratelimit/...
```

## Thread Safety

All components are thread-safe and can be used concurrently:
- `MemoryStore` uses `sync.RWMutex` for thread-safe access
- `Limiter` uses `sync.RWMutex` for tool configuration access
- Metrics use `sync.Mutex` for thread-safe updates

## Error Handling

The framework implements a "fail-open" policy:
- If rate limiting fails due to storage errors, requests are allowed
- Rate limit errors are logged at WARN level
- Store errors are logged at ERROR level

Example error handling:
```go
allowed, err := limiter.Allow(key)
if err != nil {
    // Log error but allow request (fail-open)
    logger.Warn().Err(err).Msg("rate limit check failed, allowing")
    allowed = true
}
```

## Cleanup

For memory store, cleanup is automatic:
- Cleanup runs every 5 minutes
- Entries older than 10 minutes are removed
- Cleanup can be stopped explicitly:
```go
store.Stop()
```
