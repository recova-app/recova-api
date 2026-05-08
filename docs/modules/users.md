---
title: Users Module
description: Kontrak modul users untuk profil pengguna, aturan akses pemilik data, pembaruan pengaturan, dan perlindungan privasi.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/modules/users.md
last_reviewed: 2026-05-08
---

# Users Module

Modul users mengelola data profil pengguna yang dipakai lintas fitur aplikasi.

## Responsibility

Modul users bertanggung jawab pada:

- pengambilan profil pengguna terautentikasi,
- pembaruan pengaturan profil,
- pembatasan akses berdasarkan kepemilikan akun,
- audit event pada perubahan data profil.

## Route Prefix

```text
/api/v1/users
```

## Endpoint Summary

| Method   | Path                          | Auth   | Purpose                                   |
| -------- | ----------------------------- | ------ | ----------------------------------------- |
| `GET`    | `/api/v1/users/me`            | Bearer | Ambil profil milik pengguna saat ini      |
| `PUT`    | `/api/v1/users/settings`      | Bearer | Perbarui pengaturan profil pengguna       |
| `DELETE` | `/api/v1/users/me/reset-data` | Bearer | Reset data pengguna untuk pengujian lokal |

Catatan:

- endpoint reset-data bersifat khusus development dan harus nonaktif di production.

## Profile Data Contract

Field profil minimum yang sudah teridentifikasi:

- `nickname`,
- `recovery_reason`,
- `daily_checkin_time`,
- `onboarding_completed`.

Field tambahan yang umum dibutuhkan:

- `created_at`,
- `updated_at`.

## Ownership Rules

- route `GET /me` dan `PUT /settings` hanya boleh mengakses data milik principal autentikasi saat ini,
- identifier user pada path/query eksternal tidak boleh menimpa context `user_id` dari token,
- akses data pengguna lain harus ditolak.

## Update Rules

- `nickname` wajib tervalidasi format dan panjang,
- `recovery_reason` diperlakukan sebagai data sensitif moderat,
- `daily_checkin_time` wajib valid sebagai jam lokal pengguna (format terdokumentasi),
- update parsial harus jelas: field yang tidak dikirim tidak diubah.

## Privacy and Logging Rules

- jangan log nilai mentah `recovery_reason`,
- jangan log payload penuh request update profil,
- audit log cukup menyimpan event type, user id, request id, dan ringkasan field yang berubah.

## Reset Data Guardrails

Jika endpoint reset data digunakan:

- wajib dikunci environment development,
- wajib menolak di staging/production,
- respons error harus konsisten dan aman saat route dinonaktifkan.

## Open Gaps

- aturan panjang final dan karakter yang diizinkan untuk `nickname`,
- format final `daily_checkin_time` (local time string atau timezone-aware payload),
- kebutuhan audit detail per regulasi data.

## Related Documents

- [Auth Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/auth.md)
- [Onboarding Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/onboarding.md)
- [Data Sensitivity Matrix](/Users/macbookpro/Development/recova-backend-v2/docs/references/data-sensitivity-matrix.md)

## Source Reference

- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
