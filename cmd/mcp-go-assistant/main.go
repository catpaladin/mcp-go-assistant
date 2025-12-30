package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mcp-go-assistant/internal/circuitbreaker"
	"mcp-go-assistant/internal/codereview"
	"mcp-go-assistant/internal/config"
	"mcp-go-assistant/internal/godoc"
	"mcp-go-assistant/internal/logging"
	"mcp-go-assistant/internal/metrics"
	"mcp-go-assistant/internal/ratelimit"
	"mcp-go-assistant/internal/retry"
	"mcp-go-assistant/internal/testgen"
	"mcp-go-assistant/internal/types"
	"mcp-go-assistant/internal/validations"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	toolGoDoc      = "go-doc"
	toolCodeReview = "code-review"
	toolTestGen    = "test-gen"
)

var (
	cfg                      *config.Config
	logger                   *logging.Logger
	metricsCol               *metrics.Metrics
	shutdownChan             chan os.Signal
	goDocCircuitBreaker      *circuitbreaker.CircuitBreaker
	codeReviewCircuitBreaker *circuitbreaker.CircuitBreaker
	testGenCircuitBreaker    *circuitbreaker.CircuitBreaker
	validator                *validations.Validator
	rateLimiter              *ratelimit.Limiter
	rateLimitMiddleware      *ratelimit.Middleware
	goDocRetryWrapper        *retry.RetryWrapper
	codeReviewRetryWrapper   *retry.RetryWrapper
	testGenRetryWrapper      *retry.RetryWrapper
)

// GoDocTool handles the go-doc tool invocation
func GoDocTool(ctx context.Context, request *mcp.CallToolRequest, params godoc.GoDocParams) (*mcp.CallToolResult, any, error) {
	startTime := time.Now()
	log := logger.WithNewRequestID()

	log.InfoEvent().
		Str("tool", toolGoDoc).
		Str("package_path", params.PackagePath).
		Str("symbol_name", params.SymbolName).
		Msg("processing go-doc request")

	// Check rate limit before processing
	if rateLimitMiddleware != nil {
		if err := rateLimitMiddleware.CheckRateLimit(toolGoDoc, "default"); err != nil {
			duration := time.Since(startTime)
			mcpErr := WrapRateLimitError(err, toolGoDoc)
			LogAndHandleError(log, mcpErr, toolGoDoc, duration)
			return nil, nil, mcpErr
		}
	}

	metricsCol.IncrementActiveRequest(toolGoDoc)
	defer metricsCol.DecrementActiveRequest(toolGoDoc)

	// Validate package path
	log.LogValidationAttempt("package_path", "package_path", toolGoDoc)
	metricsCol.RecordValidationAttempt("package_path", toolGoDoc)
	if err := validator.ValidatePackagePath(params.PackagePath); err != nil {
		log.LogValidationError("package_path", "package_path", params.PackagePath, toolGoDoc)
		metricsCol.RecordValidationFailure("package_path", toolGoDoc)
		mcpErr := WrapValidationError(err, toolGoDoc)
		LogAndHandleError(log, mcpErr, toolGoDoc, time.Since(startTime))
		return nil, nil, mcpErr
	}
	log.LogValidationSuccess("package_path", "package_path", toolGoDoc)

	// Validate symbol name (if provided)
	if params.SymbolName != "" {
		log.LogValidationAttempt("symbol_name", "symbol_name", toolGoDoc)
		metricsCol.RecordValidationAttempt("symbol_name", toolGoDoc)
		if err := validator.ValidateInput(params.SymbolName, "symbol_name"); err != nil {
			log.LogValidationError("symbol_name", "symbol_name", params.SymbolName, toolGoDoc)
			metricsCol.RecordValidationFailure("symbol_name", toolGoDoc)
			mcpErr := WrapValidationError(err, toolGoDoc)
			LogAndHandleError(log, mcpErr, toolGoDoc, time.Since(startTime))
			return nil, nil, mcpErr
		}
		log.LogValidationSuccess("symbol_name", "symbol_name", toolGoDoc)
	}

	// Validate working directory (if provided)
	if params.WorkingDir != "" {
		log.LogValidationAttempt("working_dir", "file_path", toolGoDoc)
		metricsCol.RecordValidationAttempt("working_dir", toolGoDoc)
		if err := validator.ValidateFilePath(params.WorkingDir); err != nil {
			log.LogValidationError("working_dir", "file_path", params.WorkingDir, toolGoDoc)
			metricsCol.RecordValidationFailure("working_dir", toolGoDoc)
			mcpErr := WrapValidationError(err, toolGoDoc)
			LogAndHandleError(log, mcpErr, toolGoDoc, time.Since(startTime))
			return nil, nil, mcpErr
		}
		log.LogValidationSuccess("working_dir", "file_path", toolGoDoc)
	}

	// Wrap call with circuit breaker
	var documentation string
	var err error

	cbErr := goDocCircuitBreaker.Call(func() error {
		// Set timeout
		if params.WorkingDir == "" {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, cfg.Tools.GoDocTimeout)
			defer cancel()
		}

		// Use retry logic if configured
		if goDocRetryWrapper != nil {
			_, retryErr := goDocRetryWrapper.DoWithData(ctx, func(attempt uint) (interface{}, error) {
				documentation, err = godoc.GetDocumentation(ctx, params)
				return nil, err
			})
			return retryErr
		}

		// No retry logic
		documentation, err = godoc.GetDocumentation(ctx, params)
		return err
	})

	if cbErr != nil {
		duration := time.Since(startTime)

		// Check if error is circuit breaker open
		var cbOpenErr *circuitbreaker.CircuitBreakerError
		if errors.As(cbErr, &cbOpenErr) && errors.Is(cbErr, circuitbreaker.ErrCircuitBreakerOpen) {
			mcpErr := WrapCircuitBreakerError(cbErr, toolGoDoc)
			LogAndHandleError(log, mcpErr, toolGoDoc, duration)
			return nil, nil, mcpErr
		}

		// Tool execution error - wrap as internal error
		mcpErr := types.WrapError(cbErr, fmt.Sprintf("failed to get documentation for %s", params.PackagePath))
		metricsCol.RecordToolCall(toolGoDoc, "error", duration)
		LogAndHandleError(log, mcpErr, toolGoDoc, duration)
		return nil, nil, mcpErr
	}

	duration := time.Since(startTime)
	metricsCol.RecordToolCall(toolGoDoc, "success", duration)
	log.InfoEvent().
		Dur("duration_ms", duration).
		Msg("go-doc request completed")

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: documentation}},
	}, nil, nil
}

// CodeReviewTool handles the code-review tool invocation
func CodeReviewTool(ctx context.Context, request *mcp.CallToolRequest, params codereview.CodeReviewParams) (*mcp.CallToolResult, *codereview.ReviewResult, error) {
	startTime := time.Now()
	log := logger.WithNewRequestID()

	log.InfoEvent().
		Str("tool", toolCodeReview).
		Str("hint", params.Hint).
		Msg("processing code-review request")

	// Check rate limit before processing
	if rateLimitMiddleware != nil {
		if err := rateLimitMiddleware.CheckRateLimit(toolCodeReview, "default"); err != nil {
			duration := time.Since(startTime)
			mcpErr := WrapRateLimitError(err, toolCodeReview)
			LogAndHandleError(log, mcpErr, toolCodeReview, duration)
			return nil, nil, mcpErr
		}
	}

	metricsCol.IncrementActiveRequest(toolCodeReview)
	defer metricsCol.DecrementActiveRequest(toolCodeReview)

	// Validate code
	log.LogValidationAttempt("code", "code_safety", toolCodeReview)
	metricsCol.RecordValidationAttempt("code_safety", toolCodeReview)
	if err := validator.ValidateCode(params.GoCode); err != nil {
		log.LogValidationError("code", "code_safety", "code", toolCodeReview)
		metricsCol.RecordValidationFailure("code_safety", toolCodeReview)
		mcpErr := WrapValidationError(err, toolCodeReview)
		LogAndHandleError(log, mcpErr, toolCodeReview, time.Since(startTime))
		return nil, nil, mcpErr
	}
	log.LogValidationSuccess("code", "code_safety", toolCodeReview)

	// Validate hint (if provided)
	if params.Hint != "" {
		log.LogValidationAttempt("hint", "hint", toolCodeReview)
		metricsCol.RecordValidationAttempt("hint", toolCodeReview)
		if err := validator.ValidateHint(params.Hint); err != nil {
			log.LogValidationError("hint", "hint", params.Hint, toolCodeReview)
			metricsCol.RecordValidationFailure("hint", toolCodeReview)
			mcpErr := WrapValidationError(err, toolCodeReview)
			LogAndHandleError(log, mcpErr, toolCodeReview, time.Since(startTime))
			return nil, nil, mcpErr
		}
		log.LogValidationSuccess("hint", "hint", toolCodeReview)
	}

	// Validate guidelines file path (if provided)
	if params.GuidelinesFile != "" {
		log.LogValidationAttempt("guidelines_file", "file_path", toolCodeReview)
		metricsCol.RecordValidationAttempt("guidelines_file", toolCodeReview)
		if err := validator.ValidateFilePath(params.GuidelinesFile); err != nil {
			log.LogValidationError("guidelines_file", "file_path", params.GuidelinesFile, toolCodeReview)
			metricsCol.RecordValidationFailure("guidelines_file", toolCodeReview)
			mcpErr := WrapValidationError(err, toolCodeReview)
			LogAndHandleError(log, mcpErr, toolCodeReview, time.Since(startTime))
			return nil, nil, mcpErr
		}
		log.LogValidationSuccess("guidelines_file", "file_path", toolCodeReview)
	}

	// Wrap call with circuit breaker
	var result *codereview.ReviewResult
	var err error

	cbErr := codeReviewCircuitBreaker.Call(func() error {
		// Set timeout
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cfg.Tools.CodeReviewTimeout)
		defer cancel()

		// Use retry logic if configured
		if codeReviewRetryWrapper != nil {
			_, retryErr := codeReviewRetryWrapper.DoWithData(ctx, func(attempt uint) (interface{}, error) {
				result, err = codereview.PerformCodeReview(ctx, params)
				return nil, err
			})
			return retryErr
		}

		// No retry logic
		result, err = codereview.PerformCodeReview(ctx, params)
		return err
	})

	if cbErr != nil {
		duration := time.Since(startTime)

		// Check if error is circuit breaker open
		var cbOpenErr *circuitbreaker.CircuitBreakerError
		if errors.As(cbErr, &cbOpenErr) && errors.Is(cbErr, circuitbreaker.ErrCircuitBreakerOpen) {
			mcpErr := WrapCircuitBreakerError(cbErr, toolCodeReview)
			LogAndHandleError(log, mcpErr, toolCodeReview, duration)
			return nil, nil, mcpErr
		}

		// Tool execution error - wrap as internal error
		mcpErr := types.WrapError(cbErr, "failed to perform code review")
		metricsCol.RecordToolCall(toolCodeReview, "error", duration)
		LogAndHandleError(log, mcpErr, toolCodeReview, duration)
		return nil, nil, mcpErr
	}

	duration := time.Since(startTime)
	metricsCol.RecordToolCall(toolCodeReview, "success", duration)
	log.InfoEvent().
		Dur("duration_ms", duration).
		Int("score", result.Score).
		Msg("code-review request completed")

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: result.String()}},
	}, result, nil
}

// TestGenTool handles the test generation tool invocation
func TestGenTool(ctx context.Context, request *mcp.CallToolRequest, params testgen.TestGenParams) (*mcp.CallToolResult, *testgen.TestGenResult, error) {
	startTime := time.Now()
	log := logger.WithNewRequestID()

	log.InfoEvent().
		Str("tool", toolTestGen).
		Str("focus", params.Focus).
		Str("package_name", params.PackageName).
		Msg("processing test-gen request")

	// Check rate limit before processing
	if rateLimitMiddleware != nil {
		if err := rateLimitMiddleware.CheckRateLimit(toolTestGen, "default"); err != nil {
			duration := time.Since(startTime)
			mcpErr := WrapRateLimitError(err, toolTestGen)
			LogAndHandleError(log, mcpErr, toolTestGen, duration)
			return nil, nil, mcpErr
		}
	}

	metricsCol.IncrementActiveRequest(toolTestGen)
	defer metricsCol.DecrementActiveRequest(toolTestGen)

	// Validate focus (if provided)
	if params.Focus != "" {
		log.LogValidationAttempt("focus", "focus", toolTestGen)
		metricsCol.RecordValidationAttempt("focus", toolTestGen)
		if err := validator.ValidateFocus(params.Focus); err != nil {
			log.LogValidationError("focus", "focus", params.Focus, toolTestGen)
			metricsCol.RecordValidationFailure("focus", toolTestGen)
			mcpErr := WrapValidationError(err, toolTestGen)
			LogAndHandleError(log, mcpErr, toolTestGen, time.Since(startTime))
			return nil, nil, mcpErr
		}
		log.LogValidationSuccess("focus", "focus", toolTestGen)
	}

	// Validate package name (if provided)
	if params.PackageName != "" {
		log.LogValidationAttempt("package_name", "package_name", toolTestGen)
		metricsCol.RecordValidationAttempt("package_name", toolTestGen)
		if err := validator.ValidatePackageName(params.PackageName); err != nil {
			log.LogValidationError("package_name", "package_name", params.PackageName, toolTestGen)
			metricsCol.RecordValidationFailure("package_name", toolTestGen)
			mcpErr := WrapValidationError(err, toolTestGen)
			LogAndHandleError(log, mcpErr, toolTestGen, time.Since(startTime))
			return nil, nil, mcpErr
		}
		log.LogValidationSuccess("package_name", "package_name", toolTestGen)
	}

	// Validate Go code
	log.LogValidationAttempt("code", "code_safety", toolTestGen)
	metricsCol.RecordValidationAttempt("code_safety", toolTestGen)
	if err := validator.ValidateCode(params.GoCode); err != nil {
		log.LogValidationError("code", "code_safety", "code", toolTestGen)
		metricsCol.RecordValidationFailure("code_safety", toolTestGen)
		mcpErr := WrapValidationError(err, toolTestGen)
		LogAndHandleError(log, mcpErr, toolTestGen, time.Since(startTime))
		return nil, nil, mcpErr
	}
	log.LogValidationSuccess("code", "code_safety", toolTestGen)

	// Wrap call with circuit breaker
	var result *testgen.TestGenResult
	var err error

	cbErr := testGenCircuitBreaker.Call(func() error {
		// Set timeout
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cfg.Tools.TestGenTimeout)
		defer cancel()

		// Use retry logic if configured
		if testGenRetryWrapper != nil {
			_, retryErr := testGenRetryWrapper.DoWithData(ctx, func(attempt uint) (interface{}, error) {
				result, err = testgen.GenerateTests(ctx, params)
				return nil, err
			})
			return retryErr
		}

		// No retry logic
		result, err = testgen.GenerateTests(ctx, params)
		return err
	})

	if cbErr != nil {
		duration := time.Since(startTime)

		// Check if error is circuit breaker open
		var cbOpenErr *circuitbreaker.CircuitBreakerError
		if errors.As(cbErr, &cbOpenErr) && errors.Is(cbErr, circuitbreaker.ErrCircuitBreakerOpen) {
			mcpErr := WrapCircuitBreakerError(cbErr, toolTestGen)
			LogAndHandleError(log, mcpErr, toolTestGen, duration)
			return nil, nil, mcpErr
		}

		// Tool execution error - wrap as internal error
		mcpErr := types.WrapError(cbErr, "failed to generate tests")
		metricsCol.RecordToolCall(toolTestGen, "error", duration)
		LogAndHandleError(log, mcpErr, toolTestGen, duration)
		return nil, nil, mcpErr
	}

	duration := time.Since(startTime)
	metricsCol.RecordToolCall(toolTestGen, "success", duration)
	log.InfoEvent().
		Dur("duration_ms", duration).
		Int("interface_count", len(result.Interfaces)).
		Msg("test-gen request completed")

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: result.String()}},
	}, result, nil
}

// classifyError classifies errors for metrics
func classifyError(err error) string {
	if err == nil {
		return "none"
	}

	errStr := err.Error()

	// Classify based on error patterns
	switch {
	case contains(errStr, "timeout", "deadline exceeded"):
		return "timeout"
	case contains(errStr, "not found", "no such file"):
		return "not_found"
	case contains(errStr, "permission", "denied"):
		return "permission"
	case contains(errStr, "parse", "syntax"):
		return "parse_error"
	default:
		return "unknown"
	}
}

// contains checks if any of the substrings are in the text
func contains(text string, substrings ...string) bool {
	textLower := text
	for _, s := range substrings {
		if s != "" && len(s) <= len(textLower) {
			// Simple substring check
			for i := 0; i <= len(textLower)-len(s); i++ {
				if textLower[i:i+len(s)] == s {
					return true
				}
			}
		}
	}
	return false
}

// FormatError formats an error for API response
func FormatError(err error, tool string, exposeDetails bool) []byte {
	if err == nil {
		return []byte(`{"code":"INTERNAL_ERROR","message":"unknown error"}`)
	}

	var mcpErr types.MCPError
	if types.IsMCPError(err) {
		mcpErr = err.(types.MCPError)
	} else {
		// Wrap non-MCPError for consistency
		mcpErr = types.WrapError(err, err.Error())
	}

	// Track error metrics
	if cfg.ErrorHandling.TrackMetrics {
		metricsCol.RecordError(mcpErr.Category(), mcpErr.Code(), tool)
	}

	// Build response
	response := map[string]interface{}{
		"code":        mcpErr.Code(),
		"message":     mcpErr.Error(),
		"category":    mcpErr.Category(),
		"status_code": mcpErr.StatusCode(),
		"timestamp":   mcpErr.Timestamp().Format(time.RFC3339),
	}

	// Include details if configured
	if exposeDetails || cfg.ErrorHandling.ExposeDetails {
		response["details"] = mcpErr.Details()
	}

	// Include stack trace if configured
	if cfg.ErrorHandling.IncludeStack {
		// In production, you'd want to capture the actual stack trace
		response["stack_trace"] = "stack trace capture not implemented"
	}

	jsonBytes, _ := mcpErr.ToJSON()
	if jsonBytes != nil {
		return jsonBytes
	}

	return []byte(`{"code":"INTERNAL_ERROR","message":"error formatting failed"}`)
}

// LogAndHandleError logs and handles errors consistently
func LogAndHandleError(log *logging.Logger, err error, tool string, duration time.Duration) error {
	if err == nil {
		return nil
	}

	// Track error metrics
	if cfg.ErrorHandling.TrackMetrics {
		var code, category string
		if types.IsMCPError(err) {
			mcpErr := err.(types.MCPError)
			code = mcpErr.Code()
			category = mcpErr.Category()
		} else {
			code = classifyError(err)
			category = "unknown"
		}
		metricsCol.RecordError(category, code, tool)
	}

	// Log error with appropriate level
	if cfg.ErrorHandling.LogAllErrors || cfg.Logging.Level == "debug" || cfg.Logging.Level == "trace" {
		if types.IsMCPError(err) {
			log.LogMCPError(err, fmt.Sprintf("%s tool call failed", tool))
		} else {
			log.ErrorEvent().
				Str("tool", tool).
				Err(err).
				Dur("duration_ms", duration).
				Msg(fmt.Sprintf("%s tool call failed", tool))
		}
	}

	return err
}

// WrapValidationError wraps validation errors as MCPError
func WrapValidationError(err error, tool string) error {
	if err == nil {
		return nil
	}

	if verr, ok := err.(*validations.ValidationError); ok {
		mcpErr := verr.ToMCPError()
		if tool != "" {
			mcpErr = types.AddDetail(mcpErr, "tool", tool)
		}
		return mcpErr
	}

	return types.WrapValidationError(err, fmt.Sprintf("%s validation failed", tool))
}

// WrapRateLimitError wraps rate limit errors as MCPError
func WrapRateLimitError(err error, tool string) error {
	if err == nil {
		return nil
	}

	if rerr, ok := err.(*ratelimit.RateLimitError); ok {
		mcpErr := rerr.ToMCPError()
		if tool != "" {
			mcpErr = types.AddDetail(mcpErr, "tool", tool)
		}
		return mcpErr
	}

	return types.WrapRateLimitError(err, fmt.Sprintf("%s rate limit exceeded", tool))
}

// WrapCircuitBreakerError wraps circuit breaker errors as MCPError
func WrapCircuitBreakerError(err error, tool string) error {
	if err == nil {
		return nil
	}

	if cerr, ok := err.(*circuitbreaker.CircuitBreakerError); ok {
		mcpErr := cerr.ToMCPError()
		if tool != "" {
			mcpErr = types.AddDetail(mcpErr, "tool", tool)
		}
		return mcpErr
	}

	return types.WrapCircuitBreakerError(err, fmt.Sprintf("%s circuit breaker open", tool))
}

// init initializes the application
func init() {
	var err error

	// Load configuration
	cfg, err = config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger, err = logging.New(
		cfg.Logging.Level,
		cfg.Logging.Format,
		cfg.Logging.OutputPath,
		cfg.Logging.NoColor,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	// Initialize metrics
	metricsCol = metrics.New()

	// Initialize validator
	validator = validations.NewValidator()
	validator.SetMaxSize(cfg.Validations.MaxInputSize)
	validator.SetAllowedChars(cfg.Validations.AllowedChars)

	logger.InfoEvent().
		Int("max_input_size", cfg.Validations.MaxInputSize).
		Str("allowed_chars", cfg.Validations.AllowedChars).
		Msg("validator initialized")

	// Initialize circuit breakers
	goDocCircuitBreaker = circuitbreaker.NewCircuitBreaker(
		"godoc",
		cfg.Tools.GoDocCircuitBreaker.ToCircuitBreakerConfig("godoc"),
	)
	logger.InfoEvent().
		Str("circuit_breaker", "godoc").
		Str("max_failures", fmt.Sprintf("%d", cfg.Tools.GoDocCircuitBreaker.MaxFailures)).
		Str("timeout", cfg.Tools.GoDocCircuitBreaker.Timeout.String()).
		Msg("godoc circuit breaker initialized")

	codeReviewCircuitBreaker = circuitbreaker.NewCircuitBreaker(
		"code-review",
		cfg.Tools.CodeReviewCircuitBreaker.ToCircuitBreakerConfig("code-review"),
	)
	logger.InfoEvent().
		Str("circuit_breaker", "code-review").
		Str("max_failures", fmt.Sprintf("%d", cfg.Tools.CodeReviewCircuitBreaker.MaxFailures)).
		Str("timeout", cfg.Tools.CodeReviewCircuitBreaker.Timeout.String()).
		Msg("code-review circuit breaker initialized")

	testGenCircuitBreaker = circuitbreaker.NewCircuitBreaker(
		"test-gen",
		cfg.Tools.TestGenCircuitBreaker.ToCircuitBreakerConfig("test-gen"),
	)
	logger.InfoEvent().
		Str("circuit_breaker", "test-gen").
		Str("max_failures", fmt.Sprintf("%d", cfg.Tools.TestGenCircuitBreaker.MaxFailures)).
		Str("timeout", cfg.Tools.TestGenCircuitBreaker.Timeout.String()).
		Msg("test-gen circuit breaker initialized")

	// Initialize rate limiter
	if cfg.RateLimit.Enabled {
		// Create rate limit config
		rlConfig := &ratelimit.Config{
			Enabled:   cfg.RateLimit.Enabled,
			Limit:     cfg.RateLimit.Limit,
			Window:    cfg.RateLimit.Window,
			Mode:      ratelimit.Mode(cfg.RateLimit.Mode),
			Algorithm: ratelimit.Algorithm(cfg.RateLimit.Algorithm),
			StoreType: ratelimit.StoreType(cfg.RateLimit.StoreType),
			KeyPrefix: "mcp",
		}

		// Create store based on type
		var store ratelimit.Store
		switch rlConfig.StoreType {
		case ratelimit.StoreMemory:
			store = ratelimit.NewMemoryStore()
		case ratelimit.StoreNoOp:
			store = ratelimit.NewNoOpStore()
		}

		// Create rate limiter
		rateLimiter = ratelimit.NewLimiter(rlConfig, store, metricsCol, logger)

		// Configure tool-specific rate limits
		for toolName, toolCfg := range cfg.RateLimit.Tools {
			if toolCfg.Enabled {
				rlToolConfig := &ratelimit.ToolConfig{
					Enabled: toolCfg.Enabled,
					Limit:   toolCfg.Limit,
					Window:  toolCfg.Window,
				}
				if err := rateLimiter.SetToolConfig(toolName, rlToolConfig); err != nil {
					logger.ErrorEvent().
						Str("tool", toolName).
						Err(err).
						Msg("failed to set tool rate limit config")
				}
				logger.InfoEvent().
					Str("tool", toolName).
					Int("limit", toolCfg.Limit).
					Dur("window", toolCfg.Window).
					Msg("tool rate limit configured")
			}
		}

		// Create middleware
		rateLimitMiddleware = ratelimit.NewMiddleware(rateLimiter, logger)

		logger.InfoEvent().
			Str("mode", cfg.RateLimit.Mode).
			Int("limit", cfg.RateLimit.Limit).
			Dur("window", cfg.RateLimit.Window).
			Str("algorithm", cfg.RateLimit.Algorithm).
			Msg("rate limiter initialized")
	}

	// Initialize retry wrappers
	if cfg.Retry.Enabled {
		// Create global retry configuration
		globalRetryConfig := cfg.Retry.ToRetryConfig()
		globalRetryer := retry.NewRetryer(globalRetryConfig)

		// GoDoc retry wrapper
		if goDocCfg, ok := cfg.Retry.Tools["godoc"]; ok && goDocCfg.Enabled {
			goDocRetryer := retry.NewRetryer(goDocCfg.ToRetryConfig())
			goDocRetryWrapper = retry.NewRetryWrapper("go-doc", goDocRetryer, logger)
			goDocRetryWrapper.SetRetryIf(retry.RetryableErrors("go-doc"))
			logger.InfoEvent().
				Str("tool", "go-doc").
				Uint("max_attempts", goDocCfg.MaxAttempts).
				Msg("go-doc retry wrapper initialized")
		} else {
			goDocRetryWrapper = retry.NewRetryWrapper("go-doc", globalRetryer, logger)
			goDocRetryWrapper.SetRetryIf(retry.RetryableErrors("go-doc"))
		}

		// CodeReview retry wrapper
		if codeReviewCfg, ok := cfg.Retry.Tools["code-review"]; ok && codeReviewCfg.Enabled {
			codeReviewRetryer := retry.NewRetryer(codeReviewCfg.ToRetryConfig())
			codeReviewRetryWrapper = retry.NewRetryWrapper("code-review", codeReviewRetryer, logger)
			codeReviewRetryWrapper.SetRetryIf(retry.RetryableErrors("code-review"))
			logger.InfoEvent().
				Str("tool", "code-review").
				Uint("max_attempts", codeReviewCfg.MaxAttempts).
				Msg("code-review retry wrapper initialized")
		} else {
			codeReviewRetryWrapper = retry.NewRetryWrapper("code-review", globalRetryer, logger)
			codeReviewRetryWrapper.SetRetryIf(retry.RetryableErrors("code-review"))
		}

		// TestGen retry wrapper
		if testGenCfg, ok := cfg.Retry.Tools["test-gen"]; ok && testGenCfg.Enabled {
			testGenRetryer := retry.NewRetryer(testGenCfg.ToRetryConfig())
			testGenRetryWrapper = retry.NewRetryWrapper("test-gen", testGenRetryer, logger)
			testGenRetryWrapper.SetRetryIf(retry.RetryableErrors("test-gen"))
			logger.InfoEvent().
				Str("tool", "test-gen").
				Uint("max_attempts", testGenCfg.MaxAttempts).
				Msg("test-gen retry wrapper initialized")
		} else {
			testGenRetryWrapper = retry.NewRetryWrapper("test-gen", globalRetryer, logger)
			testGenRetryWrapper.SetRetryIf(retry.RetryableErrors("test-gen"))
		}

		logger.InfoEvent().
			Uint("max_attempts", cfg.Retry.MaxAttempts).
			Str("strategy", cfg.Retry.Strategy).
			Msg("retry wrappers initialized")
	}

	// Initialize shutdown channel
	shutdownChan = make(chan os.Signal, 1)
}

func main() {
	// Setup graceful shutdown
	setupGracefulShutdown()

	logger.InfoEvent().
		Str("version", cfg.Server.Version).
		Str("name", cfg.Server.Name).
		Msg("starting MCP server")

	server := mcp.NewServer(&mcp.Implementation{
		Name:    cfg.Server.Name,
		Version: cfg.Server.Version,
	}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        toolGoDoc,
		Description: "Get Go documentation for packages and symbols using 'go doc' command",
	}, GoDocTool)

	mcp.AddTool(server, &mcp.Tool{
		Name:        toolCodeReview,
		Description: "Analyze Go code and provide improvement suggestions based on best practices",
	}, CodeReviewTool)

	mcp.AddTool(server, &mcp.Tool{
		Name:        toolTestGen,
		Description: "Generate Go test scaffolding including interfaces, mocks, and table-driven tests. Use focus='interfaces' for interface extraction and mocks, 'table' for table-driven tests, or 'unit' for basic unit tests.",
	}, TestGenTool)

	logger.InfoEvent().Msg("MCP server ready")

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run server in a goroutine
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.Run(ctx, &mcp.StdioTransport{})
	}()

	// Wait for shutdown signal or server error
	select {
	case err := <-serverErr:
		if err != nil {
			logger.FatalEvent().Err(err).Msg("server error")
		}
	case sig := <-shutdownChan:
		logger.InfoEvent().Str("signal", sig.String()).Msg("received shutdown signal")

		// Initiate graceful shutdown
		logger.InfoEvent().Dur("timeout", cfg.Timeouts.Shutdown).Msg("starting graceful shutdown")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Timeouts.Shutdown)
		defer shutdownCancel()

		// Cancel server context
		cancel()

		// Wait for server to stop or timeout
		select {
		case <-serverErr:
			logger.InfoEvent().Msg("server stopped gracefully")
		case <-shutdownCtx.Done():
			logger.WarnEvent().Msg("server shutdown timed out")
		}

		logger.InfoEvent().Msg("shutdown complete")
	}
}

// setupGracefulShutdown sets up signal handling for graceful shutdown
func setupGracefulShutdown() {
	signal.Notify(shutdownChan,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGHUP,
	)
}
