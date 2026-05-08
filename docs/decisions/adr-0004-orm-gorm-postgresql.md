---
title: ADR 0004 - ORM GORM PostgreSQL
description: Keputusan penggunaan GORM v2 di atas PostgreSQL untuk persistence layer backend Recova.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/decisions/adr-0004-orm-gorm-postgresql.md
last_reviewed: 2026-05-08
---

# ADR 0004 - ORM GORM PostgreSQL

## Status

Proposed

## Context

Backend membutuhkan data layer relasional dengan:

- transaksi kuat untuk auth, profile, routine, journal, dan community flows,
- query yang cukup ekspresif untuk modul domain,
- ergonomi implementasi yang tetap menjaga kontrol SQL bila dibutuhkan.

## Decision

Gunakan **PostgreSQL** sebagai database utama dan **GORM v2** sebagai ORM pada application layer.

## Decision Drivers

- PostgreSQL unggul untuk konsistensi transaksi dan relasi data,
- GORM menyediakan transaksi, nested transaction, savepoint, dan fitur query yang cukup kaya,
- kombinasi ini menekan boilerplate tanpa mengorbankan kontrol query.

## Alternatives Considered

### A1 - Raw SQL only

- plus: kontrol SQL penuh.
- minus: biaya implementasi repository dan mapping tinggi untuk modul banyak.
- hasil: tidak dipilih sebagai baseline utama.

### A2 - SQL builder tanpa ORM

- plus: lebih eksplisit daripada ORM.
- minus: tetap butuh boilerplate domain mapping yang besar.
- hasil: tidak dipilih.

### A3 - GORM v2 + PostgreSQL

- plus: produktif dan tetap mendukung query SQL eksplisit saat perlu.
- minus: risiko query implicit dan performa jika penggunaan ORM tidak disiplin.
- hasil: dipilih.

## Consequences

Konsekuensi positif:

- pengembangan repository lebih cepat,
- transaksi domain kritis lebih mudah dikelola,
- integrasi model-domain ke persistence lebih konsisten.

Konsekuensi negatif:

- perlu pedoman ketat untuk preload/join agar tidak memicu query berlebih,
- review query wajib untuk endpoint ber-volume tinggi.

## Guardrails

- repository wajib jadi satu-satunya layer yang berinteraksi dengan ORM,
- query sensitif performa wajib diinspeksi dengan SQL log/analysis,
- schema migration tetap dikelola SQL migration files, bukan auto-migrate runtime produksi.

## Related Documents

- [Tech Stack](/Users/macbookpro/Development/recova-backend-v2/docs/tech-stack.md)
- [Database](/Users/macbookpro/Development/recova-backend-v2/docs/database.md)
- [Project Structure](/Users/macbookpro/Development/recova-backend-v2/docs/project-structure.md)

## Source Reference

- [GORM Documentation](https://gorm.io/docs/)
- [PostgreSQL Current Documentation](https://www.postgresql.org/docs/current/index.html)
