package validations

import (
	"strings"
	"testing"
)

// TestNewValidator tests the validator creation
func TestNewValidator(t *testing.T) {
	v := NewValidator()

	if v == nil {
		t.Fatal("NewValidator returned nil")
	}

	if v.maxSize != 1024*1024 {
		t.Errorf("expected default maxSize %d, got %d", 1024*1024, v.maxSize)
	}

	// Check that default rules are registered
	expectedRules := []string{
		"not_empty",
		"max_length",
		"allowed_chars",
		"package_path",
		"file_path",
		"code_safety",
		"symbol_name",
	}

	for _, ruleName := range expectedRules {
		if _, ok := v.rules[ruleName]; !ok {
			t.Errorf("expected rule '%s' to be registered", ruleName)
		}
	}
}

// TestAddRule tests adding a custom rule
func TestAddRule(t *testing.T) {
	v := NewValidator()

	// Create a custom rule
	customRule := &MaxLengthRule{MaxLength: 100}
	v.AddRule("custom_max", customRule)

	if _, ok := v.rules["custom_max"]; !ok {
		t.Error("custom rule was not added")
	}

	// Note: rule.Name() returns the rule's type name, not the key it's stored under
	// The rule was stored under "custom_max" key
	if v.rules["custom_max"] == nil {
		t.Error("custom rule is nil")
	}
}

// TestAddValidator tests adding a custom validator function
func TestAddValidator(t *testing.T) {
	v := NewValidator()

	// Create a custom validator
	customValidator := func(value interface{}) error {
		if value == "invalid" {
			return NewValidationError("test", "custom", "invalid", "invalid value")
		}
		return nil
	}

	v.AddValidator("custom", customValidator)

	if _, ok := v.validators["custom"]; !ok {
		t.Error("custom validator was not added")
	}
}

// TestValidateInput tests input validation
func TestValidateInput(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name     string
		input    string
		rules    []string
		wantErr  bool
		errField string
		errRule  string
	}{
		{
			name:    "valid input with not_empty",
			input:   "test",
			rules:   []string{"not_empty"},
			wantErr: false,
		},
		{
			name:    "empty input with not_empty",
			input:   "",
			rules:   []string{"not_empty"},
			wantErr: true,
			errRule: "not_empty",
		},
		{
			name:    "whitespace input with not_empty",
			input:   "   ",
			rules:   []string{"not_empty"},
			wantErr: true,
			errRule: "not_empty",
		},
		{
			name:    "input exceeds max size",
			input:   strings.Repeat("a", 1024*1024+1),
			rules:   []string{"not_empty"},
			wantErr: true,
			errRule: "max_size",
		},
		{
			name:    "invalid rule name",
			input:   "test",
			rules:   []string{"nonexistent_rule"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateInput(tt.input, tt.rules...)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				// Check if it's a ValidationError
				if tt.errRule != "" {
					if !IsValidationError(err) {
						t.Errorf("expected ValidationError, got %T", err)
					}
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// TestValidatePackagePath tests package path validation
func TestValidatePackagePath(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid package path",
			path:    "fmt",
			wantErr: false,
		},
		{
			name:    "valid package path with subpackage",
			path:    "github.com/example/pkg/sub",
			wantErr: false,
		},
		{
			name:    "empty package path",
			path:    "",
			wantErr: true,
		},
		{
			name:    "path traversal attack",
			path:    "fmt/../os",
			wantErr: true,
		},
		{
			name:    "invalid characters",
			path:    "fmt!@#",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidatePackagePath(tt.path)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// TestValidateFilePath tests file path validation
func TestValidateFilePath(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid file path",
			path:    "path/to/file.go",
			wantErr: false,
		},
		{
			name:    "empty path",
			path:    "",
			wantErr: false, // Empty paths are allowed
		},
		{
			name:    "path traversal attack",
			path:    "../../../etc/passwd",
			wantErr: true,
		},
		{
			name:    "null byte injection",
			path:    "file\x00.go",
			wantErr: true,
		},
		{
			name:    "relative path with dot",
			path:    "./file.go",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateFilePath(tt.path)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// TestValidateCode tests code validation
func TestValidateCode(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name    string
		code    string
		wantErr bool
	}{
		{
			name:    "valid Go code",
			code:    "package main\n\nfunc main() {\n\tfmt.Println(\"Hello\")\n}",
			wantErr: false,
		},
		{
			name:    "code with dangerous pattern",
			code:    "exec.Command(\"rm -rf /\")",
			wantErr: true,
		},
		{
			name:    "code with eval",
			code:    "eval(malicious_code)",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateCode(tt.code)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// TestSetMaxSize tests setting max size
func TestSetMaxSize(t *testing.T) {
	v := NewValidator()

	v.SetMaxSize(100)

	if v.maxSize != 100 {
		t.Errorf("expected maxSize %d, got %d", 100, v.maxSize)
	}

	// Test that the rule is updated
	err := v.ValidateInput(strings.Repeat("a", 101))
	if err == nil {
		t.Error("expected error for input exceeding new max size")
	}
}

// TestSetAllowedChars tests setting allowed characters
func TestSetAllowedChars(t *testing.T) {
	v := NewValidator()

	v.SetAllowedChars("[a-zA-Z0-9]")

	if v.allowedChars != "[a-zA-Z0-9]" {
		t.Errorf("expected allowedChars %s, got %s", "[a-zA-Z0-9]", v.allowedChars)
	}

	// Test validation with special characters - need to explicitly specify allowed_chars rule
	err := v.ValidateInput("test-123", "allowed_chars")
	if err == nil {
		t.Error("expected error for input with disallowed characters")
	}
}

// TestValidateHint tests hint validation
func TestValidateHint(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name    string
		hint    string
		wantErr bool
	}{
		{
			name:    "empty hint",
			hint:    "",
			wantErr: false,
		},
		{
			name:    "valid hint",
			hint:    "Focus on error handling",
			wantErr: false,
		},
		{
			name:    "hint exceeds max length",
			hint:    strings.Repeat("a", 501),
			wantErr: true,
		},
		{
			name:    "hint with script tag",
			hint:    "Check <script>alert('xss')</script>",
			wantErr: true,
		},
		{
			name:    "hint with javascript:",
			hint:    "use javascript:void(0)",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateHint(tt.hint)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// TestValidateFocus tests focus validation
func TestValidateFocus(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name    string
		focus   string
		wantErr bool
	}{
		{
			name:    "empty focus",
			focus:   "",
			wantErr: false,
		},
		{
			name:    "valid focus: interfaces",
			focus:   "interfaces",
			wantErr: false,
		},
		{
			name:    "valid focus: unit",
			focus:   "unit",
			wantErr: false,
		},
		{
			name:    "valid focus: table",
			focus:   "table",
			wantErr: false,
		},
		{
			name:    "invalid focus",
			focus:   "invalid",
			wantErr: true,
		},
		{
			name:    "case insensitive",
			focus:   "INTERFACES",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateFocus(tt.focus)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// TestValidatePackageName tests package name validation
func TestValidatePackageName(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name    string
		pkgName string
		wantErr bool
	}{
		{
			name:    "empty package name",
			pkgName: "",
			wantErr: false,
		},
		{
			name:    "valid package name",
			pkgName: "mypackage",
			wantErr: false,
		},
		{
			name:    "package name starting with underscore",
			pkgName: "_private",
			wantErr: false,
		},
		{
			name:    "invalid package name with hyphen",
			pkgName: "my-package",
			wantErr: true,
		},
		{
			name:    "package name exceeds max length",
			pkgName: strings.Repeat("a", 101),
			wantErr: true,
		},
		{
			name:    "package name starting with number",
			pkgName: "1package",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidatePackageName(tt.pkgName)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// TestConcurrency tests concurrent access to validator
func TestConcurrency(t *testing.T) {
	v := NewValidator()

	done := make(chan bool)

	// Run concurrent validations
	for i := 0; i < 100; i++ {
		go func() {
			_ = v.ValidateInput("test", "not_empty")
			_ = v.ValidatePackagePath("fmt")
			_ = v.ValidateFilePath("test.go")
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 100; i++ {
		<-done
	}
}

// TestIsValidationError tests the IsValidationError helper
func TestIsValidationError(t *testing.T) {
	vErr := NewValidationError("field", "rule", "value", "test error")

	if !IsValidationError(vErr) {
		t.Error("expected ValidationError to be detected")
	}

	if IsValidationError(nil) {
		t.Error("expected non-ValidationError to not be detected")
	}
}

// TestValidationError tests ValidationError methods
func TestValidationError(t *testing.T) {
	vErr := NewValidationError("test_field", "test_rule", "test_value", "test message")

	if vErr.Field != "test_field" {
		t.Errorf("expected Field 'test_field', got '%s'", vErr.Field)
	}

	if vErr.Rule != "test_rule" {
		t.Errorf("expected Rule 'test_rule', got '%s'", vErr.Rule)
	}

	if vErr.Value != "test_value" {
		t.Errorf("expected Value 'test_value', got '%s'", vErr.Value)
	}

	if vErr.Message != "test message" {
		t.Errorf("expected Message 'test message', got '%s'", vErr.Message)
	}

	// Test Error method
	errStr := vErr.Error()
	expected := "validation failed for field 'test_field' (rule: test_rule): test message"
	if errStr != expected {
		t.Errorf("expected error string '%s', got '%s'", expected, errStr)
	}

	// Test Unwrap method
	if vErr.Unwrap() != nil {
		t.Error("expected Unwrap to return nil")
	}
}

// TestValidationErrorValueTruncation tests that long values are truncated
func TestValidationErrorValueTruncation(t *testing.T) {
	longValue := strings.Repeat("a", 200)
	vErr := NewValidationError("field", "rule", longValue, "test message")

	if len(vErr.Value) > 100 {
		t.Errorf("expected Value to be truncated to 100 chars, got %d", len(vErr.Value))
	}

	if !strings.HasSuffix(vErr.Value, "...") {
		t.Error("expected truncated Value to end with '...'")
	}
}
