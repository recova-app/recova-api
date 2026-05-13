---
title: Recova Backend Error Taxonomy
description: Klasifikasi error lintas-layer dan pemetaan error teknis ke kontrak error API yang aman untuk klien.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/standards/error-taxonomy.md
last_reviewed: 2026-05-08
---

# Recova Backend Error Taxonomy

Dokumen ini memetakan error teknis ke error API agar perilaku response konsisten lintas modul.

## Taxonomy Levels

| Level                | Scope                  | Tujuan                                        |
| -------------------- | ---------------------- | --------------------------------------------- |
| Domain error         | service/business rules | menyatakan pelanggaran aturan bisnis          |
| Validation error     | request validation     | menahan input invalid sebelum business logic  |
| Infrastructure error | DB/network/dependency  | menangkap kegagalan dependency eksternal      |
| Internal error       | panic/unhandled        | fallback aman saat error tidak terklasifikasi |

## API Error Contract

Semua error akhirnya harus dipetakan ke code API standar:

- `BAD_REQUEST`
- `VALIDATION_ERROR`
- `UNAUTHENTICATED`
- `FORBIDDEN`
- `NOT_FOUND`
- `CONFLICT`
- `RATE_LIMITED`
- `DOWNSTREAM_ERROR`
- `SERVICE_UNAVAILABLE`
- `INTERNAL_ERROR`

## Technical Mapping

| Source error                 | API code              | HTTP status | Notes                                      |
| ---------------------------- | --------------------- | ----------- | ------------------------------------------ |
| `validator.ValidationErrors` | `VALIDATION_ERROR`    | `422`       | field error aman untuk klien               |
| `fiber.Error` status `400`   | `BAD_REQUEST`         | `400`       | gunakan message aman                       |
| auth token invalid/missing   | `UNAUTHENTICATED`     | `401`       | jangan expose alasan sensitif              |
| authz ownership gagal        | `FORBIDDEN`           | `403`       | akses ditolak                              |
| `gorm.ErrRecordNotFound`     | `NOT_FOUND`           | `404`       | tanpa detail query internal                |
| unique constraint violation  | `CONFLICT`            | `409`       | map dari error SQLSTATE terkait uniqueness |
| rate limiter reject          | `RATE_LIMITED`        | `429`       | sertakan retry hint bila tersedia          |
| dependency timeout/error     | `DOWNSTREAM_ERROR`    | `502`       | provider eksternal gagal                   |
| `context.DeadlineExceeded`   | `DOWNSTREAM_ERROR`    | `502`       | timeout dependency atau upstream           |
| database unavailable         | `SERVICE_UNAVAILABLE` | `503`       | dependency inti tidak siap                 |
| panic/unhandled unknown      | `INTERNAL_ERROR`      | `500`       | log internal lengkap, response aman        |

## Logging and Redaction Rule

- log internal boleh menyimpan context teknis untuk investigasi.
- response ke klien hanya menampilkan data aman.
- data sensitif wajib direduksi pada log dan response.

## Error Handling Pipeline

```text
error source
  -> classify (domain/validation/infrastructure/internal)
  -> map to API error code
  -> attach requestId
  -> write structured log
  -> return safe envelope
```

## Related Documents

- [API Response Standard](/Users/macbookpro/Development/recova-backend-v2/docs/api-response-standard.md)
- [Request Lifecycle](/Users/macbookpro/Development/recova-backend-v2/docs/overview/request-lifecycle.md)
- [Architecture](/Users/macbookpro/Development/recova-backend-v2/docs/architecture.md)

## Source Reference

- [Fiber Error Handling Guide](https://docs.gofiber.io/guide/error-handling)
- [GORM Documentation](https://gorm.io/docs/)
- [/Users/macbookpro/Development/bisakerja-api/docs/api-response-standard.md](/Users/macbookpro/Development/bisakerja-api/docs/api-response-standard.md)
