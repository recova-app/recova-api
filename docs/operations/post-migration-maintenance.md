---
title: Recova Backend Post-Migration Maintenance
description: Runbook maintenance berkelanjutan pasca migrasi mencakup triage freshness docs, review threshold observability/SLO, cadence dependency update, dan backlog optimasi.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/operations/post-migration-maintenance.md
last_reviewed: 2026-05-08
---

# Recova Backend Post-Migration Maintenance

Dokumen ini mendefinisikan maintenance rutin untuk menjaga kualitas runtime aktif dan dokumentasi operasional.

## Maintenance Goals

- menjaga freshness dokumen sesuai cadence,
- memastikan review threshold alert dan target SLO tetap rutin,
- memastikan cadence update dependency tetap berjalan,
- menjaga backlog optimasi lanjutan memiliki owner dan prioritas.

## Maintenance Gate Command

Gunakan command:

- `make post-migration-maintenance`

Input minimum (wajib `done`):

- `ALERT_REVIEW_STATUS=done`
- `SLO_REVIEW_STATUS=done`
- `DEPENDENCY_CADENCE_REVIEW_STATUS=done`

Input opsional:

- `SECURITY_DOC_CADENCE_DAYS` (default: `30`),
- `GENERAL_DOC_CADENCE_DAYS` (default: `90`),
- `DEFAULT_BACKLOG_OWNER`,
- `DEFAULT_BACKLOG_DUE_DATE`.

## Freshness Triage Rule

Runner maintenance menilai stale docs menggunakan metadata `last_reviewed`:

- dokumen high-risk (security/auth/privacy/deployment): cadence maksimal `30` hari,
- dokumen kategori lain: cadence maksimal `90` hari.

## Required Output Evidence

Output default:

- `artifacts/maintenance/*-summary.log`
- `artifacts/maintenance/*-docs-freshness-report.tsv`
- `artifacts/maintenance/*-maintenance-backlog.md`
- `artifacts/maintenance/*-maintenance-report.json`

Backlog wajib memiliki:

- item action,
- owner,
- priority,
- due date.

## Schedule Recommendation

Jalankan gate maintenance minimal mingguan dan setiap ada perubahan kontrak besar.

## Related Documents

- [Documentation Maintenance Standard](/Users/macbookpro/Development/recova-backend-v2/docs/standards/documentation-maintenance.md)
- [Documentation Sync Operations](/Users/macbookpro/Development/recova-backend-v2/docs/operations/documentation-sync.md)
- [Observability](/Users/macbookpro/Development/recova-backend-v2/docs/operations/observability.md)

## Source Reference

- [Workflow syntax for GitHub Actions](https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax)
- [Store and share data with workflow artifacts](https://docs.github.com/en/actions/tutorials/store-and-share-data)
