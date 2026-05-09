---
title: Recovery Enhancement Data Plan
description: Rencana ekspansi skema data untuk statistik lanjutan, threaded comments, achievements, dan preferensi persona AI.
owner: database-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/roadmap/recovery-enhancement-data-plan.md
last_reviewed: 2026-05-09
---

# Recovery Enhancement Data Plan

Dokumen ini memetakan perubahan skema yang diperlukan agar kontrak enhancement dapat diimplementasikan secara konsisten.

## Change Set Summary

1. Threaded comments:
   - tambah `parent_comment_id` pada `community_comments`.
   - tambah `depth` pada `community_comments`.
   - tambah `reply_count` pada `community_comments`.
2. Achievements:
   - tabel baru `achievements`.
   - tabel baru `user_achievement_progress`.
3. AI persona preference:
   - tabel baru `user_ai_persona_preferences`.
4. Statistik lanjutan:
   - gunakan `check_ins`, `streaks`, `journals` sebagai source utama.
   - optional view/materialized view untuk agregasi periodik.

## Proposed Table and Column Contracts

### `community_comments` expansion

- `parent_comment_id uuid null references community_comments(id) on delete cascade`
- `depth smallint not null default 0`
- `reply_count integer not null default 0`

Constraint tambahan:

- `depth >= 0`,
- comment root wajib `parent_comment_id is null and depth = 0`,
- reply wajib `parent_comment_id is not null and depth > 0`.

### `achievements`

- `id uuid primary key default gen_random_uuid()`
- `code text not null unique`
- `title text not null`
- `description text not null`
- `category text not null`
- `threshold numeric not null`
- `is_active boolean not null default true`
- `created_at timestamptz not null default now()`
- `updated_at timestamptz not null default now()`

### `user_achievement_progress`

- `id uuid primary key default gen_random_uuid()`
- `user_id uuid not null references users(id)`
- `achievement_id uuid not null references achievements(id)`
- `progress_value numeric not null default 0`
- `unlocked_at timestamptz null`
- `last_evaluated_at timestamptz not null default now()`
- `created_at timestamptz not null default now()`
- `updated_at timestamptz not null default now()`

Constraint tambahan:

- unique `(user_id, achievement_id)`,
- `progress_value >= 0`.

### `user_ai_persona_preferences`

- `user_id uuid primary key references users(id)`
- `persona text not null`
- `updated_at timestamptz not null default now()`

Constraint tambahan:

- `persona in ('supportive','friendly','concise','direct')`.

## Index Plan

- `idx_community_comments_post_parent_created` pada `(post_id, parent_comment_id, created_at, id)`.
- `idx_community_comments_parent` pada `(parent_comment_id)`.
- `idx_user_achievement_progress_user` pada `(user_id, updated_at desc)`.
- `idx_user_achievement_progress_achievement` pada `(achievement_id)`.
- `idx_check_ins_user_date_success` pada `(user_id, check_in_date, is_successful)` untuk query statistik periodik.

## Migration Safety Rules

- lakukan perubahan `community_comments` bertahap:
  1. add nullable columns,
  2. backfill row existing (`depth=0`, `reply_count=0`),
  3. baru tambah `not null` dan constraint check.
- jalankan transaksi terpisah untuk operasi DDL berisiko lock panjang bila diperlukan.
- sediakan rollback yang memulihkan constraint/index dan kolom baru dengan urutan aman.

## Validation Plan

- migration integration test `up/down/up` wajib lulus,
- cek unique dan FK constraint dengan fixture minimal,
- cek performa query thread komentar dan statistik pada sample data sintetik,
- cek snapshot schema terhadap model GORM target.

## Related Documents

- [Statistics Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/statistics.md)
- [Community Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/community.md)
- [Achievements Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/achievements.md)
- [AI Coach Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/ai-coach.md)
- [Database Migrations](/Users/macbookpro/Development/recova-backend-v2/docs/operations/database-migrations.md)

## Source Reference

- [PostgreSQL WITH Queries](https://www.postgresql.org/docs/current/queries-with.html)
- [GORM Belongs To](https://gorm.io/docs/belongs_to.html)
