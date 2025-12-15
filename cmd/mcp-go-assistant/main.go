package main

import (
	"context"
	"log"

	"mcp-go-assistant/internal/codereview"
	"mcp-go-assistant/internal/godoc"
	"mcp-go-assistant/internal/testgen"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// GoDocTool handles the go-doc tool invocation
func GoDocTool(ctx context.Context, request *mcp.CallToolRequest, params godoc.GoDocParams) (*mcp.CallToolResult, any, error) {
	documentation, err := godoc.GetDocumentation(ctx, params)
	if err != nil {
		return nil, nil, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: documentation}},
	}, nil, nil
}

// CodeReviewTool handles the code-review tool invocation
func CodeReviewTool(ctx context.Context, request *mcp.CallToolRequest, params codereview.CodeReviewParams) (*mcp.CallToolResult, *codereview.ReviewResult, error) {
	result, err := codereview.PerformCodeReview(ctx, params)
	if err != nil {
		return nil, nil, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: result.String()}},
	}, result, nil
}

// TestGenTool handles the test generation tool invocation
func TestGenTool(ctx context.Context, request *mcp.CallToolRequest, params testgen.TestGenParams) (*mcp.CallToolResult, *testgen.TestGenResult, error) {
	result, err := testgen.GenerateTests(ctx, params)
	if err != nil {
		return nil, nil, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: result.String()}},
	}, result, nil
}

func main() {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "mcp-go-assistant",
		Version: "1.1.0",
	}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "go-doc",
		Description: "Get Go documentation for packages and symbols using 'go doc' command",
	}, GoDocTool)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "code-review",
		Description: "Analyze Go code and provide improvement suggestions based on best practices",
	}, CodeReviewTool)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "test-gen",
		Description: "Generate Go test scaffolding including interfaces, mocks, and table-driven tests. Use focus='interfaces' for interface extraction and mocks, 'table' for table-driven tests, or 'unit' for basic unit tests.",
	}, TestGenTool)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
