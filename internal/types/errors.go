package types

import (
	"encoding/json"
	"fmt"
	"time"
)

// MCPError is the interface for all standardized MCP errors
type MCPError interface {
	error
	// Code returns the error code (e.g., "VALIDATION_FAILED", "RATE_LIMIT_EXCEEDED")
	Code() string
	// Category returns the error category (e.g., "validation", "rate_limit", "internal")
	Category() string
	// StatusCode returns HTTP-like status code (e.g., 400, 429, 500)
	StatusCode() int
	// Details returns additional context as a map
	Details() map[string]interface{}
	// Unwrap returns the underlying error (if any)
	Unwrap() error
	// Timestamp returns when the error was created
	Timestamp() time.Time
	// ToJSON returns a JSON representation of the error
	ToJSON() ([]byte, error)
}

// ErrorType represents predefined error types
type ErrorType struct {
	code       string
	category   string
	statusCode int
}

// Predefined error types
var (
	ErrorTypeValidation = ErrorType{
		code:       "VALIDATION_FAILED",
		category:   "validation",
		statusCode: 400,
	}
	ErrorTypeRateLimit = ErrorType{
		code:       "RATE_LIMIT_EXCEEDED",
		category:   "rate_limit",
		statusCode: 429,
	}
	ErrorTypeCircuitBreaker = ErrorType{
		code:       "CIRCUIT_BREAKER_OPEN",
		category:   "circuit_breaker",
		statusCode: 503,
	}
	ErrorTypeInternal = ErrorType{
		code:       "INTERNAL_ERROR",
		category:   "internal",
		statusCode: 500,
	}
	ErrorTypeNotFound = ErrorType{
		code:       "NOT_FOUND",
		category:   "not_found",
		statusCode: 404,
	}
	ErrorTypeTimeout = ErrorType{
		code:       "TIMEOUT",
		category:   "timeout",
		statusCode: 408,
	}
	ErrorTypeUnauthorized = ErrorType{
		code:       "UNAUTHORIZED",
		category:   "auth",
		statusCode: 401,
	}
	ErrorTypeForbidden = ErrorType{
		code:       "FORBIDDEN",
		category:   "auth",
		statusCode: 403,
	}
	ErrorTypeBadRequest = ErrorType{
		code:       "BAD_REQUEST",
		category:   "client",
		statusCode: 400,
	}
)

// mcpError is the concrete implementation of MCPError
type mcpError struct {
	code       string
	category   string
	statusCode int
	message    string
	details    map[string]interface{}
	underlying error
	timestamp  time.Time
}

// Error implements the error interface
func (e *mcpError) Error() string {
	if e.underlying != nil {
		return fmt.Sprintf("%s: %s: %v", e.code, e.message, e.underlying)
	}
	return fmt.Sprintf("%s: %s", e.code, e.message)
}

// Code returns the error code
func (e *mcpError) Code() string {
	return e.code
}

// Category returns the error category
func (e *mcpError) Category() string {
	return e.category
}

// StatusCode returns the HTTP-like status code
func (e *mcpError) StatusCode() int {
	return e.statusCode
}

// Details returns additional context
func (e *mcpError) Details() map[string]interface{} {
	// Return a copy to prevent modification
	detailsCopy := make(map[string]interface{}, len(e.details))
	for k, v := range e.details {
		detailsCopy[k] = v
	}
	return detailsCopy
}

// Unwrap returns the underlying error
func (e *mcpError) Unwrap() error {
	return e.underlying
}

// Timestamp returns when the error was created
func (e *mcpError) Timestamp() time.Time {
	return e.timestamp
}

// ToJSON returns a JSON representation of the error
func (e *mcpError) ToJSON() ([]byte, error) {
	type jsonError struct {
		Code       string                 `json:"code"`
		Message    string                 `json:"message"`
		Category   string                 `json:"category"`
		StatusCode int                    `json:"status_code"`
		Details    map[string]interface{} `json:"details,omitempty"`
		Timestamp  string                 `json:"timestamp"`
		Underlying string                 `json:"underlying,omitempty"`
	}

	jErr := jsonError{
		Code:       e.code,
		Message:    e.message,
		Category:   e.category,
		StatusCode: e.statusCode,
		Details:    e.details,
		Timestamp:  e.timestamp.Format(time.RFC3339),
	}

	if e.underlying != nil {
		jErr.Underlying = e.underlying.Error()
	}

	return json.Marshal(jErr)
}

// NewMCPError creates a new MCPError with the given parameters
func NewMCPError(code, category string, statusCode int, message string) MCPError {
	return &mcpError{
		code:       code,
		category:   category,
		statusCode: statusCode,
		message:    message,
		details:    make(map[string]interface{}),
		timestamp:  time.Now(),
	}
}

// NewValidationError creates a validation error
func NewValidationError(message string, details ...interface{}) MCPError {
	err := NewMCPError(
		ErrorTypeValidation.code,
		ErrorTypeValidation.category,
		ErrorTypeValidation.statusCode,
		message,
	)
	return addDetails(err, details...)
}

// NewRateLimitError creates a rate limit error
func NewRateLimitError(message string, details ...interface{}) MCPError {
	err := NewMCPError(
		ErrorTypeRateLimit.code,
		ErrorTypeRateLimit.category,
		ErrorTypeRateLimit.statusCode,
		message,
	)
	return addDetails(err, details...)
}

// NewCircuitBreakerError creates a circuit breaker error
func NewCircuitBreakerError(message string, details ...interface{}) MCPError {
	err := NewMCPError(
		ErrorTypeCircuitBreaker.code,
		ErrorTypeCircuitBreaker.category,
		ErrorTypeCircuitBreaker.statusCode,
		message,
	)
	return addDetails(err, details...)
}

// NewInternalError creates an internal error
func NewInternalError(message string, details ...interface{}) MCPError {
	err := NewMCPError(
		ErrorTypeInternal.code,
		ErrorTypeInternal.category,
		ErrorTypeInternal.statusCode,
		message,
	)
	return addDetails(err, details...)
}

// NewNotFoundError creates a not found error
func NewNotFoundError(message string, details ...interface{}) MCPError {
	err := NewMCPError(
		ErrorTypeNotFound.code,
		ErrorTypeNotFound.category,
		ErrorTypeNotFound.statusCode,
		message,
	)
	return addDetails(err, details...)
}

// NewTimeoutError creates a timeout error
func NewTimeoutError(message string, details ...interface{}) MCPError {
	err := NewMCPError(
		ErrorTypeTimeout.code,
		ErrorTypeTimeout.category,
		ErrorTypeTimeout.statusCode,
		message,
	)
	return addDetails(err, details...)
}

// NewUnauthorizedError creates an unauthorized error
func NewUnauthorizedError(message string, details ...interface{}) MCPError {
	err := NewMCPError(
		ErrorTypeUnauthorized.code,
		ErrorTypeUnauthorized.category,
		ErrorTypeUnauthorized.statusCode,
		message,
	)
	return addDetails(err, details...)
}

// NewForbiddenError creates a forbidden error
func NewForbiddenError(message string, details ...interface{}) MCPError {
	err := NewMCPError(
		ErrorTypeForbidden.code,
		ErrorTypeForbidden.category,
		ErrorTypeForbidden.statusCode,
		message,
	)
	return addDetails(err, details...)
}

// NewBadRequestError creates a bad request error
func NewBadRequestError(message string, details ...interface{}) MCPError {
	err := NewMCPError(
		ErrorTypeBadRequest.code,
		ErrorTypeBadRequest.category,
		ErrorTypeBadRequest.statusCode,
		message,
	)
	return addDetails(err, details...)
}

// WrapError wraps an existing error with MCPError context
func WrapError(err error, message string, details ...interface{}) MCPError {
	if err == nil {
		return nil
	}

	// If it's already an MCPError, just update the message and add details
	if mcpErr, ok := err.(MCPError); ok {
		return updateMCPError(mcpErr, message, details...)
	}

	// Default to internal error for unknown errors
	return &mcpError{
		code:       ErrorTypeInternal.code,
		category:   ErrorTypeInternal.category,
		statusCode: ErrorTypeInternal.statusCode,
		message:    message,
		details:    parseDetails(details...),
		underlying: err,
		timestamp:  time.Now(),
	}
}

// WrapValidationError wraps an error as a validation error
func WrapValidationError(err error, message string, details ...interface{}) MCPError {
	if err == nil {
		return nil
	}
	return wrapWithCode(err, ErrorTypeValidation, message, details...)
}

// WrapRateLimitError wraps an error as a rate limit error
func WrapRateLimitError(err error, message string, details ...interface{}) MCPError {
	if err == nil {
		return nil
	}
	return wrapWithCode(err, ErrorTypeRateLimit, message, details...)
}

// WrapCircuitBreakerError wraps an error as a circuit breaker error
func WrapCircuitBreakerError(err error, message string, details ...interface{}) MCPError {
	if err == nil {
		return nil
	}
	return wrapWithCode(err, ErrorTypeCircuitBreaker, message, details...)
}

// Errorf creates a formatted error with the given code and category
func Errorf(code, category string, statusCode int, format string, args ...interface{}) MCPError {
	message := fmt.Sprintf(format, args...)
	return NewMCPError(code, category, statusCode, message)
}

// IsMCPError checks if an error is an MCPError
func IsMCPError(err error) bool {
	_, ok := err.(MCPError)
	return ok
}

// GetErrorCode extracts the error code from an error
func GetErrorCode(err error) string {
	if mcpErr, ok := err.(MCPError); ok {
		return mcpErr.Code()
	}
	return ""
}

// GetErrorCategory extracts the error category from an error
func GetErrorCategory(err error) string {
	if mcpErr, ok := err.(MCPError); ok {
		return mcpErr.Category()
	}
	return ""
}

// GetErrorStatusCode extracts the status code from an error
func GetErrorStatusCode(err error) int {
	if mcpErr, ok := err.(MCPError); ok {
		return mcpErr.StatusCode()
	}
	return 500
}

// GetErrorDetails extracts details from an error
func GetErrorDetails(err error) map[string]interface{} {
	if mcpErr, ok := err.(MCPError); ok {
		return mcpErr.Details()
	}
	return nil
}

// addDetails adds details to an MCPError
func addDetails(err MCPError, details ...interface{}) MCPError {
	if err == nil || len(details) == 0 {
		return err
	}

	if mErr, ok := err.(*mcpError); ok {
		mErr.details = parseDetails(details...)
	}
	return err
}

// parseDetails converts the variadic details into a map
func parseDetails(details ...interface{}) map[string]interface{} {
	d := make(map[string]interface{})
	for i := 0; i < len(details); i += 2 {
		if i+1 < len(details) {
			key := fmt.Sprintf("%v", details[i])
			d[key] = details[i+1]
		}
	}
	return d
}

// updateMCPError updates an existing MCPError with new message and details
func updateMCPError(err MCPError, message string, details ...interface{}) MCPError {
	if mErr, ok := err.(*mcpError); ok {
		if message != "" {
			mErr.message = message
		}
		if len(details) > 0 {
			for k, v := range parseDetails(details...) {
				mErr.details[k] = v
			}
		}
	}
	return err
}

// wrapWithCode wraps an error with a specific error type
func wrapWithCode(err error, errorType ErrorType, message string, details ...interface{}) MCPError {
	return &mcpError{
		code:       errorType.code,
		category:   errorType.category,
		statusCode: errorType.statusCode,
		message:    message,
		details:    parseDetails(details...),
		underlying: err,
		timestamp:  time.Now(),
	}
}

// AddDetail adds a key-value pair to the error details
func AddDetail(err MCPError, key string, value interface{}) MCPError {
	if err == nil {
		return nil
	}

	if mErr, ok := err.(*mcpError); ok {
		mErr.details[key] = value
	}

	return err
}
