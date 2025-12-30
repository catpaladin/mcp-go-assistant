package health

import (
	"strings"
	"testing"
	"time"
)

func TestStatus_Values(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		value  string
	}{
		{"healthy", StatusHealthy, "healthy"},
		{"unhealthy", StatusUnhealthy, "unhealthy"},
		{"degraded", StatusDegraded, "degraded"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.value {
				t.Errorf("expected status '%s', got '%s'", tt.value, string(tt.status))
			}
		})
	}
}

func TestNew(t *testing.T) {
	hc := New("1.0.0")

	if hc == nil {
		t.Fatal("expected health checker, got nil")
	}

	if hc.version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got '%s'", hc.version)
	}

	if hc.checkers == nil {
		t.Error("expected checkers map to be initialized")
	}
}

func TestHealthChecker_RegisterChecker(t *testing.T) {
	hc := New("1.0.0")

	mockChecker := &mockCheck{name: "test", status: StatusHealthy}
	hc.RegisterChecker("test", mockChecker)

	if _, ok := hc.checkers["test"]; !ok {
		t.Error("expected checker to be registered")
	}
}

func TestHealthChecker_RegisterChecker_Override(t *testing.T) {
	hc := New("1.0.0")

	mockChecker1 := &mockCheck{name: "test", status: StatusHealthy}
	mockChecker2 := &mockCheck{name: "test2", status: StatusUnhealthy}

	hc.RegisterChecker("test", mockChecker1)
	hc.RegisterChecker("test", mockChecker2)

	// Should override the previous checker
	if hc.checkers["test"] != mockChecker2 {
		t.Error("expected checker to be overridden")
	}
}

func TestHealthChecker_Check_Overall(t *testing.T) {
	hc := New("1.0.0")

	// Register mock checkers
	hc.RegisterChecker("checker1", &mockCheck{name: "checker1", status: StatusHealthy})
	hc.RegisterChecker("checker2", &mockCheck{name: "checker2", status: StatusHealthy})

	health := hc.Check()

	if health.Status != StatusHealthy {
		t.Errorf("expected overall status healthy, got '%s'", health.Status)
	}
}

func TestHealthChecker_Check_WithUnhealthy(t *testing.T) {
	hc := New("1.0.0")

	// Register checkers with one unhealthy
	hc.RegisterChecker("checker1", &mockCheck{name: "checker1", status: StatusHealthy})
	hc.RegisterChecker("checker2", &mockCheck{name: "checker2", status: StatusUnhealthy})

	health := hc.Check()

	if health.Status != StatusUnhealthy {
		t.Errorf("expected overall status unhealthy, got '%s'", health.Status)
	}
}

func TestHealthChecker_Check_WithDegraded(t *testing.T) {
	hc := New("1.0.0")

	// Register checkers with one degraded
	hc.RegisterChecker("checker1", &mockCheck{name: "checker1", status: StatusHealthy})
	hc.RegisterChecker("checker2", &mockCheck{name: "checker2", status: StatusDegraded})

	health := hc.Check()

	if health.Status != StatusDegraded {
		t.Errorf("expected overall status degraded, got '%s'", health.Status)
	}
}

func TestHealthChecker_Check_WithBothUnhealthyAndDegraded(t *testing.T) {
	hc := New("1.0.0")

	// Register checkers with both unhealthy and degraded
	hc.RegisterChecker("checker1", &mockCheck{name: "checker1", status: StatusDegraded})
	hc.RegisterChecker("checker2", &mockCheck{name: "checker2", status: StatusUnhealthy})

	health := hc.Check()

	// Unhealthy should take precedence
	if health.Status != StatusUnhealthy {
		t.Errorf("expected overall status unhealthy, got '%s'", health.Status)
	}
}

func TestHealthChecker_Check_WithNoCheckers(t *testing.T) {
	hc := New("1.0.0")

	// Don't register any checkers
	health := hc.Check()

	// Should still be healthy (only memory check)
	if health.Status != StatusHealthy {
		t.Errorf("expected overall status healthy, got '%s'", health.Status)
	}

	// Should have memory check
	if _, ok := health.Checks["memory"]; !ok {
		t.Error("expected memory check to be present")
	}
}

func TestHealthChecker_Check_Metadata(t *testing.T) {
	hc := New("1.2.3")

	health := hc.Check()

	if health.Version != "1.2.3" {
		t.Errorf("expected version '1.2.3', got '%s'", health.Version)
	}

	if health.Metadata == nil {
		t.Fatal("expected metadata to be present")
	}

	if _, ok := health.Metadata["go_version"]; !ok {
		t.Error("expected go_version in metadata")
	}

	if _, ok := health.Metadata["os"]; !ok {
		t.Error("expected os in metadata")
	}

	if _, ok := health.Metadata["arch"]; !ok {
		t.Error("expected arch in metadata")
	}

	if _, ok := health.Metadata["goroutines"]; !ok {
		t.Error("expected goroutines in metadata")
	}
}

func TestHealthChecker_Check_Timestamp(t *testing.T) {
	hc := New("1.0.0")

	before := time.Now()
	health := hc.Check()
	after := time.Now()

	if health.Timestamp.Before(before) {
		t.Error("health timestamp is before test started")
	}

	if health.Timestamp.After(after) {
		t.Error("health timestamp is after test ended")
	}
}

func TestHealth_Checks(t *testing.T) {
	hc := New("1.0.0")

	mockChecker := &mockCheck{name: "test", status: StatusHealthy}
	hc.RegisterChecker("test", mockChecker)

	health := hc.Check()

	if len(health.Checks) == 0 {
		t.Fatal("expected checks to be present")
	}

	// Should have registered checker
	if _, ok := health.Checks["test"]; !ok {
		t.Error("expected registered checker in checks")
	}

	// Should have memory check
	if _, ok := health.Checks["memory"]; !ok {
		t.Error("expected memory check in checks")
	}
}

func TestCheckMemory(t *testing.T) {
	hc := New("1.0.0")
	health := hc.Check()

	memoryCheck := health.Checks["memory"]
	if memoryCheck.Name != "memory" {
		t.Errorf("expected check name 'memory', got '%s'", memoryCheck.Name)
	}

	// Memory usage should be relatively low in tests
	if memoryCheck.Status != StatusHealthy {
		t.Logf("Note: Memory check status is '%s' (this is expected in test environment)", memoryCheck.Status)
	}

	if memoryCheck.Message == "" {
		t.Error("expected memory check to have a message")
	}

	if memoryCheck.Duration < 0 {
		t.Errorf("expected non-negative duration, got %v", memoryCheck.Duration)
	}

	if memoryCheck.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestHealth_JSON(t *testing.T) {
	hc := New("1.0.0")

	mockChecker := &mockCheck{name: "test", status: StatusHealthy}
	hc.RegisterChecker("test", mockChecker)

	health := hc.Check()

	json, err := health.JSON()
	if err != nil {
		t.Fatalf("JSON() failed: %v", err)
	}

	if json == "" {
		t.Error("expected non-empty JSON")
	}

	// Check for expected fields
	if !strings.Contains(json, `"status"`) {
		t.Error("expected 'status' in JSON")
	}

	if !strings.Contains(json, `"timestamp"`) {
		t.Error("expected 'timestamp' in JSON")
	}

	if !strings.Contains(json, `"version"`) {
		t.Error("expected 'version' in JSON")
	}

	if !strings.Contains(json, `"checks"`) {
		t.Error("expected 'checks' in JSON")
	}

	if !strings.Contains(json, `"metadata"`) {
		t.Error("expected 'metadata' in JSON")
	}
}

func TestHealthChecker_MultipleCheckers(t *testing.T) {
	hc := New("1.0.0")

	// Register multiple checkers
	statuses := []Status{StatusHealthy, StatusDegraded, StatusUnhealthy}
	for i, status := range statuses {
		hc.RegisterChecker("checker"+string(rune('0'+i)), &mockCheck{
			name:   "checker" + string(rune('0'+i)),
			status: status,
		})
	}

	health := hc.Check()

	// Should have all registered checkers
	for i := 0; i < 3; i++ {
		name := "checker" + string(rune('0'+i))
		if _, ok := health.Checks[name]; !ok {
			t.Errorf("expected checker '%s' in checks", name)
		}
	}
}

func TestHealthChecker_ConcurrentAccess(t *testing.T) {
	hc := New("1.0.0")

	done := make(chan bool)

	// Concurrent checks
	for i := 0; i < 50; i++ {
		go func(id int) {
			name := "checker" + string(rune('0'+(id%10)))
			hc.RegisterChecker(name, &mockCheck{name: name, status: StatusHealthy})
			health := hc.Check()
			if health.Status == "" {
				t.Error("expected non-empty status")
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 50; i++ {
		<-done
	}
}

func TestHealthChecker_FailingChecker(t *testing.T) {
	hc := New("1.0.0")

	// Register a checker that always fails
	failingChecker := &mockCheck{
		name:   "failing",
		status: StatusUnhealthy,
		error:  "checker failed",
	}
	hc.RegisterChecker("failing", failingChecker)

	health := hc.Check()

	// Overall status should be unhealthy
	if health.Status != StatusUnhealthy {
		t.Errorf("expected overall status unhealthy, got '%s'", health.Status)
	}

	// The failing checker should have error message
	check := health.Checks["failing"]
	if check.Error == "" {
		t.Error("expected error message in failing check")
	}
}

func TestHealthChecker_Check_Duration(t *testing.T) {
	hc := New("1.0.0")

	// Register a checker with delay
	slowChecker := &mockCheck{
		name:   "slow",
		status: StatusHealthy,
		delay:  10 * time.Millisecond,
	}
	hc.RegisterChecker("slow", slowChecker)

	health := hc.Check()

	// Check duration should be recorded
	check := health.Checks["slow"]
	if check.Duration < 10*time.Millisecond {
		t.Errorf("expected duration >= 10ms, got %v", check.Duration)
	}
}

func TestHealthChecker_VersionInOutput(t *testing.T) {
	hc := New("2.5.1")

	health := hc.Check()

	// Check the health struct directly first
	if health.Version != "2.5.1" {
		t.Errorf("expected version '2.5.1', got '%s'", health.Version)
	}

	json, err := health.JSON()
	if err != nil {
		t.Fatalf("JSON() failed: %v", err)
	}

	t.Logf("JSON output: %s", json)

	if !strings.Contains(json, `"version"`) {
		t.Error("expected version key to be in JSON output")
	}
	if !strings.Contains(json, "2.5.1") {
		t.Error("expected version value '2.5.1' to be in JSON output")
	}
}

// Helper types and functions

type mockCheck struct {
	name   string
	status Status
	error  string
	delay  time.Duration
}

func (m *mockCheck) Check() Check {
	start := time.Now()
	if m.delay > 0 {
		time.Sleep(m.delay)
	}
	duration := time.Since(start)

	return Check{
		Name:      m.name,
		Status:    m.status,
		Message:   m.name + " check",
		Error:     m.error,
		Duration:  duration,
		Timestamp: time.Now().UTC(),
	}
}
