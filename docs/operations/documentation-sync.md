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
last_reviewed: 2026-05-09
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

| Jenis perubahan                       | Dokumen minimum yang harus diperbarui                                                                                                                   |
| ------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Endpoint/contract API                 | `docs/api-reference.md`, `docs/generated/routes.md`, `docs/generated/openapi.yaml`, dokumen modul terkait                                               |
| API docs presentation                 | `scalar.config.json`, `docs/api-reference.md`, `docs/operations/api-docs-generation.md`                                                                 |
| Module structure/contract consistency | `docs/standards/module-structure-consistency.md`, `docs/standards/module-contract-consistency.md`, `docs/roadmap/internal-modules-consistency-audit.md` |
| Module consistency cleanup backlog    | `docs/roadmap/module-consistency-cleanup-backlog.md`, dokumen standar/audit konsistensi terkait                                                         |
| Auth/security                         | `docs/operations/security.md`, `docs/modules/auth.md`, standar terkait                                                                                  |
| Database/schema                       | `docs/database.md`, `docs/operations/database-migrations.md`                                                                                            |
| Deployment/runtime                    | `docs/operations/deployment.md`, `docs/operations/rollback.md`, post-deploy checks                                                                      |
| Runtime decommission                  | `docs/operations/runtime-decommission.md`, `docs/roadmap/current-runtime-inventory.md`, legacy baseline docs                                            |
| Post-migration maintenance            | `docs/operations/post-migration-maintenance.md`, `docs/standards/documentation-maintenance.md`                                                          |
| Privacy/logging                       | `docs/operations/data-privacy.md`, `docs/operations/data-retention.md`, `docs/standards/log-redaction-policy.md`                                        |
| Release validation evidence           | `docs/generated/release-confidence-report.md` atau artefak report release validation yang ekuivalen                                                     |

## Pull Request Review Gate

PR review wajib memeriksa:

- apakah ada perubahan kontrak yang belum di-dokumentasikan,
- apakah metadata dokumen valid dan `last_reviewed` diperbarui,
- apakah link internal dokumen tetap valid,
- apakah contoh payload aman dari data sensitif.
- apakah `scalar.config.json` masih sinkron dengan halaman docs yang dipakai dan artefak OpenAPI generated.
- apakah perubahan file/config memiliki test companion atau exception rationale.
- apakah `make module-consistency-check` tetap lulus saat mengubah layer/module anatomy.
- apakah backlog cleanup konsistensi module tetap sinkron dengan status exception/gap aktual.
- apakah report release confidence (E2E + performance smoke) terlampir untuk release candidate.
- apakah evidence decommission/maintenance terbaru tersedia bila runtime legacy sudah ditutup.

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
- [ ] `scalar.config.json` sinkron dengan docs pages + artefak OpenAPI generated.
- [ ] `make module-consistency-check` lulus atau exception rationale terdokumentasi.
- [ ] `make module-consistency-full-check` lulus untuk cleanup behavior-preserving lintas module.
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
