---
title: Recova Backend Verification Matrix
description: Matriks verifikasi minimum per modul dan tipe pengujian untuk memastikan readiness release dan kompatibilitas kontrak layanan.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/operations/verification-matrix.md
last_reviewed: 2026-05-08
---

# Recova Backend Verification Matrix

Matriks ini mendefinisikan tipe uji minimum per modul.

## Module to Test-Type Matrix

| Module                                 | Unit     | Handler  | Repo Integration | Contract | Smoke    | E2E      |
| -------------------------------------- | -------- | -------- | ---------------- | -------- | -------- | -------- |
| Auth                                   | required | required | required         | required | required | required |
| Users & Onboarding                     | required | required | required         | required | optional | required |
| Routine, Check-ins, Streak, Statistics | required | required | required         | required | optional | required |
| Journals                               | required | required | required         | required | optional | required |
| Community                              | required | required | required         | required | optional | required |
| Education                              | required | required | optional         | required | optional | optional |
| Daily Content                          | required | required | optional         | required | optional | optional |
| AI Coach                               | required | required | optional         | required | optional | required |
| Health                                 | optional | required | optional         | optional | required | optional |

## Gate Rules

- `required` berarti harus lulus sebelum merge release branch,
- `optional` berarti disarankan dan ditingkatkan secara bertingkat,
- kegagalan test `required` memblok release gate.

## Compatibility Focus Area

Contract tests wajib memverifikasi:

- path dan method endpoint,
- response envelope,
- status code dan error code,
- aturan auth guard,
- request id pada error envelope,
- batas payload dan validasi dasar.

Gate tambahan lintas module:

- `make module-consistency-check` untuk anatomy module, companion tests, dan boundary layer.

## Migration Readiness Checks

- migration up/down tervalidasi,
- no dirty migration state,
- query path kritikal tidak regress.

## Evidence Requirement

Untuk setiap release candidate, simpan:

- hasil suite test,
- ringkasan cakupan modul,
- daftar pengecualian test optional,
- keputusan sign-off engineering.

## Related Documents

- [Testing Strategy](/Users/macbookpro/Development/recova-backend-v2/docs/operations/testing.md)
- [Release Gates](/Users/macbookpro/Development/recova-backend-v2/docs/operations/release-gates.md)
- [Compatibility Test Plan](/Users/macbookpro/Development/recova-backend-v2/docs/roadmap/compatibility-test-plan.md)

## Source Reference

- [Go Testing Package](https://go.dev/pkg/testing/?m=old)
- [PostgreSQL Current Documentation](https://www.postgresql.org/docs/current/)
