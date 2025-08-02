# Security Guidelines for Go Code

## Input Validation

- Always validate all user inputs
- Use parameterized queries for database operations
- Sanitize data before processing
- Check for buffer overflows and boundary conditions

## Authentication & Authorization

- Never store passwords in plain text
- Use strong cryptographic libraries
- Implement proper session management
- Validate user permissions before operations

## Data Protection

- Never log sensitive information like passwords or tokens
- Use TLS for all network communications
- Encrypt sensitive data at rest
- Implement proper key management

## Error Handling

- Don't expose internal system details in error messages
- Log security events appropriately
- Handle errors gracefully without revealing system internals
- Use structured logging for security events

## Dependencies

- Regularly update dependencies to patch security vulnerabilities
- Use dependency scanning tools
- Pin dependency versions in production
- Review third-party code before inclusion

## Unsafe Operations

- Avoid using the unsafe package unless absolutely necessary
- Review all unsafe operations carefully
- Document why unsafe operations are needed
- Consider safer alternatives first