# Test Files and Examples

This directory contains test files and guidelines for testing the MCP Go Assistant functionality.

## Directory Structure

```
tests/
├── README.md           # This file
├── examples/           # Go code examples for testing
│   ├── simple_good.go      # Well-written Go code example
│   ├── test_code.go        # Basic test with documentation issues
│   ├── complex_test.go     # Complex code with multiple issues
│   ├── performance_issues.go # Code with performance problems
│   └── test_review.json    # Example JSON request
└── guidelines/         # Custom coding guidelines
    ├── guidelines.md       # General best practices
    ├── security.md         # Security-focused guidelines
    ├── performance.md      # Performance-focused guidelines
    └── project-examples/   # Example project guidelines
        ├── .mcp-guidelines.md      # General project guidelines
        ├── microservice-guidelines.md  # Microservice-specific
        ├── webapp-guidelines.md        # Web application guidelines
        └── library-guidelines.md       # Library development guidelines
```

## Testing the Go Documentation Tool

### Basic Package Documentation
```bash
./bin/client fmt
./bin/client net/http
./bin/client encoding/json
```

### Specific Symbol Documentation
```bash
./bin/client fmt Printf
./bin/client net/http Server
./bin/client encoding/json Marshal
```

## Testing the Code Review Tool

### Basic Code Review (no guidelines)
```bash
./bin/review-client tests/examples/simple_good.go
./bin/review-client tests/examples/test_code.go
./bin/review-client tests/examples/complex_test.go
./bin/review-client tests/examples/performance_issues.go
```

### Code Review with Custom Guidelines
```bash
# General guidelines
./bin/review-client tests/examples/complex_test.go tests/guidelines/guidelines.md

# Security-focused review
./bin/review-client tests/examples/complex_test.go tests/guidelines/security.md "focus on security"

# Performance-focused review
./bin/review-client tests/examples/performance_issues.go tests/guidelines/performance.md "focus on performance"
```

### Code Review with Hints
```bash
# Focus on specific areas
./bin/review-client tests/examples/complex_test.go "" "focus on maintainability"
./bin/review-client tests/examples/performance_issues.go "" "focus on performance"
./bin/review-client tests/examples/test_code.go "" "focus on documentation"
```

## Example Test Files

### `simple_good.go`
Well-structured Go code following best practices. Should receive a high score with minimal issues.

### `test_code.go`
Basic Go code with common issues like missing documentation for exported functions.

### `complex_test.go`
Complex code with multiple issues including:
- Naming convention violations
- Large structs
- Missing documentation
- Error handling problems
- Security concerns (unsafe package usage)
- Global variables

### `performance_issues.go`
Code specifically designed to test performance analysis including:
- String concatenation in loops
- Functions with too many parameters
- Deep nesting
- Inefficient algorithms

## Custom Guidelines

### `guidelines.md`
General Go best practices and coding standards.

### `security.md`
Security-focused guidelines for:
- Input validation
- Authentication/authorization
- Data protection
- Error handling
- Dependencies
- Unsafe operations

### `performance.md`
Performance-focused guidelines for:
- String operations
- Memory management
- Goroutines and concurrency
- I/O operations
- Data structures
- Profiling and monitoring

## Expected Results

### Good Code (`simple_good.go`)
- **Score**: 90-100
- **Issues**: Few or none
- **Categories**: Minimal warnings

### Basic Issues (`test_code.go`)
- **Score**: 85-95
- **Issues**: Missing documentation
- **Categories**: Documentation warnings

### Complex Issues (`complex_test.go`)
- **Score**: 60-80
- **Issues**: Multiple categories
- **Categories**: Naming, documentation, error handling, security

### Performance Issues (`performance_issues.go`)
- **Score**: 50-70
- **Issues**: Performance and structure problems
- **Categories**: Performance, structure, complexity

## Manual Testing

You can also test the MCP server directly:

```bash
# Initialize and test go-doc tool
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0.0"}}}
{"jsonrpc":"2.0","method":"notifications/initialized"}
{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"go-doc","arguments":{"package_path":"fmt"}}}' | ./bin/mcp-go-assistant
```

## Adding New Test Cases

To add new test cases:

1. **Code Examples**: Add `.go` files to `tests/examples/`
2. **Guidelines**: Add `.md` files to `tests/guidelines/`
3. **Update Documentation**: Update this README with the new test cases
4. **Test**: Verify the new examples work with both tools

## LLM Prompt Examples

### Go Documentation Prompts

Use these prompts with your MCP-enabled LLM client:

```
"Show me the documentation for the fmt package"
"What does context.WithTimeout do?"
"How do I use sync.Mutex?"
"Get the documentation for http.Handler interface"
"What methods are available on *os.File?"
```

### Code Review Prompts

#### Basic Code Review
```
"Review this Go code for best practices:

```go
[paste code from tests/examples/test_code.go]
```"
```

#### Performance-Focused Review
```
"Analyze this code for performance issues:

```go
[paste code from tests/examples/performance_issues.go]
```"
```

#### Security-Focused Review
```
"Review this code for security vulnerabilities:

```go
[paste code from tests/examples/complex_test.go]
```"
```

#### With Custom Guidelines
```
"Review this Go code following these guidelines:
- Functions should not exceed 30 lines
- Always use error wrapping
- Prefer dependency injection over globals
- Use meaningful variable names

```go
[paste your code here]
```"
```

### Expected LLM Responses

When you use these prompts, your LLM should:

1. **For Documentation**: Provide the exact Go documentation with explanations
2. **For Code Review**: Return structured analysis including:
   - Issues found with severity levels
   - Specific suggestions for improvement
   - Code quality score (0-100)
   - Metrics (lines of code, complexity, etc.)

### Testing Different Scenarios

#### Well-Written Code (`simple_good.go`)
**Prompt**: "Review this code for quality and best practices"
**Expected**: High score (90-100), minimal or no issues

#### Problematic Code (`complex_test.go`)
**Prompt**: "Analyze this code and suggest improvements"
**Expected**: Multiple issues across categories, lower score (60-80)

#### Performance Issues (`performance_issues.go`)
**Prompt**: "Review for performance problems"
**Expected**: String concatenation warnings, structure issues

## Integration Testing

For integration with MCP clients (Claude Desktop, Windsurf, etc.), use these examples to verify the tools work correctly in your development environment.

### Verification Checklist

- [ ] Documentation requests return proper Go docs
- [ ] Code review requests return JSON-formatted analysis
- [ ] Custom guidelines are applied correctly
- [ ] Error handling works for invalid requests
- [ ] Performance hints affect the analysis focus