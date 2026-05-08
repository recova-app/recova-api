---
title: Recova Backend Release Gates
description: Definisi gate release dari pre-merge hingga post-deploy untuk memastikan kesiapan teknis, keamanan, dan operasional sebelum trafik dialihkan.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/operations/release-gates.md
last_reviewed: 2026-05-08
---

# Recova Backend Release Gates

Dokumen ini mendefinisikan aturan lulus/gagal release.

## Gate A - Pre-Merge

Wajib lulus:

- lint/static analysis,
- unit tests,
- handler tests,
- dokumen yang berubah memiliki metadata valid.

## Gate B - Pre-Release Candidate

Wajib lulus:

- integration tests database,
- migration up/down verification,
- contract compatibility tests,
- security checklist terisi dan disetujui.

## Gate C - Pre-Deploy

Wajib lulus:

- build image sukses,
- vulnerability scan sesuai policy,
- deployment plan dan rollback path tersedia,
- release sign-off engineering + platform.

## Gate D - Post-Deploy

Wajib lulus:

- liveness/readiness checks,
- smoke tests endpoint kritikal,
- observability signals stabil pada window awal,
- tidak ada lonjakan error severity tinggi.

## Blocking Conditions

Release harus dibatalkan atau ditahan jika:

- migration gagal,
- smoke test kritikal gagal,
- readiness gagal stabil,
- vulnerability critical/high tanpa approval,
- contract drift tidak disetujui.

## Rollback Gate

Rollback dilakukan bila:

- Gate D gagal,
- dampak user signifikan,
- mitigasi cepat tidak menstabilkan layanan.

Rollback harus diikuti post-rollback verification:

- health checks normal,
- endpoint kritikal normal,
- error rate kembali ke baseline.

## Related Documents

- [CI/CD Operations](/Users/macbookpro/Development/recova-backend-v2/docs/operations/ci-cd.md)
- [Security Checklist](/Users/macbookpro/Development/recova-backend-v2/docs/operations/security-checklist.md)
- [Verification Matrix](/Users/macbookpro/Development/recova-backend-v2/docs/operations/verification-matrix.md)

## Source Reference

- [GitHub Actions Workflow Syntax](https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax)
