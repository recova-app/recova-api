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
| `daily_physical_challenges` | Daily physical movement challenges               | `id` (UUID PK), `title`, `description`, `is_active`, `created_at`, UNIQUE(`title`, `description`)                                 |
| `relapses`                  | Specific relapse event logs with triggers        | `id` (UUID PK), `user_id` (FK users), `check_in_id` (FK check_ins, nullable), `relapse_date`, `mood`, `commitment`, `relapse_trigger` (text[]), `created_at` |
| `achievements`              | Gamification milestone and badge definitions     | `id` (UUID PK), `code` (UNIQUE), `title`, `description`, `category`, `threshold`, `is_active`, `created_at`, `updated_at`        |
| `user_achievement_progress` | User progress and unlock states for achievements | `id` (UUID PK), `user_id` (FK users), `achievement_id` (FK achievements), `progress_value`, `unlocked_at`, `last_evaluated_at`, `created_at`, `updated_at` |
| `user_ai_persona_preferences` | User-selected AI Coach persona preferences      | `user_id` (UUID PK, FK users), `persona`, `updated_at`                                                                            |
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
6. **Journals Chronological Owner Retrieval**:
   - `CREATE INDEX idx_journals_user_created ON journals (user_id, created_at DESC)`: Optimizes user journal listing and pagination without in-memory sort.
7. **Community Feed Timeline and Category Filtering**:
   - `CREATE INDEX idx_community_posts_created_at_desc ON community_posts (created_at DESC)`: Enables index scans for global feed timeline queries.
   - `CREATE INDEX idx_community_posts_category_created_at_desc ON community_posts (category, created_at DESC)`: Optimizes category-filtered community feed retrieval.
8. **Relapse History Ordered Retrieval**:
   - `CREATE INDEX idx_relapses_user_created ON relapses (user_id, created_at DESC)`: Optimizes user relapse history listing and chronological ordering.

## Database Seeding

Baseline educational content, daily motivations, and initial recovery challenges are seeded using raw SQL scripts:

```bash
# Seed initial content into the database
make seed
```

Seed scripts are located in `migrations/seeds/` (e.g. `000001_baseline_seed.sql`) and are written with `ON CONFLICT DO NOTHING` statements so they can be re-run safely in local and development environments without corrupting existing records.

## Schema Drift & Data Contract Verification

To prevent production failures caused by out-of-sync database schemas and Go domain models:

1. **Migration Source of Truth**:
   - Sequential SQL migrations in `migrations/` are the sole authority for database schema transitions. GORM `AutoMigrate` is never used at runtime.
   - Every domain model change must be accompanied by corresponding `.up.sql` and `.down.sql` migration scripts.

2. **Automated Schema Drift Detection**:
   - `CheckSchemaDrift` inspects PostgreSQL's `information_schema.tables` and `information_schema.columns` against all models registered in `models.RegisteredModels()`.
   - Drift detection validates:
     - Table existence for all registered models.
     - Column presence for all fields defined on model structs.
     - Basic type compatibility between Go struct fields and PostgreSQL column definitions.
     - Untracked tables in the `public` database schema.

3. **Fresh Schema and Rollback Testing**:
   - Integration tests ensure that a fresh, empty database can successfully execute all forward migrations up to the latest timestamp, roll back cleanly, and re-apply without errors.
   - Baseline seed integrity is validated against the latest schema on each integration run.
