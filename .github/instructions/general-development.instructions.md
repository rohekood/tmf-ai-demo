---
description: This instruction are general rules for development in the project
applyTo: 'this must be used for all files in the project'
---

* this is monorepo
* every service have docs directory
* you have to always develope code according to following files in docs:
 * ARCHITECTURE.md
 * ANALYSIS.md
* You must always write tests for every functionality you create
 * in ui
 * in backend
* if bug is found you have to first create test and approve that test failes
 * after that you can start fixing bug
* always read through relevant docs before starting development, and ask for help if you can't find relevant docs or if something is unclear

use golint as $(go env GOPATH)/bin/golangci-lint