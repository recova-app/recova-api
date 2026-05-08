---
title: Auth Module
description: Kontrak modul autentikasi untuk login Google OAuth, akses token, refresh session, dan kontrol logout/revocation.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/modules/auth.md
last_reviewed: 2026-05-08
---

# Auth Module

## Responsibility

- memverifikasi identity token Google,
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
| `POST` | `/api/v1/auth/onboarding` | Bearer     | melengkapi onboarding awal        |
| `POST` | `/api/v1/auth/refresh`    | Cookie     | refresh access token              |
| `POST` | `/api/v1/auth/logout`     | Bearer     | mengakhiri sesi aktif             |

## Database Model

Entitas utama:

- `users` (identitas akun),
- `auth_refresh_tokens` (hash refresh token, expiry, revoked state, dan riwayat rotasi token).

Constraint minimum:

- relasi token ke `user_id` harus konsisten,
- token mentah tidak disimpan plaintext di database.

## Authentication and Authorization

- endpoint `/google` public,
- endpoint lain wajib autentikasi sesuai kontrak,
- `user_id` sumber kebenaran berasal dari token tervalidasi,
- endpoint auth tidak boleh menerima override principal dari payload klien.

## Service and Business Rules

- verifikasi claim `iss`, `aud`, `exp`, `sub` untuk token Google,
- access token berlaku pendek,
- refresh token harus dirotasi pada refresh sukses,
- logout bersifat idempotent.

## Validation Rules

- token input wajib non-empty,
- claim wajib tervalidasi sebelum issuance session,
- header auth dan cookie harus sesuai format yang didukung,
- request invalid dipetakan ke `VALIDATION_ERROR` atau `UNAUTHENTICATED`.

## Error Contract

| Condition                         | HTTP  | Error code         |
| --------------------------------- | ----- | ------------------ |
| token Google tidak valid          | `401` | `UNAUTHENTICATED`  |
| refresh token tidak valid/revoked | `401` | `UNAUTHENTICATED`  |
| payload tidak valid               | `422` | `VALIDATION_ERROR` |
| konflik sesi                      | `409` | `CONFLICT`         |
| kegagalan internal                | `500` | `INTERNAL_ERROR`   |

## Observability Contract

Log field minimum:

- `request_id`,
- `user_id` (jika ada),
- `auth_action` (`google_login`, `refresh`, `logout`),
- `status_code`.

Metrik minimum:

- auth request rate,
- auth failure rate,
- token refresh success ratio,
- p95 latency endpoint auth.

## Testing Requirements

- unit test verifikasi claim token,
- unit test rotasi refresh token,
- handler test untuk `401`, `422`, dan success path,
- integration test revocation/logout idempotent,
- contract test error envelope endpoint auth.

## Implementation Notes

- access token diterbitkan sebagai JWT `HS256` berumur pendek dengan claim `iss`, `sub`, `exp`, `iat`, dan `token_type=access`,
- refresh token diterbitkan sebagai JWT `HS256` dengan claim `token_type=refresh`,
- refresh token disimpan persisten sebagai hash SHA-256 pada tabel `auth_refresh_tokens`,
- rotasi refresh token dilakukan setiap refresh sukses (`old token revoked` -> `new token inserted`),
- logout tetap idempotent walaupun cookie refresh tidak tersedia,
- endpoint `/api/v1/auth/onboarding` diproteksi bearer token dan tidak menerima override `user_id` dari body.

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
