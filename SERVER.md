# MCPFier HTTP Server & Authentication Guide

## Overview

MCPFier supports two transport modes:

- **STDIO Mode (`--mcp`)**: Single client, no authentication, perfect for Claude Desktop
- **HTTP Server Mode (`--server`)**: Multiple clients, authentication, enterprise features

This document covers the HTTP server mode with authentication, built on the [MCP 2025-06-18 specification](https://modelcontextprotocol.io/specification/2025-06-18) using the [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go) library.

## Quick Start

### 1. Start HTTP Server

```bash
# Start with default configuration
./mcpfier --server

# Start with custom configuration
./mcpfier --server --config /path/to/config.yaml
```

### 2. Test Authentication

```bash
# Test with API key
curl -H "X-API-Key: mcpfier_dev_123456" \
     -H "Content-Type: application/json" \
     -d '{"jsonrpc":"2.0","method":"tools/list","id":1}' \
     http://localhost:8080/mcp

# Test OAuth discovery endpoints
curl http://localhost:8080/.well-known/oauth-authorization-server
curl http://localhost:8080/.well-known/oauth-protected-resource
```

## Transport Modes Comparison

| Feature            | STDIO Mode (`--mcp`)     | HTTP Server Mode (`--server`)    |
| ------------------ | ------------------------ | -------------------------------- |
| **Clients**        | Single (Claude Desktop)  | Multiple (any HTTP client)       |
| **Authentication** | None (environment-based) | API Keys, OAuth 2.1              |
| **Protocol**       | MCP over STDIO           | MCP over HTTP                    |
| **Use Case**       | Local development        | Enterprise deployment            |
| **Security**       | Process-level            | Transport-level + authentication |
| **Scalability**    | Single user              | Multi-user, multi-tenant         |

## Authentication Modes

### Simple Mode (Recommended for Getting Started)

Perfect for development and simple deployments. Uses API keys and optional basic authentication.

**Configuration:**

```yaml
server:
  http:
    auth:
      mode: "simple"
      simple:
        api_keys:
          - key: "mcpfier_dev_123456"
            name: "development"
            permissions: ["*"]  # All tools
          - key: "mcpfier_prod_789012"
            name: "production" 
            permissions: ["weather", "echo-test"]  # Limited tools
```

**Client Usage:**

```bash
# API Key in header
curl -H "X-API-Key: mcpfier_dev_123456" http://localhost:8080/mcp

# API Key in Authorization header
curl -H "Authorization: ApiKey mcpfier_dev_123456" http://localhost:8080/mcp
```

### Enterprise Mode (Full OAuth 2.1 Compliance)

Full OAuth 2.1 implementation with PKCE, role-based access control, and enterprise features.

**Configuration:**

```yaml
server:
  http:
    auth:
      mode: "enterprise"
      enterprise:
        oauth21:
          enabled: true
          issuer: "https://auth.yourcompany.com"
          audience: ["https://mcpfier.yourcompany.com/api"]
          client_id: "${OAUTH_CLIENT_ID}"
          client_secret: "${OAUTH_CLIENT_SECRET}"
        rbac:
          enabled: true
          roles:
            - name: "admin"
              permissions: ["*"]
            - name: "developer"
              permissions: ["echo-test", "weather"]
```

**Client Usage:**

```go
// Using mcp-go OAuth client
oauthConfig := client.OAuthConfig{
    ClientID:     os.Getenv("MCP_CLIENT_ID"),
    ClientSecret: os.Getenv("MCP_CLIENT_SECRET"),
    RedirectURI:  "http://localhost:8085/callback",
    Scopes:       []string{"mcp:read", "mcp:execute"},
    PKCEEnabled:  true,
}

c, err := client.NewOAuthStreamableHttpClient(
    "http://localhost:8080/mcp", oauthConfig)
```

## Configuration Reference

### Server Configuration

```yaml
server:
  http:
    port: 8080                    # HTTP server port
    host: "localhost"             # Bind address
    
    auth:
      enabled: true               # Enable authentication
      mode: "simple"              # "simple" or "enterprise"
      
    cors:
      enabled: true               # Enable CORS for web clients
      allowed_origins: ["*"]      # Allowed origins
      
    rate_limit:
      enabled: true               # Enable rate limiting
      requests_per_minute: 60     # Rate limit per client
```

### Simple Authentication

```yaml
server:
  http:
    auth:
      simple:
        # API key authentication
        api_keys:
          - key: "your-api-key-here"
            name: "client-name"
            description: "Client description"
            permissions: ["tool1", "tool2"]  # or ["*"] for all
            
        # Optional: Basic authentication
        basic_auth:
          enabled: false
          users:
            - username: "admin"
              password_hash: "$2a$10$..."  # bcrypt hash
              permissions: ["*"]
```

### Enterprise Authentication

```yaml
server:
  http:
    auth:
      enterprise:
        # OAuth 2.1 configuration
        oauth21:
          enabled: true
          issuer: "https://auth.example.com"
          audience: ["https://mcpfier.example.com/api"]
          client_id: "${OAUTH_CLIENT_ID}"
          client_secret: "${OAUTH_CLIENT_SECRET}"
          scopes: ["mcp:read", "mcp:execute"]
          
        # Role-based access control
        rbac:
          enabled: true
          roles:
            - name: "admin"
              permissions: ["*"]
            - name: "user"
              permissions: ["weather", "echo-test"]
          user_roles:
            "user@example.com": ["user"]
            "admin@example.com": ["admin"]
```

## MCP 2025-06-18 Specification Compliance

MCPFier HTTP server fully complies with the MCP 2025-06-18 authentication specification:

### ✅ OAuth 2.1 Requirements
- **Resource Server**: MCPFier acts as OAuth 2.1 resource server
- **Bearer Token Validation**: Supports `Authorization: Bearer <token>` headers
- **PKCE Support**: Full Proof Key for Code Exchange implementation
- **Token Audience Validation**: Prevents confused deputy attacks
- **Authorization Server Discovery**: RFC9728 and RFC8414 compliance

### ✅ Discovery Endpoints

**Authorization Server Metadata (RFC8414):**
```
GET /.well-known/oauth-authorization-server
```

**Protected Resource Metadata (RFC9728):**
```
GET /.well-known/oauth-protected-resource
```

### ✅ Transport-Specific Authentication
- **HTTP Transport**: Full OAuth 2.1 compliance
- **STDIO Transport**: Environment-based credentials (MCP spec compliant)

## Security Best Practices

### API Key Security
1. **Generate Strong Keys**: Use cryptographically secure random strings (32+ chars)
2. **Environment Variables**: Store keys in environment variables, never in code
3. **Key Rotation**: Regularly rotate API keys
4. **Principle of Least Privilege**: Limit permissions to minimum required

```bash
# Generate secure API key
openssl rand -base64 32
```

### OAuth 2.1 Security
1. **Use HTTPS**: Always use TLS in production
2. **PKCE Required**: Enable PKCE for all clients
3. **Short-Lived Tokens**: Use access tokens with short expiration
4. **Secure Storage**: Store refresh tokens securely

### Network Security
1. **TLS Termination**: Use reverse proxy with TLS
2. **Rate Limiting**: Prevent abuse with rate limits
3. **CORS Configuration**: Restrict origins to known clients
4. **Firewall Rules**: Limit access to trusted networks

## Deployment Scenarios

### 1. Development Environment

**Simple API Key Setup:**
```yaml
server:
  http:
    port: 8080
    auth:
      mode: "simple"
      simple:
        api_keys:
          - key: "dev_key_12345"
            name: "development"
            permissions: ["*"]
```

**Usage:**
```bash
export MCP_API_KEY="dev_key_12345"
curl -H "X-API-Key: $MCP_API_KEY" http://localhost:8080/mcp
```

### 2. Production Deployment

**Enterprise OAuth Setup:**
```yaml
server:
  http:
    port: 8080
    auth:
      mode: "enterprise"
      enterprise:
        oauth21:
          issuer: "https://auth.company.com"
          audience: ["https://mcpfier.company.com/api"]
          client_id: "${OAUTH_CLIENT_ID}"
        rbac:
          enabled: true
```

**Docker Deployment:**
```dockerfile
FROM golang:1.21-alpine AS builder
COPY . /app
WORKDIR /app
RUN go build -o mcpfier

FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/mcpfier /usr/local/bin/
COPY --from=builder /app/config.yaml /etc/mcpfier/
EXPOSE 8080
CMD ["mcpfier", "--server", "--config", "/etc/mcpfier/config.yaml"]
```

### 3. Kubernetes Deployment

**Deployment YAML:**
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mcpfier-server
spec:
  replicas: 3
  selector:
    matchLabels:
      app: mcpfier-server
  template:
    metadata:
      labels:
        app: mcpfier-server
    spec:
      containers:
      - name: mcpfier
        image: mcpfier:latest
        args: ["--server"]
        ports:
        - containerPort: 8080
        env:
        - name: OAUTH_CLIENT_ID
          valueFrom:
            secretKeyRef:
              name: mcpfier-oauth
              key: client-id
        - name: OAUTH_CLIENT_SECRET
          valueFrom:
            secretKeyRef:
              name: mcpfier-oauth
              key: client-secret
---
apiVersion: v1
kind: Service
metadata:
  name: mcpfier-service
spec:
  selector:
    app: mcpfier-server
  ports:
  - port: 80
    targetPort: 8080
  type: LoadBalancer
```

## Client Integration Examples

### 1. Go Client with OAuth

```go
package main

import (
    "context"
    "os"
    "github.com/mark3labs/mcp-go/client"
    "github.com/mark3labs/mcp-go/mcp"
)

func main() {
    tokenStore := client.NewMemoryTokenStore()
    
    oauthConfig := client.OAuthConfig{
        ClientID:     os.Getenv("MCP_CLIENT_ID"),
        ClientSecret: os.Getenv("MCP_CLIENT_SECRET"),
        RedirectURI:  "http://localhost:8085/callback",
        Scopes:       []string{"mcp:read", "mcp:execute"},
        TokenStore:   tokenStore,
        PKCEEnabled:  true,
    }
    
    c, err := client.NewOAuthStreamableHttpClient(
        "https://mcpfier.company.com/mcp", oauthConfig)
    if err != nil {
        panic(err)
    }
    
    // Initialize client (handles OAuth flow automatically)
    ctx := context.Background()
    result, err := c.Initialize(ctx, mcp.InitializeRequest{
        Params: struct {
            ProtocolVersion string                 `json:"protocolVersion"`
            Capabilities    mcp.ClientCapabilities `json:"capabilities"`
            ClientInfo      mcp.Implementation     `json:"clientInfo"`
        }{
            ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
            ClientInfo: mcp.Implementation{
                Name:    "my-mcp-client",
                Version: "1.0.0",
            },
        },
    })
    
    // Use client to call tools
    tools, err := c.ListTools(ctx, mcp.ListToolsRequest{})
    // ...
}
```

### 2. Python Client with API Key

```python
import requests
import os

class MCPFierClient:
    def __init__(self, base_url, api_key):
        self.base_url = base_url
        self.api_key = api_key
        self.headers = {
            'X-API-Key': api_key,
            'Content-Type': 'application/json'
        }
    
    def list_tools(self):
        response = requests.post(
            f"{self.base_url}/mcp",
            json={
                "jsonrpc": "2.0",
                "method": "tools/list",
                "id": 1
            },
            headers=self.headers
        )
        return response.json()
    
    def call_tool(self, tool_name, arguments=None):
        response = requests.post(
            f"{self.base_url}/mcp",
            json={
                "jsonrpc": "2.0",
                "method": "tools/call",
                "params": {
                    "name": tool_name,
                    "arguments": arguments or {}
                },
                "id": 1
            },
            headers=self.headers
        )
        return response.json()

# Usage
client = MCPFierClient(
    "http://localhost:8080", 
    os.getenv("MCP_API_KEY")
)

tools = client.list_tools()
result = client.call_tool("echo-test")
```

### 3. JavaScript/Node.js Client

```javascript
const axios = require('axios');

class MCPFierClient {
    constructor(baseUrl, apiKey) {
        this.baseUrl = baseUrl;
        this.client = axios.create({
            baseURL: baseUrl,
            headers: {
                'X-API-Key': apiKey,
                'Content-Type': 'application/json'
            }
        });
    }
    
    async listTools() {
        const response = await this.client.post('/mcp', {
            jsonrpc: '2.0',
            method: 'tools/list',
            id: 1
        });
        return response.data;
    }
    
    async callTool(toolName, arguments = {}) {
        const response = await this.client.post('/mcp', {
            jsonrpc: '2.0',
            method: 'tools/call',
            params: {
                name: toolName,
                arguments: arguments
            },
            id: 1
        });
        return response.data;
    }
}

// Usage
const client = new MCPFierClient(
    'http://localhost:8080',
    process.env.MCP_API_KEY
);

client.listTools().then(tools => console.log(tools));
client.callTool('weather').then(result => console.log(result));
```

## Monitoring and Observability

### Analytics Integration

MCPFier's analytics system tracks HTTP server usage:

```bash
# View server analytics
./mcpfier --analytics

# Sample output
{
  "total_requests": 1250,
  "authenticated_requests": 1180,
  "failed_auth": 70,
  "top_clients": [
    {
      "name": "production-api",
      "requests": 800,
      "success_rate": 99.5
    }
  ],
  "avg_response_time_ms": 45
}
```

### Health Check Endpoint

```bash
# Health check endpoint
curl http://localhost:8080/health

# Response
{
  "status": "healthy",
  "version": "1.0.0",
  "uptime": "24h15m30s",
  "active_connections": 5
}
```

### Prometheus Metrics

```yaml
# Enable Prometheus metrics
server:
  metrics:
    enabled: true
    endpoint: "/metrics"
    port: 9090
```

```bash
# Scrape metrics
curl http://localhost:9090/metrics
```

## Troubleshooting

### Common Issues

1. **Authentication Failed (401)**
   ```bash
   # Check API key
   curl -v -H "X-API-Key: your-key" http://localhost:8080/mcp
   
   # Verify key in configuration
   grep -A5 api_keys config.yaml
   ```

2. **OAuth Flow Fails**
   ```bash
   # Check authorization server metadata
   curl http://localhost:8080/.well-known/oauth-authorization-server
   
   # Verify client configuration
   echo $OAUTH_CLIENT_ID
   echo $OAUTH_CLIENT_SECRET
   ```

3. **CORS Issues**
   ```yaml
   # Update CORS configuration
   server:
     http:
       cors:
         allowed_origins: ["https://your-frontend.com"]
   ```

4. **Rate Limiting**
   ```yaml
   # Adjust rate limits
   server:
     http:
       rate_limit:
         requests_per_minute: 120
   ```

### Debug Mode

```bash
# Enable debug logging
export MCP_DEBUG=true
./mcpfier --server

# OAuth-specific debugging
export OAUTH_DEBUG=true
./mcpfier --server
```

### Log Analysis

```bash
# View authentication logs
tail -f /var/log/mcpfier/auth.log

# Analyze failed requests
grep "401\|403" /var/log/mcpfier/access.log
```

## Migration from STDIO to HTTP Server

### 1. Current STDIO Setup
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

### 2. Migrate to HTTP Server

**Step 1: Update Configuration**
```yaml
server:
  http:
    port: 8080
    auth:
      mode: "simple"
      simple:
        api_keys:
          - key: "claude_desktop_key_123"
            name: "claude-desktop"
            permissions: ["*"]
```

**Step 2: Start HTTP Server**
```bash
./mcpfier --server
```

**Step 3: Update Claude Desktop Config**
```json
{
  "mcpServers": {
    "mcpfier-http": {
      "command": "curl",
      "args": [
        "-H", "X-API-Key: claude_desktop_key_123",
        "-H", "Content-Type: application/json",
        "http://localhost:8080/mcp"
      ]
    }
  }
}
```

## Performance Optimization

### Connection Pooling
```yaml
server:
  http:
    connection_pool:
      max_idle_connections: 100
      max_connections_per_host: 10
      idle_timeout: "90s"
```

### Caching
```yaml
server:
  cache:
    enabled: true
    ttl: "5m"
    max_size: "100MB"
    redis_url: "redis://localhost:6379"
```

### Load Balancing

**HAProxy Configuration:**
```
backend mcpfier_servers
    balance roundrobin
    option httpchk GET /health
    server mcpfier1 localhost:8080 check
    server mcpfier2 localhost:8081 check
    server mcpfier3 localhost:8082 check
```

## Security Hardening

### Production Security Checklist

- [ ] **TLS/HTTPS**: Use TLS 1.3+ for all connections
- [ ] **API Key Rotation**: Implement regular key rotation
- [ ] **Rate Limiting**: Set appropriate rate limits
- [ ] **CORS Configuration**: Restrict to known origins
- [ ] **Network Security**: Use firewalls and VPNs
- [ ] **Audit Logging**: Enable comprehensive logging
- [ ] **Token Expiration**: Use short-lived access tokens
- [ ] **Security Headers**: Add security headers to responses
- [ ] **Input Validation**: Validate all input parameters
- [ ] **Error Handling**: Don't leak sensitive information in errors

### Security Headers
```yaml
server:
  http:
    security_headers:
      enabled: true
      headers:
        "Strict-Transport-Security": "max-age=31536000"
        "X-Content-Type-Options": "nosniff"
        "X-Frame-Options": "DENY"
        "X-XSS-Protection": "1; mode=block"
```

## Related Documentation

- **[MCP 2025-06-18 Specification](https://modelcontextprotocol.io/specification/2025-06-18)** - Official MCP protocol specification
- **[mark3labs/mcp-go](https://github.com/mark3labs/mcp-go)** - Go library for MCP implementation
- **[OAuth 2.1](https://datatracker.ietf.org/doc/draft-ietf-oauth-v2-1/)** - OAuth 2.1 specification
- **[RFC9728](https://datatracker.ietf.org/doc/html/rfc9728)** - OAuth 2.0 Protected Resource Metadata
- **[RFC8414](https://datatracker.ietf.org/doc/html/rfc8414)** - OAuth 2.0 Authorization Server Metadata

## Contributing

When contributing to HTTP server features:

1. **Follow MCP Specification**: Ensure compliance with MCP 2025-06-18
2. **Security First**: All changes must maintain security standards
3. **Backward Compatibility**: Don't break existing STDIO mode
4. **Documentation**: Update this document for new features
5. **Testing**: Include comprehensive auth flow tests

## Support

For issues and questions:

1. **GitHub Issues**: [mcpfier/issues](https://github.com/gleicon/mcpfier/issues)
2. **Documentation**: Check README.md and SERVER.md
3. **Community**: MCP community forums and discussions