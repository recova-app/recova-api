---
title: Recova Backend Information Architecture Standard
description: Standar struktur, navigasi, naming, dan ownership dokumentasi teknis di repository Recova Backend.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/standards/information-architecture.md
last_reviewed: 2026-05-08
---

# Recova Backend Information Architecture Standard

Dokumen ini menetapkan standar informasi untuk seluruh `docs/**` agar konsisten, mudah dinavigasi, dan siap dipakai sebagai referensi teknis mandiri.

## Top-Level Structure

Struktur root dokumentasi:

- `docs/overview.md`
- `docs/architecture.md`
- `docs/database.md`
- `docs/api/index.md`
- `docs/modules/index.md`
- `docs/operations/index.md`
- `docs/standards/index.md`
- `docs/integrations/index.md`
- `docs/decisions/index.md`
- `docs/roadmap/index.md`

Setiap section harus punya landing page `index.md` (atau file root setara untuk section tunggal) yang menjelaskan cakupan section dan tautan dokumen turunannya.

## Naming Standard

- Gunakan `kebab-case` untuk nama file dan folder.
- Gunakan suffix `index.md` untuk landing page section.
- Gunakan nama deskriptif berbasis domain, bukan berbasis aktivitas internal.

Contoh:

- `api-response-standard.md`
- `auth-session-architecture.md`
- `migration-rollback-policy.md`

## Content Standard

Setiap landing page minimal memuat:

- tujuan section,
- scope informasi,
- tautan ke dokumen terkait.

Setiap dokumen detail minimal memuat:

- konteks masalah,
- keputusan/kontrak teknis,
- dampak ke API, data, operasi, atau keamanan (jika relevan),
- referensi sumber.

## Ownership Boundary

- `docs/**` berisi dokumentasi teknis layanan.
- `tasks/**` berisi catatan eksekusi kerja internal.
- Dokumen di `docs/**` tidak boleh bergantung pada keberadaan dokumen di `tasks/**`.

## Metadata Requirement

Setiap file dalam `docs/**` wajib memiliki frontmatter:

- `title`
- `description`
- `owner`
- `reviewers`
- `doc_status`
- `source_repo`
- `source_path`
- `last_reviewed`

## Navigation Rule

- Landing page top-level harus saling terhubung.
- Dokumen detail harus punya link balik ke landing page section.
- Broken link tidak diperbolehkan.

## Review Rule

- `doc_status: draft` dipakai sampai isi dan referensi tervalidasi.
- `doc_status: active` dipakai saat isi telah disetujui reviewer.
- `last_reviewed` wajib diperbarui saat ada review substansial.

## Source Reference

Standar ini disusun dengan acuan:

- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
- [/Users/macbookpro/Development/bisakerja-api/docs/overview.md](/Users/macbookpro/Development/bisakerja-api/docs/overview.md)
- [/Users/macbookpro/Development/bisakerja-api/docs/project-structure.md](/Users/macbookpro/Development/bisakerja-api/docs/project-structure.md)
- [Go module layout guidance](https://go.dev/doc/modules/layout)
