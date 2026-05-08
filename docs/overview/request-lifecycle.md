---
title: Recova Backend Request Lifecycle
description: Alur request dari edge hingga persistence, termasuk jalur error, observability, dan aturan boundary layer.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/overview/request-lifecycle.md
last_reviewed: 2026-05-08
---

# Recova Backend Request Lifecycle

Dokumen ini menetapkan urutan eksekusi request di backend agar kontrak API stabil, mudah diuji, dan aman.

## Success Path

```text
HTTP Request
  -> Request ID middleware
  -> Security headers
  -> CORS gate
  -> Recover middleware
  -> Rate limiter
  -> Auth middleware (jika endpoint protected)
  -> Request validation
  -> Handler
  -> Service
  -> Repository
  -> PostgreSQL
  -> Service result mapping
  -> Response formatter
  -> HTTP Response
```

## Failure Path

```text
Validation/Auth/Domain/Downstream Error
  -> Centralized error mapper
  -> Error envelope
  -> Structured log (with request id)
  -> HTTP Response
```

Untuk panic:

```text
panic
  -> Recover middleware
  -> Internal error mapping
  -> Safe error envelope
  -> Structured log
```

## Middleware Contract

| Middleware       | Tujuan                            | Output minimum                               |
| ---------------- | --------------------------------- | -------------------------------------------- |
| Request ID       | korelasi request lintas log       | `x-request-id` tersedia                      |
| Security headers | baseline hardening HTTP           | header keamanan terpasang                    |
| CORS             | kontrol origin browser            | request lintas origin tervalidasi            |
| Recover          | mencegah crash process saat panic | panic terkonversi ke error aman              |
| Rate limiter     | proteksi abuse/traffic burst      | error `429` dengan envelope standar          |
| Auth             | verifikasi identitas              | principal terpasang untuk endpoint protected |
| Validation       | validasi params/query/body        | payload valid sebelum handler                |

## Handler/Service/Repository Contract

| Layer      | Input                  | Output                         | Catatan                     |
| ---------- | ---------------------- | ------------------------------ | --------------------------- |
| Handler    | request tervalidasi    | command/query ke service       | hanya mapping transport     |
| Service    | command/query domain   | result domain atau typed error | aturan bisnis dan otorisasi |
| Repository | query data terstruktur | model persistence              | tanpa response formatting   |

## Error Classification Gate

Error yang harus distandarkan:

- validation
- unauthenticated
- forbidden
- not found
- conflict
- rate limited
- downstream failure
- service unavailable
- internal

Standar detail ada di [Error Taxonomy](/Users/macbookpro/Development/recova-backend-v2/docs/standards/error-taxonomy.md) dan [API Response Standard](/Users/macbookpro/Development/recova-backend-v2/docs/api-response-standard.md).

## Observability Contract

- Semua log request/error wajib membawa request id.
- Jalur error harus menyimpan klasifikasi error yang konsisten.
- Health endpoint harus terpisah dari route bisnis.
- Error response ke klien harus aman, tidak mengandung stack trace/raw DB/raw provider payload.

## Related Documents

- [Architecture](/Users/macbookpro/Development/recova-backend-v2/docs/architecture.md)
- [Health and Readiness](/Users/macbookpro/Development/recova-backend-v2/docs/operations/health-readiness.md)
- [API Response Standard](/Users/macbookpro/Development/recova-backend-v2/docs/api-response-standard.md)

## Source Reference

- [Fiber v3 Documentation](https://docs.gofiber.io/)
- [Fiber Error Handling Guide](https://docs.gofiber.io/guide/error-handling)
- [OpenTelemetry Go Documentation](https://opentelemetry.io/docs/languages/go/)
