# MCP Go Assistant

A Model Context Protocol (MCP) server that provides various Go development tools,
including documentation access and code analysis capabilities for LLMs.

## Features

- **Package Documentation**: Get documentation for any Go package
- **Symbol Documentation**: Get specific documentation for functions, types, constants,
  and variables
- **Code Review**: Analyze Go code and provide improvement suggestions based on best
  practices
- **Test Generation**: Generate test scaffolding including interfaces, mocks, and
  table-driven tests
- **Custom Guidelines**: Support for custom coding guidelines via markdown files
- **MCP Protocol**: Full MCP server implementation with stdio transport
- **Error Handling**: Proper error responses for invalid packages or symbols

## Production Features (New!)

The MCP Go Assistant now includes enterprise-grade features:

- **Structured Logging**: JSON and console logging with request tracking, using zerolog
- **Metrics**: Prometheus-compatible metrics for monitoring and observability
- **Configuration Management**: YAML config files with environment variable overrides
- **Health Checks**: System health monitoring with custom check registration
- **Graceful Shutdown**: Proper signal handling and cleanup
- **Request Timeouts**: Configurable timeouts per tool with context propagation
- **Error Classification**: Automatic error categorization for metrics

For detailed production documentation, see
[PRODUCTION_READINESS.md](PRODUCTION_READINESS.md) and
[PRODUCTION_IMPROVEMENTS.md](PRODUCTION_IMPROVEMENTS.md).

## Project Structure

```
mcp-go-assistant/
├── cmd/
│   ├── mcp-go-assistant/   # MCP server binary
│   │   └── main.go
│   ├── client/             # Test client for go-doc
│   │   └── main.go
│   └── review-client/      # Test client for code-review
│       └── main.go
├── internal/
│   ├── godoc/              # Go doc functionality
│   │   └── godoc.go
│   └── codereview/         # Code review functionality
│       ├── types.go
│       ├── analyzer.go
│       ├── guidelines.go
│       └── codereview.go
├── tests/                  # Test files and examples
│   ├── README.md           # Test documentation
│   ├── examples/           # Go code examples
│   │   ├── simple_good.go
│   │   ├── test_code.go
│   │   ├── complex_test.go
│   │   └── performance_issues.go
│   └── guidelines/         # Custom coding guidelines
│       ├── guidelines.md
│       ├── security.md
│       ├── performance.md
│       └── project-examples/   # Example project guidelines
│           ├── .mcp-guidelines.md
│           ├── microservice-guidelines.md
│           ├── webapp-guidelines.md
│           └── library-guidelines.md
├── bin/                    # Built binaries
├── go.mod
├── CLAUDE.md
├── PROMPTS.md              # Quick prompt reference
├── TEAM_WORKFLOW.md        # Team integration guide
└── README.md
```

## Installation

### Prerequisites

- Go 1.24.0 or later
- Git (for cloning the repository)

### Clone and Build

```bash
# Clone the repository
git clone <repository-url>
cd mcp-go-assistant

# Install dependencies
go mod tidy

# Create bin directory
mkdir -p bin

# Build the MCP server
go build -o bin/mcp-go-assistant ./cmd/mcp-go-assistant

# Build the test clients (optional)
go build -o bin/client ./cmd/client
go build -o bin/review-client ./cmd/review-client
```

### Install System-wide (Optional)

```bash
# Install to Go bin directory
go install ./cmd/mcp-go-assistant

# Or copy to system PATH
sudo cp bin/mcp-go-assistant /usr/local/bin/
```

## Usage

### Running the MCP Server

The server communicates via stdio using the MCP protocol. It can be run in several ways:

#### Direct Execution

```bash
# From project directory
./bin/mcp-go-assistant

# If installed system-wide
mcp-go-assistant
```

#### With MCP Clients

Most MCP clients expect the server to be configured in their settings. The server should
be invoked with:

**Command**: `mcp-go-assistant` (or full path to binary)  
**Transport**: `stdio`

#### Claude Desktop Integration

Add to your Claude Desktop configuration
(`~/Library/Application Support/Claude/claude_desktop_config.json` on macOS):

```json
{
  "mcpServers": {
    "go-assistant": {
      "command": "/path/to/mcp-go-assistant/bin/mcp-go-assistant"
    }
  }
}
```

#### VS Code / Cursor Integration

If using with MCP-compatible VS Code extensions, configure the server path in your
extension settings.

#### Windsurf Integration

Windsurf (by Codeium) supports MCP servers through its configuration:

1. Open Windsurf settings or preferences
2. Navigate to the MCP/Extensions section
3. Add a new MCP server with:
   - **Name**: `Go Assistant`
   - **Command**: `/full/path/to/mcp-go-assistant/bin/mcp-go-assistant`
   - **Transport**: `stdio`

Alternatively, if Windsurf uses a configuration file (similar to Claude Desktop), add:

```json
{
  "mcpServers": {
    "go-assistant": {
      "command": "/full/path/to/mcp-go-assistant/bin/mcp-go-assistant",
      "args": []
    }
  }
}
```

## Tools

The MCP Go Assistant server provides three powerful tools for Go development. Each tool
is designed to help you work more efficiently with Go code.

### Tool Overview

| Tool            | Description                                         | Use Cases                                                                                       |
| --------------- | --------------------------------------------------- | ----------------------------------------------------------------------------------------------- |
| **go-doc**      | Get documentation for Go packages and symbols       | Looking up package documentation, understanding function signatures, exploring standard library |
| **code-review** | Analyze Go code for best practices and improvements | Code quality checks, performance analysis, security reviews, adherence to coding guidelines     |
| **test-gen**    | Generate test scaffolding for Go code               | Creating test files, generating interface mocks, building table-driven tests                    |

---

### go-doc Tool

**Tool Name**: `go-doc`

**Description**: Get Go documentation for packages and symbols using the standard
`go doc` command. This tool provides access to comprehensive documentation for the
standard library, third-party packages, and your own code.

#### Parameters

| Parameter      | Type   | Required | Description                                                                                  |
| -------------- | ------ | -------- | -------------------------------------------------------------------------------------------- |
| `package_path` | string | Yes      | The Go package path to query (e.g., `"fmt"`, `"net/http"`, `"github.com/user/repo/package"`) |
| `symbol_name`  | string | No       | Specific symbol within the package (e.g., `"Printf"`, `"Server"`, `"Context"`)               |
| `working_dir`  | string | No       | Optional working directory with go.mod file for external package access                      |

#### Usage Examples

**Example 1: Get package documentation**

```
"Show me the documentation for fmt package"
"What's in the net/http package?"
```

**Example 2: Get specific symbol documentation**

```
"Show me the documentation for fmt.Printf"
"Get documentation for net/http.Server"
"What methods are available on context.Context?"
"Explain sync.WaitGroup.Add"
```

**Example 3: Explore third-party packages**

```
"Show me the documentation for github.com/gin-gonic/gin"
"What does gorm.DB do?"
```

#### MCP Request Example

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "go-doc",
    "arguments": {
      "package_path": "fmt",
      "symbol_name": "Printf"
    }
  }
}
```

#### Typical Responses

The tool returns documentation including:

- Package overview
- Function signatures
- Type definitions
- Method documentation
- Examples (when available)

---

### code-review Tool

**Tool Name**: `code-review`

**Description**: Analyze Go code and provide improvement suggestions based on best
practices, security considerations, performance optimizations, and custom coding
guidelines.

#### Parameters

| Parameter            | Type   | Required | Description                                                                  |
| -------------------- | ------ | -------- | ---------------------------------------------------------------------------- |
| `go_code`            | string | Yes      | The Go code content to analyze                                               |
| `guidelines_file`    | string | No       | Path to markdown file with coding guidelines                                 |
| `guidelines_content` | string | No       | Markdown content with coding guidelines (alternative to file)                |
| `hint`               | string | No       | Specific focus area (e.g., `"performance"`, `"security"`, `"documentation"`) |

#### Usage Examples

**Example 1: General best practices review**

````
"Review this Go code for best practices:
```go
func main() {
    var name string = "world"
    fmt.Printf("Hello %s\n", name)
}
```"
````

**Example 2: Performance-focused review**

````
"Analyze this code for performance issues:
```go
func buildString(items []string) string {
    result := ""
    for _, item := range items {
        result += item + ","
    }
    return result
}
```"
````

**Example 3: Security-focused review**

````
"Check this code for security vulnerabilities:
```go
func handleInput(input string) {
    cmd := exec.Command("sh", "-c", input)
    cmd.Run()
}
```"
````

**Example 4: Documentation review**

````
"Review this code focusing on documentation:
```go
func process(data string) error {
    if data == "" {
        return errors.New("empty")
    }
    return nil
}
```"
````

**Example 5: Custom guidelines review**

````
"Review this code using these guidelines:
- Avoid using panic in production code
- Always validate input parameters
- Use context.Context for timeouts

```go
func processData(data string) {
    if data == "" {
        panic("empty data")
    }
    // process data...
}
```"
````

#### MCP Request Examples

**Basic Code Review:**

```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/call",
  "params": {
    "name": "code-review",
    "arguments": {
      "go_code": "package main\n\nfunc ExportedFunction() {\n\t// Missing documentation\n}",
      "hint": "focus on documentation and best practices"
    }
  }
}
```

**Code Review with Guidelines:**

```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "tools/call",
  "params": {
    "name": "code-review",
    "arguments": {
      "go_code": "package main\n\nfunc main() {\n\tpanic(\"error\")\n}",
      "guidelines_content": "- Avoid using panic in production code\n- Always handle errors gracefully",
      "hint": "focus on error handling"
    }
  }
}
```

**Code Review with Guidelines File:**

```json
{
  "jsonrpc": "2.0",
  "id": 4,
  "method": "tools/call",
  "params": {
    "name": "code-review",
    "arguments": {
      "go_code": "package main\n\nfunc process() error {\n\treturn nil\n}",
      "guidelines_file": "/path/to/guidelines.md",
      "hint": "focus on security"
    }
  }
}
```

#### Review Categories

The tool analyzes code in these areas:

- **Code Quality**: Naming conventions, code organization, complexity
- **Error Handling**: Proper error wrapping, panic avoidance, error messages
- **Documentation**: Function/type documentation, examples, clarity
- **Security**: Input validation, SQL injection prevention, sensitive data handling
- **Performance**: Memory usage, goroutine management, algorithmic efficiency
- **Testing**: Test coverage, test quality, edge cases
- **Concurrency**: Race conditions, mutex usage, channel patterns

---

### test-gen Tool

**Tool Name**: `test-gen`

**Description**: Generate Go test scaffolding including interfaces, mocks, and
table-driven tests. This tool helps you quickly create test files with proper structure
and best practices.

#### Parameters

| Parameter      | Type   | Required | Description                                                                                                                                   |
| -------------- | ------ | -------- | --------------------------------------------------------------------------------------------------------------------------------------------- |
| `go_code`      | string | Yes      | The Go code to generate tests for                                                                                                             |
| `focus`        | string | No       | Test generation focus: `"interfaces"` (extract interfaces and generate mocks), `"table"` (table-driven tests), or `"unit"` (basic unit tests) |
| `package_name` | string | No       | Package name for generated tests (defaults to `"package_test"`)                                                                               |

#### Usage Examples

**Example 1: Generate basic unit tests**

````
"Generate tests for this Go function:
```go
func Add(a, b int) int {
    return a + b
}
```"
````

**Example 2: Generate interface mocks**

````
"Create interface mocks for this code:
```go
type UserService interface {
    GetUser(id int) (*User, error)
    CreateUser(user *User) error
}

type Service struct {
    userRepo UserService
}
```"
````

**Example 3: Generate table-driven tests**

````
"Generate table-driven tests for:
```go
func ParseSize(s string) (int, error) {
    var size int
    _, err := fmt.Sscanf(s, "%d", &size)
    return size, err
}
```"
````

#### MCP Request Examples

**Basic Unit Tests:**

```json
{
  "jsonrpc": "2.0",
  "id": 5,
  "method": "tools/call",
  "params": {
    "name": "test-gen",
    "arguments": {
      "go_code": "package main\n\nfunc Add(a, b int) int {\n\treturn a + b\n}",
      "focus": "unit",
      "package_name": "mypackage_test"
    }
  }
}
```

**Interface Extraction and Mocks:**

```json
{
  "jsonrpc": "2.0",
  "id": 6,
  "method": "tools/call",
  "params": {
    "name": "test-gen",
    "arguments": {
      "go_code": "package main\n\ntype Database interface {\n\tQuery(query string) (string, error)\n}\n\ntype Service struct {\n\tdb Database\n}",
      "focus": "interfaces"
    }
  }
}
```

**Table-Driven Tests:**

```json
{
  "jsonrpc": "2.0",
  "id": 7,
  "method": "tools/call",
  "params": {
    "name": "test-gen",
    "arguments": {
      "go_code": "package main\n\nfunc ValidateEmail(email string) bool {\n\treturn len(email) > 5\n}",
      "focus": "table"
    }
  }
}
```

#### Generated Test Features

The `test-gen` tool creates:

- **Unit Tests**: Basic test structure with `TestFunctionName` format
- **Interface Tests**: Extracted interfaces with mock implementations using `gomock` or
  `testify`
- **Table-Driven Tests**: Structured test cases with input/output pairs
- **Test Helpers**: Common test utilities and setup functions
- **Edge Cases**: Tests for boundary conditions and error scenarios

---

## Integration Examples

### Claude Desktop Integration

**Step 1: Locate Configuration File**

Find your Claude Desktop config file:

- **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`
- **Linux**: `~/.config/Claude/claude_desktop_config.json`

**Step 2: Add Server Configuration**

```json
{
  "mcpServers": {
    "go-assistant": {
      "command": "/full/path/to/mcp-go-assistant/bin/mcp-go-assistant",
      "args": []
    }
  }
}
```

**Step 3: Restart Claude Desktop**

After adding the configuration, restart Claude Desktop to load the MCP server.

**Step 4: Use the Tools**

Once configured, you can ask Claude questions like:

````
User: "Show me the documentation for fmt.Printf"
Claude: [Uses go-doc tool] Returns documentation for fmt.Printf

User: "Review this code for performance issues:"
```go
func buildString(items []string) string {
    result := ""
    for _, item := range items {
        result += item + ","
    }
    return result
}
````

Claude: [Uses code-review tool] Provides performance analysis and suggestions

User: "Generate unit tests for this function:"

```go
func CalculateTax(price float64, rate float64) float64 {
    return price * rate / 100
}
```

Claude: [Uses test-gen tool] Generates comprehensive unit tests

````

### Windsurf Integration

**Step 1: Via Settings UI**

1. Open Windsurf
2. Go to **Settings** → **Extensions** → **MCP Servers**
3. Click **Add Server**
4. Configure:
   - **Name**: `Go Assistant`
   - **Command**: `/full/path/to/mcp-go-assistant/bin/mcp-go-assistant`
   - **Transport**: `stdio`
5. Save and restart Windsurf

**Step 2: Via Configuration File** (if supported)

Locate Windsurf's config file:
- **macOS**: `~/Library/Application Support/Windsurf/mcp_config.json`
- **Windows**: `%APPDATA%\Windsurf\mcp_config.json`
- **Linux**: `~/.config/windsurf/mcp_config.json`

Add the server configuration:

```json
{
  "mcpServers": {
    "go-assistant": {
      "command": "/full/path/to/mcp-go-assistant/bin/mcp-go-assistant",
      "args": [],
      "description": "Go development assistant with documentation, code review, and test generation"
    }
  }
}
````

**Step 3: Usage in Windsurf**

Once configured, Windsurf can help you:

- **Documentation Queries**:

  ```
  "Show me the documentation for net/http.Server"
  "What does context.WithTimeout do?"
  ```

- **Code Review**:

  ```
  "Review this handler for best practices: [select code]"
  "Check this function for security issues"
  ```

- **Test Generation**:
  ```
  "Generate tests for this service: [select code]"
  "Create mocks for this interface"
  ```

### VS Code / Cursor Integration

With MCP-compatible extensions (e.g., Continue, MCP extensions):

**Step 1: Install MCP Extension**

Search for and install an MCP-compatible extension in the VS Code Marketplace.

**Step 2: Configure MCP Server**

Add to your VS Code settings or the extension's configuration file:

```json
{
  "mcp.servers": {
    "go-assistant": {
      "command": "/full/path/to/mcp-go-assistant/bin/mcp-go-assistant",
      "args": []
    }
  }
}
```

**Step 3: Reload and Use**

Reload VS Code and start using the tools via:

- Chat interface
- Command palette commands
- Context menu options

### Other MCP Clients

The server works with any MCP-compatible client. Configure:

- **Command**: Path to the `mcp-go-assistant` binary
- **Transport**: `stdio`
- **Protocol**: MCP 2024-11-05

### Workflow Integration

**Pre-commit Code Review:**

```bash
#!/bin/bash
# .git/hooks/pre-commit

echo "Running MCP code review..."
for file in $(git diff --cached --name-only --diff-filter=AM | grep '\.go$'); do
    echo "Reviewing $file..."
    # Use your MCP client to review the file with project guidelines
done
```

**CI/CD Integration:**

```yaml
# Example GitHub Actions workflow
- name: Run Code Review
  run: |
    echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"code-review","arguments":{"go_code":"$(cat main.go)"}}}' | \
    ./bin/mcp-go-assistant
```

---

### Testing with CLI Client

For testing and development, use the included CLI client:

```bash
# Get package documentation
./bin/client fmt

# Get specific symbol documentation
./bin/client fmt Printf
./bin/client net/http Server
./bin/client encoding/json Marshal

# Test code review functionality
./bin/review-client tests/examples/test_code.go
./bin/review-client tests/examples/complex_test.go tests/guidelines/guidelines.md "focus on performance"
```

## Usage with LLMs

Once the MCP server is configured with your LLM client (Claude Desktop, Windsurf, etc.),
you can interact with it using natural language prompts.

### Go Documentation Queries

Ask for Go package and symbol documentation using natural language:

```
"Show me the documentation for the fmt package"
"What does fmt.Printf do?"
"Get documentation for net/http.Server"
"How do I use encoding/json.Marshal?"
"What methods are available on context.Context?"
"Show me the documentation for sync.WaitGroup.Add"
```

### Code Review Requests

Ask for code analysis and improvement suggestions:

````
"Review this Go code for best practices:
```go
func main() {
    var name string = "world"
    fmt.Printf("Hello %s\n", name)
}
```"

"Analyze this code for performance issues:
```go
func buildString(items []string) string {
    result := ""
    for _, item := range items {
        result += item + ","
    }
    return result
}
```"

"Review this code focusing on security:
```go
func handleInput(input string) {
    cmd := exec.Command("sh", "-c", input)
    cmd.Run()
}
```"
````

### Using Custom Guidelines

You can provide custom coding guidelines in your prompts:

````
"Review this Go code using these guidelines:
- Avoid using panic in production code
- Always validate input parameters
- Use context.Context for timeouts

```go
func processData(data string) {
    if data == "" {
        panic("empty data")
    }
    // process data...
}
```"

"Analyze this code for performance with focus on:
- String concatenation efficiency
- Memory allocations
- Algorithm complexity

```go
[your code here]
```"
````

### Advanced Usage Examples

**Documentation with Context:**

```
"I'm working with HTTP servers in Go. Show me the documentation for net/http.Server and explain the key fields I should configure."
```

**Focused Code Review:**

````
"Review this Go function for maintainability and suggest improvements:

```go
func ProcessUserData(id int, name string, email string, age int, active bool, role string, department string) error {
    if id <= 0 {
        if name == "" {
            if email == "" {
                return errors.New("invalid input")
            }
        }
    }
    // ... complex logic
}
````

Focus on: function complexity, parameter management, and error handling."

````

## Project-Based Guidelines

For teams and projects, you can maintain coding guidelines in markdown files and instruct the MCP to use them for consistent code reviews.

### Setting Up Project Guidelines

#### 1. Create Guidelines File in Your Project

Create a `.mcp-guidelines.md` file in your project root:

```markdown
# Project Coding Guidelines

## General Rules
- All functions must have documentation comments
- Use meaningful variable names (no single letters except for loop counters)
- Maximum function length: 30 lines
- Maximum function parameters: 5

## Error Handling
- Always return errors, never panic in production code
- Wrap errors with context using fmt.Errorf("operation failed: %w", err)
- Log errors at the boundary where they're handled
- Use structured logging with consistent fields

## Performance
- Use strings.Builder for string concatenation in loops
- Prefer sync.Pool for expensive object reuse
- Always profile before optimizing
- Use context.Context for cancellation and timeouts

## Security
- Validate all user inputs at API boundaries
- Use parameterized queries for database operations
- Never log sensitive information (passwords, tokens, PII)
- Implement proper authentication and authorization

## Architecture
- Follow dependency injection pattern
- Use interfaces for external dependencies
- Keep business logic separate from infrastructure concerns
- Implement proper separation of concerns
````

#### 2. Instruct the MCP to Use Your Guidelines

When asking for code reviews, reference your guidelines file:

````
"Please review this Go code using our project guidelines from .mcp-guidelines.md:

```go
func processUser(id string, name string, email string, age int, active bool) error {
    if id == "" {
        panic("user ID cannot be empty")
    }

    result := ""
    for i := 0; i < 100; i++ {
        result += "processing step " + string(i) + ", "
    }

    log.Printf("Processing user: %s", email)
    return nil
}
````

Focus on: error handling, performance, and security compliance."

````

#### 3. Alternative Guidelines File Names

You can use any of these common names:
- `.mcp-guidelines.md`
- `CODING_GUIDELINES.md`
- `docs/guidelines.md`
- `.github/CODING_STANDARDS.md`
- `CONTRIBUTING.md` (if it contains coding standards)

### Team Workflow Examples

#### Microservice Guidelines

**File: `microservice-guidelines.md`**
```markdown
# Microservice Development Guidelines

## API Design
- All endpoints must return consistent error formats
- Use OpenAPI/Swagger documentation
- Implement proper HTTP status codes
- Include request/response examples

## Observability
- Add tracing to all external calls
- Use structured logging with correlation IDs
- Implement health check endpoints
- Add metrics for business operations

## Resilience
- Implement circuit breakers for external dependencies
- Use exponential backoff for retries
- Set appropriate timeouts for all operations
- Handle graceful shutdown properly
````

**Usage:**

````
"Review this microservice handler using our microservice-guidelines.md standards:

```go
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
    id := r.URL.Query().Get("id")
    user, err := h.db.GetUser(id)
    if err != nil {
        http.Error(w, "error", 500)
        return
    }
    json.NewEncoder(w).Encode(user)
}
```"
````

#### Web Application Guidelines

**File: `webapp-guidelines.md`**

```markdown
# Web Application Guidelines

## Security

- All user inputs must be validated and sanitized
- Use CSRF protection on state-changing operations
- Implement proper session management
- Sanitize data before database operations

## Performance

- Use connection pooling for database operations
- Implement caching for frequently accessed data
- Minimize database N+1 queries
- Use appropriate HTTP caching headers

## Frontend Integration

- API responses must include CORS headers
- Use consistent date/time formats (RFC3339)
- Implement proper error response formats
- Include API versioning in URLs
```

#### Library Development Guidelines

**File: `library-guidelines.md`**

```markdown
# Library Development Guidelines

## Public API

- All exported functions and types must have comprehensive documentation
- Use semantic versioning for releases
- Maintain backward compatibility within major versions
- Provide usage examples in documentation

## Error Handling

- Define custom error types for different error categories
- Provide error wrapping for context
- Include helpful error messages for debugging
- Document all possible error conditions

## Testing

- Maintain minimum 80% test coverage
- Write table-driven tests where appropriate
- Include benchmark tests for performance-critical code
- Test all public API functions
```

### Advanced Project Configurations

#### Multiple Guidelines Files

For complex projects, you can reference multiple guideline files:

````
"Review this code using our project standards:
- General guidelines from .mcp-guidelines.md
- Security standards from docs/security-guidelines.md
- Performance requirements from docs/performance-guidelines.md

```go
[your code here]
````

Focus on security and performance aspects."

```

#### Conditional Guidelines

Reference specific sections based on code type:

```

"Review this database layer code using our guidelines from db-guidelines.md,
specifically focusing on:

- Connection management standards
- Query optimization rules
- Error handling patterns
- Transaction management

````go
[your database code here]
```"
````

For complete team integration instructions, see [TEAM_WORKFLOW.md](TEAM_WORKFLOW.md).

### Integration with Development Workflow

#### Pre-commit Hooks

Create a script that uses the MCP for automated code review:

```bash
#!/bin/bash
# .git/hooks/pre-commit

echo "Running MCP code review..."
for file in $(git diff --cached --name-only --diff-filter=AM | grep '\.go$'); do
    echo "Reviewing $file..."
    # Use your MCP client to review the file with project guidelines
done
```

## Custom Guidelines

The code review tool supports custom coding guidelines to enforce your team's specific
standards and practices.

### Three Ways to Provide Guidelines

#### 1. Inline Guidelines in Prompts (Recommended for LLM usage)

Include guidelines directly in your prompt:

````
"Review this Go code following these guidelines:
- Functions should not exceed 50 lines
- Always use error wrapping with fmt.Errorf
- Prefer dependency injection over global variables
- Use meaningful variable names (no single letters except for loops)

```go
func p(d string) error {
    if d == "" {
        return errors.New("bad input")
    }
    return nil
}
```"
````

#### 2. Guidelines Files (For CLI testing)

Create markdown files with your guidelines:

**Example: `my-guidelines.md`**

```markdown
# Team Go Guidelines

## Error Handling

- Always wrap errors with context
- Use structured logging for errors
- Never ignore errors silently

## Performance

- Use sync.Pool for expensive object reuse
- Avoid string concatenation in loops
- Profile before optimizing

## Security

- Validate all inputs at boundaries
- Use parameterized queries
- Never log sensitive data
```

Use with CLI client:

```bash
./bin/review-client mycode.go my-guidelines.md "focus on error handling"
```

#### 3. Built-in Guideline Categories

Reference existing guidelines from the `tests/guidelines/` directory:

```bash
# Security-focused review
./bin/review-client mycode.go tests/guidelines/security.md "focus on security"

# Performance-focused review
./bin/review-client mycode.go tests/guidelines/performance.md "focus on performance"

# General best practices
./bin/review-client mycode.go tests/guidelines/guidelines.md
```

### Guideline Format

Guidelines can be written in several formats:

**Bullet Points:**

```markdown
- Use gofmt to format all code
- Document all exported functions
- Handle errors explicitly
```

**Numbered Lists:**

```markdown
1. Validate input parameters
2. Use context for cancellation
3. Prefer interfaces for testability
```

**Natural Language:**

```markdown
Functions should be small and focused on a single responsibility. Always use meaningful
variable names that clearly indicate purpose. Avoid deep nesting by using early returns.
```

### Advanced Guidelines Examples

**Team-Specific Standards:**

````
"Review this code according to our team standards:
- All HTTP handlers must use middleware for auth
- Database operations must include context with timeout
- Errors must be logged with structured fields (user_id, action, etc.)
- No direct SQL strings - use only prepared statements or query builders

```go
func GetUser(w http.ResponseWriter, r *http.Request) {
    id := r.URL.Query().Get("id")
    row := db.QueryRow("SELECT * FROM users WHERE id = " + id)
    // ... rest of handler
}
```"
````

**Architecture Guidelines:**

````
"Analyze this code for architecture compliance:
- Follow hexagonal architecture pattern
- Domain logic must not import infrastructure packages
- All external dependencies must be behind interfaces
- Use dependency injection containers

```go
package user

import "database/sql"

type Service struct {
    db *sql.DB
}

func (s *Service) CreateUser(name string) error {
    _, err := s.db.Exec("INSERT INTO users (name) VALUES (?)", name)
    return err
}
```"
````

## Testing

The `tests/` directory contains comprehensive examples and guidelines for testing both
tools. See [tests/README.md](tests/README.md) for detailed testing instructions.

### Quick Test Examples

```bash
# Test go-doc tool
./bin/client fmt
./bin/client fmt Printf

# Test code review with different examples
./bin/review-client tests/examples/simple_good.go          # Should score 90-100
./bin/review-client tests/examples/test_code.go            # Missing docs
./bin/review-client tests/examples/complex_test.go         # Multiple issues
./bin/review-client tests/examples/performance_issues.go   # Performance problems

# Test with custom guidelines
./bin/review-client tests/examples/complex_test.go tests/guidelines/security.md "focus on security"
./bin/review-client tests/examples/performance_issues.go tests/guidelines/performance.md "focus on performance"
```

### MCP Tools

The server exposes two tools:

#### 1. go-doc Tool

- **Name**: `go-doc`
- **Description**: Get Go documentation for packages and symbols
- **Parameters**:
  - `package_path` (required): The Go package path to query
  - `symbol_name` (optional): Specific symbol within the package

#### 2. code-review Tool

- **Name**: `code-review`
- **Description**: Analyze Go code and provide improvement suggestions
- **Parameters**:
  - `go_code` (required): The Go code content to analyze
  - `guidelines_file` (optional): Path to markdown file with coding guidelines
  - `guidelines_content` (optional): Markdown content with coding guidelines
  - `hint` (optional): Specific focus area for the review (e.g., "performance",
    "security")

### Example MCP Requests

#### Go Documentation Request

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "go-doc",
    "arguments": {
      "package_path": "fmt",
      "symbol_name": "Printf"
    }
  }
}
```

#### Code Review Request

```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/call",
  "params": {
    "name": "code-review",
    "arguments": {
      "go_code": "package main\n\nfunc ExportedFunction() {\n\t// Missing documentation\n}",
      "hint": "focus on documentation and best practices"
    }
  }
}
```

#### Code Review with Guidelines

```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "tools/call",
  "params": {
    "name": "code-review",
    "arguments": {
      "go_code": "package main\n\nfunc main() {\n\tpanic(\"error\")\n}",
      "guidelines_content": "- Avoid using panic in production code\n- Always handle errors gracefully",
      "hint": "focus on error handling"
    }
  }
}
```

## Troubleshooting

### Common Issues

1. **"command not found"**: Ensure the binary path is correct and executable
2. **"package not found"**: Make sure Go is installed and GOPATH/modules are accessible
3. **Connection issues**: Verify the MCP client is configured for stdio transport

### Debug Mode

To debug server communications, you can manually test with:

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0.0"}}}' | ./bin/mcp-go-assistant
```

## Requirements

- Go 1.24.0 or later
- Access to Go documentation (standard library and installed packages)
- MCP-compatible client application
