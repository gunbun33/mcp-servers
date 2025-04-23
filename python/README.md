# MCP Python Server ✅

This is a production-ready Machine Comprehension Protocol (MCP) server implementation in Python. It provides a robust, scalable server that implements the MCP protocol for use with VS Code and other MCP clients.

**Status: Production Ready**

## Features

- JSON-RPC 2.0 compliant API
- Server-Sent Events (SSE) for real-time communication
- Comprehensive error handling and logging
- Prometheus metrics for monitoring
- VS Code integration
- Docker containerization

## Getting Started

### Prerequisites

- Python 3.11 or higher
- VS Code 1.60.0 or higher
- Docker (optional, for containerized deployment)

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/mcp-servers.git
   cd mcp-servers/python
   ```

2. Install dependencies:
   ```bash
   pip install -r requirements.txt
   ```

3. Run the server:
   ```bash
   uvicorn mcp_server:app --host 0.0.0.0 --port 8080 --reload
   ```

### Docker Deployment

Build and run the Docker container:

```bash
docker build -t mcp-python-agent .
docker run -p 8080:8080 -p 8081:8081 mcp-python-agent
```

Or use docker-compose:

```bash
cd ..  # Go to the root directory with docker-compose.yml
docker-compose up mcp-python
```

## Using with VS Code

VS Code can connect to the MCP server using the SSE/HTTP MCP Protocol. The server exposes the following endpoints:

- `GET /health` - Health check endpoint
- `GET /sse` - Server-Sent Events endpoint for real-time communication
- `POST /` - Main MCP endpoint for JSON-RPC requests

### Example VS Code Settings

Add the following to your VS Code settings.json:

```json
{
  "mcp.python.serverUrl": "http://localhost:8080",
  "mcp.python.enableAutoConnect": true
}
```

## API Documentation

When running in debug mode, API documentation is available at:
- Swagger UI: http://localhost:8080/docs
- ReDoc: http://localhost:8080/redoc

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| PORT | HTTP server port | 8080 |
| HOST | HTTP server host | 0.0.0.0 |
| LOG_LEVEL | Logging level (debug, info, warning, error) | info |
| DEBUG | Enable debug mode | false |
| METRICS_PORT | Prometheus metrics port | 8081 |
| ALLOWED_ORIGINS | CORS allowed origins (comma-separated) | * |

## Development

The server is built with FastAPI and follows best practices for production-ready applications:

- Comprehensive error handling
- Request validation with Pydantic
- Structured logging with Loguru
- Prometheus metrics
- Health checks for container orchestration
- Graceful shutdown

## License

MIT
