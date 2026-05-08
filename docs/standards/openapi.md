---
title: Recova Backend OpenAPI Standard
description: Standar pengelolaan kontrak OpenAPI untuk Recova Backend agar dokumentasi API terverifikasi, versioned, dan konsisten dengan perilaku runtime.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/standards/openapi.md
last_reviewed: 2026-05-08
---

# Recova Backend OpenAPI Standard

Dokumen ini mendefinisikan aturan sumber kontrak OpenAPI, tata kelola perubahan, dan mekanisme deteksi drift.

## Canonical Contract Location

Kontrak OpenAPI kanonik disimpan di repository:

- `docs/generated/openapi.yaml` atau `docs/generated/openapi.json`

Hanya satu artefak yang menjadi source of truth pada satu waktu. Format YAML atau JSON dipilih sesuai tooling aktif tim.

## OpenAPI Version Policy

- baseline kontrak mengikuti OpenAPI `3.1.x` untuk kompatibilitas ekosistem tooling saat ini,
- referensi spesifikasi terbaru tetap dipantau dari OpenAPI Initiative,
- upgrade minor/major versi spesifikasi harus melalui keputusan terdokumentasi.

## Source Strategy

Sumber kontrak dapat berasal dari salah satu pendekatan berikut:

1. spec-first: kontrak ditulis dulu lalu server mengikuti,
2. code-first: kontrak digenerate dari definisi route/handler,
3. hybrid: route metadata dari code, schema inti dari spec.

Aturan wajib:

- pendekatan yang dipakai harus konsisten lintas modul,
- perubahan endpoint harus memperbarui kontrak di commit yang sama,
- endpoint tanpa dokumentasi kontrak tidak boleh dipromosikan.

## Required Contract Scope

OpenAPI minimal harus mencakup:

- semua route publik di `/api/v1/**`,
- health endpoints yang terekspos,
- request/response schema,
- error responses standar,
- auth scheme (bearer/cookie/internal bila ada),
- contoh payload minimum per endpoint kritis.

## Contract Quality Rules

- `operationId` unik dan stabil,
- path parameter harus dideklarasikan eksplisit,
- schema harus reusable melalui `components/schemas`,
- response error mengikuti [API Response Standard](/Users/macbookpro/Development/recova-backend-v2/docs/api-response-standard.md),
- deprecated endpoint wajib diberi penanda `deprecated: true` sebelum removal.

## Drift Detection Rules

Drift terjadi jika route runtime berbeda dari kontrak OpenAPI.

Deteksi drift wajib:

- pada pull request,
- pada release candidate,
- setelah perubahan route registry.

Sinyal drift minimal:

- path/method ada di runtime tapi tidak ada di OpenAPI,
- path/method ada di OpenAPI tapi tidak terpasang di runtime,
- schema response utama berubah tanpa update kontrak.

## Review Gate

Perubahan OpenAPI harus ditinjau dari sisi:

- kompatibilitas backward,
- konsistensi error envelope,
- keamanan endpoint (auth + data exposure),
- kualitas contoh payload.

## Related Documents

- [API Reference](/Users/macbookpro/Development/recova-backend-v2/docs/api-reference.md)
- [Route Inventory](/Users/macbookpro/Development/recova-backend-v2/docs/generated/routes.md)
- [API Docs Generation](/Users/macbookpro/Development/recova-backend-v2/docs/operations/api-docs-generation.md)
- [API Response Standard](/Users/macbookpro/Development/recova-backend-v2/docs/api-response-standard.md)

## Source Reference

- [OpenAPI Specification Latest](https://spec.openapis.org/oas/latest.html)
- [OpenAPI Specification Repository](https://github.com/OAI/OpenAPI-Specification)
- [Fiber Routing Guide](https://docs.gofiber.io/guide/routing/)
- [Fiber Error Handling Guide](https://docs.gofiber.io/guide/error-handling/)
- [/Users/macbookpro/Development/bisakerja-api/docs/generated/routes.md](/Users/macbookpro/Development/bisakerja-api/docs/generated/routes.md)
