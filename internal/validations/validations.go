package validations

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
)

// Validator is the main validation struct that holds rules and configuration
type Validator struct {
	rules        map[string]ValidationRule
	validators   map[string]ValidatorFunc
	maxSize      int
	allowedChars string
	mu           sync.RWMutex
}

// NewValidator creates a new Validator with default rules
func NewValidator() *Validator {
	v := &Validator{
		rules:        make(map[string]ValidationRule),
		validators:   make(map[string]ValidatorFunc),
		maxSize:      1024 * 1024,         // 1MB default max size
		allowedChars: `[a-zA-Z0-9_ ./\-]`, // Default allowed characters
	}

	// Register default rules
	v.AddRule("not_empty", &NotEmptyRule{})
	v.AddRule("max_length", &MaxLengthRule{MaxLength: v.maxSize})
	v.AddRule("allowed_chars", &AllowedCharsRule{Pattern: v.allowedChars})
	v.AddRule("package_path", &PackagePathRule{})
	v.AddRule("file_path", &FilePathRule{})
	v.AddRule("code_safety", &CodeSafetyRule{})
	v.AddRule("symbol_name", &SymbolNameRule{})

	return v
}

// AddRule registers a validation rule
func (v *Validator) AddRule(name string, rule ValidationRule) {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.rules[name] = rule
}

// AddValidator registers a custom validator function
func (v *Validator) AddValidator(name string, fn ValidatorFunc) {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.validators[name] = fn
}

// ValidateInput validates input string against specified rules
func (v *Validator) ValidateInput(input string, rules ...string) error {
	v.mu.RLock()
	defer v.mu.RUnlock()

	// Check max size first
	if len(input) > v.maxSize {
		return NewValidationError("input", "max_size", input,
			fmt.Sprintf("input size %d exceeds maximum size %d", len(input), v.maxSize))
	}

	// Apply each rule
	for _, ruleName := range rules {
		// Check custom validators first
		if validator, ok := v.validators[ruleName]; ok {
			if err := validator(input); err != nil {
				return NewValidationError("input", ruleName, input, err.Error())
			}
			continue
		}

		// Check predefined rules
		if rule, ok := v.rules[ruleName]; ok {
			if err := rule.Validate(input); err != nil {
				return NewValidationError("input", ruleName, input, err.Error())
			}
			continue
		}

		// Rule not found
		return fmt.Errorf("validation rule '%s' not found", ruleName)
	}

	return nil
}

// ValidatePackagePath validates a Go package path
func (v *Validator) ValidatePackagePath(path string) error {
	return v.ValidateInput(path, "not_empty", "package_path")
}

// ValidateFilePath validates a file path
func (v *Validator) ValidateFilePath(path string) error {
	if path == "" {
		return nil // Empty paths are allowed
	}
	return v.ValidateInput(path, "file_path")
}

// ValidateCode validates code content
func (v *Validator) ValidateCode(code string) error {
	return v.ValidateInput(code, "code_safety")
}

// SetMaxSize sets the maximum input size
func (v *Validator) SetMaxSize(size int) {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.maxSize = size

	// Update max_length rule
	if rule, ok := v.rules["max_length"]; ok {
		if mlRule, ok := rule.(*MaxLengthRule); ok {
			mlRule.MaxLength = size
		}
	}
}

// SetAllowedChars sets the allowed character pattern
func (v *Validator) SetAllowedChars(pattern string) {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.allowedChars = pattern

	// Update allowed_chars rule
	if rule, ok := v.rules["allowed_chars"]; ok {
		if acRule, ok := rule.(*AllowedCharsRule); ok {
			acRule.Pattern = pattern
			acRule.re = nil // Reset compiled regex
		}
	}
}

// ValidateHint validates a hint parameter (e.g., for code review)
func (v *Validator) ValidateHint(hint string) error {
	if hint == "" {
		return nil // Empty hint is allowed
	}

	// Validate hint size and characters
	if len(hint) > 500 {
		return NewValidationError("hint", "max_length", hint,
			fmt.Sprintf("hint exceeds maximum length of 500 (got %d)", len(hint)))
	}

	// Check for potentially dangerous patterns
	dangerousPatterns := []string{
		"<script",
		"javascript:",
		"onerror=",
		"onload=",
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(strings.ToLower(hint), pattern) {
			return NewValidationError("hint", "dangerous_pattern", hint,
				fmt.Sprintf("hint contains potentially dangerous pattern: %s", pattern))
		}
	}

	return nil
}

// ValidateFocus validates the focus parameter for test generation
func (v *Validator) ValidateFocus(focus string) error {
	if focus == "" {
		return nil // Empty focus is allowed
	}

	validFocus := map[string]bool{
		"interfaces": true,
		"unit":       true,
		"table":      true,
	}

	if !validFocus[strings.ToLower(focus)] {
		return NewValidationError("focus", "invalid_value", focus,
			fmt.Sprintf("focus must be one of: interfaces, unit, table (got: %s)", focus))
	}

	return nil
}

// ValidatePackageName validates a Go package name
func (v *Validator) ValidatePackageName(name string) error {
	if name == "" {
		return nil // Empty package name is allowed
	}

	// Validate package name format
	// Go package names must be valid identifiers
	pattern := `^[a-zA-Z_][a-zA-Z0-9_]*$`
	if !regexp.MustCompile(pattern).MatchString(name) {
		return NewValidationError("package_name", "invalid_format", name,
			"invalid Go package name format")
	}

	// Check size
	if len(name) > 100 {
		return NewValidationError("package_name", "max_length", name,
			fmt.Sprintf("package name exceeds maximum length of 100 (got %d)", len(name)))
	}

	return nil
}
