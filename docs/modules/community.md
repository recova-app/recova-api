---
title: Community Module
description: Kontrak modul komunitas untuk posting, komentar, interaksi like, visibilitas publik, dan batasan akses pemilik data.
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

Modul community mengelola interaksi sosial pengguna berupa post, komentar, dan like.

## Responsibility

Modul community bertanggung jawab pada:

- mengambil daftar post komunitas,
- membuat post baru,
- menambahkan komentar pada post,
- mencatat like dan unlike secara konsisten,
- menjaga visibilitas identitas penulis sesuai kebijakan privasi.

Modul community tidak bertanggung jawab pada:

- autentikasi identitas pengguna,
- manajemen profil pengguna,
- evaluasi keamanan prompt AI.

## Route Prefix

```text
/api/v1/community
```

## Endpoint Summary

| Method | Path                                 | Auth   | Purpose                                     |
| ------ | ------------------------------------ | ------ | ------------------------------------------- |
| `GET`  | `/api/v1/community`                  | Bearer | Ambil feed post komunitas                   |
| `POST` | `/api/v1/community`                  | Bearer | Buat post komunitas baru                    |
| `POST` | `/api/v1/community/:postId/comments` | Bearer | Tambah komentar ke post tertentu            |
| `POST` | `/api/v1/community/:postId/like`     | Bearer | Atur state like pengguna terhadap satu post |

## Post and Comment Contract

Field minimum untuk post:

- `id`,
- `author_id`,
- `content`,
- `created_at`,
- `updated_at`.

Field minimum untuk komentar:

- `id`,
- `post_id`,
- `author_id`,
- `content`,
- `created_at`.

Aturan kontrak:

- `author_id` berasal dari auth context server, bukan dari payload klien,
- konten kosong ditolak,
- konten melanggar kebijakan moderasi ditolak atau ditandai sesuai aturan moderasi.

## Visibility Rules

Aturan visibilitas baseline:

- feed komunitas dapat ditampilkan ke pengguna terautentikasi,
- identitas penulis ditampilkan dalam bentuk profil publik terbatas,
- data sensitif profil (misalnya alasan recovery detail) tidak boleh diekspos pada feed komunitas,
- post atau komentar yang dimoderasi dapat disembunyikan dari feed publik.

## Like and Unlike Behavior

Kontrak perilaku like:

- endpoint like harus idempotent terhadap state akhir,
- satu pengguna hanya boleh memiliki satu relasi like aktif untuk satu post,
- permintaan like pada post yang sudah di-like diperlakukan sebagai no-op sukses,
- permintaan unlike pada post yang belum di-like diperlakukan sebagai no-op sukses,
- penghitungan total like harus konsisten terhadap state relasi unik `(post_id, user_id)`.

## Duplicate Like Conflict Handling

Kontrol konsistensi:

- database wajib memiliki unique constraint `(post_id, user_id)` untuk relasi like,
- ketika request paralel memicu konflik unique, service memetakan hasil ke state idempotent final,
- konflik tidak boleh menghasilkan data like duplikat,
- bila kontrak response membutuhkan sinyal eksplisit, gunakan error `CONFLICT` hanya untuk request yang benar-benar ambigu, bukan retry identik.

## Abuse and Rate-Limit Direction

Baseline proteksi:

- endpoint create post dan create comment wajib punya rate limit lebih ketat daripada read endpoint,
- endpoint like harus dibatasi untuk mencegah spam interaksi,
- payload konten yang diduga abuse perlu dipetakan ke alur moderasi,
- event abuse dicatat sebagai metadata tanpa menyimpan payload sensitif berlebihan.

## Observability and Audit Rules

- log hanya metadata operasional: `request_id`, `user_id`, `post_id`, `action`, status,
- konten mentah post/komentar tidak dicatat pada log error umum,
- audit event minimum: post created, comment created, like state changed, moderation action.

## Open Gaps

- aturan final apakah feed komunitas boleh untuk pengguna non-login,
- batas panjang final konten post/komentar,
- kebijakan final soft-delete atau hard-delete untuk post dan komentar.

## Related Documents

- [Community Moderation Baseline](/Users/macbookpro/Development/recova-backend-v2/docs/modules/community-moderation.md)
- [Users Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/users.md)
- [API Response Standard](/Users/macbookpro/Development/recova-backend-v2/docs/api-response-standard.md)
- [Data Sensitivity Matrix](/Users/macbookpro/Development/recova-backend-v2/docs/references/data-sensitivity-matrix.md)

## Source Reference

- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
- [/Users/macbookpro/Development/bisakerja-api/docs/modules/users.md](/Users/macbookpro/Development/bisakerja-api/docs/modules/users.md)
- [Fiber Limiter Middleware](https://docs.gofiber.io/middleware/limiter/)
