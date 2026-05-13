---
title: Scraper Requirement Gap Register
description: Register gap requirement untuk kebutuhan scraper atau ingestion konten eksternal pada backend Recova.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/roadmap/scraper-requirement-gap.md
last_reviewed: 2026-05-08
---

# Scraper Requirement Gap Register

Dokumen ini mencatat gap requirement terkait scraper/ingestion agar keputusan implementasi tidak didasarkan asumsi.

## Assessment Summary

Temuan dari sumber yang ada:

- belum ada kontrak eksplisit scraper pada service Recova,
- modul education dan daily content sudah ada sebagai consumer konten,
- belum ada definisi source eksternal, schedule, atau owner operasional ingestion.

Kesimpulan baseline:

- scraper belum menjadi kebutuhan terkonfirmasi,
- readiness dokumentasi disiapkan melalui template flow dan daftar keputusan yang harus ditutup.

## Gap List

| Gap ID  | Gap Description                                              | Impact                                       | Owner             | Status |
| ------- | ------------------------------------------------------------ | -------------------------------------------- | ----------------- | ------ |
| `SG-01` | Tidak ada daftar source eksternal resmi untuk konten         | ingestion tidak bisa dirancang deterministik | product + backend | open   |
| `SG-02` | Tidak ada kontrak field normalisasi untuk konten eksternal   | risiko mismatch schema dan kualitas konten   | backend           | open   |
| `SG-03` | Tidak ada jadwal ingestion dan SLA update konten             | freshness konten tidak terukur               | backend + ops     | open   |
| `SG-04` | Tidak ada keputusan owner moderasi untuk konten hasil ingest | risiko konten tidak aman dipublikasikan      | product + ops     | open   |
| `SG-05` | Tidak ada strategi rollback ingestion rusak                  | risiko konten salah tetap tampil             | backend + ops     | open   |

## Required Decisions Before Scraper Activation

- tetapkan source external yang diizinkan,
- tetapkan schema normalisasi konten,
- tetapkan policy publish dan moderasi,
- tetapkan observability dan alerting ingestion,
- tetapkan matrix ownership per environment.

## Interim Rule Until Gaps Closed

Sampai gap ditutup:

- gunakan seed/kurasi internal sebagai source konten,
- jangan aktifkan ingestion otomatis,
- dokumentasikan setiap perubahan source konten pada dokumen modul terkait.

## Related Documents

- [Data Flow Overview](/Users/macbookpro/Development/recova-backend-v2/docs/overview/data-flow.md)
- [Scraper Flow Integration](/Users/macbookpro/Development/recova-backend-v2/docs/integrations/scraper-flow.md)
- [Education Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/education.md)
- [Daily Content Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/daily-content.md)

## Source Reference

- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
- [/Users/macbookpro/Development/bisakerja-api/docs/integrations/scraper-api.md](/Users/macbookpro/Development/bisakerja-api/docs/integrations/scraper-api.md)
