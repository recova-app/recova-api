---
title: Recova Backend API Reference
description: Inventaris endpoint publik Recova Backend, klasifikasi kontrak endpoint, dan daftar gap detail request-response yang masih perlu dilengkapi.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/api-reference.md
last_reviewed: 2026-05-09
---

# Recova Backend API Reference

Dokumen ini adalah inventaris endpoint publik berdasarkan sumber yang tersedia saat ini.

## Base Path

Semua endpoint produk berada di bawah:

```text
/api/v1
```

## Client Contract Context

API saat ini dikonsumsi oleh aplikasi mobile Flutter, sehingga kontrak endpoint harus menjaga:

- stabilitas nama field JSON agar parsing model Flutter tidak mudah rusak,
- kompatibilitas perubahan (tambahan field optional diperbolehkan, penghapusan field wajib dihindari),
- konsistensi envelope success/error untuk menyederhanakan handler klien,
- message user-facing berbahasa Indonesia.

## Endpoint Inventory

| Domain    | Method   | Path                          | Summary                                       | Auth requirement  | Contract status |
| --------- | -------- | ----------------------------- | --------------------------------------------- | ----------------- | --------------- |
| auth      | `POST`   | `/auth/google`                | login/registrasi via Google token             | public            | implemented     |
| auth      | `POST`   | `/auth/onboarding`            | simpan onboarding awal pengguna               | bearer            | implemented     |
| auth      | `POST`   | `/auth/refresh`               | rotasi refresh token dan perbarui sesi akses  | cookie            | implemented     |
| auth      | `POST`   | `/auth/logout`                | akhiri sesi aktif pengguna                    | bearer            | implemented     |
| users     | `GET`    | `/users/me`                   | ambil profil pengguna                         | bearer            | implemented     |
| users     | `PUT`    | `/users/settings`             | update pengaturan profil                      | bearer            | implemented     |
| users     | `DELETE` | `/users/me/reset-data`        | reset data pengguna untuk development/testing | bearer + dev-only | implemented     |
| ai        | `POST`   | `/ai/ask-coach`               | kirim pesan ke AI Coach                       | bearer            | implemented     |
| ai        | `GET`    | `/ai/chat-history`            | ambil riwayat chat AI Coach                   | bearer            | implemented     |
| ai        | `GET`    | `/ai/summary`                 | ambil ringkasan check-in                      | bearer            | implemented     |
| ai        | `POST`   | `/ai/onboarding-analysis`     | analisis data onboarding                      | bearer            | implemented     |
| ai        | `GET`    | `/ai/persona-preferences`     | ambil preferensi persona AI                   | bearer            | implemented     |
| ai        | `PUT`    | `/ai/persona-preferences`     | perbarui preferensi persona AI                | bearer            | implemented     |
| routine   | `POST`   | `/routine/checkin`            | check-in harian                               | bearer            | implemented     |
| routine   | `GET`    | `/routine/statistics`         | ambil statistik pengguna                      | bearer            | implemented     |
| routine   | `GET`    | `/routine/relapses`           | ambil riwayat relapse                         | bearer            | implemented     |
| journals  | `GET`    | `/journals`                   | ambil daftar jurnal pribadi                   | bearer            | implemented     |
| journals  | `POST`   | `/journals`                   | buat entri jurnal                             | bearer            | implemented     |
| community | `GET`    | `/community`                  | ambil daftar postingan komunitas              | bearer            | implemented     |
| community | `POST`   | `/community`                  | buat postingan komunitas                      | bearer            | implemented     |
| community | `POST`   | `/community/:postId/comments` | tambah komentar pada postingan                | bearer            | implemented     |
| community | `POST`   | `/community/:postId/like`     | toggle suka postingan                         | bearer            | implemented     |
| education | `GET`    | `/education`                  | ambil daftar konten edukasi                   | bearer            | implemented     |
| content   | `GET`    | `/content/daily`              | ambil konten harian                           | bearer            | implemented     |

## Contract Fields Coverage

| Contract field          | Status  | Notes                                   |
| ----------------------- | ------- | --------------------------------------- |
| Method                  | covered | tersedia di sumber                      |
| Path                    | covered | tersedia di sumber                      |
| Summary                 | covered | tersedia di sumber                      |
| Auth requirement        | gap     | belum ada source eksplisit per endpoint |
| Request body schema     | gap     | belum ada source schema                 |
| Query schema            | gap     | belum ada source schema                 |
| Success response schema | gap     | belum ada source schema                 |
| Error response schema   | gap     | belum ada source schema                 |

## Gap Register

Detail berikut masih perlu source tambahan:

- daftar status code per endpoint,
- struktur request body/query per endpoint,
- struktur response sukses per endpoint,
- struktur error per endpoint,
- auth/authorization rule per endpoint.

## Related Documents

- [API Compatibility Matrix](/Users/macbookpro/Development/recova-backend-v2/docs/api-compatibility-matrix.md)
- [API Versioning Standard](/Users/macbookpro/Development/recova-backend-v2/docs/standards/api-versioning.md)
- [API Response Standard](/Users/macbookpro/Development/recova-backend-v2/docs/api-response-standard.md)

## Source Reference

- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
- [/Users/macbookpro/Development/bisakerja-api/docs/api-reference.md](/Users/macbookpro/Development/bisakerja-api/docs/api-reference.md)
