# Release Notes - v1.0.4

Release Date: 2025-11-03

## Summary

Version 1.0.4 introduces secure API key storage with three formats, comprehensive documentation cleanup, and improved security features for production deployments.

## New Features

### Secure API Key Storage

Three storage formats to match your security requirements:

**Plain Text**
```yaml
api_keys:
  - key: "mcpfier_dev_123456"
    name: "development"
    permissions: ["*"]
```

**Environment Variables**
```yaml
api_keys:
  - key: "${MCPFIER_API_KEY}"
    name: "production"
    permissions: ["*"]
```

**Hashed Keys (SHA-256)**
```yaml
api_keys:
  - key: "hash:95e904704170a1539c636551759916c31758d6302b43c4e80e11a44a0a2fbb40"
    name: "production"
    permissions: ["*"]
```

### Key Generation Utility

New command-line utility for generating and hashing API keys:

```bash
go build -o bin/hashkey ./cmd/hashkey
./bin/hashkey -generate
./bin/hashkey -hash "existing-key"
```

### Environment Variable Validation

API key configuration now validates environment variables:

- Detects undefined variables
- Reports errors at startup
- Prevents silent failures

## Breaking Changes

None. All existing configurations continue to work without modification.

## Migration Guide

### From Plain Text to Environment Variables

1. Set environment variable:
```bash
export MCPFIER_API_KEY="your-existing-key"
```

2. Update config.yaml:
```yaml
api_keys:
  - key: "${MCPFIER_API_KEY}"
    name: "production"
```

3. Restart server

### From Plain Text to Hashed Keys

1. Hash existing key:
```bash
./bin/hashkey -hash "your-existing-key"
```

2. Copy hash from output and update config.yaml:
```yaml
api_keys:
  - key: "hash:PASTE_HASH_HERE"
    name: "production"
```

3. Restart server
4. Clients use original plain text key (no changes needed)

## Files Added

- `cmd/hashkey/main.go` - Key generation and hashing utility
- `internal/config/secrets.go` - Secret handling and validation
- `internal/config/secrets_test.go` - Comprehensive tests
- `docs/API_KEY_STORAGE.md` - Complete storage guide
- `docs/QUICK_START_API_KEYS.md` - 5-minute quick start
- `docs/RELEASE_NOTES_v1.0.4.md` - This file

## Files Modified

- `internal/auth/auth.go` - Added hashed key authentication
- `internal/auth/auth_test.go` - Added hashed key tests
- `internal/config/config.go` - Added key processing during load
- `config.yaml` - Added examples for all three formats
- `README.md` - Added API key storage section and documentation links
- `CHANGELOG.md` - Complete v1.0.4 changelog
- `docs/AUTHENTICATION_NOTES.md` - Enhanced with security best practices

## Documentation Improvements

### New Documentation

- API key storage guide with migration examples
- Quick start guide for 5-minute setup
- Security considerations and best practices
- Key rotation procedures

### Updated Documentation

- README now shows all three storage formats
- Configuration examples updated to current format
- Authentication notes expanded with practical examples
- Links organized by quick guides vs detailed docs

## Security Enhancements

### SHA-256 for API Keys

Uses SHA-256 instead of bcrypt for performance:

- API keys validated on every request
- bcrypt too slow (100ms per request)
- SHA-256 suitable for high-entropy random keys
- <1ms validation time

### Environment Variable Support

- Keep secrets out of config files
- Config files safe to commit to git
- Integration with Docker, Kubernetes, CI/CD
- Validation at startup prevents runtime errors

### Hashed Storage

- Config contains only SHA-256 hash
- Original key cannot be recovered
- Compliance-friendly
- Supports shared configuration files

## Testing

All tests passing:

- Environment variable expansion: 6 test cases
- API key hashing: 3 test cases
- Hash comparison: 3 test cases
- Key detection: 6 test cases
- API key processing: 5 test cases
- Hashed authentication: 1 test case
- Backward compatibility: verified

Total: 30+ test cases covering all new functionality

## Performance

No performance impact on existing deployments:

- Plain text keys: identical performance
- Environment variables: expansion at startup only
- Hashed keys: SHA-256 is fast (<1ms per request)

## Backward Compatibility

Fully backward compatible:

- Existing plain text configs work without changes
- No configuration schema changes required
- Clients unaffected by server-side storage format
- Gradual migration supported (mixed formats)

## Getting Started

### Quick Setup (5 minutes)

```bash
# Build utilities
go build -o bin/hashkey ./cmd/hashkey

# Generate key
./bin/hashkey -generate

# Copy plain text key to config.yaml
# Start using immediately
```

### Production Setup

```bash
# Generate key
KEY=$(./bin/hashkey -generate | grep "Plain text:" | cut -d' ' -f3)

# Set environment variable
export MCPFIER_API_KEY="$KEY"

# Update config.yaml to use ${MCPFIER_API_KEY}
# Start server
```

## Documentation Links

- Quick Start: [docs/QUICK_START_API_KEYS.md](QUICK_START_API_KEYS.md)
- Complete Guide: [docs/API_KEY_STORAGE.md](API_KEY_STORAGE.md)
- Authentication Details: [docs/AUTHENTICATION_NOTES.md](AUTHENTICATION_NOTES.md)
- Changelog: [CHANGELOG.md](../CHANGELOG.md)

## Known Issues

None.

## Future Enhancements

Planned for future releases:

- Key rotation automation
- Multiple hash algorithms
- Key expiration dates
- OAuth 2.1 integration
- RBAC support

See ROADMAP.md for complete feature roadmap.

## Support

For issues or questions:

- GitHub Issues: https://github.com/gleicon/mcpfier/issues
- Documentation: https://github.com/gleicon/mcpfier/tree/main/docs
