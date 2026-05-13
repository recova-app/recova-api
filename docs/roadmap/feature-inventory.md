---
title: Recova Backend Feature Inventory
description: Inventaris capability backend saat ini beserta endpoint publik yang telah terdokumentasi untuk menjaga konsistensi kontrak layanan.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/roadmap/feature-inventory.md
last_reviewed: 2026-05-08
---

# Recova Backend Feature Inventory

Dokumen ini berisi inventaris fitur backend yang saat ini sudah dinyatakan aktif beserta endpoint publik yang tersedia dari sumber resmi layanan.

## Feature Matrix

| Domain fitur                  | Capability utama                                                                             | Endpoint group      | Endpoint publik yang tersedia                                                       |
| ----------------------------- | -------------------------------------------------------------------------------------------- | ------------------- | ----------------------------------------------------------------------------------- |
| Authentication                | Login/registrasi dengan Google token dan penyelesaian onboarding awal pengguna               | `/api/v1/auth`      | `POST /google`, `POST /onboarding`                                                  |
| User Profile                  | Ambil profil pengguna, ubah pengaturan profil, reset data pengguna untuk konteks development | `/api/v1/users`     | `GET /me`, `PUT /settings`, `DELETE /me/reset-data`                                 |
| AI Coach                      | Interaksi tanya-jawab dengan coach, riwayat chat, ringkasan check-in, analisis onboarding    | `/api/v1/ai`        | `POST /ask-coach`, `GET /chat-history`, `GET /summary`, `POST /onboarding-analysis` |
| Routine and Recovery Progress | Check-in harian, statistik progres, riwayat relapse                                          | `/api/v1/routine`   | `POST /checkin`, `GET /statistics`, `GET /relapses`                                 |
| Journaling                    | Pengambilan dan pembuatan entri jurnal pengguna                                              | `/api/v1/journals`  | `GET /`, `POST /`                                                                   |
| Community                     | Daftar dan pembuatan posting komunitas, komentar, dan like posting                           | `/api/v1/community` | `GET /`, `POST /`, `POST /:post_id/comments`, `POST /:post_id/like`                 |
| Education                     | Pengambilan konten edukasi                                                                   | `/api/v1/education` | `GET /`                                                                             |
| Daily Content                 | Pengambilan konten harian                                                                    | `/api/v1/content`   | `GET /daily`                                                                        |

## Service Capability Coverage

Daftar capability tingkat produk yang telah dinyatakan tersedia:

- autentikasi pengguna,
- manajemen profil,
- check-in harian,
- pelacakan streak,
- jurnal pribadi,
- statistik pengguna,
- komunitas,
- konten edukasi,
- AI Coach,
- konten harian.

## Data Initialization Coverage

Sumber runtime saat ini menyebut data seeding untuk domain berikut:

- users,
- profiles,
- streaks,
- check-ins,
- journals,
- community posts/comments,
- education content,
- daily motivations,
- daily challenges.

## Compatibility Notes

- Prefix endpoint publik saat ini adalah `/api/v1`.
- Tidak ada versi endpoint lain yang dinyatakan pada sumber saat ini.
- Tidak ada kontrak deprecation endpoint yang dinyatakan pada sumber saat ini.

## Known Gaps

- tidak ada definisi skema request/response per endpoint,
- tidak ada daftar status code dan error code per endpoint,
- tidak ada aturan idempotency atau pagination contract,
- tidak ada dokumentasi authorization scope per endpoint,
- tidak ada dokumentasi rate limit per endpoint.

## Related Documents

- [Current Express Baseline](/Users/macbookpro/Development/recova-backend-v2/docs/roadmap/express-baseline.md)
- [Current Runtime Inventory](/Users/macbookpro/Development/recova-backend-v2/docs/roadmap/current-runtime-inventory.md)
- [Recova Backend Documentation Overview](/Users/macbookpro/Development/recova-backend-v2/docs/overview.md)

## Source Reference

- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
