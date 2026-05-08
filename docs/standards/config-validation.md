---
title: Configuration Validation Standard
description: Standar validasi konfigurasi environment agar startup fail-fast, bebas fallback tersembunyi, dan aman untuk operasi lintas environment.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/standards/config-validation.md
last_reviewed: 2026-05-08
---

# Configuration Validation Standard

Standar ini mengatur bagaimana konfigurasi environment divalidasi sebelum service menerima request.

## Validation Principles

- fail-fast: service berhenti bila konfigurasi tidak valid,
- explicitness: setiap env required harus dideklarasikan jelas,
- no hidden fallback: tidak boleh ada default tersembunyi untuk env required,
- safe logging: error konfigurasi tidak menampilkan nilai secret.

## Required Validation Rules

1. Env required wajib ada dan non-empty.
2. Enum env wajib hanya menerima nilai yang diizinkan.
3. Numeric env wajib lolos parsing dan rentang valid.
4. Duration/TTL env wajib lolos format yang disetujui (`time.ParseDuration` + dukungan `d` untuk hari bila dipakai pada kontrak env).
5. URL env wajib valid secara sintaks.
6. Secret env wajib panjang minimum sesuai kebijakan internal.

## No-Fallback Policy

Aturan ini wajib dipenuhi:

- dilarang memakai fallback `||`, `??`, atau default implicit untuk env required,
- dilarang memetakan env kosong menjadi nilai default,
- jika env optional tidak tersedia, nilainya harus benar-benar `missing`, bukan string kosong.

## Validation Execution Point

- validasi dijalankan saat bootstrap aplikasi,
- validasi dijalankan juga pada command penting yang bergantung env (misalnya migration/seeding),
- hasil validasi gagal harus menghentikan proses dengan pesan aman.

## Error Messaging Rules

Saat validasi gagal:

- tampilkan nama variabel yang invalid,
- tampilkan alasan validasi (missing/invalid format/out of range),
- jangan tampilkan nilai raw untuk secret.

Contoh aman:

```text
CONFIG_VALIDATION_ERROR: AI_API_KEY is required and must be non-empty.
```

## Security Checks

- rahasia seperti key/token/secret tidak boleh dicetak,
- daftar env yang bersifat secret harus dikelola eksplisit,
- pengujian konfigurasi wajib memastikan redaction aktif.

## CI and Deployment Checks

- pipeline harus memverifikasi bahwa env required tersedia,
- deploy harus gagal jika ada env required yang kosong,
- environment template (`.env.example`) harus sinkron dengan skema validasi.

## Verification Checklist

- skema validasi mencakup semua env pada dokumen environment,
- tidak ada fallback tersembunyi di bootstrap config,
- invalid config menghasilkan exit non-zero,
- error output tidak membocorkan secret.

## Related Documents

- [Environment Configuration](/Users/macbookpro/Development/recova-backend-v2/docs/environment.md)
- [Environment Matrix and Runtime Profiles](/Users/macbookpro/Development/recova-backend-v2/docs/operations/environments.md)
- [Review Process Standard](/Users/macbookpro/Development/recova-backend-v2/docs/standards/review-process.md)

## Source Reference

- [/Users/macbookpro/Development/bisakerja-api/docs/environment.md](/Users/macbookpro/Development/bisakerja-api/docs/environment.md)
- [Google OAuth 2.0 Web Server Flow](https://developers.google.com/identity/protocols/oauth2/web-server)
