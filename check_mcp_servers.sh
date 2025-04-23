#!/bin/bash

# Script to check if all MCP servers are running correctly

echo "Checking MCP servers..."

# Check Python MCP server
echo -e "\n=== Checking Python MCP server ==="
curl -s http://localhost:8080/health | jq || echo "Python MCP server is not running"

# Check Go MCP server
echo -e "\n=== Checking Go MCP server ==="
# Try different health endpoints
echo "Checking SSE endpoint:"
curl -s -I http://localhost:9090/sse | head -n 1 || echo "Go MCP server SSE endpoint is not responding"
echo "\nChecking API endpoint:"
curl -s -I http://localhost:9090/ | head -n 1 || echo "Go MCP server API endpoint is not responding"

# Check Rust MCP server
echo -e "\n=== Checking Rust MCP server ==="
# Try different endpoints
echo "Checking SSE endpoint:"
curl -s -I http://localhost:8082/sse | head -n 1 || echo "Rust MCP server SSE endpoint is not responding"
echo "\nChecking root endpoint:"
curl -s -I http://localhost:8082/ | head -n 1 || echo "Rust MCP server root endpoint is not responding"

echo -e "\nDone checking MCP servers."
