package validations

import (
	"fmt"

	"mcp-go-assistant/internal/types"
)

// ValidationError represents a validation error with detailed context
type ValidationError struct {
	Field   string // Which field failed validation
	Rule    string // Which rule failed
	Value   string // The invalid value (truncated if too long)
	Message string // Human-readable error message
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	if e.Field == "" {
		return fmt.Sprintf("validation failed: %s", e.Message)
	}
	return fmt.Sprintf("validation failed for field '%s' (rule: %s): %s", e.Field, e.Rule, e.Message)
}

// Unwrap returns the underlying error (none for ValidationError)
func (e *ValidationError) Unwrap() error {
	return nil
}

// NewValidationError creates a new ValidationError
func NewValidationError(field, rule, value, message string) *ValidationError {
	// Truncate value if too long for logging
	if len(value) > 100 {
		value = value[:97] + "..."
	}
	return &ValidationError{
		Field:   field,
		Rule:    rule,
		Value:   value,
		Message: message,
	}
}

// ToMCPError converts a ValidationError to an MCPError
func (e *ValidationError) ToMCPError() types.MCPError {
	details := make(map[string]interface{})
	if e.Field != "" {
		details["field"] = e.Field
	}
	if e.Rule != "" {
		details["rule"] = e.Rule
	}
	if e.Value != "" {
		details["value"] = e.Value
	}

	return types.NewValidationError(e.Message, "validation_error", details)
}

// IsValidationError checks if an error is a ValidationError
func IsValidationError(err error) bool {
	_, ok := err.(*ValidationError)
	return ok
}
