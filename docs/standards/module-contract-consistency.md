---
title: Recova Backend Module Contract Consistency Standard
description: Standar konsistensi kontrak runtime lintas module untuk envelope response, auth guard, error mapping, observability, dan sinkronisasi OpenAPI.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/standards/module-contract-consistency.md
last_reviewed: 2026-05-09
---

# Recova Backend Module Contract Consistency Standard

Dokumen ini menetapkan baseline kontrak lintas module agar perilaku API tetap seragam kecuali ada kebutuhan domain yang jelas dan terdokumentasi.

## Runtime Contract Baseline

Wajib konsisten untuk endpoint publik:

- envelope sukses/error mengikuti [API Response Standard](/Users/macbookpro/Development/recova-backend-v2/docs/api-response-standard.md),
- mapping error mengikuti [Error Taxonomy](/Users/macbookpro/Development/recova-backend-v2/docs/standards/error-taxonomy.md),
- route domain berjalan di prefix `/api/v1`,
- endpoint protected memakai auth guard,
- request id dipropagasi ke error envelope,
- route runtime sinkron dengan OpenAPI generated.

## Ownership and Access Baseline

Endpoint user-owned wajib memakai identitas dari auth context:

- journals,
- routine dan relapse history,
- achievements progress,
- AI persona preference dan chat history,
- community write path.

Repository dan service wajib menolak akses tidak valid melalui error contract standar (`UNAUTHENTICATED`, `FORBIDDEN`, `NOT_FOUND`, `CONFLICT` sesuai konteks).

## Repository Error Handling Baseline

Repository layer wajib:

- pakai `WithContext(ctx)` di query,
- memetakan `gorm.ErrRecordNotFound` di service ke `NOT_FOUND` saat relevan,
- mempertahankan boundary transaksi hanya di jalur tulis yang perlu atomik,
- menghindari pemaparan raw error ke response.

## Observability and Redaction Baseline

- log request wajib menyertakan metadata inti (`requestId`, route, status, duration),
- payload sensitif (token, raw journal, raw prompt AI, secret) tidak boleh masuk log mentah,
- event audit penting auth/write path dicatat lewat recorder observability.

## OpenAPI and Drift Baseline

Gate wajib:

- `make openapi-check`
- `go test -count=1 ./test/contract -run '^TestContract_.*OpenAPI|^TestContract_ProtectedRoutes_Unauthenticated_ValidAgainstOpenAPI$'`

Aturan:

- drift runtime vs OpenAPI memblok merge.
- perubahan kontrak harus memperbarui source spec dan artefak generated pada commit yang sama.

## Consistency Drift Register

| Area                      | Drift                                                                | Status            | Owner                | Next Action                                                                |
| ------------------------- | -------------------------------------------------------------------- | ----------------- | -------------------- | -------------------------------------------------------------------------- |
| Pagination metadata       | List endpoint belum menaruh metadata pagination di `meta.pagination` | `needs-follow-up` | `api-contract-owner` | Tambah pola pagination envelope standar untuk endpoint list yang berlimit. |
| Repository timeout policy | Belum ada standar timeout eksplisit per query lintas semua module    | `needs-follow-up` | `platform-owner`     | Definisikan timeout policy repository + guard test konsistensi.            |

## Enforcement

Gate konsistensi lintas module:

- `make module-consistency-check`
- `make openapi-check`
- `go test ./test/contract`

## Related Documents

- [API Response Standard](/Users/macbookpro/Development/recova-backend-v2/docs/api-response-standard.md)
- [Error Taxonomy](/Users/macbookpro/Development/recova-backend-v2/docs/standards/error-taxonomy.md)
- [OpenAPI Standard](/Users/macbookpro/Development/recova-backend-v2/docs/standards/openapi.md)
- [Observability](/Users/macbookpro/Development/recova-backend-v2/docs/operations/observability.md)
- [Testing Strategy](/Users/macbookpro/Development/recova-backend-v2/docs/operations/testing.md)

## Source Reference

- [OpenAPI Specification](https://spec.openapis.org/oas/)
- [Fiber Error Handling](https://docs.gofiber.io/guide/error-handling)
- [Fiber RequestID Middleware](https://docs.gofiber.io/middleware/requestid)
- [GORM Error Handling](https://gorm.io/docs/error_handling.html)
- [GORM Transactions](https://gorm.io/docs/transactions.html)
- [OWASP Logging Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Logging_Cheat_Sheet.html)
