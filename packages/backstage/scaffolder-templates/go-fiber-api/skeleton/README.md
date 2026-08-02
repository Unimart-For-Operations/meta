# ${{ values.name }}

${{ values.description }}

## Overview

This is a Go REST API service built with the [Fiber](https://gofiber.io/) web framework.

**Owner:** ${{ values.owner }}

## Features

- ✅ High-performance HTTP server using Fiber
{% if values.enableMetrics %}
- ✅ Prometheus metrics endpoint (`/metrics`)
{% endif %}
{% if values.enableHealthChecks %}
- ✅ Health check endpoints (`/health`, `/ready`)
{% endif %}
{% if values.enableCORS %}
- ✅ CORS support
{% endif %}
- ✅ Structured logging
- ✅ Graceful shutdown
- ✅ Environment-based configuration

## Quick Start

### Prerequisites

- Go 1.23+
- Docker (optional, for containerized deployment)

### Local Development

```bash
# Install dependencies
go mod download

# Run the service
go run cmd/server/main.go

# Service will be available at http://localhost:${{ values.port }}
```

### Build

```bash
# Build binary
go build -o bin/${{ values.name }} cmd/server/main.go

# Run binary
./bin/${{ values.name }}
```

### Docker

```bash
# Build image
docker build -t ${{ values.name }}:latest .

# Run container
docker run -p ${{ values.port }}:${{ values.port }} ${{ values.name }}:latest
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/hello` | Hello world endpoint |
{% if values.enableHealthChecks %}
| GET | `/health` | Health check |
| GET | `/ready` | Readiness check |
{% endif %}
{% if values.enableMetrics %}
| GET | `/metrics` | Prometheus metrics |
{% endif %}

## Configuration

Configuration is managed via environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | HTTP server port | `${{ values.port }}` |
| `LOG_LEVEL` | Logging level (debug, info, warn, error) | `info` |
{% if values.enableCORS %}
| `CORS_ORIGINS` | Allowed CORS origins (comma-separated) | `*` |
{% endif %}

## Testing

```bash
# Run unit tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Project Structure

```
.
├── cmd/
│   └── server/
│       └── main.go          # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go        # Configuration management
│   ├── handlers/
│   │   └── handlers.go      # HTTP handlers
│   └── middleware/
│       └── logger.go        # Middleware (logging, etc.)
├── Dockerfile               # Container image definition
├── go.mod                   # Go module definition
├── go.sum                   # Go module checksums
├── Makefile                 # Build automation
└── README.md                # This file
```

## Deployment

See `k8s/` directory for Kubernetes manifests or use ArgoCD for GitOps deployment.

## Contributing

1. Create a feature branch
2. Make your changes
3. Add tests
4. Submit a pull request

## License

Copyright © ${{ "now" | date("YYYY") }} Unimart-For-Operations
