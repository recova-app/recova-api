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
last_reviewed: 2026-05-08
---

# Check-Ins Module

Dokumen ini menetapkan kontrak check-in harian pengguna.

## Route Contract

```text
POST /api/v1/routine/checkin
```

## Input Contract

Field minimum:

- `mood` (kategori atau skala mood sesuai kamus domain),
- `commitment` (teks komitmen harian),
- `submitted_at` (opsional jika server memakai receive-time sebagai sumber waktu utama).

## Validation Rules

- `mood` wajib berada pada daftar nilai yang diizinkan,
- `commitment` wajib memiliki batas panjang minimum/maksimum,
- payload kosong atau format invalid harus ditolak sebagai `VALIDATION_ERROR`.

## Idempotency Rules

Kunci idempotency:

- satu check-in per pengguna per tanggal lokal.

Perilaku duplicate:

- duplicate request hari yang sama harus dipetakan konsisten:
  - opsi A: `CONFLICT` jika check-in sudah ada,
  - opsi B: update terbatas field check-in pada hari yang sama.

Pilihan implementasi akhir wajib konsisten di API reference dan tests.

## Timezone Rules

- tanggal check-in dihitung dari timezone profil pengguna,
- data input waktu dari client tidak boleh langsung dipercaya tanpa normalisasi,
- simpan timestamp server dalam UTC dan simpan `local_date` ter-normalisasi.

## Storage Integrity Rules

- gunakan constraint unik `(user_id, local_date)`,
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
