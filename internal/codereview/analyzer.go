package codereview

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"unicode"
)

// Analyzer performs Go code analysis
type Analyzer struct {
	fset       *token.FileSet
	guidelines []string
	hint       string
}

// NewAnalyzer creates a new code analyzer
func NewAnalyzer(guidelines []string, hint string) *Analyzer {
	return &Analyzer{
		fset:       token.NewFileSet(),
		guidelines: guidelines,
		hint:       hint,
	}
}

// AnalyzeCode performs comprehensive Go code analysis
func (a *Analyzer) AnalyzeCode(code string) (*ReviewResult, error) {
	// Parse the Go code
	file, err := parser.ParseFile(a.fset, "", code, parser.ParseComments)
	if err != nil {
		return &ReviewResult{
			Summary: "Code parsing failed",
			Issues: []Issue{{
				Type:     "error",
				Category: "syntax",
				Message:  "Failed to parse Go code: " + err.Error(),
				Severity: "critical",
				Rule:     "valid-syntax",
			}},
			Score: 0,
		}, nil
	}

	result := &ReviewResult{
		Issues:      []Issue{},
		Suggestions: []Suggestion{},
		Metrics:     a.calculateMetrics(file, code),
	}

	// Perform various checks
	a.checkNaming(file, result)
	a.checkStructure(file, result)
	a.checkComments(file, result)
	a.checkErrorHandling(file, result)
	a.checkPerformance(file, result)
	a.checkSecurity(file, result)
	a.checkTestability(file, result)
	a.checkComplexity(file, result)
	a.applyCustomGuidelines(file, result)

	// Calculate overall score
	result.Score = a.calculateScore(result)
	result.Summary = a.generateSummary(result)

	return result, nil
}

// checkNaming verifies Go naming conventions
func (a *Analyzer) checkNaming(file *ast.File, result *ReviewResult) {
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			if node.Name.IsExported() && !isCapitalized(node.Name.Name) {
				result.Issues = append(result.Issues, Issue{
					Type:       "warning",
					Category:   "naming",
					Line:       a.getLine(node.Pos()),
					Message:    "Exported function should start with capital letter",
					Suggestion: "Rename to " + capitalize(node.Name.Name),
					Severity:   "medium",
					Rule:       "exported-naming",
				})
			}
			if containsUnderscore(node.Name.Name) {
				result.Issues = append(result.Issues, Issue{
					Type:       "style",
					Category:   "naming",
					Line:       a.getLine(node.Pos()),
					Message:    "Function names should use camelCase, not underscores",
					Suggestion: "Use camelCase naming convention",
					Severity:   "low",
					Rule:       "camel-case",
				})
			}
		case *ast.TypeSpec:
			if node.Name.IsExported() && !isCapitalized(node.Name.Name) {
				result.Issues = append(result.Issues, Issue{
					Type:       "warning",
					Category:   "naming",
					Line:       a.getLine(node.Pos()),
					Message:    "Exported type should start with capital letter",
					Suggestion: "Rename to " + capitalize(node.Name.Name),
					Severity:   "medium",
					Rule:       "exported-naming",
				})
			}
		case *ast.GenDecl:
			for _, spec := range node.Specs {
				if vs, ok := spec.(*ast.ValueSpec); ok {
					for _, name := range vs.Names {
						if name.IsExported() && !isCapitalized(name.Name) {
							result.Issues = append(result.Issues, Issue{
								Type:       "warning",
								Category:   "naming",
								Line:       a.getLine(name.Pos()),
								Message:    "Exported variable should start with capital letter",
								Suggestion: "Rename to " + capitalize(name.Name),
								Severity:   "medium",
								Rule:       "exported-naming",
							})
						}
					}
				}
			}
		}
		return true
	})
}

// checkStructure verifies code structure best practices
func (a *Analyzer) checkStructure(file *ast.File, result *ReviewResult) {
	functionCount := 0
	longFunctions := 0

	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			functionCount++
			if node.Body != nil {
				lineCount := a.getLine(node.End()) - a.getLine(node.Pos())
				if lineCount > 50 {
					longFunctions++
					result.Issues = append(result.Issues, Issue{
						Type:       "warning",
						Category:   "structure",
						Line:       a.getLine(node.Pos()),
						Message:    "Function is too long (>50 lines). Consider breaking it down",
						Suggestion: "Split into smaller, focused functions",
						Severity:   "medium",
						Rule:       "function-length",
					})
				}

				// Check for too many parameters
				if node.Type.Params != nil && len(node.Type.Params.List) > 5 {
					result.Issues = append(result.Issues, Issue{
						Type:       "warning",
						Category:   "structure",
						Line:       a.getLine(node.Pos()),
						Message:    "Function has too many parameters (>5)",
						Suggestion: "Consider using a struct to group related parameters",
						Severity:   "medium",
						Rule:       "parameter-count",
					})
				}
			}
		case *ast.StructType:
			if len(node.Fields.List) > 10 {
				result.Issues = append(result.Issues, Issue{
					Type:       "warning",
					Category:   "structure",
					Line:       a.getLine(node.Pos()),
					Message:    "Struct has many fields (>10). Consider if it's doing too much",
					Suggestion: "Consider breaking into smaller structs",
					Severity:   "low",
					Rule:       "struct-size",
				})
			}
		}
		return true
	})

	if functionCount > 20 {
		result.Suggestions = append(result.Suggestions, Suggestion{
			Category: "structure",
			Message:  "File contains many functions. Consider splitting into multiple files",
			Impact:   "Improved maintainability and organization",
		})
	}
}

// checkComments verifies documentation practices
func (a *Analyzer) checkComments(file *ast.File, result *ReviewResult) {
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			if node.Name.IsExported() && node.Doc == nil {
				result.Issues = append(result.Issues, Issue{
					Type:       "warning",
					Category:   "documentation",
					Line:       a.getLine(node.Pos()),
					Message:    "Exported function lacks documentation comment",
					Suggestion: "Add a comment starting with function name",
					Severity:   "medium",
					Rule:       "exported-docs",
				})
			}
		case *ast.TypeSpec:
			if node.Name.IsExported() {
				if genDecl, ok := findParentGenDecl(file, node); ok {
					if genDecl.Doc == nil {
						result.Issues = append(result.Issues, Issue{
							Type:       "warning",
							Category:   "documentation",
							Line:       a.getLine(node.Pos()),
							Message:    "Exported type lacks documentation comment",
							Suggestion: "Add a comment starting with type name",
							Severity:   "medium",
							Rule:       "exported-docs",
						})
					}
				}
			}
		}
		return true
	})
}

// checkErrorHandling verifies error handling patterns
func (a *Analyzer) checkErrorHandling(file *ast.File, result *ReviewResult) {
	ast.Inspect(file, func(n ast.Node) bool {
		if node, ok := n.(*ast.CallExpr); ok {
			// Check for ignored errors
			if parent, ok := findParentAssign(file, node); ok {
				if assign, ok := parent.(*ast.AssignStmt); ok {
					if len(assign.Lhs) > 1 {
						if ident, ok := assign.Lhs[len(assign.Lhs)-1].(*ast.Ident); ok {
							if ident.Name == "_" {
								result.Issues = append(result.Issues, Issue{
									Type:       "warning",
									Category:   "error-handling",
									Line:       a.getLine(node.Pos()),
									Message:    "Error is being ignored",
									Suggestion: "Handle error appropriately",
									Severity:   "high",
									Rule:       "error-handling",
								})
							}
						}
					}
				}
			}
		}
		return true
	})
}

// checkPerformance identifies performance issues
func (a *Analyzer) checkPerformance(file *ast.File, result *ReviewResult) {
	ast.Inspect(file, func(n ast.Node) bool {
		if node, ok := n.(*ast.RangeStmt); ok {
			// Check for string concatenation in loops
			ast.Inspect(node.Body, func(inner ast.Node) bool {
				if binExpr, ok := inner.(*ast.BinaryExpr); ok {
					if binExpr.Op == token.ADD {
						result.Issues = append(result.Issues, Issue{
							Type:       "warning",
							Category:   "performance",
							Line:       a.getLine(binExpr.Pos()),
							Message:    "String concatenation in loop can be inefficient",
							Suggestion: "Consider using strings.Builder or bytes.Buffer",
							Severity:   "medium",
							Rule:       "string-concatenation",
						})
					}
				}
				return true
			})
		}
		return true
	})
}

// checkSecurity identifies potential security issues
func (a *Analyzer) checkSecurity(file *ast.File, result *ReviewResult) {
	ast.Inspect(file, func(n ast.Node) bool {
		if node, ok := n.(*ast.CallExpr); ok {
			if sel, ok := node.Fun.(*ast.SelectorExpr); ok {
				if ident, ok := sel.X.(*ast.Ident); ok {
					// Check for unsafe operations
					if ident.Name == "unsafe" {
						result.Issues = append(result.Issues, Issue{
							Type:       "warning",
							Category:   "security",
							Line:       a.getLine(node.Pos()),
							Message:    "Use of unsafe package should be carefully reviewed",
							Suggestion: "Ensure unsafe operations are necessary and correct",
							Severity:   "high",
							Rule:       "unsafe-usage",
						})
					}
				}
			}
		}
		return true
	})
}

// checkTestability identifies testability issues
func (a *Analyzer) checkTestability(file *ast.File, result *ReviewResult) {
	hasGlobalVars := false

	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.GenDecl:
			if node.Tok == token.VAR {
				for _, spec := range node.Specs {
					if vs, ok := spec.(*ast.ValueSpec); ok {
						for _, name := range vs.Names {
							if name.IsExported() {
								hasGlobalVars = true
							}
						}
					}
				}
			}
		}
		return true
	})

	if hasGlobalVars {
		result.Suggestions = append(result.Suggestions, Suggestion{
			Category: "testability",
			Message:  "Global variables can make testing difficult",
			Example:  "Consider dependency injection or configuration structs",
			Impact:   "Improved testability and maintainability",
		})
	}
}

// checkComplexity analyzes cyclomatic complexity
func (a *Analyzer) checkComplexity(file *ast.File, result *ReviewResult) {
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			complexity := calculateCyclomaticComplexity(node)
			if complexity > 10 {
				result.Issues = append(result.Issues, Issue{
					Type:       "warning",
					Category:   "complexity",
					Line:       a.getLine(node.Pos()),
					Message:    "Function has high cyclomatic complexity",
					Suggestion: "Consider breaking down into smaller functions",
					Severity:   "medium",
					Rule:       "cyclomatic-complexity",
				})
			}
		}
		return true
	})
}

// applyCustomGuidelines applies user-provided guidelines
func (a *Analyzer) applyCustomGuidelines(file *ast.File, result *ReviewResult) {
	for _, guideline := range a.guidelines {
		// Simple pattern matching for custom guidelines
		if strings.Contains(strings.ToLower(guideline), "no panic") {
			ast.Inspect(file, func(n ast.Node) bool {
				if call, ok := n.(*ast.CallExpr); ok {
					if ident, ok := call.Fun.(*ast.Ident); ok {
						if ident.Name == "panic" {
							result.Issues = append(result.Issues, Issue{
								Type:       "warning",
								Category:   "custom",
								Line:       a.getLine(call.Pos()),
								Message:    "Custom guideline: avoid using panic",
								Suggestion: "Return an error instead",
								Severity:   "medium",
								Rule:       "custom-no-panic",
							})
						}
					}
				}
				return true
			})
		}
	}
}

// Helper functions

func (a *Analyzer) getLine(pos token.Pos) int {
	position := a.fset.Position(pos)
	return position.Line
}

func (a *Analyzer) calculateMetrics(file *ast.File, code string) Metrics {
	lines := strings.Split(code, "\n")
	functionCount := 0
	typeCount := 0
	maxComplexity := 0

	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			functionCount++
			complexity := calculateCyclomaticComplexity(node)
			if complexity > maxComplexity {
				maxComplexity = complexity
			}
		case *ast.TypeSpec:
			typeCount++
		}
		return true
	})

	maintainability := "high"
	if maxComplexity > 15 || functionCount > 30 {
		maintainability = "low"
	} else if maxComplexity > 10 || functionCount > 20 {
		maintainability = "medium"
	}

	return Metrics{
		LinesOfCode:          len(lines),
		CyclomaticComplexity: maxComplexity,
		FunctionCount:        functionCount,
		TypeCount:            typeCount,
		TestCoverage:         "unknown",
		Maintainability:      maintainability,
	}
}

func (a *Analyzer) calculateScore(result *ReviewResult) int {
	score := 100

	for _, issue := range result.Issues {
		switch issue.Severity {
		case "critical":
			score -= 20
		case "high":
			score -= 10
		case "medium":
			score -= 5
		case "low":
			score -= 2
		}
	}

	if score < 0 {
		score = 0
	}
	return score
}

func (a *Analyzer) generateSummary(result *ReviewResult) string {
	issueCount := len(result.Issues)
	suggestionCount := len(result.Suggestions)

	if issueCount == 0 {
		return "Code looks good! No major issues found."
	}

	return fmt.Sprintf("Found %d issues and %d suggestions. Overall score: %d/100",
		issueCount, suggestionCount, result.Score)
}

// Utility functions

func isCapitalized(name string) bool {
	if len(name) == 0 {
		return false
	}
	return unicode.IsUpper(rune(name[0]))
}

func capitalize(name string) string {
	if len(name) == 0 {
		return name
	}
	return strings.ToUpper(name[:1]) + name[1:]
}

func containsUnderscore(name string) bool {
	return strings.Contains(name, "_")
}

func calculateCyclomaticComplexity(fn *ast.FuncDecl) int {
	complexity := 1 // Base complexity

	ast.Inspect(fn, func(n ast.Node) bool {
		switch n.(type) {
		case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt, *ast.SwitchStmt,
			*ast.TypeSwitchStmt, *ast.CaseClause:
			complexity++
		}
		return true
	})

	return complexity
}

func findParentGenDecl(file *ast.File, target ast.Node) (*ast.GenDecl, bool) {
	var parent *ast.GenDecl
	ast.Inspect(file, func(n ast.Node) bool {
		if genDecl, ok := n.(*ast.GenDecl); ok {
			for _, spec := range genDecl.Specs {
				if spec == target {
					parent = genDecl
					return false
				}
			}
		}
		return true
	})
	return parent, parent != nil
}

func findParentAssign(file *ast.File, target ast.Node) (ast.Node, bool) {
	var parent ast.Node
	ast.Inspect(file, func(n ast.Node) bool {
		if assign, ok := n.(*ast.AssignStmt); ok {
			for _, rhs := range assign.Rhs {
				if rhs == target {
					parent = assign
					return false
				}
			}
		}
		return true
	})
	return parent, parent != nil
}
