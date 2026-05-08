---
title: Recova Backend Secure Coding Standard
description: Standar secure coding untuk handler, service, repository, dan integrasi eksternal agar perubahan kode tetap aman, terukur, dan dapat direview konsisten.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/standards/secure-coding.md
last_reviewed: 2026-05-08
---

# Recova Backend Secure Coding Standard

Standar ini menjadi acuan review kode untuk mencegah kerentanan yang umum pada layanan API.

## Core Principles

- deny-by-default untuk akses sensitif,
- validasi input sebelum logic domain,
- minimisasi data sensitif di seluruh alur,
- error aman untuk klien, detail teknis di log internal.

## Input Validation Rules

- validasi `params`, `query`, dan `body` wajib,
- tetapkan batas panjang string, ukuran array, dan nilai numerik,
- tolak payload dengan field tak dikenal untuk endpoint sensitif,
- normalisasi input penting sebelum dipakai untuk lookup.

## Authentication and Authorization Rules

- jangan percaya klaim auth dari client tanpa verifikasi signature/claims,
- ownership check dilakukan di backend,
- keputusan authz tidak boleh dipindahkan ke frontend,
- endpoint yang membuka data user lain harus punya policy eksplisit.

## Query and Persistence Rules

- gunakan query terparameter,
- dilarang interpolasi string mentah dari input user ke SQL,
- gunakan transaction untuk operasi multi-entity yang harus atomik,
- batasi kolom select untuk data sensitif.

## Secret and Sensitive Data Rules

- secret tidak boleh hardcoded,
- secret tidak boleh dicetak ke log,
- token yang dipersist harus berupa hash,
- field sensitif di response harus di-mask atau dihilangkan.

## Error Handling Rules

- gunakan error code stabil untuk kontrak API,
- response error tidak memuat stack trace,
- jangan expose error mentah database/downstream,
- sertakan request id pada error response untuk triage.

## External Integration Rules

- semua outbound call harus timeout-bounded,
- retry hanya untuk operasi idempotent,
- map error downstream ke taxonomy error internal,
- payload ke provider eksternal harus minimum-need.

## Logging Rules

- gunakan structured logging,
- wajib log `requestId`, route, status, latency,
- larang log raw request body untuk endpoint sensitif,
- redaksi otomatis harus mencakup token, password, API key, cookie, credential.

## Testing Rules for Security-Critical Changes

Perubahan pada auth, validation, ownership, limiter, atau token handling wajib menambah test untuk:

- success path,
- reject path,
- abuse path (rate-limit/invalid payload).

## Code Review Checklist

- [ ] validation lengkap,
- [ ] ownership check ada,
- [ ] query aman,
- [ ] logging aman,
- [ ] error mapping sesuai standar,
- [ ] test coverage perubahan keamanan memadai.

## Related Documents

- [Security Operations Baseline](/Users/macbookpro/Development/recova-backend-v2/docs/operations/security.md)
- [Error Taxonomy](/Users/macbookpro/Development/recova-backend-v2/docs/standards/error-taxonomy.md)
- [API Response Standard](/Users/macbookpro/Development/recova-backend-v2/docs/api-response-standard.md)

## Source Reference

- [Go Security Best Practices](https://go.dev/doc/security/)
- [GORM Security Guide](https://gorm.io/docs/security.html)
- [Fiber Error Handling Guide](https://docs.gofiber.io/guide/error-handling)
