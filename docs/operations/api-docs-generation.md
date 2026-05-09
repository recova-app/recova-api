---
title: Recova Backend API Docs Generation
description: Runbook generasi dokumentasi API untuk menghasilkan artefak OpenAPI dan route inventory tanpa ketergantungan secret production.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/operations/api-docs-generation.md
last_reviewed: 2026-05-09
---

# Recova Backend API Docs Generation

Dokumen ini menetapkan proses generasi dokumentasi API yang deterministik dan aman dijalankan di local maupun CI.

## Output Artifacts

Artefak minimal:

- source spec: `api/openapi/openapi.yaml`,
- generated spec: `docs/generated/openapi.yaml`,
- `docs/generated/routes.md`,
- Scalar docs config: `scalar.config.json`.

## Security Rule

Generasi dokumentasi API tidak boleh membutuhkan:

- secret production,
- koneksi ke database production,
- akses ke layanan eksternal berprivilege.

Gunakan konfigurasi minimal non-sensitive untuk proses build/generate.

## Standard Generation Flow

```text
1) Load safe environment profile
2) Validate source OpenAPI spec
3) Sync source spec -> generated spec
4) Build route registry dari runtime Go Fiber
5) Generate route inventory
6) Run contract lint/validation + drift check
7) Compare generated files against repository state
```

## Standard Commands

Gunakan command berikut:

- `make openapi-generate`
- `make openapi-check`
- `make scalar-check`
- `make scalar-preview` (untuk membuka runtime docs di `/docs/api`)

Command dijalankan via wrapper berikut:

- `scripts/openapi.sh` untuk generate/check OpenAPI,
- `scripts/scalar.sh` untuk validasi config Scalar dan preview runtime docs.

## Contract Drift Verification

Drift check wajib mendeteksi:

- route baru tanpa kontrak,
- route kontrak yang sudah tidak ada,
- perubahan schema utama tanpa update artefak.

Jika drift terdeteksi:

- tandai review sebagai gagal,
- perbarui artefak atau route source,
- ulang validasi sampai sinkron.

## CI Gate Expectation

Pada pipeline CI, job dokumentasi API minimal menjalankan:

- generate artifact,
- validate schema OpenAPI,
- validate route inventory,
- validate `scalar.config.json` dan keterbacaan semua filepath,
- fail jika ada perubahan artefak yang belum dikomit.

## Manual Verification Checklist

- artefak OpenAPI bisa dibaca tool validator,
- route inventory sesuai endpoint yang didokumentasikan,
- `scalar.config.json` memetakan route `type: openapi` ke `docs/generated/openapi.yaml`,
- route runtime `/openapi.yaml` dan `/docs/api` tersedia serta tidak memuat credential/token nyata,
- section auth dan error response tetap konsisten,
- metadata file (`last_reviewed`, `source_path`) valid.

## Related Documents

- [OpenAPI Standard](/Users/macbookpro/Development/recova-backend-v2/docs/standards/openapi.md)
- [Route Inventory](/Users/macbookpro/Development/recova-backend-v2/docs/generated/routes.md)
- [API Reference](/Users/macbookpro/Development/recova-backend-v2/docs/api-reference.md)
- [Testing Strategy](/Users/macbookpro/Development/recova-backend-v2/docs/operations/testing.md)

## Source Reference

- [OpenAPI Specification Latest](https://spec.openapis.org/oas/latest.html)
- [OpenAPI Initiative](https://www.openapis.org/)
- [Fiber Routing Guide](https://docs.gofiber.io/guide/routing/)
- [Scalar Docs `scalar.config.json`](https://scalar.com/products/docs/configuration/scalar.config.json)
- [Scalar Docs Navigation](https://scalar.com/products/docs/configuration/navigation)
- [/Users/macbookpro/Development/bisakerja-api/docs/generated/routes.md](/Users/macbookpro/Development/bisakerja-api/docs/generated/routes.md)
