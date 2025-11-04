# API Key Storage Guide

Version: v1.0.4

## Overview

MCPFier supports three API key storage formats. Choose the format that matches your security requirements and deployment environment.

## Storage Formats

### 1. Plain Text (Development)

Store API keys directly in config.yaml.

**When to use:**
- Local development
- Testing
- Single-user setups

**Example:**
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

**Security note:** Keys are visible in config file. Do not commit to version control.

### 2. Environment Variables (Production)

Reference environment variables in config.yaml.

**When to use:**
- Production deployments
- CI/CD pipelines
- Docker/Kubernetes environments
- When config files are committed to git

**Example:**
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

**Set environment variable:**
```bash
export MCPFIER_API_KEY="cuhw4RIKLaBkACH4hBlkfPbRHPaqxxc-AdUFV2qVQA"
```

**Security note:** Config file is safe to commit. Keys managed separately.

### 3. Hashed Keys (Maximum Security)

Store SHA-256 hash of API key in config.yaml.

**When to use:**
- Compliance requirements
- Shared configuration files
- When config backups must not expose keys
- Multi-tenant environments

**Example:**
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

**Security note:** Only hash stored. Original key cannot be recovered from config.

## Getting Started

### Step 1: Build the Key Utility

```bash
cd /path/to/mcpfier
go build -o bin/hashkey ./cmd/hashkey
```

### Step 2: Generate Your First Key

```bash
./bin/hashkey -generate
```

Output:
```
Generated API Key:
==================
Plain text: cuhw4RIKLaBkACH4hBlkfPbRHPaqxxc-AdUFV2qVQA

Hashed version (SHA-256):
hash:95e904704170a1539c636551759916c31758d6302b43c4e80e11a44a0a2fbb40

Configuration examples:
=======================

1. Plain text in config.yaml:
   api_keys:
     - key: "cuhw4RIKLaBkACH4hBlkfPbRHPaqxxc-AdUFV2qVQA"
       name: "client-name"
       permissions: ["*"]

2. Hashed in config.yaml:
   api_keys:
     - key: "hash:95e904704170a1539c636551759916c31758d6302b43c4e80e11a44a0a2fbb40"
       name: "client-name"
       permissions: ["*"]

3. Environment variable in config.yaml:
   api_keys:
     - key: "${MCPFIER_API_KEY}"
       name: "client-name"
       permissions: ["*"]

   Then set: export MCPFIER_API_KEY="cuhw4RIKLaBkACH4hBlkfPbRHPaqxxc-AdUFV2qVQA"
```

### Step 3: Choose Your Format

Pick one of the three formats shown above and update your config.yaml.

### Step 4: Share Key with Client

Send the plain text key to your client:
```
cuhw4RIKLaBkACH4hBlkfPbRHPaqxxc-AdUFV2qVQA
```

Clients use this key in their requests regardless of storage format:
```bash
curl -X POST http://localhost:8080/mcp \
  -H "X-API-Key: cuhw4RIKLaBkACH4hBlkfPbRHPaqxxc-AdUFV2qVQA" \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"tools/list","id":1}'
```

## Migration Guide

### From Plain Text to Environment Variables

1. Identify existing keys in config.yaml:
```yaml
api_keys:
  - key: "mcpfier_dev_123456"
    name: "development"
```

2. Set environment variable:
```bash
export MCPFIER_DEV_KEY="mcpfier_dev_123456"
```

3. Update config.yaml:
```yaml
api_keys:
  - key: "${MCPFIER_DEV_KEY}"
    name: "development"
```

4. Restart MCPFier server
5. Test authentication still works
6. Remove plain text key from config.yaml history (git filter-branch if needed)

### From Plain Text to Hashed Keys

1. Identify existing keys in config.yaml:
```yaml
api_keys:
  - key: "mcpfier_dev_123456"
    name: "development"
```

2. Hash the existing key:
```bash
./bin/hashkey -hash "mcpfier_dev_123456"
```

Output:
```
API Key Hash:
=============
Input:  mcpfier_dev_123456
SHA-256: 80744d96caaa2acddd1cfc0e242757867353b659e51f5b3bb44f2f2159a345d5

Configuration:
==============
api_keys:
  - key: "hash:80744d96caaa2acddd1cfc0e242757867353b659e51f5b3bb44f2f2159a345d5"
    name: "client-name"
    permissions: ["*"]
```

3. Update config.yaml with hashed value:
```yaml
api_keys:
  - key: "hash:80744d96caaa2acddd1cfc0e242757867353b659e51f5b3bb44f2f2159a345d5"
    name: "development"
```

4. Restart MCPFier server
5. Test authentication with original key (mcpfier_dev_123456)
6. Clients do not need to change anything

### From Environment Variables to Hashed Keys

1. Retrieve current key from environment:
```bash
echo $MCPFIER_DEV_KEY
```

2. Hash the key:
```bash
./bin/hashkey -hash "$(echo $MCPFIER_DEV_KEY)"
```

3. Update config.yaml with hash
4. Can now remove environment variable if desired

## Key Rotation

### Generate New Key

```bash
./bin/hashkey -generate
```

### Update Configuration

Add new key while keeping old key active:
```yaml
api_keys:
  - key: "hash:old_hash_here"
    name: "client1"
    permissions: ["*"]
  
  - key: "hash:new_hash_here"
    name: "client1"
    permissions: ["*"]
```

### Notify Clients

Send new plain text key to clients. Give deadline to switch.

### Remove Old Key

After migration period, remove old key from config.yaml:
```yaml
api_keys:
  - key: "hash:new_hash_here"
    name: "client1"
    permissions: ["*"]
```

## Validation

MCPFier validates keys during startup:

**Success:**
```
INFO: Loaded 3 API keys
INFO: HTTP server listening on :8080
```

**Failure (undefined environment variable):**
```
ERROR: API key for 'production' contains unexpanded environment variable
```

**Failure (invalid hash):**
```
ERROR: invalid hashed API key for 'client1': must be valid hex
```

## Hash an Existing Key

If you have a key in use and want to hash it:

```bash
./bin/hashkey -hash "your-existing-key-here"
```

The output shows the hash to put in config.yaml. Clients continue using the original key.

## Common Configurations

### Development Only

```yaml
server:
  http:
    auth:
      enabled: true
      mode: "simple"
      simple:
        api_keys:
          - key: "dev_local_key"
            name: "local"
            permissions: ["*"]
```

### Production with Environment Variables

```yaml
server:
  http:
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
            permissions: ["search-products", "get-weather"]
```

Shell:
```bash
export MCPFIER_CLIENT1_KEY="$(./bin/hashkey -generate | grep 'Plain text:' | cut -d' ' -f3)"
export MCPFIER_CLIENT2_KEY="$(./bin/hashkey -generate | grep 'Plain text:' | cut -d' ' -f3)"
```

### Production with Hashed Keys

```yaml
server:
  http:
    auth:
      enabled: true
      mode: "simple"
      simple:
        api_keys:
          - key: "hash:95e904704170a1539c636551759916c31758d6302b43c4e80e11a44a0a2fbb40"
            name: "client1"
            permissions: ["*"]
          
          - key: "hash:a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
            name: "client2"
            permissions: ["search-products", "get-weather"]
```

### Mixed Environment

```yaml
server:
  http:
    auth:
      enabled: true
      mode: "simple"
      simple:
        api_keys:
          # Local development
          - key: "dev_local_key"
            name: "local"
            permissions: ["*"]
          
          # Production client (environment variable)
          - key: "${PROD_CLIENT_KEY}"
            name: "prod-client"
            permissions: ["*"]
          
          # Partner integration (hashed)
          - key: "hash:95e904704170a1539c636551759916c31758d6302b43c4e80e11a44a0a2fbb40"
            name: "partner"
            permissions: ["search-products"]
```

## Security Notes

### Why SHA-256 Instead of bcrypt?

API keys are validated on every request. Performance comparison:

- bcrypt: ~100ms per validation
- SHA-256: <1ms per validation

With 100 concurrent requests:
- bcrypt: ~2GB RAM, high CPU usage
- SHA-256: minimal overhead

API keys should be high-entropy random strings (32+ bytes). SHA-256 is sufficient for properly generated keys. Use the hashkey utility to generate strong keys.

### Key Generation Best Practices

1. Use the hashkey utility (cryptographically random)
2. Minimum 32 bytes of entropy
3. Never use predictable values
4. Rotate keys periodically
5. Different key per client
6. Revoke keys when no longer needed

### Storage Best Practices

1. Development: Plain text is acceptable
2. Production: Use environment variables or hashed keys
3. Compliance: Use hashed keys
4. Never commit plain text production keys to git
5. Use .gitignore for .env files
6. Restrict config file permissions (chmod 600)

## Troubleshooting

### "contains unexpanded environment variable"

Environment variable is not set or is empty.

Fix:
```bash
export MCPFIER_API_KEY="your-key-here"
```

Verify:
```bash
echo $MCPFIER_API_KEY
```

### "invalid hashed API key: must be valid hex"

Hash is not a valid SHA-256 hash (must be 64 hex characters).

Fix: Regenerate hash with hashkey utility:
```bash
./bin/hashkey -hash "your-key"
```

### Authentication fails after migration

Verify client is using original plain text key, not the hash.

Wrong:
```bash
curl -H "X-API-Key: hash:95e904..."
```

Correct:
```bash
curl -H "X-API-Key: cuhw4RIKLaBkACH4hBlkfPbRHPaqxxc-AdUFV2qVQA"
```

### Multiple keys with same name

MCPFier allows this for key rotation. First matching key authenticates the request.

## Quick Reference

| Task | Command |
|------|---------|
| Build utility | `go build -o bin/hashkey ./cmd/hashkey` |
| Generate new key | `./bin/hashkey -generate` |
| Hash existing key | `./bin/hashkey -hash "key"` |
| Generate longer key | `./bin/hashkey -generate -length 48` |
| Set environment variable | `export MCPFIER_API_KEY="key"` |
| Test authentication | `curl -H "X-API-Key: key" http://localhost:8080/mcp` |

## Format Comparison

| Format | Security | Ease of Use | Git Safe | Audit Trail |
|--------|----------|-------------|----------|-------------|
| Plain Text | Low | High | No | Yes |
| Environment Variables | Medium | Medium | Yes | No |
| Hashed | High | Medium | Yes | Yes |

Choose based on your security requirements and operational constraints.
