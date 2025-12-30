## 🚀 Quick Start

### 1. Setup (one-time)

```bash
# Download dependencies
go mod tidy

# Build
go build -o mcp-go-assistant ./cmd/mcp-go-assistant

# Create config from example
cp config.example.yaml config.yaml
```

### 2. Run with defaults

```bash
./mcp-go-assistant
```

### 3. Run with debug logging

```bash
MCP_LOG_LEVEL=debug ./mcp-go-assistant
```

### 4. Run with custom config

```bash
MCP_CONFIG=/path/to/config.yaml ./mcp-go-assistant
```

## 📊 Monitor Metrics

### View all metrics

```bash
curl http://localhost:8080/metrics
```

### Example metrics output

```
mcp_requests_total{method="call_tool",tool="go-doc",status="success"} 150
mcp_request_duration_seconds_bucket{tool="go-doc",le="0.005"} 100
mcp_tool_calls_total{tool="code-review",status="success"} 75
mcp_uptime_seconds 1234.56
```

## 🔧 Configuration

### Default config (config.example.yaml)

```yaml
server:
  name: "mcp-go-assistant"
  version: "1.2.0"
  host: "0.0.0.0"
  port: 8080

logging:
  level: "info"
  format: "json"
  output_path: "stdout"

metrics:
  enabled: true
  path: "/metrics"

tools:
  godoc_timeout: 30s
  code_review_timeout: 60s
  test_gen_timeout: 45s

timeouts:
  default: 30s
  shutdown: 30s
  request: 60s
```

### Environment variables

```bash
MCP_LOG_LEVEL=debug
MCP_LOG_FORMAT=console
MCP_PORT=9090
MCP_METRICS_ENABLED=true
MCP_GODOC_TIMEOUT=60s
```

## 📈 Available Metrics

| Metric                         | Type      | Labels                   | Description         |
| ------------------------------ | --------- | ------------------------ | ------------------- |
| `mcp_requests_total`           | Counter   | method, tool, status     | Total requests      |
| `mcp_request_duration_seconds` | Histogram | method, tool             | Request duration    |
| `mcp_request_errors_total`     | Counter   | method, tool, error_type | Request errors      |
| `mcp_requests_active`          | Gauge     | tool                     | Active requests     |
| `mcp_tool_calls_total`         | Counter   | tool, status             | Tool invocations    |
| `mcp_tool_duration_seconds`    | Histogram | tool                     | Tool execution time |
| `mcp_tool_errors_total`        | Counter   | tool, error_type         | Tool errors         |
| `mcp_uptime_seconds`           | Gauge     | -                        | Server uptime       |

## 🎚️ Log Levels

- `trace` - Most detailed logging
- `debug` - Debug information
- `info` - General information (default)
- `warn` - Warning messages
- `error` - Error messages
- `fatal` - Fatal errors (exits)
- `panic` - Panic messages (exits)

## 🔍 Logging Example

### JSON format (production)

```json
{
  "level": "info",
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "tool": "go-doc",
  "package_path": "fmt",
  "duration_ms": 45,
  "message": "request completed successfully",
  "time": "2024-01-15T10:30:00Z"
}
```

### Console format (development)

```
10:30:00 INF request completed successfully tool=go-doc package_path=fmt duration_ms=45 request_id=550e8400-e29b-41d4-a716-446655440000
```

## 🚨 Error Types

Errors are automatically classified for metrics:

- `timeout` - Request/tool timeout
- `not_found` - Resource not found
- `permission` - Permission denied
- `parse_error` - Syntax/parse errors
- `unknown` - Other errors

## 🐛 Troubleshooting

### High error rate?

```bash
# Enable debug logging
MCP_LOG_LEVEL=debug ./mcp-go-assistant

# Check error metrics
curl http://localhost:8080/metrics | grep mcp_request_errors_total
```

### Slow responses?

```bash
# Check duration metrics
curl http://localhost:8080/metrics | grep mcp_request_duration_seconds

# Increase timeouts if needed
MCP_CODE_REVIEW_TIMEOUT=120s ./mcp-go-assistant
```

### Memory issues?

```bash
# Check health (if endpoint exposed)
# Review goroutine count in metadata
# Review memory usage in health checks
```
