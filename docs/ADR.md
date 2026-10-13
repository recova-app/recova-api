# Architecture Decision Records (ADR)

This document consolidates key architectural and design decisions adopted for the Recova Backend service.

## ADR-0001: Migration to Go Fiber Runtime

- **Status**: Accepted
- **Date**: 2026-05-08

### Context

The original Recova backend was implemented in Node.js/Express. As user growth increased and mobile check-ins created synchronized traffic spikes (especially during morning and evening routines), the service experienced higher memory footprints, slower cold starts, and runtime typing errors in complex streak calculation logic.

### Decision

Migrate the backend service entirely to **Go** using the **Fiber v3** framework. Go provides compiled performance, predictable memory utilization, native concurrency, and strong static typing.

### Consequences

- **Positive**:
  - Significant reduction in memory consumption and CPU utilization under concurrent load.
  - Sub-millisecond response times on core check-in and read endpoints.
  - Compile-time type safety preventing runtime type bugs.
- **Negative / Trade-offs**:
  - Reimplementation required rewriting existing domain logic and test suites in Go.
  - Team needed to align on idiomatic Go error handling and architectural patterns.

## ADR-0002: Modular Clean Architecture & Project Directory Layout

- **Status**: Accepted
- **Date**: 2026-05-08

### Context

To support parallel development and maintain clear module boundaries as the codebase grows, the project structure required explicit rules regarding dependency direction, domain isolation, and platform separation.

### Decision

Adopt standard Go project layout with domain modules encapsulated under `internal/modules/`:

- `cmd/api`: Runtime entrypoint only (flags, configuration loading, root bootstrapping).
- `internal/app`: Application wiring, HTTP server configuration, lifecycle management.
- `internal/modules/<domain>`: Self-contained domain modules containing handler, service, repository, model, dto, mapper, and validator.
- `internal/platform`: Infrastructure adapters (database, config, logging, observability, AI clients).
- `internal/shared`: Domain-agnostic cross-cutting primitives (standard envelopes, common error types, utilities).

### Consequences

- **Positive**:
  - Unidirectional dependency flow: `route -> handler -> service -> repository -> database`.
  - Easy to navigate, test, and refactor individual modules without ripple effects.
  - Prevents circular dependencies across domain boundaries.
- **Negative / Trade-offs**:
  - Requires maintaining DTO and Mapper boilerplate per module.

## ADR-0003: Fiber v3 HTTP Framework

- **Status**: Accepted
- **Date**: 2026-05-08

### Context

Choosing an HTTP framework for Go required balancing developer ergonomics, routing performance, middleware ecosystem, and familiarity for engineers transitioning from Express.

### Decision

Adopt **Fiber v3** (built atop Fasthttp). Fiber provides an Express-like routing and middleware API, zero-memory-allocation routing, and robust built-in middleware (CORS, recover, request ID, logger).

### Consequences

- **Positive**:
  - High request throughput and minimal memory allocations.
  - Natural migration path from Express route definitions.
  - Built-in context utilities for error handling and JSON serialization.
- **Negative / Trade-offs**:
  - Does not use standard `net/http` Handler interface directly; standard `net/http` middleware requires Fiber adaptors.

## ADR-0004: PostgreSQL Persistence with GORM ORM

- **Status**: Accepted
- **Date**: 2026-05-08

### Context

The application requires relational ACID transactions for streak tracking, relational integrity for threaded community comments, and JSONB capabilities for semi-structured onboarding answers.

### Decision

Use **PostgreSQL 16+** as the primary relational database, accessed through **GORM v1.25+** as the ORM layer.

### Consequences

- **Positive**:
  - PostgreSQL provides proven reliability, advanced indexing, and native JSONB support.
  - GORM accelerates developer velocity with typed query builders, hooks, and association preloading.
- **Negative / Trade-offs**:
  - Complex aggregations and reporting queries must use raw SQL or carefully crafted GORM expressions to avoid N+1 query patterns.

## ADR-0005: Raw SQL Migrations via golang-migrate

- **Status**: Accepted
- **Date**: 2026-05-08

### Context

While GORM provides `AutoMigrate`, automated schema mutation in production poses significant risks: lack of rollbacks, risk of unintentional data loss, unpredictable lock acquisitions, and poor auditability.

### Decision

Manage all database schema changes strictly using **`golang-migrate`** with versioned forward (`.up.sql`) and backward (`.down.sql`) migration files. GORM `AutoMigrate` is completely disabled.

### Consequences

- **Positive**:
  - Every schema change is explicitly versioned, reviewed, and tested in pull requests.
  - Deterministic and reversible schema deployments across all environments.
  - Zero surprises during production deployments.
- **Negative / Trade-offs**:
  - Developers must write SQL migrations manually whenever domain models change.

## ADR-0006: GORM Data Modeling Conventions

- **Status**: Accepted
- **Date**: 2026-05-08

### Context

Inconsistent naming conventions, primary key formats, and implicit behaviors across domain entities can cause subtle bugs and database schema drift.

### Decision

Establish strict entity modeling rules:

- **Primary Keys**: UUID v4 strings for all entity IDs (`type:uuid;primaryKey`).
- **Table Names**: Snake_case plural names (e.g. `community_posts`, `check_ins`).
- **Column Tags**: Explicit GORM tags defining column names and constraints (`gorm:"column:created_at;not null"`).
- **Soft Delete Policy**: Automatic GORM `gorm.DeletedAt` is only applied where explicit soft deletion is required by business rules; hard deletes are preferred for ephemeral data (such as revoked refresh tokens).

### Consequences

- **Positive**:
  - Clean, predictable schema naming across database, ORM models, and OpenAPI specs.
  - UUIDs prevent enumeration attacks on mobile endpoints.
- **Negative / Trade-offs**:
  - UUID primary keys require slightly more storage space and index overhead compared to sequential integers.

## ADR-0007: Dual-Token Authentication Strategy

- **Status**: Accepted
- **Date**: 2026-05-09

### Context

Mobile clients require a seamless login experience without prompting users to re-authenticate daily, while backend security requires the ability to revoke sessions instantly if credentials or devices are compromised.

### Decision

Implement a **Dual-Token Authentication Strategy**:

1. **Access Token**:
   - Short-lived HMAC-SHA256 JWT (15-minute TTL).
   - Transmitted in `Authorization: Bearer <token>` request header.
   - Verified statelessly in middleware without hitting the database on every request.
2. **Refresh Token**:
   - Long-lived cryptographically secure random token (7-day TTL).
   - Stored in PostgreSQL `auth_refresh_tokens` table as a one-way SHA-256 hash.
   - Delivered via `HttpOnly`, `SameSite=Lax`, `Secure` cookies, with JSON response fallback for mobile clients.
   - Rotated on every use: old token is immediately marked revoked and replaced with a new pair.

### Consequences

- **Positive**:
  - Fast, stateless access token verification for 99% of requests.
  - Instant session revocation capability upon logout or security breach.
  - Protection against token theft via automatic refresh token rotation.
- **Negative / Trade-offs**:
  - Requires database write on every token refresh operation (`POST /api/v1/auth/refresh`).

## ADR-0008: Automated Verification Gates & CI/CD Pipeline

- **Status**: Accepted
- **Date**: 2026-05-09

### Context

To maintain high software quality and prevent regressions in API contracts, code style, or deployment configurations, all validation steps need automated enforcement.

### Decision

Implement strict automated verification gates required on every commit and pull request:

1. `make hooks-install`: Pre-commit hook gate (`.githooks/pre-commit`) enforcing staged formatting (`gofmt`), OpenAPI synchronization, static analysis (`go vet`), and fast unit tests before commit creation.
2. `make preflight`: Toolchain and dependency verification.
3. `make fmt`: Mandatory Go formatting enforcement (`gofmt`).
4. `make lint`: Static analysis via `go vet ./...`.
5. `make test`: Full unit and package test execution.
6. `make test-integration`: Shell script and operational workflow validation.
7. `make openapi-check`: Verification that OpenAPI schemas and runtime routes have not drifted.
8. `make scalar-check`: Verification that Scalar UI configuration references valid documentation files.
9. `make module-consistency-check`: Verification that architectural layer boundaries are respected.

### Consequences

- **Positive**:
  - Catches broken contracts and architecture drift before merging to main.
  - High confidence in continuous deployments to production.
- **Negative / Trade-offs**:
  - Pull requests require passing all checks before merging, requiring discipline around OpenAPI synchronization.
