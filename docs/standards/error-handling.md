---
title: Recova Backend Error Handling Standard
description: Standar penanganan error untuk layanan Go Recova Backend meliputi wrapping, klasifikasi domain, mapping ke HTTP response, dan redaksi informasi sensitif.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/standards/error-handling.md
last_reviewed: 2026-05-08
---

# Recova Backend Error Handling Standard

Dokumen ini mendefinisikan cara membuat, membungkus, mengklasifikasi, dan memetakan error secara konsisten.

## Error Design Principles

- error harus actionable untuk operator,
- response ke klien harus aman dan tidak membocorkan detail internal,
- root cause harus tetap bisa ditelusuri melalui log/traces.

## Wrapping and Inspection Rules

- gunakan wrapping berjenjang saat menambah konteks,
- gunakan `errors.Is` untuk sentinel error,
- gunakan `errors.As` untuk typed error,
- hindari string-compare untuk pengambilan keputusan.

## Error Classification

Gunakan kategori minimal berikut:

- `VALIDATION_ERROR`
- `UNAUTHENTICATED`
- `FORBIDDEN`
- `NOT_FOUND`
- `CONFLICT`
- `RATE_LIMITED`
- `DOWNSTREAM_ERROR`
- `INTERNAL_ERROR`

Kategori harus dipetakan konsisten ke status HTTP dan envelope API.

## Layer Responsibility

- repository mengembalikan error persistence terbungkus,
- service memetakan error teknis ke domain error bila perlu,
- handler tidak membuat keputusan klasifikasi domain dari string mentah,
- HTTP error mapping dilakukan di error handler terpusat.

## Client Response Rules

- gunakan format respons error standar,
- berikan `code` stabil untuk klien,
- detail internal (query, stack trace, secret) tidak boleh dikirim ke klien,
- `request_id` harus ikut di error response.

## Logging Rules

- log error menyertakan `request_id`, kategori, dan jalur endpoint,
- data sensitif dalam payload harus di-redact,
- stack trace dicatat hanya pada level yang sesuai kebijakan runtime.

## Retry and Idempotency Rules

- error sementara downstream diberi klasifikasi yang mendukung retry,
- operasi tidak idempotent harus mencegah duplicate side effect,
- conflict idempotent diperlakukan konsisten sesuai kontrak endpoint.

## Review Checklist

- pembungkusan error tetap mempertahankan root cause,
- mapping status HTTP dan error code sesuai taxonomy,
- tidak ada data sensitif pada body error,
- test mencakup kasus error utama.

## Related Documents

- [API Response Standard](/Users/macbookpro/Development/recova-backend-v2/docs/api-response-standard.md)
- [Error Taxonomy](/Users/macbookpro/Development/recova-backend-v2/docs/standards/error-taxonomy.md)
- [Go Coding Standards](/Users/macbookpro/Development/recova-backend-v2/docs/standards/go-coding-standards.md)

## Source Reference

- [Working with Errors in Go 1.13](https://go.dev/blog/go1.13-errors)
- [Effective Go](https://go.dev/doc/effective_go)
- [Fiber Guide: Error Handling](https://docs.gofiber.io/guide/error-handling/)
