# Favo Chain Deployment Guide

## 部署最佳实践指南 (Best Practices Deployment Guide)

This document provides comprehensive deployment guidelines for Development, Testing, and Production environments.

---

## Table of Contents

1. [Overview](#overview)
2. [Environment Architecture](#environment-architecture)
3. [Development Environment](#development-environment)
4. [Testing/Staging Environment](#testingstaging-environment)
5. [Production Environment](#production-environment)
6. [CI/CD Pipeline](#cicd-pipeline)
7. [Monitoring and Logging](#monitoring-and-logging)
8. [Security Best Practices](#security-best-practices)

---

## Overview

Favo Chain supports three primary environments:

| Environment | Purpose | Branch | Trigger |
|-------------|---------|--------|---------|
| Development | Local development & feature testing | `feature/*`, `fix/*` | Manual |
| Staging/Test | Integration testing & QA | `develop` | Push to develop |
| Production | Live network deployment | `release/*`, `main` | Tag release |

### 环境概述

- **开发环境 (Development)**: 本地开发，快速迭代，单节点或4节点本地集群
- **测试环境 (Staging)**: 集成测试，模拟生产配置，自动化部署
- **生产环境 (Production)**: 正式网络，高可用配置，完整监控

---

## Environment Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                         Development                                  │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │  Local Machine / Docker Compose                              │   │
│  │  - 4 Validator Nodes (local)                                 │   │
│  │  - Single shared volume                                      │   │
│  │  - Hot reload capability                                     │   │
│  └─────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│                       Staging/DevNet                                 │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │  AWS/Cloud Infrastructure                                    │   │
│  │  - 4 Validator Nodes (separate instances)                    │   │
│  │  - Load Balancer (RPC endpoint)                              │   │
│  │  - Auto-scaling groups                                       │   │
│  │  - Automated deployment via CI/CD                            │   │
│  └─────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│                         Production                                   │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │  Kubernetes / AWS ECS / Bare Metal                           │   │
│  │  - N Validator Nodes (distributed across zones)              │   │
│  │  - High-availability load balancing                          │   │
│  │  - Persistent storage                                        │   │
│  │  - Monitoring & Alerting (Prometheus/Grafana)                │   │
│  │  - Log aggregation (ELK/CloudWatch)                          │   │
│  └─────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Development Environment

### Prerequisites

- Go 1.18.x
- Docker & Docker Compose 2+
- Node.js & npm (for favobft consensus)
- Make

### Quick Start

#### Option 1: Using Scripts (Recommended)

```bash
# Clone repository
git clone https://github.com/newton2049/favo-chain.git
cd favo-chain

# Download submodules (required for favobft)
make download-submodules

# Start local cluster with IBFT consensus
scripts/cluster ibft

# OR start with FavoBFT consensus
scripts/cluster favobft
```

#### Option 2: Using Docker Compose

```bash
# For IBFT consensus
scripts/cluster ibft --docker

# For FavoBFT consensus
scripts/cluster favobft --docker

# Stop containers
scripts/cluster ibft --docker stop

# Destroy environment
scripts/cluster ibft --docker destroy
```

### Development Configuration

Create a `.env.development` file:

```bash
# Development Environment Configuration
CHAIN_ID=51001
NETWORK_NAME=favo-chain-dev
BLOCK_GAS_LIMIT=10000000
EPOCH_SIZE=10
LOG_LEVEL=DEBUG

# Pre-mine accounts (development only)
PREMINE_ACCOUNTS=0x228466F2C715CbEC05dEAbfAc040ce3619d7CF0B:0xD3C21BCECCEDA1000000

# JSON-RPC
JSONRPC_PORT=8545
GRPC_PORT=9632
LIBP2P_PORT=1478

# Prometheus metrics
PROMETHEUS_PORT=5001
```

### Development Ports

| Service | Node 1 | Node 2 | Node 3 | Node 4 |
|---------|--------|--------|--------|--------|
| JSON-RPC | 10002 | 20002 | 30002 | 40002 |
| gRPC | 10000 | 20000 | 30000 | 40000 |
| libp2p | 30301 | 30302 | 30303 | 30304 |
| Prometheus | 10003 | 20003 | 30003 | 40003 |

### Best Practices for Development

1. **Use Debug Logging**: Enable `--log-level DEBUG` for detailed output
2. **Hot Reload**: Use `go build && ./favo-chain server` for quick iterations
3. **Small Epoch Size**: Use `--epoch-size 10` for faster testing of epoch-related features
4. **Test Accounts**: Use pre-mined accounts for testing transactions
5. **Clean State**: Regularly clean `test-chain-*` directories for fresh starts

---

## Testing/Staging Environment

### Overview

The staging environment mirrors production but allows for automated testing and QA validation.

### Automated Deployment

Staging deployments are triggered automatically when code is pushed to the `develop` branch:

```yaml
# .github/workflows/deploy.devnet.yml
on:
  push:
    branches:
      - develop
```

### Staging Configuration

Create a `.env.staging` file:

```bash
# Staging Environment Configuration
CHAIN_ID=51002
NETWORK_NAME=favo-chain-staging
BLOCK_GAS_LIMIT=20000000
EPOCH_SIZE=100
LOG_LEVEL=INFO

# Consensus
CONSENSUS=ibft  # or favobft

# Security
NUM_BLOCK_CONFIRMATIONS=2

# Monitoring
PROMETHEUS_ENABLED=true
PROMETHEUS_PORT=5001

# Network
JSONRPC_PORT=8545
GRPC_PORT=9632
LIBP2P_PORT=1478
```

### AWS Infrastructure Requirements

```bash
# Validator Auto-Scaling Groups (recommended)
VALIDATOR_ASGS="validator-1-asg,validator-2-asg,validator-3-asg,validator-4-asg"

# S3 Bucket for artifacts
FAVO_EDGE_ARTIFACT_BUCKET="favo-chain-artifacts-staging"

# ECR Repository
ECR_REPOSITORY="library/favo-chain"
```

### Load Testing

After deployment, run load tests to validate performance:

```bash
# Run EOA transfer load test
make loadtest-eoa

# Run ERC20 transfer load test
make loadtest-erc20
```

### Testing Checklist

- [ ] Unit tests pass (`make test`)
- [ ] E2E tests pass (`make test-e2e`)
- [ ] FavoBFT E2E tests pass (`make test-e2e-favobft`)
- [ ] Load tests complete successfully
- [ ] Security scan (Snyk) passes
- [ ] Code quality (SonarQube) passes

---

## Production Environment

### Overview

Production deployment requires careful planning and security considerations.

### Pre-Production Checklist

- [ ] All staging tests passed
- [ ] Security audit completed
- [ ] Backup strategy defined
- [ ] Monitoring & alerting configured
- [ ] Incident response plan documented
- [ ] Key management strategy implemented

### Production Configuration

Create a `.env.production` file:

```bash
# Production Environment Configuration
CHAIN_ID=51000
NETWORK_NAME=favo-chain-mainnet
BLOCK_GAS_LIMIT=30000000
EPOCH_SIZE=1000
LOG_LEVEL=WARN

# Security - NEVER use --insecure in production
SECRETS_MANAGER=hashicorp-vault  # or aws-ssm

# High availability
NUM_BLOCK_CONFIRMATIONS=12
MAX_PEERS=100
MAX_INBOUND_PEERS=50
MAX_OUTBOUND_PEERS=50

# Performance
BLOCK_TIME=2s
PRICE_LIMIT=1000000000  # 1 Gwei

# Monitoring
PROMETHEUS_ENABLED=true
PROMETHEUS_PORT=5001
```

### Kubernetes Deployment

For production-grade deployments, use Kubernetes:

```yaml
# See docs/deployment/kubernetes/ for full configurations
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: favo-chain-validator
spec:
  replicas: 4
  serviceName: favo-chain
  template:
    spec:
      containers:
      - name: favo-chain
        image: <ecr-registry>/favo-chain:latest
        ports:
        - containerPort: 8545
          name: jsonrpc
        - containerPort: 9632
          name: grpc
        - containerPort: 1478
          name: libp2p
        - containerPort: 5001
          name: prometheus
        volumeMounts:
        - name: data
          mountPath: /data
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 100Gi
```

### High Availability Setup

1. **Geographic Distribution**: Deploy validators across multiple availability zones
2. **Load Balancing**: Use ALB/NLB for RPC endpoint distribution
3. **Redundancy**: Maintain at least N+1 validators where N is the BFT threshold
4. **Failover**: Configure automatic failover for critical services

### Backup Strategy

```bash
# Daily backup of chain data
#!/bin/bash
DATE=$(date +%Y%m%d)
BACKUP_DIR="/backups/favo-chain/${DATE}"

# Stop node gracefully
systemctl stop favo-chain

# Create backup
tar -czvf ${BACKUP_DIR}/chain-data.tar.gz /data/favo-chain/

# Start node
systemctl start favo-chain

# Upload to S3 (or other cloud storage)
aws s3 cp ${BACKUP_DIR}/chain-data.tar.gz s3://backups/favo-chain/${DATE}/
```

### Release Process

1. Create release branch: `git checkout -b release/v1.x.x`
2. Push to trigger TestNet deployment
3. Verify TestNet stability (minimum 24-48 hours)
4. Create and push tag: `git tag v1.x.x && git push --tags`
5. GoReleaser creates binary releases
6. Deploy to production validators

---

## CI/CD Pipeline

### Pipeline Overview

```
┌───────────┐    ┌───────────┐    ┌───────────┐    ┌───────────┐
│   Lint    │───▶│   Build   │───▶│   Test    │───▶│  Security │
└───────────┘    └───────────┘    └───────────┘    └───────────┘
                                                         │
                 ┌──────────────────────────────────────┘
                 ▼
┌───────────────────────────────────────────────────────────────┐
│                    Environment Deployment                      │
├───────────────────┬───────────────────┬───────────────────────┤
│     DevNet        │     TestNet       │      Production        │
│  (develop branch) │  (release/*)      │      (tags v*.*.*)    │
└───────────────────┴───────────────────┴───────────────────────┘
```

### Workflow Files

| Workflow | Trigger | Purpose |
|----------|---------|---------|
| `ci.yml` | Pull Request | Build & Test |
| `lint.yml` | Push (non-deploy branches) | Code quality |
| `deploy.devnet.yml` | Push to develop | Deploy to DevNet |
| `deploy.testnet.yml` | Push to release/* | Deploy to TestNet |
| `release.yml` | Tag v*.*.* | Create releases |
| `security.yml` | Schedule/Manual | Security scanning |

---

## Monitoring and Logging

### Prometheus Metrics

Favo Chain exposes metrics on the Prometheus endpoint (default: `:5001`):

```bash
# Enable Prometheus in node configuration
favo-chain server --prometheus 0.0.0.0:5001
```

### Key Metrics to Monitor

| Metric | Description | Alert Threshold |
|--------|-------------|-----------------|
| `chain_head_block` | Current block number | N/A |
| `txpool_pending` | Pending transactions | > 10000 |
| `peers_count` | Connected peers | < 3 |
| `consensus_round` | Current consensus round | Stuck for > 30s |

### Grafana Dashboard

Import the provided Grafana dashboard for visualization:
- Dashboard ID: (To be configured)
- Data source: Prometheus

### Log Aggregation

Configure centralized logging:

```bash
# Example: Shipping logs to CloudWatch
favo-chain server --log-to /var/log/favo-chain/node.log 2>&1 | tee /var/log/favo-chain/node.log
```

---

## Security Best Practices

### Key Management

1. **Never use `--insecure` flag in production**
2. Use secure secret storage:
   - HashiCorp Vault
   - AWS Secrets Manager
   - Azure Key Vault

```bash
# Initialize secrets with secure storage
favo-chain secrets init --data-dir /data --secrets-manager hashicorp-vault
```

### Network Security

1. **Firewall Rules**:
   ```bash
   # Allow JSON-RPC only from trusted sources
   ufw allow from 10.0.0.0/8 to any port 8545
   
   # Allow gRPC from internal network
   ufw allow from 10.0.0.0/8 to any port 9632
   
   # Allow libp2p from anywhere (required for consensus)
   ufw allow 1478/tcp
   ```

2. **TLS/SSL**: Enable TLS for all external-facing endpoints

3. **Rate Limiting**: Implement rate limiting for JSON-RPC endpoints

### Access Control

1. Restrict SSH access to validators
2. Use bastion hosts for administrative access
3. Implement role-based access control (RBAC)
4. Enable audit logging

### Security Scanning

Security scans run automatically:
- Snyk: Dependency vulnerability scanning
- SonarQube: Code quality and security analysis
- CodeQL: GitHub's semantic code analysis

---

## Quick Reference

### Commands Summary

```bash
# Development
make build                    # Build binary
make test                     # Run unit tests
make test-e2e                 # Run E2E tests
make lint                     # Run linter
scripts/cluster ibft          # Start local cluster

# Docker
scripts/cluster ibft --docker        # Start Docker cluster
scripts/cluster ibft --docker stop   # Stop cluster
scripts/cluster ibft --docker destroy # Destroy cluster

# Production
favo-chain secrets init --data-dir /data
favo-chain genesis --consensus ibft ...
favo-chain server --data-dir /data --seal
```

### Environment Variables Reference

| Variable | Development | Staging | Production |
|----------|-------------|---------|------------|
| `CHAIN_ID` | 51001 | 51002 | 51000 |
| `LOG_LEVEL` | DEBUG | INFO | WARN |
| `EPOCH_SIZE` | 10 | 100 | 1000 |
| `BLOCK_CONFIRMATIONS` | 2 | 2 | 12 |
| `INSECURE_SECRETS` | Yes | No | **Never** |

---

## Troubleshooting

### Common Issues

1. **Nodes not connecting**
   - Check firewall rules for libp2p port (1478)
   - Verify bootnode addresses in genesis
   - Check peer discovery logs

2. **Consensus failures**
   - Ensure sufficient validators are running (minimum 2f+1)
   - Check clock synchronization (NTP)
   - Review consensus logs for errors

3. **Transaction failures**
   - Verify gas limit settings
   - Check account balance
   - Review transaction pool status

### Support

- Documentation: https://wiki.favo.technology/docs/edge/overview/
- Issues: https://github.com/newton2049/favo-chain/issues

---

*Last Updated: December 2024*
