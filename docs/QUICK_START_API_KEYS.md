# Quick Start: API Keys

## 5-Minute Setup

### 1. Build the Key Generator

```bash
cd /path/to/mcpfier
go build -o bin/hashkey ./cmd/hashkey
```

### 2. Generate a Key

```bash
./bin/hashkey -generate
```

Copy the plain text key from output.

### 3. Add to config.yaml

Open config.yaml and update the auth section:

```yaml
server:
  http:
    auth:
      enabled: true
      mode: "simple"
      simple:
        api_keys:
          - key: "paste-plain-text-key-here"
            name: "my-client"
            permissions: ["*"]
```

### 4. Start MCPFier

```bash
./bin/mcpfier --server
```

### 5. Test Authentication

```bash
curl -X POST http://localhost:8080/mcp \
  -H "X-API-Key: paste-plain-text-key-here" \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"tools/list","id":1}'
```

Should return list of available tools.

## Next Steps

### For Production

Use environment variables instead of plain text:

1. Set environment variable:
```bash
export MCPFIER_API_KEY="your-generated-key"
```

2. Update config.yaml:
```yaml
api_keys:
  - key: "${MCPFIER_API_KEY}"
    name: "production"
    permissions: ["*"]
```

3. Config file is now safe to commit to git.

### For Maximum Security

Use hashed keys:

1. Generate key:
```bash
./bin/hashkey -generate
```

2. Copy the hashed version (starts with "hash:")

3. Update config.yaml:
```yaml
api_keys:
  - key: "hash:95e904704170a1539c636551759916c31758d6302b43c4e80e11a44a0a2fbb40"
    name: "production"
    permissions: ["*"]
```

4. Share plain text key with client

5. Client uses plain text key, server stores only hash

## Common Patterns

### Multiple Clients

```yaml
api_keys:
  - key: "${CLIENT1_KEY}"
    name: "client1"
    permissions: ["*"]
  
  - key: "${CLIENT2_KEY}"
    name: "client2"
    permissions: ["get-weather", "echo-test"]
```

### Development and Production

```yaml
api_keys:
  # Local development
  - key: "dev_local_key_123"
    name: "local-dev"
    permissions: ["*"]
  
  # Production
  - key: "${PROD_API_KEY}"
    name: "production"
    permissions: ["*"]
```

### Key Rotation

Add new key while keeping old:

```yaml
api_keys:
  - key: "old_key_still_works"
    name: "client1"
    permissions: ["*"]
  
  - key: "new_key_also_works"
    name: "client1"
    permissions: ["*"]
```

Both keys work. Remove old key after clients migrate.

## Client Integration

### curl

```bash
curl -H "X-API-Key: your-key" http://localhost:8080/mcp
```

### Python

```python
import requests

headers = {
    'X-API-Key': 'your-key',
    'Content-Type': 'application/json'
}

response = requests.post(
    'http://localhost:8080/mcp',
    headers=headers,
    json={'jsonrpc': '2.0', 'method': 'tools/list', 'id': 1}
)
```

### JavaScript

```javascript
const response = await fetch('http://localhost:8080/mcp', {
  method: 'POST',
  headers: {
    'X-API-Key': 'your-key',
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    jsonrpc: '2.0',
    method: 'tools/list',
    id: 1
  })
});
```

## Troubleshooting

### 401 Unauthorized

Check:
1. Auth is enabled in config.yaml
2. Key matches exactly (copy-paste to avoid typos)
3. Server restarted after config change

### Environment variable not found

```bash
# Verify variable is set
echo $MCPFIER_API_KEY

# Set if empty
export MCPFIER_API_KEY="your-key"

# Restart server
```

### Invalid hash error

Hash must be 64 hexadecimal characters. Use hashkey utility to generate:

```bash
./bin/hashkey -hash "your-key"
```

## Complete Example

### Setup

```bash
# Build utilities
go build -o bin/mcpfier .
go build -o bin/hashkey ./cmd/hashkey

# Generate key
./bin/hashkey -generate > key.txt

# Extract plain text key
PLAIN_KEY=$(grep "Plain text:" key.txt | cut -d' ' -f3)

# Set environment variable
export MCPFIER_API_KEY="$PLAIN_KEY"
```

### config.yaml

```yaml
server:
  http:
    port: 8080
    auth:
      enabled: true
      mode: "simple"
      simple:
        api_keys:
          - key: "${MCPFIER_API_KEY}"
            name: "main-client"
            permissions: ["*"]
```

### Start Server

```bash
./bin/mcpfier --server
```

### Test

```bash
curl -X POST http://localhost:8080/mcp \
  -H "X-API-Key: $MCPFIER_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"tools/list","id":1}'
```

## Documentation

- Full guide: docs/API_KEY_STORAGE.md
- Authentication details: docs/AUTHENTICATION_NOTES.md
- Security considerations: docs/SECURITY.md
