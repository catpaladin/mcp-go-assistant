package health

import (
	"encoding/json"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// Status represents the health status of the server
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
	StatusDegraded  Status = "degraded"
)

// Health represents the overall health of the system
type Health struct {
	Status    Status            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Version   string            `json:"version"`
	Checks    map[string]Check  `json:"checks"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// Check represents a single health check
type Check struct {
	Name      string        `json:"name"`
	Status    Status        `json:"status"`
	Message   string        `json:"message,omitempty"`
	Duration  time.Duration `json:"duration,omitempty"`
	Error     string        `json:"error,omitempty"`
	Timestamp time.Time     `json:"timestamp"`
}

// Checker defines the interface for health checks
type Checker interface {
	Check() Check
}

// HealthChecker performs health checks
type HealthChecker struct {
	checkers map[string]Checker
	mu       sync.RWMutex
	version  string
}

// New creates a new health checker
func New(version string) *HealthChecker {
	return &HealthChecker{
		checkers: make(map[string]Checker),
		version:  version,
	}
}

// RegisterChecker registers a health checker
func (h *HealthChecker) RegisterChecker(name string, checker Checker) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.checkers[name] = checker
}

// Check performs all health checks and returns the overall health
func (h *HealthChecker) Check() Health {
	h.mu.RLock()
	checkers := make(map[string]Checker, len(h.checkers))
	for name, checker := range h.checkers {
		checkers[name] = checker
	}
	h.mu.RUnlock()

	checks := make(map[string]Check)
	overallStatus := StatusHealthy

	for name, checker := range checkers {
		check := checker.Check()
		checks[name] = check

		if check.Status == StatusUnhealthy {
			overallStatus = StatusUnhealthy
		} else if check.Status == StatusDegraded && overallStatus == StatusHealthy {
			overallStatus = StatusDegraded
		}
	}

	// Add built-in checks
	memoryCheck := h.checkMemory()
	checks["memory"] = memoryCheck
	if memoryCheck.Status == StatusUnhealthy {
		overallStatus = StatusUnhealthy
	}

	health := Health{
		Status:    overallStatus,
		Timestamp: time.Now().UTC(),
		Version:   h.version,
		Checks:    checks,
		Metadata: map[string]string{
			"go_version": runtime.Version(),
			"os":         runtime.GOOS,
			"arch":       runtime.GOARCH,
			"goroutines": fmt.Sprintf("%d", runtime.NumGoroutine()),
		},
	}

	return health
}

// checkMemory performs a memory health check
func (h *HealthChecker) checkMemory() Check {
	start := time.Now()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	duration := time.Since(start)

	// Check memory usage (consider unhealthy if > 90% of max memory)
	// For simplicity, we'll use a threshold of 2GB
	const maxMemory = 2 * 1024 * 1024 * 1024 // 2GB
	usedMemory := m.Sys

	status := StatusHealthy
	message := fmt.Sprintf("Memory usage: %d MB", usedMemory/(1024*1024))

	if usedMemory > maxMemory {
		status = StatusUnhealthy
		message = fmt.Sprintf("Memory usage critical: %d MB > %d MB", usedMemory/(1024*1024), maxMemory/(1024*1024))
	} else if usedMemory > maxMemory*3/4 {
		status = StatusDegraded
		message = fmt.Sprintf("Memory usage high: %d MB > %d MB", usedMemory/(1024*1024), (maxMemory*3/4)/(1024*1024))
	}

	return Check{
		Name:      "memory",
		Status:    status,
		Message:   message,
		Duration:  duration,
		Timestamp: time.Now().UTC(),
	}
}

// JSON returns the health status as JSON
func (h *Health) JSON() (string, error) {
	data, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal health status: %w", err)
	}
	return string(data), nil
}
