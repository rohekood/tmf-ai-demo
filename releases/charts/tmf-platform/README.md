# tmf-platform umbrella chart

This chart deploys all TMF AI Demo application charts in one Helm release.

## Included subcharts

- bff
- customer-management
- demo-ui
- party-management
- pocv
- product-catalog-management
- qualification
- shopping-cart

## Usage

```bash
helm dependency update releases/charts/tmf-platform

helm upgrade --install tmf-platform releases/charts/tmf-platform \
  --namespace tmf \
  --create-namespace
```

## Overriding child values

Each child chart is configured under its alias key in `values.yaml`:

- `bff`
- `customer-management`
- `demo-ui`
- `party-management`
- `pocv`
- `product-catalog-management`
- `qualification`
- `shopping-cart`

Example:

```bash
helm upgrade --install tmf-platform releases/charts/tmf-platform \
  --namespace tmf \
  --set bff.image.tag=main-123 \
  --set customer-management.image.tag=main-123
```

## Emissary ingress

You can expose both UI and BFF through Emissary with one host/IP.

Enable Emissary resources in this chart:

```bash
helm upgrade --install tmf-platform releases/charts/tmf-platform \
  --namespace tmf \
  --set emissary.enabled=true
```

This renders:

- `Mapping` for `/api/` -> `tmf-platform-bff:8080`
- `Mapping` for `/` -> `tmf-platform-demo-ui:80`

Optional:

- `--set emissary.hostname=<dns-or-ip>` to require a specific host header.
- `--set emissary.createHost=true` to create an Emissary `Host` resource.
- `--set emissary.host.tlsSecretName=<secret>` for TLS on that Host.
