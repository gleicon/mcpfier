# MCPFier Architecture

This document provides a detailed technical overview of MCPFier's architecture, design decisions, and implementation details.

## System Overview

MCPFier is designed as a bridge between traditional command-line tools and the Model Context Protocol (MCP) ecosystem. It transforms arbitrary commands and scripts into standardized MCP tools that can be consumed by LLMs and other MCP clients.

## Core Architecture

### High-Level Components

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   MCP Client    │────│   MCPFier       │────│   Execution     │
│   (LLM/Tool)    │    │   Server        │    │   Environment   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │
                              │
                       ┌─────────────────┐
                       │   Configuration │
                       │   (YAML)        │
                       └─────────────────┘
```

### Component Details

#### 1. Configuration System

- **File Format**: YAML-based configuration (`config.yaml`)
- **Hot Reload**: Currently requires restart (future enhancement)
- **Validation**: Schema validation on startup
- **Extensibility**: Supports arbitrary metadata fields

#### 2. Command Registry

- **Discovery**: Automatic command discovery from configuration
- **Registration**: Dynamic MCP tool registration
- **Lookup**: Efficient command resolution by name
- **Metadata**: Maintains command descriptions and constraints

#### 3. Execution Engine

- **Local Execution**: Direct process execution on host
- **Container Execution**: Docker-based isolated execution
- **Webhook Execution**: HTTP client for external API calls
- **Environment Management**: Per-command environment variables with expansion support
- **Timeout Handling**: Configurable execution timeouts

#### 4. MCP Protocol Handler

- **Transport**: Dual-mode (STDIO for desktop, HTTP/SSE for enterprise)
- **Message Handling**: Full MCP specification compliance
- **Tool Management**: Dynamic tool registration and invocation
- **Error Handling**: Structured error responses

#### 5. Authentication and Authorization

- **API Key Authentication**: Simple key-based auth with granular permissions
- **Permission Model**: Per-tool access control
- **Context Propagation**: Authentication context through request chain

#### 6. Rate Limiting

- **Token Bucket Algorithm**: Via golang.org/x/time/rate
- **Per-Client Limiting**: Separate limits by API key or IP address
- **Configurable**: Requests per minute and burst size

#### 7. Analytics Engine

- **Storage**: Embedded SQLite database
- **Metrics**: Command execution, HTTP requests, webhook calls
- **Web Dashboard**: Real-time analytics visualization
- **Retention**: Configurable data retention policies

#### 8. TLS/HTTPS Security

- **Manual TLS**: Bring your own certificates (cert_file, key_file)
- **Protocol**: HTTPS with HTTP/2 automatic upgrade
- **Encryption**: All traffic encrypted including API keys and tool results
- **Flexibility**: Optional (disabled by default for development)

## Data Flow

### Command Execution Flow

```
1. MCP Client Request (over HTTPS if TLS enabled)
   ↓
2. TLS Termination (if enabled)
   ↓
3. Rate Limiting Check
   ↓
4. Authentication/Authorization
   ↓
5. Protocol Parsing
   ↓
6. Command Resolution
   ↓
7. Execution Mode Selection
   ↓ (Local)        ↓ (Container)     ↓ (Webhook)
8a. Direct Exec    8b. Docker Exec    8c. HTTP Request
   ↓                ↓                  ↓
9. Result Capture (with analytics recording)
   ↓
10. MCP Response
   ↓
11. TLS Encryption (if enabled)
   ↓
12. Client Response (over HTTPS if TLS enabled)
```

### Configuration Loading Flow

```
1. Application Start
   ↓
2. YAML File Read
   ↓
3. Schema Validation
   ↓
4. Command Registration
   ↓
5. MCP Server Ready
```

## Implementation Details

### Package Structure

```
mcpfier/
├── main.go                    # CLI entry point
├── internal/
│   ├── analytics/            # Analytics and metrics
│   │   ├── analytics.go      # Analytics interface
│   │   └── sqlite.go         # SQLite implementation
│   ├── auth/                 # Authentication and authorization
│   │   └── auth.go           # API key auth implementation
│   ├── config/               # Configuration management
│   │   ├── config.go         # Types and loading logic
│   │   └── config_test.go    # Configuration tests
│   ├── executor/             # Command execution
│   │   ├── executor.go       # Executor interface and service
│   │   ├── local.go          # Local execution implementation
│   │   ├── container.go      # Container execution implementation
│   │   ├── webhook.go        # HTTP/API execution implementation
│   │   └── executor_test.go  # Execution tests
│   ├── ratelimit/            # Rate limiting
│   │   └── ratelimit.go      # Token bucket rate limiter
│   ├── server/               # MCP server implementation
│   │   ├── server.go         # STDIO MCP server
│   │   ├── http_server.go    # HTTP/SSE MCP server
│   │   └── logging.go        # Request logging middleware
│   └── setup/                # Setup and configuration helpers
│       └── setup.go          # Claude Desktop setup generation
│   └── setup/                # Setup instructions
│       └── setup.go          # Setup command implementation
└── mcpfier_test.go           # Integration tests
```

### Core Types

```go
// config/config.go
type Command struct {
    Name        string            `yaml:"name"`
    Script      string            `yaml:"script"`
    Args        []string          `yaml:"args"`
    Description string            `yaml:"description"`
    Container   string            `yaml:"container"`
    Timeout     string            `yaml:"timeout"`
    Env         map[string]string `yaml:"env"`
    Webhook     *WebhookConfig    `yaml:"webhook"`  // HTTP API configuration
}

type WebhookConfig struct {
    URL         string            `yaml:"url"`
    Method      string            `yaml:"method"`
    Headers     map[string]string `yaml:"headers"`
    Body        string            `yaml:"body"`
    Auth        *WebhookAuth      `yaml:"auth"`
    Retry       *WebhookRetry     `yaml:"retry"`
}

type Config struct {
    Commands  []Command       `yaml:"commands"`
    Server    ServerConfig    `yaml:"server"`
    Analytics AnalyticsConfig `yaml:"analytics"`
}

type HTTPConfig struct {
    Port      int             `yaml:"port"`
    Host      string          `yaml:"host"`
    TLS       TLSConfig       `yaml:"tls"`
    Auth      AuthConfig      `yaml:"auth"`
    CORS      CORSConfig      `yaml:"cors"`
    RateLimit RateLimitConfig `yaml:"rate_limit"`
}

type TLSConfig struct {
    Enabled  bool   `yaml:"enabled"`
    CertFile string `yaml:"cert_file"`
    KeyFile  string `yaml:"key_file"`
}

// executor/executor.go
type Executor interface {
    Execute(ctx context.Context, cmd *config.Command) (string, error)
}

type Service struct {
    local     *LocalExecutor
    container *ContainerExecutor
    webhook   *WebhookExecutor
    analytics analytics.Analytics
}

// server/http_server.go
type HTTPServer struct {
    config       *config.Config
    mcpServer    *server.MCPServer
    httpServer   *server.StreamableHTTPServer
    executor     *executor.Service
    analytics    analytics.Analytics
}
```

### Execution Strategies

#### Local Execution
- **Process**: `exec.CommandContext()`
- **Environment**: Inherits parent plus command-specific vars (with ${VAR} expansion)
- **Isolation**: None (same privileges as MCPFier)
- **Performance**: High (no containerization overhead)
- **Use Case**: Trusted scripts and system commands

#### Container Execution
- **Runtime**: Docker via command-line interface
- **Isolation**: Complete filesystem and process isolation
- **Resource Limits**: Configurable via Docker options
- **Performance**: Lower (containerization overhead)
- **Use Case**: Untrusted code, heavy computations

#### Webhook Execution
- **Protocol**: HTTP/HTTPS client
- **Authentication**: Bearer, API Key, Basic Auth (with env var expansion)
- **Retry Logic**: Exponential backoff with configurable retries
- **Performance**: Network-dependent
- **Use Case**: External API integration, serverless functions

### Security Model

#### Transport Security
- **TLS/HTTPS**: Optional encryption for HTTP server mode
- **Certificate Management**: Manual (bring your own certificates)
- **Protocol**: HTTPS with automatic HTTP/2 upgrade
- **Encryption**: API keys, auth tokens, and all MCP messages encrypted in transit
- **Backward Compatible**: Disabled by default, opt-in for production

#### Local Execution Security
- Runs with MCPFier process privileges
- Full access to host filesystem
- Inherits network permissions
- Suitable for trusted environments

#### Container Execution Security
- Isolated filesystem access
- No host network access by default
- Resource limits enforceable
- Temporary container lifecycle

#### Webhook Execution Security
- TLS certificate validation for external APIs
- Environment variable expansion for secure credential management
- Retry logic with backoff prevents abuse
- Request timeout enforcement

## Protocol Implementation

### MCP Tool Registration

Each YAML command becomes an MCP tool:

```go
s.AddTool(
    mcp.NewTool(cmd.Name,
        mcp.WithDescription(cmd.Description),
    ),
    func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        return server.ExecuteCommand(ctx, cmd.Name, request)
    },
)
```

### Message Handling

MCPFier handles these MCP message types:
- **initialize**: Server capabilities negotiation
- **tools/list**: Available tool enumeration
- **tools/call**: Tool execution requests

### Error Handling

Structured error responses for:
- Configuration errors
- Command not found
- Execution failures
- Timeout exceeded
- Container runtime errors

## Performance Considerations

### Local Execution Performance
- **Startup Time**: ~1ms (direct process spawn)
- **Memory Overhead**: Minimal (shared process space)
- **CPU Impact**: Low (direct execution)
- **Scalability**: High (limited by system resources)

### Container Execution Performance
- **Startup Time**: ~100-500ms (container lifecycle)
- **Memory Overhead**: Container base image + application
- **CPU Impact**: Moderate (containerization layer)
- **Scalability**: Moderate (Docker daemon limits)

### Optimization Strategies
- Container image pre-pulling
- Connection pooling (future)
- Command result caching (future)
- Parallel execution support

## Extensibility Points

### Custom Execution Backends
The execution engine is designed for extension:
- Kubernetes Job execution
- AWS Lambda functions
- Remote SSH execution
- Serverless platforms

### Enhanced Security Features
- ACME/Let's Encrypt autocert integration
- Mutual TLS (mTLS) for client certificates
- SELinux/AppArmor integration
- Resource quotas and limits
- Audit logging
- Access control policies

### Protocol Extensions
- Streaming output support
- Binary data handling
- File upload/download
- Interactive sessions

## Configuration Schema

### Current Schema

```yaml
commands:
  - name: string          # Required: Unique identifier
    script: string        # Required: Executable path
    args: []string        # Optional: Static arguments
    description: string   # Optional: Tool description
    timeout: string       # Optional: Execution timeout
    container: string     # Optional: Docker image
    env:                  # Optional: Environment variables
      KEY: string
```

### Future Schema Extensions

```yaml
commands:
  - name: string
    # ... existing fields ...
    security:             # Security configuration
      user: string        # Execution user
      capabilities: []string # Linux capabilities
      read_only: bool     # Filesystem read-only
    resources:            # Resource limits
      memory: string      # Memory limit
      cpu: string         # CPU limit
      disk: string        # Disk space limit
    network:              # Network configuration
      enabled: bool       # Network access
      allowed_hosts: []string # Allowed destinations
```

## Testing Strategy

### Unit Tests
- Configuration loading and validation
- Command execution (mocked)
- MCP protocol compliance
- Error handling scenarios

### Integration Tests
- End-to-end command execution
- Container runtime integration
- MCP client interaction
- Performance benchmarks

### Security Tests
- TLS certificate validation
- Encrypted transport verification
- Container escape prevention
- Resource limit enforcement
- Privilege escalation detection
- Network isolation validation

## Deployment Patterns

### Standalone Deployment
- Single binary execution
- Local configuration file
- Direct MCP client connection
- Optional TLS with self-signed certificates (development)

### Service Deployment
- Systemd service integration
- Process supervision
- TLS with production certificates
- Certificate renewal automation (external tools like certbot)
- Log aggregation
- Health monitoring

### Container Deployment
- Dockerized MCPFier server
- Container-in-container execution
- Kubernetes integration
- Scalable deployment
- Reverse proxy for TLS termination (recommended for production)

### Production TLS Options

**Option 1: Direct TLS (Implemented)**
```yaml
server:
  http:
    port: 8443
    tls:
      enabled: true
      cert_file: "/etc/mcpfier/certs/cert.pem"
      key_file: "/etc/mcpfier/certs/key.pem"
```
- MCPFier handles TLS directly
- Suitable for simple deployments
- Certificate management is manual

**Option 2: Reverse Proxy (Recommended)**
```
Internet → nginx/Caddy (TLS) → MCPFier (HTTP on localhost:8080)
```
- Industry standard pattern
- Easier certificate management (certbot, Caddy auto-HTTPS)
- Load balancing and caching capabilities
- MCPFier runs without TLS configuration

## Monitoring and Observability

### Metrics (Future)
- Command execution count
- Execution duration
- Error rates
- Resource utilization

### Logging
- Structured logging format
- Configurable log levels
- Request/response tracing
- Security event logging

### Health Checks
- Server liveness probes
- Configuration validation
- Dependency health checks
- Performance monitoring

## Development Guidelines

### Code Organization
- Clear separation of concerns
- Interface-driven design
- Comprehensive error handling
- Extensive documentation

### Testing Requirements
- Unit test coverage > 80%
- Integration tests for core flows
- Security test automation
- Performance regression testing

### Documentation Standards
- API documentation
- Configuration examples
- Troubleshooting guides
- Architecture decisions

## Future Enhancements

### Phase 2 Features
- Remote server mode with authentication
- Resource usage monitoring and limits
- Command chaining and workflows
- Advanced security policies

### Phase 3 Features
- Kubernetes job execution
- Scheduled task support
- LLM-to-LLM communication patterns
- Web UI for configuration management

### Long-term Vision
- Enterprise-grade security and compliance
- Multi-tenant execution environments
- Marketplace for command configurations
- Integration with major cloud platforms