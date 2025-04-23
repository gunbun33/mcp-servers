FROM python:3.11-slim

WORKDIR /app

# Set environment variables
ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    PORT=8080 \
    HOST=0.0.0.0 \
    LOG_LEVEL=info \
    METRICS_PORT=8081 \
    DEBUG=false \
    MCP_SERVER_NAME="Python FastAPI MCP" \
    MCP_SERVER_VERSION="1.0.0"

# Install dependencies
COPY requirements.txt /app/
RUN pip install --no-cache-dir -r requirements.txt

# Copy application code
COPY mcp_server.py /app/

# Create a directory for data
RUN mkdir -p /data
VOLUME /data

# Expose ports for the main API and metrics
EXPOSE 8080
EXPOSE 8081

# Healthcheck
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

# Run the server with production settings
CMD ["uvicorn", "mcp_server:app", "--host", "0.0.0.0", "--port", "8080", "--workers", "4", "--log-level", "info"]