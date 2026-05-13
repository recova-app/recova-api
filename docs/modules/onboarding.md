---
title: Onboarding Module
description: Kontrak onboarding pengguna untuk penyelesaian data awal, status completion, validasi input, dan ownership checks.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/modules/onboarding.md
last_reviewed: 2026-05-13
---

# Onboarding Module

Modul onboarding mengelola penyelesaian data awal pengguna setelah autentikasi berhasil.

## Responsibility

- menerima data onboarding awal,
- memvalidasi field onboarding,
- menyimpan status completion,
- memicu analisis AI onboarding dari jawaban user,
- menjaga idempotency saat onboarding dikirim ulang oleh client.

## Route Contract

Route onboarding saat ini berada pada:

```text
POST /api/v1/auth/onboarding
```

Owner domain tetap pada modul onboarding meskipun route berada di prefix auth.

## Input Contract

Field onboarding minimum:

- `nickname`,
- `recovery_reason`,
- `daily_checkin_time`,
- `porn_free_goal` (jumlah hari target bebas pornografi, contoh `1`, `2`, `7`).

Aturan:

- request harus datang dari user terautentikasi,
- semua field wajib divalidasi sebelum disimpan,
- data onboarding tidak boleh menerima `user_id` dari client sebagai sumber kebenaran.

Field `answers` dan `dependency_level` tetap opsional untuk kebutuhan analisis pola onboarding.

## Onboarding Completion State

State minimum:

- `PENDING`: user belum menyelesaikan onboarding,
- `COMPLETED`: data onboarding minimum sudah tervalidasi dan tersimpan.

Aturan transisi:

- `PENDING -> COMPLETED` terjadi setelah validasi sukses,
- `COMPLETED -> COMPLETED` pada submit ulang yang identik harus idempotent,
- perubahan data setelah completion diproses melalui aturan update profile.

## Ownership and Authorization

- principal token harus sama dengan owner profil yang diubah,
- user tidak boleh menyelesaikan onboarding untuk akun lain,
- kegagalan ownership harus ditangani dengan error aman.

## Data Quality Rules

- `nickname` harus unik atau memiliki strategi konflik yang terdokumentasi,
- `daily_checkin_time` harus lolos validasi format waktu,
- `recovery_reason` diperlakukan sebagai data pribadi dan tidak di-log mentah.

## Audit and Privacy Requirements

Audit event onboarding minimum:

- request id,
- user id,
- event type (`onboarding_completed` atau `onboarding_updated`),
- status hasil.

Batas privasi:

- jangan tulis `recovery_reason` mentah ke log,
- jangan expose field sensitif di error payload.

## Implementation Notes

- onboarding disimpan dengan membuat baris baru `profiles` untuk `user_id` terkait,
- onboarding memanggil analisis AI (`POST /api/v1/ai/onboarding-analysis`) secara internal dalam service flow yang sama,
- response onboarding mengembalikan payload user + `onboarding_analysis`,
- ringkasan teks AI onboarding dipersist ke `profiles.ai_summary`,
- submit onboarding ulang dengan payload identik diperlakukan idempotent (`COMPLETED -> COMPLETED`),
- submit onboarding ulang dengan payload berbeda dipetakan ke `409 CONFLICT`,
- perubahan data profil setelah onboarding selesai harus melalui endpoint users settings.

## Related Documents

- [Auth Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/auth.md)
- [Users Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/users.md)
- [Authentication and Trust Boundaries](/Users/macbookpro/Development/recova-backend-v2/docs/overview/authentication-and-trust-boundaries.md)

## Source Reference

- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
