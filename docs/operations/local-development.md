---
title: Recova Backend Local Development Runtime
description: Kontrak runtime pengembangan lokal untuk backend meliputi mode native dan container compose, kebutuhan dependency, serta aturan validasi lingkungan.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/operations/local-development.md
last_reviewed: 2026-05-08
---

# Recova Backend Local Development Runtime

Dokumen ini mendefinisikan setup runtime lokal agar pengembangan dan debugging konsisten.

## Local Runtime Profiles

### Profile A - Native Runtime

Cocok untuk iterasi cepat.

Komponen minimum:

- service backend lokal,
- PostgreSQL lokal atau instance dev terisolasi,
- env file lokal khusus development.

### Profile B - Container Compose Runtime

Cocok untuk parity environment.

Komponen minimum:

- service backend via container,
- PostgreSQL via service terpisah atau endpoint eksternal dev,
- volume terpisah untuk data non-persisten development bila diperlukan.

## Environment Rules

- gunakan `APP_ENV=local`,
- secret lokal tetap dipisah dari secret staging/production,
- env required harus valid saat startup,
- tidak boleh ada fallback tersembunyi untuk env required.

## Database Rules for Local

- gunakan database terisolasi untuk development,
- jalankan migration sebelum menjalankan flow fitur,
- seed data hanya data sintetis,
- jangan gunakan database production/staging untuk local verification.

## Local Verification Baseline

Setiap perubahan utama lokal harus memverifikasi:

- startup sukses,
- `GET /health/live` dan `GET /health/ready` sukses,
- endpoint yang diubah merespons sesuai kontrak,
- tidak ada log error kritikal saat skenario uji dasar.

## Common Troubleshooting

| Symptom          | First check                                          |
| ---------------- | ---------------------------------------------------- |
| startup gagal    | validasi env required dan format nilainya            |
| readiness gagal  | koneksi PostgreSQL, migration status, timeout config |
| auth route gagal | JWT/config cookie dan CORS origin policy             |
| AI route lambat  | timeout konfigurasi dan status provider              |

## Related Documents

- [Environment Configuration](/Users/macbookpro/Development/recova-backend-v2/docs/environment.md)
- [Database Migrations](/Users/macbookpro/Development/recova-backend-v2/docs/operations/database-migrations.md)
- [Containerization Strategy](/Users/macbookpro/Development/recova-backend-v2/docs/operations/containerization.md)

## Source Reference

- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
- [PostgreSQL Current Documentation](https://www.postgresql.org/docs/current/)
- [Docker Multi-stage Builds](https://docs.docker.com/build/building/multi-stage/)
