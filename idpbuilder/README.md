# IDP Builder

Private fork of [cnoe-io/idpbuilder](https://github.com/cnoe-io/idpbuilder) — a Kubernetes-based Internal Developer Platform builder.

## About

Spin up a complete internal developer platform using industry standard technologies like Kubernetes, ArgoCD, and Gitea with only Docker required as a dependency.

This fork is owned and customized by the `idpbuilder` GitHub org. It is **not** a GitHub fork (the fork relationship was intentionally broken to allow private hosting). Upstream changes are tracked via a dedicated `upstream` remote and cherry-picked as needed.

### Core Components

| Component | Purpose |
|-----------|---------|
| [Kind](https://kind.sigs.k8s.io/) | Local Kubernetes cluster |
| [ArgoCD](https://argo-cd.readthedocs.io/en/stable/) | GitOps deployment engine |
| [Gitea](https://about.gitea.com/) | In-cluster Git server |
| [ingress-nginx](https://kubernetes.github.io/ingress-nginx/) | Ingress controller |

### What This Fork Changes

This repo diverges from upstream to support our own infrastructure:
- Custom kustomize overlays and platform component versions
- Upstream management via `make fetch-upstream`, `make cherry-pick`, etc.
- Integration with the org-level docs sync pipeline and Obsidian vault
- Future: custom packages, cloud infrastructure targeting, private registry configs

## Building from Source

Prerequisites: Go 1.21+, Make, Docker (via Colima on macOS).

```bash
git clone git@github.com:Unimart-For-Operations/idpbuilder.git
cd idpbuilder
make build
./idpbuilder --help
```

## Usage

```bash
# Create a local IDP cluster
./idpbuilder create

# Access the UIs
# Gitea:  https://gitea.cnoe.localtest.me:8443
# ArgoCD: https://argocd.cnoe.localtest.me:8443

# Get credentials
./idpbuilder get secrets
```

## Upstream Management

This repo tracks `cnoe-io/idpbuilder` as a read-only upstream. See [docs/upstream-management.md](docs/upstream-management.md) for the full workflow.

```bash
make fetch-upstream      # Fetch latest upstream changes
make upstream-status     # Show ahead/behind counts
make log-upstream        # See new upstream commits
make cherry-pick COMMIT=abc123  # Cherry-pick a specific commit
```

## Documentation

- [Architecture](docs/Architecture/README.md) — CLI phase, controllers, build pipeline
- [Contributing](docs/Contributing/README.md) — Dev setup, build, test, component upgrades
- [Getting Started](docs/Getting-Started/README.md) — Prerequisites, first cluster, credentials
- [Upstream Management](docs/Reference/upstream-management.md) — Cherry-pick workflow, divergence tracking
- [Pluggable Packages](docs/Reference/pluggable-packages.md) — Custom ArgoCD application packaging
- [Private Registries](docs/Reference/private-registries.md) — Registry authentication configuration

## Development

See [docs/Contributing/README.md](docs/Contributing/README.md) for development setup, build process, and architecture details.

### Key Makefile Targets

```bash
make build               # Compile the binary
make test                # Run unit tests
make e2e                 # Run end-to-end tests
make embedded-resources  # Regenerate embedded manifests
make fetch-upstream      # Fetch latest from cnoe-io/idpbuilder
make upstream-status     # Show divergence from upstream
make sync-docs           # Sync docs to Obsidian vault
```

## Relationship to Upstream

| Aspect | This Repo | Upstream |
|--------|-----------|----------|
| GitHub | `idpbuilder/idpbuilder` (private) | `cnoe-io/idpbuilder` (public) |
| Go module path | `github.com/cnoe-io/idpbuilder` | `github.com/cnoe-io/idpbuilder` |
| Fork relationship | None (broken) | N/A |
| Remote name | `origin` | `upstream` |

The Go module path intentionally matches upstream to minimize import churn. This is a deliberate choice — changing it would require rewriting every import in the codebase.
