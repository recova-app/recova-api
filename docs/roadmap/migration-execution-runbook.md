---
title: Recova Backend Migration Execution Runbook
description: Runbook eksekusi migrasi backend dengan urutan kerja, gate kompatibilitas, strategi coexistence, dan kontrol keputusan cutover.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/roadmap/migration-execution-runbook.md
last_reviewed: 2026-05-08
---

# Recova Backend Migration Execution Runbook

Dokumen ini menjadi panduan operasional perpindahan layanan ke runtime target dengan kontrak API tetap kompatibel.

## Runbook Goals

- menjaga stabilitas kontrak `/api/v1`,
- meminimalkan blast radius dengan cutover bertahap,
- memastikan rollback cepat ketika gate gagal.

## Roles and Ownership

| Role               | Tanggung jawab                                            |
| ------------------ | --------------------------------------------------------- |
| Migration owner    | memimpin keputusan go/no-go per domain                    |
| API contract owner | validasi parity request/response                          |
| Database owner     | validasi compat migration + rollback direction            |
| Operations owner   | validasi deployment, observability, dan incident response |

## Execution Waves

Urutan domain disarankan berdasarkan dampak dan ketergantungan:

1. health and infrastructure routes,
2. auth and users,
3. routine, streak, statistics,
4. journals,
5. community,
6. education and daily content,
7. AI coach integration.

Domain berikutnya hanya boleh lanjut jika domain sebelumnya lulus gate.

Mapping wave cutover runtime:

- wave 64: platform/health,
- wave 65: auth/users/onboarding,
- wave 66: routine/journals/statistics,
- wave 67: community/content,
- wave 68: ai coach.

Runner lokal/CI:

- `./scripts/cutover-wave.sh 64|65|66|67|68|all`.

## Preconditions Before First Cutover

- [Compatibility Matrix](/Users/macbookpro/Development/recova-backend-v2/docs/roadmap/compatibility-matrix.md) tervalidasi untuk domain target,
- [Compatibility Test Plan](/Users/macbookpro/Development/recova-backend-v2/docs/roadmap/compatibility-test-plan.md) executable,
- kebijakan deploy dan rollback aktif,
- observability minimum aktif: request id, structured logs, health/readiness checks.

## Standard Run Sequence Per Domain

```text
1) Freeze perubahan kontrak domain
2) Build and deploy candidate runtime
3) Jalankan contract parity tests
4) Jalankan smoke tests domain
5) Aktifkan cutover trafik terkontrol
6) Pantau error/latency/mismatch
7) Tetapkan go-forward atau rollback
```

## Compatibility Gate

Domain dianggap `pass` bila:

- endpoint existence parity lulus,
- method dan status code parity lulus,
- authn/authz parity lulus,
- error envelope parity lulus,
- tidak ada anomali kritis pada metrik pasca cutover.

## Coexistence Control Rules

- satu domain write-path hanya punya satu owner runtime aktif,
- perubahan schema harus backward-compatible untuk runtime yang masih hidup,
- dual-write hanya boleh bila idempotensi dan conflict policy sudah terdokumentasi.

## Data Migration Procedure

Prinsip umum:

- utamakan pola expand-then-contract,
- hindari migrasi destruktif sebelum tidak ada dependency runtime lama,
- wajib punya rollback direction atau forward-fix plan untuk perubahan berisiko.

Detail operasional migrasi ada di:

- [Database Migrations](/Users/macbookpro/Development/recova-backend-v2/docs/operations/database-migrations.md)
- [Rollback Checklist](/Users/macbookpro/Development/recova-backend-v2/docs/roadmap/rollback-checklist.md)

## Parallel Verification Option

Saat domain berisiko tinggi, gunakan verifikasi paralel:

- mirror request sampling untuk membandingkan respons antar runtime,
- deteksi mismatch pada status code, envelope, dan field wajib,
- stop cutover jika mismatch melewati threshold yang disetujui.

## Exit Criteria

Migrasi eksekusi dinyatakan selesai jika:

- semua domain publik sudah cutover,
- seluruh gate domain berstatus pass,
- rollback rehearsal terakhir lulus,
- runtime lama tidak lagi dibutuhkan pada jalur request publik,
- gate decommission runtime legacy lulus (`make runtime-decommission`).

## Related Documents

- [Migration Strategy](/Users/macbookpro/Development/recova-backend-v2/docs/roadmap/migration-strategy.md)
- [Cutover Checklist](/Users/macbookpro/Development/recova-backend-v2/docs/roadmap/cutover-checklist.md)
- [Rollback Checklist](/Users/macbookpro/Development/recova-backend-v2/docs/roadmap/rollback-checklist.md)
- [Deployment Workflow](/Users/macbookpro/Development/recova-backend-v2/docs/operations/deployment.md)

## Source Reference

- [Migration Strategy](/Users/macbookpro/Development/recova-backend-v2/docs/roadmap/migration-strategy.md)
- [/Users/macbookpro/Development/bisakerja-api/docs/operations/deployment.md](/Users/macbookpro/Development/bisakerja-api/docs/operations/deployment.md)
- [Go Fiber Documentation](https://docs.gofiber.io/)
- [PostgreSQL Current Documentation](https://www.postgresql.org/docs/current/index.html)
