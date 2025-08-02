package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <package_path> [symbol_name]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  %s fmt\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s fmt Printf\n", os.Args[0])
		os.Exit(1)
	}

	packagePath := os.Args[1]
	var symbolName string
	if len(os.Args) > 2 {
		symbolName = os.Args[2]
	}

	// Start the MCP server as a subprocess
	cmd := exec.Command("./bin/mcp-go-assistant")

	// Set up stdio pipes
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatalf("Failed to create stdin pipe: %v", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("Failed to create stdout pipe: %v", err)
	}

	// Start the server
	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	// We'll communicate directly with JSON-RPC over stdio
	// Read from stdout and write to stdin directly
	reader := bufio.NewReader(stdout)
	writer := bufio.NewWriter(stdin)

	// Create request manually using JSON-RPC
	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo": map[string]interface{}{
				"name":    "mcp-go-assistant-client",
				"version": "1.0.0",
			},
		},
	}

	// Send initialize request
	reqData, _ := json.Marshal(request)
	fmt.Fprintf(writer, "%s\n", reqData)
	writer.Flush()

	// Read initialize response
	_, err = reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Failed to read initialize response: %v", err)
	}

	// Send initialized notification
	notification := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "notifications/initialized",
	}
	notifData, _ := json.Marshal(notification)
	fmt.Fprintf(writer, "%s\n", notifData)
	writer.Flush()

	// Prepare tool call arguments
	args := map[string]interface{}{
		"package_path": packagePath,
	}
	if symbolName != "" {
		args["symbol_name"] = symbolName
	}

	// Send tool call request
	toolRequest := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name":      "go-doc",
			"arguments": args,
		},
	}

	toolData, _ := json.Marshal(toolRequest)
	fmt.Fprintf(writer, "%s\n", toolData)
	writer.Flush()

	// Read tool response
	toolResponse, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Failed to read tool response: %v", err)
	}

	// Parse and display the response
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(toolResponse), &result); err != nil {
		log.Fatalf("Failed to parse tool response: %v", err)
	}

	// Extract and display the documentation text
	if resultData, ok := result["result"].(map[string]interface{}); ok {
		if content, ok := resultData["content"].([]interface{}); ok {
			for _, item := range content {
				if textItem, ok := item.(map[string]interface{}); ok {
					if text, ok := textItem["text"].(string); ok {
						fmt.Println(text)
					}
				}
			}
		}
	} else if errorData, ok := result["error"].(map[string]interface{}); ok {
		fmt.Printf("Error: %v\n", errorData["message"])
	} else {
		fmt.Printf("Raw response: %s", toolResponse)
	}

	// Clean up
	stdin.Close()
	stdout.Close()
	if err := cmd.Wait(); err != nil {
		log.Printf("Server exited with error: %v", err)
	}
}
