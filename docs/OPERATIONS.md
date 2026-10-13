# Recova Backend Operations Guide

## Overview

This guide covers local development workflows, testing strategies, database operations, containerization, deployment procedures, and runtime observability for Recova Backend.

## Prerequisites & Environment

Ensure the following tools are installed on your workstation:

- **Go**: Version `1.26+` (pinned in `go.mod`)
- **Make**: Standard POSIX Make
- **Docker & Docker Compose**: For containerized database and testing
- **PostgreSQL Client / psql**: For direct database inspection (optional)

Run the preflight verification script to confirm all dependencies:

```bash
make preflight
```

## Quick Start (Local Development)

### 1. Initialize Configuration

Copy the example environment configuration:

```bash
cp .env.example .env
```

Ensure the minimum required local settings are populated:

```env
APP_ENV=local
PORT=3001
DATABASE_URL=postgresql://postgres:postgres@localhost:5432/recova_local?sslmode=disable
JWT_SECRET=local-development-secret-key-32-chars-minimum
GOOGLE_CLIENT_ID=test-client-id.apps.googleusercontent.com
AI_API_KEY=mock-or-valid-gemini-api-key
```

### 2. Start PostgreSQL via Docker

```bash
docker compose -f docker-compose.local.yml up -d
```

### 3. Run Migrations & Seed Baseline Data

```bash
make migrate-up
make seed
```

### 4. Start the Application

```bash
make run
```

The server starts at `http://localhost:3001`. Verify runtime readiness:

```bash
curl http://localhost:3001/health/live
curl http://localhost:3001/health/ready
```

Explore interactive API documentation at `http://localhost:3001/docs/api`.

## Makefile Commands Reference

| Command                          | Action Description                                                      |
| -------------------------------- | ----------------------------------------------------------------------- |
| `make preflight`                 | Verifies local development tools and environment setup                  |
| `make fmt`                       | Formats all Go source files using `gofmt`                               |
| `make lint`                      | Runs Go static analysis with `go vet ./...`                             |
| `make test`                      | Executes all unit and package-level tests                               |
| `make test-integration`          | Runs shell integration test suites (`test/integration/scripts_test.sh`) |
| `make test-e2e`                  | Executes critical path end-to-end integration flows                     |
| `make test-performance`          | Runs performance and concurrency smoke tests                            |
| `make coverage`                  | Generates HTML and summary test coverage reports                        |
| `make build`                     | Compiles the standalone production binary into `./bin/recova-api`       |
| `make run`                       | Runs the API server locally loading `.env` variables                    |
| `make migrate-up`                | Applies all pending database migrations                                 |
| `make migrate-down`              | Rolls back the latest applied database migration step                   |
| `make migrate-status`            | Displays current migration version and status                           |
| `make migrate-check`             | Validates migration state against the target database                   |
| `make migrate-force VERSION=<v>` | Clears dirty migration flags by forcing a version                       |
| `make seed`                      | Applies initial seed data into PostgreSQL                               |
| `make openapi-generate`          | Regenerates OpenAPI 3.1 artifact and route inventory                    |
| `make openapi-check`             | Verifies OpenAPI artifact is fully synchronized with code               |
| `make scalar-check`              | Validates Scalar UI configuration and documentation file paths          |
| `make scalar-preview`            | Starts server with live Scalar documentation preview                    |
| `make module-consistency-check`  | Enforces layer boundaries and module consistency                        |

## Testing Strategy

Recova Backend follows a comprehensive testing pyramid:

```text
       /\
      /  \      E2E Critical Flow Tests (test/e2e)
     /────\     Contract & OpenAPI Verification (test/contract)
    /      \    Integration & Script Tests (test/integration)
   /────────\   Unit & Package Tests (internal/**/*_test.go)
  /──────────\
```

### 1. Unit Tests

Unit tests validate business logic in services, validators, mappers, and utility functions in isolation using mocks or in-memory fixtures.

```bash
make test
```

### 2. Contract & Route Consistency Tests

Located in `test/contract/`, these tests verify that:

- Runtime routes registered in Fiber match the published OpenAPI 3.1 contract.
- Module layers and directory conventions follow architectural rules.

```bash
make module-consistency-check
```

### 3. End-to-End Critical Flow Tests

Located in `test/e2e/` and `scripts/e2e-critical.sh`, these tests execute real HTTP requests against an active test server to validate:

- User registration → Google OAuth → JWT generation → refresh token rotation.
- Daily check-in creation → streak calculation → relapse reporting.
- Journal entry creation and access control.
- Community posting and threaded comment hierarchies.

```bash
make test-e2e
```

### 4. Performance Smoke Tests

Executed via `scripts/performance-smoke.sh`, validating throughput and latency boundaries on critical read/write endpoints.

```bash
make test-performance
```

## Containerization & Deployment

### Multi-Stage Dockerfile

The service uses a secure multi-stage Docker build (`Dockerfile`):

1. **Build Stage**: Compiles the Go binary on Alpine Linux with static linking (`CGO_ENABLED=0`).
2. **Runtime Stage**: Copies only the compiled binary and OpenAPI specification into an unprivileged, minimal Alpine image.

Build container image locally:

```bash
docker build -t recova-backend:latest .
```

### Deployment Profiles

- **Local Development**: `docker-compose.local.yml` (spins up local PostgreSQL).
- **Dokploy / Production**: `docker-compose.dokploy.yml` (production ingress with healthchecks and restart policies).

## Observability & Health Probes

### Health Check Endpoints

- **Liveness Probe** (`GET /health/live`):
  - Used by Kubernetes or container orchestrators to detect process lockups.
  - Returns HTTP `200` with JSON status `ok`.
- **Readiness Probe** (`GET /health/ready`):
  - Pings PostgreSQL via connection pool.
  - In production runtime, an active database probe is required at startup; unconfigured probes fail fast during server assembly.
  - Returns HTTP `200` (`"status": "ready"`) if database is healthy.
  - Returns HTTP `503` (`"status": "not_ready"`) if database connection fails, signaling load balancers to halt traffic routing.

### Prometheus Metrics

- **Endpoint**: `GET /metrics`
- Exposes runtime Prometheus metrics:
  - HTTP requests, duration histograms, and errors partitioned by route and status class (`recova_http_requests_total`, `recova_http_request_duration_seconds`, `recova_http_errors_total`).
  - Database operation latency and status partitioned by operation and table (`recova_db_operation_duration_seconds`).
  - AI provider latency partitioned by provider and model (`recova_ai_request_duration_seconds`) and persona usage counters (`recova_ai_persona_usage_total`).
  - Sensitive action audit event counters (`recova_audit_events_total`).
  - Standard Go runtime process metrics (memory allocation, goroutines, garbage collection).

### Structured Logging & Request Correlation

- Request correlation is maintained via `X-Request-ID` across structured logs, audit events, and API error envelopes. Distributed tracing spans are not implemented; correlation relies on indexed request IDs in structured logs.
- Logs are emitted in structured JSON format with configurable log levels (`LOG_LEVEL=debug|info|warn|error`).
- Request logs capture `requestId`, `method`, `path`, `status`, `latencyMs`, and `userId` (when authenticated).

## Common Troubleshooting

### 1. Startup Fails with `config error`

- **Cause**: Missing or invalid required environment variable.
- **Fix**: Check `.env` against the reference table in [ENVIRONMENT.md](ENVIRONMENT.md). Verify all required fields (`JWT_SECRET`, `AI_API_KEY`, `DATABASE_URL`) are populated.

### 2. Database Connection Refused

- **Cause**: PostgreSQL container is not running or network port is blocked.
- **Fix**: Check `docker ps`. Restart local database with `docker compose -f docker-compose.local.yml up -d`. Verify database credentials in `DATABASE_URL`.

### 3. OpenAPI Check Fails in CI (`openapi-check failed`)

- **Cause**: Route handlers or schemas were modified without regenerating the OpenAPI artifacts.
- **Fix**: Run `make openapi-generate` and commit the updated `docs/generated/openapi.yaml` and `docs/generated/routes.md`.

### 4. Migration Dirty State (`dirty database version`)

- **Cause**: A previous migration failed midway due to syntax error or constraint conflict.
- **Fix**: Inspect PostgreSQL error, fix migration script, then force the clean version with `make migrate-force VERSION=<target_version>`.
