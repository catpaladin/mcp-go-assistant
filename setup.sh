#!/bin/bash

# MCP Go Assistant - Setup Script
# This script helps set up the production-grade MCP Go Assistant

set -e

echo "🚀 Setting up MCP Go Assistant..."
echo ""

# Step 1: Download dependencies
echo "📦 Downloading dependencies..."
go mod tidy
echo "✅ Dependencies downloaded"
echo ""

# Step 2: Build the application
echo "🔨 Building the application..."
go build -o mcp-go-assistant ./cmd/mcp-go-assistant
echo "✅ Application built successfully"
echo ""

# Step 3: Create configuration file if it doesn't exist
if [ ! -f "config.yaml" ]; then
    echo "📝 Creating configuration file from example..."
    cp config.example.yaml config.yaml
    echo "✅ Configuration file created: config.yaml"
    echo "   You can edit config.yaml to customize settings"
else
    echo "ℹ️  Configuration file already exists: config.yaml"
fi
echo ""

# Step 4: Run tests
echo "🧪 Running tests..."
go test ./... || {
    echo "⚠️  Some tests failed, but build succeeded"
}
echo ""

echo "✅ Setup complete!"
echo ""
echo "📋 Quick Start:"
echo "   Run with default config:     ./mcp-go-assistant"
echo "   Run with custom config:     MCP_CONFIG=config.yaml ./mcp-go-assistant"
echo "   Run with debug logging:     MCP_LOG_LEVEL=debug ./mcp-go-assistant"
echo "   Run on custom port:          MCP_PORT=9090 ./mcp-go-assistant"
echo ""
echo "📊 View Metrics:"
echo "   curl http://localhost:8080/metrics"
echo ""
echo "📚 Documentation:"
echo "   - PRODUCTION_READINESS.md - Implementation summary"
echo "   - PRODUCTION_IMPROVEMENTS.md - Detailed documentation"
echo "   - config.example.yaml - Configuration reference"
echo ""
echo "🎉 MCP Go Assistant is ready!"
