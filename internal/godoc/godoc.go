package godoc

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GoDocParams represents the parameters for the go-doc tool
type GoDocParams struct {
	PackagePath string `json:"package_path" jsonschema:"description:The Go package path to query documentation for"`
	SymbolName  string `json:"symbol_name,omitempty" jsonschema:"description:Optional symbol name within the package to get specific documentation"`
	WorkingDir  string `json:"working_dir,omitempty" jsonschema:"description:Optional working directory with go.mod file for external package access"`
}

// GetDocumentation executes the go doc command and returns the documentation
func GetDocumentation(ctx context.Context, params GoDocParams) (string, error) {
	if params.PackagePath == "" {
		return "", fmt.Errorf("package_path is required")
	}

	// Build the go doc command
	args := []string{"doc"}

	if params.SymbolName != "" {
		// When symbol is specified, use format: go doc package.symbol
		args = append(args, fmt.Sprintf("%s.%s", params.PackagePath, params.SymbolName))
	} else {
		// When only package is specified
		args = append(args, params.PackagePath)
	}

	// Execute the command
	cmd := exec.CommandContext(ctx, "go", args...)

	// Set working directory to enable external package access
	workingDir := params.WorkingDir
	if workingDir == "" {
		// Try to find a go.mod file in current directory or parents
		workingDir = findGoModule()
	}
	if workingDir != "" {
		cmd.Dir = workingDir
	}

	output, err := cmd.CombinedOutput()

	if err != nil {
		// Include both error and output for better debugging
		outputStr := strings.TrimSpace(string(output))
		if outputStr != "" {
			return "", fmt.Errorf("go doc failed: %v\nOutput: %s", err, outputStr)
		}
		return "", fmt.Errorf("go doc failed: %v", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// findGoModule searches for a go.mod file starting from the current directory
// and walking up the directory tree
func findGoModule() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}

	dir := cwd
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached the root directory
			break
		}
		dir = parent
	}

	return ""
}
