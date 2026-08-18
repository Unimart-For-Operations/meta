# Nix Sandbox Live Terminal — Technical Architecture

The nix-sandbox feature provisions ephemeral, browser-accessible Nix sandboxes
on the IDP platform. Each sandbox is a long-running Kubernetes pod running the
`nixos/nix` image, and the primary access path is a **live terminal** — an
in-browser shell rendered by [ttyd](https://github.com/tsl0922/ttyd) that
`kubectl exec`s into the sandbox pod.

This document is the authoritative technical reference for how that terminal is
built, deployed, secured, and accessed. It is written for operators who need to
debug it and for developers extending the scaffolder template.

---

## 1. High-level data flow

```
Browser ──HTTPS──► Ingress ──► Service ──► ttyd container
                                            │
                                            │ kubectl exec (pods/exec)
                                            ▼
                                        nix container
                                        (sandbox pod)
```

1. A user opens `https://<name>-terminal.<host>:8443/` in a browser.
2. The Ingress (nginx) routes the host to the terminal Service.
3. The Service load-balances to the ttyd container in the terminal pod.
4. ttyd serves its web UI (WebSocket-based) on port `7681` and, at connection
   time, launches the command configured in its entrypoint.
5. The entrypoint runs `kubectl exec -it deploy/<name> -c nix -- /bin/sh`,
   which attaches a shell **inside the sandbox pod's `nix` container**.
6. The ServiceAccount bound to the terminal pod has a Role granting
   `pods/exec` on the `sandboxes` namespace, so the exec is authorized.

The "live terminal" is therefore four cooperating layers:

| Layer | Where it lives | What it does |
|-------|----------------|--------------|
| Image | `containers/terminal/` | ttyd binary + kubectl binary + entrypoint |
| Wiring | `internal/builder/`, `cmd/helpers.go` | builds the image, loads it into Kind |
| Manifests | `packages/nix-sandboxes/sandboxes/<name>/terminal.yaml` | Deployment, Service, Ingress, RBAC |
| Template | `packages/backstage/scaffolder-templates/nix-sandbox/` | how new sandboxes get these manifests |

---

## 2. The terminal image

Source: `containers/terminal/Dockerfile` and `containers/terminal/terminal-entrypoint.sh`.

### 2.1 Base image

`FROM node:24-trixie-slim` — a small Debian trixie-based image that already
ships `ca-certificates` and `curl` for the download steps.

### 2.2 Binaries

Neither `ttyd` nor `kubectl` is packaged in the Debian repositories, so both are
fetched as static binaries at build time:

- **ttyd 1.7.7** (`ARG TTYD_VERSION=1.7.7`): downloaded from the upstream
  GitHub release as `ttyd.x86_64`. ttyd is a C web server that exposes a
  terminal in the browser over a WebSocket (the `xterm.js` client is bundled
  into the binary, so there is no frontend code in this repo).
- **kubectl v1.33.1** (`ARG KUBECTL_VERSION=v1.33.1`): downloaded from
  `dl.k8s.io`. The version pins to the cluster's Kubernetes version
  (the localdev cluster runs v1.33.1).

Builds use BuildKit cache mounts for apt (`--mount=type=cache`) to keep
rebuilds fast during `unimart reload` iterations.

### 2.3 Entrypoint

`containers/terminal/terminal-entrypoint.sh`:

```sh
#!/usr/bin/env sh
set -e

SANDBOX_NAME="${SANDBOX_NAME:-}"
SANDBOX_CONTAINER="${SANDBOX_CONTAINER:-nix}"
SHELL_CMD="${SHELL_CMD:-/bin/sh}"

if [ -n "${SANDBOX_NAME}" ]; then
  exec ttyd --writable --port 7681 \
    kubectl exec -it "deploy/${SANDBOX_NAME}" -c "${SANDBOX_CONTAINER}" -- "${SHELL_CMD}"
fi

exec ttyd --writable --port 7681 /bin/sh
```

Key behaviors:

- `SANDBOX_NAME` is set from the terminal Deployment's environment and names
  the sandbox Deployment to exec into (`deploy/<name>`).
- `SANDBOX_CONTAINER` defaults to `nix`, the primary container in the sandbox
  pod.
- `--writable` lets the browser terminal run interactive commands (not just
  read-only PTYs). This is the reason the entrypoint uses `--writable` and not
  the default read-only mode.
- The `exec` at the front replaces the shell: ttyd **is** the process that owns
  port 7681 and the terminal session.
- If `SANDBOX_NAME` is empty, ttyd falls back to a local `/bin/sh` (useful for
  standalone debugging of the image without a cluster).

`kubectl` inside the container talks to the cluster's API server. It relies on
the pod's **ServiceAccount token** being mounted at the standard
`/var/run/secrets/kubernetes.io/serviceaccount` path — no `kubeconfig` is
shipped in the image.

---

## 3. Build and load wiring (unimart CLI)

The terminal image is not a platform service; it is a **custom image built
during `unimart open` / `unimart reload`**, alongside `backstage-platform`.
The sandbox runtime image (`sandbox-tty`, §4.1) is built in the same pass.

### 3.1 `internal/builder/builder.go`

`BuildTerminal(orgDir string, verbose bool) error` (builder.go:85) mirrors
`BuildBackstagePlatform`:

- Resolves `containers/terminal` under the org dir; errors if missing.
- Runs `docker build -t terminal:latest -f Dockerfile .` with the build
  context set to the terminal dir.
- Streams output only when the CLI is verbose.

`BuildSandbox(orgDir string, verbose bool) error` (builder.go:120) builds the
sandbox runtime image from `containers/sandbox`:

- `containers/sandbox/flake.nix` defines a pinned `ttyProfile` (`pkgs.buildEnv`
  of the bare-minimum TTY toolset: zsh, tmux, nvim, yazi, starship, fzf,
  zoxide, bat, eza, git, ripgrep, fd, openssh, opencode).
- The Dockerfile builds that profile inside a `nixos/nix` build stage and
  copies the store + profile into the final image, so sandbox pods start with
  the toolset pre-baked (no per-pod install).
- Minimal rc files (`zshrc`, `tmux.conf`, `starship.toml`) are baked in so the
  sandbox shell feels like a real tty-engineer workstation.
- `BuildSandbox` auto-copies the user's AstroNvim config
  (`cmdr/home/04-modules/tui/graduated/nvim/nvim-astro`) into the build
  context before `docker build`, so the sandbox's `nvim` launches nvim-astro
  (plugins lazy-fetch on first run) and its tmux keybindings match the cmdr
  tmux module (prefix `C-Space`, vim pane nav, `|`/`-` splits). The copy is
  skipped with a warning when cmdr is absent.

### 3.2 `cmd/helpers.go`

- `buildCustomImages` (helpers.go:201) now builds both images: it warns and
  skips `terminal` if `containers/terminal` is absent (backwards-compatible
  with orgs that predate the terminal).
- `loadCustomImages` (helpers.go:235) loads `["backstage-platform:latest",
  "terminal:latest", "sandbox-tty:latest"]` into the Kind cluster via
  `kind load docker-image`.

This is why, after a fresh `unimart open`, `docker exec localdev-control-plane
crictl images` shows `docker.io/library/terminal` — the image is present in the
Kind node and the terminal Deployment can pull it with `imagePullPolicy:
IfNotPresent`.

---

## 4. Sandbox manifests

Two sources of truth for the same manifests:

1. **Running example**: `packages/nix-sandboxes/sandboxes/example/`
2. **Scaffold source**: `packages/backstage/scaffolder-templates/nix-sandbox/skeleton/`

Both are identical in structure; the skeleton uses Backstage `${{ values.* }}`
templating. The running example is the "reference implementation" that the
scaffolder produces.

### 4.1 The sandbox Deployment

`packages/nix-sandboxes/sandboxes/example/deployment.yaml`:

- `Deployment example-sandbox` in namespace `sandboxes`.
- One container, `nix`, running `docker.io/library/sandbox-tty:latest` with
  `imagePullPolicy: IfNotPresent`. The image is built from `containers/sandbox`
  during `unimart open`/`reload` and carries the bare-minimum TTY toolset
  (zsh, tmux, nvim, yazi, starship, fzf, zoxide, bat, eza, git, opencode), the
  user's nvim-astro editor config, and tmux keybindings matching cmdr.
- Command keeps the pod alive: `echo "Nix sandbox 'example' ready"; sleep
  infinity`.
- Resource requests/limits: `100m/256Mi` request, `1 CPU/1Gi` limit.

The sandbox pod itself is intentionally inert — the terminal sidecar is the
access point.

### 4.2 The terminal Deployment + Service + Ingress

`packages/nix-sandboxes/sandboxes/example/terminal.yaml`:

**Deployment** `example-sandbox-terminal`:

- Container `terminal` uses image `docker.io/library/terminal:latest` with
  `imagePullPolicy: IfNotPresent`.
- Environment:
  - `SANDBOX_NAME=example-sandbox` → tells the entrypoint which Deployment to
    exec into.
  - `SANDBOX_CONTAINER=nix` → which container carries the shell.
  - `SHELL_CMD=/bin/zsh` → the interactive shell launched by ttyd (the baked
    profile's zsh), so the browser terminal is a proper TTY environment.
- Command `["/terminal-entrypoint.sh"]`.
- Exposes container port `7681` (ttyd's listen port).
- Carries the label `backstage.io/kubernetes-id: example-sandbox`, which ties
  the terminal pod to the sandbox's Backstage catalog entity for the
  kubernetes plugin.

**Service** `example-sandbox-terminal`:

- Selects `app: example-sandbox-terminal`.
- Maps service port `80` → target port `7681` (named `http`).

**Ingress** `example-sandbox-terminal`:

- `ingressClassName: nginx`.
- Rule host: `example-sandbox-terminal.cnoe.localtest.me`, path `/` with
  `PathType: Prefix`.
- TLS for that host using the shared `idpbuilder-cert` secret.

### 4.3 RBAC (the exec authorization)

`packages/nix-sandboxes/sandboxes/example/rbac.yaml`:

- `ServiceAccount example-sandbox`.
- `Role example-sandbox-exec` in namespace `sandboxes`:
  - `apps/deployments`: `get`, `list`
  - `pods`: `get`, `list`
  - `pods/exec`: `create`
- `RoleBinding example-sandbox-exec` binds the Role to the ServiceAccount.

Both the sandbox Deployment and the terminal Deployment declare
`serviceAccountName: example-sandbox`. This is the critical security property:

> The terminal's `kubectl exec` is authorized because the pod's ServiceAccount
> has a Role granting `pods/exec create` scoped to the `sandboxes` namespace.
> The Role only grants `get`/`list` on deployments/pods plus exec — it cannot
> modify any workload.

### 4.4 The catalog entity (not deployed)

`catalog-info.yaml` per sandbox is a Backstage Component entity, **not** a
Kubernetes manifest. The `nix-sandboxes` ArgoCD Application excludes it from
sync:

```yaml
directory:
  recurse: true
  exclude: "sandboxes/**/catalog-info.yaml"
```

It carries `backstage.io/kubernetes-id` and a link to the terminal URL, so the
sandbox shows up in the Backstage catalog with the terminal one click away.

---

## 5. The scaffolder template

Source: `packages/backstage/scaffolder-templates/nix-sandbox/template.yaml`.

When a user scaffolds a new sandbox in Backstage, three steps run:

1. **`fetch:template`** — renders `./skeleton` with the user's values
   (`name`, `description`, `owner`, `image`, `terminalHost`) into the working
   directory. `copyWithoutTemplating` protects binary assets.
2. **`gitea:commit`** — appends the rendered manifests to the shared
   `cnoe://nix-sandboxes` repo under `sandboxes/<name>/` via the custom
   `gitea:commit` action, committing to `main`. The action emits
   `catalogInfoUrl` as output.
3. **`catalog:register`** — registers the committed `catalog-info.yaml` as a
   Component in Backstage, using the `catalogInfoUrl` from step 2 (marked
   `optional: true`).

The template's **output links** give the user the exact terminal URL after a
successful scaffold:

```
https://${{ parameters.name }}-terminal.${{ parameters.terminalHost }}
```

The mirror that Backstage actually reads is `repositories/backstage-templates/
nix-sandbox/` in this repo (published to gitea; gitignored in meta so it does
not double-track). The `terminalHost` parameter defaults to
`cnoe.localtest.me`, so the advertised URL is
`https://<name>-terminal.cnoe.localtest.me:8443`.

> **Note:** both the scaffold output link and the per-sandbox `catalog-info.yaml`
> link now include the `:8443` port. The IDP's nginx ingress is only reachable
> through the host's mapped port `8443` (see §6.3), so without that port the
> browser hits `127.0.0.1:443` where nothing listens.

---

## 6. Access path, step by step

### 6.1 DNS resolution

`*.cnoe.localtest.me` hosts are resolved to loopback by the host's local DNS.
On the NixOS workstation (`strix-nix`) this is now done with **NetworkManager's
dnsmasq plugin**:

- `networking.networkmanager.dns = "dnsmasq"` — NM spawns its own dnsmasq and
  writes `nameserver 127.0.0.1` to `/etc/resolv.conf`.
- `/etc/NetworkManager/dnsmasq.d/idp-localtest.conf` contains
  `address=/cnoe.localtest.me/127.0.0.1` — a wildcard that resolves **any**
  `*.cnoe.localtest.me` host to `127.0.0.1`.

This replaces the old per-host `/etc/hosts` pins and means newly scaffolded
sandboxes resolve automatically, with no hosts file edits.

### 6.2 Cluster routing

- The Kind node (`localdev-control-plane`) publishes the nginx ingress
  controller's HTTPS listener on host port `8443`
  (`127.0.0.1:8443->443/tcp`).
- The nginx controller matches the `Host: example-sandbox-terminal.cnoe.localtest.me`
  header to the sandbox terminal Ingress and forwards to the Service, which
  forwards to ttyd on `7681`.

### 6.3 Verifying end-to-end

```bash
# DNS: the wildcard resolves any sandbox host to loopback
getent hosts example-sandbox-terminal.cnoe.localtest.me
# → 127.0.0.1

# Routing + TLS: ingress answers through host port 8443
curl -sk -o /dev/null -w '%{http_code}\n' \
  -H 'Host: example-sandbox-terminal.cnoe.localtest.me' \
  https://127.0.0.1:8443/
# → 200  (ttyd HTML)

# Full URL in a browser:
#   https://example-sandbox-terminal.cnoe.localtest.me:8443/
```

### 6.4 Direct exec (CLI alternative)

The same shell is reachable without the browser:

```bash
kubectl exec -it -n sandboxes deploy/example-sandbox -c nix -- /bin/sh
```

This does not require the terminal Deployment at all; the sandbox pod is
directly exec-able because the same ServiceAccount Role governs API access
(operator credentials, not the pod's, are used here).

---

## 7. Debugging guide

### 7.1 "Browser terminal shows nothing / connection reset"

Check the chain outward-to-in:

```bash
# 1. Pods up?
kubectl get pods -n sandboxes
# 2. Service has endpoints?
kubectl get endpoints -n sandboxes example-sandbox-terminal
# 3. Ingress routed?
kubectl get ingress -n sandboxes
# 4. ttyd actually listening / entrypoint exec working?
kubectl logs -n sandboxes deploy/example-sandbox-terminal -c terminal --tail=50
```

A healthy log shows ttyd's "start server at ... port 7681" line. If ttyd
crashes on startup, the usual cause is a **kubectl exec failure**: check the
`pods/exec` RBAC (§4.3) and that `SANDBOX_NAME` matches an existing sandbox
Deployment in the same namespace.

### 7.2 "Permission denied when exec'ing"

- The terminal pod's ServiceAccount (`example-sandbox`) must have the
  `example-sandbox-exec` RoleBinding.
- The Role needs `pods/exec` `create` — `pods` `get/list` alone is not enough
  for exec.
- Exec runs in the **same namespace** as the sandbox (`sandboxes`); a
  namespace-scoped Role cannot cross namespaces.

### 7.3 "I changed the template but the catalog still shows the old one"

- Backstage reads the published `backstage-templates` repo, not
  `packages/backstage/` directly. The mirror is
  `repositories/backstage-templates/nix-sandbox/` and must be re-synced
  (`rm -rf` + `cp -r` from `packages/...`) and committed/pushed there.
- The catalog location in `packages/backstage/values.yaml` must point at the
  published repo's raw URL; after changing it, the backstage app-config
  configmap must regenerate (the app must re-render — see the vendored chart
  note in §8).

### 7.4 "New sandbox host doesn't resolve"

- Confirm the dnsmasq wildcard is active: `cat /etc/resolv.conf` should show
  `nameserver 127.0.0.1`; `getent hosts <any>.cnoe.localtest.me` should return
  `127.0.0.1`.
- If it returns nothing, the config switch hasn't been applied yet
  (`make switch` in cmdr, requires sudo).

---

## 8. Related infrastructure quirks (context for operators)

### 8.1 Vendored backstage chart

The backstage ArgoCD Application renders its Helm chart in-cluster. The chart's
dependency (`backstage` 2.8.2) is served from `https://backstage.github.io/
charts`, which the air-gapped cluster's repo-server cannot reach. The chart is
therefore **vendored** as an unpacked subchart at
`packages/backstage/charts/backstage/` (with its own vendored `common` and
`postgresql` deps), so `helm template` renders fully offline. Regenerating the
backstage app-config (e.g. to add a new catalog location) depends on this
vendored chart being in the seeded gitea repo.

### 8.2 Repos are private; tokens are injected

All org repos are private. Backstage's `UrlReader` authenticates against Gitea
using the `GITEA_TOKEN` mounted from the `backstage-secrets` secret, which is
why catalog `url` locations succeed despite the repos being private.

### 8.3 Stray-file hygiene

The scaffolder template root must only contain `template.yaml` and `skeleton/`.
Earlier syncs accidentally copied skeleton files into the template root; those
are dead weight and confuse consumers reading the repo raw. Keep the template
root clean.

---

## 9. Reference index

| Concern | File |
|---------|------|
| Terminal image | `containers/terminal/Dockerfile`, `terminal-entrypoint.sh` |
| Sandbox runtime image | `containers/sandbox/Dockerfile`, `flake.nix`, `flake.lock`, `rc/` |
| Image build (CLI) | `internal/builder/builder.go` (`BuildTerminal`, `BuildSandbox`) |
| Image build/load orchestration | `cmd/helpers.go` (`buildCustomImages`, `loadCustomImages`) |
| Running example sandbox | `packages/nix-sandboxes/sandboxes/example/*.yaml` |
| ArgoCD app for sandboxes | `packages/nix-sandboxes.yaml` |
| Scaffolder template | `packages/backstage/scaffolder-templates/nix-sandbox/template.yaml` + `skeleton/` |
| Published template mirror | `repositories/backstage-templates/nix-sandbox/` (gitignored in meta) |
| Catalog location registration | `packages/backstage/values.yaml` |
| Vendored chart (offline render) | `packages/backstage/charts/backstage/` |
| Wildcard DNS (host) | cmdr `home/02-hosts/nixos/strix-nix/system.nix` |
