---
title: Recova Backend Module Documentation Template
description: Template baku dokumentasi modul untuk memastikan setiap domain memiliki kontrak API, model data, auth, error, observability, dan verifikasi test yang konsisten.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/standards/module-documentation-template.md
last_reviewed: 2026-05-08
---

# Recova Backend Module Documentation Template

Dokumen ini adalah format baku untuk menulis `docs/modules/*.md`.

## Required Metadata

Setiap dokumen modul wajib memiliki metadata standar:

- `title`
- `description`
- `owner`
- `reviewers`
- `doc_status`
- `source_repo`
- `source_path`
- `last_reviewed`

## Required Section Order

Gunakan urutan section ini agar seragam lintas modul:

1. `## Responsibility`
2. `## API Contract`
3. `## Database Model`
4. `## Authentication and Authorization`
5. `## Service and Business Rules`
6. `## Validation Rules`
7. `## Error Contract`
8. `## Observability Contract`
9. `## Testing Requirements`
10. `## Open Gaps`
11. `## Related Documents`
12. `## Source Reference`

## Minimal Content Rules Per Section

- `Responsibility`: cakupan dan non-cakupan modul.
- `API Contract`: route prefix, endpoint summary, request/response ringkas.
- `Database Model`: entitas utama, constraint, relasi penting.
- `Authentication and Authorization`: auth class + ownership rules.
- `Service and Business Rules`: idempotency, transaction boundary, dan aturan domain.
- `Validation Rules`: field-level rule utama, limit, format.
- `Error Contract`: mapping error code dan status HTTP utama.
- `Observability Contract`: log fields, metrics, trace attributes.
- `Testing Requirements`: unit, integration, contract, dan edge-case minimum.

## Reusable Skeleton

```md
## Responsibility

## API Contract

## Database Model

## Authentication and Authorization

## Service and Business Rules

## Validation Rules

## Error Contract

## Observability Contract

## Testing Requirements

## Open Gaps

## Related Documents

## Source Reference
```

## Quality Checklist

- setiap section required terisi,
- istilah endpoint/path konsisten dengan API reference,
- tidak ada konteks task internal,
- referensi sumber relevan dan terbaru,
- data sensitif tidak ditulis mentah.

## Related Documents

- [Modules Index](/Users/macbookpro/Development/recova-backend-v2/docs/modules/index.md)
- [API Reference](/Users/macbookpro/Development/recova-backend-v2/docs/api-reference.md)
- [OpenAPI Standard](/Users/macbookpro/Development/recova-backend-v2/docs/standards/openapi.md)
- [Testing Conventions](/Users/macbookpro/Development/recova-backend-v2/docs/standards/testing-conventions.md)

## Source Reference

- [/Users/macbookpro/Development/bisakerja-api/docs/modules/users.md](/Users/macbookpro/Development/bisakerja-api/docs/modules/users.md)
- [/Users/macbookpro/Development/bisakerja-api/docs/modules/jobs.md](/Users/macbookpro/Development/bisakerja-api/docs/modules/jobs.md)
- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
