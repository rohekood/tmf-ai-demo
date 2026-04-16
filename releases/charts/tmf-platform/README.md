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
- `customerManagement`
- `demoUi`
- `partyManagement`
- `pocv`
- `productCatalogManagement`
- `qualification`
- `shoppingCart`

Example:

```bash
helm upgrade --install tmf-platform releases/charts/tmf-platform \
  --namespace tmf \
  --set bff.image.tag=main-123 \
  --set customerManagement.image.tag=main-123
```
