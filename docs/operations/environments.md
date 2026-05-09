---
title: Environment Matrix and Runtime Profiles
description: Matrix konfigurasi local, test, staging, dan production untuk operasi layanan Recova Backend.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/operations/environments.md
last_reviewed: 2026-05-09
---

# Environment Matrix and Runtime Profiles

Dokumen ini menjelaskan tujuan tiap environment dan baseline konfigurasi runtime yang harus dipenuhi.

## Environment Purpose

| Environment  | Purpose                          | Data rule                                               |
| ------------ | -------------------------------- | ------------------------------------------------------- |
| `local`      | pengembangan harian              | data lokal/disposable saja                              |
| `test`       | automated test dan contract test | database terisolasi khusus test                         |
| `staging`    | validasi pre-production          | konfigurasi mirip production dengan secret non-produksi |
| `production` | runtime pengguna akhir           | secret terkelola, kontrol keamanan paling ketat         |

## Configuration Baseline Matrix

| Configuration area  | local                         | test                      | staging                         | production                        |
| ------------------- | ----------------------------- | ------------------------- | ------------------------------- | --------------------------------- |
| `APP_ENV`           | `local`                       | `test`                    | `staging`                       | `production`                      |
| `NODE_ENV`          | `development`                 | `test`                    | `production`                    | `production`                      |
| DB target           | local container/instance      | isolated test DB          | managed DB non-prod             | managed DB prod                   |
| cookie secure       | `false`                       | `false`                   | `true`                          | `true`                            |
| CORS origins        | localhost allowlist           | test-runner origin only   | staging frontend allowlist      | production frontend allowlist     |
| AI provider         | provider utama atau mock aman | deterministic test mode   | provider nyata dengan guardrail | provider nyata + fallback policy  |
| observability sinks | console + local collector     | test logs + artifact logs | centralized non-prod monitoring | centralized production monitoring |

## Runtime Rules per Environment

### Local

- gunakan placeholder secret non-empty untuk menjaga validasi tetap ketat,
- boleh memakai data seed lokal,
- endpoint development-only boleh aktif dengan guard eksplisit.

### Test

- semua test harus memakai database terisolasi,
- test tidak boleh mengakses secret production,
- integrasi downstream AI boleh menggunakan mock atau sandbox resmi.

### Staging

- harus mereplikasi jalur deployment production,
- semua secret berasal dari secret manager non-prod,
- semua gate observability, rate limit, dan error mapping harus aktif.
- deployment otomatis hanya dari branch `develop` dengan GitHub Environment `staging`.
- artifact runtime staging harus memakai image immutable `sha-<commit-sha>` walau branch tag `develop` tetap dipublish.

### Production

- tidak boleh memakai mock provider,
- semua secret wajib dikelola terpusat dan dapat dirotasi,
- logging harus mengikuti redaction policy,
- health dan readiness wajib dipantau kontinu.

## Deployment-Time Validation

Sebelum deploy, verifikasi:

- env required lengkap dan valid,
- tidak ada secret kosong,
- `APP_ENV` sesuai target deployment,
- database URL mengarah ke lingkungan yang benar,
- origin CORS sesuai domain frontend resmi.

## Incident Response Notes

Jika terjadi salah konfigurasi env:

- blok rollout,
- rollback ke release stabil,
- perbaiki secret/config source,
- dokumentasikan insiden konfigurasi pada postmortem operasional.

## Related Documents

- [Environment Configuration](/Users/macbookpro/Development/recova-backend-v2/docs/environment.md)
- [Configuration Validation Standard](/Users/macbookpro/Development/recova-backend-v2/docs/standards/config-validation.md)
- [Health and Readiness](/Users/macbookpro/Development/recova-backend-v2/docs/operations/health-readiness.md)

## Source Reference

- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
- [/Users/macbookpro/Development/bisakerja-api/docs/environment.md](/Users/macbookpro/Development/bisakerja-api/docs/environment.md)
