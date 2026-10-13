# Recova Backend Overview

## Introduction

Recova Backend is the core API service powering the Recova addiction recovery platform. The application is designed to support individuals overcoming behavioral and digital addictions through structured daily check-ins, routine tracking, relapse management, private journaling, community support, educational resources, and AI-driven coaching.

The backend is built as a high-performance, modular monolithic RESTful API in Go using the Fiber framework, PostgreSQL for durable persistence, and GORM as the object-relational mapper.

## System Mission & Core Value

The system provides reliable, low-latency services tailored for mobile clients (primarily Flutter):

- **Account & Session Security**: Seamless Google OAuth and manual authentication with short-lived JWT access tokens and secure rotating refresh tokens.
- **Recovery Tracking**: Resilient daily morning/evening check-ins, streak calculations, relapse logging, and aggregated progress statistics.
- **Supportive Community**: Moderated public discussions, threaded comments, reactions, and peer encouragement.
- **Reflective Journaling**: Private, mood-tagged journaling tied directly to recovery check-ins.
- **Educational & Daily Content**: Curated recovery literature, daily affirmations, and daily mental and physical challenges.
- **AI Coach**: Context-aware, empathetic AI companion supporting users through urges, onboarding analysis, and personalized guidance.

## Architecture Highlights

```text
[ Mobile App (Flutter) ]
           │
           ▼
[ Fiber HTTP Router & Middleware ]
           │
           ▼
[ Domain Modules (internal/modules/*) ]
   ├── auth          ├── community
   ├── users         ├── education
   ├── routine       ├── content
   ├── journals      ├── ai
   └── achievements
           │
     ┌─────┴──────────────┐
     ▼                    ▼
[ PostgreSQL ]    [ External Providers ]
                  ├── Google OAuth API
                  └── Google Gemini / OpenAI Compatible API
```

Key technical characteristics:

- **Language & Runtime**: Go 1.26+ with toolchain pinning.
- **HTTP Engine**: Fiber v3 with zero memory allocation routing and standard middleware.
- **Database**: PostgreSQL 16+ accessed via GORM, managed by raw SQL migrations using `golang-migrate`.
- **API Specification**: OpenAPI 3.1 contract synchronized with code and rendered via Scalar UI at `/docs/api`.
- **Observability**: Prometheus metrics at `/metrics`, structured JSON logging with request ID correlation, and health probes at `/health/live` and `/health/ready`.

## Documentation Structure

The documentation set is structured into focused, authoritative guides:

| Document                           | Purpose                                                                                  |
| ---------------------------------- | ---------------------------------------------------------------------------------------- |
| [ARCHITECTURE.md](ARCHITECTURE.md) | Runtime topology, layered architecture, module design, and cross-cutting concerns        |
| [API.md](API.md)                   | Standard response envelopes, error taxonomy, route catalog, authentication, and OpenAPI  |
| [DATABASE.md](DATABASE.md)         | PostgreSQL schema, domain entities, relationships, indexing, migrations, and seeding     |
| [ENVIRONMENT.md](ENVIRONMENT.md)   | Configuration loading, environment variables reference, TLS rules, and security policies |
| [MODULES.md](MODULES.md)           | Deep dive into each domain module: responsibilities, services, models, and endpoints     |
| [OPERATIONS.md](OPERATIONS.md)     | Local development, testing suites, containerization, deployment, and health monitoring   |
| [ADR.md](ADR.md)                   | Architecture Decision Records documenting foundational technical choices and trade-offs  |

## Core Engineering Principles

1. **Strict Layered Boundaries**: Requests flow unidirectionally: Router → Handler → Service → Repository → Database. Handlers never query databases; services never manipulate raw HTTP responses.
2. **Fail-Fast Configuration**: All required environment variables and security constraints are verified at startup. Missing or invalid configurations immediately abort application boot.
3. **Data Privacy & Integrity**: Sensitive user data (journals, AI conversations, relapse logs) is strictly protected. Passwords and refresh tokens are cryptographically hashed; secrets are never logged.
4. **Resilient Integrations**: Failures in external downstream providers (such as AI services) are contained gracefully without disrupting core non-dependent recovery flows.
5. **Contract Parity**: API contracts and runtime routes are rigorously validated via automated contract tests and OpenAPI generation.
