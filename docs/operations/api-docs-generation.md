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
last_reviewed: 2026-05-08
---

# Recova Backend API Docs Generation

Dokumen ini menetapkan proses generasi dokumentasi API yang deterministik dan aman dijalankan di local maupun CI.

## Output Artifacts

Artefak minimal:

- `docs/generated/openapi.yaml` atau `docs/generated/openapi.json`,
- `docs/generated/routes.md`.

## Security Rule

Generasi dokumentasi API tidak boleh membutuhkan:

- secret production,
- koneksi ke database production,
- akses ke layanan eksternal berprivilege.

Gunakan konfigurasi minimal non-sensitive untuk proses build/generate.

## Standard Generation Flow

```text
1) Load safe environment profile
2) Build route registry or contract source
3) Generate OpenAPI artifact
4) Generate route inventory
5) Run contract lint/validation
6) Compare generated files against repository state
```

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
- fail jika ada perubahan artefak yang belum dikomit.

## Manual Verification Checklist

- artefak OpenAPI bisa dibaca tool validator,
- route inventory sesuai endpoint yang didokumentasikan,
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
- [/Users/macbookpro/Development/bisakerja-api/docs/generated/routes.md](/Users/macbookpro/Development/bisakerja-api/docs/generated/routes.md)
