# Recova Backend Environment Configuration

## Configuration Architecture

Recova Backend implements a strict **fail-fast** configuration policy managed via `internal/platform/config/config.go`. All required configuration settings, credentials, and constraints are validated immediately upon application bootstrap.

- If any required environment variable is missing or malformed, the application logs a clear startup error and terminates immediately.
- There are no hidden default fallbacks for security credentials (e.g. `JWT_SECRET` or `AI_API_KEY`).
- Secret values are never printed in plaintext, startup logs, or error responses.

## Environment Profiles

The application recognizes four runtime profiles determined by `APP_ENV`:

| Profile      | Target Environment      | Security & Behavior Rules                                                        |
| ------------ | ----------------------- | -------------------------------------------------------------------------------- |
| `local`      | Developer workstation   | Allows `DATABASE_SSL_MODE=disable`; debug logs enabled; relaxed CORS             |
| `test`       | Automated test runners  | Uses isolated test database; in-memory mocks for external OAuth and AI           |
| `staging`    | Pre-production testing  | Production-parity constraints; requires secure cookies and SSL database          |
| `production` | Live production cluster | Strict security; `DATABASE_SSL_MODE=disable` strictly prohibited; HTTPS enforced |

## Complete Environment Variables Reference

### Application Settings

| Variable     | Required | Type   | Example Value             | Description                                              |
| ------------ | -------- | ------ | ------------------------- | -------------------------------------------------------- |
| `APP_NAME`   | Yes      | string | `recova-backend-v2`       | Service name used for logging and telemetry              |
| `APP_ENV`    | Yes      | string | `local`                   | Active profile: `local`, `test`, `staging`, `production` |
| `NODE_ENV`   | Yes      | string | `development`             | Compatibility runtime mode                               |
| `PORT`       | Yes      | int    | `3001`                    | TCP port where Fiber listens for HTTP traffic            |
| `API_PREFIX` | Yes      | string | `/api/v1`                 | Root path prefix for public application endpoints        |
| `APP_URL`    | Yes      | string | `http://localhost:3001`   | Canonical public base URL of the service                 |
| `DOCS_URL`   | No       | string | `https://docs.recova.app` | Public link to published documentation                   |

### Database Settings

| Variable                         | Required | Type   | Example Value                                                  | Description                                             |
| -------------------------------- | -------- | ------ | -------------------------------------------------------------- | ------------------------------------------------------- |
| `DATABASE_URL`                   | Yes      | string | `postgresql://user:pass@localhost:5432/recova?sslmode=disable` | Main PostgreSQL connection connection string            |
| `DATABASE_MAX_OPEN_CONNS`        | Yes      | int    | `25`                                                           | Maximum open database connections in pool               |
| `DATABASE_MAX_IDLE_CONNS`        | Yes      | int    | `10`                                                           | Maximum idle database connections in pool               |
| `DATABASE_CONN_MAX_LIFETIME_SEC` | Yes      | int    | `300`                                                          | Maximum connection reuse duration in seconds            |
| `DATABASE_SSL_MODE`              | Yes      | string | `disable` \| `require` \| `verify-ca` \| `verify-full`         | PostgreSQL SSL connection mode                          |
| `DIRECT_DATABASE_URL`            | No       | string | `postgresql://user:pass@localhost:5432/recova`                 | Direct URL bypassing connection pooler (for migrations) |

### Authentication & Sessions

| Variable                | Required | Type     | Example Value                      | Description                                           |
| ----------------------- | -------- | -------- | ---------------------------------- | ----------------------------------------------------- |
| `JWT_SECRET`            | Yes      | string   | `random-32-byte-hex-secret`        | Cryptographic secret for signing HMAC-SHA256 JWTs     |
| `JWT_ACCESS_TTL`        | Yes      | duration | `15m`                              | Access token lifespan (e.g. `15m`, `1h`)              |
| `JWT_REFRESH_TTL`       | Yes      | duration | `7d`                               | Refresh token session lifespan (e.g. `7d`, `30d`)     |
| `GOOGLE_CLIENT_ID`      | Yes      | string   | `12345.apps.googleusercontent.com` | Google OAuth client ID for token verification         |
| `AUTH_COOKIE_NAME`      | Yes      | string   | `recova_refresh`                   | Cookie name for rotating refresh token                |
| `AUTH_COOKIE_SECURE`    | Yes      | boolean  | `false` (local) / `true` (prod)    | Enforce HTTPS `Secure` flag on refresh cookie         |
| `AUTH_COOKIE_SAME_SITE` | Yes      | string   | `lax`                              | Cookie `SameSite` attribute (`lax`, `strict`, `none`) |
| `AUTH_COOKIE_DOMAIN`    | No       | string   | `recova.app`                       | Cookie domain boundary (optional)                     |

### AI Provider Settings

| Variable               | Required | Type   | Example Value                               | Description                                  |
| ---------------------- | -------- | ------ | ------------------------------------------- | -------------------------------------------- |
| `AI_PROVIDER`          | Yes      | string | `gemini` \| `openai-compatible`             | Primary AI provider for AI Coach             |
| `AI_MODEL`             | Yes      | string | `gemini-2.0-flash`                          | Language model identifier                    |
| `AI_API_KEY`           | Yes      | string | `ai-provider-api-key`                       | API key credential for AI provider           |
| `AI_BASE_URL`          | No       | string | `https://generativelanguage.googleapis.com` | Custom endpoint base URL (optional)          |
| `AI_TIMEOUT_MS`        | Yes      | int    | `10000`                                     | Maximum inference timeout in milliseconds    |
| `AI_FALLBACK_PROVIDER` | No       | string | `openai-compatible`                         | Optional secondary provider if primary fails |
| `AI_FALLBACK_MODEL`    | No       | string | `gpt-4.1-mini`                              | Model identifier for fallback provider       |

### Security & Rate Limiting

| Variable                | Required | Type   | Example Value                   | Description                                  |
| ----------------------- | -------- | ------ | ------------------------------- | -------------------------------------------- |
| `CORS_ALLOWED_ORIGINS`  | No       | string | `*` or `https://app.recova.com` | Comma-separated list of allowed CORS origins |
| `RATE_LIMIT_MAX`        | No       | int    | `100`                           | Maximum requests per time window             |
| `RATE_LIMIT_WINDOW_SEC` | No       | int    | `60`                            | Time window for rate limiter in seconds      |

### Observability

| Variable    | Required | Type   | Example Value                          | Description                     |
| ----------- | -------- | ------ | -------------------------------------- | ------------------------------- |
| `LOG_LEVEL` | Yes      | string | `info` \| `debug` \| `warn` \| `error` | Minimum logging level threshold |

## Security Invariants

### 1. Database SSL Mode Verification

- When `APP_ENV=production`, `DATABASE_SSL_MODE` is strictly required to be `require`, `verify-ca`, or `verify-full`. Setting it to `disable` will abort application startup with a `CONFIG_VALIDATION_ERROR`.
- If the `DATABASE_URL` query parameter includes `sslmode`, its value must match `DATABASE_SSL_MODE` identically.

### 2. Outbound AI Provider Validation

- In `staging` and `production`, if `AI_BASE_URL` is configured, it must strictly use the `https://` protocol scheme.
- The host must belong to an approved provider domain (e.g. `generativelanguage.googleapis.com` for Gemini).

### 3. Secret Masking

- Sensitive variables (`JWT_SECRET`, `AI_API_KEY`, passwords in database connection strings) are sanitized before configuration telemetry or debug dumps are emitted.
