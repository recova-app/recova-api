---
title: Scraper Flow Integration
description: Kontrak kesiapan alur scraper atau ingestion konten eksternal untuk education, daily content, dan enrichment data AI secara aman.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/integrations/scraper-flow.md
last_reviewed: 2026-05-08
---

# Scraper Flow Integration

Dokumen ini menetapkan kontrak kesiapan jika backend membutuhkan scraper atau pipeline ingestion konten eksternal.

## Current Baseline

Berdasarkan sumber saat ini:

- layanan memiliki modul education dan daily content,
- sumber belum mendefinisikan scraper otomatis khusus Recova,
- ingestion baseline dijalankan melalui seed data atau kurasi internal.

Karena itu, integrasi scraper diperlakukan sebagai future capability dengan gap yang terdokumentasi.

## Future Flow Template

Jika scraper diaktifkan, alur minimum yang harus dipenuhi:

```text
scheduler/manual trigger
  -> ingestion worker
      -> external source fetch
      -> normalize + validate
      -> safety/moderation pre-publish
      -> upsert content storage
      -> publish active content
      -> emit ingestion metrics and audit event
```

## Integration Contract Requirements

Kontrak minimum sebelum aktif:

- source external harus terdaftar dan tervalidasi,
- skema normalisasi konten harus terdokumentasi,
- status lifecycle (`active`/`inactive`) harus diterapkan,
- proses publish harus idempotent,
- error ingestion tidak boleh menjatuhkan endpoint user-facing.

## Security and Privacy Rules

- secret source ingestion disimpan sebagai env secret,
- token source tidak boleh dicatat pada log,
- payload raw disimpan terbatas sesuai kebutuhan debugging,
- konten yang melibatkan data personal harus melalui redaction/pseudonymization.

## Operational Controls

- timeout, retry, dan dead-letter strategy harus terdokumentasi,
- observability minimum: run status, records processed, records rejected, processing latency,
- rollback strategy wajib tersedia bila source mengirim data rusak.

## Data Quality Gates

Sebelum konten dipublikasikan:

- cek field wajib tidak kosong,
- cek konten duplikat,
- cek status moderasi,
- cek format tanggal/waktu dan locale konten.

## Gap Register Linkage

Detail gap requirement ada di:

- [Scraper Requirement Gap Register](/Users/macbookpro/Development/recova-backend-v2/docs/roadmap/scraper-requirement-gap.md)

## Related Documents

- [Data Flow Overview](/Users/macbookpro/Development/recova-backend-v2/docs/overview/data-flow.md)
- [Education Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/education.md)
- [Daily Content Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/daily-content.md)

## Source Reference

- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
- [/Users/macbookpro/Development/bisakerja-api/docs/integrations/scraper-api.md](/Users/macbookpro/Development/bisakerja-api/docs/integrations/scraper-api.md)
