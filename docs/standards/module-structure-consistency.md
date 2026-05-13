---
title: Recova Backend Module Structure Consistency Standard
description: Standar konsistensi anatomi folder module di internal/modules agar boundary layer, naming, dan test companion tetap seragam lintas domain.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/standards/module-structure-consistency.md
last_reviewed: 2026-05-09
---

# Recova Backend Module Structure Consistency Standard

Dokumen ini mendefinisikan baseline anatomi module domain di `internal/modules` dan cara menangani deviasi secara eksplisit.

## Module Inventory Baseline

Subfolder domain aktif yang wajib diaudit:

- `achievements`
- `ai`
- `auth`
- `community`
- `content`
- `education`
- `journals`
- `routine`
- `users`

## Required Anatomy

File baseline per module:

- `doc.go`
- `dto.go`
- `handler.go`
- `repository.go`
- `route.go`
- `service.go`
- `validator.go`

Kaidah layer:

- `handler` hanya orkestrasi HTTP + binding payload.
- `service` hanya business rule.
- `repository` hanya persistence access.
- `validator` hanya normalisasi/validasi input.

## Companion Test Baseline

Companion test minimum:

- `route.go` -> `route_test.go`
- `service.go` -> `service_test.go`
- `validator.go` -> `validator_test.go` (jika `validator.go` ada)
- `repository.go` -> `repository_test.go` atau `repository_integration_test.go`

## Naming and Registration Baseline

- constructor layer memakai pola `NewHandler`, `NewService`, `NewRepository`.
- mayoritas module memakai `RegisterRoutes`.
- exception yang sah:
  - `auth` memakai `RegisterCoreRoutes`.
  - `users` memakai `RegisterUserRoutes` dan `RegisterOnboardingRoute`.

## Accepted Exceptions Register

Deviasi yang diizinkan saat ini:

| Module      | Gap                                             | Status               | Owner             | Reason                                                                        |
| ----------- | ----------------------------------------------- | -------------------- | ----------------- | ----------------------------------------------------------------------------- |
| `content`   | `validator.go` belum ada                        | `accepted-exception` | `content-owner`   | Endpoint read-only tanpa payload tulis yang butuh validator khusus.           |
| `education` | `validator.go` belum ada                        | `accepted-exception` | `education-owner` | Endpoint read-only tanpa payload tulis yang butuh validator khusus.           |
| `content`   | test companion repository belum spesifik module | `accepted-exception` | `content-owner`   | Coverage persistence ditopang `seed_integration_test.go` lintas jalur konten. |

## Enforcement

Gate konsistensi wajib:

- `make module-consistency-check`
- `go test -count=1 ./test/contract -run '^TestContract_Module.*Consistency'`

Aturan:

- gap baru yang tidak ada di register harus gagal di gate.
- gap yang sudah ditutup wajib dihapus dari register.
- status gap harus eksplisit: `accepted-exception` atau `needs-follow-up`.

## Related Documents

- [Project Structure](/Users/macbookpro/Development/recova-backend-v2/docs/project-structure.md)
- [Import Boundaries Standard](/Users/macbookpro/Development/recova-backend-v2/docs/standards/import-boundaries.md)
- [Testing Conventions](/Users/macbookpro/Development/recova-backend-v2/docs/standards/testing-conventions.md)
- [Testing Strategy](/Users/macbookpro/Development/recova-backend-v2/docs/operations/testing.md)
- [Internal Modules Consistency Audit](/Users/macbookpro/Development/recova-backend-v2/docs/roadmap/internal-modules-consistency-audit.md)
- [Module Consistency Cleanup Backlog](/Users/macbookpro/Development/recova-backend-v2/docs/roadmap/module-consistency-cleanup-backlog.md)

## Source Reference

- [Organizing a Go module](https://go.dev/doc/modules/layout)
- [Package testing](https://go.dev/pkg/testing/?m=old)
- [Fiber Grouping](https://docs.gofiber.io/guide/grouping)
