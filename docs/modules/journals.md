---
title: Journals Module
description: Kontrak modul jurnal pribadi untuk create/list entry, ownership pengguna, klasifikasi privasi, serta kebijakan retensi dan penghapusan.
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

Modul journals mengelola catatan refleksi pribadi pengguna dengan kontrol privasi ketat.

## Responsibility

- membuat entri jurnal baru,
- menampilkan daftar entri jurnal milik pengguna,
- memastikan kepemilikan data pada setiap operasi,
- menerapkan batas privasi pada storage dan logging.

## Route Prefix

```text
/api/v1/journals
```

## Endpoint Summary

| Method | Path               | Auth   | Purpose                           |
| ------ | ------------------ | ------ | --------------------------------- |
| `GET`  | `/api/v1/journals` | Bearer | Ambil daftar entri jurnal sendiri |
| `POST` | `/api/v1/journals` | Bearer | Buat entri jurnal baru            |

## Data Contract

Field minimum entri jurnal:

- `content`,
- `created_at`,
- `updated_at`.

Field opsional masa depan:

- `title`,
- `tags`,
- `mood_marker`.

## Ownership Rules

- user hanya boleh membaca dan menulis jurnal miliknya sendiri,
- `user_id` sumber kebenaran diambil dari auth context,
- parameter dari client tidak boleh mengganti owner data.

## Privacy Classification

Konten jurnal diklasifikasikan sebagai data sensitif pribadi.

Aturan:

- tidak boleh masuk log aplikasi sebagai teks mentah,
- tidak boleh dikirim ke sistem eksternal tanpa kontrak eksplisit,
- error response tidak boleh memantulkan konten jurnal.

## Retention and Delete Policy

Arah kebijakan retensi:

- default: data jurnal dipertahankan sampai pengguna menghapus,
- kebijakan auto-delete jika dibutuhkan harus terdokumentasi eksplisit,
- delete policy final (soft delete/hard delete) harus konsisten dengan kebijakan privasi global.

## Future Search and Filter Direction

Ruang pengembangan:

- filter berdasarkan rentang tanggal,
- full-text search terbatas,
- agregasi statistik journaling non-personal.

Penambahan fitur ini wajib disertai kajian dampak privasi dan index strategy.

## Security and Observability Rules

- log operasional hanya menyimpan metadata (request id, user id, jumlah record),
- audit event untuk create/delete jurnal dicatat tanpa konten,
- response time dan error rate jurnal tetap dipantau sebagai metrik operasional.

## Open Gaps

- batas panjang final `content`,
- keputusan final soft delete vs hard delete,
- strategi enkripsi at-rest per kolom untuk konten jurnal.

## Related Documents

- [Users Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/users.md)
- [Data Sensitivity Matrix](/Users/macbookpro/Development/recova-backend-v2/docs/references/data-sensitivity-matrix.md)
- [API Response Standard](/Users/macbookpro/Development/recova-backend-v2/docs/api-response-standard.md)

## Source Reference

- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
