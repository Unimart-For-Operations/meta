# personal-blog

A Phoenix microservice that serves as a multi-purpose personal hub: portfolio, projects, and articles. Content is discovered and fetched from Gitea repositories using pattern-based discovery (portfolio/, articles/, projects/, README.md).

## Overview

**personal-blog** is a stateless, read-only Phoenix microservice designed for showcasing personal work across three dimensions:

- **Portfolio** — Professional experience, accomplishments, resume-like entries from `portfolio/` directories
- **Projects** — Project showcases and technical deep-dives from `projects/` directories
- **Articles** — Blog posts, technical writings, and thought pieces from `articles/` directories

Content is pulled from Gitea in real-time (with 60s ETS-based TTL caching), rendered as HTML via Earmark, and served over LiveView for a responsive, dynamic UI. No database required; all state is transient and cache-backed.

## Architecture

### Directory Structure

```
personal-blog/
├── lib/
│   ├── personal_blog/
│   │   ├── application.ex          # OTP app + supervisor tree
│   │   ├── cache.ex                # ETS-backed TTL cache
│   │   └── gitea.ex                # Read-only Gitea API client
│   └── personal_blog_web/
│       ├── endpoint.ex             # Bandit HTTP adapter
│       ├── router.ex               # Routes (/portfolio, /articles, etc.)
│       ├── doc_renderer.ex         # Earmark HTML rendering + heading anchors
│       ├── controllers/            # HTTP controllers (health checks)
│       ├── live/blog_live/         # LiveView components (home, articles, portfolio, projects)
│       └── components/layouts.ex   # Tailwind CSS layout components
├── config/
│   ├── config.exs                  # Compile-time base config
│   ├── dev.exs                     # Development overrides
│   ├── prod.exs                    # Production stub
│   ├── runtime.exs                 # Runtime env-driven config (prod only)
│   └── test.exs                    # Test config
├── assets/
│   ├── css/app.css                 # Tailwind CSS input
│   └── js/app.js                   # LiveSocket entry point
├── priv/static/assets/             # Compiled CSS/JS (generated)
├── manifests/
│   ├── deployment.yaml             # Pod spec + env vars + secrets
│   ├── service.yaml                # ClusterIP 4000
│   ├── ingress.yaml                # nginx ingress → personal-blog.cnoe.localtest.me:8443
│   ├── configmap.yaml              # Gitea URL, cache TTL, host config
│   └── serviceaccount.yaml         # K8s ServiceAccount
├── argocd/application.yaml         # ArgoCD Application (manual deploy)
├── Dockerfile                      # Multi-stage build → release binary
├── mix.exs                         # Project definition + deps + aliases
├── flake.nix                       # Nix devShell (elixir, erlang, elixir-ls)
├── Makefile                        # Org-wide convention (dev tasks + unimart delegates)
└── catalog-info.yaml               # Backstage component registration
```

### Key Modules

| Module | Purpose | Key Functions |
|--------|---------|---------------|
| `PersonalBlog.Application` | OTP supervisor tree startup | Starts Cache, Telemetry, Bandit HTTP server |
| `PersonalBlog.Cache` | ETS-backed TTL cache (60s default) | `get(key)`, `put(key, value)`, `flush()` |
| `PersonalBlog.Gitea` | Read-only Gitea API client | `list_repos()`, `list_articles()`, `get_article(repo, path)`, etc. |
| `PersonalBlogWeb.Endpoint` | Bandit HTTP server + routing | HTTP listener on port 4000 |
| `PersonalBlogWeb.Router` | Phoenix router | Routes: `/`, `/articles`, `/articles/:repo/:path`, `/portfolio`, `/projects`, `/healthz` |
| `PersonalBlogWeb.BlogLive.Home` | Home page LiveView | Displays bio from README.md + links to sections |
| `PersonalBlogWeb.BlogLive.Articles` | Articles list LiveView | Fetches & displays articles from `articles/` directories |
| `PersonalBlogWeb.BlogLive.Article` | Article detail LiveView | Fetches & renders single article markdown |
| `PersonalBlogWeb.BlogLive.Portfolio` | Portfolio entries list LiveView | Similar pattern for portfolio/ |
| `PersonalBlogWeb.BlogLive.PortfolioEntry` | Portfolio entry detail LiveView | Render single portfolio entry |
| `PersonalBlogWeb.BlogLive.Projects` | Projects list LiveView | Similar pattern for projects/ |
| `PersonalBlogWeb.BlogLive.Project` | Project detail LiveView | Render single project |
| `PersonalBlogWeb.DocRenderer` | Markdown → HTML renderer | Uses Earmark; adds heading anchors |

### Data Flow

```
User Request
    ↓
Phoenix Router (/articles, /portfolio, etc.)
    ↓
LiveView Component (Articles.ex, Portfolio.ex, etc.)
    ↓
PersonalBlog.Gitea (API client)
    ↓
PersonalBlog.Cache (ETS TTL cache)
    ↓
Gitea API (http://my-gitea-http.gitea.svc.cluster.local:3000)
    ↓
Rendered HTML (Earmark + DocRenderer + Tailwind CSS)
```

## Dependencies

- **Phoenix 1.7.14** — Web framework
- **Bandit 1.5** — HTTP server adapter (native BEAM)
- **phoenix_live_view 1.0** — Real-time UI
- **Req 0.5** — HTTP client (Gitea API calls)
- **Earmark 1.4** — Markdown → HTML
- **Jason 1.4** — JSON codec
- **Tailwind 0.2** — Standalone CSS compiler (hex package, no npm)

**No Ecto** — This is a read-only, stateless service; no database.

## Development

### Prerequisites

- Nix with `nix develop` OR Elixir 1.17 + Erlang 27 + OTP 27
- Docker (for image builds)

### Setup

```bash
nix develop
mix setup
mix phx.server
```

Visit http://localhost:4000.

### Development Workflow

```bash
# Start development server with hot reload
make server

# Run tests
make test

# Format code
make fmt

# Lint
make lint

# Full CI check (format, lint, compile, test)
make ci
```

All commands run via `make` delegate to Mix; see `Makefile` for targets.

### Environment Variables (Dev)

Not required locally; defaults are sensible:
- `PORT` — HTTP port (default: 4000)
- `GITEA_URL` — Gitea base URL (default: http://my-gitea-http.gitea.svc.cluster.local:3000)
- `GITEA_USER` — User or org whose repos to scan (default: empty)
- `BLOG_CACHE_TTL` — Cache TTL in seconds (default: 60)

## Deployment

### Build

```bash
make build
```

Produces `personal-blog:latest` Docker image (multi-stage build, ~200MB).

### Deploy to Kubernetes

The service is **deliberately not auto-deployed** in `unimart open`. Deploy manually:

```bash
# Build and load image into Kind
docker build -t personal-blog:latest .
kind load docker-image personal-blog:latest

# Apply Kubernetes manifests
kubectl apply -f manifests/

# (Optional) Set up ArgoCD for GitOps sync
kubectl apply -f argocd/application.yaml
```

### Configuration

**ConfigMap** (`manifests/configmap.yaml`):
```yaml
phx_host: personal-blog.cnoe.localtest.me
gitea_url: http://my-gitea-http.gitea.svc.cluster.local:3000
gitea_user: ""  # Set to your Gitea username or org
cache_ttl: "60"
```

**Secrets** (managed by platform automation):
```yaml
secret_key_base: <generated on first deploy>
gitea_token: <optional; for private repos>
```

### Health Check

```bash
curl http://localhost:4000/healthz
# Returns 200 OK
```

Kubernetes probes:
- **Liveness** — `/healthz` every 20s (10s initial delay)
- **Readiness** — `/healthz` every 10s (3s initial delay)

## Integration Points

### Gitea Discovery

personal-blog discovers content via pattern-based file discovery:

1. **List repos** — Fetches list of accessible repos from Gitea org
2. **Scan directories** — Looks for `portfolio/`, `articles/`, `projects/` in each repo
3. **Fetch files** — Retrieves `.md` files from those directories
4. **Render** — Converts Markdown to HTML via Earmark

**Cache behavior:**
- Successful results cached for 60s (configurable via `BLOG_CACHE_TTL`)
- Failed requests not cached (retry quickly)
- Manual flush via `/internal/cache/flush` (admin-only, not exposed)

### Backstage

Registered via `catalog-info.yaml`:
- Type: `service`
- Lifecycle: `production`
- System: `idp-platform`
- Links back to https://personal-blog.cnoe.localtest.me:8443

### ArgoCD

Application at `argocd/application.yaml`:
- Source: `Unimart-For-Operations/personal-blog` repo (in-cluster Gitea)
- Path: `manifests/`
- Auto-sync enabled (prune + self-heal)
- Deliberately **not** in `packages/` (no auto-deploy with `unimart open`)

## Content Structure

To use personal-blog, create a repository in your Gitea user/org with the following structure:

```
my-personal-repo/
├── README.md                     # Bio/intro (displayed on home)
├── portfolio/
│   ├── 2024-acme-company.md     # Job/role description
│   └── 2024-startup-cto.md
├── projects/
│   ├── my-saas-platform.md       # Project showcase
│   └── opensource-contrib.md
└── articles/
    ├── 2024-elixir-concurrency.md
    └── 2024-kubernetes-secrets.md
```

Then configure:
```bash
kubectl set env deployment/personal-blog GITEA_USER=<your-username-or-org>
```

## Performance

- **Startup time** — ~3s (BEAM VM + Bandit)
- **Request latency** — <100ms (cached), <500ms (Gitea API)
- **Memory footprint** — 128Mi (request), 384Mi (limit)
- **CPU** — 50m (request), 500m (limit)

## Troubleshooting

### No content appears

1. Check Gitea URL: `kubectl logs deployment/personal-blog | grep GITEA`
2. Verify Gitea user is set: `kubectl get configmap personal-blog-config -o yaml | grep gitea_user`
3. Check cache: Wait 60s or restart pod to clear cache
4. Check Gitea API: `curl http://my-gitea-http.gitea.svc.cluster.local:3000/api/v1/user/repos`

### Articles/projects not found

Ensure directory structure matches exactly:
- Must be named `articles/`, `portfolio/`, or `projects/`
- Must be `.md` files
- Gitea token required for private repos

### Build fails

```bash
# Clean build
docker build --no-cache -t personal-blog:latest .

# Check Dockerfile syntax
docker build --progress=plain -t personal-blog:latest .
```

## Conventions

- **Commit style** — Conventional commits + DCO sign-off (`git commit -s`)
- **Code format** — `mix format` (enforced in CI)
- **Git hooks** — Managed via `cmdr`; install via `make hooks` or `unimart deli switch`
- **Documentation** — Markdown in repo root; sync to cdc via `make sync-docs`
- **Test coverage** — No strict percentage; write tests for logic

## See Also

- [Phoenix Documentation](https://hexdocs.pm/phoenix/)
- [LiveView Guide](https://hexdocs.pm/phoenix_live_view/)
- [Gitea API Docs](https://docs.gitea.io/en-us/api-usage/)
- [docs-service AGENTS.md](/home/cmdr/repos/meta/docs-service/AGENTS.md) — Similar service; reference for patterns

## License

Same as the org (check meta repo LICENSE).
