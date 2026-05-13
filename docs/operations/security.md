---
title: Recova Backend Security Operations Baseline
description: Baseline kontrol keamanan operasional untuk API Recova Backend meliputi hardening HTTP, autentikasi, validasi input, perlindungan data sensitif, dan verifikasi sebelum rilis.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/operations/security.md
last_reviewed: 2026-05-08
---

# Recova Backend Security Operations Baseline

Dokumen ini menetapkan kontrol keamanan minimum sebelum layanan menerima trafik production.

## Security Objectives

- mencegah akses tidak sah,
- membatasi dampak input berbahaya,
- menjaga data sensitif tidak bocor melalui log atau error,
- menyediakan checklist verifikasi keamanan yang konsisten.

## Security Control Matrix

| Area               | Baseline control                                                      |
| ------------------ | --------------------------------------------------------------------- |
| CORS               | hanya origin allowlist eksplisit, tidak wildcard pada mode kredensial |
| HTTP headers       | aktifkan security headers (helmet) dengan konfigurasi eksplisit       |
| Rate limiting      | limiter berbasis endpoint group: auth, AI, dan community write        |
| Request size       | batas ukuran body untuk mencegah abuse payload besar                  |
| Auth               | validasi JWT `iss`, `aud`, `exp`, `sub`, signature, dan jenis token   |
| Secret handling    | redaksi field sensitif di log dan payload error                       |
| Data access        | query melalui repository + parameter binding aman                     |
| Dependency hygiene | scan kerentanan dependency dengan `govulncheck`                       |

## Middleware Hardening Order

Urutan middleware direkomendasikan:

1. request id,
2. recover,
3. structured logger,
4. security headers,
5. CORS,
6. limiter,
7. auth middleware,
8. input validation,
9. route handlers.

Tujuan urutan ini:

- request id tersedia untuk seluruh log/error,
- panic ditangkap lebih awal,
- security policy dipasang sebelum route bisnis,
- limiter menahan abuse sebelum beban service meningkat.

## CORS Policy

Aturan wajib:

- `AllowOrigins` harus berupa daftar origin terverifikasi,
- `AllowCredentials=true` hanya untuk origin eksplisit,
- dilarang menggunakan wildcard origin jika kredensial diaktifkan,
- perubahan origin policy harus melewati review keamanan.

## Authentication and Session Security

- access token berumur pendek,
- refresh token opaque dalam cookie `HttpOnly`,
- refresh token disimpan dalam bentuk hash,
- rotasi refresh token wajib pada setiap refresh sukses,
- logout harus melakukan revocation refresh credential.

## Password and Token Hashing

- password harus di-hash sebelum persistence,
- token yang tersimpan harus dalam bentuk hash,
- larang penyimpanan plaintext secret,
- kebijakan upgrade parameter hashing harus terdokumentasi.

## Input Validation and Abuse Protection

- semua body/query/params harus tervalidasi,
- tetapkan batas panjang string dan ukuran payload,
- tolak field tidak dikenal pada endpoint sensitif,
- endpoint auth dan AI menggunakan limiter lebih ketat.

## SQL Injection and Query Safety

Aturan inti untuk GORM:

- gunakan placeholder/argument binding,
- dilarang menyusun query dari string format input user,
- whitelist kolom untuk sorting/filter dinamis,
- validasi tipe identifier sebelum query by id.

## Sensitive Logging and Redaction

Data berikut tidak boleh muncul di log:

- password atau hash password,
- access token, refresh token, OTP, reset token,
- API key provider eksternal,
- isi jurnal mentah atau chat AI mentah,
- connection string database lengkap.

Log aman minimal berisi:

- `requestId`, method, route, status, latency,
- error code terstandar,
- dependency name untuk kegagalan downstream.

## Dependency and Build Security

- pin versi dependency,
- jalankan vulnerability scan via `make security-scan`,
- blok rilis jika ada kerentanan critical/high yang belum di-approve,
- dokumentasikan exception risk dengan owner dan batas waktu remediation.

## Security Verification

Gunakan [Security Checklist](/Users/macbookpro/Development/recova-backend-v2/docs/operations/security-checklist.md) sebagai gate wajib sebelum rilis.

## Related Documents

- [Secure Coding Standard](/Users/macbookpro/Development/recova-backend-v2/docs/standards/secure-coding.md)
- [API Response Standard](/Users/macbookpro/Development/recova-backend-v2/docs/api-response-standard.md)
- [Auth Token Strategy ADR](/Users/macbookpro/Development/recova-backend-v2/docs/decisions/adr-0007-auth-token-strategy.md)
- [Health and Readiness](/Users/macbookpro/Development/recova-backend-v2/docs/operations/health-readiness.md)

## Source Reference

- [Fiber CORS Middleware](https://docs.gofiber.io/middleware/cors/)
- [Fiber Helmet Middleware](https://docs.gofiber.io/middleware/helmet/)
- [Fiber Limiter Middleware](https://docs.gofiber.io/middleware/limiter/)
- [Fiber Recover Middleware](https://docs.gofiber.io/middleware/recover/)
- [GORM Security Guide](https://gorm.io/docs/security.html)
- [OWASP API Security Top 10](https://owasp.org/API-Security/)
