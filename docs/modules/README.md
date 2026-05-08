---
title: Recova Backend Module Documentation Guide
description: Panduan penggunaan dokumentasi modul domain Recova Backend beserta standar section wajib dan checklist readiness lintas modul.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/modules/README.md
last_reviewed: 2026-05-08
---

# Recova Backend Module Documentation Guide

Dokumen ini menjadi panduan cepat untuk menulis dan mereview dokumen modul domain.

## Source of Format

Gunakan template resmi:

- [Module Documentation Template](/Users/macbookpro/Development/recova-backend-v2/docs/standards/module-documentation-template.md)

## Required Sections

Setiap dokumen modul wajib memiliki section:

- `Responsibility`
- `API Contract`
- `Database Model`
- `Authentication and Authorization`
- `Service and Business Rules`
- `Validation Rules`
- `Error Contract`
- `Observability Contract`
- `Testing Requirements`
- `Open Gaps`

## Module List

- [Auth Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/auth.md)
- [Users Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/users.md)
- [Routine Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/routine.md)
- [Journals Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/journals.md)
- [Community Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/community.md)
- [Education Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/education.md)
- [Daily Content Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/daily-content.md)
- [AI Coach Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/ai-coach.md)
- [Health Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/health.md)

## Readiness Review Checklist

- metadata lengkap,
- seluruh section wajib tersedia,
- kontrak API sinkron dengan route inventory,
- auth/ownership rules jelas,
- error + observability + test requirements terdokumentasi,
- source reference relevan.

## Related Documents

- [Modules Index](/Users/macbookpro/Development/recova-backend-v2/docs/modules/index.md)
- [API Reference](/Users/macbookpro/Development/recova-backend-v2/docs/api-reference.md)
- [Route Inventory](/Users/macbookpro/Development/recova-backend-v2/docs/generated/routes.md)
- [Testing Strategy](/Users/macbookpro/Development/recova-backend-v2/docs/operations/testing.md)

## Source Reference

- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
- [/Users/macbookpro/Development/bisakerja-api/docs/modules/users.md](/Users/macbookpro/Development/bisakerja-api/docs/modules/users.md)
