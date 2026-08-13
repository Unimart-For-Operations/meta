# Documentation

Internal Developer Platform builder. Private repo derived from [cnoe-io/idpbuilder](https://github.com/cnoe-io/idpbuilder).

- Repository: https://github.com/Unimart-For-Operations/idpbuilder (private)
- Upstream: https://github.com/cnoe-io/idpbuilder (public, read-only reference)

## Overview

`idpbuilder create` provisions a local Kubernetes cluster (Kind) with:

- **ArgoCD** — GitOps continuous delivery
- **Gitea** — Self-hosted Git service
- **Ingress-NGINX** — Cluster ingress controller

Access the platform at:
- ArgoCD: https://argocd.cnoe.localtest.me:8443
- Gitea: https://gitea.cnoe.localtest.me:8443

## Architecture

![idpbuilder architecture](images/idpbuilder.png)

**Two-stage manifest pipeline:**

1. **Build time** (`hack/embedded-resources.sh`): Generates Kubernetes manifests from upstream sources via `kustomize build` (ArgoCD, Nginx) and `helm template` (Gitea)
2. **Compile time**: Generated manifests are embedded into the Go binary via `//go:embed`
3. **Runtime** (`Build.Run()`): Creates Kind cluster, installs core packages directly, then hands off to GitOps — pushes manifests to in-cluster Gitea, creates ArgoCD Applications

See [Architecture/README.md](Architecture/README.md) for the full two-phase design.

## Contents

| Section | Description |
|---------|-------------|
| [Architecture](Architecture/README.md) | CLI phase, controller phase, reconcilers, build pipeline |
| [Contributing](Contributing/README.md) | Dev setup, building, testing, component upgrades, DCO |
| [Getting Started](Getting-Started/README.md) | Prerequisites, first cluster, credentials |
| [Reference](Reference/README.md) | Pluggable packages, private registries, upstream management |

## Key Directories

| Path | Purpose |
|------|---------|
| `hack/argo-cd/` | ArgoCD kustomization, patches, ingress template |
| `hack/ingress-nginx/` | Nginx kustomization, service template |
| `hack/gitea/` | Gitea Helm values, ingress template |
| `hack/embedded-resources.sh` | Master script that runs all 3 generate scripts |
| `pkg/build/build.go` | Main orchestration (`Build.Run()`) |
| `pkg/controllers/localbuild/` | Core reconciler + embedded resource installers |
| `pkg/controllers/localbuild/resources/` | Generated embedded manifests |
| `pkg/controllers/gitrepository/` | GitRepository reconciler (pushes to Gitea) |
| `pkg/controllers/custompackage/` | CustomPackage reconciler (user-provided ArgoCD apps) |
| `api/` | CRD type definitions (Localbuild, GitRepository, CustomPackage) |

## Note on pluggable-packages.md

The [pluggable packages reference](Reference/pluggable-packages.md) is a **design proposal document** from upstream. Much of the proposed architecture has been implemented in the current codebase. The `cnoe://` URL scheme described in [Contributing/README.md](Contributing/README.md) is the implemented version of the `files://` scheme discussed in the proposal.
