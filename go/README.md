# Go MCP Server ⚠️

A custom Go implementation of the Model Context Protocol (MCP) server using the Gin framework. This server implements the MCP protocol with PostgreSQL integration for database operations and code assistance features.

**Status: In Progress**

_Part of the [MCP Servers](https://github.com/cploutarchou/mcp-servers) collection by Christos Ploutarchou._

## Features

- JSON-RPC 2.0 API implementation
- Server-Sent Events (SSE) endpoint for streaming
- Health check endpoint
- Prometheus metrics
- Structured logging with logrus
- CORS support
- PostgreSQL database integration
- Docker configuration
- Code assistance capabilities:
  - Code completion for multiple languages
  - Code analysis and diagnostics
  - Documentation lookup

## Getting Started

### Prerequisites

- Docker and Docker Compose
- VS Code

### Running the Server

```bash
cd /home/chris/mcp-servers/mcp-servers
docker-compose up mcp-go
```

The server will be available at http://localhost:9090 with metrics at http://localhost:9091.

**Note:** The server is currently in development and not all MCP protocol features are fully implemented.

## Using with VS Code

VS Code can connect to the MCP server using the SSE/HTTP MCP Protocol. The server exposes the following endpoints:

- `POST /` - JSON-RPC endpoint for MCP protocol requests
- `GET /sse` - Server-Sent Events endpoint for streaming
- `GET /health` - Health check endpoint
- `GET /metrics` - Prometheus metrics endpoint (on metrics port)

## Development

The Go MCP server is built with the following components:

- **Gin Framework**: High-performance HTTP web framework
- **Viper**: Configuration management
- **Logrus**: Structured logging
- **Prometheus**: Metrics collection

To build the server locally:

```bash
cd /home/chris/mcp-servers/mcp-servers/go/src
go build -o mcp-go-server ./cmd/server
```

## Code Assistance Features

The Go MCP server includes advanced code assistance capabilities:

### Code Completion

Provides intelligent code completion suggestions for multiple programming languages:

- Go
- Python
- SQL
- Generic support for other languages

**Endpoint:** `POST /codeassist/completion`

### Code Analysis

Analyzes code for potential issues, bugs, and improvement suggestions:

- Syntax errors
- Best practices violations
- Security concerns
- Performance optimizations

**Endpoint:** `POST /codeassist/analysis`

### Documentation

Provides documentation for programming language symbols, functions, and keywords:

- Function signatures and descriptions
- Usage examples
- Links to official documentation

**Endpoint:** `POST /codeassist/documentation`

## Work in Progress

The following features are currently under development:

1. **Database Integration**: Implementing full PostgreSQL integration for storing and retrieving code snippets and analysis results
2. **Language Server Protocol (LSP) Support**: Adding LSP compatibility for better IDE integration
3. **AI-Powered Suggestions**: Enhancing code completion with more advanced AI models
4. **Testing and Validation**: Comprehensive testing of all code assistance features
5. **Performance Optimization**: Improving response times for large codebases
- Various REST endpoints for database operations

**Current Limitations:**
- The health check endpoint is not implemented in the standard location
- The SSE endpoint needs further configuration to work with VS Code
- Some MCP protocol features may not be fully implemented

### Example VS Code Settings

Add the following to your VS Code settings.json:

```json
{
  "mcp.serverUrl": "http://localhost:9090",
  "mcp.enableAutoConnect": true
}
```

## Configuration

The server can be configured using the `config.yaml` file. Key configuration options:

- `server.port` - HTTP server port
- `server.host` - HTTP server host
- `server.metrics_port` - Prometheus metrics port
- `server.log_level` - Logging level
- `database.connection_string` - PostgreSQL connection string

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| CONFIG_FILE | Path to config file | /app/config.yaml |
| LOG_LEVEL | Logging level | info |
| DEBUG | Enable debug mode | false |
| MCP_SERVER_NAME | Server name | Go MCP Server |
| MCP_SERVER_VERSION | Server version | 1.0.0 |

## License

MIT License - Copyright (c) 2025 Christos Ploutarchou

See [LICENSE](../LICENSE) file for details.
