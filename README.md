# unimart

Your one-stop shop for the [Unimart-For-Operations](https://github.com/Unimart-For-Operations) organization. A unified CLI that manages workstation configuration, IDP platform lifecycle, documentation, and cross-repo coordination.

See [TOOLING.md](TOOLING.md) for the boundary between `unimart`, Make, CI, and shell completion. See [PROVISIONING.md](PROVISIONING.md) for the physical-host provisioning contract.

## Quick Start

```bash
git clone --recurse-submodules git@github.com:Unimart-For-Operations/meta.git
cd meta
make init
```

`make init` handles everything: submodule init, Nix/Homebrew prerequisites, host detection, configuration apply, and installing the `unimart` binary to `~/.local/bin/`.

After setup, reload your shell and use `unimart` directly:

```bash
exec zsh
unimart --help
```

## Open for Business

One command to bring the entire org's IDP online locally:

```
unimart open                 Full startup: prereqs → build → create → publish → browser
unimart open --skip-build    Skip the idpbuilder build step
unimart open --recreate      Tear down existing cluster first
unimart open --no-browser    Don't auto-open ArgoCD dashboard
```

This runs a 6-step sequence: check prerequisites, start the container runtime (Colima on macOS), build idpbuilder, create the IDP platform (ArgoCD + Gitea + nginx on Kind), publish repos from `repositories/` to in-cluster Gitea by default, and open the ArgoCD dashboard.

### Default Publish Directory

`unimart open` and `unimart freezer repos publish-to-gitea` look for publishable git repos in:

```bash
./repositories/
```

Each direct child folder with a `.git` entry is pushed to local Gitea under the default owner.

If `repositories/` is absent, unimart falls back to legacy org-root scanning.

### Close Up Shop

```
unimart close                Tear down the IDP platform
unimart close --stop-colima  Also stop the Colima VM (macOS)
unimart close --yes          Skip confirmation prompt
```

## Aisles

Commands are organized into store aisles:

### deli — Workstation configuration

Custom-sliced Nix/Home Manager host configs.

```
unimart deli switch          Apply Nix configuration for the current host
unimart deli doctor          Check system prerequisites and environment health
unimart deli bootstrap       Full setup: submodules, prerequisites, host detection, apply
unimart deli hosts           List available host configurations
unimart deli plan [host]     Show a non-destructive physical-host onboarding plan
```

### freezer — IDP platform lifecycle

Spin up and cool down IDP clusters.

```
unimart freezer up           Start the full IDP platform (prereqs → Colima → build → create)
unimart freezer down         Tear down the IDP platform cluster
unimart freezer status       Show cluster state, ArgoCD apps, and secrets
unimart freezer build        Build idpbuilder from source
unimart freezer doctor       Check IDP platform prerequisites
unimart freezer bootstrap    Install missing IDP platform prerequisites
unimart freezer repos        Manage repos (list, clone, status)
unimart freezer repos publish-to-gitea   Publish repos from repositories/ into in-cluster Gitea
unimart freezer config       Generate or show IDP platform configuration
unimart freezer theme        Load org theme and generate k9s/tmux configs
```

### newsstand — Documentation

Pick up the latest publications.

```
unimart newsstand sync       Sync documentation across repos to the cdc handbook
```

### stockroom — Cross-repo coordination

Back-of-house inventory management.

```
unimart stockroom status     Show submodule state (dirty, ahead/behind, current ref)
unimart stockroom drift      Check submodule pointers vs remote HEAD
unimart stockroom update     Pull latest main for all submodules
unimart stockroom sync       Push all org repos to their remotes
unimart stockroom check      Validate cross-repo contracts (CI)
```

## Repositories

| Submodule | Purpose | Language |
|-----------|---------|----------|
| [cmdr](cmdr/) | Nix flake + Home Manager workstation config | Nix |
| [idpbuilder](idpbuilder/) | Kubernetes-based IDP builder (private fork of cnoe-io/idpbuilder) | Go |
| [idpctl](idpctl/) | CLI lifecycle tool for idpbuilder (deprecated — absorbed into unimart) | Go |
| [docs](docs/) | Documentation aggregation hub (transitional — being replaced by cdc) | Shell/Markdown |
| [cdc](unimart-employee-handbooks/cdc/) | Obsidian vault — synced doc mirrors + org knowledge base | Markdown |

## Install Methods

| Method | Command | Notes |
|--------|---------|-------|
| `make init` | Full onboarding (recommended) | Installs unimart as step 6/7 |
| `make install` | Build + symlink to `~/.local/bin/` | Requires Go |
| `nix build .` | Hermetic Nix build | Requires Nix |
| `nix profile install .` | Install to Nix profile | Requires Nix |

## Make Targets

After `unimart` is installed, most Make targets delegate to it. Without `unimart`, they fall back to shell implementations.

```
make help         Show available commands
make init         Full onboarding: submodules → prereqs → config → unimart → verify
make bootstrap    Initialize submodules, verify remotes, install hooks
make build        Build the unimart binary
make install      Build and symlink to ~/.local/bin/unimart
make completion   Generate shell completion for unimart
make status       Show submodule state
make drift        Check submodule pointers vs remote HEAD
make update       Pull latest main for all submodules
make sync-docs    Trigger docs sync pipeline to cdc
make ci/check     Validate cross-repo contracts
make tag          Create org snapshot tag (VERSION=v0.x.0)
```

## Conventions

- **Private repos only** — all org repos are private on GitHub
- **Nix-first** — system managed via nix-darwin + home-manager (cmdr)
- **DCO sign-off** — all commits require `git commit -s`
- **Conventional commits** — `feat(scope):`, `fix:`, `docs:`, `refactor:`
- **Makefile convention** — consistent style across all repos (color output, hand-crafted help, status indicators)

## Tagging Strategy

Individual repos maintain their own semver tags. This meta repo uses **org snapshot tags** (`v0.1.0`, `v0.2.0`, ...) representing known-good combinations of submodule pointers. Create one with:

```bash
make tag VERSION=v0.1.0
```
