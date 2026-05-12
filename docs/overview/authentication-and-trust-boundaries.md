---
title: Authentication and Trust Boundaries
description: Definisi trust boundary autentikasi antara client, backend, Google identity provider, storage token, dan data domain.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/overview/authentication-and-trust-boundaries.md
last_reviewed: 2026-05-12
---

# Authentication and Trust Boundaries

Dokumen ini memetakan batas kepercayaan untuk flow autentikasi agar kontrol keamanan dapat diverifikasi lintas komponen.

## Trust Zones

| Zone | Component                    | Trust level                 | Notes                                                                                 |
| ---- | ---------------------------- | --------------------------- | ------------------------------------------------------------------------------------- |
| Z1   | User App (web/mobile client) | Untrusted                   | input dan token dari client selalu dianggap tidak tepercaya sebelum verifikasi server |
| Z2   | Recova Backend API           | Trusted execution           | satu-satunya komponen yang membuat keputusan auth dan ownership                       |
| Z3   | Google Identity Provider     | External trusted dependency | dipakai untuk validasi identity token dan federated sign-in                           |
| Z4   | PostgreSQL                   | Trusted data store          | menyimpan user state dan hash refresh token                                           |
| Z5   | Logs/Observability           | Restricted                  | hanya metadata aman, tanpa payload sensitif                                           |

## Boundary Rules

### Z1 -> Z2 (Client to Backend)

- semua request harus divalidasi,
- bearer token diverifikasi penuh sebelum akses resource terproteksi,
- request berbasis cookie wajib melewati CSRF guard pada unsafe methods,
- response error tidak boleh membocorkan detail kriptografi atau query internal.

### Z2 -> Z3 (Backend to Google)

- backend memverifikasi token terhadap issuer resmi Google,
- backend memastikan `aud`, `iss`, `exp`, dan signature valid,
- kegagalan verifikasi dipetakan ke `UNAUTHENTICATED`.

### Z1 -> Z2 (Manual Credential Auth)

- register manual memvalidasi `email`, `username`, `password`, dan `confirm_password`,
- login manual hanya menerima identifier (`email`/`username`) + `password`,
- password mentah hanya dipakai saat proses verifikasi hash dan tidak boleh di-log.

### Z2 -> Z4 (Backend to Database)

- write/read auth state hanya melalui repository,
- refresh token mentah tidak disimpan ke database,
- password manual hanya disimpan sebagai bcrypt hash,
- penyimpanan token persisten harus berbasis hash.

### Z2 -> Z5 (Backend to Logs)

- log boleh berisi request id, route, status, latency,
- log tidak boleh berisi JWT mentah, refresh token, atau konten jurnal,
- audit event auth dicatat terstruktur.

## Session Model

Model sesi default:

- access token pendek untuk authorisasi request,
- refresh token cookie untuk perpanjangan sesi (opsional via feature flag/env),
- rotasi refresh token setiap refresh sukses.
- register manual sukses langsung membuat sesi aktif tanpa OTP/email verification.

## Cookie Security Contract

Jika cookie dipakai:

- `HttpOnly=true`,
- `Secure=true` di production,
- `SameSite` ditetapkan eksplisit,
- domain/path dibatasi seperlunya,
- origin CORS harus allowlist eksplisit.

## CSRF and Origin Controls

Untuk endpoint cookie-based:

- validasi token CSRF pada method tidak aman,
- validasi `Origin`/`Referer` terhadap trusted origins,
- larangan wildcard origin jika credentials aktif.

## Request Identity and Auditability

- setiap request harus memiliki request-id,
- request-id diteruskan ke log dan error envelope,
- event auth penting (login, refresh, logout, revoke) wajib ter-audit.

## Failure Handling

- invalid/missing token -> `401 UNAUTHENTICATED`,
- akun manual tidak ditemukan -> `401 UNAUTHENTICATED`,
- password manual salah -> `401 UNAUTHENTICATED`,
- token valid tapi ownership gagal -> `403 FORBIDDEN` atau `404` sesuai strategi anti-enumeration,
- dependency auth eksternal gagal -> `502 DOWNSTREAM_ERROR` atau `503 SERVICE_UNAVAILABLE`.

## Related Documents

- [Auth Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/auth.md)
- [Architecture](/Users/macbookpro/Development/recova-backend-v2/docs/architecture.md)
- [Error Taxonomy](/Users/macbookpro/Development/recova-backend-v2/docs/standards/error-taxonomy.md)
- [API Response Standard](/Users/macbookpro/Development/recova-backend-v2/docs/api-response-standard.md)

## Source Reference

- [Google OAuth2 Web Server Applications](https://developers.google.com/identity/protocols/oauth2/web-server)
- [Google Backend Auth Verification](https://developers.google.com/identity/sign-in/web/backend-auth)
- [Fiber RequestID Middleware](https://docs.gofiber.io/middleware/requestid/)
- [Fiber CORS Middleware](https://docs.gofiber.io/middleware/cors/)
- [Fiber CSRF Middleware](https://docs.gofiber.io/middleware/csrf/)
- [JWT Best Current Practices (RFC 8725)](https://www.rfc-editor.org/rfc/rfc8725)
- [Go bcrypt Package](https://pkg.go.dev/golang.org/x/crypto/bcrypt)
