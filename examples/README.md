# MCPFier Client Examples

This directory contains examples of how to connect to MCPFier using different authentication methods and transports.

## Available Examples

### 1. Simple STDIO Client (Current Default)
**File:** `stdio_client/main.go`  
**Description:** Basic client using STDIO transport (no authentication required)  
**Use Case:** Local development, single-user scenarios

```bash
cd examples/stdio_client
go run main.go
```

### 2. API Key HTTP Client
**File:** `apikey_client/main.go`  
**Description:** HTTP client using API key authentication  
**Use Case:** Simple server-to-server authentication

```bash
# Start MCPFier HTTP server
./mcpfier --mcp --transport http

# Run client
cd examples/apikey_client  
export MCP_API_KEY="mcpfier_secret_123"
go run main.go
```

### 3. OAuth 2.1 HTTP Client  
**File:** `oauth_client/main.go`  
**Description:** Full OAuth 2.1 client with PKCE and browser-based auth  
**Use Case:** Enterprise integrations, multi-user scenarios

```bash
# Start MCPFier HTTP server with OAuth
./mcpfier --mcp --transport http --auth oauth

# Run client (opens browser for OAuth flow)
cd examples/oauth_client
export MCP_CLIENT_ID="your-client-id"
export MCP_CLIENT_SECRET="your-client-secret"
go run main.go
```

### 4. Multi-Transport Client
**File:** `multi_transport_client/main.go`  
**Description:** Client that can connect via both STDIO and HTTP  
**Use Case:** Development tools that support multiple deployment modes

```bash
cd examples/multi_transport_client

# Connect via STDIO
go run main.go --transport stdio

# Connect via HTTP with API key
go run main.go --transport http --api-key "mcpfier_secret_123"

# Connect via HTTP with OAuth
go run main.go --transport http --oauth
```

## Authentication Flow Examples

### API Key Authentication Flow
```
1. Client → HTTP Request with X-API-Key header
2. MCPFier → Validates API key against config
3. MCPFier → Executes tool if authorized  
4. MCPFier → Returns result to client
```

### OAuth 2.1 Authentication Flow
```
1. Client → Attempts MCP connection
2. MCPFier → Returns 401 with WWW-Authenticate header
3. Client → Opens browser to authorization URL
4. User → Completes OAuth flow in browser
5. Client → Receives authorization code
6. Client → Exchanges code for access token
7. Client → Retries MCP connection with Bearer token
8. MCPFier → Validates token and allows access
```

## Testing Authentication

### Test OAuth Flow End-to-End
```bash
# Terminal 1: Start MCPFier with OAuth
./mcpfier --mcp --transport http --auth oauth

# Terminal 2: Run OAuth client test
cd examples/oauth_client
go run main.go

# Browser will open automatically for OAuth flow
# Complete login and return to terminal
```

### Test API Key Authentication
```bash
# Terminal 1: Start MCPFier with API keys
./mcpfier --mcp --transport http --auth apikey

# Terminal 2: Test valid API key
curl -H "X-API-Key: mcpfier_secret_123" \
     -H "Content-Type: application/json" \
     -d '{"jsonrpc":"2.0","method":"tools/list","id":1}' \
     http://localhost:8080/mcp

# Terminal 3: Test invalid API key (should fail)
curl -H "X-API-Key: invalid_key" \
     -H "Content-Type: application/json" \
     -d '{"jsonrpc":"2.0","method":"tools/list","id":1}' \
     http://localhost:8080/mcp
```

## Claude Desktop Integration

### STDIO Integration (Current)
```json
{
  "mcpServers": {
    "mcpfier": {
      "command": "/path/to/mcpfier",
      "args": ["--mcp"]
    }
  }
}
```

### HTTP with API Key Integration
```json
{
  "mcpServers": {
    "mcpfier": {
      "command": "curl",
      "args": [
        "-H", "X-API-Key: mcpfier_secret_123",
        "-H", "Content-Type: application/json",
        "http://localhost:8080/mcp"
      ]
    }
  }
}
```

### HTTP with OAuth Integration
```json
{
  "mcpServers": {
    "mcpfier": {
      "command": "/path/to/oauth_client_wrapper",
      "args": ["http://localhost:8080/mcp"]
    }
  }
}
```

## Development and Testing

### Run All Examples
```bash
make test-examples
```

### Individual Example Testing
```bash
# Test STDIO client
make test-stdio-client

# Test API key client  
make test-apikey-client

# Test OAuth client
make test-oauth-client
```

### Integration Tests
```bash
# Test complete auth flows
make test-auth-flows

# Performance testing
make bench-auth-performance
```

## Configuration Examples

### MCPFier Server Configuration
```yaml
# config.yaml
server:
  transport: "http"  # or "stdio" or "both"
  
  http:
    port: 8080
    auth:
      enabled: true
      methods: ["api_key", "oauth"]
      
      api_keys:
        - key: "mcpfier_secret_123"
          name: "development_client"
          permissions: ["weather", "echo-test"]
          
      oauth:
        issuer: "http://localhost:8080"
        audience: ["http://localhost:8080/mcp"]
        jwks_uri: "http://localhost:8080/.well-known/jwks.json"

commands:
  - name: echo-test
    script: echo
    args: ["Hello from authenticated MCPFier!"]
    description: "Test command for auth verification"
```

## Troubleshooting

### Common Issues

1. **"Missing authentication" error**
   - Check API key is included in request headers
   - Verify API key matches server configuration

2. **OAuth flow fails**
   - Ensure redirect URI matches exactly
   - Check client ID and secret are correct
   - Verify authorization server is accessible

3. **Token expired** 
   - OAuth clients should automatically refresh tokens
   - Check token store implementation

### Debug Mode
```bash
# Enable debug logging
export MCP_DEBUG=true
go run main.go
```

### Verbose OAuth Logging
```bash
# See detailed OAuth flow
export OAUTH_DEBUG=true
cd examples/oauth_client
go run main.go
```

## Contributing

When adding new authentication examples:

1. Follow the existing directory structure
2. Include comprehensive error handling
3. Add configuration documentation
4. Create integration tests
5. Update this README with usage instructions

## Security Best Practices

### API Key Security
- Store API keys in environment variables, not code
- Use different keys for different environments  
- Rotate keys regularly
- Limit key permissions to minimum required

### OAuth Security
- Always use PKCE for public clients
- Validate state parameter to prevent CSRF
- Use short-lived access tokens
- Store tokens securely (encrypted at rest)

### Transport Security
- Use HTTPS in production
- Validate TLS certificates
- Consider mutual TLS for high-security environments