---
title: Check-Ins Module
description: Kontrak check-in harian untuk pencatatan mood dan komitmen dengan aturan idempotency, timezone, dan integritas data.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/modules/check-ins.md
last_reviewed: 2026-05-13
---

# Check-Ins Module

Dokumen ini menetapkan kontrak check-in harian pengguna.

## Route Contract

```text
POST /api/v1/routine/checkin
POST /api/v1/routine/relapses
```

## Input Contract

Field minimum:

- `mood` (kategori atau skala mood sesuai kamus domain),
- `is_successful` wajib `true` pada endpoint check-in,
- `commitment` (teks komitmen harian),
- `relapse_trigger` dipindah ke endpoint terpisah (`POST /api/v1/routine/relapses`),
- `mood` juga wajib pada endpoint relapse (`POST /api/v1/routine/relapses`),
- `submitted_at` (opsional jika server memakai receive-time sebagai sumber waktu utama).

## Validation Rules

- `mood` wajib berada pada daftar nilai yang diizinkan,
- `commitment` wajib memiliki batas panjang minimum/maksimum,
- `is_successful=false` tidak valid di endpoint check-in dan harus memakai endpoint relapse,
- `relapse_trigger` pada endpoint relapse boleh kirim multiple item, tiap item maksimal 500 karakter,
- `relapse_trigger` tidak boleh dikirim ke endpoint check-in,
- payload kosong atau format invalid harus ditolak sebagai `VALIDATION_ERROR`.

## Idempotency Rules

Kunci idempotency:

- satu check-in per pengguna per tanggal UTC,
- endpoint relapse boleh dipanggil tanpa check-in hari itu,
- endpoint relapse boleh dipanggil setelah check-in hari yang sama tanpa overwrite data check-in.

Perilaku duplicate:

- duplicate request pada hari UTC yang sama dikembalikan sebagai `CONFLICT` (`409`).

## Timezone Rules

- seluruh perhitungan hari check-in menggunakan tanggal UTC,
- data input waktu dari client tidak menjadi sumber utama boundary harian,
- simpan timestamp server dalam UTC.

## Storage Integrity Rules

- gunakan constraint unik `(user_id, check_in_date)` untuk check-in,
- gunakan transaksi saat write check-in + update streak/state terkait,
- error constraint duplicate dipetakan ke respons bisnis yang eksplisit.

## Logging and Privacy

- jangan log teks commitment mentah,
- log hanya metadata operasional (request id, user id, local_date, hasil operasi).

## Related Documents

- [Routine Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/routine.md)
- [Streaks Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/streaks.md)
- [API Response Standard](/Users/macbookpro/Development/recova-backend-v2/docs/api-response-standard.md)

## Source Reference

- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
