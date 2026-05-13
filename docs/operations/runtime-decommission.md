---
title: Recova Backend Runtime Decommission
description: Runbook decommission runtime legacy setelah runtime aktif stabil, termasuk gate traffic, retention evidence rollback, dan arsip konfigurasi legacy.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/operations/runtime-decommission.md
last_reviewed: 2026-05-08
---

# Recova Backend Runtime Decommission

Dokumen ini mendefinisikan langkah decommission runtime legacy secara aman dan audit-able.

## Decommission Goals

- memastikan runtime legacy keluar dari jalur request publik,
- memastikan evidence rollback rehearsal masih tersedia sesuai retention,
- mengarsipkan artefak konfigurasi runtime legacy untuk kebutuhan audit.

## Preconditions

- stabilization gate terakhir berstatus `passed`,
- rollback rehearsal terakhir berstatus `passed`,
- source of truth runtime sudah menunjuk runtime aktif (Go Fiber),
- keputusan decommission disetujui owner operasi.

## Decommission Gate Command

Gunakan command:

- `make runtime-decommission`

Input minimum:

- `LEGACY_RUNTIME_TRAFFIC_COUNT=0`,
- `ROLLBACK_EVIDENCE_DIR` menunjuk artefak rollback rehearsal,
- `RUNTIME_SOURCE_OF_TRUTH_FILE` menunjuk dokumen runtime aktif.

Input opsional:

- `LEGACY_RUNTIME_TRAFFIC_EVIDENCE_FILE` untuk menyalin bukti log traffic,
- `LEGACY_ARCHIVE_PATHS` untuk daftar path konfigurasi legacy yang diarsipkan,
- `ROLLBACK_RETENTION_DAYS` untuk window validasi evidence rollback.

## Validation Rules

Runner decommission wajib memverifikasi:

- traffic publik runtime legacy = `0`,
- minimal satu rollback report `passed` ada dalam window retention,
- dokumen source-of-truth runtime mengandung keyword runtime aktif,
- arsip legacy runtime berhasil dibuat.

## Output Evidence

Output default:

- `artifacts/decommission/*-summary.log`
- `artifacts/decommission/*-decommission-report.json`
- `artifacts/decommission/*-express-archive.tar.gz`

## Failure Rules

Decommission otomatis gagal jika:

- traffic runtime legacy tidak nol,
- evidence rollback rehearsal tidak tersedia atau stale,
- source of truth runtime tidak sinkron,
- arsip legacy tidak dapat dibuat.

## Related Documents

- [Deployment Workflow](/Users/macbookpro/Development/recova-backend-v2/docs/operations/deployment.md)
- [Rollback Runbook](/Users/macbookpro/Development/recova-backend-v2/docs/operations/rollback.md)
- [Current Runtime Inventory](/Users/macbookpro/Development/recova-backend-v2/docs/roadmap/current-runtime-inventory.md)

## Source Reference

- [Workflow syntax for GitHub Actions](https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax)
- [Store and share data with workflow artifacts](https://docs.github.com/en/actions/tutorials/store-and-share-data)
- [docker compose down](https://docs.docker.com/reference/cli/docker/compose/down/)
