---
title: Routine Module
description: Kontrak modul routine untuk check-in harian, perhitungan streak, dan statistik pengguna dengan boundary harian berbasis timezone.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/modules/routine.md
last_reviewed: 2026-05-14
---

# Routine Module

## Responsibility

- menerima check-in harian,
- menerima pencatatan relapse harian terpisah,
- menjaga konsistensi check-in per hari,
- menghitung streak,
- menyediakan statistik rutin,
- memicu rekomendasi AI otomatis ketika check-in relapse.

## API Contract

Route prefix:

```text
/api/v1/routine
```

| Method | Path                                          | Auth class | Purpose                            |
| ------ | --------------------------------------------- | ---------- | ---------------------------------- |
| `POST` | `/api/v1/routine/checkin`                     | Bearer     | simpan check-in harian             |
| `POST` | `/api/v1/routine/relapses`                    | Bearer     | simpan trigger relapse hari ini    |
| `GET`  | `/api/v1/routine/statistics`                  | Bearer     | ambil statistik rutin              |
| `GET`  | `/api/v1/routine/statistics/activity-summary` | Bearer     | ambil ringkasan aktivitas periodik |
| `GET`  | `/api/v1/routine/relapses`                    | Bearer     | ambil riwayat relapse              |

## Database Model

Entitas utama:

- `check_ins`,
- `streaks`,
- `routine_statistics` (materialized atau computed).

Constraint minimum:

- unique check-in per `(user_id, local_date)`,
- update streak dan check-in harus berada dalam transaksi yang konsisten.

## Authentication and Authorization

- seluruh endpoint routine wajib bearer auth,
- semua data berbasis `user_id` dari auth context,
- akses ke data pengguna lain ditolak.

## Service and Business Rules

- boundary harian mengikuti tanggal UTC,
- duplicate check-in pada hari UTC yang sama dikembalikan sebagai `CONFLICT` (`409`),
- `is_successful=false` tidak diproses pada endpoint check-in,
- relapse boleh dicatat walau check-in hari itu belum ada, dan kejadian relapse menutup streak aktif,
- endpoint relapse bisa dipanggil setelah check-in sukses pada hari UTC yang sama,
- check-in sukses pada hari yang sudah memiliki relapse tidak boleh membuka/menambah streak,
- jika relapse terjadi di hari berjalan maka `current_streak` pada statistik hari itu harus `0`,
- race condition check-in/relapse paralel harus ditangani aman tanpa overwrite lintas entitas.

## Validation Rules

- `mood` wajib dalam enum/format yang didukung,
- `commitment` wajib valid sesuai batas panjang,
- `is_successful` pada endpoint check-in wajib `true`,
- `mood` wajib pada endpoint `POST /api/v1/routine/relapses`,
- `relapse_trigger` hanya diterima pada endpoint `POST /api/v1/routine/relapses`,
- endpoint relapse wajib menerima minimal satu trigger, tiap item maksimal 500 karakter,
- `window_days` pada endpoint activity summary bersifat opsional dengan rentang `7..90`,
- timestamp/check-in time harus valid,
- request invalid dipetakan ke error validation standar.

## Error Contract

| Condition                        | HTTP  | Error code         |
| -------------------------------- | ----- | ------------------ |
| auth invalid/missing             | `401` | `UNAUTHENTICATED`  |
| payload invalid                  | `422` | `VALIDATION_ERROR` |
| check-in conflict non-idempotent | `409` | `CONFLICT`         |
| user data tidak ditemukan        | `404` | `NOT_FOUND`        |
| kegagalan internal               | `500` | `INTERNAL_ERROR`   |

## Observability Contract

Log field minimum:

- `request_id`,
- `user_id`,
- `local_date`,
- `routine_action`,
- `status_code`.

Metrik minimum:

- check-in success rate,
- duplicate/conflict rate,
- streak computation latency,
- p95 endpoint routine.

## Testing Requirements

- unit test perhitungan streak,
- unit test boundary timezone,
- integration test transaksi check-in + streak,
- handler test auth, validation, idempotency,
- contract test response statistics (field existing + field additive),
- test endpoint activity summary untuk default window dan validasi `window_days`.

## Related Documents

- [Check-Ins Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/check-ins.md)
- [Streaks Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/streaks.md)
- [Statistics Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/statistics.md)

## Source Reference

- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/current/)
