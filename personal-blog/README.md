# Personal Blog

A Phoenix microservice for showcasing your personal portfolio, projects, and articles. Content is discovered from Gitea repositories and rendered in a beautiful, responsive interface.

## Quick Start

### Local Development

```bash
nix develop
make setup
make server
```

Visit http://localhost:4000.

### Docker Build

```bash
make build
# Image: personal-blog:latest
```

### Kubernetes Deploy

```bash
docker build -t personal-blog:latest .
kind load docker-image personal-blog:latest
kubectl apply -f manifests/
```

Access at https://personal-blog.cnoe.localtest.me:8443 (in-cluster).

## Features

- **Portfolio** — Showcase professional experience from markdown files
- **Projects** — Highlight projects and technical work
- **Articles** — Publish blog posts and technical writing
- **Live UI** — Phoenix LiveView for responsive, interactive interface
- **Tailwind CSS** — Modern, responsive design
- **Gitea Integration** — Auto-discover content from repos
- **Caching** — ETS-backed TTL cache for performance
- **Dark Mode** — Built-in dark theme support (via Tailwind)

## Content Structure

Create directories in your Gitea repository:

```
portfolio/       # Professional entries
articles/        # Blog posts
projects/        # Project showcases
README.md        # Your bio/intro
```

## Configuration

Environment variables (production):

```bash
GITEA_URL         # Gitea API base URL (default: in-cluster)
GITEA_USER        # Username/org to scan (required)
GITEA_TOKEN       # Optional; for private repos
BLOG_CACHE_TTL    # Cache TTL in seconds (default: 60)
PHX_HOST          # Hostname for links (default: personal-blog.cnoe.localtest.me)
SECRET_KEY_BASE   # Phoenix session secret (generated on first deploy)
```

## Development

See [AGENTS.md](./AGENTS.md) for comprehensive developer guide.

### Make Targets

- `make setup` — Install dependencies
- `make server` — Start dev server with hot reload
- `make test` — Run tests
- `make fmt` — Format code
- `make lint` — Lint with Credo
- `make ci` — Full CI check
- `make build` — Build Docker image

## Architecture

- **Framework** — Phoenix 1.7 + LiveView
- **HTTP** — Bandit (native BEAM)
- **Cache** — ETS with TTL
- **Markdown** — Earmark
- **CSS** — Tailwind 3.4
- **No DB** — Stateless, read-only service

## Project Layout

```
lib/
├── personal_blog/
│   ├── application.ex   # OTP supervision tree
│   ├── cache.ex         # TTL cache
│   └── gitea.ex         # API client
└── personal_blog_web/
    ├── router.ex        # Routes
    ├── endpoint.ex      # HTTP server
    ├── doc_renderer.ex  # Markdown → HTML
    ├── controllers/     # HTTP controllers
    └── live/blog_live/  # LiveView pages
config/
├── config.exs           # Base config
├── dev.exs              # Dev overrides
├── prod.exs             # Prod overrides
├── runtime.exs          # Runtime config
└── test.exs             # Test config
assets/
├── css/app.css          # Tailwind input
└── js/app.js            # LiveSocket entry
manifests/               # Kubernetes YAML
argocd/                  # ArgoCD Application
```

## Deployment

This service is **not auto-deployed** in `unimart open`. Deploy manually when ready:

```bash
# Build locally
make build

# Load into Kind cluster
kind load docker-image personal-blog:latest

# Apply manifests
kubectl apply -f manifests/

# Apply ArgoCD for GitOps (optional)
kubectl apply -f argocd/application.yaml
```

## Links

- [AGENTS.md](./AGENTS.md) — Architecture, development guide, troubleshooting
- [Dockerfile](./Dockerfile) — Multi-stage build
- [mix.exs](./mix.exs) — Dependencies

## License

MIT
