# Production Secrets

This document describes the per-service credentials to inject for production deployments.
Local development uses `guest:guest` for all services (see `docker-compose.yml`).

## RabbitMQ

Each service reads a single `RABBITMQ_URL` environment variable.
In production, inject a K8s Secret per service containing the service-specific URL.

Run `scripts/rabbitmq-setup-prod.sh` once against the production broker to create the users.

Actual passwords are NOT stored in this repo. Store them in your secrets manager
(Vault, AWS Secrets Manager, GCP Secret Manager, etc.) and reference them in K8s Secrets.
See `.env.production.example` for the variable names expected per service.

### Per-service RabbitMQ usernames

| Service | RabbitMQ user |
|---|---|
| party-management | `tmf_party` |
| customer-management | `tmf_customer` |
| product-catalog-management | `tmf_catalog` |
| qualification | `tmf_qualification` |
| shopping-cart | `tmf_cart` |
| pocv | `tmf_pocv` |
| bff | `tmf_bff` |

### K8s Secret example (party-management)

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: party-management-secrets
type: Opaque
stringData:
  RABBITMQ_URL: "amqp://tmf_party:<password>@rabbitmq:5672/tmf"
```

Reference it in the Deployment:

```yaml
envFrom:
  - secretRef:
      name: party-management-secrets
```
