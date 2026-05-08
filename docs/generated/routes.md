---
title: Recova Backend Route Inventory
description: Runtime API route inventory used for contract coverage verification and documentation drift detection.
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

This document lists active routes from the current Go Fiber runtime.

## Summary

| Metric        | Value        |
| ------------- | ------------ |
| Total routes  | 14           |
| API prefix    | `/api/v1`    |
| Last verified | `2026-05-08` |

## Registered Routes

| Method   | Path                          | Module   |
| -------- | ----------------------------- | -------- |
| `DELETE` | `/api/v1/users/me/reset-data` | `api-v1` |
| `GET`    | `/api/v1/journals`            | `api-v1` |
| `GET`    | `/api/v1/routine/relapses`    | `api-v1` |
| `GET`    | `/api/v1/routine/statistics`  | `api-v1` |
| `GET`    | `/api/v1/users/me`            | `api-v1` |
| `GET`    | `/health/live`                | `health` |
| `GET`    | `/health/ready`               | `health` |
| `POST`   | `/api/v1/auth/google`         | `api-v1` |
| `POST`   | `/api/v1/auth/logout`         | `api-v1` |
| `POST`   | `/api/v1/auth/onboarding`     | `api-v1` |
| `POST`   | `/api/v1/auth/refresh`        | `api-v1` |
| `POST`   | `/api/v1/journals`            | `api-v1` |
| `POST`   | `/api/v1/routine/checkin`     | `api-v1` |
| `PUT`    | `/api/v1/users/settings`      | `api-v1` |

## Drift Check Use

Use this file to validate runtime route and OpenAPI contract synchronization in review and CI flows.

## Known Gap

This route inventory is auto-synced from runtime. Any mismatch with the OpenAPI contract is treated as drift and must be fixed before merge.

## Related Documents

- [OpenAPI Standard](/Users/macbookpro/Development/recova-backend-v2/docs/standards/openapi.md)
- [API Reference](/Users/macbookpro/Development/recova-backend-v2/docs/api-reference.md)
- [API Docs Generation](/Users/macbookpro/Development/recova-backend-v2/docs/operations/api-docs-generation.md)

## Source Reference

- [Fiber App API `GetRoutes`](https://docs.gofiber.io/next/api/app/)
- [OpenAPI Specification](https://spec.openapis.org/oas/latest)
