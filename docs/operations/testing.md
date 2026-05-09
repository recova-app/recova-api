---
title: Recova Backend Testing Strategy
description: Strategi pengujian lintas layer untuk API Recova Backend meliputi unit, handler, repository, contract, migration, smoke, dan end-to-end critical flow.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/operations/testing.md
last_reviewed: 2026-05-09
---

# Recova Backend Testing Strategy

Dokumen ini mendefinisikan baseline testing agar perubahan layanan aman sebelum rilis.

## Testing Goals

- memverifikasi kontrak API tetap stabil,
- memastikan aturan bisnis tetap benar,
- mencegah regresi auth/ownership/security,
- menjamin migration database aman dijalankan.

## Test-by-Default Rule

- setiap file implementasi baru wajib punya test companion pada layer yang relevan,
- setiap perubahan konfigurasi runtime/deploy/CI wajib punya verifikasi otomatis (unit/integration/smoke/scripted check),
- jika test dianggap tidak perlu, alasan teknis wajib dicatat eksplisit dan disetujui reviewer.

## Test Pyramid

| Layer                        | Fokus                            | Target utama                     |
| ---------------------------- | -------------------------------- | -------------------------------- |
| Unit tests                   | service rules, validator, mapper | feedback cepat dan deterministik |
| Handler tests                | route contract, status, envelope | validasi HTTP behavior           |
| Repository integration tests | query DB nyata + constraint      | integritas persistence           |
| Contract tests               | kompatibilitas response API      | deteksi drift kontrak            |
| Smoke tests                  | health/readiness dan startup     | gate runtime dasar               |
| E2E critical flow            | alur pengguna inti               | confidence release               |

## Coverage Baseline per Area

- auth: token validation, refresh rotation, unauthorized path,
- users/onboarding: ownership dan update validation,
- routine/check-in/streak: idempotency dan statistik,
- journals: access scope + privacy behavior,
- community: ownership/moderation rules,
- education/content: read contract dan fallback data,
- AI coach: timeout, error mapping, safety redaction.

## Enhancement Integration Scenarios

Skenario minimum pada suite release confidence:

- `make test-e2e` wajib mencakup statistik `activity-summary`, thread komentar + reply, achievements, dan persona preference AI.
- `make test-performance` wajib mencakup load smoke endpoint enhancement read-path (`activity-summary`, thread comments read, achievements catalog, persona preference read).
- hasil kedua suite wajib menghasilkan artefak report yang dapat diaudit.

## Database and Migration Verification

Wajib:

- jalankan migration pada database kosong,
- verifikasi rollback minimal satu langkah,
- pastikan query kritis lolos integration tests,
- pastikan tidak ada schema drift tidak terdokumentasi.

## Compatibility Verification

- response envelope harus konsisten,
- status code/error code harus sesuai standar,
- endpoint compatibility prioritas diuji melalui contract tests,
- perubahan breaking harus punya approval eksplisit.

## Test Data Rules

- gunakan data sintetis,
- jangan gunakan credential nyata,
- data sensitif harus disamarkan,
- fixture harus versioned dan repeatable.

## Baseline Verification Commands

Perintah verifikasi baseline:

| Command                     | Scope                                                                 |
| --------------------------- | --------------------------------------------------------------------- |
| `make test`                 | unit tests package Go                                                 |
| `make lint`                 | static checks baseline (`go vet ./...`)                               |
| `make build`                | compile gate untuk binary API                                         |
| `make preflight`            | validasi dependency command dan struktur minimum project              |
| `make test-integration`     | scripted checks untuk workflow tooling (mis. migrasi wrapper)         |
| `make migrate-check`        | validasi state migration (versi/dirtiness)                            |
| `make openapi-check`        | validasi OpenAPI source/generated + drift route runtime               |
| `make security-scan`        | vulnerability scan dependency Go via govulncheck                      |
| `make compose-smoke`        | smoke runtime container (`docker-compose.local.yml`)                  |
| `make test-e2e`             | E2E critical flow suite + report JSON                                 |
| `make test-performance`     | load/performance smoke suite + report JSON                            |
| `make release-validation`   | menjalankan seluruh gate release-validation (E2E + performance smoke) |
| `make cutover-wave WAVE=64` | gate otomatis cutover per wave (single wave)                          |
| `make cutover-all`          | gate otomatis cutover serial wave 64-68                               |
| `make stabilization-gate`   | full regression + openapi + E2E + performance smoke + evidence report |
| `make rollback-rehearsal`   | rehearsal jalur rollback dengan failure injection terkontrol          |

Default output report:

- `artifacts/release-confidence/e2e-critical-flows.json`
- `artifacts/release-confidence/performance-smoke.json`
- `artifacts/stabilization/*-stabilization-report.json`
- `artifacts/rollback-rehearsal/*-rollback-rehearsal-report.json`

## Release Readiness Testing Gate

Sebelum rilis:

1. unit + handler tests lulus,
2. integration DB + migration checks lulus,
3. contract tests kompatibilitas lulus,
4. smoke tests readiness lulus,
5. critical E2E flows lulus.

Tambahan gate:

6. semua perubahan file/config dalam scope rilis memiliki test companion atau exception rationale terdokumentasi.
7. report release confidence E2E + performance tersedia dan dapat diaudit.

## Related Documents

- [Verification Matrix](/Users/macbookpro/Development/recova-backend-v2/docs/operations/verification-matrix.md)
- [Compatibility Test Plan](/Users/macbookpro/Development/recova-backend-v2/docs/roadmap/compatibility-test-plan.md)
- [API Response Standard](/Users/macbookpro/Development/recova-backend-v2/docs/api-response-standard.md)
- [Release Confidence Report](/Users/macbookpro/Development/recova-backend-v2/docs/generated/release-confidence-report.md)

## Source Reference

- [Go Testing Package](https://pkg.go.dev/testing)
- [go Command Documentation](https://pkg.go.dev/cmd/go)
- [PostgreSQL Current Documentation](https://www.postgresql.org/docs/current/)
- [Fiber App API](https://docs.gofiber.io/api/app/)
