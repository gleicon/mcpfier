# MCPFier

MCPFier transforms any command, script, or tool into a standardized MCP (Model Context Protocol) server that LLMs can use seamlessly.

Think "GitHub Actions for MCP" - configure once, use everywhere. Check the bundled config for examples.

## Quick Start

1. **Configure commands** in `config.yaml`:

   ```yaml
   commands:
     - name: get-weather
       script: curl
       args: ["https://wttr.in/?format=3"]
       description: "Get current weather"
   ```

2. **Generate setup instructions**:

   ```bash
   ./mcpfier --setup
   ```

3. **Add to Claude Desktop** using the JSON configuration from step 2

4. **Use in Claude**: "Use the get-weather tool"

## Features

- **Universal Tool Interface**: Any script becomes an MCP tool
- **Dual Execution**: Local commands or isolated Docker containers
- **Zero Configuration**: Automatic setup for Claude Desktop
- **Enterprise Ready**: Security, logging, and resource management
- **Legacy Compatible**: Works as traditional CLI tool

## Configuration

Commands are defined in `config.yaml`:

```yaml
commands:
  # Local execution
  - name: list-files
    script: ls
    args: ["-la"]
    description: "List directory contents"
    timeout: "10s"

  # Containerized execution  
  - name: python-analysis
    script: python
    args: ["/app/analyze.py"]
    description: "Run data analysis"
    container: "python:3.9-slim"
    timeout: "5m"
    env:
      DATA_SOURCE: "production"
```

### Configuration Fields

| Field         | Required | Description                |
| ------------- | -------- | -------------------------- |
| `name`        | Yes      | Unique tool identifier     |
| `script`      | Yes      | Command or executable path |
| `args`        | No       | Command arguments          |
| `description` | No       | Tool description for LLMs  |
| `container`   | No       | Docker image for isolation |
| `timeout`     | No       | Execution timeout          |
| `env`         | No       | Environment variables      |

## Installation

```bash
git clone https://github.com/gleicon/mcpfier.git
cd mcpfier
go build -o mcpfier
```

## Usage Modes

### MCP Server (Primary)

```bash
./mcpfier --mcp    # Start MCP server
./mcpfier --setup  # Generate Claude Desktop config
```

### CLI Tool (Legacy)

```bash
./mcpfier command-name  # Execute command directly
```

## Architecture

MCPFier uses a clean modular architecture:

- **Configuration**: YAML-based command definitions with auto-discovery
- **Execution**: Pluggable backends (local, Docker, future: Kubernetes, Lambda)
- **MCP Server**: Protocol-compliant stdio server using mark3labs/mcp-go
- **Security**: Container isolation, resource limits, sandboxing

## Security

- **Local Mode**: Fast execution, same privileges as MCPFier
- **Container Mode**: Complete isolation, no host access, resource limits
- **Best Practices**: Use containers for untrusted code, define timeouts, limit privileges

## Use Cases

- **Enterprise Integration**: Expose n8n workflows, internal APIs, data pipelines
- **Development Tools**: Linting, testing, building, deployment automation  
- **Infrastructure**: Health checks, log analysis, backup operations
- **AI Agents**: Enable LLMs to use specialized tools and services

## Testing

```bash
go test ./...              # Run all tests
./mcpfier echo-test        # Test CLI mode
./mcpfier --setup          # Test MCP configuration
```

## Documentation

- **[SETUP.md](SETUP.md)** - Detailed setup instructions
- **[ARCHITECTURE.md](ARCHITECTURE.md)** - Technical architecture
- **[SECURITY.md](SECURITY.md)** - Security guide and best practices
- **[ROADMAP.md](ROADMAP.md)** - Future development plans

## Contributing

We welcome contributions! Please see our documentation for:
- Architecture decisions and design patterns
- Security considerations and testing requirements
- Roadmap and planned features

## License

MIT License - see LICENSE file for details.