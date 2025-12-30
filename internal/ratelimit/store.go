package ratelimit

import (
	"sync"
	"time"
)

// Store defines the interface for rate limiting storage backends
type Store interface {
	// Increment increments the counter for a key and returns the new count
	Increment(key string, window time.Duration) (int, error)
	// Get retrieves the current count for a key
	Get(key string) (int, error)
	// Reset resets the counter for a key
	Reset(key string) error
	// Delete removes a key from storage
	Delete(key string) error
}

// MemoryStore implements an in-memory rate limiting store with automatic cleanup
type MemoryStore struct {
	// buckets holds the rate limit buckets indexed by key
	buckets map[string]*Bucket
	// mutex provides thread-safe access
	mutex sync.RWMutex
	// cleanupInterval is how often to check for expired entries
	cleanupInterval time.Duration
	// done is used to stop the cleanup goroutine
	done chan struct{}
}

// NewMemoryStore creates a new in-memory store with periodic cleanup
func NewMemoryStore() *MemoryStore {
	store := &MemoryStore{
		buckets:         make(map[string]*Bucket),
		cleanupInterval: 5 * time.Minute,
		done:            make(chan struct{}),
	}

	// Start cleanup goroutine
	go store.cleanupLoop()

	return store
}

// Increment increments the counter for a key and returns the new count
func (s *MemoryStore) Increment(key string, window time.Duration) (int, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	now := time.Now()
	bucket, exists := s.buckets[key]

	if !exists {
		// Create new bucket
		bucket = &Bucket{
			Count:       1,
			LastUpdate:  now,
			WindowStart: now,
		}
		s.buckets[key] = bucket
		return 1, nil
	}

	// Check if the window has expired
	if now.Sub(bucket.WindowStart) >= window {
		// Reset bucket for new window
		bucket.Count = 1
		bucket.WindowStart = now
	} else {
		// Increment counter within current window
		bucket.Count++
	}

	bucket.LastUpdate = now
	return bucket.Count, nil
}

// Get retrieves the current count for a key
func (s *MemoryStore) Get(key string) (int, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	bucket, exists := s.buckets[key]
	if !exists {
		return 0, nil
	}

	return bucket.Count, nil
}

// Reset resets the counter for a key
func (s *MemoryStore) Reset(key string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.buckets, key)
	return nil
}

// Delete removes a key from storage
func (s *MemoryStore) Delete(key string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.buckets, key)
	return nil
}

// cleanupLoop periodically removes expired entries
func (s *MemoryStore) cleanupLoop() {
	ticker := time.NewTicker(s.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.cleanup()
		case <-s.done:
			return
		}
	}
}

// cleanup removes expired entries older than 10 minutes
func (s *MemoryStore) cleanup() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	now := time.Now()
	expiry := 10 * time.Minute

	for key, bucket := range s.buckets {
		if now.Sub(bucket.LastUpdate) > expiry {
			delete(s.buckets, key)
		}
	}
}

// Stop stops the cleanup goroutine
func (s *MemoryStore) Stop() {
	close(s.done)
}

// NoOpStore implements a no-op store that always allows requests
type NoOpStore struct{}

// NewNoOpStore creates a new no-op store
func NewNoOpStore() *NoOpStore {
	return &NoOpStore{}
}

// Increment returns 1 (always allow)
func (s *NoOpStore) Increment(key string, window time.Duration) (int, error) {
	return 1, nil
}

// Get returns 0
func (s *NoOpStore) Get(key string) (int, error) {
	return 0, nil
}

// Reset does nothing
func (s *NoOpStore) Reset(key string) error {
	return nil
}

// Delete does nothing
func (s *NoOpStore) Delete(key string) error {
	return nil
}
