# Recova Backend API Specification

## Design Conventions

All API endpoints follow RESTful design standards and adhere to strict contract conventions:

- **Base URL Prefix**: All public application endpoints are prefixed with `/api/v1`.
- **Content Type**: Requests and responses use `application/json; charset=utf-8` unless explicitly handling file uploads or static docs.
- **HTTP Methods**: Standard semantic verbs:
  - `GET`: Read resources idempotently.
  - `POST`: Create new resources or execute non-idempotent actions (e.g. login, check-in).
  - `PUT`: Completely replace an existing resource.
  - `PATCH`: Partially update an existing resource.
  - `DELETE`: Remove a resource.
- **Resource Naming**: Plural nouns for resource collections (e.g. `/api/v1/journals`, `/api/v1/community/posts`).

## Response Envelopes

Every JSON response emitted by the API is wrapped in a consistent top-level envelope.

### 1. Success Envelope

```json
{
  "success": true,
  "message": "Permintaan berhasil diproses",
  "data": {
    "id": "e4b17f54-58a1-4328-912c-0e2417e29bb3",
    "name": "Alex"
  },
  "meta": null
}
```

### 2. Error Envelope

```json
{
  "success": false,
  "message": "Permintaan gagal diproses",
  "data": null,
  "error": {
    "code": "VALIDATION_ERROR",
    "details": ["Field 'email' harus berupa alamat email yang valid"],
    "request_id": "req_8f17a941-6e9a-4c91-b92c-619a823b1234"
  }
}
```

### Envelope Field Specifications

| Field              | Type                      | Description                                                                          |
| ------------------ | ------------------------- | ------------------------------------------------------------------------------------ |
| `success`          | `boolean`                 | `true` for HTTP 2xx statuses; `false` for 4xx and 5xx statuses                       |
| `message`          | `string`                  | User-friendly summary message in Indonesian suitable for direct display in mobile UI |
| `data`             | `object \| array \| null` | The payload data for successful requests; strictly `null` on error responses         |
| `meta`             | `object \| null`          | Optional metadata (e.g. pagination) or `null` if not applicable                      |
| `error`            | `object \| null`          | Present only on failure; contains machine-readable error details                     |
| `error.code`       | `string`                  | Stable, machine-readable uppercase error identifier (English)                        |
| `error.details`    | `array`                   | Detailed validation errors or contextual failure reasons (Indonesian)                |
| `error.request_id` | `string`                  | Unique request ID for log correlation and tracing                                    |

## Pagination Metadata

Paginated list endpoints include pagination metadata under `meta.pagination`:

```json
{
  "meta": {
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 125,
      "totalPages": 7,
      "hasNextPage": true,
      "hasPrevPage": false
    }
  }
}
```

- Query parameters: `?page=1&limit=20` (defaults: `page=1`, `limit=10`, maximum `limit=100`).
- If `page` exceeds `totalPages`, `data` returns an empty array `[]` with updated pagination metadata.

## Standard Error Codes

The API maps business and runtime failures to stable HTTP status codes and error identifiers:

| Error Code            | HTTP Status | Meaning & Typical Scenarios                                                        |
| --------------------- | ----------- | ---------------------------------------------------------------------------------- |
| `BAD_REQUEST`         | `400`       | Malformed JSON syntax, invalid query parameters, or missing headers                |
| `VALIDATION_ERROR`    | `422`       | Payload schema validation failed (missing required fields, format violations)      |
| `UNAUTHENTICATED`     | `401`       | Missing, expired, or cryptographically invalid access token                        |
| `FORBIDDEN`           | `403`       | User authenticated but lacks permission to access or modify resource               |
| `NOT_FOUND`           | `404`       | Requested resource ID does not exist                                               |
| `CONFLICT`            | `409`       | Resource state collision (e.g. email already registered, duplicate daily check-in) |
| `RATE_LIMITED`        | `429`       | Request threshold exceeded; client must wait before retrying                       |
| `DOWNSTREAM_ERROR`    | `502`       | External dependency failure (e.g. Google OAuth or AI provider error)               |
| `SERVICE_UNAVAILABLE` | `503`       | Backend database unreachable or system in maintenance mode                         |
| `INTERNAL_ERROR`      | `500`       | Unhandled internal exception; details logged privately on server                   |

## Authentication & Session Management

Recova employs a dual-token authentication pattern combining short-lived JWT access tokens with long-lived rotating refresh tokens.

### 1. Google OAuth Sign-In

- **Endpoint**: `POST /api/v1/auth/google`
- **Payload**: `{ "id_token": "<google_id_token>" }`
- **Behavior**: Verifies the ID token against Google's token verification service. If user exists, generates session tokens; if new, automatically provisions user record.

### 2. Manual Authentication

- **Register**: `POST /api/v1/auth/register` (`email`, `password`, `username`)
- **Login**: `POST /api/v1/auth/login` (`email`, `password`)
- Passwords are securely hashed with `bcrypt` (cost 10).

### 3. Dual-Token Mechanism

- **Access Token**:
  - Format: HMAC-SHA256 JWT signed with `JWT_SECRET`.
  - Expiry: 15 minutes (`JWT_ACCESS_TTL=15m`).
  - Transmission: Sent in HTTP request header: `Authorization: Bearer <access_token>`.
- **Refresh Token**:
  - Format: Cryptographically secure random UUID / token string.
  - Storage: SHA-256 hash stored in `auth_refresh_tokens` database table with user ID, expiry, and revocation status.
  - Transmission: Delivered as an `HttpOnly`, `SameSite=Lax`, `Secure` cookie (named `recova_refresh`), with fallback JSON body response for non-cookie mobile clients.
  - Expiry: 7 days (`JWT_REFRESH_TTL=7d`).
  - Rotation: Each refresh exchange (`POST /api/v1/auth/refresh`) revokes the old refresh token and issues a new pair.

### 4. Session Termination

- **Endpoint**: `POST /api/v1/auth/logout`
- Revokes the active refresh token in the database and clears the refresh cookie.

## Endpoint Catalog

### Platform & Health Routes

| Method | Path            | Auth Required | Description                             |
| ------ | --------------- | ------------- | --------------------------------------- |
| `GET`  | `/health/live`  | No            | Process liveness probe                  |
| `GET`  | `/health/ready` | No            | Database connection readiness probe     |
| `GET`  | `/metrics`      | No            | Prometheus telemetry metrics            |
| `GET`  | `/openapi.yaml` | No            | OpenAPI 3.1 specification file          |
| `GET`  | `/docs/api`     | No            | Interactive Scalar API documentation UI |

### Authentication (`/api/v1/auth`)

| Method | Path                    | Auth Required    | Description                                     |
| ------ | ----------------------- | ---------------- | ----------------------------------------------- |
| `POST` | `/api/v1/auth/google`   | No               | Authenticate via Google ID token                |
| `POST` | `/api/v1/auth/register` | No               | Register new email/password account             |
| `POST` | `/api/v1/auth/login`    | No               | Log in with email and password                  |
| `POST` | `/api/v1/auth/refresh`  | No (Cookie/Body) | Rotate refresh token and issue new access token |
| `POST` | `/api/v1/auth/logout`   | Yes              | Revoke active refresh token session             |

### Users & Onboarding (`/api/v1/users`)

| Method  | Path                                 | Auth Required | Description                                     |
| ------- | ------------------------------------ | ------------- | ----------------------------------------------- |
| `GET`   | `/api/v1/users/me`                   | Yes           | Get authenticated user profile                  |
| `PATCH` | `/api/v1/users/me`                   | Yes           | Update profile details (username, avatar, etc.) |
| `POST`  | `/api/v1/users/me/onboarding`        | Yes           | Submit onboarding questionnaire answers         |
| `GET`   | `/api/v1/users/me/onboarding/status` | Yes           | Check user onboarding completion status         |
| `PATCH` | `/api/v1/users/me/goal`              | Yes           | Configure user porn-free target goal            |

### Routine & Relapse Tracking (`/api/v1/routine`)

| Method | Path                                          | Auth Required | Description                                      |
| ------ | --------------------------------------------- | ------------- | ------------------------------------------------ |
| `GET`  | `/api/v1/routine/check-ins/today`             | Yes           | Retrieve today's morning/evening check-in status |
| `POST` | `/api/v1/routine/check-ins`                   | Yes           | Submit daily recovery check-in                   |
| `GET`  | `/api/v1/routine/check-ins/history`           | Yes           | List historical check-in records                 |
| `POST` | `/api/v1/routine/relapse`                     | Yes           | Log a relapse event and reset active streak      |
| `GET`  | `/api/v1/routine/streaks`                     | Yes           | Get active and longest recovery streak metrics   |
| `GET`  | `/api/v1/routine/statistics`                  | Yes           | Get comprehensive recovery statistics            |
| `GET`  | `/api/v1/routine/statistics/activity-summary` | Yes           | Get aggregated activity summary                  |

### Journals (`/api/v1/journals`)

| Method   | Path                    | Auth Required | Description                             |
| -------- | ----------------------- | ------------- | --------------------------------------- |
| `GET`    | `/api/v1/journals`      | Yes           | List user's journal entries (paginated) |
| `POST`   | `/api/v1/journals`      | Yes           | Create a new private journal entry      |
| `GET`    | `/api/v1/journals/{id}` | Yes           | Retrieve a specific journal entry by ID |
| `PUT`    | `/api/v1/journals/{id}` | Yes           | Update an existing journal entry        |
| `DELETE` | `/api/v1/journals/{id}` | Yes           | Delete a journal entry                  |

### Community (`/api/v1/community`)

| Method   | Path                                    | Auth Required | Description                                          |
| -------- | --------------------------------------- | ------------- | ---------------------------------------------------- |
| `GET`    | `/api/v1/community/posts`               | Yes           | List public community posts (filterable by category) |
| `POST`   | `/api/v1/community/posts`               | Yes           | Create a community post                              |
| `GET`    | `/api/v1/community/posts/{id}`          | Yes           | View post details and reply counter                  |
| `DELETE` | `/api/v1/community/posts/{id}`          | Yes           | Delete own community post                            |
| `GET`    | `/api/v1/community/posts/{id}/comments` | Yes           | Get threaded comments and replies for a post         |
| `POST`   | `/api/v1/community/posts/{id}/comments` | Yes           | Add a comment or reply to a post                     |
| `POST`   | `/api/v1/community/posts/{id}/like`     | Yes           | Toggle like on a community post                      |

### Education & Daily Content (`/api/v1/education`, `/api/v1/content`)

| Method | Path                     | Auth Required | Description                                    |
| ------ | ------------------------ | ------------- | ---------------------------------------------- |
| `GET`  | `/api/v1/education`      | Yes           | List educational articles by category          |
| `GET`  | `/api/v1/education/{id}` | Yes           | View full educational article content          |
| `GET`  | `/api/v1/content/daily`  | Yes           | Retrieve daily motivation and daily challenges |

### AI Coach (`/api/v1/ai`)

| Method | Path                     | Auth Required | Description                                            |
| ------ | ------------------------ | ------------- | ------------------------------------------------------ |
| `POST` | `/api/v1/ai/chat`        | Yes           | Send message to AI recovery coach and receive response |
| `GET`  | `/api/v1/ai/history`     | Yes           | Retrieve AI chat history for active session            |
| `POST` | `/api/v1/ai/reset`       | Yes           | Clear AI chat history session                          |
| `GET`  | `/api/v1/ai/preferences` | Yes           | Get AI persona preferences (tone, style)               |
| `PUT`  | `/api/v1/ai/preferences` | Yes           | Update AI persona preferences                          |

### Achievements (`/api/v1/achievements`)

| Method | Path                   | Auth Required | Description                                 |
| ------ | ---------------------- | ------------- | ------------------------------------------- |
| `GET`  | `/api/v1/achievements` | Yes           | List all available badges and unlock status |

## OpenAPI Contract & Scalar Documentation

1. **Source of Truth**: The canonical OpenAPI 3.1 contract resides in `api/openapi/openapi.yaml`.
2. **Automated Generation**: Running `make openapi-generate` updates `docs/generated/openapi.yaml` and synchronizes `docs/generated/routes.md` with runtime routes registered in Fiber.
3. **Interactive Scalar UI**: When running the server, visit `http://localhost:3001/docs/api` to explore all schemas, test payloads, and view live endpoints.
