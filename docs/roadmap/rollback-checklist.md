---
title: Recova Backend Rollback Checklist
description: Checklist rollback operasional untuk mengembalikan stabilitas layanan saat cutover atau rilis migrasi tidak memenuhi gate kualitas.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/roadmap/rollback-checklist.md
last_reviewed: 2026-05-08
---

# Recova Backend Rollback Checklist

Checklist ini dipakai saat keputusan rollback diambil.

## A. Rollback Trigger Validation

- [ ] mismatch kontrak kritis terkonfirmasi.
- [ ] error rate/latency melewati threshold insiden.
- [ ] ada risiko integritas data atau auth regression.
- [ ] keputusan rollback disetujui migration owner + operations owner.

## B. Immediate Containment

- [ ] hentikan cutover tambahan.
- [ ] arahkan trafik domain terdampak ke runtime stabil sebelumnya.
- [ ] aktifkan incident bridge dan komunikasi status.
- [ ] freeze perubahan baru sampai stabil.

## C. Application Rollback

- [ ] artifact versi stabil tersedia.
- [ ] redeploy/re-route ke artifact stabil selesai.
- [ ] health checks liveness/readiness lulus.
- [ ] smoke tests endpoint domain target lulus.

## D. Database Safety Check

- [ ] status migration terakhir dipastikan.
- [ ] skema kompatibel dengan runtime yang dipulihkan.
- [ ] bila migration destruktif sudah jalan, gunakan forward-fix plan.
- [ ] restore backup hanya jika disetujui owner database.

## E. Stabilization Verification

- [ ] error rate kembali ke baseline.
- [ ] status code distribution kembali normal.
- [ ] auth dan ownership checks normal.
- [ ] observability signal stabil pada window observasi pasca rollback.

## F. Evidence and Follow-up

- [ ] root cause sementara terdokumentasi.
- [ ] timeline rollback dicatat.
- [ ] evidence rollback rehearsal terbaru tersedia (`artifacts/rollback-rehearsal/**`) dan berstatus pass.
- [ ] daftar tindakan perbaikan permanen disepakati.
- [ ] gate tambahan sebelum cutover ulang ditentukan.

## Related Documents

- [Rollback Runbook](/Users/macbookpro/Development/recova-backend-v2/docs/operations/rollback.md)
- [Migration Execution Runbook](/Users/macbookpro/Development/recova-backend-v2/docs/roadmap/migration-execution-runbook.md)
- [Cutover Checklist](/Users/macbookpro/Development/recova-backend-v2/docs/roadmap/cutover-checklist.md)

## Source Reference

- [Rollback Runbook](/Users/macbookpro/Development/recova-backend-v2/docs/operations/rollback.md)
- [Migration Strategy](/Users/macbookpro/Development/recova-backend-v2/docs/roadmap/migration-strategy.md)
- [/Users/macbookpro/Development/bisakerja-api/docs/operations/deployment.md](/Users/macbookpro/Development/bisakerja-api/docs/operations/deployment.md)
