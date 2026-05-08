---
title: Education Module
description: Kontrak modul education untuk distribusi konten edukasi aktif dengan kontrol lifecycle konten dan konsistensi payload konsumsi klien.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/modules/education.md
last_reviewed: 2026-05-08
---

# Education Module

## Responsibility

- menyediakan daftar konten edukasi aktif,
- menjaga lifecycle aktif/nonaktif konten,
- menjaga kualitas metadata konten,
- memisahkan jalur konsumsi user dan jalur pengelolaan konten.

## API Contract

Route prefix:

```text
/api/v1/education
```

| Method | Path                | Auth class | Purpose                           |
| ------ | ------------------- | ---------- | --------------------------------- |
| `GET`  | `/api/v1/education` | Bearer     | ambil daftar konten edukasi aktif |

## Database Model

Entitas utama:

- `education_contents`,
- metadata status publikasi dan timestamp.

Constraint minimum:

- hanya konten status `active` yang tampil ke user,
- konten nonaktif tetap terlacak untuk audit.

## Authentication and Authorization

- endpoint konsumsi saat ini menggunakan bearer auth,
- perubahan konten hanya melalui jalur internal terotorisasi,
- data editorial tidak boleh diekspos ke endpoint konsumsi.

## Service and Business Rules

- sorting konten harus deterministik,
- konten tidak aktif tidak boleh muncul di payload,
- fallback data harus aman jika stok konten terbatas.

## Validation Rules

- `title` dan ringkasan konten wajib valid,
- `status` hanya nilai yang diizinkan,
- atribut bahasa/label konten harus sesuai format,
- payload invalid ditolak sebelum persist.

## Error Contract

| Condition                             | HTTP  | Error code         |
| ------------------------------------- | ----- | ------------------ |
| auth invalid/missing                  | `401` | `UNAUTHENTICATED`  |
| payload invalid (internal write flow) | `422` | `VALIDATION_ERROR` |
| konten tidak ditemukan                | `404` | `NOT_FOUND`        |
| kegagalan internal                    | `500` | `INTERNAL_ERROR`   |

## Observability Contract

Log field minimum:

- `request_id`,
- `content_id`,
- `education_action`,
- `status_code`.

Metrik minimum:

- content fetch rate,
- response latency,
- error rate endpoint education.

## Testing Requirements

- unit test filter status konten,
- unit test validator metadata,
- integration test query konten aktif,
- contract test response list education,
- regression test fallback behavior saat data minim.

## Open Gaps

- keputusan final apakah route education bisa public,
- strategi pagination awal,
- kebutuhan final multi-bahasa dan fallback.

## Related Documents

- [Daily Content Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/daily-content.md)
- [Scraper Flow Integration](/Users/macbookpro/Development/recova-backend-v2/docs/integrations/scraper-flow.md)
- [Data Flow Overview](/Users/macbookpro/Development/recova-backend-v2/docs/overview/data-flow.md)

## Source Reference

- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
