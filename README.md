# MCP Servers

A collection of Model Context Protocol (MCP) servers implemented in different programming languages for use with VS Code and other MCP clients.

## Overview

This repository contains MCP server implementations in three languages:

| Server | Status | Description |
|--------|--------|-------------|
| [Python MCP Server](./python/) | ✅ Production Ready | FastAPI-based MCP server with comprehensive features |
| [Go MCP Server](./go/) | ⚠️ In Progress | Go-based MCP server using centralmind/gateway |
| [Rust MCP Server](./rust/) | ⚠️ In Progress | Terraform-focused MCP server using tfmcp |

## What is MCP?

The Model Context Protocol (MCP) is a protocol designed for communication between AI agents and tools/services. It enables AI systems to interact with various tools and services through a standardized interface.

## Getting Started

### Prerequisites

- Docker and Docker Compose
- VS Code (for client integration)

### Running the Servers

To run all servers:

```bash
docker-compose up
```

To run only the Python server (recommended for production use):

```bash
docker-compose up mcp-python
```

## Server Status

### Python MCP Server (✅ Production Ready)

The Python MCP server is fully operational and production-ready:

- Implements the complete Model Context Protocol
- Provides SSE and JSON-RPC endpoints
- Includes comprehensive error handling and logging
- Features Prometheus metrics for monitoring
- Includes health checks for container orchestration

**Endpoints:**
- `GET /health` - Health check endpoint
- `GET /sse` - Server-Sent Events endpoint for real-time communication
- `POST /` - Main MCP endpoint for JSON-RPC requests

### Go MCP Server (⚠️ In Progress)

The Go MCP server is partially operational:

- Server starts successfully
- SSE endpoint responds (but with 405 Method Not Allowed)
- API endpoint responds (but with 302 Found, redirecting)
- Requires further configuration to fully support the MCP protocol

**Endpoints:**
- `GET /sse` - SSE endpoint (needs further configuration)
- `GET /` - API documentation

### Rust MCP Server (⚠️ In Progress)

The Rust MCP server requires further development:

- Based on the tfmcp tool for Terraform integration
- Currently doesn't fully support the HTTP interface needed for MCP
- Requires further investigation and potentially custom development

## VS Code Integration

VS Code can connect to the MCP servers using the SSE/HTTP MCP Protocol. Add the following to your VS Code settings.json:

```json
{
  "mcp.python.serverUrl": "http://localhost:8080",
  "mcp.python.enableAutoConnect": true
}
```

## Contributing

Contributions are welcome! Here are some ways you can contribute:

- Improve the existing server implementations
- Add new server implementations in other languages
- Enhance documentation and examples
- Report bugs and suggest features

## License

MIT
