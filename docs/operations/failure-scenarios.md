---
title: Recova Backend Failure Scenarios
description: Katalog skenario kegagalan layanan beserta signal observability, dampak bisnis, pemeriksaan awal, dan aksi mitigasi pertama.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/operations/failure-scenarios.md
last_reviewed: 2026-05-08
---

# Recova Backend Failure Scenarios

Dokumen ini mendeskripsikan skenario gagal utama untuk mempercepat deteksi dan respons awal.

## Scenario Matrix

| Scenario                    | Primary signal                        | User impact                             | First checks                                          | First action                                                    |
| --------------------------- | ------------------------------------- | --------------------------------------- | ----------------------------------------------------- | --------------------------------------------------------------- |
| Database unavailable        | readiness `503`, DB timeout spike     | hampir semua endpoint gagal             | DB connectivity, pool saturation, credential validity | aktifkan incident channel, batasi deploy, pulihkan konektivitas |
| Migration mismatch          | startup error/migration failure logs  | service tidak siap menerima trafik      | migration version, dirty state, schema drift          | hentikan deploy, jalankan recovery migration plan               |
| Auth token validation spike | `401` spike + auth error logs         | pengguna gagal login/akses route privat | JWT issuer/audience/secret mismatch, clock skew       | verifikasi config auth, rollback konfigurasi yang berubah       |
| Rate limit storm            | `429` spike pada endpoint tertentu    | request ditolak masif                   | key distribution, bot traffic, limiter policy         | aktifkan mitigasi traffic + tuning limiter terkontrol           |
| Panic burst                 | recover panic count naik              | error `500` melonjak                    | stack signature aggregate, release diff, hot path     | aktifkan rollback release jika impact meluas                    |
| AI provider degraded        | timeout/downstream error naik         | fitur AI lambat/gagal                   | provider status, timeout config, request volume       | aktifkan fallback policy atau degrade graceful                  |
| CORS misconfiguration       | error browser preflight meningkat     | klien web gagal akses API               | origin allowlist, credentials setting, env drift      | koreksi origin policy dan redeploy konfigurasi                  |
| Secret rotation mismatch    | auth/downstream errors setelah rotasi | endpoint terkait gagal                  | secret version sync antar komponen                    | rollback/rollforward secret dengan prosedur terkontrol          |

## Classification Rule

- **P1**: core API tidak tersedia untuk mayoritas user.
- **P2**: fungsi inti terganggu signifikan, sebagian user gagal.
- **P3**: fungsi non-kritis atau terbatas pada subset endpoint.

## Safety Rule During Incident

- dilarang mengaktifkan logging raw payload sensitif,
- perubahan hotfix harus minimal dan ter-review,
- setiap tindakan harus punya timestamp dan owner.

## Related Documents

- [Observability](/Users/macbookpro/Development/recova-backend-v2/docs/operations/observability.md)
- [Incident Triage](/Users/macbookpro/Development/recova-backend-v2/docs/operations/incident-triage.md)
- [Database Migrations](/Users/macbookpro/Development/recova-backend-v2/docs/operations/database-migrations.md)

## Source Reference

- [Fiber Recover Middleware](https://docs.gofiber.io/middleware/recover/)
- [PostgreSQL Current Documentation](https://www.postgresql.org/docs/current/)
- [OpenTelemetry Go Instrumentation](https://opentelemetry.io/docs/languages/go/instrumentation/)
