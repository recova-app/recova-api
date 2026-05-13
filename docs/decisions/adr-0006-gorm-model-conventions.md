---
title: ADR 0006 - GORM Model Conventions
description: Keputusan konvensi model GORM dan pemisahan tanggung jawab antara model struct, migration SQL, serta boundary transaksi.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/decisions/adr-0006-gorm-model-conventions.md
last_reviewed: 2026-05-08
---

# ADR 0006 - GORM Model Conventions

## Status

Proposed

## Context

Data layer membutuhkan konvensi model yang konsisten agar:

- skema mudah dipelihara,
- perilaku query stabil,
- perubahan schema aman untuk deployment inkremental,
- boundary transaksi lintas modul tetap terkendali.

## Decision

Gunakan konvensi GORM standar untuk naming (`snake_case`, table plural), timestamp (`CreatedAt`, `UpdatedAt`), dan soft delete terpilih (`DeletedAt`) dengan aturan berikut:

- model mendeklarasikan intent struktur,
- migration SQL menjadi sumber kebenaran final untuk schema production,
- transaksi dibuka di service layer dan repository menerima handle `tx`.

## Decision Drivers

- konvensi GORM mempercepat implementasi tanpa konfigurasi berulang,
- PostgreSQL membutuhkan definisi constraint/index yang eksplisit dan dapat diaudit,
- pemisahan model intent vs migration SQL menurunkan risiko drift schema production.

## Alternatives Considered

### A1 - Andalkan `AutoMigrate` untuk production

- plus: setup awal cepat.
- minus: kontrol skema, review perubahan, dan rollback lemah.
- hasil: ditolak.

### A2 - Raw SQL model mapping penuh tanpa ORM convention

- plus: kontrol penuh.
- minus: boilerplate tinggi dan biaya maintenance besar.
- hasil: tidak dipilih sebagai baseline.

### A3 - Konvensi GORM + migration SQL eksplisit

- plus: produktif, konsisten, dan tetap audit-friendly.
- minus: butuh disiplin menulis migration serta review lintas layer.
- hasil: dipilih.

## Consequences

Konsekuensi positif:

- standar model lintas modul seragam,
- migrasi schema production lebih aman,
- transaction boundary lebih jelas.

Konsekuensi negatif:

- tim harus disiplin menjaga sinkronisasi model dan migration,
- perlu quality gate tambahan untuk index/constraint dan soft delete.

## Guardrails

- dilarang memakai `AutoMigrate` sebagai mekanisme migrasi production,
- perubahan struktur tabel wajib lewat migration SQL berpasangan `up/down`,
- indeks unik tidak boleh diduplikasi manual bila sudah dibentuk constraint,
- nested transaction hanya untuk kasus savepoint yang terdokumentasi.

## Related Documents

- [GORM Modeling Standard](/Users/macbookpro/Development/recova-backend-v2/docs/standards/gorm-modeling.md)
- [Database](/Users/macbookpro/Development/recova-backend-v2/docs/database.md)
- [Database Migrations](/Users/macbookpro/Development/recova-backend-v2/docs/operations/database-migrations.md)

## Source Reference

- [GORM Conventions](https://gorm.io/docs/conventions.html)
- [GORM Declaring Models](https://gorm.io/docs/models.html)
- [GORM Transactions](https://gorm.io/docs/transactions.html)
- [PostgreSQL Constraints](https://www.postgresql.org/docs/current/ddl-constraints.html)
