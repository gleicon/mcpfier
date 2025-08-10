# Docker Deployment

MCPFier can be deployed as a Docker container with persistent storage for configuration and analytics data.

## Building the Image

```bash
# Build the image
docker build -t mcpfier:latest .

# Tag for registry (optional)
docker tag mcpfier:latest your-registry.com/mcpfier:latest

# Push to registry (optional)
docker push your-registry.com/mcpfier:latest
```

## Directory Setup

Create the required directories on your host system:

```bash
# Create directory structure
mkdir -p /opt/mcpfier/{config,data,logs}

# Set proper permissions
sudo chown -R 1001:1001 /opt/mcpfier
```

## Configuration File

Place your configuration file at `/opt/mcpfier/config/config.yaml`:

```yaml
commands:
  - name: echo-test
    script: echo
    args: ["Hello from container"]
    description: "Test command"

server:
  http:
    enabled: true
    host: "0.0.0.0"  # Important: bind to all interfaces in container
    port: 8080
    auth:
      enabled: true
      mode: "simple"
      api_keys:
        "mcpfier_container_key":
          name: "Container Key"
          permissions: ["*"]

analytics:
  enabled: true
  database_path: "/app/data/analytics.db"  # Container path
  retention_days: 30
```

## Running the Container

### Basic Run

```bash
docker run -d \
  --name mcpfier \
  -p 8080:8080 \
  -v /opt/mcpfier/config:/app/config:ro \
  -v /opt/mcpfier/data:/app/data \
  -v /opt/mcpfier/logs:/app/logs \
  mcpfier:latest
```

### Production Run with Additional Options

```bash
docker run -d \
  --name mcpfier \
  --restart unless-stopped \
  -p 8080:8080 \
  -v /opt/mcpfier/config:/app/config:ro \
  -v /opt/mcpfier/data:/app/data \
  -v /opt/mcpfier/logs:/app/logs \
  --memory=512m \
  --cpus=1 \
  --health-cmd="wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1" \
  --health-interval=30s \
  --health-timeout=3s \
  --health-retries=3 \
  mcpfier:latest
```

## Volume Mounts Explained

| Host Path             | Container Path | Purpose             | Mount Type |
| --------------------- | -------------- | ------------------- | ---------- |
| `/opt/mcpfier/config` | `/app/config`  | Configuration files | Read-only  |
| `/opt/mcpfier/data`   | `/app/data`    | Analytics database  | Read-write |
| `/opt/mcpfier/logs`   | `/app/logs`    | Application logs    | Read-write |

## Container Management

### View Logs

```bash
# View container logs
docker logs mcpfier

# Follow logs in real-time
docker logs -f mcpfier

# View last 100 lines
docker logs --tail 100 mcpfier
```

### Health Check

```bash
# Check container health status
docker inspect mcpfier | grep -A 10 "Health"

# Manual health check
docker exec mcpfier wget --no-verbose --tries=1 --spider http://localhost:8080/health
```

### Container Shell Access

```bash
# Execute shell in running container
docker exec -it mcpfier /bin/sh

# Check running processes
docker exec mcpfier ps aux

# View mounted volumes
docker exec mcpfier ls -la /app/
```

### Stopping and Cleanup

```bash
# Stop container
docker stop mcpfier

# Remove container
docker rm mcpfier

# Remove image
docker rmi mcpfier:latest
```

## Analytics Access

Access the analytics dashboard at:

```bash
http://your-host:8080/mcpfier/analytics
```

Use the API key configured in your `config.yaml` for authentication.

## Security Considerations

### Container Security

- Container runs as non-root user (UID 1001)
- Read-only configuration mount prevents tampering
- Resource limits prevent resource exhaustion
- Health checks enable monitoring integration

### Network Security

- Bind to specific interfaces in production
- Use reverse proxy with TLS termination
- Implement firewall rules as needed
- Consider container network isolation

### Data Security

- Ensure host directory permissions are restrictive (700/750)
- Regular backup of `/opt/mcpfier/data` directory
- Rotate API keys regularly
- Monitor analytics for suspicious activity

## Troubleshooting

### Permission Issues

```bash
# Fix ownership
sudo chown -R 1001:1001 /opt/mcpfier

# Fix permissions
sudo chmod -R 755 /opt/mcpfier
sudo chmod 600 /opt/mcpfier/config/config.yaml
```

### Container Won't Start

```bash
# Check container logs
docker logs mcpfier

# Verify configuration file
docker run --rm -v /opt/mcpfier/config:/app/config:ro mcpfier:latest ./mcpfier --config /app/config/config.yaml --help
```

### Database Issues

```bash
# Check database file
ls -la /opt/mcpfier/data/

# Reset analytics database (caution: deletes all data)
sudo rm /opt/mcpfier/data/analytics.db*
```

## Integration with Container Orchestration

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mcpfier
spec:
  replicas: 1
  selector:
    matchLabels:
      app: mcpfier
  template:
    metadata:
      labels:
        app: mcpfier
    spec:
      containers:
      - name: mcpfier
        image: mcpfier:latest
        ports:
        - containerPort: 8080
        volumeMounts:
        - name: config
          mountPath: /app/config
          readOnly: true
        - name: data
          mountPath: /app/data
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
        resources:
          limits:
            memory: "512Mi"
            cpu: "500m"
          requests:
            memory: "256Mi"
            cpu: "250m"
      volumes:
      - name: config
        configMap:
          name: mcpfier-config
      - name: data
        persistentVolumeClaim:
          claimName: mcpfier-data
```

## Registry Usage

### Public Registry

```bash
# Pull from registry
docker pull your-registry.com/mcpfier:latest

# Run from registry
docker run -d \
  --name mcpfier \
  -p 8080:8080 \
  -v /opt/mcpfier/config:/app/config:ro \
  -v /opt/mcpfier/data:/app/data \
  -v /opt/mcpfier/logs:/app/logs \
  your-registry.com/mcpfier:latest
```

### Private Registry

```bash
# Login to private registry
docker login your-private-registry.com

# Pull and run
docker pull your-private-registry.com/mcpfier:latest
docker run -d --name mcpfier \
  -p 8080:8080 \
  -v /opt/mcpfier/config:/app/config:ro \
  -v /opt/mcpfier/data:/app/data \
  your-private-registry.com/mcpfier:latest
```
