package codereview

import "encoding/json"

// CodeReviewParams represents the parameters for the code-review tool
type CodeReviewParams struct {
	GoCode            string `json:"go_code" jsonschema:"description:The Go code content to analyze"`
	GuidelinesFile    string `json:"guidelines_file,omitempty" jsonschema:"description:Optional path to markdown file with coding guidelines"`
	GuidelinesContent string `json:"guidelines_content,omitempty" jsonschema:"description:Optional markdown content with coding guidelines"`
	Hint              string `json:"hint,omitempty" jsonschema:"description:Optional hint or specific focus area for the review"`
}

// ReviewResult represents the complete result of a code review
type ReviewResult struct {
	Summary     string       `json:"summary"`
	Issues      []Issue      `json:"issues"`
	Suggestions []Suggestion `json:"suggestions"`
	Score       int          `json:"score"` // 0-100 score
	Metrics     Metrics      `json:"metrics"`
}

// Issue represents a code issue found during review
type Issue struct {
	Type       string `json:"type"`       // "error", "warning", "style"
	Category   string `json:"category"`   // "naming", "structure", "performance", etc.
	Line       int    `json:"line"`       // Line number (0 if not specific)
	Column     int    `json:"column"`     // Column number (0 if not specific)
	Message    string `json:"message"`    // Description of the issue
	Suggestion string `json:"suggestion"` // How to fix it
	Severity   string `json:"severity"`   // "low", "medium", "high", "critical"
	Rule       string `json:"rule"`       // Which Go best practice rule
}

// Suggestion represents a general improvement suggestion
type Suggestion struct {
	Category string `json:"category"` // "performance", "readability", "maintainability"
	Message  string `json:"message"`  // Description of the suggestion
	Example  string `json:"example"`  // Code example if applicable
	Impact   string `json:"impact"`   // Expected impact of the change
}

// Metrics represents code quality metrics
type Metrics struct {
	LinesOfCode          int    `json:"lines_of_code"`
	CyclomaticComplexity int    `json:"cyclomatic_complexity"`
	FunctionCount        int    `json:"function_count"`
	TypeCount            int    `json:"type_count"`
	TestCoverage         string `json:"test_coverage"`   // "unknown" if not determinable
	Maintainability      string `json:"maintainability"` // "low", "medium", "high"
}

// String returns a formatted JSON string of the ReviewResult
func (r *ReviewResult) String() string {
	jsonData, _ := json.MarshalIndent(r, "", "  ")
	return string(jsonData)
}
