# Recova Backend — Agent Instructions

Recova Backend (`recova-backend-v2`) is the authoritative RESTful API service powering the Recova addiction recovery platform, primarily consumed by a Flutter mobile client. Built in Go using Fiber v3, PostgreSQL (Neon Serverless in production), GORM, and `golang-migrate`. Scope covers domain business logic, OAuth & JWT authentication, routine check-ins, journal entries, recovery education, and AI coach interactions. Non-goals: hosting mobile UI views or running client-side computation.

## Authority and source of truth

- System architecture, topology, and layered constraints: `docs/ARCHITECTURE.md`
- HTTP contracts, envelopes, and error codes: `docs/API.md` and `api/openapi/openapi.yaml`
- Database schema, migrations, indexing, and GORM models: `docs/DATABASE.md`
- Configuration variables and security policies: `docs/ENVIRONMENT.md`
- Domain module specifications: `docs/MODULES.md`
- Operational tasks, local testing, and developer runbooks: `docs/OPERATIONS.md`
- Production deployment, SSH release runner, and VPS topology: `docs/DEPLOYMENT.md`

## Stack and runtime

- **Language & Compiler**: Go `1.26+` (pinned in `go.mod`, toolchain managed).
- **HTTP Engine**: Fiber `v3` (`github.com/gofiber/fiber/v3`).
- **ORM & Persistence**: GORM (`gorm.io/gorm`) with PostgreSQL driver (`gorm.io/driver/postgres`).
- **Migrations**: Raw versioned SQL files via `golang-migrate/migrate/v4`.
- **API Documentation**: OpenAPI 3.1 specification rendered via Scalar UI (`/docs/api`).
- **Containerization**: Multi-stage Docker build (`golang:1.26-alpine` builder, `alpine:3.22` unprivileged runner).
- **Production Target**: Standalone Docker Compose stack behind Host Nginx (`127.0.0.1:3001`), Cloudflare Tunnel, and Neon Serverless PostgreSQL.

## Repository map and boundaries

| Path                         | Responsibility                                        | Boundary / Ownership Rule                                                         |
| :--------------------------- | :---------------------------------------------------- | :-------------------------------------------------------------------------------- |
| `cmd/api/`                   | Application entrypoint (`main.go`)                    | Bootstraps configuration, database, router, and lifecycle signals                 |
| `cmd/tools/`                 | Developer CLI tools (`openapi`, `scalar`, `coverage`) | Generates OpenAPI artifacts and documentation configs                             |
| `internal/app/`              | Bootstrap, HTTP router assembly, and lifecycle        | Wires dependencies and middleware; no domain business logic                       |
| `internal/modules/<domain>/` | Self-contained domain modules                         | Encapsulates handler, service, repository, model, dto, mapper, and validator      |
| `internal/platform/`         | Cross-cutting technical capabilities                  | `config`, `database`, `ai`, `logger`, `observability`, `openapi`                  |
| `internal/shared/`           | Shared primitives and error envelopes                 | Domain-agnostic error types (`errs`), response envelopes, and utility functions   |
| `migrations/`                | Versioned SQL migrations (`*.up.sql`, `*.down.sql`)   | Authoritative database schema; manual SQL only                                    |
| `migrations/seeds/`          | Baseline development & test seeds                     | Fixture records for testing and local bootstrapping                               |
| `api/openapi/`               | Source OpenAPI 3.1 contract                           | Base API contract definitions and schemas                                         |
| `docs/generated/`            | Auto-generated API documentation                      | Do not edit manually; update via `make openapi-generate`                          |
| `tools/deploy/`              | VPS deployment and snapshot utilities                 | `deploy.sh` (release runner) and `backup.sh` (disaster recovery)                  |
| `tools/nginx/`               | Host Nginx reverse proxy configuration template       | Hardened proxy pass to `127.0.0.1:3001`                                           |
| `.github/workflows/`         | CI/CD automation pipelines                            | `ci-pr.yml` (PR validation gate), `ci-cd.yml` (production release over Tailscale) |

## Working rules

### 1. Unidirectional Layer Discipline

Within every domain module in `internal/modules/<domain>/`:

- **Router → Handler → Service → Repository → Database**.
- **Service Layer**: Owns business logic, validation, and domain orchestration. **Strictly forbidden** to import Fiber (`github.com/gofiber/fiber/v3`) or access HTTP context directly.
- **Handler Layer**: Owns request parsing, validation binding, and response formatting. **Strictly forbidden** to query the database or GORM handles directly.
- **Repository Layer**: Owns SQL queries, GORM operations, and row mapping. **Strictly forbidden** to format HTTP response envelopes or define API models.

### 2. Client Envelopes & Language Boundaries

- All HTTP responses must adhere to standard JSON envelope: `{"success": bool, "message": string, "data": ..., "meta": ..., "error": ...}`.
- **Machine identifiers** (JSON fields, enum values, `error.code`) must be English (`snake_case` or `UPPER_SNAKE_CASE`).
- **User-facing messages** (`message`, `error.details`) must be in Indonesian for consistency with Flutter mobile client UI.
- **Internal code artifacts** (identifiers, types, logs, docs, commit messages) must be strictly English.

### 3. Database Schema & Migration Invariants

- Never use GORM AutoMigrate in production or migration paths.
- Every domain model change must be accompanied by corresponding `.up.sql` and `.down.sql` files in `migrations/`.
- Ensure migrations are idempotent (`IF NOT EXISTS`, `IF EXISTS`) and backward compatible.

### 4. OpenAPI Contract Synchronization

- Runtime routes and schemas must strictly match OpenAPI 3.1 artifacts.
- When modifying route signatures or handler contracts, run `make openapi-generate` and commit both code and generated docs.

## Request workflow

1. **Inspect**: Check `git status -sb` and read relevant domain files in `internal/modules/<domain>/` and contracts in `docs/`.
2. **Plan**: State a concise plan for changes spanning multiple modules, migrations, or contract boundaries.
3. **Execute**: Make surgical edits respecting layer boundaries and local conventions.
4. **Verify**: Run focused unit tests, `go vet`, `make openapi-check`, and `git diff --check`.
5. **Sanitize**: Ensure zero secrets, credentials, internal IPs, or private host ports are committed.

## Testing and verification

| Change Surface                 | Minimum Safe Verification Command                                |
| :----------------------------- | :--------------------------------------------------------------- |
| Any Go code edit               | `make fmt && make lint`                                          |
| Domain service/unit logic      | `go test -v ./internal/modules/<domain>/...`                     |
| Route, handler, or API schema  | `make openapi-generate && make openapi-check`                    |
| Module architectural boundary  | `make module-consistency-check`                                  |
| Database model / SQL migration | `make schema-check` (or `make migrate-up && make migrate-check`) |
| Pre-commit verification        | `make preflight && make lint && make test`                       |
| Binary build verification      | `make build`                                                     |

## Security, data, and external operations

- **Zero Secret Exposure**: Never log, print, echo, commit, or hardcode secrets (`JWT_SECRET`, `AI_API_KEY`, passwords, bearer tokens). Reference via environment variables only.
- **Masking**: Passwords and refresh tokens must be cryptographically hashed; sensitive logs must be redacted.
- **Outbound AI Endpoints**: In `production`, `AI_BASE_URL` must use `https://` and belong to trusted domains or explicitly whitelisted hosts via `AI_ALLOWED_HOSTS`.
- **Database SSL Mode**: In `production`, `DATABASE_SSL_MODE` must strictly be `require`, `verify-ca`, or `verify-full`.
- **Destructive Operations**: Dropping tables, running destructive migrations, or mutating VPS production states requires explicit operator approval per `~/.pi/agent/protocols/safety.md`.

## Stop conditions

Stop, report the blocker, and ask before proceeding when:

- Proposed change breaks existing mobile API response envelopes or removes public endpoints without deprecation strategy.
- Required environment variable, migration version, or external secret is unavailable.
- Layer boundary must be violated to resolve an issue (e.g. service needing HTTP headers).
- Uncommitted unrelated dirty files exist in the target working tree.

## Completion checklist

- [ ] Requested scope satisfied without pattern drift.
- [ ] Layer boundaries respected (Handler → Service → Repository).
- [ ] Response envelopes maintain English technical keys and Indonesian user messages.
- [ ] OpenAPI schema synchronized (`make openapi-check` passed).
- [ ] Unit/module tests pass (`go test ./...`).
- [ ] Binary compiles cleanly (`make build`).
- [ ] No secrets, internal IPs, or machine-specific paths committed.
