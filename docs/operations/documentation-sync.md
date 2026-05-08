---
title: Recova Backend Documentation Sync Operations
description: Aturan sinkronisasi perubahan implementasi dan dokumentasi agar kontrak teknis tetap konsisten selama pengembangan layanan.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/operations/documentation-sync.md
last_reviewed: 2026-05-08
---

# Recova Backend Documentation Sync Operations

Dokumen ini mengatur sinkronisasi perubahan kode, kontrak teknis, dan artefak dokumentasi operasional.

## Sync Objectives

- mencegah drift antara implementasi dan dokumentasi,
- memastikan perubahan kontrak terdokumentasi di work item yang sama,
- menjaga jejak review tetap audit-able.

## Mandatory Sync Rule

Setiap perubahan berikut wajib diikuti update dokumen terkait:

- kontrak API atau response envelope,
- auth/security behavior,
- model data, migration, dan retention behavior,
- deployment/rollback/testing workflow,
- observability dan log redaction policy.

Perubahan dianggap belum selesai bila dokumen terdampak belum diperbarui.

Aturan tambahan:

- setiap file kode atau konfigurasi baru default wajib punya test companion,
- jika test tidak dibuat, wajib ada exception rationale singkat yang menjelaskan alasan teknis dan bukti verifikasi alternatif.

## Change-to-Docs Mapping

| Jenis perubahan             | Dokumen minimum yang harus diperbarui                                                                            |
| --------------------------- | ---------------------------------------------------------------------------------------------------------------- |
| Endpoint/contract API       | `docs/api-reference.md`, `docs/generated/routes.md`, dokumen modul terkait                                       |
| Auth/security               | `docs/operations/security.md`, `docs/modules/auth.md`, standar terkait                                           |
| Database/schema             | `docs/database.md`, `docs/operations/database-migrations.md`                                                     |
| Deployment/runtime          | `docs/operations/deployment.md`, `docs/operations/rollback.md`, post-deploy checks                               |
| Privacy/logging             | `docs/operations/data-privacy.md`, `docs/operations/data-retention.md`, `docs/standards/log-redaction-policy.md` |
| Release validation evidence | `docs/generated/release-confidence-report.md` atau artefak report release validation yang ekuivalen              |

## Pull Request Review Gate

PR review wajib memeriksa:

- apakah ada perubahan kontrak yang belum di-dokumentasikan,
- apakah metadata dokumen valid dan `last_reviewed` diperbarui,
- apakah link internal dokumen tetap valid,
- apakah contoh payload aman dari data sensitif.
- apakah perubahan file/config memiliki test companion atau exception rationale.
- apakah report release confidence (E2E + performance smoke) terlampir untuk release candidate.

## OpenAPI and Route Inventory Cadence

- setiap perubahan route wajib memperbarui route inventory,
- setiap perubahan kontrak request/response wajib memperbarui sumber OpenAPI,
- validasi drift harus dijalankan sebelum merge.

## ADR Sync Rule

Jika keputusan arsitektur atau tooling inti berubah:

- buat atau update dokumen ADR terkait,
- tautkan ADR pada dokumen operasional/standar yang terdampak,
- pastikan keputusan lama diberi status superseded jika tidak berlaku.

## Documentation Debt Triage

Debt dokumentasi harus dicatat sebagai item terstruktur:

- severity (`critical`, `high`, `medium`, `low`),
- owner,
- deadline koreksi,
- dampak jika tidak diperbaiki.

Debt `critical` dan `high` wajib diselesaikan sebelum rilis mayor.

## Sync Verification Checklist

- [ ] perubahan kode memiliki update dokumen terdampak.
- [ ] metadata wajib lengkap di dokumen baru/diubah.
- [ ] route inventory/OpenAPI sinkron dengan kontrak runtime.
- [ ] tidak ada konten sensitif mentah pada contoh log/payload.
- [ ] action item documentation debt tercatat bila ada gap.
- [ ] evidence release confidence terbaru tersedia untuk candidate rilis.

## Related Documents

- [Documentation Maintenance Standard](/Users/macbookpro/Development/recova-backend-v2/docs/standards/documentation-maintenance.md)
- [API Docs Generation](/Users/macbookpro/Development/recova-backend-v2/docs/operations/api-docs-generation.md)
- [Documentation Quality Audit](/Users/macbookpro/Development/recova-backend-v2/docs/roadmap/documentation-quality-audit.md)

## Source Reference

- [tasks/lessons.md](/Users/macbookpro/Development/recova-backend-v2/tasks/lessons.md)
- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
