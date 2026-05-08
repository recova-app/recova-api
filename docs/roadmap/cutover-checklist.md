---
title: Recova Backend Cutover Checklist
description: Checklist keputusan cutover untuk memindahkan trafik domain API secara aman dengan verifikasi kontrak, data, dan operasi.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/roadmap/cutover-checklist.md
last_reviewed: 2026-05-08
---

# Recova Backend Cutover Checklist

Checklist ini dipakai sebelum, saat, dan sesudah perpindahan trafik domain endpoint.

## A. Pre-Cutover Gate

- [ ] contract parity test untuk domain target lulus.
- [ ] authn/authz parity test lulus.
- [ ] migration yang relevan sudah diterapkan dan tervalidasi.
- [ ] observability wajib aktif (`requestId`, structured logs, health/readiness).
- [ ] rollback route switch tervalidasi di environment non-production.
- [ ] owner on-call dan channel incident aktif.

## B. Change Freeze Confirmation

- [ ] tidak ada perubahan breaking pada kontrak domain target.
- [ ] tidak ada migration destruktif yang belum punya forward-fix.
- [ ] backlog perubahan non-kritis ditunda sampai cutover selesai.

## C. Cutover Execution

- [ ] aktifkan perpindahan trafik bertahap untuk domain target.
- [ ] pantau status code distribution pada endpoint inti.
- [ ] pantau error envelope mismatch antar runtime.
- [ ] pantau latency P95/P99 terhadap baseline sebelum cutover.
- [ ] dokumentasikan timestamp mulai cutover.

## D. Post-Cutover Verification

- [ ] smoke checks domain target lulus.
- [ ] tidak ada lonjakan error kritis di window observasi awal.
- [ ] tidak ada mismatch data ownership atau auth scope.
- [ ] dashboard operasional menunjukkan stabilitas.
- [ ] keputusan final `go-forward` atau `rollback` dicatat.

## E. Evidence Capture

- [ ] commit/artifact identifier runtime aktif dicatat.
- [ ] hasil test parity dilampirkan.
- [ ] ringkasan metrik pasca cutover dicatat.
- [ ] keputusan owner tercatat bersama alasan teknis.

## Escalation Rule

Lakukan rollback segera bila salah satu kondisi terjadi:

- mismatch kontrak kritis pada endpoint publik,
- error rate melampaui threshold insiden,
- regresi keamanan/authn/authz,
- migration side effect menyebabkan risiko integritas data.

## Related Documents

- [Migration Execution Runbook](/Users/macbookpro/Development/recova-backend-v2/docs/roadmap/migration-execution-runbook.md)
- [Rollback Checklist](/Users/macbookpro/Development/recova-backend-v2/docs/roadmap/rollback-checklist.md)
- [Compatibility Test Plan](/Users/macbookpro/Development/recova-backend-v2/docs/roadmap/compatibility-test-plan.md)

## Source Reference

- [Compatibility Matrix](/Users/macbookpro/Development/recova-backend-v2/docs/roadmap/compatibility-matrix.md)
- [Deployment Workflow](/Users/macbookpro/Development/recova-backend-v2/docs/operations/deployment.md)
- [/Users/macbookpro/Development/bisakerja-api/docs/operations/deployment.md](/Users/macbookpro/Development/bisakerja-api/docs/operations/deployment.md)
