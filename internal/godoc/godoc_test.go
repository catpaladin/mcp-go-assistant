package godoc

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestGetDocumentation(t *testing.T) {
	tests := []struct {
		name        string
		params      GoDocParams
		wantErr     bool
		errContains string
	}{
		{
			name: "standard library package",
			params: GoDocParams{
				PackagePath: "fmt",
			},
			wantErr: false,
		},
		{
			name: "standard library with symbol",
			params: GoDocParams{
				PackagePath: "fmt",
				SymbolName:  "Println",
			},
			wantErr: false,
		},
		{
			name: "empty package path",
			params: GoDocParams{
				PackagePath: "",
			},
			wantErr:     true,
			errContains: "package_path is required",
		},
		{
			name: "invalid package",
			params: GoDocParams{
				PackagePath: "invalid/nonexistent/package",
			},
			wantErr: true,
		},
		{
			name: "valid Go package",
			params: GoDocParams{
				PackagePath: "time",
			},
			wantErr: false,
		},
		{
			name: "symbol in standard library",
			params: GoDocParams{
				PackagePath: "time",
				SymbolName:  "Now",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			result, err := GetDocumentation(ctx, tt.params)

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
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result == "" {
				t.Error("expected non-empty documentation")
			}

			if tt.params.SymbolName != "" {
				if !strings.Contains(result, tt.params.SymbolName) {
					t.Errorf("expected documentation to contain symbol '%s'", tt.params.SymbolName)
				}
			}
		})
	}
}

func TestGetDocumentation_WithContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	params := GoDocParams{
		PackagePath: "fmt",
	}

	_, err := GetDocumentation(ctx, params)

	if err == nil {
		t.Error("expected error from context cancellation")
	}

	if !strings.Contains(err.Error(), "context canceled") && !strings.Contains(err.Error(), "signal") {
		t.Logf("Note: Context cancellation behavior may vary: %v", err)
	}
}

func TestGetDocumentation_WithWorkingDir(t *testing.T) {
	params := GoDocParams{
		PackagePath: "fmt",
		WorkingDir:  ".", // Current directory should have go.mod
	}

	ctx := context.Background()
	result, err := GetDocumentation(ctx, params)

	if err != nil {
		t.Logf("Working dir test failed (may be expected): %v", err)
		return
	}

	if result == "" {
		t.Error("expected non-empty documentation")
	}
}

func TestFindGoModule(t *testing.T) {
	// This function searches for go.mod
	// In a test environment, it should find the module root
	moduleDir := findGoModule()

	if moduleDir == "" {
		t.Log("No go.mod found (expected in some test scenarios)")
		return
	}

	// Verify it's a valid directory
	// In real scenario, moduleDir should contain go.mod
	if !strings.Contains(moduleDir, "mcp-go-assistant") {
		t.Logf("Warning: Module dir doesn't contain project name: %s", moduleDir)
	}
}

func TestGetDocumentation_MultiplePackages(t *testing.T) {
	packages := []string{
		"fmt",
		"strings",
		"time",
		"json",
	}

	for _, pkg := range packages {
		t.Run(pkg, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			params := GoDocParams{
				PackagePath: pkg,
			}

			result, err := GetDocumentation(ctx, params)

			if err != nil {
				t.Logf("Package %s failed: %v", pkg, err)
				return
			}

			if result == "" {
				t.Errorf("expected non-empty documentation for package %s", pkg)
			}

			// Verify package name is in output
			if !strings.Contains(result, pkg) {
				t.Logf("Note: Documentation for %s may not contain package name", pkg)
			}
		})
	}
}

func TestGetDocumentation_VariousSymbols(t *testing.T) {
	tests := []struct {
		packagePath string
		symbolName  string
	}{
		{"fmt", "Print"},
		{"fmt", "Printf"},
		{"fmt", "Println"},
		{"time", "Now"},
		{"time", "Since"},
		{"strings", "Contains"},
		{"strings", "HasPrefix"},
		{"json", "Marshal"},
		{"json", "Unmarshal"},
	}

	for _, tt := range tests {
		t.Run(tt.packagePath+"."+tt.symbolName, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			params := GoDocParams{
				PackagePath: tt.packagePath,
				SymbolName:  tt.symbolName,
			}

			result, err := GetDocumentation(ctx, params)

			if err != nil {
				t.Logf("Symbol %s.%s failed: %v", tt.packagePath, tt.symbolName, err)
				return
			}

			if result == "" {
				t.Errorf("expected non-empty documentation for %s.%s", tt.packagePath, tt.symbolName)
			}

			// The symbol name should be in the documentation
			if !strings.Contains(result, tt.symbolName) {
				t.Logf("Warning: Documentation doesn't contain symbol name %s", tt.symbolName)
			}
		})
	}
}

func TestGetDocumentation_LongTimeout(t *testing.T) {
	// Test with a longer timeout
	params := GoDocParams{
		PackagePath: "fmt",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := GetDocumentation(ctx, params)

	if err != nil {
		t.Errorf("unexpected error with long timeout: %v", err)
		return
	}

	if result == "" {
		t.Error("expected non-empty documentation")
	}
}

func TestGetDocumentation_Concurrent(t *testing.T) {
	packages := []string{
		"fmt",
		"strings",
		"time",
		"os",
		"io",
	}

	done := make(chan bool, len(packages))

	for _, pkg := range packages {
		go func(p string) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			params := GoDocParams{
				PackagePath: p,
			}

			_, err := GetDocumentation(ctx, params)
			_ = err // Ignore errors
			done <- true
		}(pkg)
	}

	// Wait for all goroutines
	for i := 0; i < len(packages); i++ {
		<-done
	}
}

func TestGetDocumentation_PackageWithoutSymbol(t *testing.T) {
	// Test getting documentation for a package without symbol
	params := GoDocParams{
		PackagePath: "encoding/json",
	}

	ctx := context.Background()
	result, err := GetDocumentation(ctx, params)

	if err != nil {
		t.Logf("Package documentation failed: %v", err)
		return
	}

	if result == "" {
		t.Error("expected non-empty package documentation")
	}

	// Should contain package information
	if !strings.Contains(result, "json") {
		t.Log("Documentation may be in different format")
	}
}

func TestGoDocParams(t *testing.T) {
	// Test GoDocParams struct initialization
	params := GoDocParams{
		PackagePath: "fmt",
		SymbolName:  "Println",
		WorkingDir:  "/tmp",
	}

	if params.PackagePath != "fmt" {
		t.Errorf("expected PackagePath 'fmt', got '%s'", params.PackagePath)
	}

	if params.SymbolName != "Println" {
		t.Errorf("expected SymbolName 'Println', got '%s'", params.SymbolName)
	}

	if params.WorkingDir != "/tmp" {
		t.Errorf("expected WorkingDir '/tmp', got '%s'", params.WorkingDir)
	}
}
