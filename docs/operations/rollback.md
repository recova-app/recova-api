---
title: Recova Backend Rollback Runbook
description: Prosedur rollback aplikasi dan database untuk memulihkan layanan setelah deploy gagal atau regresi pasca-rilis.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/operations/rollback.md
last_reviewed: 2026-05-09
---

# Recova Backend Rollback Runbook

Dokumen ini menetapkan jalur rollback operasional agar pemulihan layanan cepat, aman, dan konsisten.

## Rollback Scope

Rollback mencakup dua domain:

- aplikasi runtime (artifact/version),
- skema/data database.

Keduanya tidak selalu bisa diputar balik bersamaan. Keputusan rollback harus berbasis kompatibilitas skema.

## Rollback Decision Matrix

| Kondisi                                  | Tindakan utama                                                                      |
| ---------------------------------------- | ----------------------------------------------------------------------------------- |
| regresi aplikasi, skema masih kompatibel | rollback aplikasi ke artifact stabil sebelumnya                                     |
| migration gagal sebelum trafik penuh     | hentikan rollout, pulihkan skema bila aman, lalu rollback aplikasi                  |
| migration destruktif sudah diterapkan    | hindari rollback skema langsung; gunakan forward-fix + rollback aplikasi kompatibel |
| insiden kritis keamanan                  | aktifkan emergency mitigation, batasi trafik, deploy hotfix terverifikasi           |

## Rollback Preconditions

- artifact stabil sebelumnya tersedia dan tervalidasi,
- status migrasi terbaru diketahui,
- backup database terkini tersedia untuk skenario pemulihan data,
- kanal komunikasi insiden aktif,
- owner on-call teridentifikasi.

## Application Rollback Procedure

```text
1) Freeze new rollout
2) Route traffic to last known good artifact
3) Restart affected workloads safely
4) Run liveness/readiness checks
5) Verify key APIs
6) Monitor error and latency stabilization
```

Aturan:

- rollback aplikasi tidak boleh menulis skema baru,
- jika ada cache berisi format baru yang inkompatibel, purge terkontrol sebelum trafik penuh.
- untuk jalur staging berbasis image registry, rollback utama dilakukan dengan mengembalikan `APP_IMAGE` ke tag immutable sebelumnya lalu `docker compose up -d --wait`.

## Database Rollback Constraints

- rollback schema hanya dilakukan bila migration menyediakan down path yang aman,
- perubahan destruktif (drop/rename/irreversible transform) wajib mengutamakan restore backup atau forward-fix,
- keputusan rollback database harus disetujui owner database.

Prioritas keamanan data:

1. lindungi integritas data,
2. pulihkan availability,
3. perbaiki kontrak aplikasi.

## Forward-Fix Path

Gunakan forward-fix jika rollback schema tidak aman:

- pertahankan skema saat ini,
- deploy patch aplikasi yang kompatibel,
- lakukan migrasi korektif non-destruktif,
- validasi ulang melalui post-deploy checks.

## Emergency Hotfix Path

Hotfix darurat digunakan jika layanan tidak dapat menunggu siklus rilis normal.

Syarat minimum hotfix:

- perubahan sekecil mungkin,
- scope difokuskan ke akar masalah,
- smoke checks wajib,
- review ringkas oleh minimal satu reviewer berwenang.

## Validation After Rollback

Wajib verifikasi:

- health endpoints hijau,
- endpoint kritis auth, users, routine, journals, community, ai berfungsi,
- error rate turun ke baseline,
- tidak ada migration lock tersisa.

## Rollback Evidence

Catat bukti operasi:

- waktu insiden,
- versi yang gagal dan versi yang dipulihkan,
- status migrasi,
- hasil verifikasi,
- tindakan lanjutan pencegahan.

## Rollback Rehearsal

Rollback rehearsal wajib dijalankan berkala pada environment non-production untuk memvalidasi jalur rollback tetap executable.

Runner lokal/CI:

- `make rollback-rehearsal`

Input penting:

- `RECOVA_DB_INTEGRATION_URL` wajib menunjuk database `*_test`,
- `ROLLBACK_REHEARSAL_COMMAND` wajib berisi command rollback yang ingin direhearse,
- `ROLLBACK_REHEARSAL_WAVE` memilih domain wave (`65|66|67|68`).

Output evidence:

- `artifacts/rollback-rehearsal/*-summary.log`
- `artifacts/rollback-rehearsal/*-rollback-rehearsal-report.json`

## Related Documents

- [Deployment Workflow](/Users/macbookpro/Development/recova-backend-v2/docs/operations/deployment.md)
- [Post-Deploy Checks](/Users/macbookpro/Development/recova-backend-v2/docs/operations/post-deploy-checks.md)
- [Database Migrations](/Users/macbookpro/Development/recova-backend-v2/docs/operations/database-migrations.md)
- [Incident Triage](/Users/macbookpro/Development/recova-backend-v2/docs/operations/incident-triage.md)

## Source Reference

- [/Users/macbookpro/Development/bisakerja-api/docs/operations/deployment.md](/Users/macbookpro/Development/bisakerja-api/docs/operations/deployment.md)
- [golang-migrate](https://github.com/golang-migrate/migrate)
- [PostgreSQL Backup and Restore](https://www.postgresql.org/docs/current/backup.html)
