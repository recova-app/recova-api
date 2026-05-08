---
title: Recova Backend Local Development Runtime
description: Kontrak runtime pengembangan lokal untuk backend meliputi mode native dan container compose, kebutuhan dependency, serta aturan validasi lingkungan.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/operations/local-development.md
last_reviewed: 2026-05-08
---

# Recova Backend Local Development Runtime

Dokumen ini mendefinisikan setup runtime lokal agar pengembangan dan debugging konsisten.

## Local Runtime Profiles

### Profile A - Native Runtime

Cocok untuk iterasi cepat.

Komponen minimum:

- service backend lokal,
- PostgreSQL lokal atau instance dev terisolasi,
- env file lokal khusus development.

### Profile B - Container Compose Runtime

Cocok untuk parity environment.

Komponen minimum:

- service backend via container,
- PostgreSQL via service terpisah di compose,
- volume data lokal terpisah,
- file runtime compose: `docker-compose.local.yml`.

## Environment Rules

- gunakan `APP_ENV=local`,
- secret lokal tetap dipisah dari secret staging/production,
- env required harus valid saat startup,
- tidak boleh ada fallback tersembunyi untuk env required.

## Database Rules for Local

- gunakan database terisolasi untuk development,
- jalankan migration sebelum menjalankan flow fitur,
- seed data hanya data sintetis,
- jangan gunakan database production/staging untuk local verification.

## Local Bootstrap Commands

Perintah standar pengembangan lokal:

| Command                 | Purpose                                                           |
| ----------------------- | ----------------------------------------------------------------- |
| `make preflight`        | validasi dependency tooling dan baseline struktur project         |
| `make fmt`              | format seluruh file Go dengan `gofmt`                             |
| `make lint`             | static analysis baseline via `go vet ./...`                       |
| `make test`             | menjalankan unit test package Go                                  |
| `make test-integration` | menjalankan scripted integration checks untuk tooling workflow    |
| `make build`            | build binary API ke `./bin/recova-api`                            |
| `make run`              | menjalankan API lokal dari `cmd/api`                              |
| `make migrate-up`       | apply migration menggunakan wrapper script                        |
| `make migrate-down`     | rollback migration (default 1 langkah) menggunakan wrapper script |
| `make migrate-status`   | menampilkan versi migration saat ini                              |
| `make migrate-check`    | validasi state migration tidak dirty                              |
| `make seed`             | menjalankan seed reference data minimal                           |
| `make compose-smoke`    | smoke test compose lokal (`api` + `db`) dengan cleanup otomatis   |

Catatan build:

- layout repository menggunakan direktori `api/` untuk kontrak OpenAPI;
- karena itu build artifact diarahkan eksplisit ke `./bin/recova-api` agar tidak bentrok dengan direktori `api/`.

Catatan env local:

- target `make run`, `make migrate-*`, dan `make seed` auto-load env dari `.env` melalui `scripts/with-env.sh`,
- jika ingin memakai file env lain, gunakan `ENV_FILE=<path> make <target>`.
- compose smoke default memakai `ENV_FILE=.env.example`; override bisa lewat `ENV_FILE=<path> make compose-smoke`.

## Local Verification Baseline

Setiap perubahan utama lokal harus memverifikasi:

- startup sukses,
- `GET /health/live` dan `GET /health/ready` sukses,
- endpoint yang diubah merespons sesuai kontrak,
- tidak ada log error kritikal saat skenario uji dasar.

## Common Troubleshooting

| Symptom          | First check                                          |
| ---------------- | ---------------------------------------------------- |
| startup gagal    | validasi env required dan format nilainya            |
| readiness gagal  | koneksi PostgreSQL, migration status, timeout config |
| auth route gagal | JWT/config cookie dan CORS origin policy             |
| AI route lambat  | timeout konfigurasi dan status provider              |

## Related Documents

- [Environment Configuration](/Users/macbookpro/Development/recova-backend-v2/docs/environment.md)
- [Database Migrations](/Users/macbookpro/Development/recova-backend-v2/docs/operations/database-migrations.md)
- [Containerization Strategy](/Users/macbookpro/Development/recova-backend-v2/docs/operations/containerization.md)

## Source Reference

- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
- [PostgreSQL Current Documentation](https://www.postgresql.org/docs/current/)
- [Docker Multi-stage Builds](https://docs.docker.com/build/building/multi-stage/)
- [Go Toolchains](https://go.dev/doc/toolchain)
