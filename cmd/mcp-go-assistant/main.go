package main

import (
	"context"
	"fmt"
	"log"
	"mcp-go-assistant/internal/codereview"
	"mcp-go-assistant/internal/godoc"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// GoDocTool handles the go-doc tool invocation
func GoDocTool(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[godoc.GoDocParams]) (*mcp.CallToolResultFor[any], error) {
	// Get documentation using the internal godoc package
	documentation, err := godoc.GetDocumentation(ctx, params.Arguments)
	if err != nil {
		return &mcp.CallToolResultFor[any]{
			IsError: true,
			Content: []mcp.Content{&mcp.TextContent{
				Text: fmt.Sprintf("Error getting documentation: %v", err),
			}},
		}, nil
	}

	// Return successful result
	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{
			Text: documentation,
		}},
	}, nil
}

// CodeReviewTool handles the code-review tool invocation
func CodeReviewTool(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[codereview.CodeReviewParams]) (*mcp.CallToolResultFor[any], error) {
	// Perform code review using the internal codereview package
	result, err := codereview.PerformCodeReview(ctx, params.Arguments)
	if err != nil {
		return &mcp.CallToolResultFor[any]{
			IsError: true,
			Content: []mcp.Content{&mcp.TextContent{
				Text: fmt.Sprintf("Error performing code review: %v", err),
			}},
		}, nil
	}

	// Return the review result as JSON
	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{
			Text: result.String(),
		}},
	}, nil
}

func main() {
	// Create MCP server
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "mcp-go-assistant",
		Version: "1.0.0",
	}, nil)

	// Register the go-doc tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "go-doc",
		Description: "Get Go documentation for packages and symbols using 'go doc' command",
	}, GoDocTool)

	// Register the code-review tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "code-review",
		Description: "Analyze Go code and provide improvement suggestions based on best practices",
	}, CodeReviewTool)

	// Run server with stdio transport
	if err := server.Run(context.Background(), mcp.NewStdioTransport()); err != nil {
		log.Fatal(err)
	}
}
