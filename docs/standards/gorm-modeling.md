---
title: GORM Modeling Standard
description: Standar pemodelan GORM untuk penamaan skema, timestamps, soft delete, relasi, indeks, dan boundary transaksi di backend Recova.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/standards/gorm-modeling.md
last_reviewed: 2026-05-08
---

# GORM Modeling Standard

Dokumen ini menetapkan aturan pemodelan GORM agar struktur data konsisten, aman, dan selaras dengan perilaku PostgreSQL.

## Scope

Standar ini mencakup:

- konvensi nama model dan kolom,
- aturan primary key, timestamps, dan soft delete,
- relasi dan foreign key,
- indeks dan unique constraints,
- boundary transaksi pada service dan repository.

## Model Naming Conventions

Aturan default:

- gunakan struct Go `PascalCase` untuk model,
- gunakan nama tabel `snake_case` plural,
- gunakan nama kolom `snake_case`.

Aturan override:

- override nama tabel hanya bila perlu kompatibilitas kontrak data lama,
- jangan gunakan nama tabel dinamis lewat `TableName()` karena hasil fungsi di-cache,
- untuk kasus tabel dinamis administratif, gunakan `Scopes` atau `Table(...)` secara eksplisit.

## Primary Key Rules

- setiap model wajib memiliki field identifier eksplisit,
- field identifier wajib diberi tag `primaryKey`,
- tipe identifier ditetapkan per entitas domain dan harus konsisten lintas relasi,
- perubahan tipe identifier lintas modul harus melalui ADR baru.

## Timestamps and Timezone

- gunakan `CreatedAt` dan `UpdatedAt` untuk audit waktu,
- simpan waktu dalam UTC,
- dilarang menyimpan waktu lokal tanpa offset,
- jika butuh kontrol granular, gunakan `autoCreateTime`/`autoUpdateTime` secara eksplisit.

## Soft Delete Policy

- soft delete hanya digunakan untuk entitas yang butuh pemulihan data,
- gunakan `gorm.DeletedAt` dan indeks pada kolom delete marker,
- entitas yang menyimpan data sangat sensitif dapat memakai hard delete terkontrol,
- keputusan soft delete per entitas harus terdokumentasi di dokumen modul terkait.

## Relationships and Foreign Keys

- foreign key wajib didefinisikan eksplisit,
- aturan `ON UPDATE` dan `ON DELETE` wajib ditentukan pada migration SQL,
- relasi antar modul domain tidak boleh mem-bypass repository milik modul owner,
- preload relasi hanya untuk kebutuhan response yang terukur.

## Index and Constraint Rules

- jadikan migration SQL sebagai sumber kebenaran final untuk index/constraint production,
- gunakan tag GORM `index`/`uniqueIndex` untuk ekspresi intent model,
- jangan membuat indeks manual yang menduplikasi unique constraint,
- untuk kebutuhan uniqueness dengan soft delete, gunakan partial unique index di PostgreSQL (`WHERE deleted_at IS NULL`) pada migration SQL.

## Transaction Boundary Rules

Aturan transaksi:

- transaksi dibuka di layer service/use case, bukan di handler,
- semua repository dalam satu use case harus memakai handle `tx` yang sama,
- I/O eksternal (HTTP, queue publish, provider AI) tidak boleh berada di dalam transaksi DB utama,
- nested transaction hanya boleh dipakai jika memang butuh savepoint rollback parsial.

## Migration Compatibility Rules

- perubahan schema wajib kompatibel untuk rollout inkremental,
- perubahan destruktif wajib pola expand-then-contract,
- setiap migration harus punya `up` dan `down` yang bisa diverifikasi,
- `AutoMigrate` tidak boleh menjadi mekanisme migrasi production.

## Quality Gate

Checklist sebelum model dinyatakan siap:

- nama tabel/kolom mengikuti konvensi,
- key dan foreign key eksplisit,
- indeks dan constraints sudah ada di SQL migration,
- boundary transaksi tidak melanggar aturan I/O,
- perubahan terdokumentasi di modul yang terdampak.

## Related Documents

- [Database](/Users/macbookpro/Development/recova-backend-v2/docs/database.md)
- [Database Migrations](/Users/macbookpro/Development/recova-backend-v2/docs/operations/database-migrations.md)
- [ADR 0004 ORM GORM PostgreSQL](/Users/macbookpro/Development/recova-backend-v2/docs/decisions/adr-0004-orm-gorm-postgresql.md)
- [ADR 0006 GORM Model Conventions](/Users/macbookpro/Development/recova-backend-v2/docs/decisions/adr-0006-gorm-model-conventions.md)

## Source Reference

- [GORM Conventions](https://gorm.io/docs/conventions.html)
- [GORM Declaring Models](https://gorm.io/docs/models.html)
- [GORM Transactions](https://gorm.io/docs/transactions.html)
- [GORM Database Indexes](https://gorm.io/docs/indexes.html)
- [PostgreSQL Constraints](https://www.postgresql.org/docs/current/ddl-constraints.html)
- [PostgreSQL Unique Indexes](https://www.postgresql.org/docs/current/indexes-unique.html)
