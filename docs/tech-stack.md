---
title: Recova Backend Tech Stack
description: Pilihan stack target backend Go, kategori package inti, risiko, dan prinsip seleksi dependency produksi.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/tech-stack.md
last_reviewed: 2026-05-08
---

# Recova Backend Tech Stack

Dokumen ini memformalkan pilihan stack backend target dan kriteria seleksi package produksi.

## Core Runtime Stack

| Area           | Target                        | Rationale                                                             |
| -------------- | ----------------------------- | --------------------------------------------------------------------- |
| Language       | Go `1.25+`                    | mengikuti requirement framework target dan tooling modern             |
| HTTP framework | Fiber `v3`                    | performa tinggi, middleware ekosistem matang, model handler sederhana |
| Database       | PostgreSQL                    | transactional consistency dan dukungan relasional kuat                |
| ORM            | GORM `v2`                     | produktivitas query + kontrol transaksi + ekosistem luas              |
| Migration      | golang-migrate `v4`           | SQL migration up/down yang eksplisit dan audit-friendly               |
| Validation     | go-playground/validator `v10` | validasi struct/tag mapan dan banyak dipakai                          |
| JWT            | golang-jwt/jwt `v5`           | library JWT Go yang maintained dan umum                               |
| Logging        | `zap` (default)               | structured logging performa tinggi                                    |
| API contract   | OpenAPI `3.1`                 | kontrak API machine-readable dan reviewable                           |
| Observability  | OpenTelemetry Go              | standar telemetry lintas vendor                                       |

## Supporting Tooling Direction

| Category   | Direction                                                      |
| ---------- | -------------------------------------------------------------- |
| Testing    | `go test` + test table pattern + integration tests berbasis DB |
| Linting    | `golangci-lint` sebagai quality gate                           |
| Formatting | `gofmt` dan `goimports`                                        |
| Container  | multi-stage Docker build                                       |

## Package Selection Criteria

Semua package final harus lolos kriteria:

- maintained aktif,
- dokumentasi resmi jelas,
- dukungan ekosistem stabil,
- bisa diuji secara deterministik,
- kompatibel dengan lisensi repository.

## Risk and Mitigation by Category

| Category       | Risiko                                        | Mitigasi                                    |
| -------------- | --------------------------------------------- | ------------------------------------------- |
| HTTP framework | lock-in middleware API                        | isolasi middleware setup di layer bootstrap |
| ORM            | query implicit dan kompleksitas eager-loading | repository pattern + query review           |
| Migration      | salah urut migrasi atau dirty state           | SQL up/down eksplisit + runbook recovery    |
| Validation     | mismatch schema request vs DTO                | central validator strategy + contract tests |
| JWT            | misconfiguration expiry/signing               | config fail-fast + auth tests               |
| Logging        | kebocoran data sensitif                       | redaction policy + forbidden field checks   |
| OpenAPI        | drift antara implementasi dan spec            | contract-first review + CI drift check      |

## Decision Links

- [ADR 0003 HTTP Framework Fiber](/Users/macbookpro/Development/recova-backend-v2/docs/decisions/adr-0003-http-framework-fiber.md)
- [ADR 0004 ORM GORM PostgreSQL](/Users/macbookpro/Development/recova-backend-v2/docs/decisions/adr-0004-orm-gorm-postgresql.md)
- [ADR 0005 Database Migration Tool](/Users/macbookpro/Development/recova-backend-v2/docs/decisions/adr-0005-database-migration-tool.md)

## Source Reference

- [Fiber v3 Documentation](https://docs.gofiber.io/)
- [GORM Documentation](https://gorm.io/docs/)
- [golang-migrate](https://github.com/golang-migrate/migrate)
- [PostgreSQL Current Documentation](https://www.postgresql.org/docs/current/index.html)
- [go-playground/validator v10](https://pkg.go.dev/github.com/go-playground/validator/v10)
- [golang-jwt/jwt v5](https://pkg.go.dev/github.com/golang-jwt/jwt/v5)
- [OpenAPI Specification](https://spec.openapis.org/oas/)
- [OpenTelemetry Go Documentation](https://opentelemetry.io/docs/languages/go/)
- [/Users/macbookpro/Development/bisakerja-api/docs/tech-stack.md](/Users/macbookpro/Development/bisakerja-api/docs/tech-stack.md)
