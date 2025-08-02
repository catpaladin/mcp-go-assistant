# MCP Go Assistant

A Model Context Protocol (MCP) server that provides various Go development tools, including documentation access and code analysis capabilities for LLMs.

## Features

- **Package Documentation**: Get documentation for any Go package
- **Symbol Documentation**: Get specific documentation for functions, types, constants, and variables
- **Code Review**: Analyze Go code and provide improvement suggestions based on best practices
- **Custom Guidelines**: Support for custom coding guidelines via markdown files
- **MCP Protocol**: Full MCP server implementation with stdio transport
- **Error Handling**: Proper error responses for invalid packages or symbols

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

Most MCP clients expect the server to be configured in their settings. The server should be invoked with:

**Command**: `mcp-go-assistant` (or full path to binary)  
**Transport**: `stdio`

#### Claude Desktop Integration

Add to your Claude Desktop configuration (`~/Library/Application Support/Claude/claude_desktop_config.json` on macOS):

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

If using with MCP-compatible VS Code extensions, configure the server path in your extension settings.

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

Once the MCP server is configured with your LLM client (Claude Desktop, Windsurf, etc.), you can interact with it using natural language prompts.

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

"Review this database layer code using our guidelines from db-guidelines.md, specifically focusing on:

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

The code review tool supports custom coding guidelines to enforce your team's specific standards and practices.

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
Functions should be small and focused on a single responsibility.
Always use meaningful variable names that clearly indicate purpose.
Avoid deep nesting by using early returns.
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

The `tests/` directory contains comprehensive examples and guidelines for testing both tools. See [tests/README.md](tests/README.md) for detailed testing instructions.

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
  - `hint` (optional): Specific focus area for the review (e.g., "performance", "security")

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

## Integration Examples

### Claude Desktop

1. Locate your Claude Desktop config file:
   - **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
   - **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`
   - **Linux**: `~/.config/Claude/claude_desktop_config.json`

2. Add the server configuration:

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

3. Restart Claude Desktop

### Windsurf

Windsurf by Codeium supports MCP servers for enhanced AI capabilities:

1. **Via Settings UI**:
   - Open Windsurf
   - Go to **Settings** → **Extensions** → **MCP Servers**
   - Click **Add Server**
   - Configure:
     - **Name**: `Go Documentation`
     - **Command**: `/full/path/to/mcp-go-assistant/bin/mcp-go-assistant`
     - **Transport**: `stdio`
   - Save and restart Windsurf

2. **Via Configuration File** (if supported):

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
         "description": "Go development assistant server"
       }
     }
   }
   ```

3. **Usage in Windsurf**:
   Once configured, you can ask Windsurf questions like:
   - "Show me the documentation for the fmt package"
   - "What does fmt.Printf do?"
   - "Review this Go code for best practices: [paste code]"
   - "Analyze this function for performance issues: [paste code]"

### Other MCP Clients

This server can be integrated with any MCP-compatible client by configuring:

- **Command**: Path to the `mcp-go-assistant` binary
- **Transport**: `stdio`
- **Protocol**: MCP 2024-11-05

### Client-Specific Usage Tips

#### Claude Desktop

After configuration, you can seamlessly ask:

- "Get Go documentation for any package or function"
- "Review my Go code and suggest improvements"
- "Check this code for security issues"

**Example Conversation:**

```
You: "Show me how to use sync.WaitGroup"

Claude: I'll get the documentation for sync.WaitGroup for you.
[Uses go-doc tool to fetch documentation]

Here's how to use sync.WaitGroup...
```

#### Windsurf

Windsurf integrates the tools into its development workflow:

- Code review suggestions appear in the editor
- Documentation is accessible via chat
- Guidelines can be enforced during development

**Example Usage:**

```
You: "Review this function I just wrote and suggest improvements"
[Select code in editor]

Windsurf: I'll analyze your code for best practices.
[Uses code-review tool]

I found several areas for improvement...
```

#### Cursor/VS Code

With MCP extensions, you can:

- Get documentation without leaving the editor
- Receive real-time code review feedback
- Apply custom guidelines to your projects

#### Other AI Development Tools

The server works with any MCP-compatible tool:

- **Continue**: Code review in VS Code
- **Aider**: Documentation and code analysis
- **Custom integrations**: Build your own tools

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

