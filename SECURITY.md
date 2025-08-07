# MCPFier Security Guide

This document outlines the security considerations, best practices, and configuration guidelines for deploying MCPFier in production environments.

## Security Overview

MCPFier operates in two distinct execution modes, each with different security characteristics:

1. **Local Execution**: Commands run directly on the host system
2. **Container Execution**: Commands run within isolated Docker containers

## Threat Model

### Potential Threats

1. **Code Injection**: Malicious command arguments or environment variables
2. **Privilege Escalation**: Commands attempting to gain elevated privileges
3. **Resource Exhaustion**: Commands consuming excessive CPU, memory, or disk
4. **Data Exfiltration**: Unauthorized access to sensitive files or network resources
5. **Lateral Movement**: Using MCPFier as a pivot point for further attacks

### Attack Vectors

- Malicious MCP clients sending crafted requests
- Compromised command configurations
- Container escape attempts
- Resource exhaustion attacks
- Network-based attacks on exposed services

## Security Architecture

### Defense in Depth

MCPFier implements multiple layers of security:

```
┌─────────────────────────────────────────┐
│ Application Layer Security              │
├─────────────────────────────────────────┤
│ Container/Process Isolation             │
├─────────────────────────────────────────┤
│ Operating System Security               │
├─────────────────────────────────────────┤
│ Network Security                        │
└─────────────────────────────────────────┘
```

### Security Boundaries

1. **MCP Protocol Boundary**: Input validation and sanitization
2. **Configuration Boundary**: Command whitelist and validation
3. **Execution Boundary**: Process/container isolation
4. **System Boundary**: Operating system access controls

## Local Execution Security

### Security Characteristics

- **Isolation Level**: None
- **Privilege Level**: Same as MCPFier process
- **Filesystem Access**: Full host access
- **Network Access**: Full network access
- **Resource Limits**: Operating system defaults

### Security Recommendations

#### 1. Process Privileges
Run MCPFier with minimal privileges:

```bash
# Create dedicated user
sudo useradd -r -s /bin/false mcpfier

# Run with restricted privileges
sudo -u mcpfier ./mcpfier --mcp
```

#### 2. Filesystem Permissions
Restrict filesystem access:

```bash
# Limit readable directories
sudo chown -R root:mcpfier /opt/mcpfier/
sudo chmod -R 750 /opt/mcpfier/

# Read-only configuration
sudo chmod 644 /opt/mcpfier/config.yaml
```

#### 3. Network Restrictions
Use firewall rules to limit network access:

```bash
# Allow only necessary outbound connections
sudo iptables -A OUTPUT -u mcpfier -p tcp --dport 80 -j ACCEPT
sudo iptables -A OUTPUT -u mcpfier -p tcp --dport 443 -j ACCEPT
sudo iptables -A OUTPUT -u mcpfier -j DROP
```

## Container Execution Security

### Security Characteristics

- **Isolation Level**: Complete process and filesystem isolation
- **Privilege Level**: Configurable (root by default)
- **Filesystem Access**: Container-only (no host access)
- **Network Access**: Isolated by default
- **Resource Limits**: Configurable via Docker

### Container Security Best Practices

#### 1. Non-Root Execution
Configure containers to run as non-root users:

```yaml
commands:
  - name: secure-python
    script: python
    args: ["-c", "import os; print(f'UID: {os.getuid()}')"]
    container: python:3.9-slim
    description: "Python execution with non-root user"
```

Use custom Dockerfile:

```dockerfile
FROM python:3.9-slim
RUN useradd -r -u 1000 appuser
USER appuser
WORKDIR /home/appuser
```

#### 2. Resource Limits
Implement resource constraints:

```bash
# Docker run with limits (implemented internally)
docker run --rm \
  --memory=128m \
  --cpus=0.5 \
  --disk-quota=100m \
  python:3.9-slim
```

#### 3. Security Options
Add security-focused Docker options:

```yaml
# Future configuration enhancement
commands:
  - name: secure-command
    container: alpine:latest
    security:
      read_only: true
      no_new_privileges: true
      user: "1000:1000"
      capabilities:
        drop: ["ALL"]
        add: ["NET_BIND_SERVICE"]
```

#### 4. Network Isolation
Disable network access when not needed:

```bash
# Future implementation
docker run --rm --network=none alpine:latest
```

## Configuration Security

### Secure Configuration Practices

#### 1. Configuration File Protection
Protect the configuration file:

```bash
# Restrict configuration access
sudo chown root:mcpfier /opt/mcpfier/config.yaml
sudo chmod 640 /opt/mcpfier/config.yaml

# Prevent unauthorized modifications
sudo chattr +i /opt/mcpfier/config.yaml
```

#### 2. Command Validation
Validate all command configurations:

```yaml
commands:
  # Good: Specific, constrained command
  - name: safe-listing
    script: ls
    args: ["-la", "/app/data"]
    container: alpine:latest
    timeout: "10s"

  # Bad: Overly permissive command
  - name: dangerous-shell
    script: sh
    args: ["-c"]  # Allows arbitrary code execution
```

#### 3. Environment Variable Security
Protect sensitive environment variables:

```yaml
commands:
  - name: database-query
    script: /app/query
    container: app:latest
    env:
      # Good: Non-sensitive configuration
      DB_HOST: "db.internal"
      DB_PORT: "5432"
      # Bad: Embedded secrets (use external secret management)
      # DB_PASSWORD: "secret123"
```

### Configuration Validation

Implement pre-deployment validation:

```bash
# Future validation script
./mcpfier --validate-config config.yaml
```

## Authentication and Authorization

### Current State
MCPFier currently operates without authentication, suitable for:
- Local development environments
- Trusted internal networks
- Single-user systems

### Future Authentication Features

#### 1. API Key Authentication
```yaml
server:
  auth:
    type: api_key
    keys:
      - name: "client1"
        key: "sk-..."
        permissions: ["execute"]
```

#### 2. mTLS Authentication
```yaml
server:
  auth:
    type: mtls
    ca_cert: "/path/to/ca.crt"
    server_cert: "/path/to/server.crt"
    server_key: "/path/to/server.key"
```

#### 3. Role-Based Access Control
```yaml
roles:
  - name: "developer"
    commands: ["list-files", "run-tests"]
  - name: "admin"
    commands: ["*"]

users:
  - name: "john"
    roles: ["developer"]
```

## Monitoring and Alerting

### Security Monitoring

#### 1. Execution Monitoring
Monitor command execution patterns:

```bash
# Log all command executions
tail -f /var/log/mcpfier/audit.log | grep "EXEC"
```

#### 2. Resource Monitoring
Track resource usage:

```bash
# Monitor container resource usage
docker stats --format "{{.Container}}: {{.CPUPerc}} {{.MemUsage}}"
```

#### 3. Network Monitoring
Monitor network connections:

```bash
# Monitor outbound connections from MCPFier
netstat -pln | grep mcpfier
```

### Security Alerts

Configure alerts for security events:

```yaml
# Future alerting configuration
alerts:
  - name: "command_execution_failure"
    condition: "error_rate > 10%"
    action: "notify_admin"
  
  - name: "resource_exhaustion"
    condition: "memory_usage > 90%"
    action: "kill_container"
```

## Incident Response

### Security Incident Procedures

#### 1. Detection
- Monitor for unusual command execution patterns
- Track resource usage anomalies
- Watch for authentication failures
- Monitor container escape attempts

#### 2. Containment
- Immediately stop MCPFier server
- Isolate affected containers
- Preserve logs for analysis
- Document timeline of events

#### 3. Investigation
- Analyze command execution logs
- Review configuration changes
- Examine container runtime logs
- Check system security logs

#### 4. Recovery
- Restore from known-good configuration
- Apply security patches
- Update monitoring rules
- Restart services with enhanced monitoring

### Forensic Logging

Enable comprehensive logging:

```yaml
# Future logging configuration
logging:
  level: info
  format: json
  outputs:
    - type: file
      path: /var/log/mcpfier/mcpfier.log
    - type: syslog
      facility: local0
  
  audit:
    enabled: true
    events: ["command_execution", "config_change", "auth_failure"]
```

## Compliance Considerations

### Regulatory Requirements

#### SOC 2 Compliance
- Implement access controls
- Enable audit logging
- Monitor system integrity
- Document security procedures

#### GDPR Compliance
- Implement data minimization
- Enable data portability
- Provide deletion capabilities
- Document data processing

#### HIPAA Compliance
- Encrypt data in transit and at rest
- Implement access controls
- Enable audit logging
- Sign Business Associate Agreements

### Security Standards

#### CIS Controls
- Inventory of authorized software
- Secure configuration management
- Continuous vulnerability management
- Controlled use of administrative privileges

#### NIST Framework
- Identify assets and risks
- Protect against threats
- Detect security events
- Respond to incidents
- Recover from disruptions

## Security Testing

### Vulnerability Scanning

Regular security assessments:

```bash
# Container image scanning
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock \
  aquasec/trivy:latest image python:3.9-slim

# Static code analysis
gosec ./...

# Dependency scanning
go list -json -m all | nancy sleuth
```

### Penetration Testing

Conduct regular penetration tests:
- Command injection attempts
- Container escape testing
- Privilege escalation attempts
- Resource exhaustion testing

### Security Automation

Integrate security into CI/CD:

```yaml
# GitHub Actions security workflow
name: Security Scan
on: [push, pull_request]
jobs:
  security:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Run Gosec
        uses: securecodewarrior/github-action-gosec@master
```

## Deployment Security

### Secure Deployment Practices

#### 1. Minimal Installation
- Use minimal base images
- Remove unnecessary packages
- Disable unused services
- Apply security patches regularly

#### 2. Network Security
- Use private networks
- Implement firewall rules
- Enable TLS encryption
- Monitor network traffic

#### 3. Secret Management
- Use external secret stores
- Rotate secrets regularly
- Avoid hardcoded credentials
- Implement least-privilege access

### Production Hardening

#### System Hardening
```bash
# Disable unnecessary services
sudo systemctl disable cups bluetooth

# Configure kernel parameters
echo "kernel.dmesg_restrict = 1" >> /etc/sysctl.conf
echo "net.ipv4.conf.all.send_redirects = 0" >> /etc/sysctl.conf

# Apply security updates
sudo apt update && sudo apt upgrade -y
```

#### Docker Hardening
```bash
# Enable user namespaces
echo '{"userns-remap": "default"}' > /etc/docker/daemon.json
sudo systemctl restart docker

# Limit container resources by default
echo '{"default-ulimits": {"nproc": {"Name": "nproc", "Hard": 1024, "Soft": 1024}}}' > /etc/docker/daemon.json
```

## Security Roadmap

### Phase 1 (Current)
- Basic container isolation
- Configuration validation
- Execution logging

### Phase 2 (Next)
- Authentication and authorization
- Resource limits enforcement
- Security policy engine
- Audit logging

### Phase 3 (Future)
- Advanced threat detection
- Automated incident response
- Compliance reporting
- Security orchestration

## Conclusion

Security is a critical aspect of MCPFier deployment. Organizations should:

1. Choose appropriate execution modes based on threat model
2. Implement defense in depth strategies
3. Monitor and log security events
4. Regular security assessments and updates
5. Follow industry best practices and compliance requirements

Regular review and updates of security configurations are essential to maintain a secure MCPFier deployment.