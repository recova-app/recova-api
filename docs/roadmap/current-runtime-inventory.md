---
title: Recova Backend Current Runtime Inventory
description: Inventaris runtime aktif backend Recova mencakup command operasional, workflow migrasi, quality gate, dan artefak deployment pada jalur Go Fiber.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/roadmap/current-runtime-inventory.md
last_reviewed: 2026-05-08
---

# Recova Backend Current Runtime Inventory

Dokumen ini mencatat kontrak runtime yang aktif pada repository saat ini agar operasi dan verifikasi berjalan konsisten.

## Runtime Baseline

- Language: Go.
- HTTP framework: Fiber.
- Data store: PostgreSQL.
- ORM: GORM.
- API prefix: `/api/v1`.

## Local Runtime Commands

| Mode                     | Command      | Catatan                                                                   |
| ------------------------ | ------------ | ------------------------------------------------------------------------- |
| Run API lokal            | `make run`   | Memuat env via `scripts/with-env.sh` lalu menjalankan `go run ./cmd/api`. |
| Build binary             | `make build` | Output binary: `./bin/recova-api`.                                        |
| Quality lint             | `make lint`  | Menjalankan `go vet ./...`.                                               |
| Unit + integration tests | `make test`  | Menjalankan `go test ./...`.                                              |

## Database Workflow Inventory

| Workflow                | Command                          | Tujuan                                  |
| ----------------------- | -------------------------------- | --------------------------------------- |
| Apply migration         | `make migrate-up`                | Menerapkan migration SQL terbaru.       |
| Rollback migration      | `make migrate-down`              | Rollback satu langkah migration.        |
| Migration status        | `make migrate-status`            | Melihat version migration aktif.        |
| Migration health check  | `make migrate-check`             | Validasi migration version tidak dirty. |
| Force migration version | `make migrate-force VERSION=<n>` | Recovery migration state terkontrol.    |
| Seed baseline data      | `make seed`                      | Menjalankan seed non-secret idempotent. |

## Contract and Release Validation Inventory

| Gate                      | Command                           | Output evidence                                         |
| ------------------------- | --------------------------------- | ------------------------------------------------------- | --- | --- | --- | ----- | ----------------------- |
| OpenAPI drift check       | `make openapi-check`              | Validasi route dan contract source.                     |
| Script integration checks | `make test-integration`           | Validasi runner shell operasional.                      |
| E2E critical flows        | `make test-e2e`                   | `artifacts/release-confidence/e2e-critical-flows.json`. |
| Performance smoke         | `make test-performance`           | `artifacts/release-confidence/performance-smoke.json`.  |
| Cutover verification      | `make cutover-wave WAVE=<64       | 65                                                      | 66  | 67  | 68  | all>` | `artifacts/cutover/**`. |
| Stabilization gate        | `make stabilization-gate`         | `artifacts/stabilization/**`.                           |
| Rollback rehearsal        | `make rollback-rehearsal`         | `artifacts/rollback-rehearsal/**`.                      |
| Runtime decommission      | `make runtime-decommission`       | `artifacts/decommission/**`.                            |
| Maintenance review        | `make post-migration-maintenance` | `artifacts/maintenance/**`.                             |

## Container Workflow Inventory

| Workflow            | Command               | Catatan                                              |
| ------------------- | --------------------- | ---------------------------------------------------- |
| Compose smoke       | `make compose-smoke`  | Validasi startup stack `api` + `db`.                 |
| Staging deploy gate | `make staging-deploy` | Migration dry-run, seed idempotency, readiness gate. |

## Active CI/CD Workflow Inventory

| Workflow file                                    | Purpose                                                                 |
| ------------------------------------------------ | ----------------------------------------------------------------------- |
| `.github/workflows/ci.yml`                       | quality, DB, security, compose smoke, image build, staging deploy gate. |
| `.github/workflows/cutover-waves.yml`            | cutover wave manual + evidence artifact.                                |
| `.github/workflows/stabilization-rollback.yml`   | stabilization gate dan rollback rehearsal manual.                       |
| `.github/workflows/decommission-maintenance.yml` | runtime decommission gate + maintenance review cadence.                 |

## Legacy Runtime Status

- Runtime lama berbasis Express tidak lagi menjadi runtime publik aktif.
- Artefak referensi runtime lama dipertahankan sebagai arsip pada direktori `references/`.
- Baseline historis runtime lama dicatat pada [Legacy Express Baseline](/Users/macbookpro/Development/recova-backend-v2/docs/roadmap/express-baseline.md).

## Related Documents

- [Deployment Workflow](/Users/macbookpro/Development/recova-backend-v2/docs/operations/deployment.md)
- [Release Gates](/Users/macbookpro/Development/recova-backend-v2/docs/operations/release-gates.md)
- [Migration Execution Runbook](/Users/macbookpro/Development/recova-backend-v2/docs/roadmap/migration-execution-runbook.md)
- [Documentation Maintenance Standard](/Users/macbookpro/Development/recova-backend-v2/docs/standards/documentation-maintenance.md)

## Source Reference

- [Makefile](/Users/macbookpro/Development/recova-backend-v2/Makefile)
- [scripts/](/Users/macbookpro/Development/recova-backend-v2/scripts)
- [CI workflow](/Users/macbookpro/Development/recova-backend-v2/.github/workflows/ci.yml)
