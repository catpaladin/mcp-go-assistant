# Team Workflow Guide for MCP Go Assistant

This guide shows how teams can integrate the MCP Go Assistant into their development workflow using project-based guidelines.

## Setting Up Team Guidelines

### 1. Create Team Guidelines File

Place a `.mcp-guidelines.md` file in your project root:

```markdown
# [Your Team/Project] Go Guidelines

## Code Standards
- Follow our naming conventions
- Use our error handling patterns
- Implement our logging standards
- Follow our testing requirements

## Architecture Patterns
- Use dependency injection
- Follow clean architecture
- Implement proper interfaces
- Use our approved libraries

## Security Requirements
- Validate all inputs
- Use approved authentication methods
- Follow our data protection standards
- Implement proper audit logging
```

### 2. Common Prompt Patterns for Teams

#### Code Review in Pull Requests
```
"Please review this Go code using our project guidelines from .mcp-guidelines.md:

[paste the code from PR]

Focus on: [specific areas of concern for this PR]"
```

#### Pre-commit Review
```
"Review this new function according to our .mcp-guidelines.md before I commit:

```go
[your new function]
```

Make sure it follows our error handling and testing standards."
```

#### Architecture Review
```
"Review this service implementation using our microservice-guidelines.md:

```go
[service code]
```

Focus on: API design, error handling, and observability requirements."
```

## Integration with Development Tools

### 1. IDE Integration (VS Code/Cursor)

Create a workspace snippet for quick reviews:

```json
{
  "Review with Guidelines": {
    "prefix": "mcpreview",
    "body": [
      "Please review this Go code using our project guidelines from .mcp-guidelines.md:",
      "",
      "```go",
      "$CLIPBOARD",
      "```",
      "",
      "Focus on: ${1:error handling, performance, security}"
    ],
    "description": "Review code with project guidelines"
  }
}
```

### 2. Git Hooks Integration

Create a pre-commit hook script:

```bash
#!/bin/bash
# .git/hooks/pre-commit

echo "Running MCP code review on staged Go files..."

# Get staged Go files
staged_files=$(git diff --cached --name-only --diff-filter=AM | grep '\.go$')

if [ -z "$staged_files" ]; then
    echo "No Go files to review"
    exit 0
fi

for file in $staged_files; do
    echo "Reviewing $file..."
    # Here you would integrate with your MCP client
    # For example, using a CLI tool that interfaces with MCP
    # mcp-review --file "$file" --guidelines ".mcp-guidelines.md"
done
```

### 3. CI/CD Pipeline Integration

Add to your GitHub Actions workflow:

```yaml
name: Code Review
on:
  pull_request:
    paths:
      - '**/*.go'

jobs:
  code-review:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup MCP Go Assistant
        run: |
          # Setup your MCP client
          # This could be a custom action or script
          
      - name: Review Changed Files
        run: |
          # Get changed Go files
          files=$(git diff --name-only ${{ github.event.pull_request.base.sha }} ${{ github.sha }} | grep '\.go$')
          
          for file in $files; do
            echo "Reviewing $file with project guidelines..."
            # Use your MCP integration to review the file
            # Post results as PR comments
          done
```

## Team-Specific Workflows

### 1. Backend Team Workflow

**Guidelines File**: `backend-guidelines.md`
- API design standards
- Database patterns
- Error handling
- Logging requirements

**Common Prompts**:
```
"Review this API handler using our backend-guidelines.md:
[code]
Focus on: API design and error handling"

"Check this database layer against our backend-guidelines.md:
[code]
Focus on: query patterns and transaction handling"
```

### 2. Frontend Team Workflow

**Guidelines File**: `frontend-api-guidelines.md`
- API client patterns
- Error handling for UI
- State management
- Component architecture

**Common Prompts**:
```
"Review this API client code using our frontend-api-guidelines.md:
[code]
Focus on: error handling and state management"
```

### 3. Platform Team Workflow

**Guidelines File**: `platform-guidelines.md`
- Infrastructure code standards
- Deployment patterns
- Monitoring and observability
- Security requirements

**Common Prompts**:
```
"Review this infrastructure code using our platform-guidelines.md:
[code]
Focus on: security and observability"
```

## Guidelines Management

### 1. Guidelines File Organization

For complex projects, organize guidelines into multiple files:

```
docs/
├── guidelines/
│   ├── general.md           # General coding standards
│   ├── api-design.md        # API design guidelines
│   ├── database.md          # Database patterns
│   ├── security.md          # Security requirements
│   ├── testing.md           # Testing standards
│   └── deployment.md        # Deployment guidelines
└── README.md
```

### 2. Versioning Guidelines

Version your guidelines with your project:

```markdown
# Project Guidelines v2.1.0

## Changelog
- v2.1.0: Added new error handling patterns
- v2.0.0: Updated API design standards
- v1.0.0: Initial guidelines

## Current Standards
[guidelines content]
```

### 3. Guidelines Review Process

1. **Proposal**: Team member proposes guideline changes
2. **Review**: Team reviews using MCP assistant
3. **Testing**: Test new guidelines on sample code
4. **Adoption**: Update guidelines file and communicate changes

## Onboarding New Team Members

### 1. Guidelines Introduction

New team members should review the guidelines with MCP:

```
"I'm new to the team. Can you explain our project guidelines from .mcp-guidelines.md and show me examples of code that follows these standards?"
```

### 2. Code Review Training

Use MCP to train new members:

```
"Review this code example using our .mcp-guidelines.md and explain what issues you find and why they matter for our project:

[training code example]"
```

### 3. Guidelines Verification

New members can test their understanding:

```
"I wrote this function following our .mcp-guidelines.md. Does it meet our standards?

[their code]

Please provide specific feedback on adherence to our guidelines."
```

## Best Practices

### 1. Keep Guidelines Current
- Review guidelines quarterly
- Update based on team learnings
- Remove outdated practices

### 2. Make Guidelines Specific
- Provide concrete examples
- Explain the "why" behind rules
- Include anti-patterns to avoid

### 3. Integrate with Tools
- Use MCP in code reviews
- Integrate with CI/CD
- Make it part of the development flow

### 4. Train the Team
- Regular guidelines review sessions
- Share MCP prompt patterns
- Encourage guideline feedback

## Troubleshooting

### Common Issues

1. **Guidelines not being applied**: Ensure file path is correct in prompts
2. **Inconsistent reviews**: Update guidelines to be more specific
3. **Guidelines conflicts**: Prioritize guidelines by importance
4. **Team resistance**: Show value through examples and metrics

### Getting Help

- Review example guidelines in `tests/guidelines/project-examples/`
- Use the prompt templates in `PROMPTS.md`
- Test guidelines with the CLI tools before rolling out