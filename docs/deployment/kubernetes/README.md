# Kubernetes Deployment Configuration for Favo Chain
# Kubernetes 部署配置

This directory contains Kubernetes manifests for deploying Favo Chain in a production environment.

## Prerequisites

- Kubernetes 1.21+
- kubectl configured
- Persistent storage provisioner (e.g., AWS EBS, GCE PD)
- Ingress controller (e.g., nginx-ingress, AWS ALB)

## Files

| File | Description |
|------|-------------|
| `namespace.yaml` | Namespace definition |
| `configmap.yaml` | Configuration data |
| `secret.yaml` | Secret configuration (template) |
| `statefulset.yaml` | Validator StatefulSet |
| `service.yaml` | Service definitions |
| `ingress.yaml` | Ingress configuration |

## Quick Start

```bash
# Create namespace
kubectl apply -f namespace.yaml

# Create ConfigMap and Secrets
kubectl apply -f configmap.yaml
kubectl apply -f secret.yaml  # Edit with your secrets first!

# Deploy validators
kubectl apply -f statefulset.yaml

# Create services
kubectl apply -f service.yaml

# (Optional) Create ingress
kubectl apply -f ingress.yaml
```

## Customization

1. Update `configmap.yaml` with your genesis configuration
2. Update `secret.yaml` with validator private keys (use Kubernetes secrets management)
3. Adjust `statefulset.yaml` resources based on your requirements
4. Configure `ingress.yaml` with your domain and TLS certificates

## Production Recommendations

1. **Use Separate Nodes**: Deploy each validator on a separate Kubernetes node
2. **Anti-Affinity**: Configure pod anti-affinity to spread validators
3. **Resource Limits**: Set appropriate CPU and memory limits
4. **Monitoring**: Deploy Prometheus ServiceMonitor
5. **Backups**: Configure persistent volume snapshots
