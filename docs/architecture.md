---
title: Recova Backend Architecture
description: Batas arsitektur layanan, alur request, dan daftar dokumen arsitektur teknis Recova Backend.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/architecture.md
last_reviewed: 2026-05-08
---

# Recova Backend Architecture

Halaman ini adalah titik masuk untuk dokumentasi arsitektur backend Recova. Fokus utama adalah boundary antarlayer, kontrak antarmodul, penanganan error, observability baseline, dan aturan dependency direction.

## Architecture Scope

- request lifecycle dari edge ke data layer,
- pemisahan tanggung jawab transport, domain, dan persistence,
- kontrak error dan response,
- batas integrasi eksternal.

## Related Sections

- [Documentation Overview](/Users/macbookpro/Development/recova-backend-v2/docs/overview.md)
- [Project Standards](/Users/macbookpro/Development/recova-backend-v2/docs/standards/index.md)
- [Architecture Decisions](/Users/macbookpro/Development/recova-backend-v2/docs/decisions/index.md)
- [Operations](/Users/macbookpro/Development/recova-backend-v2/docs/operations/index.md)
