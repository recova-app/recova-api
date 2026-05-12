---
title: ADR 0007 - Auth Token Strategy
description: Keputusan strategi sesi autentikasi berbasis access JWT pendek dan refresh token cookie ter-rotasi.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/decisions/adr-0007-auth-token-strategy.md
last_reviewed: 2026-05-12
---

# ADR 0007 - Auth Token Strategy

## Status

Proposed

## Context

Layanan memerlukan strategi sesi yang:

- aman untuk client web/mobile,
- mendukung durasi sesi pengguna yang wajar,
- tetap membatasi dampak kebocoran token,
- kompatibel dengan alur login Google OAuth dan login manual email/username/password.

## Decision

Gunakan model hybrid:

- access token JWT berumur pendek untuk request authorization,
- refresh token opaque dalam cookie `HttpOnly` untuk perpanjangan sesi,
- rotasi refresh token setiap refresh sukses,
- penyimpanan refresh token persisten dalam bentuk hash.

## Decision Drivers

- token pendek menurunkan blast radius jika access token bocor,
- refresh token cookie meminimalkan paparan token ke JavaScript,
- rotasi refresh token mempersempit peluang replay,
- kompatibel dengan praktik OAuth web server dan kontrol CSRF/CORS.

## Alternatives Considered

### A1 - Access JWT saja tanpa refresh

- plus: implementasi sederhana.
- minus: UX buruk karena login ulang sering.
- hasil: tidak dipilih.

### A2 - Long-lived JWT tanpa rotasi

- plus: minim endpoint sesi.
- minus: risiko keamanan tinggi jika token bocor.
- hasil: ditolak.

### A3 - Server session state penuh tanpa JWT

- plus: revocation sederhana.
- minus: scaling stateful lebih kompleks lintas instance.
- hasil: tidak dipilih sebagai baseline.

### A4 - Access JWT pendek + refresh cookie ter-rotasi

- plus: keseimbangan keamanan, UX, dan operasional.
- minus: membutuhkan endpoint refresh/logout dan kontrol cookie+CSRF lebih ketat.
- hasil: dipilih.

## Consequences

Konsekuensi positif:

- kontrol sesi lebih aman dibanding JWT long-lived,
- logout/revocation dapat dieksekusi tegas pada refresh layer,
- tetap kompatibel untuk client yang memakai bearer access token.

Konsekuensi negatif:

- menambah kompleksitas middleware dan storage sesi,
- perlu disiplin tinggi pada CORS/CSRF untuk endpoint cookie.

## Guardrails

- whitelist algoritma JWT secara eksplisit,
- verifikasi `iss`, `aud`, `exp`, `sub` wajib,
- refresh token mentah tidak boleh disimpan di DB/log,
- password manual harus di-hash (bcrypt) dan tidak boleh tersimpan/log sebagai plaintext,
- `AllowCredentials=true` hanya dengan origin allowlist,
- wildcard origin dilarang saat credentials aktif,
- endpoint reset data dev-only tidak boleh aktif production.

## Related Documents

- [Auth Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/auth.md)
- [Authentication and Trust Boundaries](/Users/macbookpro/Development/recova-backend-v2/docs/overview/authentication-and-trust-boundaries.md)

## Source Reference

- [Google OAuth2 Web Server Applications](https://developers.google.com/identity/protocols/oauth2/web-server)
- [Google Backend Auth Verification](https://developers.google.com/identity/sign-in/web/backend-auth)
- [JWT Best Current Practices (RFC 8725)](https://www.rfc-editor.org/rfc/rfc8725)
- [Fiber CORS Middleware](https://docs.gofiber.io/middleware/cors/)
- [Fiber CSRF Middleware](https://docs.gofiber.io/middleware/csrf/)
- [Fiber JWT Middleware](https://docs.gofiber.io/contrib/v3_jwt_v1.x.x/jwt/)
