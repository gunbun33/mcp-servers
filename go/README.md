# Go MCP Server ⚠️

This is a Go implementation of the Machine Comprehension Protocol (MCP) server using the centralmind/gateway image. It provides a server that implements parts of the MCP protocol for use with VS Code and other MCP clients.

**Status: In Progress**

## Features

- JSON-RPC 2.0 API (partially implemented)
- Server-Sent Events (SSE) endpoint available but needs configuration
- Swagger UI for API documentation
- PostgreSQL database integration
- Docker configuration

## Getting Started

### Prerequisites

- Docker and Docker Compose
- VS Code

### Running the Server

```bash
cd /home/chris/mcp-servers/mcp-servers
docker-compose -f docker-compose.yml -f go/docker-compose.override.yml up mcp-go
```

The server will be available at http://localhost:9090 with metrics at http://localhost:9091.

**Note:** The server is currently in development and not all MCP protocol features are fully implemented.

## Using with VS Code

VS Code can connect to the MCP server using the SSE/HTTP MCP Protocol. The server exposes the following endpoints:

- `GET /sse` - Server-Sent Events endpoint (currently returns 405 Method Not Allowed)
- `GET /` - API documentation and Swagger UI
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
