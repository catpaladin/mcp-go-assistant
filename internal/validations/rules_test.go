package validations

import (
	"strings"
	"testing"
)

// TestNotEmptyRule tests the NotEmptyRule
func TestNotEmptyRule(t *testing.T) {
	rule := &NotEmptyRule{}

	tests := []struct {
		name      string
		value     interface{}
		wantErr   bool
		errString string
	}{
		{
			name:    "valid string",
			value:   "test",
			wantErr: false,
		},
		{
			name:      "empty string",
			value:     "",
			wantErr:   true,
			errString: "value cannot be empty",
		},
		{
			name:      "whitespace string",
			value:     "   ",
			wantErr:   true,
			errString: "value cannot be empty",
		},
		{
			name:      "non-string value",
			value:     123,
			wantErr:   true,
			errString: "value must be a string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rule.Validate(tt.value)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				if tt.errString != "" && !strings.Contains(err.Error(), tt.errString) {
					t.Errorf("expected error to contain '%s', got '%s'", tt.errString, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// TestMaxLengthRule tests the MaxLengthRule
func TestMaxLengthRule(t *testing.T) {
	rule := &MaxLengthRule{MaxLength: 10}

	tests := []struct {
		name      string
		value     interface{}
		wantErr   bool
		errString string
	}{
		{
			name:    "string within limit",
			value:   "test",
			wantErr: false,
		},
		{
			name:    "string at limit",
			value:   strings.Repeat("a", 10),
			wantErr: false,
		},
		{
			name:      "string exceeds limit",
			value:     strings.Repeat("a", 11),
			wantErr:   true,
			errString: "exceeds maximum length",
		},
		{
			name:      "non-string value",
			value:     123,
			wantErr:   true,
			errString: "value must be a string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rule.Validate(tt.value)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				if tt.errString != "" && !strings.Contains(err.Error(), tt.errString) {
					t.Errorf("expected error to contain '%s', got '%s'", tt.errString, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// TestAllowedCharsRule tests the AllowedCharsRule
func TestAllowedCharsRule(t *testing.T) {
	rule := &AllowedCharsRule{Pattern: `[a-zA-Z0-9]+`}

	tests := []struct {
		name      string
		value     interface{}
		wantErr   bool
		errString string
	}{
		{
			name:    "valid alphanumeric",
			value:   "Test123",
			wantErr: false,
		},
		{
			name:      "invalid character",
			value:     "test-123",
			wantErr:   true,
			errString: "invalid characters",
		},
		{
			name:      "non-string value",
			value:     123,
			wantErr:   true,
			errString: "must be a string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rule.Validate(tt.value)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				if tt.errString != "" && !strings.Contains(err.Error(), tt.errString) {
					t.Errorf("expected error to contain '%s', got '%s'", tt.errString, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// TestPackagePathRule tests the PackagePathRule
func TestPackagePathRule(t *testing.T) {
	rule := &PackagePathRule{}

	tests := []struct {
		name      string
		value     interface{}
		wantErr   bool
		errString string
	}{
		{
			name:    "valid stdlib package",
			value:   "fmt",
			wantErr: false,
		},
		{
			name:    "valid external package",
			value:   "github.com/example/pkg",
			wantErr: false,
		},
		{
			name:      "empty string",
			value:     "",
			wantErr:   true,
			errString: "cannot be empty",
		},
		{
			name:      "path traversal",
			value:     "fmt/../os",
			wantErr:   true,
			errString: "cannot contain '..'",
		},
		{
			name:      "non-string value",
			value:     123,
			wantErr:   true,
			errString: "value must be a string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rule.Validate(tt.value)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				if tt.errString != "" && !strings.Contains(err.Error(), tt.errString) {
					t.Errorf("expected error to contain '%s', got '%s'", tt.errString, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// TestFilePathRule tests the FilePathRule
func TestFilePathRule(t *testing.T) {
	rule := &FilePathRule{}

	tests := []struct {
		name      string
		value     interface{}
		wantErr   bool
		errString string
	}{
		{
			name:    "empty path",
			value:   "",
			wantErr: false,
		},
		{
			name:    "valid relative path",
			value:   "path/to/file.go",
			wantErr: false,
		},
		{
			name:      "path traversal",
			value:     "../etc/passwd",
			wantErr:   true,
			errString: "cannot contain '..'",
		},
		{
			name:      "null byte injection",
			value:     "file\x00.go",
			wantErr:   true,
			errString: "cannot contain null bytes",
		},
		{
			name:      "non-string value",
			value:     123,
			wantErr:   true,
			errString: "value must be a string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rule.Validate(tt.value)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				if tt.errString != "" && !strings.Contains(err.Error(), tt.errString) {
					t.Errorf("expected error to contain '%s', got '%s'", tt.errString, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// TestCodeSafetyRule tests the CodeSafetyRule
func TestCodeSafetyRule(t *testing.T) {
	rule := &CodeSafetyRule{}

	tests := []struct {
		name      string
		value     interface{}
		wantErr   bool
		errString string
	}{
		{
			name:    "safe code",
			value:   "package main\n\nfunc main() {\n\tfmt.Println(\"Hello\")\n}",
			wantErr: false,
		},
		{
			name:      "code with exec.Command",
			value:     "exec.Command(\"rm -rf /\")",
			wantErr:   true,
			errString: "dangerous pattern",
		},
		{
			name:      "code with eval",
			value:     "eval(malicious_code)",
			wantErr:   true,
			errString: "dangerous pattern",
		},
		{
			name:      "non-string value",
			value:     123,
			wantErr:   true,
			errString: "value must be a string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rule.Validate(tt.value)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				if tt.errString != "" && !strings.Contains(err.Error(), tt.errString) {
					t.Errorf("expected error to contain '%s', got '%s'", tt.errString, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// TestSymbolNameRule tests the SymbolNameRule
func TestSymbolNameRule(t *testing.T) {
	rule := &SymbolNameRule{}

	tests := []struct {
		name      string
		value     interface{}
		wantErr   bool
		errString string
	}{
		{
			name:    "empty symbol",
			value:   "",
			wantErr: false,
		},
		{
			name:    "valid symbol",
			value:   "Println",
			wantErr: false,
		},
		{
			name:    "valid private symbol",
			value:   "_private",
			wantErr: false,
		},
		{
			name:      "invalid symbol with dash",
			value:     "my-symbol",
			wantErr:   true,
			errString: "invalid Go symbol name",
		},
		{
			name:      "non-string value",
			value:     123,
			wantErr:   true,
			errString: "value must be a string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rule.Validate(tt.value)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				if tt.errString != "" && !strings.Contains(err.Error(), tt.errString) {
					t.Errorf("expected error to contain '%s', got '%s'", tt.errString, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// TestRuleName tests the Name method of all rules
func TestRuleName(t *testing.T) {
	rules := []struct {
		rule ValidationRule
		name string
	}{
		{&NotEmptyRule{}, "not_empty"},
		{&MaxLengthRule{}, "max_length"},
		{&AllowedCharsRule{}, "allowed_chars"},
		{&PackagePathRule{}, "package_path"},
		{&FilePathRule{}, "file_path"},
		{&CodeSafetyRule{}, "code_safety"},
		{&SymbolNameRule{}, "symbol_name"},
	}

	for _, tt := range rules {
		if tt.rule.Name() != tt.name {
			t.Errorf("expected rule name '%s', got '%s'", tt.name, tt.rule.Name())
		}
	}
}
