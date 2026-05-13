---
title: Database Seeding
description: Standar seeding database untuk data referensi, fixture pengujian, dan data bootstrap development secara aman dan repeatable.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/operations/database-seeding.md
last_reviewed: 2026-05-13
---

# Database Seeding

Dokumen ini menetapkan standar seeding agar data awal environment dapat direproduksi tanpa merusak kontrak data aplikasi.

## Seeding Objectives

- menyediakan data bootstrap untuk development,
- menyediakan fixture deterministik untuk integration testing,
- memisahkan data referensi dari data skenario uji,
- mencegah data sensitif masuk seed.

## Data Classes

### Reference Data

Contoh:

- konten edukasi default,
- daily motivations,
- daily challenges.
- achievement catalog.

Aturan:

- boleh disimpan di repository,
- harus versioned,
- perubahan harus dicatat pada dokumen modul terkait.

### Development Fixture Data

Contoh:

- akun pengguna uji,
- profile dan check-in simulasi,
- journals dan interaksi komunitas untuk debugging lokal.

Aturan:

- hanya untuk lingkungan development/test,
- dilarang berisi kredensial nyata,
- gunakan value sintetis yang jelas bukan data produksi.

## Environment Rules

- development: seeding diperbolehkan,
- test/integration: seeding wajib deterministik,
- staging: hanya data referensi minimal,
- production: seeding manual dibatasi untuk reference bootstrap yang terdokumentasi.

## Idempotency and Re-run Policy

- seeding harus aman dijalankan ulang,
- gunakan natural key atau unique key untuk mencegah duplikasi,
- operasi update harus bersifat targeted, bukan truncate massal tanpa kontrol.

## Execution Order

Urutan minimum:

1. jalankan migration schema,
2. seed reference data,
3. seed fixture data non-produksi bila diperlukan,
4. verifikasi integrity constraints.

## Baseline Seed Runner

Seed reference default disimpan di:

- `migrations/seeds/000001_baseline_seed.sql`

Eksekusi lokal:

- `make seed`

Aturan runner:

- membutuhkan `DATABASE_URL`,
- menjalankan `psql` dengan `ON_ERROR_STOP=1`,
- script seed wajib idempotent (`ON CONFLICT DO NOTHING` untuk data referensi).

## Baseline User Fixture (Manual Auth)

Baseline seeder (`migrations/seeds/000001_baseline_seed.sql`) memakai akun manual berbasis email/password.

Kredensial fixture development:

- email: sesuai seed `users.email` (contoh: `andre.wijaya@gmail.com`),
- username: sesuai seed `users.username` (contoh: `andre_wijaya`),
- password default: `Recova123!` (disimpan sebagai bcrypt hash pada `users.password_hash`).

Aturan:

- hanya untuk development/test, bukan production/staging publik,
- password plaintext fixture hanya boleh muncul di dokumentasi internal, bukan di SQL,
- update fixture user wajib menjaga idempotency key `ON CONFLICT (email)` agar rerun aman.

## Security and Privacy Rules

- dilarang menaruh secret/token/API key di seed file,
- journal/chat simulasi tidak boleh berisi data personal nyata,
- hash password fixture bila skenario email/password diaktifkan,
- log seeding tidak boleh menampilkan payload sensitif penuh.

## Verification Checklist

- seluruh seed script dapat dijalankan ulang tanpa duplikasi,
- foreign key integrity valid,
- data referensi aktif tersedia untuk endpoint read,
- tidak ada data rahasia atau kredensial produksi.

## Automated Staging Check

Runner `scripts/staging-deploy.sh` menjalankan verifikasi seeding otomatis:

1. jalankan seed pass pertama,
2. simpan row count `education_contents`, `daily_motivations`, `daily_challenges`, `achievements`,
3. jalankan seed pass kedua,
4. pastikan row count tidak berubah (idempotent),
5. pastikan tidak ada duplicate content pada tabel reference harian dan tidak ada duplicate `achievements.code`.

## Baseline Catalog Minimum

Untuk baseline reference data non-production, jumlah minimum saat verifikasi seeding:

- `education_contents >= 8`,
- `daily_motivations >= 10`,
- `daily_challenges >= 10`,
- `achievements >= 10`.

## Related Documents

- [Database](/Users/macbookpro/Development/recova-backend-v2/docs/database.md)
- [Database Migrations](/Users/macbookpro/Development/recova-backend-v2/docs/operations/database-migrations.md)
- [Domain Entities Reference](/Users/macbookpro/Development/recova-backend-v2/docs/references/domain-entities.md)

## Source Reference

- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
- [PostgreSQL Current Documentation](https://www.postgresql.org/docs/current/index.html)
