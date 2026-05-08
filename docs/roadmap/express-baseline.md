---
title: Recova Backend Current Express Baseline
description: Baseline perilaku backend saat ini berdasarkan sumber yang tersedia untuk menjaga konsistensi kontrak sebelum perubahan lanjutan.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/roadmap/express-baseline.md
last_reviewed: 2026-05-08
---

# Recova Backend Current Express Baseline

Dokumen ini merangkum perilaku layanan backend saat ini berdasarkan sumber yang telah tersedia. Tujuannya adalah menyediakan baseline kontrak teknis agar perubahan berikutnya tetap dapat ditelusuri dan diverifikasi.

## Source Coverage

| Source                                                                                       | Cakupan                                                                                               |
| -------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------- |
| [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md) | Deskripsi fitur, endpoint publik, variabel environment, alur runtime, Docker, dan script operasional. |

Semua isi pada halaman ini ditarik dari sumber di atas. Jika detail belum ada di sumber, bagian tersebut dicatat sebagai gap.

## Service Runtime Baseline

- Runtime aplikasi menggunakan Node.js.
- Framework HTTP menggunakan Express.js.
- Bahasa utama menggunakan TypeScript.
- Persistence utama menggunakan PostgreSQL.
- ORM yang didokumentasikan adalah Prisma.
- Prefix endpoint publik berada di `/api/v1`.

## Public Capability Baseline

Capability aktif yang sudah dideklarasikan:

- autentikasi pengguna berbasis Google OAuth dan JWT,
- manajemen profil dan onboarding,
- check-in harian dan pelacakan streak,
- jurnal pribadi,
- statistik pengguna,
- komunitas (post, komentar, like),
- konten edukasi,
- AI Coach,
- konten harian (motivasi dan tantangan).

## API Surface Baseline

Seluruh endpoint publik berada di bawah prefix `/api/v1` dengan grup berikut:

- `/auth`
- `/users`
- `/ai`
- `/routine`
- `/journals`
- `/community`
- `/education`
- `/content`

Rincian endpoint per grup dicatat pada [Feature Inventory](/Users/macbookpro/Development/recova-backend-v2/docs/roadmap/feature-inventory.md).

## Runtime Operation Baseline

Perilaku runtime yang sudah terdokumentasi:

- mode development lokal menjalankan `npm run dev`,
- mode production menjalankan build lalu start,
- tersedia workflow migrasi database,
- tersedia workflow seeding data,
- tersedia workflow container untuk development dan production.

Rincian kontrak runtime dicatat pada [Current Runtime Inventory](/Users/macbookpro/Development/recova-backend-v2/docs/roadmap/current-runtime-inventory.md).

## Data and Integration Baseline

- Integrasi autentikasi eksternal: Google OAuth.
- Integrasi AI: Google Gemini sebagai layanan utama dan OpenAI sebagai alternatif opsional.
- Kontrak koneksi database menggunakan `DATABASE_URL` untuk PostgreSQL.

## Known Gaps

Area berikut belum dapat dipastikan dari sumber saat ini dan memerlukan source tambahan:

- daftar status code dan kontrak error per endpoint,
- skema request/response rinci per endpoint,
- detail middleware aktif (authz, validation, rate limit, logging),
- aturan session/token lifecycle rinci,
- kebijakan observability operasional (metrics, tracing, structured log schema),
- kontrak deployment non-Docker (jika ada),
- detail non-functional requirement (SLO, timeout, retry policy).

## Change Control Rule

Saat sumber utama berubah, halaman ini wajib diperbarui bersama dokumen inventaris fitur dan runtime agar baseline tetap sinkron.

## Related Documents

- [Recova Backend Documentation Overview](/Users/macbookpro/Development/recova-backend-v2/docs/overview.md)
- [Feature Inventory](/Users/macbookpro/Development/recova-backend-v2/docs/roadmap/feature-inventory.md)
- [Current Runtime Inventory](/Users/macbookpro/Development/recova-backend-v2/docs/roadmap/current-runtime-inventory.md)

## Source Reference

- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
