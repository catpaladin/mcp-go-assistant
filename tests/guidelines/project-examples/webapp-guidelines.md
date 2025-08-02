# Web Application Guidelines

Guidelines for building secure, performant web applications in Go.

## Security
- All user inputs must be validated and sanitized
- Use CSRF protection on state-changing operations
- Implement proper session management with secure cookies
- Sanitize data before database operations to prevent SQL injection
- Use Content Security Policy (CSP) headers
- Implement XSS protection measures
- Use HTTPS in production environments
- Implement proper password hashing (bcrypt, scrypt, argon2)

## Authentication & Authorization
- Use secure session management
- Implement proper password policies
- Use multi-factor authentication for sensitive operations
- Implement proper logout functionality
- Use secure cookie attributes (HttpOnly, Secure, SameSite)
- Implement account lockout after failed attempts
- Use JWT tokens with proper expiration

## Input Validation
- Validate all user inputs on both client and server side
- Use allowlists rather than blocklists for validation
- Implement proper file upload validation
- Validate file types, sizes, and content
- Sanitize user-generated content before display
- Use parameterized queries for database operations

## Performance
- Use connection pooling for database operations
- Implement caching for frequently accessed data
- Minimize database N+1 queries
- Use appropriate HTTP caching headers
- Implement CDN for static assets
- Use compression for responses (gzip)
- Optimize database queries and indexes

## Frontend Integration
- API responses must include appropriate CORS headers
- Use consistent date/time formats (RFC3339)
- Implement proper error response formats
- Include API versioning in URLs (/api/v1/)
- Use JSON for API responses unless other formats required
- Implement proper HTTP status codes

## Session Management
- Use secure session storage (Redis, database)
- Implement proper session timeouts
- Use cryptographically secure session IDs
- Implement session invalidation on logout
- Use proper session cookie attributes
- Implement concurrent session limits

## Error Handling
- Never expose internal error details to users
- Log all errors with sufficient context
- Implement user-friendly error messages
- Use structured error responses
- Implement proper error page handling
- Log security-related events

## Logging & Monitoring
- Log all authentication attempts
- Log all administrative actions
- Implement audit trails for sensitive operations
- Use structured logging with consistent fields
- Monitor for suspicious activities
- Implement alerting for critical errors

## Database Security
- Use least privilege principle for database access
- Encrypt sensitive data at rest
- Use parameterized queries exclusively
- Implement proper database backup and recovery
- Use database connection encryption
- Regularly update database software

## File Handling
- Validate file uploads thoroughly
- Store uploaded files outside web root
- Implement virus scanning for uploads
- Use content-type validation
- Implement file size limits
- Generate unique filenames to prevent conflicts

## Content Management
- Implement proper content sanitization
- Use templating engines with auto-escaping
- Validate and sanitize rich text content
- Implement content approval workflows
- Use versioning for content changes
- Implement proper content backup strategies