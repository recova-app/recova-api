# Recova Backend Domain Modules

## Architecture Overview

All business capabilities in Recova Backend are encapsulated inside domain modules located in `internal/modules/`. Each module is self-contained and enforces strict separation of concerns across standard architectural layers.

## Standard Module Anatomy

Each module follows a consistent layout:

```text
internal/modules/<module-name>/
├── dto.go         # Request and response data transfer objects with validation tags
├── handler.go     # HTTP controller: unmarshals requests, invokes service, formats response
├── mapper.go      # Pure translation functions converting Models <-> DTOs
├── model.go       # GORM domain entity structs and database mappings
├── repository.go  # Database persistence layer executing GORM queries
├── route.go       # Route registrations and middleware bindings
├── service.go     # Core business rules, validation, and domain orchestration
└── validator.go   # Custom validation rules for input payloads
```

### Layer Interaction Invariants

1. **Routes & Middleware**: Bind route paths to handler methods and attach authentication or rate-limiting guards.
2. **Handlers**:
   - Parse URL parameters, query strings, and request bodies into DTOs.
   - Execute validator checks.
   - Invoke service methods with typed domain input.
   - Return responses using standardized envelope helpers (`response.Success`, `response.Error`).
   - Must never perform database operations directly.
3. **Services**:
   - Encapsulate business logic, streak calculations, access controls, and transaction boundaries.
   - Consume and return domain types or DTOs.
   - Must never import the Fiber framework or interact directly with HTTP request contexts.
4. **Repositories**:
   - Execute database queries through GORM instances.
   - Map database rows to domain models.
   - Strictly encapsulate SQL and database dialect details.

## Domain Modules Catalog

### 1. Auth Module (`internal/modules/auth`)

Provides authentication, credential validation, and session lifecycle management.

- **Responsibilities**:
  - Exchange Google ID tokens for session tokens using Google Identity services.
  - Handle manual user registration and login with `bcrypt` password hashing.
  - Generate HMAC-SHA256 JWT access tokens (15-minute validity).
  - Generate, rotate, and revoke SHA-256 hashed refresh tokens (7-day validity).
  - Deliver refresh tokens via secure HTTP-only cookies and JSON fallback.
- **Key Models**: `AuthRefreshToken`
- **Endpoints**:
  - `POST /api/v1/auth/google`: Authenticate via Google ID token.
  - `POST /api/v1/auth/register`: Register with email, username, and password.
  - `POST /api/v1/auth/login`: Authenticate with email and password.
  - `POST /api/v1/auth/refresh`: Rotate refresh token and issue new access token.
  - `POST /api/v1/auth/logout`: Revoke active session tokens.

### 2. Users Module (`internal/modules/users`)

Manages user identity profiles, onboarding progression, and recovery goals.

- **Responsibilities**:
  - Retrieve and update user profile information (display name, avatar URL, bio).
  - Store and validate onboarding questionnaire answers (addiction triggers, baseline habits).
  - Track onboarding completion status.
  - Configure personalized recovery goals (e.g. porn-free targets).
- **Key Models**: `User`, `Profile`
- **Endpoints**:
  - `GET /api/v1/users/me`: Get current user profile.
  - `PATCH /api/v1/users/me`: Update user profile details.
  - `POST /api/v1/users/me/onboarding`: Submit onboarding responses.
  - `GET /api/v1/users/me/onboarding/status`: Inspect onboarding status.
  - `PATCH /api/v1/users/me/goal`: Update recovery target goal.

### 3. Routine Module (`internal/modules/routine`)

The core recovery tracking engine responsible for daily check-ins, relapse management, and streak calculations.

- **Responsibilities**:
  - Record morning and evening daily check-ins (mood, craving level, notes).
  - Enforce one check-in per user per calendar day via database constraints.
  - Calculate active streaks, longest streaks, and milestone achievements.
  - Handle relapse reporting: record relapse timestamp, trigger reasons, and reset active streak counter while preserving historical streak records.
  - Compute recovery statistics (success rates over 7, 30, and 90 days).
  - Provide aggregated activity summaries for dashboard charts.
- **Key Models**: `CheckIn`, `Streak`
- **Endpoints**:
  - `GET /api/v1/routine/check-ins/today`: Get check-in status for today.
  - `POST /api/v1/routine/check-ins`: Submit daily check-in.
  - `GET /api/v1/routine/check-ins/history`: Retrieve historical check-ins.
  - `POST /api/v1/routine/relapse`: Report a relapse event and reset streak.
  - `GET /api/v1/routine/streaks`: Retrieve active and longest streak metrics.
  - `GET /api/v1/routine/statistics`: Comprehensive recovery statistics.
  - `GET /api/v1/routine/statistics/activity-summary`: Aggregated activity summary.

### 4. Journals Module (`internal/modules/journals`)

Allows users to maintain private, reflective journal entries throughout their recovery journey.

- **Responsibilities**:
  - Provide full CRUD operations for private journal entries.
  - Associate journals with specific check-in records or stand-alone dates.
  - Tag journals with emotional moods (e.g. happy, anxious, urged, grateful).
  - Ensure strict user isolation (users can only access their own journals).
- **Key Models**: `Journal`
- **Endpoints**:
  - `GET /api/v1/journals`: List journal entries with pagination.
  - `POST /api/v1/journals`: Create a new private journal entry.
  - `GET /api/v1/journals/{id}`: Read journal entry details.
  - `PUT /api/v1/journals/{id}`: Update an existing journal entry.
  - `DELETE /api/v1/journals/{id}`: Delete a journal entry.

### 5. Community Module (`internal/modules/community`)

Fosters peer support and shared recovery experiences through moderated discussion forums.

- **Responsibilities**:
  - Enable post creation across standardized categories (`saran`, `motivasi`, `cerita`, `pertanyaan`, `bantuan`).
  - Support anonymous posting to protect user privacy.
  - Provide multi-level threaded comments and replies via hierarchical `parent_comment_id` references.
  - Manage post like toggles with idempotent counters.
- **Key Models**: `CommunityPost`, `CommunityComment`, `CommunityPostLike`
- **Endpoints**:
  - `GET /api/v1/community/posts`: List community posts with category filters.
  - `POST /api/v1/community/posts`: Publish a new community post.
  - `GET /api/v1/community/posts/{id}`: Get post details.
  - `DELETE /api/v1/community/posts/{id}`: Remove own post.
  - `GET /api/v1/community/posts/{id}/comments`: Retrieve threaded comments for a post.
  - `POST /api/v1/community/posts/{id}/comments`: Post a comment or reply.
  - `POST /api/v1/community/posts/{id}/like`: Toggle like on a post.

### 6. Education Module (`internal/modules/education`)

Provides recovery literature and habit-rebuilding knowledge articles.

- **Responsibilities**:
  - Serve categorized articles covering neurobiology of addiction, trigger management, and coping mechanisms.
  - Track estimated reading time and article metadata.
- **Key Models**: `EducationContent`
- **Endpoints**:
  - `GET /api/v1/education`: List educational articles by category.
  - `GET /api/v1/education/{id}`: Read full article content.

### 7. Content Module (`internal/modules/content`)

Delivers daily rotating motivational quotes, behavioral challenges, and physical movement goals.

- **Responsibilities**:
  - Rotate daily quotes based on current calendar date.
  - Serve daily mental and physical recovery challenges.
- **Key Models**: `DailyMotivation`, `DailyChallenge`, `DailyPhysicalChallenge`
- **Endpoints**:
  - `GET /api/v1/content/daily`: Retrieve today's daily motivation and challenges.

### 8. AI Coach Module (`internal/modules/ai`)

Provides interactive recovery coaching and emotional support via an AI companion.

- **Responsibilities**:
  - Orchestrate prompts and conversational history with the configured AI provider (Gemini or OpenAI).
  - Maintain context-aware recovery guidance tailored to the user's active streak and recent mood.
  - Manage AI persona preferences (tone: compassionate, direct, scientific).
  - Enforce content safety guardrails and prompt sanitization.
- **Key Models**: `AIChat`
- **Endpoints**:
  - `POST /api/v1/ai/chat`: Send prompt and receive response from AI Coach.
  - `GET /api/v1/ai/history`: Fetch chat session history.
  - `POST /api/v1/ai/reset`: Clear active chat conversation history.
  - `GET /api/v1/ai/preferences`: Retrieve AI persona preferences.
  - `PUT /api/v1/ai/preferences`: Update AI persona preferences.

### 9. Achievements Module (`internal/modules/achievements`)

Celebrates recovery milestones through automated badge unlocking.

- **Responsibilities**:
  - Evaluate user streak milestones (3 days, 7 days, 14 days, 30 days, 90 days, 1 year).
  - Serve available and unlocked achievement badges.
- **Endpoints**:
  - `GET /api/v1/achievements`: List all achievements and user unlock status.

## Inter-Module Boundaries

- Modules must communicate with other modules via public service interfaces, never by querying another module's repository or database tables directly.
- Shared domain types and cross-cutting error definitions reside in `internal/shared/` to eliminate circular dependencies between domain modules.
