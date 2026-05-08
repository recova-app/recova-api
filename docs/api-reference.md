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
last_reviewed: 2026-05-08
---

# Recova Backend API Reference

Dokumen ini adalah inventaris endpoint publik berdasarkan sumber yang tersedia saat ini.

## Base Path

Semua endpoint produk berada di bawah:

```text
/api/v1
```

## Endpoint Inventory

| Domain    | Method   | Path                          | Summary                                       | Auth requirement  | Contract status |
| --------- | -------- | ----------------------------- | --------------------------------------------- | ----------------- | --------------- |
| auth      | `POST`   | `/auth/google`                | login/registrasi via Google token             | gap               | documented      |
| auth      | `POST`   | `/auth/onboarding`            | simpan onboarding awal pengguna               | gap               | documented      |
| users     | `GET`    | `/users/me`                   | ambil profil pengguna                         | gap               | documented      |
| users     | `PUT`    | `/users/settings`             | update pengaturan profil                      | gap               | documented      |
| users     | `DELETE` | `/users/me/reset-data`        | reset data pengguna untuk development/testing | dev-only (stated) | documented      |
| ai        | `POST`   | `/ai/ask-coach`               | kirim pesan ke AI Coach                       | gap               | documented      |
| ai        | `GET`    | `/ai/chat-history`            | ambil riwayat chat AI Coach                   | gap               | documented      |
| ai        | `GET`    | `/ai/summary`                 | ambil ringkasan check-in                      | gap               | documented      |
| ai        | `POST`   | `/ai/onboarding-analysis`     | analisis data onboarding                      | gap               | documented      |
| routine   | `POST`   | `/routine/checkin`            | check-in harian                               | gap               | documented      |
| routine   | `GET`    | `/routine/statistics`         | ambil statistik pengguna                      | gap               | documented      |
| routine   | `GET`    | `/routine/relapses`           | ambil riwayat relapse                         | gap               | documented      |
| journals  | `GET`    | `/journals`                   | ambil daftar jurnal pribadi                   | gap               | documented      |
| journals  | `POST`   | `/journals`                   | buat entri jurnal                             | gap               | documented      |
| community | `GET`    | `/community`                  | ambil daftar postingan komunitas              | gap               | documented      |
| community | `POST`   | `/community`                  | buat postingan komunitas                      | gap               | documented      |
| community | `POST`   | `/community/:postId/comments` | tambah komentar pada postingan                | gap               | documented      |
| community | `POST`   | `/community/:postId/like`     | like postingan                                | gap               | documented      |
| education | `GET`    | `/education`                  | ambil daftar konten edukasi                   | gap               | documented      |
| content   | `GET`    | `/content/daily`              | ambil konten harian                           | gap               | documented      |

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
