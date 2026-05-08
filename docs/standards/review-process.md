---
title: Recova Backend Documentation Review Process
description: Proses review dokumentasi teknis Recova Backend, termasuk role, tanggung jawab, dan review gate untuk area kritis.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/standards/review-process.md
last_reviewed: 2026-05-08
---

# Recova Backend Documentation Review Process

Dokumen ini mendefinisikan governance review untuk dokumentasi teknis agar setiap kontrak penting memiliki owner, reviewer, dan gate persetujuan yang jelas.

## Default Role Assignment

| Peran                               | Tanggung jawab                                                                                       |
| ----------------------------------- | ---------------------------------------------------------------------------------------------------- |
| `owner`                             | Menjaga akurasi isi dokumen, memperbarui metadata, memicu review saat ada perubahan perilaku sistem. |
| `engineering-lead`                  | Memvalidasi akurasi arsitektur, konsistensi lintas modul, dan kesesuaian keputusan teknis.           |
| `platform-docs-maintainer`          | Memvalidasi kualitas dokumentasi, konsistensi struktur, metadata, dan navigasi.                      |
| Reviewer domain tambahan (opsional) | Memvalidasi area spesifik seperti security, deployment, atau database bila ada dampak domain tinggi. |

Default owner untuk dokumen baru adalah `backend-owner` sampai owner domain spesifik ditetapkan.

## Review Entry Criteria

Dokumen dapat masuk review bila:

- metadata lengkap dan valid,
- scope, kontrak, dan batasan ditulis jelas,
- referensi sumber utama tercantum,
- tautan antar dokumen valid.

## Review Gate Matrix

| Area dokumen | Trigger review wajib                                         | Reviewer minimum                                                           | Kriteria lulus                                                               |
| ------------ | ------------------------------------------------------------ | -------------------------------------------------------------------------- | ---------------------------------------------------------------------------- |
| Architecture | perubahan alur request, boundary layer, dependency direction | `engineering-lead`, `platform-docs-maintainer`                             | diagram/alur sinkron dengan implementasi target, failure path terdokumentasi |
| Auth         | perubahan login/session/token/ownership/authz                | `engineering-lead` + reviewer domain security (jika tersedia)              | aturan auth jelas, error contract auth terdokumentasi, risiko abuse dibahas  |
| Database     | perubahan schema, ownership data, migration behavior         | `engineering-lead` + reviewer domain database (jika tersedia)              | ownership tabel jelas, constraint utama dicatat, dampak migration terpetakan |
| API          | perubahan endpoint, payload, status code, error code         | `engineering-lead`, `platform-docs-maintainer`                             | request/response konsisten, backward compatibility jelas                     |
| Security     | perubahan kontrol keamanan, secret handling, hardening       | reviewer domain security (jika tersedia) + `engineering-lead`              | kontrol keamanan terukur, asumsi risiko eksplisit, mitigasi terdokumentasi   |
| Deployment   | perubahan release flow, rollback, runtime dependency         | `engineering-lead` + reviewer domain operations/deployment (jika tersedia) | prosedur deploy/rollback dapat dieksekusi, pre/post-check jelas              |

## Review Flow

1. Owner mengubah dokumen dan set `doc_status: review`.
2. Reviewer memeriksa isi teknis, metadata, dan referensi.
3. Jika ada gap, reviewer kembalikan ke owner untuk revisi.
4. Setelah gate lulus, owner ubah status ke `active` dan update `last_reviewed`.
5. Jika dokumen digantikan, ubah status ke `deprecated` dan tautkan dokumen pengganti.

## Blocking Conditions

Dokumen tidak boleh dipromosikan ke `active` bila:

- metadata wajib tidak lengkap,
- tidak ada owner atau reviewer role,
- kontrak teknis inti masih ambigu,
- ada conflict dengan dokumen aktif lain tanpa keputusan resolusi.

## Review Output Requirement

Hasil review minimal menyebutkan:

- status akhir (`active`, tetap `review`, atau `deprecated`),
- ringkasan gap dan tindakan lanjut,
- tanggal review (`last_reviewed`) yang diperbarui.

## Related Documents

- [Documentation Metadata Standard](/Users/macbookpro/Development/recova-backend-v2/docs/standards/metadata-standard.md)
- [Freshness and Lifecycle Standard](/Users/macbookpro/Development/recova-backend-v2/docs/standards/freshness-and-lifecycle.md)
- [Information Architecture Standard](/Users/macbookpro/Development/recova-backend-v2/docs/standards/information-architecture.md)

## Source Reference

- [/Users/macbookpro/Development/bisakerja-api/docs/overview.md](/Users/macbookpro/Development/bisakerja-api/docs/overview.md)
- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
