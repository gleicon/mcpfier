# Changelog

All notable changes to MCPFier will be documented in this file.

## [1.0.4] - 2025-11-03

### Added
- Three API key storage formats: plain text, environment variables, hashed (SHA-256)
- Environment variable expansion in API keys using ${VAR_NAME} syntax
- Hashed API key support with hash: prefix for secure storage
- Key generation and hashing utility: cmd/hashkey/main.go
- Comprehensive API key authentication tests
- AUTHENTICATION_NOTES.md documenting current auth implementation and security best practices
- STRATEGY.md consolidating all strategic planning
- internal/config/secrets.go for secure key handling

### Changed
- Verified and documented that API key authentication is fully working
- Cleaned up docs folder (removed implementation plans for completed features)
- Updated version to v1.0.4 across documentation
- Enhanced internal/auth/auth.go to support hashed key validation
- Updated config processing to expand environment variables and handle hashed keys

### Fixed
- Corrected documentation that incorrectly stated auth was not implemented
- Authentication has been working since v0.1.2, now properly documented
- Environment variable expansion now properly validates undefined variables

### Security
- API keys can now be stored as SHA-256 hashes in config.yaml
- Environment variable support prevents committing secrets to git
- Backward compatible: existing plain text configs continue to work
- SHA-256 chosen over bcrypt for performance (API keys validated on every request)
- Cryptographically random key generation with base64url encoding

### Documentation
- Consolidated 9 strategic analysis documents into single STRATEGY.md
- Updated AUTHENTICATION_NOTES.md with complete implementation details, security considerations, and migration guide
- Added examples for all three key storage formats
- Documented key generation utility usage
- Removed obsolete TLS_IMPLEMENTATION_PLAN.md (feature complete)
- Removed obsolete test_logging_analytics.md

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
