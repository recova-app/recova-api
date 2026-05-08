---
title: Recova Backend CI/CD Operations
description: Strategi pipeline CI/CD untuk validasi kualitas kode, verifikasi migrasi, pengujian kontrak, build image, dan kontrol deployment yang deterministik.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/operations/ci-cd.md
last_reviewed: 2026-05-08
---

# Recova Backend CI/CD Operations

Dokumen ini menetapkan desain pipeline CI/CD agar setiap perubahan tervalidasi sebelum deployment.

## Pipeline Objectives

- quality gate konsisten untuk setiap perubahan,
- mencegah deploy ketika migration atau test gagal,
- menghasilkan artefak container yang reproducible,
- memisahkan verifikasi pre-merge dan pre-deploy.

## Deterministic Stage Order

1. checkout + dependency restore,
2. format/lint/static analysis,
3. unit dan handler tests,
4. integration tests + migration verification,
5. contract compatibility tests,
6. docs/generator checks (jika diaktifkan),
7. build container image,
8. security/dependency scan,
9. deploy gate evaluation,
10. deployment + smoke checks.

## Gate Rules

- langkah berikutnya tidak boleh berjalan jika langkah sebelumnya gagal,
- migration failure wajib memblok deployment,
- security scan severity critical/high memblok deployment kecuali exception disetujui,
- contract drift tanpa approval memblok release.

## Required Controls

| Control               | Rule                                                 |
| --------------------- | ---------------------------------------------------- |
| Job dependency        | gunakan dependency antar job eksplisit               |
| Concurrency           | cegah deploy paralel ke target environment yang sama |
| Artifact immutability | deploy menggunakan image tag immutable               |
| Secret handling       | secret hanya dari pipeline secret store              |
| Rollback trigger      | rollback path tersedia jika smoke check gagal        |

## Post-Deploy Verification

Setelah deployment:

- cek `/health/live` dan `/health/ready`,
- jalankan smoke endpoint prioritas,
- pantau error/latency pada window observasi awal,
- dokumentasikan hasil release gate.

## Related Documents

- [Release Gates](/Users/macbookpro/Development/recova-backend-v2/docs/operations/release-gates.md)
- [Testing Strategy](/Users/macbookpro/Development/recova-backend-v2/docs/operations/testing.md)
- [Containerization Strategy](/Users/macbookpro/Development/recova-backend-v2/docs/operations/containerization.md)
- [ADR 0008 CI/CD Strategy](/Users/macbookpro/Development/recova-backend-v2/docs/decisions/adr-0008-ci-cd-strategy.md)

## Source Reference

- [GitHub Actions Workflow Syntax](https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax)
- [GitHub Actions Concurrency](https://docs.github.com/en/actions/concepts/workflows-and-actions/concurrency)
