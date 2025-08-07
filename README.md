# MCPFier

MCPFier is a Model Context Protocol (MCP) server that enables secure, configurable execution of commands and scripts through a standardized interface. It transforms YAML-defined commands into MCP tools that can be consumed by any MCP-compatible client or LLM, supporting both local execution and containerized isolation.

## Overview

MCPFier bridges the gap between internal tooling and LLM accessibility by providing:

- **Unified Tool Interface**: Convert any script, command, or workflow into standardized MCP tools
- **Dual Execution Modes**: Local execution for trusted environments, containerized execution for isolation
- **Security by Design**: Sandboxed execution with configurable resource limits and access control
- **Enterprise Ready**: Replace fragmented tooling landscapes with a single, manageable interface

## Features

- **MCP Protocol Compliance**: Full implementation using mark3labs/mcp-go framework
- **YAML Configuration**: Simple, declarative command definitions
- **Container Support**: Execute commands in Docker containers for isolation
- **Environment Management**: Per-command environment variable configuration
- **Legacy Compatibility**: Maintains backward compatibility with command-line usage
- **Timeout Controls**: Configurable execution timeouts for reliability
- **Error Handling**: Comprehensive error reporting and logging

## Architecture

### Core Components

1. **Command Registry**: Loads and manages commands from YAML configuration
2. **Execution Engine**: Handles both local and containerized command execution
3. **MCP Server**: Exposes commands as standardized MCP tools via stdio transport
4. **Security Layer**: Provides isolation and resource management (containerized mode)

### Execution Modes

- **Local Mode**: Direct execution on the host system (fast, less isolated)
- **Container Mode**: Execution within Docker containers (secure, isolated)

## Installation

### Prerequisites

- Go 1.21 or later
- Docker (for containerized execution)

### Build from Source

```bash
git clone https://github.com/gleicon/mcpfier.git
cd mcpfier
go build -o mcpfier
```

### Install Dependencies

```bash
go mod download
```

## Configuration

MCPFier uses a YAML configuration file (`config.yaml`) to define available commands. Each command becomes an MCP tool automatically.

### Basic Configuration Structure

```yaml
commands:
  - name: command-name
    script: /path/to/executable
    args: ["arg1", "arg2"]
    description: "Human-readable description"
    timeout: "30s"
    container: "optional/docker:image"
    env:
      KEY: "value"
```

### Configuration Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Unique identifier for the command/tool |
| `script` | string | Yes | Path to executable or command to run |
| `args` | []string | No | Arguments to pass to the script |
| `description` | string | No | Description shown to MCP clients |
| `timeout` | string | No | Execution timeout (e.g., "30s", "5m") |
| `container` | string | No | Docker image for containerized execution |
| `env` | map[string]string | No | Environment variables |

### Example Configurations

#### Local Command Execution
```yaml
commands:
  - name: list-files
    script: ls
    args: ["-la"]
    description: "List files in current directory"
    timeout: "10s"
```

#### Containerized Execution
```yaml
commands:
  - name: python-analysis
    script: python
    args: ["/app/analyze.py"]
    description: "Run data analysis in isolated Python environment"
    container: "python:3.9-slim"
    timeout: "5m"
    env:
      PYTHONPATH: "/app"
      DATA_SOURCE: "production"
```

#### Web Screenshot Tool
```yaml
commands:
  - name: screenshot
    script: "/usr/bin/chromium"
    args: ["--headless", "--disable-gpu", "--screenshot"]
    description: "Capture webpage screenshots"
    container: "browserless/chrome:latest"
    timeout: "60s"
    env:
      DISPLAY: ":99"
```

## Usage

### MCP Server Mode

Start MCPFier as an MCP server (primary mode):

```bash
./mcpfier --mcp
```

The server will:
- Load configuration from `config.yaml`
- Register each command as an MCP tool
- Listen on stdio for MCP protocol messages
- Execute commands based on client requests

### Legacy Command Mode

MCPFier maintains backward compatibility for direct command execution:

```bash
./mcpfier command-name
```

This mode executes the specified command directly without MCP protocol overhead.

### Integration with MCP Clients

MCPFier can be integrated with any MCP-compatible client. Example client configuration:

```json
{
  "name": "mcpfier",
  "command": "/path/to/mcpfier",
  "args": ["--mcp"],
  "type": "stdio"
}
```

## Security Considerations

### Local Execution Security
- Commands run with the same privileges as the MCPFier process
- Suitable for trusted environments and internal tooling
- Consider running MCPFier with restricted user privileges

### Container Execution Security
- Commands execute within isolated Docker containers
- No access to host filesystem by default
- Network isolation available through Docker configuration
- Resource limits can be imposed via container runtime

### Best Practices
1. Use containerized execution for untrusted or external commands
2. Define explicit timeouts for all commands
3. Limit environment variable exposure
4. Use non-root containers when possible
5. Implement proper logging and monitoring

## Testing

Run the test suite:

```bash
go test ./...
```

Tests cover:
- Configuration loading and validation
- Command execution (both local and containerized)
- MCP protocol compliance
- Error handling scenarios

## Use Cases

### Enterprise Integration
- Expose n8n workflows as MCP tools
- Provide LLM access to internal APIs and databases
- Standardize data pipeline execution
- Enable secure AI agent interactions

### Development Tools
- Code analysis and linting tools
- Build and deployment automation
- Testing framework integration
- Documentation generation

### Data Processing
- ETL pipeline execution
- Report generation
- Data validation and cleaning
- Analytics and visualization

### Infrastructure Operations
- Health checks and monitoring
- Log analysis and alerting
- Backup and recovery operations
- System administration tasks

## Roadmap

### Phase 1 (Current)
- Basic MCP server implementation
- Local and containerized execution
- YAML configuration support

### Phase 2 (Planned)
- Remote server capabilities with authentication
- Resource usage monitoring and limits
- Command chaining and workflows
- Advanced security policies

### Phase 3 (Future)
- Kubernetes job execution
- Scheduled task support
- LLM-to-LLM communication patterns
- Web UI for configuration management

## API Reference

### MCP Tools

Each configured command becomes an MCP tool with:
- **Name**: Matches the command name from configuration
- **Description**: Uses the description field or auto-generates
- **Input Schema**: Accepts additional arguments (future enhancement)
- **Output**: Returns command output as text content

### Error Handling

MCPFier returns structured error responses for:
- Command not found
- Execution failures
- Timeout exceeded
- Container runtime errors

## Troubleshooting

### Common Issues

**Command not found**: Verify the command name exists in `config.yaml`

**Container execution fails**: Ensure Docker is running and the specified image exists

**Permission denied**: Check file permissions and user privileges

**Timeout exceeded**: Increase timeout value or optimize command performance

### Debug Mode

Enable verbose logging by setting environment variables:
```bash
export MCP_DEBUG=1
./mcpfier --mcp
```

## Contributing

We welcome contributions to MCPFier. Please:

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Submit a pull request with clear description

## License

MCPFier is released under the MIT License. See LICENSE file for details.

## Support

For issues, feature requests, or questions:
- Open an issue on GitHub
- Check existing documentation
- Review configuration examples