# Test Suite Documentation

## Overview

This document describes the comprehensive test suite for the MCP Go Assistant. The test suite covers all packages and targets 80%+ code coverage.

## Running Tests

### Run All Tests
```bash
go test ./...
```

### Run Tests for Specific Package
```bash
go test ./internal/config
go test ./internal/logging
go test ./internal/metrics
```

### Run Tests with Coverage
```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Run Tests with Race Detector
```bash
go test -race ./...
```

### Run Tests with Verbose Output
```bash
go test -v ./...
```

### Run Integration Tests
```bash
go test -tags=integration ./...
```

## Test Coverage

### Package Coverage Goals

| Package               | Target Coverage | Status |
|-----------------------|----------------|---------|
| internal/config       | 80%+          | ✅      |
| internal/logging      | 80%+          | ✅      |
| internal/metrics      | 80%+          | ✅      |
| internal/health      | 80%+          | ✅      |
| internal/ratelimit   | 80%+          | ✅      |
| internal/circuitbreaker | 80%+         | ✅      |
| internal/retry       | 80%+          | ✅      |
| internal/validations | 80%+          | ✅      |
| internal/types       | 80%+          | ✅      |
| internal/codereview  | 80%+          | ✅      |
| internal/godoc       | 80%+          | ✅      |
| internal/testgen     | 80%+          | ✅      |

### Viewing Coverage Report

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View coverage in HTML
go tool cover -html=coverage.out

# View coverage percentage
go tool cover -func=coverage.out
```

## Test Structure

### Unit Tests

Unit tests are located in `[package]_test.go` files and test individual functions, methods, and components.

**Example:**
```go
func TestConfig_Validate(t *testing.T) {
    tests := []struct {
        name    string
        config  *Config
        wantErr bool
    }{
        // test cases...
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
```

### Integration Tests

Integration tests are located in `[package]_integration_test.go` files and test interactions between components.

**Build Tag:**
```go
// +build integration

package config

func TestConfig_Integration(t *testing.T) {
    // integration test code
}
```

### Table-Driven Tests

Most tests use table-driven patterns for testing multiple scenarios:

```go
tests := []struct {
    name     string
    input    string
    expected string
    wantErr  bool
}{
    {"case 1", "input1", "expected1", false},
    {"case 2", "input2", "expected2", true},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        // test implementation
    })
}
```

## Test Utilities

The `internal/testutil` package provides common test helpers:

```go
import "mcp-go-assistant/internal/testutil"

// Create temporary directory
dir := testutil.TempDir(t)

// Create temporary file
path := testutil.TempFile(t, dir, "config.yaml", content)

// Wait for condition
testutil.Wait(t, condition, timeout, message)

// Assert eventually
testutil.AssertEventually(t, condition, timeout, msg)

// Setup environment variables
cleanup := testutil.SetupEnv(t, map[string]string{
    "MCP_CONFIG": "/path/to/config.yaml",
})
defer cleanup()
```

## Testing Patterns

### Testing Error Cases

```go
t.Run("error case", func(t *testing.T) {
    err := functionThatErrors()
    if err == nil {
        t.Error("expected error, got nil")
    }

    if !strings.Contains(err.Error(), "expected message") {
        t.Errorf("expected error message to contain 'expected message', got '%s'", err.Error())
    }
})
```

### Testing Concurrent Access

```go
func TestConcurrentAccess(t *testing.T) {
    done := make(chan bool)

    for i := 0; i < 100; i++ {
        go func(id int) {
            // concurrent operation
            done <- true
        }(i)
    }

    for i := 0; i < 100; i++ {
        <-done
    }
}
```

### Testing Context Cancellation

```go
func TestContextCancellation(t *testing.T) {
    ctx, cancel := context.WithCancel(context.Background())
    go func() {
        time.Sleep(10 * time.Millisecond)
        cancel()
    }()

    err := operationWithContext(ctx)
    if !errors.Is(err, context.Canceled) {
        t.Errorf("expected context.Canceled error, got %v", err)
    }
}
```

### Testing with Test Data

Create test fixtures in `[package]/testdata/` directories:

```
internal/config/testdata/
    valid_config.yaml
    invalid_config.yaml
    minimal_config.yaml
```

Use in tests:
```go
content, err := os.ReadFile("testdata/valid_config.yaml")
if err != nil {
    t.Fatalf("failed to read test data: %v", err)
}
```

## Test Files by Package

### internal/config
- `config_test.go` - Configuration loading, validation, environment variables

### internal/logging
- `logging_test.go` - Logger creation, log levels, formatting, structured logging

### internal/metrics
- `metrics_test.go` - Metrics creation, recording, collection, Prometheus format

### internal/health
- `health_test.go` - Health checks, status reporting, metadata

### internal/ratelimit
- `ratelimit_test.go` - Rate limiting, storage backends, per-tool limits, middleware

### internal/circuitbreaker
- `circuitbreaker_test.go` - Circuit breaker states, transitions, concurrent access

### internal/retry
- `retry_test.go` - Retry logic, backoff strategies, error handling, context cancellation

### internal/validations
- `validations_test.go` - Validation rules, sanitization, tool-specific validations

### internal/types
- `errors_test.go` - Error types, error wrapping, error formatting

### internal/codereview
- `codereview_test.go` - Analyzer functions, guideline matching, code review generation

### internal/godoc
- `godoc_test.go` - Go doc command execution, output parsing

### internal/testgen
- `testgen_test.go` - Test generation, interface extraction, mock generation

### internal/testutil
- `testutil.go` - Common test helpers (no tests, utilities only)

## Best Practices

### 1. Use Table-Driven Tests
Prefer table-driven tests for multiple scenarios with similar setup.

### 2. Use Subtests for Grouping
Use `t.Run()` for grouping related test cases.

### 3. Test Success and Error Cases
Always test both success and error paths.

### 4. Test Edge Cases
Include tests for edge cases like empty inputs, nil values, boundary conditions.

### 5. Use Descriptive Names
Use descriptive test names that clearly state what is being tested.

### 6. Clean Up Resources
Use `t.Cleanup()` or `defer` to clean up resources.

### 7. Test Concurrent Access
Test concurrent access where applicable using goroutines and wait groups.

### 8. Use Assert Helpers
Use `assert` or `require` from testify for cleaner assertions.

### 9. Mock External Dependencies
Use mock implementations for external dependencies.

### 10. Keep Tests Independent
Ensure tests don't depend on each other and can run in any order.

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.23.0'

      - name: Run tests
        run: go test -v -race -coverprofile=coverage.out ./...

      - name: Check coverage
        run: go tool cover -func=coverage.out | grep total

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out
```

## Troubleshooting

### Tests Timing Out

If tests timeout:
1. Check for infinite loops
2. Increase timeouts in context.WithTimeout
3. Add logging to identify slow operations

### Race Conditions

If race detector fails:
1. Add `-race` flag to test command
2. Look for data races
3. Add mutexes or atomic operations

### Low Coverage

If coverage is low:
1. Run `go test -coverprofile=coverage.out ./...`
2. View `go tool cover -html=coverage.out`
3. Identify untested lines
4. Add test cases for uncovered code

### Flaky Tests

If tests are flaky:
1. Add proper cleanup
2. Use deterministic test data
3. Add delays for async operations
4. Increase timeouts

## Contributing

When adding new packages or features:

1. Create corresponding test file
2. Target 80%+ coverage
3. Use table-driven tests
4. Test error cases
5. Test edge cases
6. Test concurrent access
7. Add to coverage report
8. Update this documentation

## Resources

- [Go Testing Guide](https://golang.org/pkg/testing/)
- [Table-Driven Tests](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)
- [Testify](https://github.com/stretchr/testify)
- [Go Coverage](https://blog.golang.org/cover)
