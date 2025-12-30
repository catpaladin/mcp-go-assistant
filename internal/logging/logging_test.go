package logging

import (
	"bytes"
	"os"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name       string
		level      string
		format     string
		outputPath string
		noColor    bool
		wantErr    bool
	}{
		{
			name:       "valid json logger to stdout",
			level:      "info",
			format:     "json",
			outputPath: "stdout",
			noColor:    false,
			wantErr:    false,
		},
		{
			name:       "valid console logger to stderr",
			level:      "debug",
			format:     "console",
			outputPath: "stderr",
			noColor:    true,
			wantErr:    false,
		},
		{
			name:       "invalid log level",
			level:      "invalid",
			format:     "json",
			outputPath: "stdout",
			noColor:    false,
			wantErr:    true,
		},
		{
			name:       "case insensitive log level",
			level:      "INFO",
			format:     "json",
			outputPath: "stdout",
			noColor:    false,
			wantErr:    false,
		},
		{
			name:       "case insensitive format",
			level:      "debug",
			format:     "CONSOLE",
			outputPath: "stdout",
			noColor:    false,
			wantErr:    false,
		},
		{
			name:       "warn alias",
			level:      "warning",
			format:     "json",
			outputPath: "stdout",
			noColor:    false,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := New(tt.level, tt.format, tt.outputPath, tt.noColor)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if logger == nil {
				t.Fatal("expected logger, got nil")
			}
		})
	}
}

func TestNew_WithFileOutput(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-log-*.log")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	_ = tmpFile.Close()

	logger, err := New("info", "json", tmpFile.Name(), false)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	// Log a message
	logger.Info("test message")

	// Give it time to flush
	time.Sleep(100 * time.Millisecond)

	// Read the file and check for the message
	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	if !bytes.Contains(content, []byte("test message")) {
		t.Errorf("log file does not contain expected message: %s", string(content))
	}
}

func TestNew_WithInvalidPath(t *testing.T) {
	_, err := New("info", "json", "/invalid/path/that/does/not/exist.log", false)
	if err == nil {
		t.Error("expected error for invalid path, got nil")
	}
}

func TestLogger_WithField(t *testing.T) {
	logger, err := New("info", "json", "stdout", false)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	newLogger := logger.WithField("key", "value")
	if newLogger == nil {
		t.Fatal("expected logger, got nil")
	}

	// Verify new logger is different instance
	if newLogger == logger {
		t.Error("expected new logger instance")
	}
}

func TestLogger_WithFields(t *testing.T) {
	logger, err := New("info", "json", "stdout", false)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	fields := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
		"key3": true,
	}

	newLogger := logger.WithFields(fields)
	if newLogger == nil {
		t.Fatal("expected logger, got nil")
	}
}

func TestLogger_WithError(t *testing.T) {
	logger, err := New("info", "json", "stdout", false)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	t.Run("with error", func(t *testing.T) {
		newLogger := logger.WithError(&testError{})
		if newLogger == nil {
			t.Fatal("expected logger, got nil")
		}
	})

	t.Run("with nil error", func(t *testing.T) {
		newLogger := logger.WithError(nil)
		// Should return same logger instance
		if newLogger != logger {
			t.Error("expected same logger instance for nil error")
		}
	})
}

func TestLogger_WithRequestID(t *testing.T) {
	logger, err := New("info", "json", "stdout", false)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	requestID := "test-request-id-123"
	newLogger := logger.WithRequestID(requestID)

	if newLogger == nil {
		t.Fatal("expected logger, got nil")
	}
}

func TestLogger_WithNewRequestID(t *testing.T) {
	logger, err := New("info", "json", "stdout", false)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	newLogger := logger.WithNewRequestID()

	if newLogger == nil {
		t.Fatal("expected logger, got nil")
	}

	// Verify we can get different request IDs
	newLogger2 := logger.WithNewRequestID()
	if newLogger == newLogger2 {
		t.Error("expected different logger instances for different request IDs")
	}
}

func TestLogger_LogLevels(t *testing.T) {
	logger, err := New("debug", "console", "", false)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	// Note: This is a simplified test as zerolog doesn't easily support output redirection after creation
	// In real testing, we'd use zerolog's test utilities

	t.Run("trace", func(_ *testing.T) {
		logger.Trace("trace message")
	})

	t.Run("tracef", func(t *testing.T) {
		logger.Tracef("trace %s", "formatted")
	})

	t.Run("debug", func(t *testing.T) {
		logger.Debug("debug message")
	})

	t.Run("debugf", func(t *testing.T) {
		logger.Debugf("debug %s", "formatted")
	})

	t.Run("info", func(t *testing.T) {
		logger.Info("info message")
	})

	t.Run("infof", func(t *testing.T) {
		logger.Infof("info %s", "formatted")
	})

	t.Run("warn", func(t *testing.T) {
		logger.Warn("warn message")
	})

	t.Run("warnf", func(t *testing.T) {
		logger.Warnf("warn %s", "formatted")
	})

	t.Run("error", func(t *testing.T) {
		logger.Error("error message")
	})

	t.Run("errorf", func(t *testing.T) {
		logger.Errorf("error %s", "formatted")
	})
}

func TestLogger_LogRequest(t *testing.T) {
	logger, err := New("info", "json", "stdout", false)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	params := map[string]interface{}{
		"param1": "value1",
		"param2": 123,
	}

	logger.LogRequest("test_method", "test_tool", params)

	// Should not panic
}

func TestLogger_LogResponse(t *testing.T) {
	logger, err := New("info", "json", "stdout", false)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	t.Run("success", func(t *testing.T) {
		logger.LogResponse("test_method", "test_tool", 100*time.Millisecond, nil)
	})

	t.Run("error", func(t *testing.T) {
		err := &testError{}
		logger.LogResponse("test_method", "test_tool", 200*time.Millisecond, err)
	})
}

func TestLogger_LogToolCall(t *testing.T) {
	logger, err := New("info", "json", "stdout", false)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	t.Run("success", func(t *testing.T) {
		logger.LogToolCall("test_tool", 150*time.Millisecond, true)
	})

	t.Run("failure", func(t *testing.T) {
		logger.LogToolCall("test_tool", 200*time.Millisecond, false)
	})
}

func TestLogger_LogValidationAttempt(t *testing.T) {
	logger, err := New("debug", "json", "stdout", false)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	logger.LogValidationAttempt("test_field", "test_rule", "test_tool")
}

func TestLogger_LogValidationSuccess(t *testing.T) {
	logger, err := New("debug", "json", "stdout", false)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	logger.LogValidationSuccess("test_field", "test_rule", "test_tool")
}

func TestLogger_LogValidationError(t *testing.T) {
	logger, err := New("info", "json", "stdout", false)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	logger.LogValidationError("test_field", "test_rule", "test_value", "test_tool")
}

func TestLogger_LogMCPError(t *testing.T) {
	logger, err := New("info", "json", "stdout", false)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	t.Run("with nil error", func(t *testing.T) {
		logger.LogMCPError(nil, "test message")
	})

	t.Run("with non-MCP error", func(t *testing.T) {
		logger.LogMCPError(&testError{}, "test message")
	})
}

func TestLogger_LogMCPErrorWithFields(t *testing.T) {
	logger, err := New("info", "json", "stdout", false)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	t.Run("with nil error", func(t *testing.T) {
		fields := map[string]interface{}{
			"key": "value",
		}
		logger.LogMCPErrorWithFields(nil, fields)
	})

	t.Run("with non-MCP error", func(t *testing.T) {
		fields := map[string]interface{}{
			"key": "value",
		}
		logger.LogMCPErrorWithFields(&testError{}, fields)
	})
}

func TestLogger_EventMethods(t *testing.T) {
	// Use debug level to ensure DebugEvent returns non-nil event
	logger, err := New("debug", "json", "stdout", false)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	t.Run("InfoEvent", func(t *testing.T) {
		event := logger.InfoEvent()
		if event == nil {
			t.Error("expected event, got nil")
		}
		event.Msg("test")
	})

	t.Run("ErrorEvent", func(t *testing.T) {
		event := logger.ErrorEvent()
		if event == nil {
			t.Error("expected event, got nil")
		}
		event.Msg("test")
	})

	t.Run("WarnEvent", func(t *testing.T) {
		event := logger.WarnEvent()
		if event == nil {
			t.Error("expected event, got nil")
		}
		event.Msg("test")
	})

	t.Run("DebugEvent", func(t *testing.T) {
		event := logger.DebugEvent()
		if event == nil {
			t.Error("expected event, got nil")
		}
		event.Msg("test")
	})

	t.Run("FatalEvent", func(t *testing.T) {
		// Note: FatalEvent will call os.Exit, so we can't test it directly
		event := logger.FatalEvent()
		if event == nil {
			t.Error("expected event, got nil")
		}
	})
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		name    string
		level   string
		wantErr bool
	}{
		{"trace", "trace", false},
		{"debug", "debug", false},
		{"info", "info", false},
		{"warn", "warn", false},
		{"warning", "warning", false},
		{"error", "error", false},
		{"fatal", "fatal", false},
		{"panic", "panic", false},
		{"case insensitive", "DEBUG", false},
		{"invalid", "invalid", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseLogLevel(tt.level)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseLogLevel(%q) error = %v, wantErr %v", tt.level, err, tt.wantErr)
			}
		})
	}
}

func TestLogger_ConcurrentAccess(t *testing.T) {
	logger, err := New("info", "json", "stdout", false)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	done := make(chan bool)

	// Run concurrent log operations
	for i := 0; i < 100; i++ {
		go func(id int) {
			logger.WithField("id", id).Info("concurrent message")
			logger.WithFields(map[string]interface{}{"id": id}).Info("concurrent message 2")
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}
}

func TestLogger_ChainedWithMethods(t *testing.T) {
	logger, err := New("info", "json", "stdout", false)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	// Test chaining
	chainedLogger := logger.
		WithField("field1", "value1").
		WithFields(map[string]interface{}{"field2": "value2"}).
		WithRequestID("req-123").
		WithError(&testError{})

	if chainedLogger == nil {
		t.Fatal("expected logger, got nil")
	}

	chainedLogger.Info("chained message")
}

// Helper types and functions

type testError struct{}

func (e *testError) Error() string {
	return "test error"
}

func (e *testError) Code() string {
	return "TEST_ERROR"
}

func (e *testError) Category() string {
	return "test"
}

func (e *testError) StatusCode() int {
	return 500
}

func (e *testError) Details() map[string]interface{} {
	return map[string]interface{}{
		"test": "details",
	}
}

func (e *testError) Unwrap() error {
	return nil
}

func (e *testError) Timestamp() time.Time {
	return time.Now()
}

func (e *testError) ToJSON() ([]byte, error) {
	return []byte(`{"code":"TEST_ERROR"}`), nil
}
