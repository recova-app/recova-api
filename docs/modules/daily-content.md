---
title: Daily Content Module
description: Kontrak modul konten harian untuk motivasi dan tantangan harian, termasuk aturan lifecycle, seed data, dan konsistensi publikasi.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/modules/daily-content.md
last_reviewed: 2026-05-08
---

# Daily Content Module

Modul daily-content menyediakan konten motivasi harian dan tantangan harian untuk menjaga engagement pengguna.

## Responsibility

- menyajikan konten harian aktif,
- menjamin satu set konten harian yang konsisten untuk tanggal tertentu,
- mengelola lifecycle aktif/nonaktif item konten,
- mendukung strategi seed data dan update berkala.

## Route Prefix

```text
/api/v1/content
```

## Endpoint Summary

| Method | Path                    | Auth   | Purpose                                          |
| ------ | ----------------------- | ------ | ------------------------------------------------ |
| `GET`  | `/api/v1/content/daily` | Bearer | Ambil motivasi harian dan tantangan harian aktif |

## Daily Payload Contract

Komponen minimum response:

- `date`,
- `motivation`:
  - `id`,
  - `title` atau `text`,
  - `status`,
- `challenge`:
  - `id`,
  - `title`,
  - `instruction`,
  - `status`.

## Publish and Lifecycle Rules

- hanya item berstatus `active` yang boleh disajikan,
- item `inactive` tidak boleh masuk payload harian,
- pergantian konten harian mengikuti boundary tanggal layanan yang terdokumentasi,
- perubahan konten harian harus audit-able.

## Seed Strategy

- seed data awal wajib menyediakan cakupan minimal beberapa hari,
- generator seed harus deterministik pada environment test,
- source eksternal untuk konten harian harus melewati validasi editorial sebelum aktif.

## Ownership Rules

- user tidak dapat menulis langsung daily content,
- daily content dikelola oleh jalur internal/editorial,
- data konsumsi user terhadap daily content (misal completion challenge) berada pada modul lain dan tidak mengubah source konten.

## Localization Direction

- dukung field `language_code` untuk ekspansi multi-bahasa,
- fallback bahasa harus eksplisit,
- translasi machine-only tanpa review tidak dipublikasikan.

## Observability Rules

- metrik minimum: request volume, latency, error rate,
- log konten harian cukup menyimpan `content_id`, `date`, `request_id`,
- hindari logging payload lengkap jika tidak diperlukan.

## Open Gaps

- timezone final untuk boundary konten harian lintas wilayah,
- strategi rotasi konten bila stok konten aktif habis,
- kontrak format final untuk instruksi challenge multimedia.

## Related Documents

- [Education Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/education.md)
- [Data Flow Overview](/Users/macbookpro/Development/recova-backend-v2/docs/overview/data-flow.md)
- [Scraper Requirement Gap Register](/Users/macbookpro/Development/recova-backend-v2/docs/roadmap/scraper-requirement-gap.md)

## Source Reference

- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
