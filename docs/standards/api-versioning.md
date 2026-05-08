---
title: Recova Backend API Versioning Standard
description: Aturan versioning API publik, klasifikasi perubahan kontrak, dan syarat kompatibilitas pada jalur /api/v1.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/standards/api-versioning.md
last_reviewed: 2026-05-08
---

# Recova Backend API Versioning Standard

Dokumen ini menetapkan aturan versioning API publik backend Recova.

## Current Version Contract

- prefix publik aktif: `/api/v1`
- semua endpoint produk wajib ter-mount di bawah prefix ini
- perubahan kontrak harus mengevaluasi dampak klien sebelum rilis

## Change Classification

| Class                | Contoh                                                | Dampak versioning                              |
| -------------------- | ----------------------------------------------------- | ---------------------------------------------- |
| Non-breaking         | tambah field optional response                        | tetap di `v1`                                  |
| Behavioral-sensitive | perubahan semantic field lama                         | butuh compatibility review ketat               |
| Breaking             | ganti path/method, hapus field wajib, ubah tipe field | butuh versi baru atau migration plan eksplisit |

## Versioning Rules

- endpoint yang sudah dipakai klien aktif diprioritaskan kompatibel.
- perubahan breaking tidak boleh masuk tanpa keputusan compatibility yang tertulis.
- jika versi baru dibuka, endpoint lama punya window deprecation yang jelas.
- dokumentasi API wajib diperbarui di perubahan yang menyentuh kontrak.

## Deprecation Rules

- endpoint yang akan dipensiunkan harus diberi status deprecate di compatibility matrix.
- alasan dan rencana transisi harus tertulis sebelum implementasi perubahan.
- endpoint dev-only tidak boleh aktif pada lingkungan produksi.

## Contract Checklist Before Release

- endpoint path/method sesuai kontrak,
- response envelope sesuai standar,
- error code mapping sesuai taxonomy,
- compatibility status endpoint sudah diperbarui,
- gap dokumentasi tidak ditutup dengan asumsi.

## Related Documents

- [API Reference](/Users/macbookpro/Development/recova-backend-v2/docs/api-reference.md)
- [API Compatibility Matrix](/Users/macbookpro/Development/recova-backend-v2/docs/api-compatibility-matrix.md)
- [API Response Standard](/Users/macbookpro/Development/recova-backend-v2/docs/api-response-standard.md)

## Source Reference

- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
- [OpenAPI Specification](https://spec.openapis.org/oas/)
