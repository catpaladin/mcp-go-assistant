package codereview

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// PerformCodeReview analyzes Go code and returns improvement suggestions
func PerformCodeReview(ctx context.Context, params CodeReviewParams) (*ReviewResult, error) {
	// Validate input
	if params.GoCode == "" {
		return nil, fmt.Errorf("go_code parameter is required")
	}

	// Parse guidelines
	var guidelines []string
	parser := NewGuidelinesParser()

	// Load guidelines from file if provided
	if params.GuidelinesFile != "" {
		if _, err := os.Stat(params.GuidelinesFile); err == nil {
			fileGuidelines, err := parser.ParseFile(params.GuidelinesFile)
			if err != nil {
				return nil, fmt.Errorf("failed to parse guidelines file: %v", err)
			}
			guidelines = append(guidelines, fileGuidelines...)
		} else {
			return nil, fmt.Errorf("guidelines file not found: %s", params.GuidelinesFile)
		}
	}

	// Parse guidelines from content if provided
	if params.GuidelinesContent != "" {
		contentGuidelines := parser.ParseContent(params.GuidelinesContent)
		guidelines = append(guidelines, contentGuidelines...)
	}

	// Add default guidelines if none provided
	if len(guidelines) == 0 {
		guidelines = GetDefaultGuidelines()
	}

	// Create analyzer with guidelines and hint
	analyzer := NewAnalyzer(guidelines, params.Hint)

	// Perform analysis
	result, err := analyzer.AnalyzeCode(params.GoCode)
	if err != nil {
		return nil, fmt.Errorf("code analysis failed: %v", err)
	}

	// Add hint-specific analysis if provided
	if params.Hint != "" {
		addHintSpecificAnalysis(result, params.Hint)
	}

	return result, nil
}

// addHintSpecificAnalysis adds analysis based on the provided hint
func addHintSpecificAnalysis(result *ReviewResult, hint string) {
	hintLower := strings.ToLower(hint)
	
	if strings.Contains(hintLower, "performance") {
		result.Suggestions = append(result.Suggestions, Suggestion{
			Category: "performance",
			Message:  "Focus on performance: Look for opportunities to optimize algorithms, reduce allocations, and improve efficiency",
			Impact:   "Better runtime performance and resource utilization",
		})
	}
	
	if strings.Contains(hintLower, "security") {
		result.Suggestions = append(result.Suggestions, Suggestion{
			Category: "security",
			Message:  "Focus on security: Validate inputs, handle sensitive data carefully, and avoid common vulnerabilities",
			Impact:   "Improved application security and reduced attack surface",
		})
	}
	
	if strings.Contains(hintLower, "test") {
		result.Suggestions = append(result.Suggestions, Suggestion{
			Category: "testability",
			Message:  "Focus on testability: Make functions pure, inject dependencies, and avoid global state",
			Impact:   "Easier testing and better code reliability",
		})
	}
	
	if strings.Contains(hintLower, "maintainability") || strings.Contains(hintLower, "readability") {
		result.Suggestions = append(result.Suggestions, Suggestion{
			Category: "maintainability",
			Message:  "Focus on maintainability: Use clear naming, add documentation, and simplify complex logic",
			Impact:   "Easier code maintenance and team collaboration",
		})
	}
}