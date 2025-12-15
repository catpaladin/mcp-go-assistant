package testgen

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// GenerateTests analyzes Go code and generates test scaffolding
func GenerateTests(ctx context.Context, params TestGenParams) (*TestGenResult, error) {
	if params.GoCode == "" {
		return nil, fmt.Errorf("go_code parameter is required")
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", params.GoCode, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Go code: %v", err)
	}

	pkgName := params.PackageName
	if pkgName == "" {
		pkgName = file.Name.Name + "_test"
	}

	result := &TestGenResult{
		Interfaces:  []Interface{},
		Suggestions: []string{},
	}

	focus := strings.ToLower(params.Focus)

	switch focus {
	case "interfaces", "interface", "mock", "mocks":
		generateInterfacesAndMocks(file, pkgName, result)
	case "table", "table-driven":
		generateTableDrivenTests(file, pkgName, result)
	default:
		generateUnitTests(file, pkgName, result)
	}

	return result, nil
}

// generateInterfacesAndMocks extracts interfaces from concrete types and generates mocks
func generateInterfacesAndMocks(file *ast.File, pkgName string, result *TestGenResult) {
	var testCode strings.Builder
	var mockCode strings.Builder

	testCode.WriteString(fmt.Sprintf("package %s\n\n", pkgName))
	mockCode.WriteString(fmt.Sprintf("package %s\n\n", pkgName))

	// Find struct types and their methods
	typesMethods := make(map[string][]Method)
	structTypes := make(map[string]bool)

	// First pass: identify struct types
	ast.Inspect(file, func(n ast.Node) bool {
		if ts, ok := n.(*ast.TypeSpec); ok {
			if _, isStruct := ts.Type.(*ast.StructType); isStruct {
				structTypes[ts.Name.Name] = true
			}
		}
		return true
	})

	// Second pass: collect methods for each type
	ast.Inspect(file, func(n ast.Node) bool {
		if fd, ok := n.(*ast.FuncDecl); ok {
			if fd.Recv != nil && len(fd.Recv.List) > 0 {
				recvType := getReceiverTypeName(fd.Recv.List[0].Type)
				if recvType != "" && fd.Name.IsExported() {
					method := Method{
						Name:    fd.Name.Name,
						Params:  formatFieldList(fd.Type.Params),
						Returns: formatFieldList(fd.Type.Results),
					}
					typesMethods[recvType] = append(typesMethods[recvType], method)
				}
			}
		}
		return true
	})

	// Generate interfaces and mocks for types with methods
	for typeName, methods := range typesMethods {
		if len(methods) == 0 {
			continue
		}

		ifaceName := typeName + "er"
		if strings.HasSuffix(typeName, "er") {
			ifaceName = typeName + "Interface"
		}

		iface := Interface{
			Name:    ifaceName,
			Methods: methods,
			ForType: typeName,
		}
		result.Interfaces = append(result.Interfaces, iface)

		// Generate interface definition
		testCode.WriteString(fmt.Sprintf("// %s defines the interface for %s\n", ifaceName, typeName))
		testCode.WriteString(fmt.Sprintf("type %s interface {\n", ifaceName))
		for _, m := range methods {
			if m.Returns != "" {
				testCode.WriteString(fmt.Sprintf("\t%s(%s) %s\n", m.Name, m.Params, m.Returns))
			} else {
				testCode.WriteString(fmt.Sprintf("\t%s(%s)\n", m.Name, m.Params))
			}
		}
		testCode.WriteString("}\n\n")

		// Generate mock implementation
		mockName := "Mock" + typeName
		mockCode.WriteString(fmt.Sprintf("// %s is a mock implementation of %s\n", mockName, ifaceName))
		mockCode.WriteString(fmt.Sprintf("type %s struct {\n", mockName))
		for _, m := range methods {
			mockCode.WriteString(fmt.Sprintf("\t%sFunc func(%s)", m.Name, m.Params))
			if m.Returns != "" {
				mockCode.WriteString(fmt.Sprintf(" %s", m.Returns))
			}
			mockCode.WriteString("\n")
		}
		mockCode.WriteString("}\n\n")

		// Generate mock method implementations
		for _, m := range methods {
			mockCode.WriteString(fmt.Sprintf("func (m *%s) %s(%s)", mockName, m.Name, m.Params))
			if m.Returns != "" {
				mockCode.WriteString(fmt.Sprintf(" %s", m.Returns))
			}
			mockCode.WriteString(" {\n")
			if m.Returns != "" {
				mockCode.WriteString(fmt.Sprintf("\treturn m.%sFunc(%s)\n", m.Name, extractParamNames(m.Params)))
			} else {
				mockCode.WriteString(fmt.Sprintf("\tm.%sFunc(%s)\n", m.Name, extractParamNames(m.Params)))
			}
			mockCode.WriteString("}\n\n")
		}
	}

	if len(typesMethods) == 0 {
		result.Suggestions = append(result.Suggestions, "No exported methods found on struct types. Consider adding methods to generate interfaces.")
		testCode.WriteString("// No interfaces could be extracted from the provided code.\n")
		testCode.WriteString("// Ensure your code has struct types with exported methods.\n")
	}

	result.TestCode = testCode.String()
	result.MockCode = mockCode.String()
}

// generateUnitTests generates basic unit test scaffolding
func generateUnitTests(file *ast.File, pkgName string, result *TestGenResult) {
	var testCode strings.Builder

	testCode.WriteString(fmt.Sprintf("package %s\n\n", pkgName))
	testCode.WriteString("import (\n\t\"testing\"\n)\n\n")

	funcCount := 0
	ast.Inspect(file, func(n ast.Node) bool {
		if fd, ok := n.(*ast.FuncDecl); ok {
			if fd.Name.IsExported() && fd.Recv == nil {
				funcCount++
				testCode.WriteString(fmt.Sprintf("func Test%s(t *testing.T) {\n", fd.Name.Name))
				testCode.WriteString("\t// Arrange\n")
				testCode.WriteString("\t// TODO: Set up test inputs\n\n")
				testCode.WriteString("\t// Act\n")
				testCode.WriteString(fmt.Sprintf("\t// TODO: Call %s with inputs\n\n", fd.Name.Name))
				testCode.WriteString("\t// Assert\n")
				testCode.WriteString("\t// TODO: Verify expected outcomes\n")
				testCode.WriteString("}\n\n")
			}
		}
		return true
	})

	if funcCount == 0 {
		result.Suggestions = append(result.Suggestions, "No exported functions found. Tests are typically written for exported functions.")
		testCode.WriteString("// No exported functions found to generate tests for.\n")
	}

	result.TestCode = testCode.String()
}

// generateTableDrivenTests generates table-driven test scaffolding
func generateTableDrivenTests(file *ast.File, pkgName string, result *TestGenResult) {
	var testCode strings.Builder

	testCode.WriteString(fmt.Sprintf("package %s\n\n", pkgName))
	testCode.WriteString("import (\n\t\"testing\"\n)\n\n")

	funcCount := 0
	ast.Inspect(file, func(n ast.Node) bool {
		if fd, ok := n.(*ast.FuncDecl); ok {
			if fd.Name.IsExported() && fd.Recv == nil {
				funcCount++
				params := formatFieldList(fd.Type.Params)
				returns := formatFieldList(fd.Type.Results)

				testCode.WriteString(fmt.Sprintf("func Test%s(t *testing.T) {\n", fd.Name.Name))
				testCode.WriteString("\ttests := []struct {\n")
				testCode.WriteString("\t\tname string\n")

				// Add input fields based on function params
				if params != "" {
					for _, p := range strings.Split(params, ", ") {
						parts := strings.Fields(p)
						if len(parts) >= 2 {
							testCode.WriteString(fmt.Sprintf("\t\t%s %s\n", parts[0], parts[1]))
						}
					}
				}

				// Add expected output fields
				if returns != "" {
					testCode.WriteString(fmt.Sprintf("\t\twant %s\n", returns))
				}
				testCode.WriteString("\t\twantErr bool\n")
				testCode.WriteString("\t}{\n")
				testCode.WriteString("\t\t{\n")
				testCode.WriteString("\t\t\tname: \"success case\",\n")
				testCode.WriteString("\t\t\t// TODO: Fill in test case values\n")
				testCode.WriteString("\t\t},\n")
				testCode.WriteString("\t\t{\n")
				testCode.WriteString("\t\t\tname: \"error case\",\n")
				testCode.WriteString("\t\t\twantErr: true,\n")
				testCode.WriteString("\t\t\t// TODO: Fill in test case values\n")
				testCode.WriteString("\t\t},\n")
				testCode.WriteString("\t}\n\n")
				testCode.WriteString("\tfor _, tt := range tests {\n")
				testCode.WriteString("\t\tt.Run(tt.name, func(t *testing.T) {\n")
				testCode.WriteString(fmt.Sprintf("\t\t\t// TODO: Call %s and verify results\n", fd.Name.Name))
				testCode.WriteString("\t\t})\n")
				testCode.WriteString("\t}\n")
				testCode.WriteString("}\n\n")
			}
		}
		return true
	})

	if funcCount == 0 {
		result.Suggestions = append(result.Suggestions, "No exported functions found for table-driven tests.")
		testCode.WriteString("// No exported functions found to generate tests for.\n")
	}

	result.TestCode = testCode.String()
}

// Helper functions

func getReceiverTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			return ident.Name
		}
	}
	return ""
}

func formatFieldList(fl *ast.FieldList) string {
	if fl == nil || len(fl.List) == 0 {
		return ""
	}

	var parts []string
	for _, f := range fl.List {
		typeStr := formatType(f.Type)
		if len(f.Names) == 0 {
			parts = append(parts, typeStr)
		} else {
			for _, name := range f.Names {
				parts = append(parts, fmt.Sprintf("%s %s", name.Name, typeStr))
			}
		}
	}
	return strings.Join(parts, ", ")
}

func formatType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + formatType(t.X)
	case *ast.SelectorExpr:
		return formatType(t.X) + "." + t.Sel.Name
	case *ast.ArrayType:
		if t.Len == nil {
			return "[]" + formatType(t.Elt)
		}
		return "[...]" + formatType(t.Elt)
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", formatType(t.Key), formatType(t.Value))
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.FuncType:
		return "func(" + formatFieldList(t.Params) + ")"
	case *ast.ChanType:
		return "chan " + formatType(t.Value)
	default:
		return "any"
	}
}

func extractParamNames(params string) string {
	if params == "" {
		return ""
	}
	var names []string
	for _, p := range strings.Split(params, ", ") {
		parts := strings.Fields(p)
		if len(parts) >= 1 {
			names = append(names, parts[0])
		}
	}
	return strings.Join(names, ", ")
}
