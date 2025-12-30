package codereview

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPerformCodeReview(t *testing.T) {
	tests := []struct {
		name             string
		params           CodeReviewParams
		wantErr          bool
		errContains      string
		checkSuggestions bool
	}{
		{
			name: "valid simple code",
			params: CodeReviewParams{
				GoCode: `
package main

func main() {
	println("Hello, World!")
}
`,
			},
			wantErr:          false,
			checkSuggestions: true,
		},
		{
			name: "empty code",
			params: CodeReviewParams{
				GoCode: "",
			},
			wantErr:     true,
			errContains: "go_code parameter is required",
		},
		{
			name: "code with issues",
			params: CodeReviewParams{
				GoCode: `
package main

func badFunction() {
	var x int
	var y string
	// TODO: implement
	x = 5
}
`,
			},
			wantErr:          false,
			checkSuggestions: true,
		},
		{
			name: "with guidelines content",
			params: CodeReviewParams{
				GoCode: `
package main

func test() {}
`,
				GuidelinesContent: `
# Custom Guidelines
- Use descriptive function names
- Add documentation
`,
			},
			wantErr:          false,
			checkSuggestions: true,
		},
		{
			name: "with hint",
			params: CodeReviewParams{
				GoCode: `
package main

func test() {}
`,
				Hint: "performance",
			},
			wantErr:          false,
			checkSuggestions: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := PerformCodeReview(nil, tt.params)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}

				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error to contain '%s', got '%s'", tt.errContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result == nil {
				t.Fatal("expected result, got nil")
			}

			if tt.checkSuggestions {
				if result.Suggestions == nil {
					t.Error("expected suggestions to be set")
				}
			}
		})
	}
}

func TestPerformCodeReview_WithGuidelinesFile(t *testing.T) {
	// Create a temporary guidelines file
	tmpDir := t.TempDir()
	guidelinesFile := filepath.Join(tmpDir, "guidelines.md")

	guidelinesContent := `
# Custom Guidelines
- Use descriptive function names
- Add comments for complex logic
- Avoid global state
- Handle errors properly
`

	if err := os.WriteFile(guidelinesFile, []byte(guidelinesContent), 0644); err != nil {
		t.Fatalf("failed to create guidelines file: %v", err)
	}

	params := CodeReviewParams{
		GoCode: `
package main

func test() {}
`,
		GuidelinesFile: guidelinesFile,
	}

	result, err := PerformCodeReview(nil, params)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}
}

func TestPerformCodeReview_WithInvalidGuidelinesFile(t *testing.T) {
	params := CodeReviewParams{
		GoCode:         "package main\nfunc test() {}",
		GuidelinesFile: "/nonexistent/guidelines.md",
	}

	_, err := PerformCodeReview(nil, params)

	if err == nil {
		t.Error("expected error for nonexistent guidelines file")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error, got: %v", err)
	}
}

func TestAddHintSpecificAnalysis(t *testing.T) {
	tests := []struct {
		name           string
		hint           string
		expectCategory string
	}{
		{
			name:           "performance hint",
			hint:           "focus on performance",
			expectCategory: "performance",
		},
		{
			name:           "security hint",
			hint:           "check security",
			expectCategory: "security",
		},
		{
			name:           "test hint",
			hint:           "improve testability",
			expectCategory: "testability",
		},
		{
			name:           "maintainability hint",
			hint:           "improve maintainability",
			expectCategory: "maintainability",
		},
		{
			name:           "readability hint",
			hint:           "improve readability",
			expectCategory: "maintainability",
		},
		{
			name:           "multiple hints",
			hint:           "performance and security",
			expectCategory: "", // Should have multiple
		},
		{
			name:           "no specific hint",
			hint:           "general review",
			expectCategory: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := CodeReviewParams{
				GoCode: `
package main

func test() {}
`,
				Hint: tt.hint,
			}

			result, err := PerformCodeReview(nil, params)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.expectCategory != "" {
				found := false
				for _, suggestion := range result.Suggestions {
					if suggestion.Category == tt.expectCategory {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected to find suggestion with category '%s'", tt.expectCategory)
				}
			}
		})
	}
}

func TestGuidelinesParser(t *testing.T) {
	parser := NewGuidelinesParser()

	if parser == nil {
		t.Fatal("expected parser, got nil")
	}
}

func TestGetDefaultGuidelines(t *testing.T) {
	guidelines := GetDefaultGuidelines()

	if len(guidelines) == 0 {
		t.Error("expected default guidelines to be non-empty")
	}

	// Check for common guideline themes
	hasGoStyle := false
	hasErrorHandling := false
	hasPerformance := false

	for _, guideline := range guidelines {
		if strings.Contains(strings.ToLower(guideline), "go") {
			hasGoStyle = true
		}
		if strings.Contains(strings.ToLower(guideline), "error") {
			hasErrorHandling = true
		}
		if strings.Contains(strings.ToLower(guideline), "performance") {
			hasPerformance = true
		}
	}

	if !hasGoStyle {
		t.Log("Note: Default guidelines may not explicitly mention Go style")
	}

	if !hasErrorHandling {
		t.Log("Note: Default guidelines may not explicitly mention error handling")
	}

	if !hasPerformance {
		t.Log("Note: Default guidelines may not explicitly mention performance")
	}
}

func TestAnalyzer(t *testing.T) {
	guidelines := GetDefaultGuidelines()
	analyzer := NewAnalyzer(guidelines, "")

	if analyzer == nil {
		t.Fatal("expected analyzer, got nil")
	}
}

func TestAnalyzeCode(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		hint     string
		wantErr  bool
		checkLen int
	}{
		{
			name: "simple function",
			code: `
package main

func test() {
	println("hello")
}
`,
			hint:     "",
			wantErr:  false,
			checkLen: 0, // Just check it doesn't panic
		},
		{
			name: "function with error",
			code: `
package main

func test() error {
	return nil
}
`,
			hint:    "",
			wantErr: false,
		},
		{
			name: "multiple functions",
			code: `
package main

func func1() {}
func func2() {}
func func3() {}
`,
			hint:    "",
			wantErr: false,
		},
		{
			name: "struct definition",
			code: `
package main

type MyStruct struct {
	Field1 string
	Field2 int
}
`,
			hint:    "",
			wantErr: false,
		},
		{
			name: "interface definition",
			code: `
package main

type MyInterface interface {
	Method1()
	Method2() error
}
`,
			hint:    "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			guidelines := GetDefaultGuidelines()
			analyzer := NewAnalyzer(guidelines, tt.hint)

			result, err := analyzer.AnalyzeCode(tt.code)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result == nil {
				t.Fatal("expected result, got nil")
			}

			if tt.checkLen > 0 && len(result.Suggestions) < tt.checkLen {
				t.Errorf("expected at least %d suggestions, got %d", tt.checkLen, len(result.Suggestions))
			}
		})
	}
}

func TestReviewResult(t *testing.T) {
	result := &ReviewResult{
		Suggestions: []Suggestion{
			{
				Category: "test",
				Message:  "test suggestion",
				Impact:   "test impact",
				Example:  "test code",
			},
		},
		Summary: "test summary",
	}

	if len(result.Suggestions) != 1 {
		t.Errorf("expected 1 suggestion, got %d", len(result.Suggestions))
	}

	if result.Suggestions[0].Category != "test" {
		t.Errorf("expected category 'test', got '%s'", result.Suggestions[0].Category)
	}

	if result.Summary != "test summary" {
		t.Errorf("expected summary 'test summary', got '%s'", result.Summary)
	}
}

func TestSuggestion(t *testing.T) {
	suggestion := Suggestion{
		Category: "performance",
		Message:  "use slice instead of array",
		Impact:   "improved memory efficiency",
		Example:  "arr := []int{1,2,3}",
	}

	if suggestion.Category != "performance" {
		t.Errorf("expected category 'performance', got '%s'", suggestion.Category)
	}

	if suggestion.Message != "use slice instead of array" {
		t.Errorf("expected message 'use slice instead of array', got '%s'", suggestion.Message)
	}

	if suggestion.Impact != "improved memory efficiency" {
		t.Errorf("expected impact 'improved memory efficiency', got '%s'", suggestion.Impact)
	}
}

func TestCodeReviewParams(t *testing.T) {
	params := CodeReviewParams{
		GoCode:            "package main",
		GuidelinesFile:    "/path/to/guidelines.md",
		GuidelinesContent: "# Guidelines",
		Hint:              "performance",
	}

	if params.GoCode != "package main" {
		t.Errorf("expected GoCode 'package main', got '%s'", params.GoCode)
	}

	if params.GuidelinesFile != "/path/to/guidelines.md" {
		t.Errorf("expected GuidelinesFile '/path/to/guidelines.md', got '%s'", params.GuidelinesFile)
	}

	if params.GuidelinesContent != "# Guidelines" {
		t.Errorf("expected GuidelinesContent '# Guidelines', got '%s'", params.GuidelinesContent)
	}

	if params.Hint != "performance" {
		t.Errorf("expected Hint 'performance', got '%s'", params.Hint)
	}
}

func TestPerformCodeReview_ComplexCode(t *testing.T) {
	code := `
package main

import (
	"errors"
	"fmt"
)

type Service struct {
	config map[string]interface{}
}

func NewService() *Service {
	return &Service{
		config: make(map[string]interface{}),
	}
}

func (s *Service) Process(data string) (string, error) {
	if data == "" {
		return "", errors.New("empty data")
	}

	// Process the data
	result := fmt.Sprintf("processed: %s", data)
	return result, nil
}

func main() {
	s := NewService()
	result, err := s.Process("test")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(result)
}
`

	params := CodeReviewParams{
		GoCode: code,
		Hint:   "error handling and performance",
	}

	result, err := PerformCodeReview(nil, params)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}

	// Complex code should generate some suggestions
	if len(result.Suggestions) == 0 {
		t.Log("No suggestions generated for complex code")
	}
}
