---
title: Community Module
description: Kontrak modul komunitas untuk posting, komentar, dan interaksi like dengan kontrol ownership, moderasi, dan anti-abuse.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/modules/community.md
last_reviewed: 2026-05-08
---

# Community Module

## Responsibility

- menyediakan feed komunitas,
- membuat post,
- menambahkan komentar,
- mengelola state like/unlike secara konsisten,
- menerapkan baseline moderasi konten.

## API Contract

Route prefix:

```text
/api/v1/community
```

| Method | Path                                 | Auth class | Purpose              |
| ------ | ------------------------------------ | ---------- | -------------------- |
| `GET`  | `/api/v1/community`                  | Bearer     | ambil feed komunitas |
| `POST` | `/api/v1/community`                  | Bearer     | buat post baru       |
| `POST` | `/api/v1/community/:postId/comments` | Bearer     | tambah komentar      |
| `POST` | `/api/v1/community/:postId/like`     | Bearer     | toggle like/unlike   |

## Database Model

Entitas utama:

- `community_posts`,
- `community_comments`,
- `community_likes`.

Constraint minimum:

- unique like per `(post_id, user_id)`,
- comment terkait post yang valid,
- relasi user ownership terjaga untuk post/comment.

## Authentication and Authorization

- endpoint community wajib bearer auth,
- ownership diterapkan untuk aksi perubahan data,
- user tidak dapat memodifikasi konten milik user lain tanpa izin moderasi.

## Service and Business Rules

- endpoint like bersifat toggle (`like` <-> `unlike`) dan konsisten terhadap state akhir,
- post/comment yang melanggar kebijakan dapat ditandai atau disembunyikan,
- rate limit lebih ketat pada endpoint write.

## Validation Rules

- konten post/comment wajib non-empty,
- batas panjang konten ditegakkan,
- `postId` harus valid,
- payload invalid dipetakan ke `VALIDATION_ERROR`.

## Error Contract

| Condition                    | HTTP  | Error code         |
| ---------------------------- | ----- | ------------------ |
| auth invalid/missing         | `401` | `UNAUTHENTICATED`  |
| akses tidak diizinkan        | `403` | `FORBIDDEN`        |
| post/comment tidak ditemukan | `404` | `NOT_FOUND`        |
| payload invalid              | `422` | `VALIDATION_ERROR` |
| conflict state               | `409` | `CONFLICT`         |
| kegagalan internal           | `500` | `INTERNAL_ERROR`   |

## Observability Contract

Log field minimum:

- `request_id`,
- `user_id`,
- `post_id`,
- `community_action`,
- `status_code`.

Metrik minimum:

- post/comment creation rate,
- like operation rate,
- moderation action count,
- p95 latency endpoint community.

## Testing Requirements

- unit test idempotency like/unlike,
- unit test validator konten,
- integration test unique constraint likes,
- handler test auth/ownership,
- contract test error mapping komunitas.

## Open Gaps

- batas final panjang konten,
- kebijakan delete post/comment final.

## Related Documents

- [Community Moderation Baseline](/Users/macbookpro/Development/recova-backend-v2/docs/modules/community-moderation.md)
- [Users Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/users.md)
- [API Response Standard](/Users/macbookpro/Development/recova-backend-v2/docs/api-response-standard.md)

## Source Reference

- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
- [Fiber Limiter Middleware](https://docs.gofiber.io/middleware/limiter/)
