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
- **Environment Management**: Per-command environment variables
- **Timeout Handling**: Configurable execution timeouts

#### 4. MCP Protocol Handler
- **Transport**: Stdio-based communication
- **Message Handling**: Full MCP specification compliance
- **Tool Management**: Dynamic tool registration and invocation
- **Error Handling**: Structured error responses

## Data Flow

### Command Execution Flow

```
1. MCP Client Request
   ↓
2. Protocol Parsing
   ↓
3. Command Resolution
   ↓
4. Execution Mode Selection
   ↓ (Local)        ↓ (Container)
5a. Direct Exec    5b. Docker Exec
   ↓                ↓
6. Result Capture
   ↓
7. MCP Response
   ↓
8. Client Response
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

### Core Types

```go
// Command represents a configurable command
type Command struct {
    Name        string            `yaml:"name"`        // Unique identifier
    Script      string            `yaml:"script"`      // Executable path
    Args        []string          `yaml:"args"`        // Static arguments
    Description string            `yaml:"description"` // Human description
    Container   string            `yaml:"container"`   // Docker image
    Timeout     string            `yaml:"timeout"`     // Execution timeout
    Env         map[string]string `yaml:"env"`         // Environment vars
}

// Config holds the complete configuration
type Config struct {
    Commands []Command `yaml:"commands"`
}

// MCPFierServer manages the server instance
type MCPFierServer struct {
    config *Config
}
```

### Execution Strategies

#### Local Execution
- **Process**: `exec.CommandContext()`
- **Environment**: Inherits parent + command-specific vars
- **Isolation**: None (same privileges as MCPFier)
- **Performance**: High (no containerization overhead)

#### Container Execution
- **Runtime**: Docker via command-line interface
- **Isolation**: Complete filesystem and process isolation
- **Resource Limits**: Configurable via Docker options
- **Performance**: Lower (containerization overhead)

### Security Model

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
- Container escape prevention
- Resource limit enforcement
- Privilege escalation detection
- Network isolation validation

## Deployment Patterns

### Standalone Deployment
- Single binary execution
- Local configuration file
- Direct MCP client connection

### Service Deployment
- Systemd service integration
- Process supervision
- Log aggregation
- Health monitoring

### Container Deployment
- Dockerized MCPFier server
- Container-in-container execution
- Kubernetes integration
- Scalable deployment

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