---
title: Auth Module
description: Kontrak modul autentikasi untuk login Google OAuth, register/login manual, akses token, refresh session, dan kontrol logout/revocation.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/modules/auth.md
last_reviewed: 2026-05-12
---

# Auth Module

## Responsibility

- memverifikasi identity token Google,
- membuat akun manual berbasis email/username/password,
- menerbitkan access token,
- mengelola refresh token jika mode refresh aktif,
- menangani logout dan revocation sesi,
- menyediakan middleware auth untuk endpoint terproteksi.

Di luar cakupan modul ini: manajemen konten domain (routine, journals, community).

## API Contract

Route prefix:

```text
/api/v1/auth
```

| Method | Path                      | Auth class | Purpose                           |
| ------ | ------------------------- | ---------- | --------------------------------- |
| `POST` | `/api/v1/auth/google`     | Public     | login/registrasi via Google token |
| `POST` | `/api/v1/auth/register`   | Public     | registrasi akun manual + sesi     |
| `POST` | `/api/v1/auth/login`      | Public     | login manual via email/username   |
| `POST` | `/api/v1/auth/onboarding` | Bearer     | melengkapi onboarding awal        |
| `POST` | `/api/v1/auth/refresh`    | Cookie     | refresh access token              |
| `POST` | `/api/v1/auth/logout`     | Bearer     | mengakhiri sesi aktif             |

## Database Model

Entitas utama:

- `users` (identitas akun),
- `auth_refresh_tokens` (hash refresh token, expiry, revoked state, dan riwayat rotasi token).

Constraint minimum:

- akun Google memakai `google_id`; akun manual boleh `google_id` kosong (`NULL`),
- `email` unik global,
- `username` unik untuk akun manual,
- password disimpan hanya sebagai `password_hash`,
- relasi token ke `user_id` harus konsisten,
- token mentah tidak disimpan plaintext di database.

## Authentication and Authorization

- endpoint `/google` public,
- endpoint `/register` dan `/login` public,
- endpoint lain wajib autentikasi sesuai kontrak,
- `user_id` sumber kebenaran berasal dari token tervalidasi,
- endpoint auth tidak boleh menerima override principal dari payload klien.

## Service and Business Rules

- verifikasi claim `iss`, `aud`, `exp`, `sub` untuk token Google,
- register manual validasi `email`, `username`, `password`, `confirm_password`,
- login manual menerima satu identifier (`identifier` atau fallback `email`/`username`) + `password`,
- register manual langsung membuat sesi aktif (tanpa OTP/email verification),
- access token berlaku pendek,
- refresh token harus dirotasi pada refresh sukses,
- logout bersifat idempotent.

## Validation Rules

- token input wajib non-empty,
- `password` minimal 8 karakter dan maksimal 72 byte,
- `confirm_password` wajib sama dengan `password`,
- `username` hanya boleh alfanumerik + underscore,
- claim wajib tervalidasi sebelum issuance session,
- header auth dan cookie harus sesuai format yang didukung,
- request invalid dipetakan ke `VALIDATION_ERROR` atau `UNAUTHENTICATED`.

## Error Contract

| Condition                         | HTTP  | Error code         |
| --------------------------------- | ----- | ------------------ |
| token Google tidak valid          | `401` | `UNAUTHENTICATED`  |
| akun manual tidak ditemukan       | `401` | `UNAUTHENTICATED`  |
| password manual tidak cocok       | `401` | `UNAUTHENTICATED`  |
| email atau username sudah dipakai | `409` | `CONFLICT`         |
| mismatch confirm password         | `422` | `VALIDATION_ERROR` |
| refresh token tidak valid/revoked | `401` | `UNAUTHENTICATED`  |
| payload tidak valid               | `422` | `VALIDATION_ERROR` |
| konflik sesi                      | `409` | `CONFLICT`         |
| kegagalan internal                | `500` | `INTERNAL_ERROR`   |

## Observability Contract

Log field minimum:

- `request_id`,
- `user_id` (jika ada),
- `auth_action` (`manual_register`, `manual_login`, `google_login`, `refresh`, `logout`),
- `status_code`.

Metrik minimum:

- auth request rate,
- auth failure rate,
- token refresh success ratio,
- p95 latency endpoint auth.

## Testing Requirements

- unit test verifikasi claim token,
- unit test validator register/login manual,
- unit test rotasi refresh token,
- unit test service register/login success + negative path,
- handler test untuk `401`, `422`, dan success path,
- repository integration test untuk unique `email`/`username` + penyimpanan `password_hash`,
- integration test revocation/logout idempotent,
- contract test error envelope endpoint auth.

## Implementation Notes

- access token diterbitkan sebagai JWT `HS256` berumur pendek dengan claim `iss`, `sub`, `exp`, `iat`, dan `token_type=access`,
- refresh token diterbitkan sebagai JWT `HS256` dengan claim `token_type=refresh`,
- refresh token disimpan persisten sebagai hash SHA-256 pada tabel `auth_refresh_tokens`,
- rotasi refresh token dilakukan setiap refresh sukses (`old token revoked` -> `new token inserted`),
- logout tetap idempotent walaupun cookie refresh tidak tersedia,
- endpoint `/api/v1/auth/onboarding` diproteksi bearer token dan tidak menerima override `user_id` dari body.
- hash password manual memakai bcrypt (`GenerateFromPassword`/`CompareHashAndPassword`) dengan batas bcrypt maksimal 72 byte.

## Related Documents

- [Users Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/users.md)
- [Authentication and Trust Boundaries](/Users/macbookpro/Development/recova-backend-v2/docs/overview/authentication-and-trust-boundaries.md)
- [API Response Standard](/Users/macbookpro/Development/recova-backend-v2/docs/api-response-standard.md)

## Source Reference

- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
- [Google Backend Auth Verification](https://developers.google.com/identity/sign-in/web/backend-auth)
- [JWT Best Current Practices](https://www.rfc-editor.org/rfc/rfc8725)
- [Google ID Token Validation for Go](https://pkg.go.dev/google.golang.org/api/idtoken)
- [Go JWT Package](https://pkg.go.dev/github.com/golang-jwt/jwt/v5)
- [Go bcrypt Package](https://pkg.go.dev/golang.org/x/crypto/bcrypt)
- [PostgreSQL Unique Indexes](https://www.postgresql.org/docs/current/indexes-unique.html)
