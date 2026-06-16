---
title: Recova Backend Deployment Workflow
description: Runbook deployment layanan Recova Backend mencakup topology rollout, urutan migrasi database, injeksi konfigurasi, dan smoke verification.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/operations/deployment.md
last_reviewed: 2026-05-09
---

# Recova Backend Deployment Workflow

Dokumen ini mendefinisikan alur deploy layanan secara aman dan terverifikasi, dari persiapan artefak sampai trafik stabil.

## Deployment Goals

- artefak yang dirilis harus reproducible,
- perubahan skema database harus terkontrol,
- downtime harus minimal,
- rollback aplikasi harus cepat,
- dampak insiden harus terlokalisasi.

## Topology Options

| Topology   | Kapan dipakai                                           | Karakteristik                                                                      |
| ---------- | ------------------------------------------------------- | ---------------------------------------------------------------------------------- |
| Rolling    | default environment dengan resource terbatas            | kapasitas dipindah secara gradual ke versi baru, validasi dilakukan selama rollout |
| Blue/Green | rilis berisiko tinggi atau perubahan besar pada runtime | environment baru disiapkan penuh, trafik dipindah setelah smoke checks lulus       |

Aturan pemilihan:

- gunakan rolling untuk perubahan kecil/menengah,
- gunakan blue/green untuk perubahan berisiko tinggi atau saat rollback instan dibutuhkan.

## Deployment Preconditions

Sebelum deploy, wajib tersedia:

- image/container artifact yang immutable (tag commit atau digest),
- konfigurasi environment tervalidasi,
- migration scripts yang sudah lolos verifikasi,
- backup database terbaru untuk rilis dengan perubahan skema,
- release gates lulus sesuai [Release Gates](/Users/macbookpro/Development/recova-backend-v2/docs/operations/release-gates.md).

## Environment Injection Rules

- seluruh konfigurasi runtime disuplai dari secret manager atau environment runtime,
- secret tidak boleh di-hardcode di image atau repository,
- variabel wajib harus fail-fast saat startup,
- perbedaan konfigurasi antar-environment harus eksplisit dan terdokumentasi di [Environment](/Users/macbookpro/Development/recova-backend-v2/docs/environment.md).

## Standard Deployment Sequence

```text
1) Build and publish immutable artifact
2) Validate environment configuration
3) Run migration precheck
4) Apply forward migrations
5) Roll out application version
6) Execute post-deploy checks
7) Promote deployment status to healthy
```

## Compose-Based Staging Deployment Runner

Untuk staging berbasis Docker Compose, gunakan runner otomatis:

- `make staging-deploy` atau `./scripts/staging-deploy.sh`.

Runner ini mengeksekusi urutan berikut secara deterministik:

1. validasi konfigurasi compose (`docker compose config -q`),
2. bootstrap dependency database,
3. apply migration dan migration health check,
4. migration dry-run (`down 1 -> up`) pada database staging disposable,
5. seed reference data dua kali untuk verifikasi idempotency,
6. integrity checks data referensi,
7. startup service API dengan readiness gate,
8. smoke check health endpoint.

Jika salah satu langkah gagal, deployment dianggap gagal dan stack dibersihkan otomatis (kecuali `KEEP_STACK=true`).

## Remote Staging Deployment from `develop`

Untuk jalur staging yang mereplikasi deployment production-style:

- trigger deploy dari branch `develop` melalui workflow `.github/workflows/deploy-staging.yml`,
- build image ke GHCR dengan dua tag:
  - mutable branch tag: `develop`,
  - immutable release tag: `sha-<commit-sha>`,
- deploy ke host staging via `scripts/deploy/remote-deploy.sh`.

Remote runner mengeksekusi urutan:

1. sinkronisasi checkout remote ke `origin/develop` sebelum menulis env file agar script deploy terbaru selalu dipakai,
2. tulis env file staging dari GitHub secret,
3. validasi precondition host (`git`, `docker`, `curl`, repo clean),
4. validasi `APP_ENV=staging` dan `DATABASE_URL` dari env file staging,
5. update `APP_IMAGE` ke tag immutable `sha-<commit-sha>`,
6. pull image terbaru dan apply migration (`up` + `check`),
7. `docker compose up -d --wait` untuk service API,
8. smoke checks: liveness, readiness, OpenAPI route, unauthorized reject pada route protected,
9. diagnostics otomatis saat gagal (`compose ps`, log tail, migration status, health output).

Catatan database URL:

- env staging tetap memakai format canonical `postgresql://`,
- nilai env file boleh memakai quote pembungkus,
- runner deploy menghapus quote pembungkus sebelum mengekspor `DATABASE_URL`,
- wrapper migrasi dapat memakai `postgres://` hanya sebagai argumen internal `golang-migrate`.
- health/OpenAPI/protected-route smoke utama memakai loopback host (`127.0.0.1:<port>`) agar deploy tidak bergantung pada kemampuan VPS melakukan hairpin request ke domain publiknya sendiri.
- public health smoke bersifat opsional melalui `RUN_PUBLIC_SMOKE=true`; diagnostics public health hanya informasi dan tidak menggagalkan deploy.
- semua smoke dan diagnostics memakai bounded `curl` timeout agar target yang tidak reachable tidak menggantung lama.

Catatan: `docker-compose.local.yml` tidak digunakan sebagai target runtime deploy staging/production.

## Dokploy Production Deployment

Jalur production target memakai Dokploy + GHCR dan dipisahkan dari jalur staging SSH Compose existing.

Alur production:

1. push ke `main`, tag `v*.*.*`, atau `workflow_dispatch` pada `.github/workflows/deploy-production.yml`,
2. GitHub Actions build image dan push ke `ghcr.io/recova-app/backend-v2`,
3. image selalu punya tag immutable `sha-<commit-sha>` dan pointer tag `main` atau `v*.*.*`,
4. GitHub Environment `production` memberi approval sebelum deploy,
5. Dokploy menarik image memakai manifest `docker-compose.dokploy.yml`,
6. smoke checks publik memverifikasi `/health/live`, `/health/ready`, `/openapi.yaml`, dan route protected menolak request tanpa auth.

Kontrak Dokploy:

- compose production: `docker-compose.dokploy.yml`,
- image: `ghcr.io/recova-app/backend-v2:${IMAGE_TAG}`,
- default deploy/rollback anchor: `IMAGE_TAG=sha-<commit-sha>`,
- env runtime disuplai melalui Dokploy Environment dan dibaca container lewat `env_file: .env`,
- API hanya memakai `expose: "3001"`; domain publik dikonfigurasi melalui Dokploy Domains UI,
- healthcheck container memakai `wget` ke `/health/live`, sesuai binary yang tersedia di image.

Secret/variable contract:

| Name                             | Location               | Purpose                                                                                    |
| -------------------------------- | ---------------------- | ------------------------------------------------------------------------------------------ |
| `DOKPLOY_WEBHOOK_URL`            | GitHub Secret          | opsi trigger redeploy via webhook                                                          |
| `DOKPLOY_URL`                    | GitHub Variable/Secret | base URL panel Dokploy untuk API deploy                                                    |
| `DOKPLOY_API_TOKEN`              | GitHub Secret          | token API Dokploy                                                                          |
| `DOKPLOY_APPLICATION_ID`         | GitHub Variable/Secret | target application ID bila memakai API deploy                                              |
| `CF_ACCESS_CLIENT_ID`            | GitHub Secret          | opsional; Cloudflare Access service token client ID untuk panel Dokploy protected          |
| `CF_ACCESS_CLIENT_SECRET`        | GitHub Secret          | opsional; Cloudflare Access service token client secret untuk panel Dokploy protected      |
| `PRODUCTION_DOMAIN`              | GitHub Variable        | domain publik API tanpa skema                                                              |
| `BACKUP_EVIDENCE_URL`            | GitHub Variable        | opsional untuk migration non-destructive; direkomendasikan sebagai bukti backup production |
| `APPROVE_DESTRUCTIVE_MIGRATIONS` | GitHub Variable        | wajib `true` hanya jika gate mendeteksi migration destructive                              |
| `IMAGE_TAG`                      | Dokploy Environment    | tag image immutable yang akan dijalankan                                                   |

### Dokploy Git Source Setup Langsung

Gunakan source **Git** ketika GitHub App Dokploy bermasalah atau repository cukup ditarik langsung dari Git URL. Jalur ini tidak bergantung pada GitHub provider/account integration Dokploy.

Public repository:

1. buka Dokploy Application/Compose service production,
2. pilih provider **Git**,
3. isi repository URL HTTPS:

```text
https://github.com/<owner>/<repo>.git
```

4. isi branch production, default `main`,
5. simpan dan jalankan deploy manual pertama dari Dokploy UI.

Private repository:

1. buka Dokploy **SSH Keys**,
2. buat SSH key baru dan gunakan **Generate RSA SSH Key**,
3. copy public key,
4. buka GitHub repo → **Settings** → **Deploy keys** → **Add deploy key**,
5. paste public key Dokploy,
6. jangan aktifkan **Allow write access** karena Dokploy hanya perlu pull,
7. kembali ke Dokploy Application/Compose service,
8. pilih provider **Git**,
9. isi repository URL SSH:

```text
git@github.com:<owner>/<repo>.git
```

10. isi branch production, default `main`,
11. simpan dan jalankan deploy manual pertama dari Dokploy UI.

Watch Paths:

- untuk repository single-app, isi `**` agar semua perubahan bisa memicu deploy,
- untuk menghindari deploy karena dokumentasi/CI, gunakan pola berikut:

```text
**
!README.md
!docs/**
!.github/**
```

- untuk monorepo, batasi ke folder service dan file build terkait, contoh:

```text
apps/api/**
package.json
package-lock.json
Dockerfile
docker-compose.dokploy.yml
```

Auto deploy dengan provider **Git**:

1. aktifkan auto-deploy/webhook pada service Dokploy jika tersedia,
2. copy webhook URL dari Dokploy,
3. buka GitHub repo → **Settings** → **Webhooks** → **Add webhook**,
4. isi **Payload URL** dengan webhook URL Dokploy,
5. pilih content type `application/json`,
6. pilih event **Just the push event**,
7. simpan webhook,
8. push commit kecil ke branch production dan pastikan Dokploy menerima trigger.

Jika webhook tidak dipakai, deploy tetap bisa dijalankan manual dari Dokploy UI atau melalui workflow manual GitHub `deploy-production` setelah operator memastikan `IMAGE_TAG` Dokploy sudah sesuai.

### First-Time Dokploy Setup dengan Git dan Compose

Deploy manual pertama wajib dilakukan untuk memvalidasi clone repository, compose path, environment, GHCR pull, domain, dan healthcheck sebelum webhook atau workflow manual dijadikan jalur operasional.

#### 1. Pastikan image GHCR tersedia

Jalankan workflow production atau push ke `main` sampai image berikut tersedia:

```text
ghcr.io/recova-app/backend-v2:sha-<commit-sha>
```

Catat tag immutable yang akan dipromote:

```text
IMAGE_TAG=sha-<commit-sha>
```

#### 2. Buat token GHCR untuk pull image

Buat GitHub Personal Access Token dengan scope minimal:

```text
read:packages
```

Jika package private berada di organization, pastikan token memiliki akses ke package `ghcr.io/recova-app/backend-v2`.

#### 3. Tambahkan registry GHCR di Dokploy

Dokploy → **Registry** → add registry:

```text
Name: ghcr
Registry URL: ghcr.io
Username: <github-username>
Password/Token: <PAT read:packages>
```

Simpan dan gunakan test pull bila tersedia.

#### 4. Setup Git SSH key Dokploy

1. Dokploy → **SSH Keys**,
2. buat SSH key baru,
3. generate RSA SSH key,
4. copy public key,
5. GitHub repo → **Settings** → **Deploy keys** → **Add deploy key**,
6. isi title `dokploy-production`,
7. paste public key Dokploy,
8. pastikan **Allow write access** tidak aktif.

#### 5. Buat Docker Compose service di Dokploy

Buat service baru sebagai **Docker Compose** dengan source **Git**.

Isi konfigurasi source:

```text
Repository URL: git@github.com:<owner>/<repo>.git
Branch: main
Docker Compose Path: docker-compose.dokploy.yml
Watch Paths: **
```

Jika ingin menghindari redeploy karena perubahan docs atau CI, gunakan Watch Paths:

```text
**
!README.md
!docs/**
!.github/**
```

Jangan memakai build type **Dockerfile** untuk jalur ini. File `docker-compose.dokploy.yml` sudah menarik image dari GHCR melalui `IMAGE_TAG`.

#### 6. Isi Environment Dokploy

Minimal environment production:

```text
IMAGE_TAG=sha-<commit-sha>
APP_ENV=production
NODE_ENV=production
PORT=3001
```

Tambahkan seluruh environment aplikasi yang wajib sesuai `docs/environment.md`, misalnya koneksi database, JWT secret, provider AI, dan konfigurasi integrasi lain. Secret harus disimpan di Dokploy Environment, bukan di repository.

#### 7. Setup domain

Dokploy service → **Domains**:

```text
Domain: api.example.com
Service: api
Port: 3001
HTTPS: enabled
```

Compose production memakai `expose: "3001"`; jangan menambahkan `ports:` host untuk API, database, atau Redis.

#### 8. Jalankan deploy manual pertama

Klik **Deploy** dari Dokploy UI dan cek log berikut:

- Git clone berhasil,
- `docker-compose.dokploy.yml` ditemukan,
- image GHCR berhasil di-pull,
- container `api` running,
- healthcheck menjadi healthy,
- domain mengarah ke service `api` port `3001`.

Jika image pull gagal, cek registry GHCR, token `read:packages`, package visibility, dan nilai `IMAGE_TAG`.

#### 9. Jalankan smoke test publik

```bash
curl https://api.example.com/health/live
curl https://api.example.com/health/ready
curl -I https://api.example.com/openapi.yaml
curl -i https://api.example.com/api/v1/users/me
```

Route protected tanpa auth harus mengembalikan `401` atau `403`.

#### 10. Pilih jalur operasional setelah deploy pertama

Opsi manual Dokploy:

1. GitHub Actions build image,
2. catat `sha-<commit-sha>`,
3. update `IMAGE_TAG` di Dokploy,
4. klik **Deploy**,
5. jalankan smoke test.

Opsi workflow manual GitHub:

1. setup GitHub Environment `production`,
2. aktifkan **Required reviewers**,
3. isi `DOKPLOY_WEBHOOK_URL` atau pasangan API `DOKPLOY_URL`, `DOKPLOY_API_TOKEN`, `DOKPLOY_APPLICATION_ID`,
4. jika panel Dokploy dilindungi Cloudflare Zero Trust Access, isi `CF_ACCESS_CLIENT_ID` dan `CF_ACCESS_CLIENT_SECRET`,
5. isi `PRODUCTION_DOMAIN=api.example.com`,
6. isi `APPROVE_DESTRUCTIVE_MIGRATIONS=false`,
7. buka GitHub → **Actions** → **Deploy Production Dokploy**,
8. jalankan **Run workflow** dengan `trigger_dokploy=true` hanya setelah `IMAGE_TAG` Dokploy sesuai tag target,
9. approve environment `production`,
10. tunggu smoke test workflow selesai.

Checklist setup pertama:

```text
[ ] GHCR image tersedia
[ ] Dokploy Registry ghcr.io bisa pull image
[ ] Git deploy key bisa clone repository
[ ] Source type Docker Compose
[ ] Source provider Git
[ ] Docker Compose Path docker-compose.dokploy.yml
[ ] IMAGE_TAG berisi sha-<commit-sha>
[ ] env production lengkap
[ ] domain api -> service api port 3001
[ ] deploy manual pertama healthy
[ ] smoke test publik lulus
[ ] jalur deploy berikutnya dipilih: manual Dokploy atau workflow GitHub
```

### GitHub Environment Setup dengan Dokploy Webhook

Gunakan webhook saat Dokploy sudah menyediakan deploy URL seperti:

```text
https://panel.example.com/api/deploy/<redacted-token>
```

Setup GitHub:

1. buka GitHub repo → **Settings** → **Environments** → buat atau pilih `production`,
2. aktifkan **Required reviewers** untuk gate approval production,
3. tambah **Environment Secret**:

```text
DOKPLOY_WEBHOOK_URL=https://panel.example.com/api/deploy/<redacted-token>
```

4. jangan simpan webhook sebagai variable karena token URL bersifat secret,
5. jika memakai webhook, tidak perlu mengisi `DOKPLOY_URL`, `DOKPLOY_API_TOKEN`, atau `DOKPLOY_APPLICATION_ID`,
6. jika webhook berada di balik Cloudflare Zero Trust Access, tambahkan **Environment Secret** `CF_ACCESS_CLIENT_ID` dan `CF_ACCESS_CLIENT_SECRET` dari Access service token,
7. tambahkan **Repository Variables**:

```text
PRODUCTION_DOMAIN=api.example.com
APPROVE_DESTRUCTIVE_MIGRATIONS=false
```

8. `BACKUP_EVIDENCE_URL` opsional untuk migration non-destructive dan bisa diisi saat backup evidence tersedia.

Dengan webhook mode, GitHub Actions hanya memanggil webhook setelah image berhasil dipush dan approval production diberikan. Dokploy akan redeploy memakai `IMAGE_TAG` yang sedang terset di Environment Dokploy. Jika `CF_ACCESS_CLIENT_ID` dan `CF_ACCESS_CLIENT_SECRET` terisi, workflow mengirim header `CF-Access-Client-Id` dan `CF-Access-Client-Secret` pada request webhook/API agar Cloudflare Access mengizinkan request CI tanpa login email interaktif.

Manual Dokploy setup:

1. tambahkan GHCR registry `ghcr.io` di Dokploy Registry dengan PAT scope minimal untuk pull package,
2. buat Application/Compose service production dari `docker-compose.dokploy.yml`,
3. pilih source **Git** jika tidak memakai GitHub App Dokploy,
4. isi repository URL SSH `git@github.com:<owner>/<repo>.git` untuk private repository atau HTTPS untuk public repository,
5. isi branch `main`,
6. isi Watch Paths sesuai scope service; gunakan `**` untuk single-app repository,
7. isi Environment Dokploy, termasuk `IMAGE_TAG=sha-<commit-sha>` untuk promote production,
8. tambahkan Domain dengan service `api` dan container port `3001`,
9. pastikan tidak ada `ports:` host untuk API, database, atau Redis,
10. jalankan deploy manual pertama dari Dokploy UI untuk memvalidasi clone, env, registry pull, healthcheck, dan domain,
11. sebelum deploy migration production, review output migration gate; isi `BACKUP_EVIDENCE_URL` bila backup evidence tersedia, dan set `APPROVE_DESTRUCTIVE_MIGRATIONS=true` hanya jika migration destructive sudah direview.

### Manual Production Workflow Setup

Workflow `.github/workflows/deploy-production.yml` bisa dijalankan manual melalui GitHub Actions untuk build/push image, menjalankan migration safety gate, approval production, trigger Dokploy, dan smoke test publik.

Setup sekali:

1. GitHub repo → **Settings** → **Environments** → buat `production`,
2. aktifkan **Required reviewers** untuk approval manual sebelum deploy job,
3. tambahkan Environment Secret `DOKPLOY_WEBHOOK_URL` jika Dokploy webhook dipakai,
4. jika tidak memakai webhook, isi `DOKPLOY_URL`, `DOKPLOY_API_TOKEN`, dan `DOKPLOY_APPLICATION_ID`,
5. jika panel Dokploy dilindungi Cloudflare Zero Trust Access, buat policy **Service Auth** untuk service token dan simpan `CF_ACCESS_CLIENT_ID` + `CF_ACCESS_CLIENT_SECRET` sebagai Environment Secret,
6. tambahkan Repository Variable `PRODUCTION_DOMAIN=api.example.com`,
7. tambahkan Repository Variable `APPROVE_DESTRUCTIVE_MIGRATIONS=false`,
8. opsional tambahkan `BACKUP_EVIDENCE_URL` saat migration non-destructive punya bukti backup.

Cara menjalankan manual:

1. buka GitHub → **Actions** → **Deploy Production Dokploy**,
2. klik **Run workflow**,
3. pilih branch/tag sumber,
4. kosongkan `image_tag` untuk build dan promote commit saat ini sebagai `sha-<commit-sha>`,
5. isi `image_tag=sha-<commit-sha>` hanya jika ingin redeploy image immutable existing,
6. set `trigger_dokploy=true` hanya jika Environment Dokploy `IMAGE_TAG` sudah sama dengan tag yang akan dipromote,
7. approve environment `production` saat GitHub meminta approval,
8. tunggu smoke test `/health/live`, `/health/ready`, `/openapi.yaml`, dan protected route selesai.

Jika menggunakan provider **Git** langsung di Dokploy dan bukan image GHCR, workflow manual GitHub tidak mengubah source Dokploy. Gunakan workflow hanya untuk gate/smoke terkontrol, atau deploy manual dari Dokploy UI setelah push branch production.

## Cutover Wave Runner

Untuk eksekusi cutover domain secara berurutan gunakan:

- `make cutover-wave WAVE=64` untuk satu wave,
- `make cutover-all` atau `./scripts/cutover-wave.sh all` untuk wave 64-68 serial.

Perilaku runner:

- fail-fast: wave berikutnya tidak berjalan saat wave aktif gagal,
- optional rollback command saat gagal (`RUN_ROLLBACK_ON_FAILURE=true` + `CUTOVER_ROLLBACK_COMMAND`),
- evidence otomatis per eksekusi di `artifacts/cutover/` (summary + log per-wave + report E2E wave domain).

## Database Migration Order

Aturan urutan migrasi:

1. apply migration sebelum trafik penuh ke versi aplikasi baru,
2. gunakan pola expand-then-contract untuk perubahan yang memengaruhi kontrak aktif,
3. perubahan destruktif hanya setelah masa kompatibilitas selesai,
4. migration failure harus menghentikan rollout,
5. migration non-destructive boleh lanjut tanpa `BACKUP_EVIDENCE_URL` pada tahap awal, tetapi gate akan memberi warning,
6. migration destructive wajib berhenti sampai `APPROVE_DESTRUCTIVE_MIGRATIONS=true` diset setelah review.

Jika migration gagal:

- hentikan perpindahan trafik,
- jalankan prosedur rollback sesuai [Rollback Runbook](/Users/macbookpro/Development/recova-backend-v2/docs/operations/rollback.md),
- dokumentasikan penyebab sebelum retry deploy.

## Smoke Verification

Setelah deploy, minimal lakukan:

- `GET /health/live` sukses,
- `GET /health/ready` sukses,
- endpoint kritis auth dan data utama merespons sesuai kontrak,
- error rate dan latency tidak melewati baseline alarm.

Daftar lengkap ada di [Post-Deploy Checks](/Users/macbookpro/Development/recova-backend-v2/docs/operations/post-deploy-checks.md).

## Runtime Decommission Gate

Setelah runtime aktif stabil dan rollback rehearsal tervalidasi, jalankan gate decommission runtime legacy:

- `make runtime-decommission`

Gate ini memverifikasi:

- traffic publik runtime legacy sudah nol,
- evidence rollback rehearsal tersedia dalam window retention,
- arsip konfigurasi legacy berhasil disimpan.

Evidence output default disimpan pada:

- `artifacts/decommission/**`.

## Deployment Evidence

Setiap deploy harus punya bukti:

- artifact identifier (tag/digest),
- migration version yang diterapkan,
- timestamp deploy,
- hasil smoke checks,
- keputusan promote/rollback.

## Security and Access Rules

- akses deploy hanya untuk role terotorisasi,
- kredensial deploy dipisah per environment,
- aktivitas deploy harus audit-able,
- akses shell ke runtime produksi dibatasi sesuai prinsip least privilege.

## Related Documents

- [Release Gates](/Users/macbookpro/Development/recova-backend-v2/docs/operations/release-gates.md)
- [CI/CD Operations](/Users/macbookpro/Development/recova-backend-v2/docs/operations/ci-cd.md)
- [Database Migrations](/Users/macbookpro/Development/recova-backend-v2/docs/operations/database-migrations.md)
- [Rollback Runbook](/Users/macbookpro/Development/recova-backend-v2/docs/operations/rollback.md)
- [Post-Deploy Checks](/Users/macbookpro/Development/recova-backend-v2/docs/operations/post-deploy-checks.md)
- [Runtime Decommission](/Users/macbookpro/Development/recova-backend-v2/docs/operations/runtime-decommission.md)

## Source Reference

- [/Users/macbookpro/Development/bisakerja-api/docs/operations/deployment.md](/Users/macbookpro/Development/bisakerja-api/docs/operations/deployment.md)
- [Deploying with GitHub Actions](https://docs.github.com/actions/deployment/deploying-with-github-actions)
- [Fiber App API](https://docs.gofiber.io/api/app/)
- [Docker Multi-stage Builds](https://docs.docker.com/build/building/multi-stage/)
- [docker compose pull](https://docs.docker.com/reference/cli/docker/compose/pull/)
- [docker compose up](https://docs.docker.com/reference/cli/docker/compose/up/)
- [golang-migrate](https://github.com/golang-migrate/migrate)
- [golang-migrate PostgreSQL Driver](https://github.com/golang-migrate/migrate/tree/master/database/postgres)
- [PostgreSQL Backup and Restore](https://www.postgresql.org/docs/current/backup.html)
