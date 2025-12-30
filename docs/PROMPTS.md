# Quick Prompt Reference

## Go Documentation

### Basic Documentation
```
"Show me the documentation for [package]"
"What does [package.function] do?"
"How do I use [package.type]?"
```

### Examples
```
"Show me the documentation for fmt"
"What does fmt.Printf do?"
"How do I use sync.WaitGroup?"
"Get documentation for net/http.Server"
"What methods are available on context.Context?"
```

## Code Review

### Basic Review
```
"Review this Go code for best practices:
```go
[your code here]
```"
```

### Focused Reviews
```
"Analyze this code for performance issues:"
"Review this code for security vulnerabilities:"
"Check this code for maintainability:"
"Review focusing on error handling:"
```

### With Custom Guidelines

#### Inline Guidelines
```
"Review this Go code following these guidelines:
- [guideline 1]
- [guideline 2]
- [guideline 3]

```go
[your code here]
```"
```

#### Reference Guidelines Files
```
"Please review this Go code using our project guidelines from .mcp-guidelines.md:

```go
[your code here]
```"

"Review this microservice code using our microservice-guidelines.md standards:

```go
[your code here]
```"

"Analyze this library code according to our library-guidelines.md:

```go
[your code here]
```"
```

## Advanced Examples

### Performance Analysis
```
"Analyze this function for performance bottlenecks and suggest optimizations:

```go
func buildString(items []string) string {
    result := ""
    for _, item := range items {
        result += item + ","
    }
    return result
}
```"
```

### Security Review
```
"Review this HTTP handler for security issues:

```go
func handleInput(w http.ResponseWriter, r *http.Request) {
    input := r.URL.Query().Get("cmd")
    cmd := exec.Command("sh", "-c", input)
    output, _ := cmd.Output()
    w.Write(output)
}
```"
```

### Architecture Review
```
"Analyze this code for adherence to clean architecture principles:
- Domain logic should be independent of frameworks
- Use dependency inversion
- Separate concerns clearly

```go
[your code here]
```"
```

## Project-Based Guidelines

### Using Project Guidelines Files

#### General Project Guidelines
```
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
```

Focus on: error handling, performance, and security compliance."
```

#### Microservice Guidelines
```
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
```

Focus on: API design, error handling, and observability."
```

#### Web Application Guidelines
```
"Review this web handler using our webapp-guidelines.md standards:

```go
func loginHandler(w http.ResponseWriter, r *http.Request) {
    username := r.FormValue("username")
    password := r.FormValue("password")
    
    if username == "admin" && password == "admin" {
        http.SetCookie(w, &http.Cookie{
            Name:  "session",
            Value: "user123",
        })
        w.Write([]byte("Login successful"))
    }
}
```

Focus on: security, input validation, and session management."
```

#### Library Guidelines
```
"Review this library code using our library-guidelines.md standards:

```go
func ProcessData(data string) string {
    return strings.ToUpper(data)
}

func ProcessDataAdvanced(data string, options map[string]interface{}) (string, error) {
    // complex processing
    return result, nil
}
```

Focus on: API design, documentation, and error handling."
```

### Multiple Guidelines Files
```
"Review this database layer code using our project standards:
- General guidelines from .mcp-guidelines.md
- Database standards from db-guidelines.md
- Security requirements from security-guidelines.md

```go
func (r *UserRepository) GetUser(ctx context.Context, id string) (*User, error) {
    query := "SELECT * FROM users WHERE id = " + id
    row := r.db.QueryRowContext(ctx, query)
    var user User
    err := row.Scan(&user.ID, &user.Name, &user.Email)
    return &user, err
}
```

Focus on: security, database best practices, and error handling."
```

## Team-Specific Guidelines

### Error Handling Standards
```
"Review this code according to our error handling standards from error-guidelines.md:

```go
[your code here]
```"
```

### API Design Guidelines
```
"Review this API handler using our api-guidelines.md standards:

```go
[your code here]
```"
```

## Expected Responses

### Documentation Queries
- Exact Go documentation text
- Usage examples
- Parameter explanations
- Return value descriptions

### Code Reviews
- JSON-formatted analysis
- Issues categorized by type
- Severity levels (low, medium, high, critical)
- Specific improvement suggestions
- Overall quality score (0-100)
- Code metrics (complexity, line count, etc.)

## Tips for Better Results

1. **Be Specific**: Instead of "review my code", say "review for performance issues"
2. **Provide Context**: Mention what the code is supposed to do
3. **Use Guidelines**: Include your team's specific standards
4. **Ask Follow-ups**: Request clarification on suggestions
5. **Iterate**: Apply suggestions and ask for re-review