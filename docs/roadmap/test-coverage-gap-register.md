---
title: Recova Backend Test Coverage Gap Register
description: Register gap coverage dan checklist perluasan test lintas area kritikal (auth, users, routine, journals, community, content, AI, achievements, OpenAPI, dan operasi).
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/roadmap/test-coverage-gap-register.md
last_reviewed: 2026-05-12
---

# Test Coverage Gap Register

Dokumen ini memetakan baseline coverage dan gap test per area agar penambahan test terarah, terukur, dan tidak mengorbankan determinisme suite.

## Scope dan Definisi

- fokus metrik: statement coverage untuk package produksi `./internal/...`.
- suite yang butuh database (contoh E2E) dipisahkan dari baseline coverage default.
- angka coverage dipakai sebagai indikator, bukan pengganti verifikasi perilaku pada jalur kritikal.

## Cara Generate Coverage Baseline

Command:

- `make coverage`
- atau: `./scripts/coverage.sh`

Output:

- `artifacts/coverage/internal.out` (coverprofile)
- `artifacts/coverage/internal.func.txt` (ringkasan fungsi)
- `artifacts/coverage/by-area.json` (agregasi per area)
- `artifacts/coverage/by-area.md` (tabel ringkas)

Catatan:

- default memakai `go test -short` dan scope `./internal/...`.
- jika ingin include suite non-short, set `COVERAGE_SHORT=0` (mungkin butuh `RECOVA_DB_INTEGRATION_URL`).

## Coverage Snapshot (2026-05-12)

Baseline dari `artifacts/coverage/by-area.md`:

| area                     | percent |
| ------------------------ | ------- |
| modules/auth             | 65.4%   |
| modules/users            | 43.0%   |
| modules/routine          | 56.8%   |
| modules/journals         | 60.0%   |
| modules/community        | 49.6%   |
| modules/content          | 62.1%   |
| modules/education        | 62.9%   |
| modules/ai               | 59.9%   |
| modules/achievements     | 40.2%   |
| platform/database        | 43.9%   |
| platform/logger          | 92.9%   |
| platform/observability   | 57.2%   |
| platform/openapi         | 79.9%   |
| shared/errs              | 59.7%   |
| total (`./internal/...`) | 57.0%   |

## Inventory Test Suite (Entry Points)

Unit/validator/handler/service/repository:

- `internal/modules/*/*_test.go`
- `internal/platform/*/*_test.go`
- `internal/app/http/*_test.go`

Repository integration (butuh DB, auto-skip bila env tidak ada):

- `internal/platform/database/integration_test.go`
- `internal/modules/*/repository_integration_test.go`

Contract dan OpenAPI quality:

- `test/contract/openapi_contract_test.go`
- `internal/platform/openapi/*_test.go`
- `scripts/openapi.sh` + `make openapi-check`

E2E critical flows (butuh DB):

- `test/e2e/critical_flows_test.go`
- `scripts/e2e-critical.sh` + `make test-e2e`

Performance smoke (butuh DB):

- `test/performance/smoke_test.go`
- `scripts/performance-smoke.sh` + `make test-performance`

Module consistency harness:

- `test/contract/module_consistency_test.go`
- `scripts/module-consistency.sh` + `make module-consistency-check`

## Gap Matrix (Behavior-First)

| area                   | baseline | coverage sudah ada                                                    | gap prioritas                                                                                               | aksi berikutnya                                                                                             |
| ---------------------- | -------- | --------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------- |
| auth (manual)          | 65.4%    | validator + service + route: register/login, conflict, refresh/logout | tambah skenario refresh subject mismatch, refresh token missing, dan assert cookie flags pada token manager | tambah test unit/service + handler untuk refresh mismatch/missing; tambah test token manager cookie parsing |
| auth (google)          | 65.4%    | route + service: login token invalid/valid (via fake verifier)        | verifikasi mapping error downstream verifier dan anti-leak log token                                        | tambah test service untuk verifier timeout/downstream error mapping + redaction                             |
| users/me + onboarding  | 43.0%    | route unauthorized + service happy-path minimal                       | gap: validation edge-cases onboarding, not-found mapping, ownership checks                                  | tambah table-driven test validator + service untuk skenario invalid input dan not-found                     |
| routine/check-in/stats | 56.8%    | route + service: check-in/statistics dasar                            | gap: idempotency dan window-days validation edge                                                            | tambah test validator/service untuk window_days invalid dan idempotent check-in                             |
| journals               | 60.0%    | route unauthorized + create/list happy-path                           | gap: pagination/limit validation dan akses scope                                                            | tambah test handler/service untuk limit bounds + ownership                                                  |
| community              | 49.6%    | route + service: create/comment/reply/like                            | gap: moderation/ownership negative path                                                                     | tambah test untuk forbidden path (edit/delete by non-owner) bila endpoint tersedia                          |
| content/education      | 62.x%    | seed + read contract                                                  | gap: fallback behavior saat data kosong                                                                     | tambah test repository/service untuk empty table dan ordering                                               |
| AI coach               | 59.9%    | E2E safe response + telemetry basic                                   | gap: timeout/downstream error mapping detail                                                                | tambah unit test provider wrapper untuk timeout/cancel + redaction                                          |
| achievements           | 40.2%    | route + service baseline                                              | gap: unlock rules edge + repeat evaluation                                                                  | tambah unit test engine untuk threshold boundary dan replay safety                                          |
| OpenAPI quality        | 79.9%    | drift/path tests + contract validation                                | gap: tambah contoh contract untuk error codes penting di auth/users                                         | tambah contract tests (tanpa DB) untuk validation/unauthenticated/parity lebih luas                         |
| ops smoke              | n/a      | compose smoke + readiness                                             | gap: coverage deploy smoke scripts (non-Go)                                                                 | tambah scripted assertions pada `scripts/compose-smoke.sh` bila ada regresi ditemukan                       |

## Source Reference

- [Testing Strategy](/Users/macbookpro/Development/recova-backend-v2/docs/operations/testing.md)
- [Testing Conventions](/Users/macbookpro/Development/recova-backend-v2/docs/standards/testing-conventions.md)
- [CI/CD Operations](/Users/macbookpro/Development/recova-backend-v2/docs/operations/ci-cd.md)
