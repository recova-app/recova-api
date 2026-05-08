---
title: Recova Backend Data Flow Overview
description: Peta aliran data utama backend Recova dari request client, proses domain, penyimpanan data, integrasi AI, dan jalur ingestion konten.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/overview/data-flow.md
last_reviewed: 2026-05-08
---

# Recova Backend Data Flow Overview

Dokumen ini merangkum aliran data lintas modul backend dan boundary dependency eksternal.

## High-Level Flow

```text
Client App
  -> HTTP API (/api/v1)
      -> middleware (request id, auth, validation, rate limit)
      -> module handler (auth/users/routine/journals/community/education/content/ai)
      -> service layer
      -> repository layer
      -> PostgreSQL

AI-specific path:
Client App -> /api/v1/ai/* -> AI service abstraction -> AI provider -> safe response mapping -> Client App

Content ingestion path:
Curated source/seed process -> content repository -> education & daily content endpoints
```

## Core Data Domains

- identity and profile data,
- routine and streak data,
- journal data,
- community interaction data,
- education and daily content data,
- AI conversation metadata and derived insights.

## Ownership Boundaries

| Boundary                                 | Owner          |
| ---------------------------------------- | -------------- |
| request authentication and authorization | Recova Backend |
| core transactional data                  | Recova Backend |
| AI inference execution                   | AI Provider    |
| external content ingestion automation    | gap / TBD      |

## Data Safety Boundaries

- data sensitif (`L3`, `L4`) tidak boleh dicatat mentah pada log,
- context AI harus diperkecil ke field minimum,
- community payload perlu moderasi sebelum publikasi,
- konten edukasi/harian harus melalui lifecycle status sebelum tampil.

## Ingestion Direction for Content

Kondisi saat ini dari sumber yang tersedia:

- requirement scraper/ingestion otomatis untuk konten belum dinyatakan eksplisit,
- baseline aman: konten edukasi dan konten harian berasal dari seed atau kurasi internal,
- jika ingestion otomatis ditambahkan, flow wajib didokumentasikan sebagai integrasi terpisah dengan kontrol validasi, moderasi, dan observability.

## Failure Mapping Overview

- kegagalan DB -> `503 SERVICE_UNAVAILABLE`,
- kegagalan provider AI -> `502 DOWNSTREAM_ERROR` atau `503 SERVICE_UNAVAILABLE`,
- validasi input gagal -> `422 VALIDATION_ERROR`,
- konflik data idempotensi -> `409 CONFLICT` bila tidak dapat dipetakan ke no-op aman.

## Open Gaps

- belum ada sumber rinci untuk pipeline scraper konten otomatis,
- belum ada kontrak final scheduler ingestion,
- belum ada owner final untuk operasi ingestion lintas-environment.

## Related Documents

- [Architecture](/Users/macbookpro/Development/recova-backend-v2/docs/architecture.md)
- [AI Provider Integration](/Users/macbookpro/Development/recova-backend-v2/docs/integrations/ai-provider.md)
- [Scraper Flow Integration](/Users/macbookpro/Development/recova-backend-v2/docs/integrations/scraper-flow.md)
- [Scraper Requirement Gap Register](/Users/macbookpro/Development/recova-backend-v2/docs/roadmap/scraper-requirement-gap.md)

## Source Reference

- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
- [/Users/macbookpro/Development/bisakerja-api/docs/integrations/scraper-api.md](/Users/macbookpro/Development/bisakerja-api/docs/integrations/scraper-api.md)
