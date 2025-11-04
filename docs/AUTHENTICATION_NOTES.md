# Authentication Implementation Status

Version: v1.0.4
Last Updated: 2025-11-02

## Current Implementation

API key authentication is fully implemented and working in MCPFier.

### What's Implemented

1. **API Key Configuration** (config.yaml)

MCPFier supports three API key storage formats:

**A. Plain Text Keys** (simple, for development):
```yaml
server:
  http:
    auth:
      enabled: true
      mode: "simple"
      simple:
        api_keys:
          - key: "mcpfier_dev_123456"
            name: "development"
            permissions: ["*"]
```

**B. Environment Variables** (recommended for production):
```yaml
server:
  http:
    auth:
      enabled: true
      mode: "simple"
      simple:
        api_keys:
          - key: "${MCPFIER_API_KEY}"
            name: "production"
            permissions: ["*"]
```
Then set: `export MCPFIER_API_KEY="your-secure-key"`

**C. Hashed Keys** (most secure, one-way storage):
```yaml
server:
  http:
    auth:
      enabled: true
      mode: "simple"
      simple:
        api_keys:
          - key: "hash:95e904704170a1539c636551759916c31758d6302b43c4e80e11a44a0a2fbb40"
            name: "production"
            permissions: ["*"]
```
Use `./bin/hashkey -generate` to create hashed keys.

2. **Authentication Middleware** (internal/auth/auth.go)
- Extracts API key from X-API-Key header or Authorization header
- Validates against configured keys
- Returns 401 Unauthorized for invalid/missing keys
- Adds authentication context to request

3. **Permission Checking** (internal/server/http_server.go)
- Validates user has permission for each tool/command
- Supports wildcard permissions ("*")
- Supports per-tool permissions
- Returns permission denied error if unauthorized

4. **HTTP Server Integration**
- Auth middleware applied to all MCP endpoints
- Health check endpoint excluded from auth
- Analytics endpoint excluded from auth
- Logs authentication method and success/failure

### Code Locations

- Authentication logic: internal/auth/auth.go
- HTTP server integration: internal/server/http_server.go
- Configuration structures: internal/config/config.go
- Secrets handling: internal/config/secrets.go
- Key generation utility: cmd/hashkey/main.go
- Context functions: auth.ContextFunc() for mcp-go integration

### How It Works

1. Client sends request with API key:
   - Header: X-API-Key: mcpfier_dev_123456
   - Or: Authorization: ApiKey mcpfier_dev_123456

2. Auth middleware validates key:
   - Checks if key exists in config
   - Creates AuthContext with user info and permissions
   - Adds context to request

3. Command execution checks permissions:
   - Extracts AuthContext from request
   - Verifies user has permission for tool
   - Returns error if permission denied

4. Analytics tracks auth events:
   - Records auth method (api_key)
   - Records success/failure
   - Tracks per-client usage

### Testing

Test with curl:
```bash
# Valid API key
curl -X POST http://localhost:8080/mcp \
  -H "X-API-Key: mcpfier_dev_123456" \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"tools/list","id":1}'

# Invalid API key (should return 401)
curl -X POST http://localhost:8080/mcp \
  -H "X-API-Key: invalid_key" \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"tools/list","id":1}'

# No API key (should return 401)
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"tools/list","id":1}'
```

## What's NOT Implemented

The following are future enhancements (not currently needed):

- JWT token authentication
- OAuth2/OIDC integration
- RBAC (role-based access control)
- Multi-tenant isolation
- Session management
- Token refresh/rotation
- MCP Gateway proxy features

These features are documented in future roadmap but are not required for current use cases.

## Key Management Utility

Generate and hash API keys using the `hashkey` utility:

```bash
# Build the utility
go build -o bin/hashkey ./cmd/hashkey

# Generate a new random API key (with hash)
./bin/hashkey -generate

# Hash an existing API key
./bin/hashkey -hash "your-existing-key"

# Generate longer key (48 bytes)
./bin/hashkey -generate -length 48
```

The utility outputs:
- Plain text key (share with clients)
- Hashed version (store in config.yaml)
- Configuration examples for all three formats

## Configuration Examples

### Development Setup (Plain Text)
```yaml
auth:
  enabled: true
  mode: "simple"
  simple:
    api_keys:
      - key: "dev_key_123"
        name: "dev-environment"
        permissions: ["*"]
```

### Production Setup (Environment Variables)
```yaml
auth:
  enabled: true
  mode: "simple"
  simple:
    api_keys:
      - key: "${MCPFIER_CLIENT1_KEY}"
        name: "client1"
        permissions: ["*"]
      
      - key: "${MCPFIER_CLIENT2_KEY}"
        name: "client2"
        permissions: ["search-products", "get-product"]
```

Then in your environment:
```bash
export MCPFIER_CLIENT1_KEY="cuhw4RIKLaBkACH4hBlkfPbRHPaqxxc-AdUFV2qVQA"
export MCPFIER_CLIENT2_KEY="another-secure-random-key"
```

### High Security Setup (Hashed Keys)
```yaml
auth:
  enabled: true
  mode: "simple"
  simple:
    api_keys:
      - key: "hash:95e904704170a1539c636551759916c31758d6302b43c4e80e11a44a0a2fbb40"
        name: "client1"
        permissions: ["*"]
      
      - key: "hash:a1b2c3d4e5f6..."
        name: "client2-readonly"
        permissions: ["search-products", "get-product"]
```

### Mixed Setup (Different Formats)
```yaml
auth:
  enabled: true
  mode: "simple"
  simple:
    api_keys:
      # Development key (plain text)
      - key: "dev_key_local"
        name: "local-dev"
        permissions: ["*"]
      
      # Production key from environment
      - key: "${PROD_API_KEY}"
        name: "prod-client"
        permissions: ["*"]
      
      # Shared key with hash
      - key: "hash:95e904704170a1539c636551759916c31758d6302b43c4e80e11a44a0a2fbb40"
        name: "partner-integration"
        permissions: ["search-products"]
```

### Authentication Disabled (Local Development Only)
```yaml
auth:
  enabled: false
```

## Security Considerations

### Key Storage Formats

1. **Plain Text Keys**: Suitable for development only. Keys are stored in config.yaml in readable form.

2. **Environment Variables**: Recommended for production. Keys stored in environment, not in config files. Config files can be committed to git safely.

3. **Hashed Keys**: Most secure. Config file contains only SHA-256 hash. Even if config is compromised, original key cannot be recovered. Clients still use plain text key for requests.

### Why SHA-256 (Not bcrypt)?

- API keys are validated on every request (high frequency)
- bcrypt is intentionally slow (~100ms per validation)
- 100 concurrent requests with bcrypt = ~2GB RAM usage
- SHA-256 is fast (<1ms) and sufficient for high-entropy random keys
- API keys should be cryptographically random (32+ bytes), unlike user passwords

### Best Practices

1. **Generate Strong Keys**: Use `./bin/hashkey -generate` to create cryptographically random keys
2. **Use Environment Variables in Production**: Never commit production keys to git
3. **Rotate Keys Regularly**: Update keys periodically, especially after team member changes
4. **Use Hashed Storage**: For shared configs or compliance requirements
5. **Limit Permissions**: Use specific tool permissions instead of wildcard ("*") when possible
6. **Monitor Access**: Use analytics to track API key usage and detect anomalies

### Migration Path

Existing configurations with plain text keys continue to work. To migrate:

1. Generate new keys: `./bin/hashkey -generate`
2. Update config.yaml with hashed or environment variable format
3. Share plain text key with clients (they don't change their code)
4. Restart mcpfier server
5. Test authentication still works

## Summary

API key authentication is production-ready in MCPFier v1.0.4:
- Three storage formats (plain text, environment variables, hashed)
- Middleware validates all requests
- Permission checking per tool
- Analytics tracking
- Key generation utility included
- Backward compatible with existing configs
- No TODO items remaining

For future OAuth2/RBAC features, see ROADMAP.md.
