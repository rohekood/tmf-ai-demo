# releases/

Helm charts and Flux CD GitOps manifests for the TMF AI Demo platform.

## Directory Structure

```
releases/
├── charts/                          # Helm charts for all custom services
│   ├── bff/                         # Backend-for-Frontend (port 8080)
│   ├── customer-management/         # Customer Management service (port 8081)
│   ├── demo-ui/                     # Frontend SPA served by nginx (port 80)
│   ├── party-management/            # Party Management service (port 8080)
│   ├── pocv/                        # POCV consumer service (no HTTP)
│   ├── product-catalog-management/  # Product Catalog consumer service (no HTTP)
│   ├── qualification/               # Qualification consumer + Redis cache (no HTTP)
│   └── shopping-cart/               # Shopping Cart consumer service (no HTTP)
└── flux/                            # Flux CD manifests
    ├── namespace.yaml               # tmf namespace
    ├── kustomization.yaml           # Root kustomization (apply this to bootstrap)
    ├── tmf-secrets.example.yaml     # Secret template — fill in real values, do not commit
    ├── sources/
    │   ├── gitrepository.yaml       # GitRepository: this repo (chart source)
    │   └── bitnami-helmrepository.yaml  # HelmRepository: bitnami (infra charts)
    ├── infrastructure/
    │   ├── postgres-release.yaml    # PostgreSQL HelmRelease (bitnami)
    │   ├── rabbitmq-release.yaml    # RabbitMQ HelmRelease (bitnami)
    │   └── redis-release.yaml       # Redis HelmRelease (bitnami, for qualification cache)
    └── apps/
        ├── bff.yaml
        ├── customer-management.yaml
        ├── demo-ui.yaml
        ├── party-management.yaml
        ├── pocv.yaml
        ├── product-catalog-management.yaml
        ├── qualification.yaml
        └── shopping-cart.yaml
```

## Prerequisites

1. A running Kubernetes cluster with [Flux CD v2](https://fluxcd.io/docs/installation/) bootstrapped.
2. Helm 3 installed locally (for testing/dry-runs).
3. Container images built and pushed to `ghcr.io/rohekood/tmf-ai-demo/<service-name>`.

## Quick Start with Flux

### 1. Create the Secret

Copy and fill in `releases/flux/tmf-secrets.example.yaml`, then apply it:

```bash
cp releases/flux/tmf-secrets.example.yaml /tmp/tmf-secrets.yaml
# edit /tmp/tmf-secrets.yaml — set real passwords and connection strings
kubectl apply -f /tmp/tmf-secrets.yaml
```

> **Never commit a Secret with real credentials.** Use [Sealed Secrets](https://github.com/bitnami-labs/sealed-secrets), [SOPS](https://fluxcd.io/docs/guides/mozilla-sops/), or an external secrets operator.

### 2. Bootstrap Flux (if not already done)

```bash
flux bootstrap github \
  --owner=rohekood \
  --repository=tmf-ai-demo \
  --branch=main \
  --path=releases/flux
```

### 3. Apply the Kustomization manually (alternative)

```bash
kubectl apply -k releases/flux/
```

Flux will reconcile the `GitRepository`, install infrastructure (Postgres, RabbitMQ, Redis), then deploy all application services in dependency order.

## Secret Keys Reference

The `tmf-secrets` Secret must contain the following keys:

| Key | Used by |
|-----|---------|
| `CUSTOMER_DB_URL` | customer-management |
| `PARTY_DB_URL` | party-management |
| `CATALOG_DB_URL` | product-catalog-management |
| `CART_DB_URL` | shopping-cart |
| `QUALIFICATION_DB_URL` | qualification |
| `POCV_DB_URL` | pocv |
| `POSTGRES_PASSWORD` | postgres HelmRelease |
| `RABBITMQ_URL` | all backend services (uses `tmf` user) |
| `RABBITMQ_PASSWORD` | rabbitmq HelmRelease |

## Customising Values

Override Helm values in each `releases/flux/apps/<service>.yaml` under `spec.values`, for example:

```yaml
spec:
  values:
    replicaCount: 2
    image:
      tag: "1.2.3"
    resources:
      requests:
        cpu: 100m
        memory: 128Mi
      limits:
        cpu: 500m
        memory: 256Mi
```

## Local Dry-Run

```bash
helm template customer-management releases/charts/customer-management \
  --set env.secretName=tmf-secrets \
  --set image.tag=latest
```
