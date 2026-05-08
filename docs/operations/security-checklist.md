---
title: Recova Backend Security Checklist
description: Checklist verifikasi keamanan operasional sebelum deployment untuk memastikan hardening minimum, proteksi data sensitif, dan kontrol akses berjalan sesuai standar.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/operations/security-checklist.md
last_reviewed: 2026-05-08
---

# Recova Backend Security Checklist

Checklist ini digunakan sebelum deployment agar kontrol keamanan minimum tervalidasi.

## 1. Identity and Access

- [ ] JWT validation memverifikasi `iss`, `aud`, `exp`, `sub`, dan signature.
- [ ] Refresh token disimpan dalam bentuk hash.
- [ ] Endpoint logout/revoke menghapus credential aktif.
- [ ] Endpoint internal/admin terlindungi authz eksplisit.

## 2. HTTP and Transport Security

- [ ] CORS menggunakan origin allowlist eksplisit.
- [ ] Tidak ada wildcard origin saat credentials aktif.
- [ ] Security headers aktif.
- [ ] TLS termination pada environment non-local tervalidasi.

## 3. Request Protection

- [ ] Batas ukuran request body aktif.
- [ ] Rate limit global aktif.
- [ ] Rate limit ketat diterapkan untuk auth dan AI endpoint.
- [ ] Endpoint upload memvalidasi MIME type dan ukuran file.

## 4. Input and Query Safety

- [ ] Semua body/query/params tervalidasi schema.
- [ ] Field tak dikenal ditolak pada endpoint sensitif.
- [ ] Query DB memakai parameter binding aman.
- [ ] Sorting/filter dinamis menggunakan whitelist.

## 5. Data Protection and Logging

- [ ] Password/token/API key tidak pernah ditulis ke log.
- [ ] Error response tidak memuat stack trace atau payload sensitif.
- [ ] Log menyertakan request id untuk korelasi.
- [ ] Audit event kritis auth tercatat dengan metadata aman.

## 6. Dependency and Build Security

- [ ] Dependency scan dijalankan pada pipeline.
- [ ] Tidak ada temuan critical/high tanpa risk acceptance tertulis.
- [ ] Image/container tidak menyertakan secret.
- [ ] Semua dependency ter-pin dan lockfile tervalidasi.

## 7. Runtime and Operational Readiness

- [ ] `/health/live` dan `/health/ready` berfungsi sesuai kontrak.
- [ ] Secret runtime disuplai dari env/secret manager, bukan hardcoded.
- [ ] Runbook incident keamanan tersedia dan teruji tabletop.
- [ ] Monitoring error/rate-limit/auth-failure aktif.

## Sign-off

| Role              | Name | Date | Status |
| ----------------- | ---- | ---- | ------ |
| Engineering Lead  |      |      |        |
| Security Reviewer |      |      |        |
| Platform Reviewer |      |      |        |

## Related Documents

- [Security Operations Baseline](/Users/macbookpro/Development/recova-backend-v2/docs/operations/security.md)
- [Secure Coding Standard](/Users/macbookpro/Development/recova-backend-v2/docs/standards/secure-coding.md)
- [Release Gates](/Users/macbookpro/Development/recova-backend-v2/docs/operations/release-gates.md)

## Source Reference

- [Fiber Security Middleware Guide](https://docs.gofiber.io/blog/tags/cors)
- [OWASP API Security Top 10](https://owasp.org/API-Security/)
