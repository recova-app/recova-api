# Recova Backend Database Architecture

## Overview

Recova Backend relies on PostgreSQL 16+ as its single source of truth for persistent data storage. Database operations are handled using GORM v1.25+ as the Object-Relational Mapping (ORM) layer, with raw SQL migrations managed via `golang-migrate`.

## Connection Management & Pooling

Database connectivity is managed through `internal/platform/database/gorm.go`. Connection pool settings are configured via environment variables and initialized during application bootstrap:

| Setting              | Env Variable                     | Default / Recommended               | Purpose                                                                      |
| -------------------- | -------------------------------- | ----------------------------------- | ---------------------------------------------------------------------------- |
| Max Open Connections | `DATABASE_MAX_OPEN_CONNS`        | `25`                                | Caps total active and idle connections to prevent exhausting database limits |
| Max Idle Connections | `DATABASE_MAX_IDLE_CONNS`        | `10`                                | Keeps warm connections ready to minimize connection handshake latency        |
| Max Lifetime         | `DATABASE_CONN_MAX_LIFETIME_SEC` | `300`                               | Re-establishes connections periodically to prevent stale TCP sockets         |
| SSL Mode             | `DATABASE_SSL_MODE`              | `disable` (local), `require` (prod) | Enforces TLS transport encryption for database queries                       |

### Production TLS Security Policy

- In production (`APP_ENV=production`), `DATABASE_SSL_MODE` is strictly forbidden from using `disable`.
- The connection URL `DATABASE_URL` must include a valid `sslmode` parameter that matches `DATABASE_SSL_MODE`. Any mismatch halts startup immediately.

## Migration System

Database schema changes are executed strictly using forward and backward SQL migrations via `golang-migrate`. GORM's `AutoMigrate` is deliberately disabled in production to guarantee deterministic, auditable schema transitions.

### Migration Directory Structure

Migrations are stored in `migrations/` with UTC timestamped file pairs:

```text
migrations/
├── 20260508090000_create_core_schema.up.sql
├── 20260508090000_create_core_schema.down.sql
├── 20260508103000_create_auth_refresh_tokens.up.sql
├── 20260508103000_create_auth_refresh_tokens.down.sql
├── 20260509100000_add_checkins_statistics_index.up.sql
├── 20260509100000_add_checkins_statistics_index.down.sql
├── 20260509113000_add_community_threaded_comments.up.sql
├── 20260509113000_add_community_threaded_comments.down.sql
├── 20260512110000_add_manual_auth_columns.up.sql
├── 20260512110000_add_manual_auth_columns.down.sql
├── 20260513101500_add_porn_free_goal_to_users.up.sql
└── 20260513101500_add_porn_free_goal_to_users.down.sql
```

### Migration Management Commands

```bash
# Apply all pending up migrations
make migrate-up

# Roll back the latest migration step
make migrate-down

# View current schema version and migration status
make migrate-status

# Verify that current database schema matches migrations
make migrate-check

# Force database migration dirty state resolution
make migrate-force VERSION=20260513101500
```

## Core Schema & Entity Model

```text
┌──────────────┐       1:1       ┌──────────────┐
│    users     ├─────────────────┤   profiles   │
└──┬───┬───┬───┘                 └──────────────┘
   │   │   │
   │   │   │ 1:N                 ┌──────────────┐
   │   │   ├─────────────────────┤   streaks    │
   │   │   │                     └──────────────┘
   │   │   │ 1:N                 ┌──────────────┐
   │   │   ├─────────────────────┤  check_ins   │
   │   │   │                     └──┬───────────┘
   │   │   │                        │ 1:1
   │   │   │                        ▼
   │   │   │ 1:N                 ┌──────────────┐
   │   │   ├─────────────────────┤   journals   │
   │   │   │                     └──────────────┘
   │   │   │ 1:N                 ┌──────────────────────┐
   │   │   ├─────────────────────┤ auth_refresh_tokens  │
   │   │   │                     └──────────────────────┘
   │   │   │ 1:N                 ┌──────────────┐
   │   │   ├─────────────────────┤   ai_chats   │
   │   │   │                     └──────────────┘
   │   │   │ 1:N
   │   │   ▼
   │   │ ┌──────────────────────┐
   │   │ │   community_posts    │
   │   │ └──┬────────────────┬──┘
   │   │    │ 1:N            │ 1:N
   │   ▼    ▼                ▼
   │ ┌──────────────────────┐┌──────────────────────┐
   └─┤  community_comments  ││ community_post_likes │
     └──────────────────────┘└──────────────────────┘
```

### Table Definitions & Responsibilities

| Table Name                  | Entity Purpose                                   | Key Columns & Constraints                                                                                                          |
| --------------------------- | ------------------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------- |
| `users`                     | Core user identity & authentication credentials  | `id` (UUID PK), `email` (UNIQUE), `username` (UNIQUE, nullable), `password_hash`, `google_id` (UNIQUE, nullable), `porn_free_goal` |
| `profiles`                  | Onboarding survey answers & user profile         | `id` (UUID PK), `user_id` (FK users, UNIQUE), `display_name`, `avatar_url`, `onboarding_answers` (JSONB)                           |
| `streaks`                   | Active and historical recovery streaks           | `id` (UUID PK), `user_id` (FK users), `current_streak`, `longest_streak`, `started_at`, `last_relapse_at`                          |
| `check_ins`                 | Morning and evening recovery status logs         | `id` (UUID PK), `user_id` (FK users), `check_in_date`, `time_of_day`, `mood`, `craving_level`, `is_successful`                     |
| `journals`                  | Private self-reflection journal entries          | `id` (UUID PK), `user_id` (FK users), `check_in_id` (FK check_ins, UNIQUE, nullable), `title`, `content`, `mood`                   |
| `community_posts`           | Peer recovery discussions and stories            | `id` (UUID PK), `user_id` (FK users), `title`, `content`, `category`, `is_anonymous`, `like_count`, `comment_count`                |
| `community_comments`        | Threaded discussion comments and replies         | `id` (UUID PK), `post_id` (FK posts), `user_id` (FK users), `parent_comment_id` (FK comments, self-ref), `content`                 |
| `community_post_likes`      | Many-to-many relationship of user likes          | `id` (UUID PK), `post_id` (FK posts), `user_id` (FK users), UNIQUE(`post_id`, `user_id`)                                           |
| `education_contents`        | Expert educational recovery articles             | `id` (UUID PK), `title`, `category`, `reading_time_minutes`, `content`, `source_url`                                               |
| `daily_motivations`         | Rotating daily motivational quotes               | `id` (UUID PK), `content` (UNIQUE), `author`                                                                                       |
| `daily_challenges`          | Daily behavioral and mental challenges           | `id` (UUID PK), `content` (UNIQUE), `category`                                                                                     |
| `daily_physical_challenges` | Daily physical movement challenges               | `id` (UUID PK), `title`, `description`, `duration_minutes`, UNIQUE(`title`, `description`)                                         |
| `ai_chats`                  | Conversation history with the AI Coach           | `id` (UUID PK), `user_id` (FK users), `role` (`user` \| `assistant`), `content`, `session_id`                                      |
| `auth_refresh_tokens`       | Rotating cryptographically hashed refresh tokens | `id` (UUID PK), `user_id` (FK users), `token_hash` (UNIQUE), `expires_at`, `revoked_at`                                            |

## Key Indexes & Performance Constraints

To sustain fast queries on mobile read flows and high-volume check-ins:

1. **Daily Check-In Uniqueness**:
   - `UNIQUE (user_id, check_in_date)`: Prevents multiple check-in submissions for the same calendar date.
2. **Periodic Statistics Aggregation**:
   - `CREATE INDEX idx_checkins_statistics ON check_ins (user_id, check_in_date, is_successful)`: Accelerates 7-day, 30-day, and all-time success rate aggregations.
3. **Threaded Community Comments**:
   - `CREATE INDEX idx_community_comments_thread ON community_comments (post_id, parent_comment_id, created_at, id)`: Enables low-latency nested comment hierarchy tree fetches.
4. **AI Conversation Retrieval**:
   - `CREATE INDEX idx_ai_chats_user_history ON ai_chats (user_id, created_at)`: Optimizes retrieval of the user's recent chat context.
5. **Session Token Validation**:
   - `CREATE INDEX idx_refresh_tokens_active ON auth_refresh_tokens (user_id, revoked_at, expires_at)`: Speeds up active session lookups and bulk revocation during logout.

## Database Seeding

Baseline educational content, daily motivations, and initial recovery challenges are seeded using raw SQL scripts:

```bash
# Seed initial content into the database
make seed
```

Seed scripts are located in `migrations/seeds/` (e.g. `000001_baseline_seed.sql`) and are written with `ON CONFLICT DO NOTHING` statements so they can be re-run safely in local and development environments without corrupting existing records.
