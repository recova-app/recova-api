---
title: Recova Backend Deployment Workflow
description: Runbook deployment layanan Recova Backend mencakup topology rollout, urutan migrasi database, injeksi konfigurasi, dan smoke verification.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/operations/deployment.md
last_reviewed: 2026-05-08
---

# Recova Backend Deployment Workflow

Dokumen ini mendefinisikan alur deploy layanan secara aman dan terverifikasi, dari persiapan artefak sampai trafik stabil.

## Deployment Goals

- artefak yang dirilis harus reproducible,
- perubahan skema database harus terkontrol,
- downtime harus minimal,
- rollback aplikasi harus cepat,
- dampak insiden harus terlokalisasi.

## Topology Options

| Topology   | Kapan dipakai                                           | Karakteristik                                                                      |
| ---------- | ------------------------------------------------------- | ---------------------------------------------------------------------------------- |
| Rolling    | default environment dengan resource terbatas            | kapasitas dipindah secara gradual ke versi baru, validasi dilakukan selama rollout |
| Blue/Green | rilis berisiko tinggi atau perubahan besar pada runtime | environment baru disiapkan penuh, trafik dipindah setelah smoke checks lulus       |

Aturan pemilihan:

- gunakan rolling untuk perubahan kecil/menengah,
- gunakan blue/green untuk perubahan berisiko tinggi atau saat rollback instan dibutuhkan.

## Deployment Preconditions

Sebelum deploy, wajib tersedia:

- image/container artifact yang immutable (tag commit atau digest),
- konfigurasi environment tervalidasi,
- migration scripts yang sudah lolos verifikasi,
- backup database terbaru untuk rilis dengan perubahan skema,
- release gates lulus sesuai [Release Gates](/Users/macbookpro/Development/recova-backend-v2/docs/operations/release-gates.md).

## Environment Injection Rules

- seluruh konfigurasi runtime disuplai dari secret manager atau environment runtime,
- secret tidak boleh di-hardcode di image atau repository,
- variabel wajib harus fail-fast saat startup,
- perbedaan konfigurasi antar-environment harus eksplisit dan terdokumentasi di [Environment](/Users/macbookpro/Development/recova-backend-v2/docs/environment.md).

## Standard Deployment Sequence

```text
1) Build and publish immutable artifact
2) Validate environment configuration
3) Run migration precheck
4) Apply forward migrations
5) Roll out application version
6) Execute post-deploy checks
7) Promote deployment status to healthy
```

## Database Migration Order

Aturan urutan migrasi:

1. apply migration sebelum trafik penuh ke versi aplikasi baru,
2. gunakan pola expand-then-contract untuk perubahan yang memengaruhi kontrak aktif,
3. perubahan destruktif hanya setelah masa kompatibilitas selesai,
4. migration failure harus menghentikan rollout.

Jika migration gagal:

- hentikan perpindahan trafik,
- jalankan prosedur rollback sesuai [Rollback Runbook](/Users/macbookpro/Development/recova-backend-v2/docs/operations/rollback.md),
- dokumentasikan penyebab sebelum retry deploy.

## Smoke Verification

Setelah deploy, minimal lakukan:

- `GET /health/live` sukses,
- `GET /health/ready` sukses,
- endpoint kritis auth dan data utama merespons sesuai kontrak,
- error rate dan latency tidak melewati baseline alarm.

Daftar lengkap ada di [Post-Deploy Checks](/Users/macbookpro/Development/recova-backend-v2/docs/operations/post-deploy-checks.md).

## Deployment Evidence

Setiap deploy harus punya bukti:

- artifact identifier (tag/digest),
- migration version yang diterapkan,
- timestamp deploy,
- hasil smoke checks,
- keputusan promote/rollback.

## Security and Access Rules

- akses deploy hanya untuk role terotorisasi,
- kredensial deploy dipisah per environment,
- aktivitas deploy harus audit-able,
- akses shell ke runtime produksi dibatasi sesuai prinsip least privilege.

## Related Documents

- [Release Gates](/Users/macbookpro/Development/recova-backend-v2/docs/operations/release-gates.md)
- [CI/CD Operations](/Users/macbookpro/Development/recova-backend-v2/docs/operations/ci-cd.md)
- [Database Migrations](/Users/macbookpro/Development/recova-backend-v2/docs/operations/database-migrations.md)
- [Rollback Runbook](/Users/macbookpro/Development/recova-backend-v2/docs/operations/rollback.md)
- [Post-Deploy Checks](/Users/macbookpro/Development/recova-backend-v2/docs/operations/post-deploy-checks.md)

## Source Reference

- [/Users/macbookpro/Development/bisakerja-api/docs/operations/deployment.md](/Users/macbookpro/Development/bisakerja-api/docs/operations/deployment.md)
- [Fiber App API](https://docs.gofiber.io/api/app/)
- [Docker Multi-stage Builds](https://docs.docker.com/build/building/multi-stage/)
- [golang-migrate](https://github.com/golang-migrate/migrate)
- [PostgreSQL Backup and Restore](https://www.postgresql.org/docs/current/backup.html)
