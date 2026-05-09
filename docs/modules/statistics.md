---
title: Statistics Module
description: Kontrak statistik pemulihan pengguna mencakup streak, relapse metrics, konsistensi check-in, tren mood, dan ringkasan progres periodik.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/modules/statistics.md
last_reviewed: 2026-05-09
---

# Statistics Module

Dokumen ini mendefinisikan kontrak statistik pemulihan pengguna untuk domain routine.

## Route Contract

```text
GET /api/v1/routine/statistics
```

```text
GET /api/v1/routine/statistics/activity-summary
```

## Response Contract

Field minimum `GET /api/v1/routine/statistics`:

- `currentStreak`,
- `longestStreak`,
- `totalCheckins`,
- `streakCalendar`,
- `relapseCount`,
- `relapseRate`,
- `recoverySuccessRate`,
- `checkinConsistencyScore`,
- `weeklyProgress`,
- `monthlyProgress`,
- `moodTrend`.

Field minimum `GET /api/v1/routine/statistics/activity-summary`:

- `recentActivity`,
- `windowDays`,
- `successfulCheckins`,
- `relapses`,
- `activeDays`.

Contoh struktur `moodTrend`:

- `[]` berisi `{date, dominantMood, successfulRatio}`.

Contoh struktur `weeklyProgress`/`monthlyProgress`:

- `{windowDays, currentSuccessfulCheckins, previousSuccessfulCheckins, delta, deltaRate}`.

## Computation Rules

- statistik dibaca dari sumber data check-in dan streak yang konsisten,
- `totalCheckins` menghitung check-in sukses valid non-duplicate,
- `relapseCount` menghitung check-in gagal valid non-duplicate,
- `relapseRate` = `relapseCount / (successfulCheckins + relapseCount)`,
- `recoverySuccessRate` = `successfulCheckins / (successfulCheckins + relapseCount)`,
- `checkinConsistencyScore` memakai rasio hari aktif pada rolling 30 hari,
- `weeklyProgress` dan `monthlyProgress` dihitung dari baseline window sebelumnya (week-over-week dan month-over-month),
- semua angka statistik harus non-negatif,
- data null pada pengguna baru harus dimap ke nilai nol yang aman.
- boundary hitung harian memakai UTC,
- semua rasio wajib dibulatkan konsisten (misalnya 2 desimal) sebelum dikirim ke client.
- endpoint `activity-summary` memakai default `windowDays=30` jika query tidak dikirim.

## Freshness Rules

- statistik harus mencerminkan state terbaru setelah check-in sukses,
- perubahan relapse wajib tercermin pada metrik statistik dalam SLA freshness yang sama,
- jika pipeline asynchronous digunakan, kontrak eventual consistency harus disebutkan eksplisit di API response metadata.

## Ownership Rules

- pengguna hanya boleh membaca statistik miliknya sendiri,
- akses lintas pengguna tanpa otorisasi eksplisit tidak diperbolehkan.

## Observability Rules

- log statistik tidak boleh menyertakan konten sensitif pengguna,
- mismatch nilai statistik vs histori check-in harus bisa ditelusuri lewat request id dan audit internal,
- log diagnostik statistik hanya boleh menyimpan agregat dan identifier teknis tanpa payload jurnal/chat.

## Compatibility Rules

- field statistik existing (`currentStreak`, `longestStreak`, `totalCheckins`, `streakCalendar`) tetap dipertahankan untuk kompatibilitas klien,
- field baru bersifat additive dan optional-safe pada klien lama,
- penghapusan atau rename field statistik existing harus dianggap breaking change.

## Related Documents

- [Routine Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/routine.md)
- [Check-Ins Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/check-ins.md)
- [Streaks Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/streaks.md)
- [Data Sensitivity Matrix](/Users/macbookpro/Development/recova-backend-v2/docs/references/data-sensitivity-matrix.md)

## Source Reference

- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
- [PostgreSQL WITH Queries](https://www.postgresql.org/docs/current/queries-with.html)
