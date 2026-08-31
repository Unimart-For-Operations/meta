# idpbuilder Organization

This is the meta coordination repo and unified CLI (`unimart`) for the [Unimart-For-Operations](https://github.com/Unimart-For-Operations) GitHub organization. `cmdr` is tracked as a git submodule (`ignore = dirty`); `idpbuilder` is absorbed in-tree as a tracked directory.

## unimart CLI

The `unimart` binary is a Go CLI built from this repo. It is the primary interface after initial setup. Commands are organized into store aisles:

| Aisle | Domain | Key Commands |
|-------|--------|--------------|
| `deli` | Workstation config (Nix/HM) | `switch`, `doctor`, `bootstrap`, `hosts` |
| `freezer` | IDP platform lifecycle | `up`, `down`, `status`, `doctor`, `repos`, `repos-publish-to-gitea`, `docs` |
| `stockroom` | Cross-repo coordination | `check` |

Top-level commands (not under any aisle):

| Command | Purpose |
|---------|---------|
| `open` | Bring the full IDP platform online (prereqs, create, publish, browser) |
| `close` | Tear down the IDP platform (symmetric inverse of `open`) |
| `reload` | Reconcile platform changes without teardown (re-run create + re-publish repos) |
| `version` | Print version information |

Source layout:
- `main.go` — entry point
- `cmd/` — Cobra commands (root, aisle parents, subcommands)
- `internal/host/` — host auto-detection (scans cmdr meta.nix files)
- `internal/platform/` — platform detection, command execution utilities
- `internal/submodule/` — dynamic submodule discovery (parses `.gitmodules` at runtime)
- `internal/idp/` — in-process idpbuilder wiring (reuses its create/delete cobra commands)
- `internal/builder/` — container image build/load helpers (backstage-platform, Kind image loading)
- `internal/cluster/` — Kind cluster inspection (ArgoCD apps, secrets, Gitea token)
- `internal/colima/` — Colima VM lifecycle (start, stop, status, socket path)
- `internal/container/` — container image loading (Kind + Podman)
- `internal/gitea/` — Gitea API client (repos, SSH keys, auth)
- `internal/prereqs/` — prerequisite checks and installers (Go, Docker, Kind, kubectl, Colima, Podman)
- `internal/repos/` — org repo discovery and Gitea publish
- `internal/theme/` — theme loading, k9s skin and tmux status generation

idpbuilder is absorbed into this repo as a tracked directory (`idpbuilder/`, a
nested Go module `github.com/cnoe-io/idpbuilder`) and compiled directly into the
unimart binary via `replace github.com/cnoe-io/idpbuilder => ./idpbuilder` in
`go.mod`. There is no separate idpbuilder binary build step; `open`/`reload`/
`freezer up` run idpbuilder's create engine in-process (`internal/idp`), and
`close`/`freezer down` run its delete engine. Custom create flags map 1:1 to
idpbuilder flags via the curated flags on those commands, plus a `--` passthrough.

## Repositories

| Repo | Purpose | Language | Status |
|------|---------|----------|--------|
| [cmdr](cmdr/) | Nix flake + Home Manager workstation config | Nix | Active |
| [docs-service](docs-service/) | Phoenix microservice serving org docs from Gitea (deployed via `unimart freezer docs up`, not `unimart open`) | Elixir | Active |
| [cdc](unimart-employee-handbooks/cdc/) | Obsidian vault — synced doc mirrors + org knowledge base (hosted at `github.com/idpbuilder/cdc`) | Markdown | Active |
| [idpbuilder](idpbuilder/) | Kubernetes-based IDP builder, absorbed in-tree as a nested Go module (fork of cnoe-io/idpbuilder) | Go | Active |

## Distribution

unimart is distributed through three channels:

| Channel | Mechanism | When |
|---------|-----------|------|
| **Nix (primary)** | cmdr's flake imports `github:Unimart-For-Operations/meta` and includes `unimart` in `home.packages` via `04-modules/cli/graduated/unimart/` | Every `make switch` on every host |
| **CLI-first bootstrap** | `unimart deli bootstrap` / `go run . deli bootstrap` handles prerequisites and host apply | Fresh clones or first-time machine setup |
| **make install** | `go build` + symlink to `~/.local/bin/unimart` | Development iteration |

### Nix distribution details

The meta flake exposes `packages.<system>.unimart` via `buildGoModule`. cmdr's flake references it as a flake input (`meta.url = "git+ssh://git@github.com/Unimart-For-Operations/meta.git"` — uses git+ssh:// for private repos, not github: which would require a token). The unimart module at `cmdr/home/04-modules/cli/graduated/unimart/default.nix` pulls the package into `home.packages`. Any host with `features = ["cli" ...]` gets unimart automatically.

**Note on cli sub-features:** The `cli` feature is a convenience bundle that imports `cli-core`, `cli-languages`, `cli-containers`, `cli-devops`, and `cli-org`. The `_template/meta.nix` documents these sub-features, but they are not registered in the feature engine's `featureModules` map — only the bundle `cli` is usable as a top-level feature. To use individual components, compose them manually or use the bundle.

**Version pinning**: cmdr's `flake.lock` pins the meta commit. To bump: push meta changes, then run `nix flake update meta` in cmdr + `make switch`.

**Circular dependency note**: meta has cmdr as a submodule; cmdr references meta's flake by GitHub URL (not submodule path). Nix evaluates these independently — no real cycle. But version bumps need care: always push meta first, then update cmdr's lock.

## Conventions

- **Nix-first**: The user's system is Nix-managed (nix-darwin + home-manager via cmdr). All CLI tooling via Nix. Homebrew only for Colima on macOS.
- **DCO sign-off**: All repos require `git commit -s` for Developer Certificate of Origin.
- **Git submodules**: `cmdr` (`ignore = dirty`) and `unimart-employee-handbooks/cdc` (Obsidian vault, hosted at `github.com/idpbuilder/cdc`) are the active submodules; `idpbuilder` is a tracked directory. Use `unimart deli bootstrap` or `go run . deli bootstrap` for first-time init.
- **Commit style**: Conventional commits — `feat(scope):`, `fix:`, `docs:`, `refactor:`.
- **Makefile convention**: All repos use a consistent Makefile style — color output, `.DEFAULT_GOAL := help`, hand-crafted sectioned help, `@`-silenced commands, `## description` comments, `[pass]/[fail]/[warn]` status indicators. After `unimart` is installed, Make targets delegate to it.
- **Upstream sync**: cnoe-io/idpbuilder changes are cherry-picked, not merged. The `upstream` remote (`git@github.com:cnoe-io/idpbuilder.git`) and the `fetch-upstream` / `upstream-status` / `log-upstream` / `log-upstream-detail` / `diff-upstream` / `cherry-pick COMMIT=<sha>` targets live in the **meta** Makefile. idpbuilder's history in this repo is flat (one absorb commit), so an ancestry-based diff is impossible — compare via `HEAD:idpbuilder ↔ upstream/main` tree diff and apply with `git apply --directory=idpbuilder/`. Load the `upstream-mgmt` skill for the workflow.

## Development

```bash
go build ./...                 # Compile (fast check for errors)
go test ./...                  # Run all tests
make install                   # Build + symlink to ~/.local/bin/unimart
unimart deli doctor            # Validate local machine health
unimart deli bootstrap         # Full setup for a fresh clone or host
unimart stockroom check        # Run contract validation across org
```

**Go version:** See `go.mod` for the minimum version. The Nix devShell (`nix develop`) provides the correct Go toolchain.

**Adding a new command:** Follow the Cobra pattern in `cmd/`. Each aisle parent (`deli.go`, `freezer.go`, etc.) is a `cobra.Command` with no `RunE`. Subcommands register themselves via `init()` with `parentCmd.AddCommand()`. See `cmd/freezer_up.go` for the current pattern.

**Org directory resolution:** `resolveOrgDir()` in `cmd/root.go` detects the org root. It tries `--org-dir` flag → `UNIMART_ORG_DIR` env → walk up from CWD looking for `.gitmodules`.

## Infrastructure

### Physical Hosts

| Host | Platform | System | Username |
|------|----------|--------|----------|
| `apple-studio-m2-max` | macOS | `aarch64-darwin` | `cmdr` |
| `apple-macbook-m3-pro` | macOS | `aarch64-darwin` | `mortimera` |
| `cmdr` | Arch Linux | `x86_64-linux` | `cmdr` |
| `cachyos` | Arch Linux | `x86_64-linux` | `cmdr` |

**Known issue**: Both `arch/cmdr` and `arch/cachyos` have `username = "cmdr"`, so host auto-detection on Linux may select the wrong host based on directory enumeration order.

### Git Hooks

All hooks are Nix-managed via `cmdr/home/04-modules/cli/graduated/git/default.nix` and deployed to `~/.githooks/` via `unimart deli switch`.

| Hook | Gates | Speed |
|------|-------|-------|
| `pre-commit` | nix fmt, go fmt, go vet, gitleaks, theme lint (cmdr) | Fast |
| `commit-msg` | conventional commit, DCO sign-off | Instant |
| `pre-push` | go build, go test, nix flake check | Slow |

**Comment character**: `core.commentChar = ";"` — so `##` markdown headers in commit messages survive `commit.cleanup`.

## Key Paths

```
~/repos/Unimart-For-Operations/meta/    This directory (meta repo)
├── AGENTS.md                            This file
├── main.go                              CLI entry point
├── go.mod / go.sum                      Go module
├── flake.nix                            Nix packaging + devShell + overlay
├── cmd/                                 Cobra command tree
│   ├── root.go                          Root command, color helpers, org-dir resolution
│   ├── version.go                       Version command with ldflags injection
│   ├── open.go                          Top-level: open for business (7-step IDP startup)
│   ├── close.go                         Top-level: close up shop (symmetric inverse of open)
│   ├── reload.go                        Top-level: reconcile changes without teardown
│   ├── helpers.go                       Shared prereq/docker/build helpers
│   ├── deli.go                          Aisle parent
│   ├── switch.go                        deli switch (Nix apply)
│   ├── hosts.go                         deli hosts (list host configs)
│   ├── doctor.go                        deli doctor (prerequisite checks)
│   ├── bootstrap.go                     deli bootstrap (full setup)
│   ├── freezer.go                       Aisle parent (--container-runtime flag)
│   ├── freezer_up.go                    freezer up (4-step platform startup)
│   ├── freezer_down.go                  freezer down (cluster teardown)
│   ├── freezer_status.go               freezer status (cluster + ArgoCD + secrets)
│   ├── freezer_doctor.go               freezer doctor (prerequisite checks)
│   ├── freezer_bootstrap.go            freezer bootstrap (install prerequisites)
│   ├── freezer_repos.go                 freezer repos (list/clone/status)
│   ├── freezer_repos_publish.go        freezer repos publish-to-gitea
│   ├── freezer_docs.go                  freezer docs (up/status/down/open — docs microservice)
│   ├── stockroom.go                     Aisle parent
│   └── stockroom_check.go              stockroom check (CI contract validation)
├── internal/
│   ├── host/detect.go                   Host auto-detection (scans meta.nix)
│   ├── platform/detect.go              Platform detection, command execution
│   ├── platform/browser.go             Platform-aware OpenBrowser(url)
│   ├── submodule/submodule.go          Dynamic submodule discovery (.gitmodules parser)
│   ├── idp/idp.go                      In-process idpbuilder wiring (SetLogger, Run, Delete)
│   ├── builder/builder.go              BuildBackstagePlatform(), KindDeleteCluster(), LoadImageIntoKind()
│   ├── cluster/cluster.go              IsClusterRunning(), GetArgoApps(), GetSecrets(), GetGiteaAdminToken()
│   ├── colima/colima.go                Start(), Stop(), IsRunning(), EnsureDockerHost(), SocketPath()
│   ├── container/runtime.go            LoadImageIntoKind(), GatherImagesFromIdpbuilder()
│   ├── gitea/gitea.go                  RepoExists(), CreateRepo(), ListUserKeys(), auth helpers
│   ├── prereqs/                         Prerequisite checks and installers
│   │   ├── prereqs.go                  CheckResult, Platform(), CommandExists(), HasNix(), HasBrew()
│   │   ├── docker.go                   CheckDocker(), CheckColima(), InstallDocker()
│   │   ├── go.go                       CheckGo(), InstallGo()
│   │   ├── kind.go                     CheckKind(), InstallKind()
│   │   ├── kubectl.go                  CheckKubectl()
│   │   ├── podman.go                   CheckPodman(), InstallPodman()
│   │   └── workspace.go               CheckWorkspaceIdpbuilder()
│   ├── repos/repos.go                  ListRemote(), ListLocal(), Clone(), SetRemoteAndPush()
│   └── theme/                           Theme loading and config generation
│       ├── theme.go                     LoadFromOrg(), GenerateK9sSkin(), GenerateTmuxStatus()
│       └── fixtures/theme-sample.json  Test fixture
├── scripts/setup.sh                     make init bootstrap script (7 steps)
├── scripts/mirror-platform-repos.sh    3-way mirror sync (Gitea→mirrors→GitHub)
├── Makefile                             HAS_UNIMART delegation + shell fallbacks
├── packages/                            Custom ArgoCD Application YAMLs (passed to idpbuilder -p)
├── containers/                          Test infrastructure
├── repositories/                        Publish-to-Gitea symlink directory
├── mirrors/                             Pull-only clones of the 6 platform-generated
│                                       repos (backup sink; sync via scripts/mirror-platform-repos.sh)
├── cmdr/                                Nix workstation config (submodule)
├── docs-service/                        Phoenix docs microservice (submodule pending; see repositories/)
├── unimart-employee-handbooks/cdc/      Obsidian vault / synced doc mirrors (submodule)
└── idpbuilder/                          IDP builder (tracked directory, nested Go module)
```

>
> **Note**: `idpbuilder` is no longer a submodule — it was absorbed into this repo as a tracked directory (see above). `docs-service` is pending submodule registration. Active submodules in `.gitmodules`: `cmdr` and `unimart-employee-handbooks/cdc`.

## Working in This Directory

This is the org coordination hub. Component repos are managed as follows:

- **Submodules** (`cmdr`, `unimart-employee-handbooks/cdc`): Git submodules with `ignore = dirty` where applicable. Each has its own `AGENTS.md`.
- **Tracked directories** (`idpbuilder`): Absorbed in-tree, not submodules. Has its own `AGENTS.md`.
- **Standalone** (`docs-service`): Tracked directory, pending submodule conversion. Has its own `AGENTS.md`.

When working across repos, read the target repo's `AGENTS.md` first — it has repo-specific architecture, build/test commands, and conventions.

| Entry Point | Read This First | Then This |
|-------------|----------------|-----------|
| `meta/` (this dir) | This file | Target repo's `AGENTS.md` |
| `cmdr/` | `cmdr/AGENTS.md` | This file for org context |
| `idpbuilder/` | `idpbuilder/AGENTS.md` | This file for org context |

For on-demand deep knowledge, load these skills:
- `unimart-dev` — Build, test, add commands, dev workflow
- `nix-modules` — cmdr's tiered module system, host discovery, meta.nix
- `upstream-mgmt` — idpbuilder's cherry-pick workflow from cnoe-io
- `makefile-convention` — Shared Makefile style guide
