package testgen

import (
	"context"
	"strings"
	"testing"
)

func TestGenerateTests(t *testing.T) {
	tests := []struct {
		name        string
		params      TestGenParams
		wantErr     bool
		errContains string
		checkTest   bool
		checkMock   bool
	}{
		{
			name: "empty go code",
			params: TestGenParams{
				GoCode: "",
			},
			wantErr:     true,
			errContains: "go_code parameter is required",
		},
		{
			name: "valid simple function",
			params: TestGenParams{
				GoCode: `package main

func TestSomething(t *testing.T) {
	// test
}
`,
			},
			wantErr:   false,
			checkTest: true,
		},
		{
			name: "with package name",
			params: TestGenParams{
				GoCode:      `package main`,
				PackageName: "custom_test",
			},
			wantErr:   false,
			checkTest: true,
		},
		{
			name: "with interface focus",
			params: TestGenParams{
				GoCode: `package main

type MyStruct struct{}

func (m *MyStruct) Method1() {}
func (m *MyStruct) Method2() error { return nil }
`,
				Focus: "interfaces",
			},
			wantErr:   false,
			checkMock: true,
		},
		{
			name: "with table focus",
			params: TestGenParams{
				GoCode: `package main

func TestFunction() {
}
`,
				Focus: "table",
			},
			wantErr:   false,
			checkTest: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GenerateTests(context.TODO(), tt.params)

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

			if tt.checkTest && result.TestCode == "" {
				t.Error("expected test code to be generated")
			}

			if tt.checkMock && result.MockCode == "" {
				t.Error("expected mock code to be generated")
			}
		})
	}
}

func TestGenerateTests_DefaultPackageName(t *testing.T) {
	code := `package testpackage`

	params := TestGenParams{
		GoCode: code,
	}

	result, err := GenerateTests(context.TODO(), params)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}

	// Default package name should be testpackage_test
	if !strings.Contains(result.TestCode, "package testpackage_test") {
		t.Errorf("expected default package name 'testpackage_test' in test code")
	}
}

func TestGenerateInterfacesAndMocks(t *testing.T) {
	code := `package main

type Service struct{}

func (s *Service) Method1() {}
func (s *Service) Method2(a string) int { return 0 }
func (s *Service) Method3(a, b int) (string, error) { return "", nil }
`

	params := TestGenParams{
		GoCode:      code,
		PackageName: "testpkg",
		Focus:       "interfaces",
	}

	result, err := GenerateTests(context.TODO(), params)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}

	// Should generate interfaces
	if len(result.Interfaces) == 0 {
		t.Error("expected interfaces to be generated")
	}

	// Check for Service interface
	foundService := false
	for _, iface := range result.Interfaces {
		if iface.ForType == "Service" {
			foundService = true
			if len(iface.Methods) != 3 {
				t.Errorf("expected 3 methods for Service interface, got %d", len(iface.Methods))
			}
		}
	}

	if !foundService {
		t.Error("expected Service interface to be found")
	}

	// Test code should contain interface definition
	if !strings.Contains(result.TestCode, "type Serviceer interface") {
		t.Error("expected Serviceer interface in test code")
	}

	// Mock code should be generated
	if result.MockCode == "" {
		t.Error("expected mock code to be generated")
	}

	if !strings.Contains(result.MockCode, "type MockService struct") {
		t.Error("expected MockService in mock code")
	}
}

func TestGenerateUnitTests(t *testing.T) {
	code := `package main

func Function1() {}
func Function2() error { return nil }
func Function3(a string) int { return 0 }
`

	params := TestGenParams{
		GoCode:      code,
		PackageName: "testpkg",
	}

	result, err := GenerateTests(context.TODO(), params)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}

	// Test code should be generated
	if result.TestCode == "" {
		t.Error("expected test code to be generated")
	}

	// Should generate tests for exported functions
	if !strings.Contains(result.TestCode, "func TestFunction1") {
		t.Error("expected TestFunction1 in test code")
	}

	if !strings.Contains(result.TestCode, "func TestFunction2") {
		t.Error("expected TestFunction2 in test code")
	}

	if !strings.Contains(result.TestCode, "func TestFunction3") {
		t.Error("expected TestFunction3 in test code")
	}
}

func TestGenerateTableDrivenTests(t *testing.T) {
	code := `package main

func CalculateSum(a, b int) int {
	return a + b
}

func ProcessData(s string) error {
	if s == "" {
		return nil
	}
	return nil
}
`

	params := TestGenParams{
		GoCode:      code,
		PackageName: "testpkg",
		Focus:       "table",
	}

	result, err := GenerateTests(context.TODO(), params)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}

	if result.TestCode == "" {
		t.Error("expected test code to be generated")
	}

	// Should contain table-driven test structure
	if !strings.Contains(result.TestCode, "tests := []struct") {
		t.Error("expected table-driven test structure")
	}

	if !strings.Contains(result.TestCode, "for _, tt := range tests") {
		t.Error("expected range loop over tests")
	}
}

func TestGenerateTests_NoExportedFunctions(t *testing.T) {
	code := `package main

func unexportedFunction() {}
`

	params := TestGenParams{
		GoCode: code,
	}

	result, err := GenerateTests(context.TODO(), params)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}

	// Should have suggestions since no exported functions
	if len(result.Suggestions) == 0 {
		t.Error("expected suggestions for no exported functions")
	}
}

func TestGenerateTests_NoStructsWithMethods(t *testing.T) {
	code := `package main

type MyType struct {
	Field string
}
`

	params := TestGenParams{
		GoCode: code,
		Focus:  "interfaces",
	}

	result, err := GenerateTests(context.TODO(), params)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}

	// Should have suggestions since no interfaces can be generated
	if len(result.Suggestions) == 0 {
		t.Error("expected suggestions for no struct methods")
	}
}

func TestGetReceiverTypeName(t *testing.T) {
	code := `package main

type MyStruct struct{}

func (m MyStruct) Method1() {}
func (m *MyStruct) Method2() {}
`

	// Parse the code
	// This tests the helper function indirectly
	params := TestGenParams{
		GoCode: code,
		Focus:  "interfaces",
	}

	result, err := GenerateTests(context.TODO(), params)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}

	// Should identify MyStruct as having methods
	if len(result.Interfaces) == 0 {
		t.Error("expected interfaces for MyStruct")
	}
}

func TestFormatFieldList(t *testing.T) {
	code := `package main

func Function(a string, b int, c bool) error {
	return nil
}
`

	params := TestGenParams{
		GoCode: code,
		Focus:  "table",
	}

	result, err := GenerateTests(context.TODO(), params)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}

	// Test code should contain function parameters
	if !strings.Contains(result.TestCode, "a string") {
		t.Error("expected parameter 'a string' in test code")
	}

	if !strings.Contains(result.TestCode, "b int") {
		t.Error("expected parameter 'b int' in test code")
	}

	if !strings.Contains(result.TestCode, "c bool") {
		t.Error("expected parameter 'c bool' in test code")
	}
}

func TestFormatType(t *testing.T) {
	code := `package main

func Function(a []string, b map[string]int, c chan error) {
}
`

	params := TestGenParams{
		GoCode: code,
		Focus:  "table",
	}

	result, err := GenerateTests(context.TODO(), params)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}

	// Should handle complex types
	if !strings.Contains(result.TestCode, "a []string") {
		t.Error("expected 'a []string' in test code")
	}

	if !strings.Contains(result.TestCode, "b map[string]int") {
		t.Error("expected 'b map[string]int' in test code")
	}
}

func TestGenerateTests_CaseInsensitiveFocus(t *testing.T) {
	code := `package main

type Service struct{}

func (s *Service) Method() {}
`

	focuses := []string{"interfaces", "INTERFACES", "Interfaces"}

	for _, focus := range focuses {
		t.Run(focus, func(t *testing.T) {
			params := TestGenParams{
				GoCode: code,
				Focus:  focus,
			}

			result, err := GenerateTests(context.TODO(), params)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result == nil {
				t.Fatal("expected result, got nil")
			}

			// Should generate interfaces for any case
			if !strings.Contains(result.TestCode, "interface") {
				t.Error("expected interface generation for case-insensitive focus")
			}
		})
	}
}

func TestGenerateTestResult(t *testing.T) {
	result := &TestGenResult{
		Interfaces: []Interface{
			{
				Name:    "TestInterface",
				ForType: "TestType",
				Methods: []Method{
					{
						Name:    "Method1",
						Params:  "a string",
						Returns: "error",
					},
				},
			},
		},
		TestCode:    "test code content",
		MockCode:    "mock code content",
		Suggestions: []string{"suggestion 1", "suggestion 2"},
	}

	if len(result.Interfaces) != 1 {
		t.Errorf("expected 1 interface, got %d", len(result.Interfaces))
	}

	if result.Interfaces[0].Name != "TestInterface" {
		t.Errorf("expected interface name 'TestInterface', got '%s'", result.Interfaces[0].Name)
	}

	if result.TestCode != "test code content" {
		t.Errorf("expected test code 'test code content', got '%s'", result.TestCode)
	}

	if result.MockCode != "mock code content" {
		t.Errorf("expected mock code 'mock code content', got '%s'", result.MockCode)
	}

	if len(result.Suggestions) != 2 {
		t.Errorf("expected 2 suggestions, got %d", len(result.Suggestions))
	}
}

func TestInterface(t *testing.T) {
	iface := Interface{
		Name:    "MyInterface",
		ForType: "MyType",
		Methods: []Method{
			{
				Name:    "Method1",
				Params:  "param1 string",
				Returns: "error",
			},
			{
				Name:    "Method2",
				Params:  "",
				Returns: "",
			},
		},
	}

	if iface.Name != "MyInterface" {
		t.Errorf("expected name 'MyInterface', got '%s'", iface.Name)
	}

	if iface.ForType != "MyType" {
		t.Errorf("expected ForType 'MyType', got '%s'", iface.ForType)
	}

	if len(iface.Methods) != 2 {
		t.Errorf("expected 2 methods, got %d", len(iface.Methods))
	}
}

func TestMethod(t *testing.T) {
	method := Method{
		Name:    "TestMethod",
		Params:  "a int, b string",
		Returns: "error",
	}

	if method.Name != "TestMethod" {
		t.Errorf("expected name 'TestMethod', got '%s'", method.Name)
	}

	if method.Params != "a int, b string" {
		t.Errorf("expected params 'a int, b string', got '%s'", method.Params)
	}

	if method.Returns != "error" {
		t.Errorf("expected returns 'error', got '%s'", method.Returns)
	}
}
