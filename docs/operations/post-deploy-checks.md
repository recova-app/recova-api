---
title: Recova Backend Post-Deploy Checks
description: Checklist verifikasi pasca-deploy untuk memastikan health layanan, kontrak API inti, serta indikator observability berada pada baseline aman.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/operations/post-deploy-checks.md
last_reviewed: 2026-05-09
---

# Recova Backend Post-Deploy Checks

Dokumen ini berisi checklist pasca-deploy yang wajib dijalankan sebelum rilis dinyatakan stabil.

## Execution Rule

- jalankan checks segera setelah rollout selesai,
- gunakan environment target yang sama dengan trafik user,
- simpan hasil check sebagai bukti rilis.

## Health and Readiness Checks

| Check               | Expected result                      |
| ------------------- | ------------------------------------ |
| `GET /health/live`  | `200` dan respons liveness sukses    |
| `GET /health/ready` | `200` dan dependency readiness sehat |

Jika readiness gagal:

- jangan lanjutkan promote release,
- evaluasi rollback atau forward-fix sesuai [Rollback Runbook](/Users/macbookpro/Development/recova-backend-v2/docs/operations/rollback.md).

## Core API Smoke Checks

Minimal endpoint yang diverifikasi:

- `POST /api/v1/auth/google`
- `GET /api/v1/users/me`
- `POST /api/v1/routine/checkin`
- `GET /api/v1/journals`
- `GET /api/v1/community`
- `GET /api/v1/education`
- `GET /api/v1/content/daily`
- `POST /api/v1/ai/ask-coach`

Aturan verifikasi:

- status code sesuai kontrak,
- error envelope tetap konsisten,
- tidak ada lonjakan `5xx` pada endpoint inti.

## Database and Migration Checks

- pastikan migration target version tercapai,
- cek tidak ada lock migrasi tertinggal,
- cek query error kritis setelah deploy.

Untuk staging berbasis Compose, verifikasi ini dijalankan otomatis lewat `scripts/staging-deploy.sh` sebelum release dipromosikan.

Untuk jalur cutover domain bertahap, verifikasi per-wave dieksekusi lewat `scripts/cutover-wave.sh` dengan evidence di `artifacts/cutover/`.

Untuk window stabilisasi pasca-cutover, jalankan `make stabilization-gate` dan simpan evidence di `artifacts/stabilization/`.

## Observability Checks

Verifikasi sinyal minimum:

- request latency `p95` tidak melewati threshold rilis,
- error rate total tidak melonjak signifikan,
- log panic/recover tidak menunjukkan spike,
- trace sampling tetap aktif sesuai konfigurasi.

## Security and Access Checks

- header keamanan tetap aktif,
- CORS policy sesuai allowlist environment,
- endpoint development-only tetap nonaktif di non-development.

## Decision Gate

| Kondisi                                                  | Keputusan                        |
| -------------------------------------------------------- | -------------------------------- |
| seluruh check lulus                                      | promote deployment               |
| check gagal tapi bisa diperbaiki cepat tanpa risiko data | hotfix terkontrol + ulang checks |
| check gagal dan berdampak ke availability/kontrak        | rollback sesuai runbook          |

## Evidence to Record

- waktu dan environment deploy,
- artifact id,
- commit SHA source deploy,
- image tag immutable (`sha-<commit-sha>`),
- migration version hasil `migrate version/check`,
- hasil tiap check,
- keputusan akhir (promote/hotfix/rollback),
- PIC verifikasi.

## Related Documents

- [Deployment Workflow](/Users/macbookpro/Development/recova-backend-v2/docs/operations/deployment.md)
- [Rollback Runbook](/Users/macbookpro/Development/recova-backend-v2/docs/operations/rollback.md)
- [Health and Readiness](/Users/macbookpro/Development/recova-backend-v2/docs/operations/health-readiness.md)
- [Observability](/Users/macbookpro/Development/recova-backend-v2/docs/operations/observability.md)

## Source Reference

- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
- [/Users/macbookpro/Development/bisakerja-api/docs/operations/deployment.md](/Users/macbookpro/Development/bisakerja-api/docs/operations/deployment.md)
- [Fiber Health Check Middleware](https://docs.gofiber.io/middleware/healthcheck/)
- [Fiber Recover Middleware](https://docs.gofiber.io/middleware/recover/)
- [OpenTelemetry Go](https://opentelemetry.io/docs/languages/go/)
