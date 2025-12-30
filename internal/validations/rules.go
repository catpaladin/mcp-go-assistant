package validations

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	// goIdentifierPattern is the regex pattern for valid Go identifiers
	goIdentifierPattern = `^[a-zA-Z_][a-zA-Z0-9_]*$`
)

// ValidationRule defines the interface for validation rules
type ValidationRule interface {
	Validate(value interface{}) error
	Name() string
}

// ValidatorFunc is a function that validates a value
type ValidatorFunc func(value interface{}) error

// NotEmptyRule validates that a string is not empty
type NotEmptyRule struct{}

// Validate checks if the value is not empty
func (r *NotEmptyRule) Validate(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("value must be a string")
	}
	if strings.TrimSpace(str) == "" {
		return fmt.Errorf("value cannot be empty")
	}
	return nil
}

// Name returns the rule name
func (r *NotEmptyRule) Name() string {
	return "not_empty"
}

// MaxLengthRule validates that a string does not exceed maximum length
type MaxLengthRule struct {
	MaxLength int
}

// Validate checks if the value does not exceed maximum length
func (r *MaxLengthRule) Validate(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("value must be a string")
	}
	if len(str) > r.MaxLength {
		return fmt.Errorf("value exceeds maximum length of %d (got %d)", r.MaxLength, len(str))
	}
	return nil
}

// Name returns the rule name
func (r *MaxLengthRule) Name() string {
	return "max_length"
}

// AllowedCharsRule validates that a string contains only allowed characters
type AllowedCharsRule struct {
	Pattern string // Regular expression pattern
	re      *regexp.Regexp
}

// Validate checks if the value contains only allowed characters
func (r *AllowedCharsRule) Validate(value interface{}) error {
	if r.re == nil {
		r.re = regexp.MustCompile("^" + r.Pattern + "$")
	}

	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("value must be a string")
	}
	if !r.re.MatchString(str) {
		return fmt.Errorf("value contains invalid characters (pattern: %s)", r.Pattern)
	}
	return nil
}

// Name returns the rule name
func (r *AllowedCharsRule) Name() string {
	return "allowed_chars"
}

// PackagePathRule validates Go package paths
type PackagePathRule struct{}

// Validate checks if the value is a valid Go package path
func (r *PackagePathRule) Validate(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("value must be a string")
	}
	if strings.TrimSpace(str) == "" {
		return fmt.Errorf("package path cannot be empty")
	}

	// Prevent path traversal
	if strings.Contains(str, "..") {
		return fmt.Errorf("package path cannot contain '..'")
	}

	// Basic Go package path validation
	// Must start with a letter or underscore, contain only valid characters
	pattern := `^[a-zA-Z_][a-zA-Z0-9_]*(\.[a-zA-Z_][a-zA-Z0-9_]*)*/[a-zA-Z_][a-zA-Z0-9_./-]*$|^std$|^builtin$|^fmt$`
	if !regexp.MustCompile(pattern).MatchString(str) {
		return fmt.Errorf("invalid Go package path format: %s", str)
	}

	return nil
}

// Name returns the rule name
func (r *PackagePathRule) Name() string {
	return "package_path"
}

// FilePathRule validates file paths and prevents traversal attacks
type FilePathRule struct{}

// Validate checks if the value is a valid file path
func (r *FilePathRule) Validate(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("value must be a string")
	}

	// Allow empty paths (will be handled by caller)
	if str == "" {
		return nil
	}

	// Prevent path traversal attempts
	if strings.Contains(str, "..") {
		return fmt.Errorf("file path cannot contain '..' to prevent traversal attacks")
	}

	// Prevent null bytes
	if strings.Contains(str, "\x00") {
		return fmt.Errorf("file path cannot contain null bytes")
	}

	// Basic path validation - allow common path characters
	// This is intentionally permissive, actual file access checks should happen at OS level
	validChars := `^[a-zA-Z0-9_./\-][a-zA-Z0-9_./\-]*$`
	if !regexp.MustCompile(validChars).MatchString(str) {
		return fmt.Errorf("file path contains invalid characters")
	}

	return nil
}

// Name returns the rule name
func (r *FilePathRule) Name() string {
	return "file_path"
}

// CodeSafetyRule validates code for basic safety patterns
type CodeSafetyRule struct{}

// Validate checks if the code is safe for analysis
func (r *CodeSafetyRule) Validate(value interface{}) error {
	code, ok := value.(string)
	if !ok {
		return fmt.Errorf("value must be a string")
	}

	// Check for suspicious patterns that might indicate injection attempts
	dangerousPatterns := []string{
		"`rm -rf`",
		"exec.Command(",
		"os.system(",
		"subprocess.call(",
		"eval(",
		"`__import__`",
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(code, pattern) {
			return fmt.Errorf("code contains potentially dangerous pattern: %s", pattern)
		}
	}

	return nil
}

// Name returns the rule name
func (r *CodeSafetyRule) Name() string {
	return "code_safety"
}

// SymbolNameRule validates Go symbol names
type SymbolNameRule struct{}

// Validate checks if the value is a valid Go symbol name
func (r *SymbolNameRule) Validate(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("value must be a string")
	}

	// Allow empty symbols (will query all symbols in package)
	if str == "" {
		return nil
	}

	// Validate symbol name format
	// Go identifiers: start with letter or underscore, followed by letters, digits, or underscores
	if !regexp.MustCompile(goIdentifierPattern).MatchString(str) {
		return fmt.Errorf("invalid Go symbol name: %s", str)
	}

	return nil
}

// Name returns the rule name
func (r *SymbolNameRule) Name() string {
	return "symbol_name"
}
