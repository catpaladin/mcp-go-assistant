package testgen

// TestGenParams represents the parameters for the test generation tool
type TestGenParams struct {
	GoCode      string `json:"go_code" jsonschema:"description:The Go code to generate tests for"`
	PackageName string `json:"package_name,omitempty" jsonschema:"description:Package name for generated tests (defaults to package_test)"`
	Focus       string `json:"focus,omitempty" jsonschema:"description:Focus area: 'interfaces' for interface extraction and mocks or 'unit' for unit tests or 'table' for table-driven tests"`
}

// TestGenResult represents the result of test generation
type TestGenResult struct {
	TestCode    string      `json:"test_code"`
	MockCode    string      `json:"mock_code,omitempty"`
	Interfaces  []Interface `json:"interfaces,omitempty"`
	Suggestions []string    `json:"suggestions,omitempty"`
}

// Interface represents an extracted or generated interface
type Interface struct {
	Name    string   `json:"name"`
	Methods []Method `json:"methods"`
	ForType string   `json:"for_type,omitempty"`
}

// Method represents a method signature
type Method struct {
	Name    string `json:"name"`
	Params  string `json:"params"`
	Returns string `json:"returns"`
}

// String returns the test code as a string
func (r *TestGenResult) String() string {
	result := r.TestCode
	if r.MockCode != "" {
		result += "\n\n// --- Mock Implementations ---\n\n" + r.MockCode
	}
	return result
}
