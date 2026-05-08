---
title: Users Module
description: Kontrak modul users untuk profil pengguna, pengaturan akun, dan kontrol kepemilikan data profil.
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

## Responsibility

- mengambil profil pengguna saat ini,
- memperbarui pengaturan profil,
- menegakkan ownership data,
- mencatat audit perubahan profil.

## API Contract

Route prefix:

```text
/api/v1/users
```

| Method   | Path                          | Auth class        | Purpose                               |
| -------- | ----------------------------- | ----------------- | ------------------------------------- |
| `GET`    | `/api/v1/users/me`            | Bearer            | ambil profil pengguna saat ini        |
| `PUT`    | `/api/v1/users/settings`      | Bearer            | update pengaturan profil              |
| `DELETE` | `/api/v1/users/me/reset-data` | Bearer (dev-only) | reset data user untuk pengujian lokal |

## Database Model

Entitas utama:

- `users`,
- `profiles` (state completion onboarding + jawaban onboarding),
- `auth_refresh_tokens` (ikut dibersihkan pada reset data development).

Constraint minimum:

- satu profil aktif per `user_id`,
- field sensitif tidak boleh terekspos langsung ke publik.

## Authentication and Authorization

- semua endpoint modul users wajib bearer auth,
- akses hanya ke resource milik principal saat ini,
- `user_id` dari auth context tidak boleh ditimpa payload klien,
- endpoint reset-data hanya aktif di environment development.

## Service and Business Rules

- update profil bersifat parsial terkontrol,
- perubahan field kritis dicatat sebagai audit event,
- reset data harus idempotent dan aman untuk re-run.

## Validation Rules

- `nickname` wajib valid format dan panjang,
- `daily_checkin_time` wajib valid sesuai format waktu yang didukung,
- payload kosong atau field ilegal ditolak,
- field tak dikenal tidak boleh diam-diam diabaikan tanpa kebijakan eksplisit.

## Error Contract

| Condition              | HTTP  | Error code         |
| ---------------------- | ----- | ------------------ |
| auth invalid/missing   | `401` | `UNAUTHENTICATED`  |
| akses bukan milik user | `403` | `FORBIDDEN`        |
| user tidak ditemukan   | `404` | `NOT_FOUND`        |
| payload invalid        | `422` | `VALIDATION_ERROR` |
| kegagalan internal     | `500` | `INTERNAL_ERROR`   |

## Observability Contract

Log field minimum:

- `request_id`,
- `user_id`,
- `profile_action`,
- `status_code`.

Metrik minimum:

- profile read/update rate,
- profile update validation failure rate,
- p95 latency endpoint users.

## Testing Requirements

- unit test validator profil,
- handler test ownership + auth failure,
- integration test update profil dan persistensi,
- test guard reset-data non-development,
- contract test error envelope users routes.

## Implementation Notes

- `nickname` tervalidasi 3-50 karakter,
- `daily_checkin_time` menerima format `HH:mm` dan disimpan pada kolom `users.check_in_time`,
- update settings menerima alias legacy (`userWhy`, `checkinTime`) agar tetap kompatibel dengan payload klien lama,
- endpoint reset-data hanya aktif untuk environment lokal/development dan mengembalikan `403 FORBIDDEN` di environment lain.

## Related Documents

- [Auth Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/auth.md)
- [Data Sensitivity Matrix](/Users/macbookpro/Development/recova-backend-v2/docs/references/data-sensitivity-matrix.md)
- [API Response Standard](/Users/macbookpro/Development/recova-backend-v2/docs/api-response-standard.md)

## Source Reference

- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
