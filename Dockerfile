# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o mcpfier .

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -g 1001 mcpfier && \
    adduser -D -u 1001 -G mcpfier mcpfier

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /build/mcpfier .

# Create directories for mounted volumes
RUN mkdir -p /app/config /app/data /app/logs && \
    chown -R mcpfier:mcpfier /app

# Switch to non-root user
USER mcpfier

# Expose HTTP port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Default command (HTTP server mode)
CMD ["./mcpfier", "--server", "--config", "/app/config/config.yaml"]