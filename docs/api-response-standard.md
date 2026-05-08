---
title: Recova Backend API Response Standard
description: Standar envelope response sukses dan error, metadata pagination, serta aturan keamanan payload API publik.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/api-response-standard.md
last_reviewed: 2026-05-08
---

# Recova Backend API Response Standard

Dokumen ini mendefinisikan bentuk response API publik yang konsisten.

## Envelope Contract

### Success Envelope

```json
{
  "success": true,
  "message": "Request completed",
  "data": {},
  "meta": null
}
```

### Error Envelope

```json
{
  "success": false,
  "message": "Request failed",
  "data": null,
  "error": {
    "code": "VALIDATION_ERROR",
    "details": [],
    "requestId": "req_123"
  }
}
```

## Field Rules

| Field             | Rule                                                     |
| ----------------- | -------------------------------------------------------- |
| `success`         | wajib ada pada semua response JSON                       |
| `message`         | ringkas, aman, tidak mengandung detail internal sensitif |
| `data`            | objek/array pada sukses, `null` pada error               |
| `meta`            | metadata tambahan atau `null`                            |
| `error.code`      | machine-readable code stabil                             |
| `error.details`   | detail aman untuk klien, tanpa stack/internal payload    |
| `error.requestId` | wajib untuk korelasi log                                 |

## Pagination Metadata

Untuk endpoint list:

```json
{
  "meta": {
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 120,
      "totalPages": 6,
      "hasNextPage": true,
      "hasPrevPage": false
    }
  }
}
```

Aturan:

- `page >= 1`
- `limit >= 1`
- `total >= 0`
- nilai invalid dipetakan ke `VALIDATION_ERROR`

## Error Code Set

| Error code            | HTTP status |
| --------------------- | ----------- |
| `BAD_REQUEST`         | `400`       |
| `VALIDATION_ERROR`    | `422`       |
| `UNAUTHENTICATED`     | `401`       |
| `FORBIDDEN`           | `403`       |
| `NOT_FOUND`           | `404`       |
| `CONFLICT`            | `409`       |
| `RATE_LIMITED`        | `429`       |
| `DOWNSTREAM_ERROR`    | `502`       |
| `SERVICE_UNAVAILABLE` | `503`       |
| `INTERNAL_ERROR`      | `500`       |

## Security Rules

Response API tidak boleh mengandung:

- stack trace,
- raw error driver database,
- raw payload provider eksternal,
- token/kredensial,
- konten sensitif pengguna (misal isi jurnal/chat) di message error umum.

## Mapping Guidance (High-Level)

- validasi input gagal -> `VALIDATION_ERROR` (`422`)
- token tidak valid/tidak ada -> `UNAUTHENTICATED` (`401`)
- akses tidak diizinkan -> `FORBIDDEN` (`403`)
- data tidak ditemukan -> `NOT_FOUND` (`404`)
- unique/conflict domain -> `CONFLICT` (`409`)
- limiter trigger -> `RATE_LIMITED` (`429`)
- provider eksternal gagal -> `DOWNSTREAM_ERROR` (`502`) atau `SERVICE_UNAVAILABLE` (`503`)
- panic/unhandled -> `INTERNAL_ERROR` (`500`)

## Related Documents

- [Error Taxonomy](/Users/macbookpro/Development/recova-backend-v2/docs/standards/error-taxonomy.md)
- [API Reference](/Users/macbookpro/Development/recova-backend-v2/docs/api-reference.md)
- [Request Lifecycle](/Users/macbookpro/Development/recova-backend-v2/docs/overview/request-lifecycle.md)

## Source Reference

- [/Users/macbookpro/Development/bisakerja-api/docs/api-response-standard.md](/Users/macbookpro/Development/bisakerja-api/docs/api-response-standard.md)
- [Fiber Error Handling Guide](https://docs.gofiber.io/guide/error-handling)
