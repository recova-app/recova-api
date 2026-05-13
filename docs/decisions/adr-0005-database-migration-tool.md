---
title: ADR 0005 - Database Migration Tool
description: Keputusan penggunaan golang-migrate berbasis SQL up/down sebagai mekanisme migrasi skema database.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/decisions/adr-0005-database-migration-tool.md
last_reviewed: 2026-05-08
---

# ADR 0005 - Database Migration Tool

## Status

Proposed

## Context

Backend membutuhkan migration workflow yang:

- deterministik di local/CI/production,
- punya jejak perubahan skema yang audit-friendly,
- mendukung rollback direction yang eksplisit.

## Decision

Gunakan **golang-migrate v4** dengan migration files SQL `up` dan `down`.

## Decision Drivers

- format up/down SQL jelas dan mudah direview,
- tooling stabil, mature, dan umum dipakai di ekosistem Go,
- cocok untuk strategi migrasi yang memerlukan kontrol perubahan schema yang ketat.

## Alternatives Considered

### A1 - ORM auto-migrate runtime

- plus: setup awal cepat.
- minus: kontrol perubahan schema lemah untuk production.
- hasil: ditolak.

### A2 - gormigrate

- plus: integrasi dekat dengan GORM.
- minus: untuk skala perubahan schema yang lebih tinggi, SQL migration tool dedicated lebih aman dan eksplisit.
- hasil: tidak dipilih sebagai default.

### A3 - golang-migrate SQL up/down

- plus: deterministic, explicit, dan auditable.
- minus: butuh disiplin penulisan migration script.
- hasil: dipilih.

## Consequences

Konsekuensi positif:

- rollback/forward-fix direction terdokumentasi jelas,
- review schema change lebih mudah lintas tim,
- eksekusi migrasi lebih konsisten di pipeline deploy.

Konsekuensi negatif:

- tim wajib disiplin menulis script `down` yang valid,
- butuh runbook recovery untuk dirty migration state.

## Guardrails

- setiap migrasi wajib punya file `up` dan `down`,
- migrasi destruktif wajib punya rencana kompatibilitas data,
- verifikasi migrasi wajib dijalankan sebelum deploy,
- auto-migrate runtime produksi dilarang.

## Related Documents

- [Tech Stack](/Users/macbookpro/Development/recova-backend-v2/docs/tech-stack.md)
- [Database](/Users/macbookpro/Development/recova-backend-v2/docs/database.md)

## Source Reference

- [golang-migrate](https://github.com/golang-migrate/migrate)
- [PostgreSQL Current Documentation](https://www.postgresql.org/docs/current/index.html)
