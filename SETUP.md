# MCPFier Claude Desktop Setup Guide

## Quick Setup

1. **Build MCPFier**:
   ```bash
   go build -o mcpfier
   ```

2. **Install system-wide config**:
   ```bash
   mkdir -p ~/.mcpfier
   cp config.yaml ~/.mcpfier/
   ```

3. **Get setup instructions**:
   ```bash
   ./mcpfier --setup
   ```

4. **Add to Claude Desktop**: Copy the JSON configuration from step 3 to your Claude Desktop MCP settings

## Claude Desktop Configuration

Open Claude Desktop settings and add this to your MCP servers configuration:

```json
{
  "mcpServers": {
    "mcpfier": {
      "command": "/full/path/to/mcpfier",
      "args": ["--mcp"]
    }
  }
}
```

Replace `/full/path/to/mcpfier` with the actual path shown by `./mcpfier --setup`.

## Configuration File Locations

MCPFier searches for `config.yaml` in this order:

1. Current working directory: `./config.yaml`
2. Next to executable: `/path/to/mcpfier/config.yaml`  
3. User home directory: `~/.mcpfier/config.yaml`
4. User mcpfier directory: `~/mcpfier/config.yaml`

## Testing

After setup, test in Claude Desktop with:

- "Use the echo-test tool"
- "Use the get-weather tool"
- "Use the list-files tool"

## Troubleshooting

### Config File Not Found

```
2025/08/07 14:12:19 Failed to create MCP server: failed to load config: open config.yaml: no such file or directory
```

**Solution**: Copy config to user directory:

```bash
mkdir -p ~/.mcpfier
cp config.yaml ~/.mcpfier/
```

### Permission Denied

**Solution**: Make binary executable:
```bash
chmod +x mcpfier
```

### Docker Commands Fail

**Solution**: Pull required images:
```bash
docker pull python:3.9-slim
docker pull browserless/chrome:latest
```

### Server Disconnects Immediately

**Solution**: Check Claude Desktop logs and ensure:

- Binary path is correct
- Config file exists and is readable
- No syntax errors in config.yaml

## Advanced Configuration

### Custom Config Location

Set environment variable:

```bash
export MCPFIER_CONFIG="/custom/path/config.yaml"
```

### System-wide Installation

```bash
# Install binary
sudo cp mcpfier /usr/local/bin/

# Install config
sudo mkdir -p /etc/mcpfier
sudo cp config.yaml /etc/mcpfier/

# Use in Claude Desktop
{
  "mcpServers": {
    "mcpfier": {
      "command": "/usr/local/bin/mcpfier",
      "args": ["--mcp"]
    }
  }
}
```
