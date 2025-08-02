# Go Code Guidelines

## Best Practices

- Avoid using panic in production code
- Always validate input parameters
- Use context.Context for timeouts and cancellation
- Prefer composition over inheritance
- Write meaningful commit messages

## Performance Guidelines

1. Use strings.Builder for string concatenation
2. Avoid unnecessary allocations in loops
3. Profile before optimizing
4. Use sync.Pool for expensive objects

## Security Rules

- Never log sensitive information
- Validate all user inputs
- Use HTTPS for all communications
- Keep dependencies up to date