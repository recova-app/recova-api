---
title: Recova Backend Documentation Overview
description: Ringkasan tujuan dokumentasi, batas layanan, dan peta navigasi dokumen Recova Backend.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/overview.md
last_reviewed: 2026-05-08
---

# Recova Backend Documentation Overview

Dokumentasi ini menjadi sumber acuan teknis untuk layanan backend Recova. Isi dokumen berfokus pada kontrak API, batas arsitektur, model data, operasi layanan, standar penulisan kode dan dokumen, serta keputusan teknis yang memengaruhi perilaku sistem.

## Service Scope

Layanan backend menangani:

- autentikasi pengguna,
- profil dan onboarding,
- check-in harian, streak, jurnal, dan statistik,
- komunitas,
- konten edukasi dan konten harian,
- AI Coach,
- penyimpanan data PostgreSQL.

Rangkuman capability diambil dari [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md).

## Documentation Map

- [Architecture](/Users/macbookpro/Development/recova-backend-v2/docs/architecture.md)
- [Authentication and Trust Boundaries](/Users/macbookpro/Development/recova-backend-v2/docs/overview/authentication-and-trust-boundaries.md)
- [Data Flow Overview](/Users/macbookpro/Development/recova-backend-v2/docs/overview/data-flow.md)
- [API](/Users/macbookpro/Development/recova-backend-v2/docs/api/index.md)
- [Database](/Users/macbookpro/Development/recova-backend-v2/docs/database.md)
- [Environment Configuration](/Users/macbookpro/Development/recova-backend-v2/docs/environment.md)
- [Modules](/Users/macbookpro/Development/recova-backend-v2/docs/modules/index.md)
- [Operations](/Users/macbookpro/Development/recova-backend-v2/docs/operations/index.md)
- [Standards](/Users/macbookpro/Development/recova-backend-v2/docs/standards/index.md)
- [Integrations](/Users/macbookpro/Development/recova-backend-v2/docs/integrations/index.md)
- [Decisions](/Users/macbookpro/Development/recova-backend-v2/docs/decisions/index.md)
- [Roadmap](/Users/macbookpro/Development/recova-backend-v2/docs/roadmap/index.md)

## Documentation Rules

- Setiap halaman wajib memiliki metadata frontmatter.
- Setiap keputusan atau kontrak harus punya referensi sumber.
- Jika sumber belum memadai, tulis gap secara eksplisit, jangan menebak.
- Dokumen ini berdiri sendiri dan tidak bergantung pada catatan tugas internal.
