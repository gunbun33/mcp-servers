# Rust Terraform MCP Server ⚠️

This is a Rust implementation of the Machine Comprehension Protocol (MCP) server for Terraform using the tfmcp tool. It aims to provide a server that implements the MCP protocol for use with VS Code and other MCP clients.

**Status: In Progress**

## Features

- Terraform integration for infrastructure as code
- JSON-RPC messaging support (via stdin/stdout)
- Docker configuration
- Sample Terraform project generation

**Planned Features (Not Yet Implemented):**
- HTTP server for MCP protocol
- Server-Sent Events (SSE) for real-time communication
- Comprehensive error handling and logging
- Prometheus metrics for monitoring

## Getting Started

### Prerequisites

- Docker and Docker Compose
- VS Code

### Running the Server

```bash
cd /home/chris/mcp-servers/mcp-servers
docker-compose up mcp-rust
```

The server is designed to run at http://localhost:8082 with metrics at http://localhost:8081, but the HTTP interface is currently not fully implemented.

**Note:** The server is currently in development and significant work is needed to implement the full MCP protocol over HTTP.

## Current Status and Limitations

The Rust MCP server is currently in early development:

- The `tfmcp` tool primarily works via stdin/stdout JSON-RPC, not HTTP
- HTTP endpoints are not yet implemented
- The server can generate and work with Terraform configurations
- Additional development is needed to fully implement the MCP protocol over HTTP

## Future Work

To make this server production-ready, the following work is needed:

- Implement HTTP server functionality
- Add SSE support for real-time communication
- Create proper health check endpoints
- Implement the full MCP protocol specification
- Add comprehensive error handling and logging

### Example VS Code Settings

Add the following to your VS Code settings.json:

```json
{
  "mcp.terraform.serverUrl": "http://localhost:8082",
  "mcp.terraform.enableAutoConnect": true
}
```

## Configuration

The server can be configured using the `config.toml` file. Key configuration options:

- `server.port` - HTTP server port
- `server.host` - HTTP server host
- `server.metrics_port` - Prometheus metrics port
- `server.log_level` - Logging level
- `terraform.version` - Terraform version
- `terraform.state_path` - Path to store Terraform state files

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| PORT | HTTP server port | 8080 |
| HOST | HTTP server host | 0.0.0.0 |
| LOG_LEVEL | Logging level | info |
| DEBUG | Enable debug mode | false |
| METRICS_PORT | Prometheus metrics port | 8081 |
| MCP_SERVER_NAME | Server name | Rust Terraform MCP Server |
| MCP_SERVER_VERSION | Server version | 1.0.0 |
