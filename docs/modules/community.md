---
title: Community Module
description: Kontrak modul komunitas untuk posting, komentar bertingkat, dan interaksi like dengan kontrol ownership, moderasi, dan anti-abuse.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/modules/community.md
last_reviewed: 2026-05-10
---

# Community Module

## Responsibility

- menyediakan feed komunitas,
- membuat post,
- menambahkan komentar dan reply bertingkat,
- mengelola state like/unlike secara konsisten,
- menerapkan baseline moderasi konten.

## API Contract

Route prefix:

```text
/api/v1/community
```

| Method | Path                                                    | Auth class | Purpose               |
| ------ | ------------------------------------------------------- | ---------- | --------------------- |
| `GET`  | `/api/v1/community`                                     | Bearer     | ambil feed komunitas  |
| `POST` | `/api/v1/community`                                     | Bearer     | buat post baru        |
| `POST` | `/api/v1/community/:postId/comments`                    | Bearer     | tambah komentar       |
| `GET`  | `/api/v1/community/:postId/comments`                    | Bearer     | ambil thread komentar |
| `POST` | `/api/v1/community/:postId/comments/:commentId/replies` | Bearer     | tambah reply komentar |
| `POST` | `/api/v1/community/:postId/like`                        | Bearer     | toggle like/unlike    |

## Database Model

Entitas utama:

- `community_posts`,
- `community_comments`,
- `community_likes`.

Constraint minimum:

- unique like per `(post_id, user_id)`,
- comment terkait post yang valid,
- relasi user ownership terjaga untuk post/comment,
- reply memakai `parent_comment_id` pada tabel komentar,
- depth thread komentar dibatasi kebijakan maksimum `2`,
- reply wajib berada pada post yang sama dengan parent comment.

## Authentication and Authorization

- endpoint community wajib bearer auth,
- ownership diterapkan untuk aksi perubahan data,
- user tidak dapat memodifikasi konten milik user lain tanpa izin moderasi.

## Service and Business Rules

- endpoint like bersifat toggle (`like` <-> `unlike`) dan konsisten terhadap state akhir,
- komentar top-level memiliki `depth = 0`,
- reply komentar menaikkan depth parent + 1,
- reply yang menghasilkan depth di atas `2` ditolak dengan `VALIDATION_ERROR`,
- setiap node komentar menyertakan `reply_count` agar client tidak perlu full scan,
- post/comment yang melanggar kebijakan dapat ditandai atau disembunyikan,
- rate limit lebih ketat pada endpoint write.

## Validation Rules

- konten post/comment wajib non-empty,
- kategori post yang diizinkan: `saran`, `motivasi`, `cerita`, `pertanyaan`, `bantuan`,
- batas panjang konten ditegakkan,
- `postId` harus valid,
- `commentId` parent reply harus valid,
- parent reply harus berada di post yang sama,
- payload invalid dipetakan ke `VALIDATION_ERROR`.

## Threading Policy

- struktur komentar memakai adjacency list (`parent_comment_id`) dengan traversal rekursif terkontrol,
- batas kedalaman thread adalah `2` (root depth `0`, reply maksimal depth `2`),
- query thread wajib deterministic by `created_at` lalu `id`,
- response thread memuat `depth`, `parentCommentId`, `replyCount`,
- implementasi harus aman terhadap siklus data (cycle) lewat constraint/check query traversal.

## Error Contract

| Condition                    | HTTP  | Error code         |
| ---------------------------- | ----- | ------------------ |
| auth invalid/missing         | `401` | `UNAUTHENTICATED`  |
| akses tidak diizinkan        | `403` | `FORBIDDEN`        |
| post/comment tidak ditemukan | `404` | `NOT_FOUND`        |
| payload invalid              | `422` | `VALIDATION_ERROR` |
| conflict state               | `409` | `CONFLICT`         |
| melebihi rate limit          | `429` | `RATE_LIMITED`     |
| kegagalan internal           | `500` | `INTERNAL_ERROR`   |

## Observability Contract

Log field minimum:

- `request_id`,
- `user_id`,
- `post_id`,
- `comment_id`,
- `community_action`,
- `status_code`.

Metrik minimum:

- post/comment creation rate,
- comment-reply depth distribution,
- like operation rate,
- moderation action count,
- p95 latency endpoint community.

## Testing Requirements

- unit test idempotency like/unlike,
- unit test validator konten,
- unit test validator parent comment + depth policy,
- integration test unique constraint likes,
- integration test threaded comment query (depth, ordering, ownership),
- handler test auth/ownership,
- contract test error mapping komunitas.

## Open Gaps

- kebijakan delete post/comment final,
- pagination thread untuk dataset komentar besar.

## Related Documents

- [Community Moderation Baseline](/Users/macbookpro/Development/recova-backend-v2/docs/modules/community-moderation.md)
- [Users Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/users.md)
- [API Response Standard](/Users/macbookpro/Development/recova-backend-v2/docs/api-response-standard.md)

## Source Reference

- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
- [Fiber Limiter Middleware](https://docs.gofiber.io/middleware/limiter/)
