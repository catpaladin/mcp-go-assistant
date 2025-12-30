package logging

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"mcp-go-assistant/internal/types"
)

// Logger wraps zerolog.Logger with additional functionality
type Logger struct {
	logger zerolog.Logger
}

// New creates a new structured logger
func New(level string, format string, outputPath string, noColor bool) (*Logger, error) {
	// Set log level
	logLevel, err := parseLogLevel(level)
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}
	zerolog.SetGlobalLevel(logLevel)

	// Configure output writer
	var output io.Writer
	if outputPath == "" || outputPath == "stdout" {
		output = os.Stdout
	} else if outputPath == "stderr" {
		output = os.Stderr
	} else {
		file, err := os.OpenFile(outputPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		output = file
	}

	// Configure console format
	if strings.ToLower(format) == "console" {
		output = zerolog.ConsoleWriter{
			Out:        output,
			NoColor:    noColor,
			TimeFormat: time.RFC3339,
		}
	}

	// Create logger
	logger := zerolog.New(output).
		With().
		Timestamp().
		Logger()

	// Add default fields
	logger = logger.With().
		Str("service", "mcp-go-assistant").
		Logger()

	return &Logger{logger: logger}, nil
}

// parseLogLevel parses a log level string
func parseLogLevel(level string) (zerolog.Level, error) {
	switch strings.ToLower(level) {
	case "trace":
		return zerolog.TraceLevel, nil
	case "debug":
		return zerolog.DebugLevel, nil
	case "info":
		return zerolog.InfoLevel, nil
	case "warn", "warning":
		return zerolog.WarnLevel, nil
	case "error":
		return zerolog.ErrorLevel, nil
	case "fatal":
		return zerolog.FatalLevel, nil
	case "panic":
		return zerolog.PanicLevel, nil
	default:
		return zerolog.InfoLevel, fmt.Errorf("unknown log level: %s", level)
	}
}

// WithRequestID returns a logger with a request ID field
func (l *Logger) WithRequestID(requestID string) *Logger {
	return &Logger{
		logger: l.logger.With().Str("request_id", requestID).Logger(),
	}
}

// WithNewRequestID generates and adds a new request ID
func (l *Logger) WithNewRequestID() *Logger {
	requestID := uuid.New().String()
	return &Logger{
		logger: l.logger.With().Str("request_id", requestID).Logger(),
	}
}

// WithField returns a logger with an additional field
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return &Logger{
		logger: l.logger.With().Interface(key, value).Logger(),
	}
}

// WithFields returns a logger with multiple additional fields
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	ctx := l.logger.With()
	for k, v := range fields {
		ctx = ctx.Interface(k, v)
	}
	return &Logger{logger: ctx.Logger()}
}

// WithError returns a logger with an error field
func (l *Logger) WithError(err error) *Logger {
	if err == nil {
		return l
	}
	return &Logger{
		logger: l.logger.With().Err(err).Logger(),
	}
}

// Trace logs a trace message
func (l *Logger) Trace(msg string) {
	l.logger.Trace().Msg(msg)
}

// Tracef logs a formatted trace message
func (l *Logger) Tracef(format string, args ...interface{}) {
	l.logger.Trace().Msgf(format, args...)
}

// Debug logs a debug message
func (l *Logger) Debug(msg string) {
	l.logger.Debug().Msg(msg)
}

// Debugf logs a formatted debug message
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.logger.Debug().Msgf(format, args...)
}

// Info logs an info message
func (l *Logger) Info(msg string) {
	l.logger.Info().Msg(msg)
}

// Infof logs a formatted info message
func (l *Logger) Infof(format string, args ...interface{}) {
	l.logger.Info().Msgf(format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string) {
	l.logger.Warn().Msg(msg)
}

// Warnf logs a formatted warning message
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.logger.Warn().Msgf(format, args...)
}

// Error logs an error message
func (l *Logger) Error(msg string) {
	l.logger.Error().Msg(msg)
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.logger.Error().Msgf(format, args...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string) {
	l.logger.Fatal().Msg(msg)
}

// Fatalf logs a formatted fatal message and exits
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.logger.Fatal().Msgf(format, args...)
}

// InfoEvent returns a zerolog event for info level
func (l *Logger) InfoEvent() *zerolog.Event {
	return l.logger.Info()
}

// ErrorEvent returns a zerolog event for error level
func (l *Logger) ErrorEvent() *zerolog.Event {
	return l.logger.Error()
}

// WarnEvent returns a zerolog event for warn level
func (l *Logger) WarnEvent() *zerolog.Event {
	return l.logger.Warn()
}

// DebugEvent returns a zerolog event for debug level
func (l *Logger) DebugEvent() *zerolog.Event {
	return l.logger.Debug()
}

// FatalEvent returns a zerolog event for fatal level
func (l *Logger) FatalEvent() *zerolog.Event {
	return l.logger.Fatal()
}

// LogRequest logs an incoming request
func (l *Logger) LogRequest(method string, tool string, params interface{}) {
	l.logger.Info().
		Str("method", method).
		Str("tool", tool).
		Interface("params", params).
		Msg("incoming request")
}

// LogResponse logs a response
func (l *Logger) LogResponse(method string, tool string, duration time.Duration, err error) {
	event := l.logger.Info().
		Str("method", method).
		Str("tool", tool).
		Dur("duration_ms", duration)

	if err != nil {
		event = event.Err(err)
		event.Msg("request completed with error")
	} else {
		event.Msg("request completed successfully")
	}
}

// LogToolCall logs a tool invocation
func (l *Logger) LogToolCall(tool string, duration time.Duration, success bool) {
	status := "success"
	if !success {
		status = "failed"
	}

	l.logger.Info().
		Str("tool", tool).
		Dur("duration_ms", duration).
		Str("status", status).
		Msg("tool call completed")
}

// LogValidationAttempt logs a validation attempt
func (l *Logger) LogValidationAttempt(field, rule, tool string) {
	l.logger.Debug().
		Str("field", field).
		Str("rule", rule).
		Str("tool", tool).
		Msg("validation attempt")
}

// LogValidationSuccess logs a successful validation
func (l *Logger) LogValidationSuccess(field, rule, tool string) {
	l.logger.Debug().
		Str("field", field).
		Str("rule", rule).
		Str("tool", tool).
		Msg("validation succeeded")
}

// LogValidationError logs a validation failure
func (l *Logger) LogValidationError(field, rule, value, tool string) {
	l.logger.Warn().
		Str("field", field).
		Str("rule", rule).
		Str("value", value).
		Str("tool", tool).
		Msg("validation failed")
}

// LogMCPError logs an MCPError with structured fields
func (l *Logger) LogMCPError(err error, msg string) {
	if err == nil {
		return
	}

	if mcpErr, ok := err.(types.MCPError); ok {
		l.logger.Error().
			Str("error_code", mcpErr.Code()).
			Str("error_category", mcpErr.Category()).
			Int("status_code", mcpErr.StatusCode()).
			Interface("error_details", mcpErr.Details()).
			Time("error_timestamp", mcpErr.Timestamp()).
			Msg(msg)
	} else {
		l.logger.Error().Err(err).Msg(msg)
	}
}

// LogMCPErrorWithFields logs an MCPError with additional fields
func (l *Logger) LogMCPErrorWithFields(err error, fields map[string]interface{}) {
	if err == nil {
		return
	}

	ctx := l.logger.Error()
	for k, v := range fields {
		ctx = ctx.Interface(k, v)
	}

	if mcpErr, ok := err.(types.MCPError); ok {
		ctx = ctx.Str("error_code", mcpErr.Code()).
			Str("error_category", mcpErr.Category()).
			Int("status_code", mcpErr.StatusCode()).
			Interface("error_details", mcpErr.Details()).
			Time("error_timestamp", mcpErr.Timestamp())
	} else {
		ctx = ctx.Err(err)
	}

	ctx.Msg("error occurred")
}
