---
title: Recova Backend Project Structure
description: Struktur repository Go, anatomi modul, dan arah dependensi untuk menjaga boundary layer tetap konsisten.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/project-structure.md
last_reviewed: 2026-05-09
---

# Recova Backend Project Structure

Dokumen ini menetapkan layout repository target untuk backend Go agar modul domain, platform layer, dan operasi tetap terpisah jelas.

## Target Layout

```text
.
|-- cmd/
|   `-- api/
|       `-- main.go
|
|-- internal/
|   |-- app/
|   |   |-- bootstrap/
|   |   |-- http/
|   |   `-- lifecycle/
|   |
|   |-- modules/
|   |   |-- auth/
|   |   |-- users/
|   |   |-- routine/
|   |   |-- journals/
|   |   |-- community/
|   |   |-- education/
|   |   |-- content/
|   |   `-- ai/
|   |
|   |-- platform/
|   |   |-- config/
|   |   |-- logger/
|   |   |-- database/
|   |   `-- observability/
|   |
|   `-- shared/
|       |-- errs/
|       |-- response/
|       |-- types/
|       `-- util/
|
|-- api/
|   `-- openapi/
|
|-- migrations/
|-- docker-compose.staging.yml
|-- scripts/
|-- bin/ (generated local build artifact, ignored by git)
|-- test/
|   |-- unit/
|   |-- integration/
|   `-- contract/
`-- docs/
```

## Directory Responsibilities

| Path                         | Responsibility                                                  |
| ---------------------------- | --------------------------------------------------------------- |
| `cmd/api`                    | entrypoint runtime, wiring bootstrap, startup/shutdown          |
| `internal/app`               | assembly aplikasi dan lifecycle service                         |
| `internal/modules`           | domain modules dan business use-cases                           |
| `internal/platform`          | adapter infrastruktur (config, db, logger, observability)       |
| `internal/shared`            | komponen lintas-modul yang bebas ketergantungan domain spesifik |
| `api/openapi`                | kontrak OpenAPI sumber kebenaran API                            |
| `migrations`                 | SQL migration files up/down                                     |
| `docker-compose.staging.yml` | compose runtime deploy staging production-style                 |
| `bin`                        | output binary lokal hasil command build                         |
| `test`                       | test suite unit/integration/contract                            |

## Module Anatomy

Setiap modul domain mengikuti pola:

```text
internal/modules/<module>/
  handler.go
  service.go
  repository.go
  model.go
  dto.go
  mapper.go
  validator.go
  route.go
```

Aturan:

- `handler` menerima HTTP input tervalidasi dan memanggil service.
- `service` memegang aturan bisnis dan orkestrasi.
- `repository` memegang akses persistence.
- `mapper` hanya transformasi data.
- `validator` hanya validasi input domain atau DTO.

## Dependency Direction

Arah dependensi yang diizinkan:

```text
route -> handler -> service -> repository -> database
service -> platform adapters (jika perlu)
```

Larangan:

- `cmd` mengandung domain logic.
- `handler` mengakses database langsung.
- `repository` melakukan response formatting.
- modul domain mengimpor file private dari modul domain lain.

## `cmd` Rule

`cmd/api/main.go` hanya boleh berisi:

- load konfigurasi,
- bootstrap dependency container,
- start HTTP server,
- graceful shutdown.

`cmd` tidak boleh memuat:

- rules domain,
- query database,
- mapping response API.

## Naming Rules

| Item              | Rule                 | Example                       |
| ----------------- | -------------------- | ----------------------------- |
| Directory         | lowercase            | `internal/modules/routine`    |
| File              | role-based lowercase | `service.go`, `repository.go` |
| Exported type     | PascalCase           | `RoutineService`              |
| Unexported helper | camelCase            | `mapCheckinRequest`           |

## Related Documents

- [Import Boundaries Standard](/Users/macbookpro/Development/recova-backend-v2/docs/standards/import-boundaries.md)
- [Architecture](/Users/macbookpro/Development/recova-backend-v2/docs/architecture.md)
- [Tech Stack](/Users/macbookpro/Development/recova-backend-v2/docs/tech-stack.md)
- [ADR 0002 Go Project Layout](/Users/macbookpro/Development/recova-backend-v2/docs/decisions/adr-0002-go-project-layout.md)

## Source Reference

- [Go Module Layout](https://go.dev/doc/modules/layout)
- [golang-standards/project-layout](https://github.com/golang-standards/project-layout)
- [/Users/macbookpro/Development/bisakerja-api/docs/project-structure.md](/Users/macbookpro/Development/bisakerja-api/docs/project-structure.md)
