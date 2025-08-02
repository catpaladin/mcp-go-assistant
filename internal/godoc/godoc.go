package godoc

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// GoDocParams represents the parameters for the go-doc tool
type GoDocParams struct {
	PackagePath string `json:"package_path" jsonschema:"description:The Go package path to query documentation for"`
	SymbolName  string `json:"symbol_name,omitempty" jsonschema:"description:Optional symbol name within the package to get specific documentation"`
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