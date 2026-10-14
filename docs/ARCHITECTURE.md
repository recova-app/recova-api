# Recova Backend Architecture

## Overview

Recova Backend is architected as a modular monolith written in Go using the Fiber v3 HTTP web framework. The service is designed for high concurrency, low latency, predictable resource consumption, and clear architectural boundaries between domain modules.

## Runtime Topology

```text
               Mobile Client (Flutter)
                         │
                         ▼ HTTPS
          [ Ingress / Reverse Proxy ]
                         │
                         ▼
        ┌──────────────────────────────────┐
        │       Recova Backend (Go)        │
        │                                  │
        │  Fiber Router & Middleware Pipeline│
        │  Domain Modules (internal/modules)│
        │  Platform Adapters               │
        └───────┬──────────────┬───────────┘
                │              │
       PostgreSQL 16     External APIs
       (Persistence)     ├── Google OAuth
                         └── AI Provider (Gemini / OpenAI)
```

### Runtime Topology Rules

- **Single Entry Point**: Mobile clients communicate solely with the Recova Backend API. Clients never access PostgreSQL or external cloud services directly.
- **Authoritative Gateway**: The backend is the single source of truth for authentication, authorization, business validation, streak calculations, and data persistence.
- **External Dependencies**: External third-party providers (Google OAuth, Gemini, OpenAI) are accessed through platform adapters with strict timeouts and circuit breaking to prevent cascading failures.

## Client Context & Contract Rules

The primary consumer of this backend is a mobile application developed in Flutter. This imposes specific architectural constraints:

- **Network Resilience**: Payloads are kept compact, concise, and deterministic to perform reliably over mobile data connections.
- **Consistent Response Envelopes**: All endpoints adhere to a standardized JSON response envelope (`success`, `message`, `data`, `meta`, `error`).
- **Language Boundaries**:
  - Technical machine-readable identifiers (field names, enum keys, `error.code`) are strictly English uppercase/snake_case.
  - End-user facing messages (`message`, `error.details` descriptions) are in Indonesian to maintain consistency with the client application UI.
  - Internal logging, error traces, and developer documentation are strictly English.

## Layered Architecture

Within each domain module, code is organized into distinct unidirectional layers:

```text
[ HTTP Router & Middleware ]
             │
             ▼
    [ Handler / Controller ]
             │
             ▼
        [ Service ]
             │
             ▼
       [ Repository ]
             │
             ▼
       [ PostgreSQL ]
```

### Layer Responsibilities and Constraints

| Layer          | Primary Responsibilities                                                               | Strict Constraints                                                       |
| -------------- | -------------------------------------------------------------------------------------- | ------------------------------------------------------------------------ |
| **Router**     | URL path mapping, HTTP method binding, module middleware attachment                    | No business logic; no database queries                                   |
| **Middleware** | Request ID injection, CORS, recovery, rate limiting, authentication guards             | Must not perform heavy domain computation                                |
| **Handler**    | HTTP request parsing, payload unmarshaling, input validation, HTTP response formatting | Must not query the database directly; must not contain business rules    |
| **Service**    | Domain logic, business rule validation, authorization rules, transaction orchestration | Must not import Fiber or access HTTP request/response contexts directly  |
| **Repository** | Data persistence, SQL/GORM queries, transaction execution, entity mapping              | Must not perform HTTP response formatting; must not define API contracts |
| **Database**   | ACID transactional storage, schema constraints, indexes, foreign keys                  | Direct access strictly encapsulated by repositories                      |

## Project Layout

The repository follows standard Go project conventions:

```text
.
├── cmd/
│   ├── api/                     # Service entrypoint (main.go)
│   └── tools/                   # CLI utilities (openapi, scalar, coverage)
├── internal/
│   ├── app/                     # Application bootstrap, routing, lifecycle
│   │   ├── bootstrap/           # Dependency injection and service assembly
│   │   ├── http/                # Server setup, middleware pipeline, docs routes
│   │   └── lifecycle/           # Graceful startup and shutdown signals
│   ├── modules/                 # Self-contained domain modules
│   │   ├── achievements/        # Milestone and achievement tracking
│   │   ├── ai/                  # AI Coach conversations and analysis
│   │   ├── auth/                # Authentication, token generation, OAuth
│   │   ├── community/           # Forum posts, threaded comments, reactions
│   │   ├── content/             # Daily affirmations and challenges
│   │   ├── education/           # Educational recovery articles
│   │   ├── journals/            # Personal reflective journals
│   │   ├── routine/             # Daily check-ins, streaks, relapse tracking
│   │   └── users/               # User profiles, goals, onboarding
│   ├── platform/                # Infrastructure adapters
│   │   ├── ai/                  # AI client implementations (Gemini, OpenAI)
│   │   ├── config/              # Fail-fast configuration loader and validator
│   │   ├── database/            # GORM initialization and connection pooling
│   │   ├── logger/              # Structured logging adapter
│   │   ├── observability/       # Prometheus metrics collector
│   │   └── openapi/             # OpenAPI artifact validation
│   └── shared/                  # Cross-cutting utilities
│       ├── errs/                # Standard domain and application errors
│       ├── response/            # Standard JSON response envelope builders
│       ├── types/               # Common primitive types and paginators
│       └── util/                # Hash helpers, string validators, time utilities
├── api/
│   └── openapi/                 # Source OpenAPI 3.1 schema specification
├── migrations/                  # Versioned SQL migration files (up/down)
├── scripts/                     # Operational, testing, and CI automation scripts
├── test/                        # Automated test suites
│   ├── contract/                # API contract and route consistency tests
│   ├── e2e/                     # End-to-end critical flow tests
│   ├── fixtures/                # Reusable test fixtures and seeds
│   ├── harness/                 # In-memory test harnesses (HTTP, DB)
│   └── performance/             # Baseline performance smoke tests
└── docs/                        # Authoritative technical documentation
    └── generated/               # Generated OpenAPI artifacts and route inventories
```

## Cross-Cutting Concerns

### 1. Centralized Error Handling

All unhandled panics and errors bubble up through Fiber's custom error handler located in `internal/app/http/error_handler.go`.

- Domain errors map to specific HTTP status codes and stable machine-readable codes (e.g. `VALIDATION_ERROR`, `UNAUTHENTICATED`, `NOT_FOUND`).
- Unexpected system panics are recovered, logged with full stack traces internally, and returned to clients as a safe, sanitized `500 INTERNAL_ERROR` envelope without exposing internal implementation details.

### 2. Request Correlation & Observability

- Every incoming HTTP request is assigned a unique `X-Request-ID` via middleware (using incoming client header or generating a new UUID).
- The request ID is stored in the Fiber local context and attached to all structured log statements and response error envelopes.
- Prometheus metrics are recorded via middleware and exposed at `/metrics` (request counts, latency histograms, database connection pool gauges, AI provider latency and fallback counters, status code distributions).

### 3. Graceful Shutdown & Application Lifecycle

The service listens for OS termination signals (`SIGINT`, `SIGTERM`):

1. Stop accepting new HTTP requests on the network listener.
2. Allow active in-flight requests to complete within a bounded grace period (10 seconds).
3. Drain and safely close database connection pools.
4. Flush pending telemetry and log buffers before terminating.

### 4. Security Hardening

- **Security Headers**: Standard defense-in-depth headers applied across all responses (`X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `X-XSS-Protection: 1; mode=block`).
- **CORS**: Configurable domain allowlist restricting web and mobile origins.
- **Rate Limiting**: Tiered rate limits applied to sensitive routes (stricter limits on auth and AI endpoints; relaxed limits on general read endpoints).

## Failure Containment

- **AI Provider Outages**: External AI provider latency or downtime is isolated via request timeouts (default 10s), composite deadline budgets (allocating 70% to primary and reserving 30% for fallback), and automated fallback metrics telemetry. AI failures return structured `DOWNSTREAM_ERROR` envelopes and never crash the server or block unrelated database transactions.
- **Database Degradation**: When PostgreSQL becomes unreachable, the readiness probe (`/health/ready`) immediately reports `503 Service Unavailable`, preventing the ingress load balancer from routing traffic to an unhealthy node while allowing the process to attempt recovery.
- **Validation Gating**: Strict DTO binding and validation occur in the handler layer before triggering any service logic or database queries, shedding invalid load early.

## Health and Readiness Probes

The backend exposes standard cloud-native health endpoints:

- `GET /health/live`: Liveness probe. Returns HTTP `200` if the Fiber process is running and able to handle HTTP events.
- `GET /health/ready`: Readiness probe. Pings PostgreSQL via `db.PingContext()`. Returns HTTP `200` with status `ready` if the database responds, or HTTP `503` with status `unhealthy` if database connectivity is lost.
