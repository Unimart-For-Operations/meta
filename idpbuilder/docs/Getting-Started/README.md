# Getting Started

## Prerequisites

- **Go 1.21+** — for building from source
- **Make** — for build automation
- **Docker** — via Colima on macOS or Docker Engine on Linux
- **Minimum hardware**: 4 CPU cores and 4 GiB of RAM

See [Contributing/README.md](../Contributing/README.md) for detailed install instructions.

## Quick Start

```bash
# Clone and build
git clone git@github.com:Unimart-For-Operations/idpbuilder.git
cd idpbuilder
make build

# Create a local IDP cluster
./idpbuilder create

# Get credentials
./idpbuilder get secrets
```

## What Gets Deployed

`idpbuilder create` provisions a [Kind](https://kind.sigs.k8s.io/) cluster with:

| Component | Purpose | URL |
|-----------|---------|-----|
| [ArgoCD](https://argo-cd.readthedocs.io/en/stable/) | GitOps deployment engine | https://argocd.cnoe.localtest.me:8443 |
| [Gitea](https://about.gitea.com/) | In-cluster Git server | https://gitea.cnoe.localtest.me:8443 |
| [ingress-nginx](https://kubernetes.github.io/ingress-nginx/) | Ingress controller | N/A (internal) |

Components are deployed as ArgoCD Applications with Gitea repositories as sources.

## Verify

```bash
# Check all ArgoCD applications are synced
kubectl get application -n argocd

# View Gitea repositories
open https://gitea.cnoe.localtest.me:8443/explore/repos

# View ArgoCD dashboard
open https://argocd.cnoe.localtest.me:8443/applications
```

## Teardown

```bash
./idpbuilder delete
```

Or use [idpctl](https://github.com/Unimart-For-Operations/idpctl) for a managed lifecycle:

```bash
idpctl down
idpctl down --stop-colima  # also stop the Colima VM (macOS)
```
