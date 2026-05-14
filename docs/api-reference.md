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
last_reviewed: 2026-05-15
---

# Recova Backend API Reference

Dokumen ini adalah inventaris endpoint publik berdasarkan sumber yang tersedia saat ini.

## Base Path

Semua endpoint produk berada di bawah:

```text
/api/v1
```

## OpenAPI And Scalar Surfaces

Runtime dan repository sekarang memakai satu artefak OpenAPI generated yang sama.

| Surface                  | Path/Route                    | Purpose                                                                       |
| ------------------------ | ----------------------------- | ----------------------------------------------------------------------------- |
| Generated OpenAPI        | `docs/generated/openapi.yaml` | kontrak machine-readable kanonik untuk review, validasi drift, dan tooling    |
| Runtime OpenAPI endpoint | `/openapi.yaml`               | endpoint raw OpenAPI YAML untuk konsumsi tooling/browser pada service runtime |
| Runtime Scalar reference | `/docs/api`                   | dokumentasi API interaktif berbasis Scalar API Reference                      |
| Scalar Docs config       | `scalar.config.json`          | peta navigasi repo docs + OpenAPI untuk preview/publish Scalar Docs           |

Halaman `/docs/api` merender Scalar dari CDN resmi dan memuat kontrak dari `/openapi.yaml`.
Tidak ada prefilled credentials atau token pada konfigurasi UI.

## Client Contract Context

API saat ini dikonsumsi oleh aplikasi mobile Flutter, sehingga kontrak endpoint harus menjaga:

- stabilitas nama field JSON agar parsing model Flutter tidak mudah rusak,
- kompatibilitas perubahan (tambahan field optional diperbolehkan, penghapusan field wajib dihindari),
- konsistensi envelope success/error untuk menyederhanakan handler klien,
- message user-facing berbahasa Indonesia.

## Endpoint Inventory

| Domain       | Method   | Path                                                 | Summary                                       | Auth requirement  | Contract status |
| ------------ | -------- | ---------------------------------------------------- | --------------------------------------------- | ----------------- | --------------- |
| platform     | `GET`    | `/health/live`                                       | cek status liveness service                   | public/internal   | implemented     |
| platform     | `GET`    | `/health/ready`                                      | cek status readiness dependency               | public/internal   | implemented     |
| platform     | `GET`    | `/metrics`                                           | expose metrik prometheus                      | public/internal   | implemented     |
| platform     | `GET`    | `/openapi.yaml`                                      | ambil OpenAPI YAML runtime                    | public            | implemented     |
| platform     | `GET`    | `/docs/api`                                          | tampilkan API reference interaktif            | public            | implemented     |
| auth         | `POST`   | `/auth/google`                                       | login/registrasi via Google token             | public            | implemented     |
| auth         | `POST`   | `/auth/register`                                     | registrasi akun manual                        | public            | implemented     |
| auth         | `POST`   | `/auth/login`                                        | login akun manual                             | public            | implemented     |
| auth         | `POST`   | `/auth/onboarding`                                   | simpan onboarding awal + analisis AI response | bearer            | implemented     |
| auth         | `POST`   | `/auth/refresh`                                      | rotasi refresh token dan perbarui sesi akses  | cookie            | implemented     |
| auth         | `POST`   | `/auth/logout`                                       | akhiri sesi aktif pengguna                    | bearer            | implemented     |
| users        | `GET`    | `/users/me`                                          | ambil profil pengguna                         | bearer            | implemented     |
| users        | `PUT`    | `/users/settings`                                    | update pengaturan profil                      | bearer            | implemented     |
| users        | `DELETE` | `/users/me/reset-data`                               | reset data pengguna untuk development/testing | bearer + dev-only | implemented     |
| ai           | `POST`   | `/ai/ask-coach`                                      | kirim pesan ke AI Coach                       | bearer            | implemented     |
| ai           | `POST`   | `/ai/relapse-solution`                               | analisis trigger relapse + solusi terbaik     | bearer            | implemented     |
| ai           | `GET`    | `/ai/chat-history`                                   | ambil riwayat chat AI Coach                   | bearer            | implemented     |
| ai           | `GET`    | `/ai/summary`                                        | ambil ringkasan check-in                      | bearer            | implemented     |
| ai           | `POST`   | `/ai/onboarding-analysis`                            | analisis data onboarding                      | bearer            | implemented     |
| ai           | `GET`    | `/ai/persona-preferences`                            | ambil preferensi persona AI                   | bearer            | implemented     |
| ai           | `PUT`    | `/ai/persona-preferences`                            | perbarui preferensi persona AI                | bearer            | implemented     |
| routine      | `POST`   | `/routine/checkin`                                   | check-in harian                               | bearer            | implemented     |
| routine      | `POST`   | `/routine/relapses`                                  | catat relapse harian                          | bearer            | implemented     |
| routine      | `GET`    | `/routine/statistics`                                | ambil statistik pengguna                      | bearer            | implemented     |
| routine      | `GET`    | `/routine/statistics/activity-summary`               | ambil ringkasan aktivitas periodik            | bearer            | implemented     |
| routine      | `GET`    | `/routine/relapses`                                  | ambil riwayat relapse                         | bearer            | implemented     |
| routine      | `GET`    | `/routine/relapses/statistics`                       | ambil statistik relapse lengkap               | bearer            | implemented     |
| achievements | `GET`    | `/achievements/catalog`                              | ambil katalog achievement aktif               | bearer            | implemented     |
| achievements | `GET`    | `/achievements/progress`                             | ambil progres achievement user                | bearer            | implemented     |
| achievements | `GET`    | `/achievements/unlocked`                             | ambil daftar achievement yang sudah terbuka   | bearer            | implemented     |
| journals     | `GET`    | `/journals`                                          | ambil daftar jurnal pribadi                   | bearer            | implemented     |
| journals     | `POST`   | `/journals`                                          | buat entri jurnal                             | bearer            | implemented     |
| community    | `GET`    | `/community`                                         | ambil daftar postingan komunitas              | bearer            | implemented     |
| community    | `POST`   | `/community`                                         | buat postingan komunitas                      | bearer            | implemented     |
| community    | `GET`    | `/community/{post_id}/comments`                      | ambil thread komentar                         | bearer            | implemented     |
| community    | `POST`   | `/community/{post_id}/comments`                      | tambah komentar pada postingan                | bearer            | implemented     |
| community    | `POST`   | `/community/{post_id}/comments/{comment_id}/replies` | tambah balasan komentar                       | bearer            | implemented     |
| community    | `POST`   | `/community/{post_id}/like`                          | toggle suka postingan                         | bearer            | implemented     |
| education    | `GET`    | `/education`                                         | ambil daftar konten edukasi                   | bearer            | implemented     |
| content      | `GET`    | `/content/daily`                                     | ambil konten harian                           | bearer            | implemented     |

## Contract Fields Coverage

| Contract field          | Status  | Notes                                                                       |
| ----------------------- | ------- | --------------------------------------------------------------------------- |
| Method                  | covered | tersedia di sumber                                                          |
| Path                    | covered | tersedia di sumber                                                          |
| Summary                 | covered | tersedia di sumber                                                          |
| Auth requirement        | covered | terspesifikasi di `docs/generated/openapi.yaml`                             |
| Request body schema     | covered | terspesifikasi per endpoint di OpenAPI                                      |
| Query schema            | covered | terspesifikasi per endpoint di OpenAPI                                      |
| Success response schema | covered | tervalidasi contract test + OpenAPI                                         |
| Error response schema   | covered | tervalidasi contract test + OpenAPI                                         |
| Platform routes         | covered | `/health/*`, `/metrics`, `/openapi.yaml`, `/docs/api` sudah diinventarisasi |

## Education Response Notes

- response `GET /api/v1/education` mengandung field `type` pada setiap item konten,
- nilai `type` dibatasi ke `artikel` atau `video`,
- label kategori user-facing pada payload menggunakan format natural (spasi), bukan snake_case.

## Gap Register

Tidak ada gap terbuka untuk kontrak endpoint utama per audit 2026-05-15.
Detail query/response dan status code kanonik berada di `docs/generated/openapi.yaml` dan tervalidasi oleh `test/contract/openapi_contract_test.go`.

## Related Documents

- [API Compatibility Matrix](/Users/macbookpro/Development/recova-backend-v2/docs/api-compatibility-matrix.md)
- [API Versioning Standard](/Users/macbookpro/Development/recova-backend-v2/docs/standards/api-versioning.md)
- [API Response Standard](/Users/macbookpro/Development/recova-backend-v2/docs/api-response-standard.md)
- [API Docs Generation](/Users/macbookpro/Development/recova-backend-v2/docs/operations/api-docs-generation.md)
- [Flow and Endpoint Audit (2026-05-14)](/Users/macbookpro/Development/recova-backend-v2/docs/operations/flow-and-endpoint-audit-2026-05-14.md)

## Source Reference

- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
- [/Users/macbookpro/Development/bisakerja-api/docs/api-reference.md](/Users/macbookpro/Development/bisakerja-api/docs/api-reference.md)
- [Scalar API Reference Getting Started](https://scalar.com/products/api-references/getting-started)
- [Scalar Docs Navigation](https://scalar.com/products/docs/configuration/navigation)
