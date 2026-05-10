---
title: Recova Backend Compatibility Test Plan
description: Rencana verifikasi kompatibilitas kontrak API agar perilaku endpoint prioritas tetap konsisten selama transisi runtime layanan.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/roadmap/compatibility-test-plan.md
last_reviewed: 2026-05-09
---

# Recova Backend Compatibility Test Plan

Dokumen ini memandu pengujian kompatibilitas kontrak API publik.

## Objectives

- menjaga kompatibilitas endpoint prioritas,
- mendeteksi drift response lebih awal,
- mengurangi regresi frontend akibat perubahan backend.

## Scope Endpoint Prioritas

- `/api/v1/auth/*`
- `/api/v1/users/*`
- `/api/v1/routine/*`
- `/api/v1/journals/*`
- `/api/v1/community/*`
- `/api/v1/education/*`
- `/api/v1/content/daily`
- `/api/v1/ai/*`
- `/api/v1/achievements/*`
- `/health/live`, `/health/ready`

## Compatibility Assertions

Setiap endpoint prioritas diuji untuk:

- method + path contract,
- kebutuhan auth,
- status code success/error,
- response envelope shape,
- error code mapping utama,
- pagination metadata (untuk list endpoint).
- additive fields tidak merusak parser klien lama.

Assertion tambahan untuk enhancement:

- statistik lanjutan: field existing tetap ada, field baru additive,
- threaded comments: depth/reply_count/parent_comment_id konsisten,
- achievements: katalog dan progress memakai code stabil,
- AI persona preference: whitelist persona + fallback default aman.

## Test Strategy

1. buat fixture request/response per endpoint prioritas,
2. jalankan contract tests otomatis pada CI,
3. bandingkan hasil aktual vs expected contract,
4. blok release jika ada drift tidak disetujui.

## Drift Policy

- drift yang tidak disengaja: harus diperbaiki sebelum release,
- drift yang disengaja: wajib update dokumen API + announce breaking/non-breaking status,
- breaking change: butuh sign-off lintas reviewer.

## Success Criteria

- seluruh contract tests endpoint prioritas lulus,
- tidak ada perubahan shape envelope tanpa approval,
- perubahan error code terjustifikasi dan terdokumentasi.

## Related Documents

- [API Reference](/Users/macbookpro/Development/recova-backend-v2/docs/api-reference.md)
- [API Response Standard](/Users/macbookpro/Development/recova-backend-v2/docs/api-response-standard.md)
- [Testing Strategy](/Users/macbookpro/Development/recova-backend-v2/docs/operations/testing.md)

## Source Reference

- [Current Runtime Inventory](/Users/macbookpro/Development/recova-backend-v2/docs/roadmap/current-runtime-inventory.md)
- [Express Baseline](/Users/macbookpro/Development/recova-backend-v2/docs/roadmap/express-baseline.md)
- [Feature Inventory](/Users/macbookpro/Development/recova-backend-v2/docs/roadmap/feature-inventory.md)
