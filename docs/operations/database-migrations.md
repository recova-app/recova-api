---
title: Database Migrations
description: Runbook migrasi skema PostgreSQL berbasis SQL up/down dengan golang-migrate, termasuk verifikasi, rollback, dan recovery dirty state.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/operations/database-migrations.md
last_reviewed: 2026-05-08
---

# Database Migrations

Dokumen ini mendefinisikan workflow migrasi database agar perubahan skema aman, dapat diaudit, dan konsisten antar lingkungan.

## Migration Strategy

- gunakan file migration SQL `up` dan `down`,
- semua perubahan schema harus melalui migration files yang versioned,
- satu migration harus fokus pada satu perubahan logis,
- hindari menggabungkan banyak perubahan berisiko tinggi dalam satu migration.

## Tooling Baseline

Migration runner:

- `golang-migrate` CLI/library,
- source migration berbasis filesystem repository,
- target database PostgreSQL.

## File Naming Convention

Gunakan format timestamp + deskripsi singkat:

```text
20260508103000_create_users_table.up.sql
20260508103000_create_users_table.down.sql
```

Aturan:

- deskripsi file harus menjelaskan intent schema,
- pasangan `up` dan `down` wajib tersedia,
- hindari nama ambigu seperti `update_table`.

## Authoring Rules

- migration harus idempotent terhadap kondisi awal yang diketahui,
- operasi berisiko lock besar harus dipisah dan diuji,
- constraint/index naming harus eksplisit,
- SQL harus ditulis agar bisa direview tanpa konteks tambahan.

## Execution Workflow

### Local Development

1. buat pasangan migration `up` dan `down`.
2. jalankan `up` ke database development.
3. verifikasi schema hasil perubahan.
4. jalankan `down` untuk memastikan rollback valid.
5. jalankan `up` ulang untuk memastikan jalur deploy normal.

### CI Validation

CI wajib memverifikasi:

- semua file migration valid,
- urutan migration dapat dijalankan penuh,
- rollback minimal satu langkah sukses,
- tidak ada drift schema tak terdokumentasi.

### Production Rollout

Urutan minimum:

1. backup readiness check,
2. pastikan tidak ada migration pending yang belum direview,
3. jalankan `up` secara serial,
4. monitor lock duration, error, dan latency aplikasi,
5. jika gagal, gunakan strategi rollback/forward-fix sesuai insiden.

## Rollback and Forward-Fix

Kebijakan:

- rollback dipakai jika risiko data minimal dan `down` tervalidasi,
- forward-fix dipakai untuk migration destruktif atau perubahan data sensitif,
- pilihan rollback vs forward-fix harus dicatat pada log insiden.

## Dirty State Recovery

Jika migration gagal dan status menjadi dirty:

1. hentikan rollout lanjutan,
2. identifikasi migration versi gagal,
3. perbaiki SQL penyebab gagal,
4. sinkronkan versi dengan perintah `force V` hanya setelah kondisi schema diverifikasi manual,
5. jalankan migration lanjutan setelah state bersih.

Catatan penting:

- `force V` hanya mengatur versi migrasi dan mengabaikan dirty flag,
- `force V` tidak mengeksekusi migration SQL, jadi verifikasi manual wajib.

## Change Safety Rules

- gunakan pola expand-then-contract untuk perubahan breaking,
- hindari drop column langsung pada rilis yang sama dengan perpindahan pembacaan data,
- untuk unique requirements, utamakan constraint/partial index yang eksplisit,
- jangan mengandalkan `AutoMigrate` sebagai pengganti migration SQL production.

## Verification Checklist

Sebelum perubahan schema dianggap siap:

- pasangan `up/down` tersedia,
- `up -> down -> up` lolos di lingkungan uji,
- constraint/index tervalidasi,
- runbook rollback/forward-fix jelas,
- dampak ke modul terdokumentasi.

## Related Documents

- [Database](/Users/macbookpro/Development/recova-backend-v2/docs/database.md)
- [Database Seeding](/Users/macbookpro/Development/recova-backend-v2/docs/operations/database-seeding.md)
- [GORM Modeling Standard](/Users/macbookpro/Development/recova-backend-v2/docs/standards/gorm-modeling.md)
- [ADR 0005 Database Migration Tool](/Users/macbookpro/Development/recova-backend-v2/docs/decisions/adr-0005-database-migration-tool.md)

## Source Reference

- [golang-migrate CLI Usage](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate)
- [PostgreSQL Constraints](https://www.postgresql.org/docs/current/ddl-constraints.html)
- [PostgreSQL Indexes](https://www.postgresql.org/docs/current/indexes.html)
