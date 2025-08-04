# Multi-layer test image for Nexus registry testing
FROM alpine:3.18 AS base
LABEL maintainer="test@example.com"
LABEL description="Multi-layer test image for registry testing"

# Layer 1: Install basic utilities
RUN apk update && \
    apk add --no-cache \
    curl \
    wget \
    ca-certificates \
    && rm -rf /var/cache/apk/*

# Layer 2: Create application directories and users
RUN addgroup -g 1001 appgroup && \
    adduser -D -u 1001 -G appgroup appuser && \
    mkdir -p /app/bin /app/config /app/data /app/logs && \
    chown -R appuser:appgroup /app

# Layer 3: Add configuration files
COPY <<EOF /app/config/app.conf
# Application configuration
app_name=nexus-test-app
version=1.0.0
environment=development
log_level=info
EOF

# Layer 4: Install additional tools
RUN apk add --no-cache \
    bash \
    jq \
    git \
    && rm -rf /var/cache/apk/*

# Layer 5: Create a simple application script
COPY <<EOF /app/bin/start.sh
#!/bin/bash
set -e

echo "Starting Nexus Test Application..."
echo "Version: $(cat /app/config/app.conf | grep version | cut -d'=' -f2)"
echo "Environment: $(cat /app/config/app.conf | grep environment | cut -d'=' -f2)"

# Create a test file with timestamp
echo "Application started at: $(date)" > /app/logs/startup.log

# Keep container running
while true; do
    echo "App is running... $(date)" >> /app/logs/heartbeat.log
    sleep 30
done
EOF

RUN chmod +x /app/bin/start.sh

# Layer 6: Add some test data
RUN echo "Test data file 1" > /app/data/test1.txt && \
    echo "Test data file 2" > /app/data/test2.txt && \
    echo "Configuration backup" > /app/data/config.bak

# Layer 7: Set final permissions and workspace
WORKDIR /app
USER appuser

# Expose a port for testing
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD test -f /app/logs/heartbeat.log || exit 1

# Default command
CMD ["/app/bin/start.sh"]