# Recova Backend API (Go Fiber)

Main backend service for Recova, handling authentication, profiles, routines/check-ins, journaling, community features, educational content, AI Coach, and API observability.

Active runtime for this repository: **Go + Fiber**.
Legacy runtime archive (Node/Express): [README.md](https://github.com/recova-app/backend/blob/main/README.md).

## Table of Contents

- [Service Scope](#service-scope)
- [Domain Modules](#domain-modules)
- [Tech Stack](#tech-stack)
- [Quick Start](#quick-start)
- [Environment Configuration](#environment-configuration)
- [Make Commands](#make-commands)
- [Testing and Quality Gates](#testing-and-quality-gates)
- [OpenAPI and API Docs](#openapi-and-api-docs)
- [Deployment Overview](#deployment-overview)
- [Repository Structure](#repository-structure)
- [Documentation Map](#documentation-map)
- [Quick Troubleshooting](#quick-troubleshooting)

## Service Scope

This service handles:

- Google Login + JWT session flow (`/api/v1/auth/*`)
- User profiles, onboarding, and settings (`/api/v1/users/*`)
- Routine check-ins, relapse tracking, and activity statistics (`/api/v1/routine/*`)
- Personal journals (`/api/v1/journals`)
- Community posts, comments, replies, and likes (`/api/v1/community*`)
- Educational content and daily content (`/api/v1/education`, `/api/v1/content/daily`)
- AI Coach chat, summaries, onboarding analysis, and persona preferences (`/api/v1/ai/*`)
- Health checks, metrics, and OpenAPI runtime routes (`/health/*`, `/metrics`, `/openapi.yaml`, `/docs/api`)

## Domain Modules

Active modules are located in `internal/modules/`:

- `auth`
- `users`
- `routine`
- `journals`
- `community`
- `education`
- `content`
- `ai`
- `achievements`

Runtime route inventory is automatically synchronized in: `docs/generated/routes.md`.

## Tech Stack

| Area              | Technology                               |
| ----------------- | ---------------------------------------- |
| Language          | Go 1.25.x (toolchain pinned in `go.mod`) |
| HTTP Framework    | Fiber v3.1.x                             |
| Database          | PostgreSQL                               |
| ORM               | GORM                                     |
| Migration         | golang-migrate (SQL up/down)             |
| API Contract      | OpenAPI 3.1                              |
| API Docs UI       | Scalar                                   |
| Authentication    | JWT + Google OAuth token verification    |
| Metrics           | Prometheus client                        |
| Build/Task Runner | Make                                     |
| Container Runtime | Docker + Docker Compose                  |

## Quick Start

### 1) Prerequisites

- Go `>= 1.25`
- Make
- Git
- Docker + Docker Compose (required for local DB/container profile)

Check the baseline setup:

```bash
make preflight
```

### 2) Setup Environment

```bash
cp .env.example .env
```

Update the minimum required values:

- `JWT_SECRET`
- `GOOGLE_CLIENT_ID`
- `AI_API_KEY`
- `DATABASE_URL`

Default local configuration:

- `PORT=3001`
- `API_PREFIX=/api/v1`
- `APP_ENV=local`

### 3) Choose Local Runtime Profile

#### Option A - Native API + DB in Docker (recommended for development)

```bash
docker compose -f docker-compose.local.yml up -d db
make migrate-up
make seed
make run
```

#### Option B - Full Docker Compose (API + DB)

```bash
docker compose -f docker-compose.local.yml up -d --build
make migrate-up
make seed
docker compose -f docker-compose.local.yml logs -f api
```

### 4) Quick Verification

```bash
curl -fsS http://localhost:3001/health/live
curl -fsS http://localhost:3001/health/ready
curl -I http://localhost:3001/docs/api
```

## Environment Configuration

Complete references are available in:

- `docs/environment.md`
- `.env.example`

Main environment variable groups:

- Application: `APP_NAME`, `APP_ENV`, `NODE_ENV`, `PORT`, `API_PREFIX`, `APP_URL`, `DOCS_URL`
- Database: `DATABASE_URL`, `DIRECT_DATABASE_URL`, pooling, SSL mode
- Authentication: `JWT_SECRET`, token TTL, cookie policy, `GOOGLE_CLIENT_ID`
- AI: `AI_PROVIDER`, `AI_MODEL`, `AI_API_KEY`, fallback provider/model
- Security: CORS, rate limiting, request body limit
- Observability: log level, request ID header, health timeout

Notes:

- `make run`, `make migrate-*`, and `make seed` automatically load `.env` via `scripts/with-env.sh`
- Override the env file using: `ENV_FILE=.env.staging make run`

## Make Commands

All tasks are available in the `Makefile`. Core commands:

### Development

- `make preflight` check tools and project structure baseline
- `make fmt` format Go code (`gofmt -w .`)
- `make lint` static checks (`go vet ./...`)
- `make build` build binary `./bin/recova-api`
- `make run` run the local API

### Database

- `make migrate-up`
- `make migrate-down`
- `make migrate-status`
- `make migrate-check`
- `make migrate-force VERSION=<n>`
- `make seed`

### API Contract and Docs

- `make openapi-generate`
- `make openapi-check`
- `make openapi-autogen`
- `make openapi-autogen-watch`
- `make openapi-autogen-install-hook`
- `make scalar-check`
- `make scalar-preview`

### Testing and Validation

- `make test`
- `make test-integration`
- `make test-e2e`
- `make test-performance`
- `make coverage`
- `make release-validation`
- `make module-consistency-check`
- `make module-consistency-full-check`
- `make security-scan`
- `make compose-smoke`

### Deployment Operations

- `make staging-deploy`
- `make cutover-wave WAVE=<64|65|66|67|68>`
- `make cutover-all`
- `make stabilization-gate`
- `make rollback-rehearsal`
- `make runtime-decommission`
- `make post-migration-maintenance`

## Testing and Quality Gates

Baseline checks before merge:

```bash
make fmt
make lint
make test
make test-integration
make openapi-check
make scalar-check
make module-consistency-check
```

For higher confidence before release:

```bash
make release-validation
make module-consistency-full-check
```

Detailed strategy: `docs/operations/testing.md`.

## OpenAPI and API Docs

API contract source:

- `api/openapi/openapi.yaml`

Generated artifacts:

- `docs/generated/openapi.yaml`
- `docs/generated/routes.md`
- `scalar.config.json` (Scalar navigation + docs wiring)

Runtime endpoints:

- Raw OpenAPI: `GET /openapi.yaml`
- Scalar UI: `GET /docs/api`

Recommended workflow:

```bash
make openapi-generate
make openapi-check
```

Notes:

- `make openapi-autogen-watch` maintains `.openapi-watch-files` and regenerates artifacts on change.
- Runbook: `docs/operations/api-docs-generation.md`.

## Deployment Overview

### Staging Compose Runner (local/runner environment)

```bash
make staging-deploy
```

Script: `scripts/staging-deploy.sh`
Compose target: `docker-compose.staging.yml`

### Remote Staging from `develop` Branch

Main CI/CD workflow:

- `.github/workflows/deploy-staging.yml`
- `scripts/deploy/remote-deploy.sh`

Deployment smoke checks include:

- `/health/live`
- `/health/ready`
- `/openapi.yaml`

Complete reference: `docs/operations/deployment.md`.

## Repository Structure

```text
.
├── api/                       # OpenAPI source
├── artifacts/                 # Test/deployment artifact outputs
├── cmd/                       # Application/tool entrypoints
├── docs/                      # Technical documentation
├── internal/                  # Application code (app, modules, platform, shared)
├── migrations/                # SQL migrations + seeds
├── scripts/                   # Dev/ops automation scripts
├── test/                      # Integration harness and validation scripts
├── docker-compose.local.yml   # Local runtime
├── docker-compose.staging.yml # Staging runtime
├── Dockerfile                 # Production-style image build
├── scalar.config.json         # Scalar docs configuration
└── Makefile                   # Task runner
```

## Documentation Map

Start here:

- `docs/overview.md`

Important documents:

- Architecture: `docs/architecture.md`
- API: `docs/api/index.md` and `docs/api-reference.md`
- Database: `docs/database.md`
- Environment: `docs/environment.md`
- Project structure: `docs/project-structure.md`
- Tech stack: `docs/tech-stack.md`
- Local development: `docs/operations/local-development.md`
- Testing: `docs/operations/testing.md`
- Deployment: `docs/operations/deployment.md`
- Security: `docs/operations/security.md`
- API docs generation: `docs/operations/api-docs-generation.md`
- Documentation sync: `docs/operations/documentation-sync.md`
- Modules index: `docs/modules/index.md`

## Quick Troubleshooting

- `config error` during startup: check required environment variables in `.env` and `docs/environment.md`
- Missing `migrate` binary: scripts will automatically fallback to the Docker image `migrate/migrate:v4.19.0` if Docker is available
- `openapi-check` failed: run `make openapi-generate` and commit changes inside `docs/generated/*`
- `health/ready` failed: check DB connectivity and migration status (`make migrate-status`, `make migrate-check`)
