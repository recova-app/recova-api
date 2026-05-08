---
title: Education Module
description: Kontrak modul konten edukasi untuk penyajian daftar materi, lifecycle aktif/nonaktif, dan ownership pengelolaan konten.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/modules/education.md
last_reviewed: 2026-05-08
---

# Education Module

Modul education menyediakan konten edukasi yang relevan untuk perjalanan pemulihan pengguna.

## Responsibility

Modul education bertanggung jawab pada:

- menyajikan daftar konten edukasi aktif,
- menjaga metadata konten agar konsisten,
- mendukung lifecycle aktif/nonaktif konten,
- menjaga batas ownership antara data konten dan modul konsumennya.

## Route Prefix

```text
/api/v1/education
```

## Endpoint Summary

| Method | Path                | Auth   | Purpose                           |
| ------ | ------------------- | ------ | --------------------------------- |
| `GET`  | `/api/v1/education` | Bearer | Ambil daftar konten edukasi aktif |

## Content Contract

Field minimum konten edukasi:

- `id`,
- `title`,
- `summary`,
- `content_body` atau `content_url`,
- `status` (`active`/`inactive`),
- `published_at`,
- `updated_at`.

Field opsional arah lanjut:

- `tags`,
- `difficulty_level`,
- `reading_time_minutes`,
- `language_code`.

## Content Lifecycle

Aturan lifecycle baseline:

- hanya konten `active` yang muncul di endpoint publik,
- konten `inactive` tetap tersimpan untuk audit histori,
- perubahan status konten harus tercatat pada audit event,
- konten yang ditarik tidak boleh hilang dari referensi internal tanpa jejak.

## Seed and Source Strategy

Arah baseline:

- konten awal dapat dimuat melalui seed data terkurasi,
- sumber konten eksternal harus melalui proses verifikasi editorial sebelum dipublikasi,
- jika ingestion otomatis dipakai di masa depan, jalurnya harus mengikuti dokumen data-flow dan integrasi scraper.

## Ownership Model

- source of truth konten berada pada storage backend,
- modul client hanya mengonsumsi payload yang sudah dipublikasi,
- perubahan konten dilakukan melalui jalur admin/internal yang terdokumentasi terpisah.

## Localization Direction

- payload dianjurkan menyertakan `language_code`,
- fallback bahasa harus eksplisit bila konten lokal tidak tersedia,
- translasi otomatis tanpa review editorial tidak boleh langsung dipublikasikan.

## Observability Rules

- log akses cukup menyimpan metadata (`content_id`, `request_id`, status),
- jangan simpan payload konten penuh pada log request standar,
- metrik minimum: request count, latency, error rate.

## Open Gaps

- kontrak final apakah endpoint education bersifat public tanpa auth,
- format final `content_body` vs `content_url` untuk berbagai tipe konten,
- kebutuhan pagination dan filtering awal.

## Related Documents

- [Daily Content Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/daily-content.md)
- [Data Flow Overview](/Users/macbookpro/Development/recova-backend-v2/docs/overview/data-flow.md)
- [Scraper Flow Integration](/Users/macbookpro/Development/recova-backend-v2/docs/integrations/scraper-flow.md)

## Source Reference

- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
- [/Users/macbookpro/Development/bisakerja-api/docs/integrations/scraper-api.md](/Users/macbookpro/Development/bisakerja-api/docs/integrations/scraper-api.md)
