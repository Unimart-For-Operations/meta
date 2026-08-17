# Nix Sandboxes

Shared GitOps repository for ephemeral Nix sandbox workloads.

The `unimart open` flow seeds this repository from `packages/nix-sandboxes/`.
The Scaffolder `Nix Sandbox` template then appends a new `sandboxes/<name>/`
directory via its `gitea:commit` action. The `nix-sandboxes` ArgoCD
Application syncs this repo with `directory.recurse: true`, so every
directory under `sandboxes/` is deployed automatically.

Structure:

```
sandboxes/
  <name>/          # one directory per sandbox (created by the template)
    deployment.yaml   # the nix container (stdin/tty enabled)
    terminal.yaml     # in-browser terminal (ttyd) exposing the container
    rbac.yaml         # ServiceAccount + pod/exec permissions for the terminal
    catalog-info.yaml # Backstage Component entity (excluded from ArgoCD sync)
```

Note: this repository only contains Kubernetes manifests. `catalog-info.yaml`
is a Backstage entity, not a Kubernetes resource, so the `nix-sandboxes`
ArgoCD Application excludes it via `spec.source.directory.exclude`. Backstage
picks it up through the Scaffolder template's `catalog:register` step against
the committed `sandboxes/<name>/catalog-info.yaml`.

To add a sandbox by hand, create a directory under `sandboxes/<name>/`
with plain manifests and commit it to `main`.