package validations

import (
	"strings"
	"testing"
)

// TestSanitizeCode tests the SanitizeCode function
func TestSanitizeCode(t *testing.T) {
	tests := []struct {
		name      string
		code      string
		want      string
		wantErr   bool
		errString string
	}{
		{
			name:    "empty code",
			code:    "",
			want:    "",
			wantErr: false,
		},
		{
			name:    "normal code",
			code:    "package main\n\nfunc main() {\n\tfmt.Println(\"Hello\")\n}",
			want:    "package main\n\nfunc main() {\n\tfmt.Println(\"Hello\")\n}",
			wantErr: false,
		},
		{
			name:    "code with trailing whitespace",
			code:    "func main() {   \n\tfmt.Println() \t \n}",
			want:    "func main() {\n\tfmt.Println()\n}",
			wantErr: false,
		},
		{
			name:    "code with null bytes",
			code:    "package main\x00",
			want:    "package main",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SanitizeCode(tt.code)

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
				if got != tt.want {
					t.Errorf("expected '%s', got '%s'", tt.want, got)
				}
			}
		})
	}
}

// TestSanitizePath tests the SanitizePath function
func TestSanitizePath(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		want      string
		wantErr   bool
		errString string
	}{
		{
			name:    "empty path",
			path:    "",
			want:    "",
			wantErr: false,
		},
		{
			name:    "normal path",
			path:    "path/to/file.go",
			want:    "path/to/file.go",
			wantErr: false,
		},
		{
			name:    "path with backslashes",
			path:    "path\\to\\file.go",
			want:    "path/to/file.go",
			wantErr: false,
		},
		{
			name:      "path traversal",
			path:      "../etc/passwd",
			want:      "",
			wantErr:   true,
			errString: "cannot contain '..'",
		},
		{
			name:      "null byte injection",
			path:      "file\x00.go",
			want:      "",
			wantErr:   true,
			errString: "cannot contain null bytes",
		},
		{
			name:    "path with leading/trailing dots",
			path:    "./path/to/file/",
			want:    "path/to/file",
			wantErr: false,
		},
		{
			name:    "path with duplicate slashes",
			path:    "path//to///file.go",
			want:    "path/to/file.go",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SanitizePath(tt.path)

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
				if got != tt.want {
					t.Errorf("expected '%s', got '%s'", tt.want, got)
				}
			}
		})
	}
}

// TestSanitizeSymbol tests the SanitizeSymbol function
func TestSanitizeSymbol(t *testing.T) {
	tests := []struct {
		name      string
		symbol    string
		want      string
		wantErr   bool
		errString string
	}{
		{
			name:    "empty symbol",
			symbol:  "",
			want:    "",
			wantErr: false,
		},
		{
			name:    "valid symbol",
			symbol:  "Println",
			want:    "Println",
			wantErr: false,
		},
		{
			name:    "symbol with whitespace",
			symbol:  "  Println  ",
			want:    "Println",
			wantErr: false,
		},
		{
			name:      "invalid symbol with dash",
			symbol:    "my-symbol",
			want:      "",
			wantErr:   true,
			errString: "invalid Go symbol name format",
		},
		{
			name:      "symbol starting with number",
			symbol:    "1symbol",
			want:      "",
			wantErr:   true,
			errString: "invalid Go symbol name format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SanitizeSymbol(tt.symbol)

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
				if got != tt.want {
					t.Errorf("expected '%s', got '%s'", tt.want, got)
				}
			}
		})
	}
}

// TestStripComments tests the StripComments function
func TestStripComments(t *testing.T) {
	tests := []struct {
		name string
		code string
		want string
	}{
		{
			name: "empty code",
			code: "",
			want: "",
		},
		{
			name: "code without comments",
			code: "package main\n\nfunc main() {\n\tfmt.Println(\"Hello\")\n}",
			want: "package main\n\nfunc main() {\n\tfmt.Println(\"Hello\")\n}",
		},
		{
			name: "code with single-line comments",
			code: "package main\n\n// This is a comment\nfunc main() {\n\tfmt.Println(\"Hello\") // another comment\n}",
			want: "package main\n\n\nfunc main() {\n\tfmt.Println(\"Hello\") \n}",
		},
		{
			name: "code with multi-line comments",
			code: "package main\n\n/*\nMulti-line comment\n*/\nfunc main() {\n\tfmt.Println(\"Hello\")\n}",
			want: "package main\n\n\nfunc main() {\n\tfmt.Println(\"Hello\")\n}",
		},
		{
			name: "code with mixed comments",
			code: "package main\n\n/* comment */\nfunc main() {\n\t// inline comment\n\tfmt.Println(\"Hello\")\n}",
			want: "package main\n\n\nfunc main() {\n\t\n\tfmt.Println(\"Hello\")\n}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StripComments(tt.code)
			if got != tt.want {
				t.Errorf("expected '%s', got '%s'", tt.want, got)
			}
		})
	}
}

// TestEscapeSpecialChars tests the EscapeSpecialChars function
func TestEscapeSpecialChars(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty input",
			input: "",
			want:  "",
		},
		{
			name:  "normal text",
			input: "Hello World",
			want:  "Hello World",
		},
		{
			name:  "text with backslashes",
			input: "path\\to\\file",
			want:  "path\\\\to\\\\file",
		},
		{
			name:  "text with double quotes",
			input: `say "hello"`,
			want:  `say \"hello\"`,
		},
		{
			name:  "text with single quotes",
			input: "say 'hello'",
			want:  "say \\'hello\\'",
		},
		{
			name:  "text with newlines",
			input: "line1\nline2",
			want:  "line1\\nline2",
		},
		{
			name:  "text with tabs",
			input: "tab\ttab",
			want:  "tab\\ttab",
		},
		{
			name:  "mixed special characters",
			input: `path\to "file"\nline`,
			want:  `path\\to \"file\"\\nline`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EscapeSpecialChars(tt.input)
			if got != tt.want {
				t.Errorf("expected '%s', got '%s'", tt.want, got)
			}
		})
	}
}
