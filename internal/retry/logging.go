package retry

import (
	"fmt"
	"time"

	"mcp-go-assistant/internal/logging"
)

// Logger wraps the logging.Logger for retry-specific logging
type Logger struct {
	logger *logging.Logger
}

// NewLogger creates a new retry logger
func NewLogger(logger *logging.Logger) *Logger {
	return &Logger{logger: logger}
}

// LogRetryAttempt logs a retry attempt
func (l *Logger) LogRetryAttempt(tool string, attempt uint, err error, delay time.Duration) {
	if l.logger == nil {
		return
	}

	l.logger.DebugEvent().
		Str("tool", tool).
		Uint("attempt", attempt).
		Err(err).
		Dur("delay_ms", delay).
		Msg("retry attempt")
}

// LogRetrySuccess logs a successful retry
func (l *Logger) LogRetrySuccess(tool string, attempt uint, totalDuration time.Duration) {
	if l.logger == nil {
		return
	}

	l.logger.InfoEvent().
		Str("tool", tool).
		Uint("attempt", attempt).
		Dur("total_duration_ms", totalDuration).
		Msg("retry succeeded")
}

// LogRetryExhausted logs that all retry attempts were exhausted
func (l *Logger) LogRetryExhausted(tool string, attempts uint, totalDelay time.Duration, err error) {
	if l.logger == nil {
		return
	}

	l.logger.WarnEvent().
		Str("tool", tool).
		Uint("attempts", attempts).
		Dur("total_delay_ms", totalDelay).
		Err(err).
		Msg("retry attempts exhausted")
}

// LogRetryCancelled logs that retry was cancelled due to context cancellation
func (l *Logger) LogRetryCancelled(tool string, attempts uint) {
	if l.logger == nil {
		return
	}

	l.logger.DebugEvent().
		Str("tool", tool).
		Uint("attempts", attempts).
		Msg("retry cancelled")
}

// LogRetrySkipped logs that retry was skipped (non-retryable error)
func (l *Logger) LogRetrySkipped(tool string, attempt uint, err error) {
	if l.logger == nil {
		return
	}

	l.logger.DebugEvent().
		Str("tool", tool).
		Uint("attempt", attempt).
		Err(err).
		Msg("retry skipped (non-retryable error)")
}

// LogRetryInit logs retry initialization
func (l *Logger) LogRetryInit(tool string, maxAttempts uint, strategy string, initialDelay time.Duration) {
	if l.logger == nil {
		return
	}

	l.logger.InfoEvent().
		Str("tool", tool).
		Uint("max_attempts", maxAttempts).
		Str("strategy", strategy).
		Dur("initial_delay_ms", initialDelay).
		Msg("retry initialized")
}

// LogRetryConfig logs retry configuration
func (l *Logger) LogRetryConfig(tool string, config *Config) {
	if l.logger == nil {
		return
	}

	l.logger.DebugEvent().
		Str("tool", tool).
		Uint("max_attempts", config.MaxAttempts).
		Dur("initial_delay_ms", config.InitialDelay).
		Dur("max_delay_ms", config.MaxDelay).
		Float64("multiplier", config.Multiplier).
		Bool("jitter", config.Jitter).
		Str("strategy", config.Strategy).
		Msg("retry configuration")
}

// FormatRetryError formats a retry error for logging
func FormatRetryError(err error) string {
	if retryErr, ok := err.(*RetryError); ok {
		return fmt.Sprintf("retry failed after %d attempts (total delay: %v): %v",
			retryErr.Attempts, retryErr.TotalDelay, retryErr.OriginalError)
	}
	return err.Error()
}
