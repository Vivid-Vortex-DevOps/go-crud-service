# go-crud-service

Production-grade Go CRUD microservice for the GitOps Terraform AKS Pipeline POC.

## Overview

A REST API service for managing products, built with Go. Demonstrates:

- Clean architecture (handler → service → repository)
- PostgreSQL via `pgx/v5`
- Database migrations via `golang-migrate`
- Structured JSON logging via `log/slog`
- OpenTelemetry instrumentation (metrics + traces)
- Prometheus metrics endpoint
- Kubernetes-ready health probes
- Multi-stage Docker build (distroless, non-root)
- Helm chart with environment-specific values
- GitHub Actions CI/CD with JFrog publishing

## API

| Method | Path | Description |
|---|---|---|
| `POST` | `/api/v1/products` | Create a product |
| `GET` | `/api/v1/products` | List all products |
| `GET` | `/api/v1/products/{id}` | Get product by ID |
| `PUT` | `/api/v1/products/{id}` | Update a product |
| `DELETE` | `/api/v1/products/{id}` | Delete a product |
| `GET` | `/health/live` | Liveness probe |
| `GET` | `/health/ready` | Readiness probe (checks DB) |
| `GET` | `/metrics` | Prometheus metrics |

## Product Schema

```json
{
  "id": "uuid",
  "name": "string",
  "description": "string",
  "price": 9.99,
  "quantity": 10,
  "createdAt": "2024-01-01T00:00:00Z",
  "updatedAt": "2024-01-01T00:00:00Z"
}
```

## Repository Structure

```
go-crud-service/
├── cmd/
│   └── server/
│       └── main.go          ← Entry point, dependency wiring
├── internal/
│   ├── config/              ← Environment variable configuration
│   ├── handler/             ← HTTP handlers, middleware, routing
│   ├── model/               ← Domain types (Product, request/response DTOs)
│   ├── repository/          ← PostgreSQL data access layer
│   ├── service/             ← Business logic layer
│   └── telemetry/           ← OpenTelemetry setup
├── migrations/              ← SQL migration files (golang-migrate)
├── deployment/
│   ├── helm/                ← Helm chart
│   └── values/
│       ├── dev.yaml
│       ├── qa.yaml
│       ├── staging.yaml
│       └── prod.yaml
├── docs/                    ← Documentation
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── go.sum
├── VERSION
└── .github/
    └── workflows/
```

## Local Development

### Prerequisites

- Go 1.23+
- Docker + Docker Compose
- `golangci-lint` (optional, for linting)

### Run with Docker Compose

```bash
# Copy environment template
cp .env.example .env
# Edit .env and set POSTGRES_PASSWORD

docker compose up
```

The service starts at `http://localhost:8080`.

### Run without Docker

```bash
# Start PostgreSQL only
docker compose up postgres -d

export DATABASE_URL="postgres://gouser:yourpassword@localhost:5432/go_service_db?sslmode=disable"
export SERVER_PORT="8080"
export ENVIRONMENT="dev"

go run ./cmd/server/
```

### Run Tests

```bash
# Unit tests only (no Docker required)
go test ./...

# Integration tests (requires Docker for Testcontainers)
go test --tags=integration ./...

# With coverage
go test -cover ./...
```

### Lint

```bash
golangci-lint run
go vet ./...
```

## Configuration

All configuration is via environment variables. No defaults for secrets.

| Variable | Required | Default | Description |
|---|---|---|---|
| `DATABASE_URL` | Yes | — | PostgreSQL connection string |
| `SERVER_PORT` | No | `8080` | HTTP server port |
| `ENVIRONMENT` | No | `dev` | Environment name (dev/qa/staging/prod) |
| `LOG_LEVEL` | No | `info` | Log level (debug/info/warn/error) |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | No | — | OTLP endpoint for traces |
| `OTEL_SERVICE_NAME` | No | `go-crud-service` | Service name for telemetry |

## Docker

```bash
# Build
docker build -t go-crud-service:local .

# Run
docker run -p 8080:8080 \
  -e DATABASE_URL="postgres://gouser:pass@host/go_service_db?sslmode=disable" \
  go-crud-service:local
```

The image runs as a non-root user (`nonroot:nonroot`) using Google's distroless base.

## Deployment

See [docs/deployment.md](docs/deployment.md) for Kubernetes deployment via ArgoCD and Helm.

## Documentation Index

| Topic | Document |
|---|---|
| API Reference | [docs/api.md](docs/api.md) |
| Local Development | [docs/local-development.md](docs/local-development.md) |
| Docker | [docs/docker.md](docs/docker.md) |
| Testing | [docs/testing.md](docs/testing.md) |
| Deployment | [docs/deployment.md](docs/deployment.md) |
| Observability | [docs/observability.md](docs/observability.md) |
| Troubleshooting | [docs/troubleshooting.md](docs/troubleshooting.md) |

## Related Repositories

| Repository | Purpose |
|---|---|
| [cloud-platform-infra](https://github.com/Vivid-Vortex-DevOps/cloud-platform-infra) | Azure infrastructure + GitOps |
| [springboot-crud-service](https://github.com/Vivid-Vortex-DevOps/springboot-crud-service) | Spring Boot equivalent service |
