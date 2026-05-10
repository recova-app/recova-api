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

- `current_streak`,
- `longest_streak`,
- `total_checkins`,
- `streak_calendar`,
- `relapse_count`,
- `relapse_rate`,
- `recovery_success_rate`,
- `checkin_consistency_score`,
- `weekly_progress`,
- `monthly_progress`,
- `mood_trend`.

Field minimum `GET /api/v1/routine/statistics/activity-summary`:

- `recent_activity`,
- `window_days`,
- `successful_checkins`,
- `relapses`,
- `active_days`.

Contoh struktur `mood_trend`:

- `[]` berisi `{date, dominantMood, successfulRatio}`.

Contoh struktur `weekly_progress`/`monthly_progress`:

- `{window_days, current_successful_checkins, previous_successful_checkins, delta, delta_rate}`.

## Computation Rules

- statistik dibaca dari sumber data check-in dan streak yang konsisten,
- `total_checkins` menghitung check-in sukses valid non-duplicate,
- `relapse_count` menghitung check-in gagal valid non-duplicate,
- `relapse_rate` = `relapse_count / (successful_checkins + relapse_count)`,
- `recovery_success_rate` = `successful_checkins / (successful_checkins + relapse_count)`,
- `checkin_consistency_score` memakai rasio hari aktif pada rolling 30 hari,
- `weekly_progress` dan `monthly_progress` dihitung dari baseline window sebelumnya (week-over-week dan month-over-month),
- semua angka statistik harus non-negatif,
- data null pada pengguna baru harus dimap ke nilai nol yang aman.
- boundary hitung harian memakai UTC,
- semua rasio wajib dibulatkan konsisten (misalnya 2 desimal) sebelum dikirim ke client.
- endpoint `activity-summary` memakai default `window_days=30` jika query tidak dikirim.

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

- field statistik existing (`current_streak`, `longest_streak`, `total_checkins`, `streak_calendar`) tetap dipertahankan untuk kompatibilitas klien,
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
