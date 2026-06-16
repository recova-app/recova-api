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
last_reviewed: 2026-05-09
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
- critical flow E2E suite (`make test-e2e`),
- performance/load smoke suite (`make test-performance`),
- report release confidence tersimpan pada artefak pipeline,
- compose runtime smoke (`api` + `db`) lulus,
- security checklist terisi dan disetujui.

## Gate C - Pre-Deploy

Wajib lulus:

- build image sukses,
- vulnerability scan sesuai policy,
- deployment plan dan rollback path tersedia,
- gate environment staging melewati approval/protection yang ditetapkan,
- dry-run deploy staging lulus (`scripts/staging-deploy.sh`) termasuk migration dry-run, seed idempotency, integrity checks, dan readiness checks,
- remote deploy staging (`.github/workflows/deploy-staging.yml`) sukses dari branch `develop` dengan image immutable `sha-<commit-sha>`,
- Dokploy production compose valid dengan `IMAGE_TAG=sha-<commit-sha>` dan tanpa public `ports:` API,
- GitHub Environment `production` punya required reviewer sebelum `.github/workflows/deploy-production.yml` boleh redeploy,
- migration safety gate lulus; migration non-destructive boleh lanjut tanpa backup evidence dengan warning, sedangkan migration destructive wajib approval eksplisit,
- last-good image SHA dan rollback command/path tercatat,
- release sign-off engineering + platform.

## Gate D - Post-Deploy

Wajib lulus:

- liveness/readiness checks,
- smoke tests endpoint kritikal,
- observability signals stabil pada window awal,
- tidak ada lonjakan error severity tinggi.

Untuk window stabilisasi pasca-cutover domain besar:

- jalankan `make stabilization-gate`,
- simpan evidence `artifacts/stabilization/**`,
- pastikan rollback rehearsal terbaru berstatus `passed`.

Setelah stabilisasi pass, jalankan gate decommission runtime legacy:

- `make runtime-decommission`,
- simpan evidence `artifacts/decommission/**`.

Untuk maintenance berkelanjutan pasca decommission:

- jalankan `make post-migration-maintenance`,
- simpan evidence `artifacts/maintenance/**`,
- pastikan backlog maintenance berisi owner + priority.

## Blocking Conditions

Release harus dibatalkan atau ditahan jika:

- migration gagal,
- smoke test kritikal gagal,
- E2E critical flow tidak lulus,
- performance smoke melewati threshold,
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
- evidence rollback rehearsal tersedia pada `artifacts/rollback-rehearsal/**`.
- evidence runtime decommission tersedia pada `artifacts/decommission/**`.

## Related Documents

- [CI/CD Operations](/Users/macbookpro/Development/recova-backend-v2/docs/operations/ci-cd.md)
- [Security Checklist](/Users/macbookpro/Development/recova-backend-v2/docs/operations/security-checklist.md)
- [Verification Matrix](/Users/macbookpro/Development/recova-backend-v2/docs/operations/verification-matrix.md)

## Source Reference

- [GitHub Actions Workflow Syntax](https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax)
- [Store and Share Data with Workflow Artifacts](https://docs.github.com/en/actions/tutorials/store-and-share-data)
