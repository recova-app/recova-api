---
title: Recova Backend Legacy Express Baseline
description: Catatan historis runtime Express sebagai arsip referensi setelah runtime publik dipindahkan ke Go Fiber.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: deprecated
source_repo: recova-backend-v2
source_path: docs/roadmap/express-baseline.md
last_reviewed: 2026-05-08
---

# Recova Backend Legacy Express Baseline

Dokumen ini hanya menyimpan baseline historis runtime lama untuk kebutuhan audit dan perbandingan kontrak.

## Lifecycle Status

- Status dokumen: `deprecated`.
- Runtime publik aktif: Go Fiber.
- Baseline ini tidak boleh dipakai sebagai acuan deploy runtime aktif.

## Historical Snapshot

Ringkasan runtime lama (historis):

- Language: Node.js + TypeScript.
- HTTP framework: Express.
- ORM: Prisma.
- Prefix endpoint publik: `/api/v1`.

## Archived Sources

Artefak historis runtime lama disimpan pada:

- direktori `references/`,
- artifact hasil decommission pada `artifacts/decommission/**`.

## Usage Rules

- gunakan dokumen ini hanya untuk:
  - investigasi historis,
  - audit transisi runtime,
  - komparasi kontrak lama vs runtime aktif.
- gunakan [Current Runtime Inventory](/Users/macbookpro/Development/recova-backend-v2/docs/roadmap/current-runtime-inventory.md) untuk keputusan operasional saat ini.

## Related Documents

- [Current Runtime Inventory](/Users/macbookpro/Development/recova-backend-v2/docs/roadmap/current-runtime-inventory.md)
- [Migration Execution Runbook](/Users/macbookpro/Development/recova-backend-v2/docs/roadmap/migration-execution-runbook.md)
- [Runtime Decommission Runbook](/Users/macbookpro/Development/recova-backend-v2/docs/operations/runtime-decommission.md)

## Source Reference

- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
