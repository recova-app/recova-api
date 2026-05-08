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
  "message": "Permintaan berhasil diproses",
  "data": {},
  "meta": null
}
```

### Error Envelope

```json
{
  "success": false,
  "message": "Permintaan gagal diproses",
  "data": null,
  "error": {
    "code": "VALIDATION_ERROR",
    "details": [],
    "requestId": "req_123"
  }
}
```

## Field Rules

| Field             | Rule                                                                        |
| ----------------- | --------------------------------------------------------------------------- |
| `success`         | wajib ada pada semua response JSON                                          |
| `message`         | ringkas, aman, dan **berbahasa Indonesia** untuk konsumsi user aplikasi     |
| `data`            | objek/array pada sukses, `null` pada error                                  |
| `meta`            | metadata tambahan atau `null`                                               |
| `error.code`      | machine-readable code stabil, tetap English uppercase                       |
| `error.details`   | detail aman untuk klien, gunakan bahasa Indonesia untuk teks human-readable |
| `error.requestId` | wajib untuk korelasi log                                                    |

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

## Language and Client Rules

- teks untuk konsumsi pengguna (`message`, `error.details` human-readable) harus memakai bahasa Indonesia,
- identifier teknis tetap English (`error.code`, nama field JSON, enum internal),
- jangan campur istilah user-facing English jika padanan Indonesia tersedia,
- pertahankan konsistensi agar client mobile tidak perlu fallback parsing per bahasa.

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
- [Flutter Internationalization](https://docs.flutter.dev/ui/internationalization)
