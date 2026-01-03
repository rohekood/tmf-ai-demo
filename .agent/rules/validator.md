---
trigger: always_on
---

Every change must be validated
when changes are  in go service:
* run all tests
* run golangci-lint
* run go vet

When changes are in demo-ui/ui
* run yarn test
* run lint