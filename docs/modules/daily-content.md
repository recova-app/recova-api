---
title: Daily Content Module
description: Kontrak modul daily content untuk motivasi, tantangan harian, dan tantangan fisik harian dengan aturan lifecycle konten, boundary tanggal, dan konsistensi konsumsi.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/modules/daily-content.md
last_reviewed: 2026-05-15
---

# Daily Content Module

## Responsibility

- menyediakan motivasi, tantangan harian, dan tantangan fisik harian,
- memastikan konten harian konsisten per tanggal,
- mengelola lifecycle konten aktif/nonaktif,
- menjaga keterpisahan data konten vs data interaksi user.

## API Contract

Route prefix:

```text
/api/v1/content
```

| Method | Path                    | Auth class | Purpose                   |
| ------ | ----------------------- | ---------- | ------------------------- |
| `GET`  | `/api/v1/content/daily` | Bearer     | ambil konten harian aktif |

## Database Model

Entitas utama:

- `daily_motivations`,
- `daily_challenges`,
- `daily_physical_challenges`,
- mapping tanggal konten harian.

Constraint minimum:

- satu set konten aktif per tanggal layanan,
- konten nonaktif tidak boleh muncul di output endpoint.

## Authentication and Authorization

- endpoint konsumsi saat ini memakai bearer auth,
- update konten harian hanya pada jalur internal terotorisasi,
- user tidak dapat menulis source konten harian.

## Service and Business Rules

- pemilihan konten mengikuti boundary tanggal yang terdokumentasi,
- boundary tanggal layanan menggunakan hari UTC agar deterministik lintas environment runtime,
- jika stok konten kurang, fallback harus deterministik,
- status konten harus audit-able saat berubah.

## Validation Rules

- field `motivation` wajib valid dan non-empty,
- field `challenge.title`, `challenge.description`, `physical_challenge.title`, dan `physical_challenge.description` wajib valid dan non-empty,
- `status` wajib nilai yang diizinkan,
- format tanggal konten harus konsisten,
- payload invalid ditolak.

## Error Contract

| Condition                             | HTTP  | Error code         |
| ------------------------------------- | ----- | ------------------ |
| auth invalid/missing                  | `401` | `UNAUTHENTICATED`  |
| konten tidak tersedia                 | `404` | `NOT_FOUND`        |
| payload invalid (internal write flow) | `422` | `VALIDATION_ERROR` |
| kegagalan internal                    | `500` | `INTERNAL_ERROR`   |

## Observability Contract

Log field minimum:

- `request_id`,
- `content_date`,
- `daily_content_action`,
- `status_code`.

Metrik minimum:

- daily content fetch rate,
- fallback usage count,
- latency endpoint daily content,
- error rate.

## Testing Requirements

- unit test selector konten per tanggal,
- unit test validator challenge/motivation,
- integration test query konten aktif,
- contract test bentuk payload `daily`,
- regression test fallback saat stok menipis.

## Open Gaps

- timezone final boundary tanggal lintas wilayah,
- format final konten challenge multimedia,
- kebijakan rotasi saat konten habis.

## Related Documents

- [Education Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/education.md)
- [Data Flow Overview](/Users/macbookpro/Development/recova-backend-v2/docs/overview/data-flow.md)

## Source Reference

- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
