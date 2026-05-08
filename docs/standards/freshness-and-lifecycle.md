---
title: Recova Backend Documentation Freshness and Lifecycle Standard
description: Standar status lifecycle, transisi status, dan aturan freshness review untuk menjaga dokumentasi Recova Backend tetap akurat.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/standards/freshness-and-lifecycle.md
last_reviewed: 2026-05-08
---

# Recova Backend Documentation Freshness and Lifecycle Standard

Dokumen ini menetapkan status lifecycle dokumen, cara transisi antarstatus, dan aturan freshness review agar dokumentasi tetap relevan dengan perilaku sistem.

## Lifecycle Status

| Status       | Arti                                  | Kapan dipakai                                                            |
| ------------ | ------------------------------------- | ------------------------------------------------------------------------ |
| `draft`      | konten awal, belum direview formal    | dokumen baru atau perubahan besar belum melalui review lintas peran      |
| `review`     | konten siap ditinjau                  | konten sudah lengkap, menunggu persetujuan reviewer wajib                |
| `active`     | konten menjadi acuan operasional      | reviewer wajib menyetujui dan isi mencerminkan perilaku sistem saat ini  |
| `deprecated` | konten tidak lagi menjadi acuan utama | perilaku diganti, endpoint ditutup, atau dokumen digantikan dokumen lain |

## Transition Rule

| Dari         | Ke           | Syarat minimum                                                 |
| ------------ | ------------ | -------------------------------------------------------------- |
| `draft`      | `review`     | struktur lengkap, metadata valid, referensi sumber tercantum   |
| `review`     | `active`     | review gate lulus dan tidak ada blocker terbuka                |
| `active`     | `review`     | ada perubahan perilaku signifikan yang belum tervalidasi ulang |
| `active`     | `deprecated` | ada pengganti resmi atau fitur sudah dihentikan                |
| `deprecated` | `review`     | dokumen dihidupkan kembali karena kontrak kembali berlaku      |

## Freshness Policy

- `last_reviewed` wajib diupdate saat review substansial selesai.
- Dokumen `active` direview berkala maksimal tiap 90 hari.
- Dokumen pada domain berisiko tinggi (`security`, `auth`, `deployment`) direview maksimal tiap 30 hari.
- Dokumen `draft` yang tidak diperbarui lebih dari 30 hari harus diputuskan: lanjutkan ke `review` atau tandai `deprecated`.

## Staleness Handling

Dokumen dianggap stale bila melewati interval review yang berlaku.

Tindakan minimum:

1. Ubah `doc_status` ke `review` bila isi masih relevan namun butuh validasi ulang.
2. Tambah catatan gap di isi dokumen jika ada bagian belum terverifikasi.
3. Jadwalkan review oleh owner dan reviewer role terkait.
4. Bila kontrak tidak relevan lagi, pindahkan ke `deprecated` dan tautkan pengganti.

## Lifecycle Ownership

- Owner bertanggung jawab memicu review saat ada perubahan kode, kontrak API, skema data, atau operasi.
- Reviewer bertanggung jawab memvalidasi akurasi lintas domain sebelum status `active`.
- Perubahan besar tanpa update lifecycle dianggap drift dokumentasi dan harus diperbaiki pada work item yang sama.

## Review Cadence Matrix

| Kategori dokumen               | Status target normal  | Maksimum umur review |
| ------------------------------ | --------------------- | -------------------- |
| Architecture                   | `active`              | 90 hari              |
| API contract                   | `active`              | 90 hari              |
| Database schema/contract       | `active`              | 90 hari              |
| Security & auth                | `active`              | 30 hari              |
| Deployment & operations kritis | `active`              | 30 hari              |
| Roadmap/non-operational plan   | `draft` atau `review` | 90 hari              |

## Related Documents

- [Documentation Metadata Standard](/Users/macbookpro/Development/recova-backend-v2/docs/standards/metadata-standard.md)
- [Review Process Standard](/Users/macbookpro/Development/recova-backend-v2/docs/standards/review-process.md)
- [Standards Index](/Users/macbookpro/Development/recova-backend-v2/docs/standards/index.md)

## Source Reference

- [/Users/macbookpro/Development/bisakerja-api/docs/overview.md](/Users/macbookpro/Development/bisakerja-api/docs/overview.md)
- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
