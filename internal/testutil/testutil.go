package testutil

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TempDir creates a temporary directory for tests and returns its path
// It also registers a cleanup function to remove the directory after the test
func TempDir(t *testing.T) string {
	t.Helper()

	dir, err := os.MkdirTemp("", "mcp-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	t.Cleanup(func() {
		os.RemoveAll(dir)
	})

	return dir
}

// TempFile creates a temporary file with the given content
func TempFile(t *testing.T, dir, name, content string) string {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	return path
}

// CreateConfigFile creates a temporary YAML config file
func CreateConfigFile(t *testing.T, content string) string {
	t.Helper()

	dir := TempDir(t)
	return TempFile(t, dir, "config.yaml", content)
}

// BufferWriter returns a buffer that implements io.Writer
func BufferWriter() *bytes.Buffer {
	return &bytes.Buffer{}
}

// Wait waits for a condition or timeout
func Wait(t *testing.T, condition func() bool, timeout time.Duration, message string) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		if condition() {
			return
		}

		if time.Now().After(deadline) {
			t.Fatalf("timeout waiting for condition: %s", message)
		}

		<-ticker.C
	}
}

// AssertEventually asserts that a condition becomes true within the timeout
func AssertEventually(t *testing.T, condition func() bool, timeout time.Duration, msg string) {
	t.Helper()

	start := time.Now()
	for time.Since(start) < timeout {
		if condition() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}

	t.Fatalf("condition not met within %v: %s", timeout, msg)
}

// CaptureOutput captures stdout or stderr during a function execution
func CaptureOutput(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String()
}

// SetupEnv sets environment variables and returns a cleanup function
func SetupEnv(t *testing.T, envVars map[string]string) func() {
	t.Helper()

	oldValues := make(map[string]string)
	for key, value := range envVars {
		oldValues[key] = os.Getenv(key)
		if err := os.Setenv(key, value); err != nil {
			t.Fatalf("failed to set env var %s: %v", key, err)
		}
	}

	return func() {
		for key, oldValue := range oldValues {
			if oldValue == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, oldValue)
			}
		}
	}
}
