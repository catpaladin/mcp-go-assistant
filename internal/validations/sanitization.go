package validations

import (
	"regexp"
	"strings"
)

// SanitizeCode removes dangerous patterns from code
func SanitizeCode(code string) (string, error) {
	if code == "" {
		return "", nil
	}

	// Remove null bytes
	sanitized := strings.ReplaceAll(code, "\x00", "")

	// Trim excessive whitespace
	lines := strings.Split(sanitized, "\n")
	cleanLines := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimRight(line, " \t")
		cleanLines = append(cleanLines, trimmed)
	}
	sanitized = strings.Join(cleanLines, "\n")

	return sanitized, nil
}

// SanitizePath prevents path traversal and normalizes path
func SanitizePath(path string) (string, error) {
	if path == "" {
		return "", nil
	}

	// Prevent path traversal
	if strings.Contains(path, "..") {
		return "", NewValidationError("path", "traversal", path, "path cannot contain '..' to prevent traversal attacks")
	}

	// Prevent null bytes
	if strings.Contains(path, "\x00") {
		return "", NewValidationError("path", "null_byte", path, "path cannot contain null bytes")
	}

	// Normalize path separators to forward slashes
	sanitized := strings.ReplaceAll(path, "\\", "/")

	// Remove leading/trailing slashes and dots
	sanitized = strings.Trim(sanitized, "./")

	// Remove duplicate slashes
	for strings.Contains(sanitized, "//") {
		sanitized = strings.ReplaceAll(sanitized, "//", "/")
	}

	return sanitized, nil
}

// SanitizeSymbol validates and sanitizes symbol names
func SanitizeSymbol(symbol string) (string, error) {
	if symbol == "" {
		return "", nil
	}

	// Trim whitespace
	sanitized := strings.TrimSpace(symbol)

	// Validate symbol name format
	pattern := `^[a-zA-Z_][a-zA-Z0-9_]*$`
	if !regexp.MustCompile(pattern).MatchString(sanitized) {
		return "", NewValidationError("symbol", "format", sanitized, "invalid Go symbol name format")
	}

	return sanitized, nil
}

// StripComments removes single-line and multi-line comments from code
func StripComments(code string) string {
	if code == "" {
		return ""
	}

	// Remove multi-line comments /* ... */
	multiLineCommentRE := regexp.MustCompile(`/\*[\s\S]*?\*/`)
	code = multiLineCommentRE.ReplaceAllString(code, "")

	// Remove single-line comments // ...
	singleLineCommentRE := regexp.MustCompile(`//.*`)
	code = singleLineCommentRE.ReplaceAllString(code, "")

	return code
}

// EscapeSpecialChars escapes potentially dangerous characters
func EscapeSpecialChars(input string) string {
	if input == "" {
		return ""
	}

	// Escape backslashes
	escaped := strings.ReplaceAll(input, "\\", "\\\\")

	// Escape quotes
	escaped = strings.ReplaceAll(escaped, "\"", "\\\"")
	escaped = strings.ReplaceAll(escaped, "'", "\\'")

	// Escape newlines and tabs for logging
	escaped = strings.ReplaceAll(escaped, "\n", "\\n")
	escaped = strings.ReplaceAll(escaped, "\r", "\\r")
	escaped = strings.ReplaceAll(escaped, "\t", "\\t")

	return escaped
}
