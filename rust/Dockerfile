FROM rust:1.86-slim-bookworm AS builder
WORKDIR /app
RUN apt-get update && \
    apt-get install -y pkg-config libssl-dev git wget unzip ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# Install the latest version of tfmcp with specific features
RUN cargo install --git https://github.com/nwiizo/tfmcp.git --features "metrics logging"

FROM debian:bookworm-slim

# Install runtime dependencies
RUN apt-get update && \
    apt-get install -y libssl3 wget unzip curl ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# Install Terraform
ENV TERRAFORM_VERSION=1.7.5
RUN wget https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip && \
    unzip terraform_${TERRAFORM_VERSION}_linux_amd64.zip && \
    mv terraform /usr/local/bin/ && \
    rm terraform_${TERRAFORM_VERSION}_linux_amd64.zip

# Set environment variables
ENV PORT=8080 \
    HOST=0.0.0.0 \
    LOG_LEVEL=info \
    METRICS_PORT=8081 \
    DEBUG=false \
    MCP_SERVER_NAME="Rust Terraform MCP Server" \
    MCP_SERVER_VERSION="1.0.0"

# Create directories for data, logs, and config
RUN mkdir -p /data/terraform /var/log/tfmcp /etc/tfmcp
VOLUME /data

# Copy the binary from the builder stage
COPY --from=builder /usr/local/cargo/bin/tfmcp /usr/local/bin/

# Copy configuration
COPY config/config.toml /etc/tfmcp/

# Expose ports for the main API and metrics
EXPOSE 8080
EXPOSE 8081

# Healthcheck
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

# Create a directory for Terraform files
RUN mkdir -p /root/terraform

# Run the server in HTTP mode
CMD ["tfmcp", "mcp", "--http"]