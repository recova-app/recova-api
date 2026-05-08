---
title: Statistics Module
description: Kontrak statistik pengguna terkait rutin harian, termasuk current streak, longest streak, total check-in, dan aturan konsistensi agregasi.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/modules/statistics.md
last_reviewed: 2026-05-08
---

# Statistics Module

Dokumen ini mendefinisikan kontrak statistik pengguna untuk domain routine.

## Route Contract

```text
GET /api/v1/routine/statistics
```

## Response Contract

Field minimum:

- `current_streak`,
- `longest_streak`,
- `total_checkins`.

Field opsional masa depan:

- `last_checkin_date`,
- `relapse_count`,
- ringkasan tren periodik.

## Computation Rules

- statistik dibaca dari sumber data check-in dan streak yang konsisten,
- `total_checkins` menghitung event check-in valid non-duplicate,
- semua angka statistik harus non-negatif,
- data null pada pengguna baru harus dimap ke nilai nol yang aman.

## Freshness Rules

- statistik harus mencerminkan state terbaru setelah check-in sukses,
- jika pipeline asynchronous digunakan, kontrak eventual consistency harus disebutkan eksplisit di API response metadata.

## Ownership Rules

- pengguna hanya boleh membaca statistik miliknya sendiri,
- akses lintas pengguna tanpa otorisasi eksplisit tidak diperbolehkan.

## Observability Rules

- log statistik tidak boleh menyertakan konten sensitif pengguna,
- mismatch nilai statistik vs histori check-in harus bisa ditelusuri lewat request id dan audit internal.

## Related Documents

- [Routine Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/routine.md)
- [Check-Ins Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/check-ins.md)
- [Streaks Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/streaks.md)

## Source Reference

- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
