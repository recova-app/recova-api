---
title: Journals Module
description: Kontrak modul jurnal pribadi untuk pembuatan dan pembacaan entri jurnal dengan kontrol privasi, ownership, dan retensi aman.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/modules/journals.md
last_reviewed: 2026-05-08
---

# Journals Module

## Responsibility

- membuat entri jurnal,
- menampilkan daftar entri jurnal milik user,
- menjaga privasi konten jurnal,
- menegakkan ownership data jurnal.

## API Contract

Route prefix:

```text
/api/v1/journals
```

| Method | Path               | Auth class | Purpose                        |
| ------ | ------------------ | ---------- | ------------------------------ |
| `GET`  | `/api/v1/journals` | Bearer     | ambil daftar jurnal milik user |
| `POST` | `/api/v1/journals` | Bearer     | buat entri jurnal baru         |

## Database Model

Entitas utama:

- `journals` (`id`, `user_id`, `content`, `created_at`, `updated_at`),
- index berdasarkan `user_id` dan waktu pembuatan.

Constraint minimum:

- jurnal selalu terkait satu `user_id`,
- skema delete (soft/hard) harus konsisten dengan kebijakan retensi.

## Authentication and Authorization

- seluruh endpoint journals wajib bearer auth,
- akses data hanya untuk `user_id` pemilik,
- parameter klien tidak boleh menimpa owner data.

## Service and Business Rules

- konten jurnal diperlakukan sebagai data sensitif,
- create/list menjaga urutan hasil konsisten,
- operasi penghapusan atau reset data harus audit-able jika diaktifkan.

## Validation Rules

- `content` wajib non-empty,
- batas panjang konten harus ditegakkan,
- field tidak dikenal ditolak sesuai kebijakan validator,
- payload invalid dipetakan ke `VALIDATION_ERROR`.

## Error Contract

| Condition                | HTTP  | Error code         |
| ------------------------ | ----- | ------------------ |
| auth invalid/missing     | `401` | `UNAUTHENTICATED`  |
| akses bukan milik user   | `403` | `FORBIDDEN`        |
| resource tidak ditemukan | `404` | `NOT_FOUND`        |
| payload invalid          | `422` | `VALIDATION_ERROR` |
| kegagalan internal       | `500` | `INTERNAL_ERROR`   |

## Observability Contract

Log field minimum:

- `request_id`,
- `user_id`,
- `journal_action`,
- `status_code`.

Metrik minimum:

- create journal rate,
- list journal latency,
- validation failure rate,
- error rate modul journals.

Konten mentah jurnal tidak boleh masuk log standar.

## Testing Requirements

- unit test validator jurnal,
- handler test auth + ownership,
- integration test persist/list jurnal,
- test redaction log untuk konten sensitif,
- contract test response envelope.

## Open Gaps

- keputusan final soft-delete vs hard-delete,
- kebutuhan enkripsi tambahan untuk konten jurnal,
- retention period final data jurnal.

## Related Documents

- [Users Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/users.md)
- [Data Sensitivity Matrix](/Users/macbookpro/Development/recova-backend-v2/docs/references/data-sensitivity-matrix.md)
- [API Response Standard](/Users/macbookpro/Development/recova-backend-v2/docs/api-response-standard.md)

## Source Reference

- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
