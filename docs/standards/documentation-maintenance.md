---
title: Recova Backend Documentation Maintenance Standard
description: Standar pemeliharaan dokumentasi meliputi cadence review, proses update ADR, pengelolaan docs debt, dan kontrol freshness lintas domain.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/standards/documentation-maintenance.md
last_reviewed: 2026-05-08
---

# Recova Backend Documentation Maintenance Standard

Dokumen ini mendefinisikan standar pemeliharaan agar dokumentasi tetap akurat sepanjang siklus pengembangan.

## Maintenance Principles

- dokumentasi adalah bagian dari definisi selesai pekerjaan,
- update dokumen dilakukan bersamaan dengan perubahan perilaku sistem,
- kualitas dokumen diukur melalui audit yang berulang dan terukur.

## Ownership Model

| Peran                  | Tanggung jawab                                |
| ---------------------- | --------------------------------------------- |
| Owner dokumen          | menjaga akurasi konten domain                 |
| Reviewer lintas domain | memastikan konsistensi kontrak lintas dokumen |
| Engineering lead       | memutuskan prioritas docs debt kritis         |

## Review Cadence

| Kategori                             | Cadence maksimum                |
| ------------------------------------ | ------------------------------- |
| Security, auth, privacy, deployment  | 30 hari                         |
| API, database, architecture, modules | 90 hari                         |
| Roadmap dan governance               | 90 hari atau saat scope berubah |

Jika dokumen melewati cadence, `doc_status` diturunkan ke `review` sampai tervalidasi ulang.

## ADR Maintenance Rule

- keputusan baru yang memengaruhi arsitektur harus masuk ADR baru atau update ADR existing,
- ADR lama yang tidak berlaku diberi status pengganti dan ditautkan ke ADR baru,
- dokumen operasional dan standar terkait wajib diperbarui agar tidak bertentangan.

## OpenAPI and API Docs Cadence

- route change: update route inventory pada work item yang sama,
- contract change: update sumber OpenAPI pada work item yang sama,
- release candidate: jalankan validasi drift API docs.

## Documentation Debt Management

Debt dokumentasi harus punya:

- deskripsi gap,
- severity,
- owner,
- target perbaikan,
- status progres.

Prioritas eksekusi:

1. critical
2. high
3. medium
4. low

## Maintenance Automation Gate

Untuk menjaga cadence review tetap konsisten, jalankan gate maintenance:

- `make post-migration-maintenance`

Gate ini menghasilkan evidence:

- report freshness dokumen,
- backlog maintenance dengan owner + priority,
- ringkasan review alert/SLO/dependency cadence.

## Standalone Quality Rule

Dokumen pemeliharaan wajib memastikan:

- konten dapat dipahami tanpa catatan kerja internal,
- istilah dan kontrak konsisten lintas dokumen,
- referensi teknis mengarah ke sumber resmi dan valid.

## Maintenance Verification

Checklist periodik:

- [ ] dokumen `active` telah direview sesuai cadence.
- [ ] link/path internal valid.
- [ ] metadata wajib lengkap dan akurat.
- [ ] docs debt register diperbarui.
- [ ] konflik antar dokumen inti terselesaikan.
- [ ] perubahan file/config baru memiliki test companion atau exception rationale terdokumentasi.

## Related Documents

- [Freshness and Lifecycle Standard](/Users/macbookpro/Development/recova-backend-v2/docs/standards/freshness-and-lifecycle.md)
- [Documentation Metadata Standard](/Users/macbookpro/Development/recova-backend-v2/docs/standards/metadata-standard.md)
- [Documentation Sync Operations](/Users/macbookpro/Development/recova-backend-v2/docs/operations/documentation-sync.md)
- [Documentation Quality Audit](/Users/macbookpro/Development/recova-backend-v2/docs/roadmap/documentation-quality-audit.md)

## Source Reference

- [/Users/macbookpro/Development/bisakerja-api/docs/operations/documentation-sync.md](/Users/macbookpro/Development/bisakerja-api/docs/operations/documentation-sync.md)
- [tasks/lessons.md](/Users/macbookpro/Development/recova-backend-v2/tasks/lessons.md)
- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
