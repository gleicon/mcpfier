# TLS/HTTPS Implementation Plan for MCPFier

## Current State

### What MCPFier Currently Does
- HTTP server runs on `http.ListenAndServe(addr, handler)` (line 218 of http_server.go)
- No TLS/HTTPS support
- Plain text HTTP only
- SECURITY.md mentions "Configure HTTPS for production" but no implementation

### What mcp-go Provides
- `StreamableHTTPServer` uses standard Go http.Server
- No built-in TLS configuration
- Expects you to handle TLS at your level (which is correct design)
- Works as http.Handler, so compatible with any TLS setup

### Security Implications
Current setup transmits in plaintext:
- API keys in headers
- Tool execution results
- Authentication tokens for upstream APIs
- All MCP protocol messages

For production deployment, this is a critical gap.

## Implementation Options

### Option 1: Reverse Proxy (Recommended for Production)
Put mcpfier behind nginx/Caddy/Traefik.

**Pros**:
- Industry standard pattern
- Separation of concerns (mcpfier focuses on MCP, proxy handles TLS)
- Easier certificate management
- Load balancing support
- Battle-tested

**Cons**:
- Extra infrastructure component
- More complex deployment
- Not plug-and-play for simple use cases

**Current recommendation in SECURITY.md**: "Deploy behind reverse proxy with TLS"

### Option 2: Manual TLS (Bring Your Own Certificates)
Add config for cert/key files, use `http.ListenAndServeTLS()`.

**Pros**:
- Simple implementation
- Full control over certificates
- Works with any certificate source
- No external dependencies

**Cons**:
- User must manage certificates manually
- No automatic renewal
- Certificate expiry is user's problem

### Option 3: ACME/Let's Encrypt with autocert
Automatic certificate provisioning and renewal.

**Pros**:
- Zero-config TLS for public servers
- Automatic renewal
- Best user experience for simple deployments
- Production-grade security with no manual steps

**Cons**:
- Requires public DNS and ports 80/443
- Rate limits (50 certs per domain per week)
- Not suitable for internal/private servers
- Cache directory management
- More complex implementation

### Option 4: Hybrid Approach (RECOMMENDED)
Support all three: manual TLS, autocert, and reverse proxy.

**Pros**:
- Flexibility for all deployment scenarios
- Simple use cases get autocert
- Enterprise gets manual control
- Production gets reverse proxy
- Maintains simplicity for basic users

**Cons**:
- More code to maintain
- More configuration options
- More documentation needed

## Recommended Implementation Plan

### Phase 1: Manual TLS Support (v1.1)

Add configuration for bringing your own certificates:

```yaml
server:
  http:
    port: 8080
    host: "localhost"
    tls:
      enabled: false
      cert_file: "/path/to/cert.pem"
      key_file: "/path/to/key.pem"
```

**Code changes**:
1. Add TLSConfig to config.go
2. Modify http_server.go Start() method
3. Use http.ListenAndServeTLS() when enabled

**Effort**: 2-3 hours
**Risk**: Low
**Files**: 2 modified

### Phase 2: ACME/autocert Support (v1.2)

Add automatic Let's Encrypt certificates:

```yaml
server:
  http:
    port: 443  # HTTPS default
    host: "mcpfier.example.com"
    tls:
      enabled: true
      mode: "autocert"
      autocert:
        domains: ["mcpfier.example.com"]
        cache_dir: "/var/cache/mcpfier/certs"
        email: "admin@example.com"  # Optional, for renewal notices
```

**Requirements**:
- Public DNS pointing to server
- Ports 80 and 443 accessible
- Valid domain name
- Persistent cache directory

**Code changes**:
1. Add autocert config structs
2. Create autocert manager setup
3. Handle HTTP-01 challenge on port 80
4. Serve HTTPS on port 443
5. Manage certificate cache

**Effort**: 4-6 hours
**Risk**: Medium (rate limits, DNS requirements)
**Files**: 3 modified, 1 new (internal/tls/)

### Phase 3: Documentation and Testing (v1.3)

1. Update SECURITY.md with TLS guidance
2. Add TLS examples to SERVER.md
3. Create deployment guides for each scenario
4. Add TLS testing scripts
5. Document certificate renewal process

**Effort**: 3-4 hours
**Files**: Multiple docs

## Detailed Implementation

### Config Schema

```go
// internal/config/config.go

type TLSConfig struct {
    Enabled  bool          `yaml:"enabled"`
    Mode     string        `yaml:"mode"`      // "manual" or "autocert"
    CertFile string        `yaml:"cert_file"` // For manual mode
    KeyFile  string        `yaml:"key_file"`  // For manual mode
    Autocert *AutocertConfig `yaml:"autocert"` // For autocert mode
}

type AutocertConfig struct {
    Domains  []string `yaml:"domains"`
    CacheDir string   `yaml:"cache_dir"`
    Email    string   `yaml:"email"`
}

type HTTPConfig struct {
    Port      int             `yaml:"port"`
    Host      string          `yaml:"host"`
    TLS       TLSConfig       `yaml:"tls"`
    Auth      AuthConfig      `yaml:"auth"`
    CORS      CORSConfig      `yaml:"cors"`
    RateLimit RateLimitConfig `yaml:"rate_limit"`
}
```

### Server Implementation

```go
// internal/server/http_server.go

func (s *HTTPServer) Start() error {
    addr := fmt.Sprintf("%s:%d", s.config.Server.HTTP.Host, s.config.Server.HTTP.Port)
    
    // Build middleware stack
    handler := s.buildHandler()
    
    // Start server based on TLS configuration
    if s.config.Server.HTTP.TLS.Enabled {
        return s.startTLS(addr, handler)
    }
    
    log.Printf("MCPFier HTTP server starting on %s (no TLS)", addr)
    return http.ListenAndServe(addr, handler)
}

func (s *HTTPServer) startTLS(addr string, handler http.Handler) error {
    tlsCfg := s.config.Server.HTTP.TLS
    
    switch tlsCfg.Mode {
    case "manual":
        return s.startManualTLS(addr, handler, tlsCfg)
    case "autocert":
        return s.startAutocert(handler, tlsCfg.Autocert)
    default:
        return fmt.Errorf("unsupported TLS mode: %s", tlsCfg.Mode)
    }
}

func (s *HTTPServer) startManualTLS(addr string, handler http.Handler, cfg TLSConfig) error {
    if cfg.CertFile == "" || cfg.KeyFile == "" {
        return fmt.Errorf("cert_file and key_file required for manual TLS")
    }
    
    log.Printf("MCPFier HTTPS server starting on %s (manual TLS)", addr)
    return http.ListenAndServeTLS(addr, cfg.CertFile, cfg.KeyFile, handler)
}

func (s *HTTPServer) startAutocert(handler http.Handler, cfg *AutocertConfig) error {
    if cfg == nil || len(cfg.Domains) == 0 {
        return fmt.Errorf("domains required for autocert mode")
    }
    
    // Expand cache directory path
    cacheDir := cfg.CacheDir
    if cacheDir == "" {
        cacheDir = "~/.mcpfier/certs"
    }
    cacheDir = expandPath(cacheDir)
    
    // Ensure cache directory exists
    if err := os.MkdirAll(cacheDir, 0700); err != nil {
        return fmt.Errorf("failed to create cert cache dir: %w", err)
    }
    
    // Create autocert manager
    m := &autocert.Manager{
        Prompt:      autocert.AcceptTOS,
        Cache:       autocert.DirCache(cacheDir),
        HostPolicy:  autocert.HostWhitelist(cfg.Domains...),
        Email:       cfg.Email,
    }
    
    // Start HTTP-01 challenge handler on port 80
    go func() {
        log.Printf("Starting HTTP-01 challenge handler on :80")
        http.ListenAndServe(":80", m.HTTPHandler(nil))
    }()
    
    // Configure TLS server
    server := &http.Server{
        Addr:      ":443",
        Handler:   handler,
        TLSConfig: m.TLSConfig(),
    }
    
    log.Printf("MCPFier HTTPS server starting on :443 (autocert for %v)", cfg.Domains)
    return server.ListenAndServeTLS("", "")
}
```

### Example Configurations

#### Development (No TLS)
```yaml
server:
  http:
    port: 8080
    host: "localhost"
    # No TLS config - plain HTTP
```

#### Staging (Manual TLS with Self-Signed Cert)
```yaml
server:
  http:
    port: 8443
    host: "0.0.0.0"
    tls:
      enabled: true
      mode: "manual"
      cert_file: "/etc/mcpfier/certs/cert.pem"
      key_file: "/etc/mcpfier/certs/key.pem"
```

Generate self-signed cert:
```bash
openssl req -x509 -newkey rsa:4096 -nodes \
  -keyout key.pem -out cert.pem -days 365 \
  -subj "/CN=mcpfier.local"
```

#### Production (Let's Encrypt)
```yaml
server:
  http:
    # autocert overrides port to 443
    tls:
      enabled: true
      mode: "autocert"
      autocert:
        domains: ["mcp.example.com"]
        cache_dir: "/var/cache/mcpfier/certs"
        email: "admin@example.com"
```

Requirements:
- DNS A record: mcp.example.com → server IP
- Firewall: ports 80 and 443 open
- No reverse proxy (conflicts with autocert)

#### Enterprise (Reverse Proxy)
```yaml
server:
  http:
    port: 8080
    host: "127.0.0.1"  # Bind to localhost only
    # No TLS - handled by nginx/Caddy
```

Nginx config:
```nginx
server {
    listen 443 ssl http2;
    server_name mcp.example.com;
    
    ssl_certificate /etc/letsencrypt/live/mcp.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/mcp.example.com/privkey.pem;
    
    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## Testing Plan

### Manual TLS Testing
```bash
# Generate test cert
openssl req -x509 -newkey rsa:2048 -nodes \
  -keyout /tmp/test-key.pem -out /tmp/test-cert.pem -days 1 \
  -subj "/CN=localhost"

# Update config.yaml
# Start server
./mcpfier --server

# Test with curl (ignore self-signed)
curl -k https://localhost:8443/health

# Test MCP client connection
```

### Autocert Testing
```bash
# Prerequisites:
# 1. Domain pointing to server
# 2. Ports 80/443 open
# 3. Update config with real domain

# Start server
sudo ./mcpfier --server  # Needs root for ports 80/443

# Check cert provisioning
curl https://yourdomain.com/health

# Verify certificate
openssl s_client -connect yourdomain.com:443 -servername yourdomain.com
```

### Certificate Renewal Testing
Autocert renews automatically when cert is <30 days from expiry.

Force renewal test:
```bash
# Delete cached cert
rm -rf /var/cache/mcpfier/certs/*

# Restart server (will request new cert)
sudo ./mcpfier --server
```

## Security Considerations

### Certificate Storage
- Cache directory should be 0700 permissions
- Certificate files should be 0600 permissions
- Never commit certificates to version control
- Rotate certificates regularly if manual mode

### Rate Limits
Let's Encrypt limits:
- 50 certificates per registered domain per week
- 5 duplicate certificates per week
- 300 new orders per account per 3 hours

Mitigation:
- Use persistent cache directory
- Don't delete cache unnecessarily
- Use staging environment first
- Monitor certificate expiry

### Production Recommendations

1. Use reverse proxy for large deployments
2. Use autocert for simple single-server deployments
3. Use manual TLS for corporate environments with existing PKI
4. Always use TLS in production
5. Monitor certificate expiry dates
6. Set up renewal alerts
7. Keep cache directory backed up

## Migration Path

### Current Users (HTTP Only)
No breaking changes. TLS is opt-in.

### Enabling TLS
1. Choose mode based on deployment
2. Update config.yaml
3. Test in staging first
4. Deploy to production
5. Update client configurations to use https://
6. Update firewall rules if using autocert

## Documentation Updates

### README.md
Add TLS quick start section

### SECURITY.md  
Expand TLS recommendations with examples

### SERVER.md
Add complete TLS setup guide

### New: TLS_GUIDE.md
Detailed guide for each mode with troubleshooting

## Dependencies

```go
import "golang.org/x/crypto/acme/autocert"
```

Already in standard toolchain, no new external dependencies.

## Timeline Estimate

| Phase | Component | Effort | Risk |
|-------|-----------|--------|------|
| 1 | Manual TLS config structs | 1h | Low |
| 1 | Manual TLS implementation | 1h | Low |
| 1 | Manual TLS testing | 1h | Low |
| 2 | Autocert config structs | 1h | Low |
| 2 | Autocert implementation | 2h | Medium |
| 2 | Autocert testing | 2h | Medium |
| 3 | Documentation | 3h | Low |
| 3 | Examples and guides | 2h | Low |

**Total: 13-15 hours**

## Recommendation

Implement in order:
1. **Phase 1 first** (manual TLS) - low risk, immediate value for corporate users
2. **Phase 2 second** (autocert) - higher value for simple deployments
3. **Phase 3 throughout** - document as you build

Start with Phase 1 since it's simple and unblocks production use cases immediately.

Phase 2 (autocert) is nice-to-have but not critical since reverse proxy is industry standard for production.

## Future Enhancements

- Certificate monitoring and alerts
- Automatic cert rotation for manual mode
- Integration with cloud provider certificate managers (AWS ACM, etc.)
- mTLS (mutual TLS) for client certificate authentication
- Certificate pinning options
