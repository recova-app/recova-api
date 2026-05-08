---
title: Recova Backend Route Inventory
description: Inventaris route API Recova Backend untuk verifikasi coverage kontrak dan deteksi drift dokumentasi.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/generated/routes.md
last_reviewed: 2026-05-08
generated_by: manual-route-inventory
generated_at: 2026-05-08T00:00:00Z
source_commit: pending
---

# Recova Backend Route Inventory

Dokumen ini adalah inventaris route aktif berdasarkan sumber referensi layanan saat ini.

## Summary

| Metric        | Value        |
| ------------- | ------------ |
| Total routes  | 24           |
| API prefix    | `/api/v1`    |
| Last verified | `2026-05-08` |

## Registered Routes

| Method   | Path                                 | Module          |
| -------- | ------------------------------------ | --------------- |
| `GET`    | `/health/live`                       | `health`        |
| `GET`    | `/health/ready`                      | `health`        |
| `POST`   | `/api/v1/auth/google`                | `auth`          |
| `POST`   | `/api/v1/auth/onboarding`            | `auth`          |
| `POST`   | `/api/v1/auth/refresh`               | `auth`          |
| `POST`   | `/api/v1/auth/logout`                | `auth`          |
| `GET`    | `/api/v1/users/me`                   | `users`         |
| `PUT`    | `/api/v1/users/settings`             | `users`         |
| `DELETE` | `/api/v1/users/me/reset-data`        | `users`         |
| `POST`   | `/api/v1/ai/ask-coach`               | `ai`            |
| `GET`    | `/api/v1/ai/chat-history`            | `ai`            |
| `GET`    | `/api/v1/ai/summary`                 | `ai`            |
| `POST`   | `/api/v1/ai/onboarding-analysis`     | `ai`            |
| `POST`   | `/api/v1/routine/checkin`            | `routine`       |
| `GET`    | `/api/v1/routine/statistics`         | `routine`       |
| `GET`    | `/api/v1/routine/relapses`           | `routine`       |
| `GET`    | `/api/v1/journals`                   | `journals`      |
| `POST`   | `/api/v1/journals`                   | `journals`      |
| `GET`    | `/api/v1/community`                  | `community`     |
| `POST`   | `/api/v1/community`                  | `community`     |
| `POST`   | `/api/v1/community/:postId/comments` | `community`     |
| `POST`   | `/api/v1/community/:postId/like`     | `community`     |
| `GET`    | `/api/v1/education`                  | `education`     |
| `GET`    | `/api/v1/content/daily`              | `daily-content` |

## Drift Check Use

Gunakan file ini untuk:

- cek cepat perubahan route antar-commit,
- sinkronisasi dengan OpenAPI,
- validasi route coverage pada review dokumentasi.

## Known Gap

Inventaris ini masih berbasis referensi dokumentasi layanan. Setelah route registry Go Fiber tersedia, artefak ini harus digenerate otomatis dari source runtime.

## Related Documents

- [OpenAPI Standard](/Users/macbookpro/Development/recova-backend-v2/docs/standards/openapi.md)
- [API Reference](/Users/macbookpro/Development/recova-backend-v2/docs/api-reference.md)
- [API Docs Generation](/Users/macbookpro/Development/recova-backend-v2/docs/operations/api-docs-generation.md)

## Source Reference

- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
- [/Users/macbookpro/Development/bisakerja-api/docs/generated/routes.md](/Users/macbookpro/Development/bisakerja-api/docs/generated/routes.md)
