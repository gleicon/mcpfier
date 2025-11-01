# Changelog

All notable changes to MCPFier will be documented in this file.

## [0.1.3] - 2025-10-31

### Added
- TLS/HTTPS support with custom certificates (manual mode)
- Environment variable expansion in webhook authentication (${VAR_NAME} syntax)
- Rate limiting middleware with token bucket algorithm
- Support for per-client rate limits (by API key or IP address)
- Configurable requests per minute and burst size
- ROADMAP.md for feature planning and versioning
- TLS implementation plan in docs/TLS_IMPLEMENTATION_PLAN.md

### Changed
- Updated mcp-go library from v0.37.0 to v0.42.0
- Cleaned up config.yaml by removing unimplemented OAuth 2.1 placeholders
- Updated ARCHITECTURE.md with new components and data flow
- Enhanced README with rate limiting and environment variable documentation
- Moved AUTHENTICATION_PLAN.md to docs/ directory for future reference
- Moved test_logging_analytics.md to docs/ directory

### Fixed
- Config schema now correctly indicates only "simple" auth mode is implemented
- Improved documentation clarity around authentication vs authorization

### Security
- TLS/HTTPS support encrypts all traffic including API keys and tool results
- Environment variables now properly expanded in webhook auth (Bearer, API Key, Basic)
- Rate limiting protects against abuse and DoS attacks
- No secrets hardcoded in configuration files

## [0.1.2] - 2025-08-10

### Added
- Three execution modes: Local, Container, Webhook/API
- Dual transport: STDIO and HTTP/SSE
- API key authentication with per-tool permissions
- Embedded SQLite analytics
- Real-time web dashboard
- CORS support
- Health check endpoints
- Comprehensive logging and error handling
- Docker support
- Claude Desktop integration

### Documentation
- README.md with quick start guide
- ARCHITECTURE.md with technical details
- SECURITY.md with best practices
- SERVER.md for HTTP mode setup
- DOCKER.md for containerized deployment
- SETUP.md for detailed installation
