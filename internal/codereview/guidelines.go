package codereview

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

// GuidelinesParser parses markdown guidelines
type GuidelinesParser struct{}

// NewGuidelinesParser creates a new guidelines parser
func NewGuidelinesParser() *GuidelinesParser {
	return &GuidelinesParser{}
}

// ParseFile parses guidelines from a markdown file
func (p *GuidelinesParser) ParseFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var guidelines []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if guideline := p.extractGuideline(line); guideline != "" {
			guidelines = append(guidelines, guideline)
		}
	}

	return guidelines, scanner.Err()
}

// ParseContent parses guidelines from markdown content
func (p *GuidelinesParser) ParseContent(content string) []string {
	var guidelines []string
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if guideline := p.extractGuideline(line); guideline != "" {
			guidelines = append(guidelines, guideline)
		}
	}

	return guidelines
}

// extractGuideline extracts actionable guidelines from markdown lines
func (p *GuidelinesParser) extractGuideline(line string) string {
	// Remove markdown formatting
	line = strings.TrimSpace(line)

	// Skip empty lines
	if line == "" {
		return ""
	}

	// Skip markdown headers
	if strings.HasPrefix(line, "#") {
		return ""
	}

	// Skip code blocks
	if strings.HasPrefix(line, "```") {
		return ""
	}

	// Extract from bullet points
	if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
		return strings.TrimSpace(line[2:])
	}

	// Extract from numbered lists
	if matched := numberListRegex.MatchString(line); matched {
		parts := numberListRegex.Split(line, 2)
		if len(parts) > 1 {
			return strings.TrimSpace(parts[1])
		}
	}

	// Extract from guidelines that contain specific keywords
	lowerLine := strings.ToLower(line)
	if containsGuidelineKeywords(lowerLine) {
		return line
	}

	return ""
}

// containsGuidelineKeywords checks if a line contains guideline-related keywords
func containsGuidelineKeywords(line string) bool {
	keywords := []string{
		"should", "must", "avoid", "prefer", "use", "don't", "always", "never",
		"ensure", "make sure", "consider", "recommended", "best practice",
		"convention", "standard", "guideline", "rule", "requirement",
	}

	for _, keyword := range keywords {
		if strings.Contains(line, keyword) {
			return true
		}
	}

	return false
}

// GetDefaultGuidelines returns a set of standard Go best practices
func GetDefaultGuidelines() []string {
	return []string{
		"Use gofmt to format code consistently",
		"Follow standard Go naming conventions (camelCase for unexported, PascalCase for exported)",
		"Document all exported functions, types, and variables",
		"Handle errors explicitly, don't ignore them",
		"Use interfaces to define behavior",
		"Prefer small, focused functions over large ones",
		"Avoid deep nesting and complex conditionals",
		"Use meaningful variable and function names",
		"Organize code into logical packages",
		"Write tests for public APIs",
		"Use context.Context for cancellation and timeouts",
		"Prefer composition over inheritance",
		"Use channels for communication between goroutines",
		"Avoid global variables when possible",
		"Use defer for cleanup operations",
		"Validate input parameters",
		"Use appropriate data types and avoid type conversions",
		"Follow the principle of least privilege",
		"Keep dependencies minimal",
		"Use build tags for conditional compilation",
	}
}

// Compile regex pattern for numbered lists
var numberListRegex = regexp.MustCompile(`^\d+\.\s+`)
