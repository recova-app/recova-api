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
generated_by: cmd-tools-openapi
generated_at: 2026-05-08T00:00:00Z
---

# Recova Backend Route Inventory

Dokumen ini adalah inventaris route aktif berdasarkan runtime Go Fiber saat ini.

## Summary

| Metric        | Value        |
| ------------- | ------------ |
| Total routes  | 9            |
| API prefix    | `/api/v1`    |
| Last verified | `2026-05-08` |

## Registered Routes

| Method   | Path                          | Module   |
| -------- | ----------------------------- | -------- |
| `DELETE` | `/api/v1/users/me/reset-data` | `api-v1` |
| `GET`    | `/api/v1/users/me`            | `api-v1` |
| `GET`    | `/health/live`                | `health` |
| `GET`    | `/health/ready`               | `health` |
| `POST`   | `/api/v1/auth/google`         | `api-v1` |
| `POST`   | `/api/v1/auth/logout`         | `api-v1` |
| `POST`   | `/api/v1/auth/onboarding`     | `api-v1` |
| `POST`   | `/api/v1/auth/refresh`        | `api-v1` |
| `PUT`    | `/api/v1/users/settings`      | `api-v1` |

## Drift Check Use

Gunakan file ini untuk validasi sinkronisasi route runtime dan kontrak OpenAPI pada proses review maupun CI.

## Known Gap

Inventaris route ini disinkronkan otomatis dari runtime. Perbedaan terhadap kontrak OpenAPI diperlakukan sebagai drift dan harus diperbaiki sebelum merge.

## Related Documents

- [OpenAPI Standard](/Users/macbookpro/Development/recova-backend-v2/docs/standards/openapi.md)
- [API Reference](/Users/macbookpro/Development/recova-backend-v2/docs/api-reference.md)
- [API Docs Generation](/Users/macbookpro/Development/recova-backend-v2/docs/operations/api-docs-generation.md)

## Source Reference

- [Fiber App API `GetRoutes`](https://docs.gofiber.io/next/api/app/)
- [OpenAPI Specification](https://spec.openapis.org/oas/latest)
