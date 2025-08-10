# Security

MCPFier implements multiple layers of security to protect against unauthorized access and execution risks.

## Authentication

### API Key Authentication

MCPFier supports API key-based authentication with granular permission control:

```yaml
server:
  http:
    auth:
      enabled: true
      mode: "simple"
      api_keys:
        "mcpfier_prod_key":
          name: "Production Key"
          permissions: ["list-files", "check-status"]
        "mcpfier_admin_key":
          name: "Admin Key" 
          permissions: ["*"]
```

### Permission Model

- **Granular Control**: Each API key can be restricted to specific commands
- **Wildcard Permissions**: Use `["*"]` for full access (admin keys only)
- **Principle of Least Privilege**: Grant minimum necessary permissions

## Execution Security

### Three Execution Modes

1. **Local Execution**: Runs with MCPFier process privileges
2. **Container Execution**: Complete isolation using Docker
3. **Webhook Execution**: HTTP client calls to external APIs

### Container Isolation

Container mode provides complete isolation:

```yaml
commands:
  - name: isolated-task
    script: python
    args: ["/app/script.py"]
    container: "python:3.9-slim"
    timeout: "30s"
```

**Container Security Features:**
- No host filesystem access
- Network isolation options
- Resource limits (CPU, memory)
- Automatic cleanup after execution
- Read-only container filesystems

### Webhook Security

When calling external APIs:

```yaml
commands:
  - name: api-call
    webhook:
      url: "https://api.example.com/endpoint"
      method: "POST"
      auth:
        type: "bearer"
        token: "${API_TOKEN}"
      headers:
        Content-Type: "application/json"
```

**Webhook Security Measures:**
- TLS certificate validation
- Authentication header support
- Request timeout enforcement
- Retry policy limits
- Input validation

## Configuration Security

### Sensitive Data Handling

- **Environment Variables**: Use `${VAR_NAME}` for sensitive data
- **File Permissions**: Restrict config file access (600/640)
- **Secret Management**: Never commit secrets to version control

### Configuration Validation

- Schema validation on startup
- Command name uniqueness enforcement
- Timeout limit validation
- Permission syntax checking

## Network Security

### HTTP Server

- **TLS Support**: Configure HTTPS for production deployments
- **CORS Policy**: Configurable cross-origin request handling
- **Request Validation**: JSON-RPC 2.0 compliance checking
- **Rate Limiting**: Configurable request rate limits

### Logging and Monitoring

- **Access Logging**: Common Log Format for HTTP requests
- **Analytics Database**: SQLite with secure file permissions
- **Error Tracking**: Detailed error categorization
- **Audit Trail**: Command execution logging

## Best Practices

### Development Environment

- Use development API keys with limited permissions
- Enable analytics to monitor usage patterns
- Test container isolation with untrusted code
- Validate webhook endpoints before deployment

### Production Environment

- **Authentication**: Always enable authentication in production
- **Container Mode**: Use containers for all untrusted operations
- **Network Security**: Deploy behind reverse proxy with TLS
- **Monitoring**: Enable comprehensive logging and analytics
- **Resource Limits**: Configure timeouts and resource constraints
- **Key Rotation**: Regularly rotate API keys
- **Access Control**: Implement network-level access restrictions

### Configuration Management

- Store configuration files with restricted permissions (600)
- Use environment variables for sensitive data
- Implement configuration versioning and backup
- Validate configuration changes before deployment

### Incident Response

- Monitor analytics for unusual patterns
- Implement log aggregation for security events
- Configure alerting for authentication failures
- Maintain incident response procedures

## Vulnerability Management

### Regular Updates

- Keep MCPFier updated to latest version
- Update container base images regularly  
- Monitor dependencies for security advisories
- Apply security patches promptly

### Security Testing

- Test authentication mechanisms
- Validate container isolation
- Verify network security configurations
- Perform regular security assessments

## Compliance Considerations

### Data Privacy

- Analytics data retention policies
- Request logging data handling
- User consent for monitoring
- Data encryption at rest and in transit

### Access Control

- Role-based permission model
- Audit logging requirements
- Access review procedures
- Privileged access management

## Reporting Security Issues

Report security vulnerabilities through the project's issue tracker. Include:

- Detailed vulnerability description
- Steps to reproduce
- Impact assessment
- Suggested mitigation

Security issues will be addressed with priority and may result in emergency releases.