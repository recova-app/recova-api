---
title: Routine Module
description: Kontrak modul routine untuk check-in harian, streak, statistik pengguna, dan aturan konsistensi harian berbasis timezone.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/modules/routine.md
last_reviewed: 2026-05-08
---

# Routine Module

Modul routine mengelola interaksi kebiasaan harian pengguna: check-in, streak, dan statistik ringkas.

## Responsibility

- menerima check-in harian,
- menjaga konsistensi satu check-in per hari per pengguna,
- menghitung dan memperbarui streak,
- menyajikan statistik rutin pengguna.

## Route Prefix

```text
/api/v1/routine
```

## Endpoint Summary

| Method | Path                         | Auth   | Purpose                         |
| ------ | ---------------------------- | ------ | ------------------------------- |
| `POST` | `/api/v1/routine/checkin`    | Bearer | Simpan check-in harian pengguna |
| `GET`  | `/api/v1/routine/statistics` | Bearer | Ambil ringkasan statistik rutin |
| `GET`  | `/api/v1/routine/relapses`   | Bearer | Ambil riwayat relapse pengguna  |

## Data Contract Summary

Input check-in minimum:

- `mood`,
- `commitment`,
- `timestamp` atau waktu server penerimaan.

Output statistik minimum:

- `current_streak`,
- `longest_streak`,
- `total_checkins`.

## Daily Boundary and Timezone

Aturan harian:

- boundary harian dihitung berdasarkan timezone pengguna,
- jika timezone pengguna belum tersedia, gunakan default layanan yang terdokumentasi,
- perubahan timezone harus memperhatikan dampak pada boundary hari berjalan.

## Consistency Rules

- check-in harian bersifat idempotent untuk kombinasi `(user_id, local_date)`,
- duplicate submit pada hari yang sama harus menghasilkan perilaku konsisten (reject atau update terbatas) sesuai kontrak check-ins,
- pembaruan streak harus terjadi atomik bersama pencatatan check-in.

## Race Condition Prevention

- gunakan unique constraint untuk kunci harian pengguna,
- lakukan write check-in + update streak dalam satu transaksi,
- tangani konflik insert sebagai kasus idempotency, bukan error internal.

## Privacy Rules

- jangan log konten mentah commitment jika dianggap sensitif,
- log hanya metadata ringkas (request id, user id, status).

## Related Documents

- [Check-Ins Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/check-ins.md)
- [Streaks Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/streaks.md)
- [Statistics Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/statistics.md)
- [Database](/Users/macbookpro/Development/recova-backend-v2/docs/database.md)

## Source Reference

- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
