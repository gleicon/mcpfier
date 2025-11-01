# MCPFier Roadmap

## Current Version: v1.0

MCPFier is a simple, modular MCP API gateway. This roadmap outlines implemented features and future plans.

## Implemented Features

### Core Functionality
- Three execution modes: Local, Container, Webhook/API
- Dual transport: STDIO (Claude Desktop) and HTTP/SSE (Enterprise)
- YAML-based configuration with auto-discovery
- Environment variable expansion for secrets (${VAR_NAME})
- Comprehensive error handling and timeout support

### Security
- API key authentication with per-tool permissions
- Rate limiting (token bucket algorithm)
- Container isolation for untrusted code
- Request logging and audit trail

### Operations
- Embedded SQLite analytics
- Real-time web dashboard
- Health check endpoints
- Configurable data retention

### API Gateway Features
- HTTP client for external APIs
- Retry logic with exponential backoff
- Multiple auth methods (Bearer, API Key, Basic)
- Request/response logging

## Planned Features

### v1.1 (Next Release)

#### MCP Resources Capability
- Expose webhook results as MCP resources
- Resource caching layer
- URI-based resource access

#### Request Transformation
- jq-style transformations for request/response
- Template engine for dynamic values
- Path parameter substitution

#### Enhanced Observability
- OpenTelemetry integration
- Prometheus metrics endpoint
- Structured JSON logging

### v1.2 (Medium Term)

#### Advanced Security
- JWT token validation
- OIDC/OAuth 2.1 support
- Role-based access control (RBAC)
- Certificate-based authentication

#### Gateway Features
- Circuit breaker pattern for webhooks
- Request/response caching
- Load balancing for multiple backends
- API composition (multi-tool workflows)

#### Developer Experience
- Hot configuration reload
- Configuration validation CLI
- Built-in testing framework
- Plugin system for custom executors

### v2.0 (Long Term)

See docs/AUTHENTICATION_PLAN.md for detailed vision of MCPFier as a full MCP Gateway.

#### Enterprise Features
- Multi-tenancy support
- Distributed tracing
- Advanced analytics and ML insights
- Service mesh integration

#### Protocol Extensions
- MCP Prompts capability
- MCP Sampling capability
- Streaming responses
- Bidirectional communication

#### Ecosystem
- Kubernetes operator
- Terraform provider
- Pre-built integration catalog
- Managed service offering

## Contributing

We welcome contributions aligned with our core principles:
- Keep it simple
- Stay modular
- Maintain small codebase
- Focus on API gateway use case

See individual feature proposals in docs/ directory.

## Version History

### v1.0 (Current)
- Initial release with core functionality
- Three execution modes
- API key authentication
- Rate limiting
- Embedded analytics
- Environment variable expansion

For detailed technical architecture, see ARCHITECTURE.md.
