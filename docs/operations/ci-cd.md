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

Urutan ini diimplementasikan di `.github/workflows/ci.yml` melalui job:

1. `quality-gates`,
2. `database-gates`,
3. `container-compose-smoke`,
4. `image-build`,
5. `deploy-gate-staging` (hanya branch `main` atau manual `workflow_dispatch`).

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

## Workflow Scope dan Trigger

Workflow CI aktif pada:

- `pull_request` ke `main` dan `develop`,
- `push` ke `main` dan `develop`,
- `workflow_dispatch` untuk menjalankan gate staging manual.

Kontrol concurrency:

- `concurrency.group = ci-${workflow}-${ref}`,
- `cancel-in-progress = true` untuk membatalkan run lama pada ref yang sama.

## Gate Coverage Implementasi

`quality-gates`:

- `make preflight`,
- gofmt check (`gofmt -l .` wajib kosong),
- `make lint`,
- `make test`,
- `make openapi-check`,
- `make test-integration`,
- `make security-scan`,
- `make build`.

`database-gates`:

- PostgreSQL service container,
- `make migrate-up`,
- `make migrate-check`,
- rollback smoke (`make migrate-down` lalu `make migrate-up`),
- `go test ./...` dengan `RECOVA_DB_INTEGRATION_URL`.

`container-compose-smoke`:

- menjalankan `scripts/compose-smoke.sh` berbasis `docker-compose.local.yml`.

`image-build`:

- `docker build` dengan metadata `VERSION`, `VCS_REF`, dan `BUILD_DATE`.

`deploy-gate-staging`:

- mensyaratkan semua job sebelumnya sukses,
- dieksekusi pada environment `staging` agar approval/protection rule bisa diterapkan dari GitHub Environment.

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
- [GitHub Actions Concurrency](https://docs.github.com/en/actions/how-tos/write-workflows/choose-when-workflows-run/control-workflow-concurrency)
- [GitHub Actions PostgreSQL Service Containers](https://docs.github.com/en/actions/tutorials/use-containerized-services/create-postgresql-service-containers)
- [actions/setup-go](https://github.com/actions/setup-go)
