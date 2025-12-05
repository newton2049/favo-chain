# Docker Compose Configuration Examples
# Docker Compose 配置示例

This directory contains Docker Compose configurations for different environments.

## Files

| File | Description |
|------|-------------|
| `docker-compose.development.yml` | Development environment with debug logging |
| `docker-compose.staging.yml` | Staging environment with production-like settings |
| `docker-compose.production.yml` | Production-grade configuration |

## Usage

### Development

```bash
# Start development environment
docker-compose -f docker-compose.development.yml up -d

# View logs
docker-compose -f docker-compose.development.yml logs -f

# Stop and clean up
docker-compose -f docker-compose.development.yml down -v
```

### Staging

```bash
# Create .env file with staging configuration
cp ../env.staging.example .env

# Start staging environment
docker-compose -f docker-compose.staging.yml up -d
```

### Production

⚠️ **Warning**: Production Docker Compose is provided for reference only. 
For actual production deployments, use Kubernetes or managed container services.

```bash
# Ensure secrets are properly configured
# Start production environment
docker-compose -f docker-compose.production.yml up -d
```

## Customization

1. Copy the appropriate environment example file to `.env`
2. Update values as needed
3. Run the corresponding docker-compose file

## Notes

- Development uses `--insecure` flag for convenience
- Staging and Production should use proper secret management
- Production configuration includes resource limits and health checks
