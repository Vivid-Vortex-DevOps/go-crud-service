# Local Development

## Prerequisites

- Go 1.23+
- Docker + Docker Compose
- `make` (optional)

## Quick Start with Docker Compose

```bash
# Start PostgreSQL + application
DATABASE_PASSWORD=localpassword docker compose up

# Application available at:
# http://localhost:8080/health/live
# http://localhost:8080/api/v1/products
# http://localhost:8080/metrics
```

## Running Without Docker Compose

```bash
# Start only PostgreSQL
DATABASE_PASSWORD=localpassword docker compose up -d postgres

# Run application
export DATABASE_URL="postgres://crud_user:localpassword@localhost:5432/go_service_db?sslmode=disable"
export PORT=8080
go run ./cmd/server/...
```

## Environment Variables

| Variable | Required | Default | Description |
|---|---|---|---|
| `DATABASE_URL` | Yes | — | PostgreSQL connection string |
| `PORT` | No | `8080` | HTTP server port |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | No | `""` | OTLP gRPC endpoint (empty = tracing disabled) |
| `OTEL_SERVICE_VERSION` | No | `dev` | Service version tag in traces |

## Running Tests

```bash
# Unit tests only (no external dependencies)
go test ./internal/...

# Integration tests (requires Docker for Testcontainers)
go test -tags=integration ./...

# With coverage
go test -cover ./internal/...
```

## Code Structure

```
cmd/server/main.go          — entry point, wires everything together
internal/
  config/                   — environment variable loading
  model/                    — domain types (Product, errors)
  repository/               — database access interface + PostgreSQL impl
  service/                  — business logic, UUID generation, validation
  handler/                  — HTTP handlers, middleware, response helpers
  telemetry/                — OpenTelemetry setup (metrics + traces)
migrations/                 — SQL migration files (embedded in binary)
deployment/helm/            — Kubernetes Helm chart
```

## Database Migrations

Migrations run automatically at startup via `golang-migrate`.  
Migration files are embedded in the binary (`//go:embed`), so no external files are needed.

To add a migration:
```bash
# Create new migration files
touch migrations/000003_add_category.up.sql
touch migrations/000003_add_category.down.sql
# Edit the SQL files, then restart the app
```

## Linting

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
golangci-lint run ./...
```

## Common Issues

### `DATABASE_URL is required`
You forgot to set the env var. Set it or use Docker Compose.

### `connection refused` on startup
PostgreSQL isn't ready yet. Kubernetes handles this with restarts (restartPolicy: Always).  
Locally, wait a few seconds and retry.

### Migration fails: `pq: role "crud_user" does not exist`
The Docker Compose init script creates the user. Ensure you're using the full compose setup, not a bare PostgreSQL container.
