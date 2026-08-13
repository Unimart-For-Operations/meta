# Contributing Guide

## Setting Up a Development Environment

Prerequisites:
1. **Go 1.21+** — via Nix (`nix profile install nixpkgs#go`) or [official installer](https://go.dev/doc/install)
2. **Make** — via system package manager
3. **Docker** — via Colima on macOS (`brew install colima && colima start`) or Docker Engine on Linux

Clone and build:

```bash
git clone git@github.com:Unimart-For-Operations/idpbuilder.git
cd idpbuilder
make build
./idpbuilder --help
```

Ensure your Docker daemon is running: `docker images` should not error out.

## Building from the Main Branch

1. Checkout the main branch: `git checkout main`
2. Build the binary: `make build` (compiles Go, generates CRDs, embeds manifests)
3. Verify: `./idpbuilder --help`

## Testing

```bash
# Run a local IDP cluster
./idpbuilder create
```

This creates a Kind cluster and deploys:

1. **[Kind](https://kind.sigs.k8s.io/)** cluster
2. **[ArgoCD](https://argo-cd.readthedocs.io/en/stable/)** — GitOps deployment engine
3. **[Gitea](https://about.gitea.com/)** — in-cluster Git server
4. **[ingress-nginx](https://kubernetes.github.io/ingress-nginx/)** — ingress controller

They are deployed as ArgoCD Applications with Gitea repositories as sources.

Accessible UIs:
- Gitea: https://gitea.cnoe.localtest.me:8443/explore/repos
- ArgoCD: https://argocd.cnoe.localtest.me:8443/applications

### Getting Credentials

```bash
idpbuilder get secrets

# Equivalent to:
kubectl -n argocd get secret argocd-initial-admin-secret
kubectl get secrets -n gitea gitea-admin-secret
kubectl get secrets -A -l cnoe.io/cli-secret=true
```

Verify all ArgoCD applications are synced: `kubectl get application -n argocd`

## Upgrading a Core Component

To upgrade ArgoCD, Gitea, or ingress-nginx:

1. Check the current version in `hack/<component>/kustomization.yaml`
2. Bump the version in the kustomization file
3. Review patches for any needed changes (new files, deletions, modifications)
4. Regenerate manifests: run the component's `hack/<component>/generate-manifests.sh`
5. Build: `make build`
6. Test locally and with `make e2e`
7. Update documentation

**Notes:**
- Some components may require updating Go library versions in `go.mod` (e.g., `code.gitea.io/sdk/gitea`)
- For ArgoCD, we use a separate project packaging a subset of the ArgoCD API: [cnoe-io/argocd-api](https://github.com/cnoe-io/argocd-api)

## Commits

This repository requires [DCO sign-off](https://developercertificate.org/) on all commits: `git commit -s`
