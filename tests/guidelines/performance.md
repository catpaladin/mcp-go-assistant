# Performance Guidelines for Go Code

## String Operations

- Use strings.Builder for multiple string concatenations
- Avoid string concatenation in loops
- Use bytes.Buffer for binary data manipulation
- Pre-allocate builders with known capacity when possible

## Memory Management

- Reuse objects with sync.Pool for expensive allocations
- Avoid unnecessary allocations in hot paths
- Use slices efficiently - understand capacity vs length
- Be mindful of slice growth patterns

## Goroutines and Concurrency

- Don't create goroutines without bounds
- Use worker pools for managing goroutine lifecycles
- Use channels appropriately - buffered vs unbuffered
- Avoid goroutine leaks by ensuring proper cleanup

## I/O Operations

- Use buffered I/O for small, frequent operations
- Implement timeouts for network operations
- Use connection pooling for database operations
- Consider async I/O patterns where appropriate

## Data Structures

- Choose appropriate data structures for your use case
- Use maps for key-value lookups
- Consider slice vs linked list trade-offs
- Benchmark different approaches for critical paths

## Profiling and Monitoring

- Use go tool pprof for performance analysis
- Profile before optimizing
- Monitor memory usage and garbage collection
- Use benchmarks to validate optimizations