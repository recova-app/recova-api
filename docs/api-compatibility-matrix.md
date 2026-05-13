---
title: Recova Backend API Compatibility Matrix
description: Status kompatibilitas endpoint publik dan aturan perubahan kontrak API untuk menjaga kestabilan integrasi klien.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/api-compatibility-matrix.md
last_reviewed: 2026-05-08
---

# Recova Backend API Compatibility Matrix

Dokumen ini menetapkan status kompatibilitas setiap endpoint publik.

## Status Definitions

| Status      | Arti                                                            |
| ----------- | --------------------------------------------------------------- |
| `preserve`  | endpoint dipertahankan pada path+method yang sama               |
| `rename`    | endpoint dipindah nama/path dengan compatibility plan           |
| `deprecate` | endpoint dipensiunkan dengan window kompatibilitas              |
| `redesign`  | kontrak endpoint dirancang ulang dan membutuhkan migration plan |
| `unknown`   | belum ada data cukup untuk memutuskan                           |

## Matrix

| Method   | Path                                  | Status      | Notes                                            |
| -------- | ------------------------------------- | ----------- | ------------------------------------------------ |
| `POST`   | `/api/v1/auth/google`                 | `preserve`  | kontrak login publik inti                        |
| `POST`   | `/api/v1/auth/onboarding`             | `preserve`  | onboarding flow inti                             |
| `GET`    | `/api/v1/users/me`                    | `preserve`  | profil pengguna aktif                            |
| `PUT`    | `/api/v1/users/settings`              | `preserve`  | pengaturan pengguna aktif                        |
| `DELETE` | `/api/v1/users/me/reset-data`         | `deprecate` | dev-only endpoint; tidak boleh aktif di produksi |
| `POST`   | `/api/v1/ai/ask-coach`                | `preserve`  | endpoint inti AI Coach                           |
| `GET`    | `/api/v1/ai/chat-history`             | `preserve`  | kontrak histori AI tetap                         |
| `GET`    | `/api/v1/ai/summary`                  | `preserve`  | ringkasan rutin terhubung AI                     |
| `POST`   | `/api/v1/ai/onboarding-analysis`      | `preserve`  | analisis onboarding AI                           |
| `POST`   | `/api/v1/routine/checkin`             | `preserve`  | kontrak check-in harian inti                     |
| `GET`    | `/api/v1/routine/statistics`          | `preserve`  | statistik pengguna inti                          |
| `GET`    | `/api/v1/routine/relapses`            | `preserve`  | histori relapse inti                             |
| `GET`    | `/api/v1/journals`                    | `preserve`  | jurnal pribadi inti                              |
| `POST`   | `/api/v1/journals`                    | `preserve`  | penulisan jurnal inti                            |
| `GET`    | `/api/v1/community`                   | `preserve`  | feed komunitas inti                              |
| `POST`   | `/api/v1/community`                   | `preserve`  | posting komunitas inti                           |
| `POST`   | `/api/v1/community/:post_id/comments` | `preserve`  | interaksi komentar inti                          |
| `POST`   | `/api/v1/community/:post_id/like`     | `preserve`  | interaksi like inti                              |
| `GET`    | `/api/v1/education`                   | `preserve`  | konten edukasi inti                              |
| `GET`    | `/api/v1/content/daily`               | `preserve`  | konten harian inti                               |

## Compatibility Guardrails

- path dan method endpoint `preserve` tidak boleh berubah.
- perubahan format response untuk endpoint `preserve` harus backward-compatible.
- endpoint `deprecate` wajib punya aturan disable yang aman untuk produksi.
- endpoint berstatus `unknown` tidak boleh diubah sebelum source kontrak cukup.

## Related Documents

- [API Reference](/Users/macbookpro/Development/recova-backend-v2/docs/api-reference.md)
- [API Versioning Standard](/Users/macbookpro/Development/recova-backend-v2/docs/standards/api-versioning.md)
- [Roadmap Compatibility Matrix](/Users/macbookpro/Development/recova-backend-v2/docs/roadmap/compatibility-matrix.md)

## Source Reference

- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
