---
title: Recova Backend Documentation Metadata Standard
description: Standar metadata frontmatter untuk seluruh dokumentasi teknis Recova Backend agar punya owner, status, dan jejak review yang jelas.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/standards/metadata-standard.md
last_reviewed: 2026-05-08
---

# Recova Backend Documentation Metadata Standard

Dokumen ini menetapkan format metadata wajib pada setiap dokumen di `docs/**` agar kepemilikan, status, dan akuntabilitas review dapat dilacak konsisten.

## Required Frontmatter

Semua dokumen wajib memiliki blok frontmatter YAML di bagian paling atas.

```md
---
title: Nama Halaman
description: Ringkasan satu kalimat tentang isi halaman.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/example.md
last_reviewed: 2026-05-08
---
```

## Field Definition

| Field           | Type              | Wajib | Aturan                                                                                       |
| --------------- | ----------------- | ----- | -------------------------------------------------------------------------------------------- |
| `title`         | string            | ya    | Judul dokumen yang deskriptif, tidak ambigu, sesuai isi halaman.                             |
| `description`   | string            | ya    | Ringkasan satu kalimat, fokus pada kontrak/tujuan teknis dokumen.                            |
| `owner`         | string            | ya    | Pemilik utama akurasi dokumen; default `backend-owner` bila belum ada owner domain spesifik. |
| `reviewers`     | list[string]      | ya    | Daftar role reviewer; minimal 1 role.                                                        |
| `doc_status`    | string            | ya    | Salah satu: `draft`, `review`, `active`, `deprecated`.                                       |
| `source_repo`   | string            | ya    | Nama repository sumber dokumen; untuk repo ini: `recova-backend-v2`.                         |
| `source_path`   | string            | ya    | Path relatif file di repository ini, harus sama dengan lokasi aktual file.                   |
| `last_reviewed` | date (YYYY-MM-DD) | ya    | Tanggal review substansial terakhir.                                                         |

## Metadata Rules

- Frontmatter harus berada di awal file tanpa konten lain di atasnya.
- `source_path` harus diperbarui bila file dipindahkan atau di-rename.
- `last_reviewed` diperbarui saat ada perubahan substansial atau review formal.
- `reviewers` harus berisi role, bukan nama personal, agar rotasi tim tidak merusak governance.
- Dokumen tanpa metadata lengkap tidak boleh dipromosikan ke status `active`.

## Default Values

Default yang dipakai saat membuat dokumen baru:

- `owner: backend-owner`
- `reviewers:`
  - `engineering-lead`
  - `platform-docs-maintainer`
- `doc_status: draft`
- `source_repo: recova-backend-v2`

Role tambahan dapat ditambahkan sesuai domain (misalnya `security-lead` atau `platform-engineer`) tanpa menghapus role reviewer inti bila masih relevan.

## Validation Checklist

Sebelum merge perubahan dokumen:

- semua field wajib tersedia,
- `doc_status` valid,
- `source_path` akurat,
- `last_reviewed` format tanggal valid,
- owner dan reviewer sesuai domain dokumen.

## Related Documents

- [Freshness and Lifecycle Standard](/Users/macbookpro/Development/recova-backend-v2/docs/standards/freshness-and-lifecycle.md)
- [Review Process Standard](/Users/macbookpro/Development/recova-backend-v2/docs/standards/review-process.md)
- [Information Architecture Standard](/Users/macbookpro/Development/recova-backend-v2/docs/standards/information-architecture.md)
- [Documentation Overview](/Users/macbookpro/Development/recova-backend-v2/docs/overview.md)

## Source Reference

- [/Users/macbookpro/Development/bisakerja-api/docs/overview.md](/Users/macbookpro/Development/bisakerja-api/docs/overview.md)
- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
