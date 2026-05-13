---
title: Recova Backend Documentation Quality Audit
description: Metode audit kualitas dokumentasi untuk memverifikasi metadata, konsistensi istilah, validitas referensi, dan kelayakan dokumen sebagai kontrak teknis.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/roadmap/documentation-quality-audit.md
last_reviewed: 2026-05-08
---

# Recova Backend Documentation Quality Audit

Dokumen ini mendefinisikan cara audit kualitas untuk seluruh dokumen teknis agar siap dipakai sebagai sumber keputusan implementasi.

## Audit Scope

- seluruh dokumen pada `docs/**`,
- fokus pada dokumen kontrak: architecture, API, database, security, operations, modules, roadmap,
- dokumen generated ikut dicek untuk sinkronisasi sumber.

## Quality Gates

| Gate                    | Kriteria pass                                  |
| ----------------------- | ---------------------------------------------- |
| Metadata completeness   | semua field frontmatter wajib terisi dan valid |
| Link and path validity  | tautan internal aktif, path sumber akurat      |
| Terminology consistency | istilah teknis konsisten lintas dokumen        |
| Standalone quality      | isi dapat dipahami tanpa konteks task internal |
| Gap traceability        | semua gap kritis punya owner dan next action   |
| Contract coherence      | tidak ada kontradiksi besar antar dokumen inti |

## Audit Procedure

1. kumpulkan daftar dokumen target berdasarkan kategori,
2. validasi frontmatter terhadap metadata standard,
3. validasi tautan internal dan referensi sumber,
4. audit terminologi lintas dokumen inti,
5. review gap register dan status penyelesaiannya,
6. keluarkan status audit: `pass`, `pass-with-action`, atau `fail`.

## Severity Model

| Severity | Definisi                                                   | Dampak                               |
| -------- | ---------------------------------------------------------- | ------------------------------------ |
| Critical | kontradiksi kontrak atau metadata hilang pada dokumen inti | blok implementasi                    |
| High     | gap kritis tanpa owner/aksi, link penting rusak            | blok review lanjutan                 |
| Medium   | inkonsistensi istilah atau struktur yang membingungkan     | butuh perbaikan sebelum status aktif |
| Low      | masalah editorial minor                                    | tidak memblokir, tetap dicatat       |

## Standalone Content Check

Dokumen dianggap standalone jika:

- tidak bergantung pada catatan eksekusi internal,
- tidak menggunakan referensi implisit ke pekerjaan sementara,
- menjelaskan kontrak teknis dan batasan secara langsung,
- tidak mencampur instruksi tooling yang tidak relevan dengan isi teknis.

## Audit Output Template

Setiap audit menghasilkan:

- ringkasan cakupan dokumen,
- daftar temuan per severity,
- daftar tindakan dan owner,
- batas waktu koreksi,
- keputusan status akhir audit.

## Re-Audit Trigger

Audit ulang wajib dijalankan ketika:

- ada perubahan besar pada kontrak API/DB/security,
- ada penambahan domain modul baru,
- terjadi insiden yang menunjukkan dokumentasi tidak akurat,
- ada pergantian kebijakan deploy atau auth.

## Related Documents

- [Benchmark Parity Report](/Users/macbookpro/Development/recova-backend-v2/docs/roadmap/benchmark-parity-report.md)
- [Documentation Maintenance Standard](/Users/macbookpro/Development/recova-backend-v2/docs/standards/documentation-maintenance.md)
- [Metadata Standard](/Users/macbookpro/Development/recova-backend-v2/docs/standards/metadata-standard.md)

## Source Reference

- [tasks/lessons.md](/Users/macbookpro/Development/recova-backend-v2/tasks/lessons.md)
- [/Users/macbookpro/Development/bisakerja-api/docs/overview.md](/Users/macbookpro/Development/bisakerja-api/docs/overview.md)
- [/Users/macbookpro/Development/bisakerja-api/docs/architecture.md](/Users/macbookpro/Development/bisakerja-api/docs/architecture.md)
