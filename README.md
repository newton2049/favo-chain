
![Banner](.github/banner.jpg)

---

# Favo Chain

A blockchain framework for building EVM-compatible networks.

## Quick Start

### Prerequisites

- Go 1.18.x
- Docker & Docker Compose 2+ (for containerized deployment)
- Node.js & npm (for favobft consensus)

### Build

```bash
make build
```

### Run Local Cluster

```bash
# Using scripts (recommended)
scripts/cluster ibft

# Using Docker Compose
scripts/cluster ibft --docker
```

## Documentation

- [Deployment Guide](docs/deployment/DEPLOYMENT_GUIDE.md) - Comprehensive deployment best practices for Development, Testing, and Production environments
- [Docker Local Setup](docker/README.md) - Local Docker development guide
- [Scripts Usage](scripts/README.md) - Running from local binary

## Environment Configurations

| Environment | Documentation | Configuration Example |
|-------------|---------------|----------------------|
| Development | [Guide](docs/deployment/DEPLOYMENT_GUIDE.md#development-environment) | [env.development.example](docs/deployment/env.development.example) |
| Staging | [Guide](docs/deployment/DEPLOYMENT_GUIDE.md#testingstaging-environment) | [env.staging.example](docs/deployment/env.staging.example) |
| Production | [Guide](docs/deployment/DEPLOYMENT_GUIDE.md#production-environment) | [env.production.example](docs/deployment/env.production.example) |

## Deployment Options

### Docker Compose

- [Development](docs/deployment/docker-compose/docker-compose.development.yml)
- [Staging](docs/deployment/docker-compose/docker-compose.staging.yml)
- [Production](docs/deployment/docker-compose/docker-compose.production.yml) (reference only)

### Kubernetes

See [Kubernetes deployment guide](docs/deployment/kubernetes/README.md) for production-grade deployments.

## CI/CD Pipeline

The repository includes automated workflows for:

| Workflow | Trigger | Description |
|----------|---------|-------------|
| CI | Pull Request | Build and test |
| DevNet | Push to `develop` | Deploy to development network |
| TestNet | Push to `release/*` | Deploy to test network |
| Release | Tag `v*.*.*` | Create releases |

## Testing

```bash
# Unit tests
make test

# E2E tests
make test-e2e

# FavoBFT E2E tests
make test-e2e-favobft
```

## License

Copyright 2022 Favo Technology

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
