---
title: Recova Backend Environment Configuration
description: Strategi konfigurasi environment untuk layanan Recova Backend dengan validasi fail-fast, pemisahan environment, dan pengelolaan secret aman.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/environment.md
last_reviewed: 2026-05-08
---

# Recova Backend Environment Configuration

Dokumen ini mendefinisikan model konfigurasi environment untuk layanan backend Recova.

Prinsip inti:

- semua env required harus tervalidasi saat startup,
- env required tidak boleh punya fallback tersembunyi,
- secret tidak boleh dicetak pada log,
- setiap environment memiliki nilai eksplisit dan terpisah.

## Environment Groups

| Group          | Purpose                                                               |
| -------------- | --------------------------------------------------------------------- |
| Application    | mode runtime, port, base URL, dan prefix API                          |
| Database       | koneksi PostgreSQL dan parameter koneksi                              |
| Authentication | JWT signing, Google OAuth, dan cookie/session policy                  |
| AI             | provider AI, model, API key, base URL, timeout, dan fallback opsional |
| Security       | CORS allowlist, rate limit, dan request body limit                    |
| Observability  | log level, request id header, health timeout                          |

## Application Variables

| Variable     | Required | Example value             | Notes                                             |
| ------------ | -------- | ------------------------- | ------------------------------------------------- |
| `APP_NAME`   | Yes      | `recova-backend-v2`       | identitas service untuk startup log dan telemetry |
| `APP_ENV`    | Yes      | `local`                   | allowed: `local`, `test`, `staging`, `production` |
| `NODE_ENV`   | Yes      | `development`             | mode runtime standar ecosystem                    |
| `PORT`       | Yes      | `3001`                    | port HTTP service                                 |
| `API_PREFIX` | Yes      | `/api/v1`                 | prefix endpoint publik                            |
| `DOCS_URL`   | Yes      | `https://docs.recova.app` | URL dokumentasi publik bila tersedia              |
| `APP_URL`    | Yes      | `http://localhost:3001`   | base URL backend untuk callback/link              |

## Database Variables

| Variable                         | Required | Example value                                       | Notes                  |
| -------------------------------- | -------- | --------------------------------------------------- | ---------------------- |
| `DATABASE_URL`                   | Yes      | `postgresql://user:pass@host:5432/db`               | koneksi utama aplikasi |
| `DATABASE_MAX_OPEN_CONNS`        | Yes      | `25`                                                | batas koneksi terbuka  |
| `DATABASE_MAX_IDLE_CONNS`        | Yes      | `10`                                                | batas koneksi idle     |
| `DATABASE_CONN_MAX_LIFETIME_SEC` | Yes      | `300`                                               | umur maksimum koneksi  |
| `DATABASE_SSL_MODE`              | Yes      | `disable` / `require` / `verify-ca` / `verify-full` | mode SSL PostgreSQL    |

Jika deployment memisahkan URL direct migration, definisikan juga:

- `DIRECT_DATABASE_URL`.

Aturan URL database:

- gunakan `postgresql://` sebagai format canonical pada env aplikasi dan workflow,
- nilai boleh ditulis tanpa quote atau dengan quote pembungkus (`DATABASE_URL="postgresql://..."`),
- loader lokal dan runner deploy harus menghapus quote pembungkus sebelum mengekspor env,
- wrapper migrasi boleh menormalisasi URL khusus argumen `golang-migrate` tanpa mengubah env asli.

## Authentication Variables

| Variable                | Required | Example value                       | Notes                                                |
| ----------------------- | -------- | ----------------------------------- | ---------------------------------------------------- |
| `JWT_SECRET`            | Yes      | `replace-with-strong-secret`        | secret signing access token                          |
| `JWT_ACCESS_TTL`        | Yes      | `15m`                               | masa berlaku access token (format durasi Go)         |
| `JWT_REFRESH_TTL`       | Yes      | `7d`                                | masa berlaku refresh token (`d` untuk hari didukung) |
| `GOOGLE_CLIENT_ID`      | Yes      | `123456.apps.googleusercontent.com` | validasi token Google OAuth                          |
| `AUTH_COOKIE_NAME`      | Yes      | `recova_refresh`                    | nama cookie refresh                                  |
| `AUTH_COOKIE_SECURE`    | Yes      | `false` (local), `true` (prod)      | cookie secure wajib true di production               |
| `AUTH_COOKIE_SAME_SITE` | Yes      | `lax`                               | gunakan `none` hanya dengan secure cookie            |
| `AUTH_COOKIE_DOMAIN`    | No       | `recova.app`                        | opsional sesuai topology domain                      |

## AI Variables

| Variable               | Required | Example value                  | Notes                             |
| ---------------------- | -------- | ------------------------------ | --------------------------------- |
| `AI_PROVIDER`          | Yes      | `gemini` / `openai-compatible` | pemilihan provider utama          |
| `AI_MODEL`             | Yes      | `gemini-2.0-flash`             | nama model utama                  |
| `AI_API_KEY`           | Yes      | `replace-with-provider-key`    | credential provider AI            |
| `AI_BASE_URL`          | No       | `https://api.openai.com/v1`    | base URL provider non-default     |
| `AI_TIMEOUT_MS`        | Yes      | `10000`                        | timeout request inferensi         |
| `AI_FALLBACK_PROVIDER` | No       | `openai-compatible`            | provider cadangan jika diaktifkan |
| `AI_FALLBACK_MODEL`    | No       | `gpt-4.1-mini`                 | model cadangan jika diaktifkan    |

Kompatibilitas nilai lama pada source saat ini:

- `GEMINI_API_KEY`, `GEMINI_MODEL`, `OPENAI_API_KEY`, `OPENAI_MODEL`, `OPENAI_BASE_URL` dapat dipetakan ke contract unified di atas selama masa transisi konfigurasi.

## Security Variables

| Variable               | Required | Example value           | Notes                                 |
| ---------------------- | -------- | ----------------------- | ------------------------------------- |
| `CORS_ORIGINS`         | Yes      | `http://localhost:5173` | daftar origin terpisah koma           |
| `RATE_LIMIT_WINDOW_MS` | Yes      | `60000`                 | window limiter global                 |
| `RATE_LIMIT_MAX`       | Yes      | `120`                   | limit request per window              |
| `AUTH_RATE_LIMIT_MAX`  | Yes      | `10`                    | limit lebih ketat untuk endpoint auth |
| `AI_RATE_LIMIT_MAX`    | Yes      | `20`                    | limit lebih ketat untuk endpoint AI   |
| `REQUEST_BODY_LIMIT`   | Yes      | `1mb`                   | batas ukuran payload HTTP             |

## Observability Variables

| Variable                  | Required | Example value  | Notes                         |
| ------------------------- | -------- | -------------- | ----------------------------- |
| `LOG_LEVEL`               | Yes      | `info`         | level structured logger       |
| `REQUEST_ID_HEADER`       | Yes      | `x-request-id` | header korelasi request       |
| `HEALTH_CHECK_TIMEOUT_MS` | Yes      | `2000`         | timeout cek dependency health |

## Secret Management Rules

- secret hanya boleh di-load dari secret manager atau env injection runtime,
- secret tidak boleh ditulis ke source control,
- `.env.example` hanya berisi placeholder aman non-rahasia,
- rotasi secret harus memiliki prosedur dan owner operasional.

## Fail-Fast Rules

- startup harus gagal jika env required hilang atau invalid,
- string kosong dianggap invalid untuk env required,
- nilai enum di luar daftar valid harus ditolak,
- parsing numeric/duration invalid harus menghentikan startup.
- format duration mengikuti `time.ParseDuration` Go dan mendukung suffix `d` (hari) untuk kebutuhan token TTL.

## Local Command Env Loading

Untuk workflow development lokal:

- `make run`, `make migrate-*`, dan `make seed` memuat `.env` otomatis melalui `scripts/with-env.sh`,
- jika perlu profile env berbeda, set `ENV_FILE` saat menjalankan target make.

## Related Documents

- [Environment Matrix and Runtime Profiles](/Users/macbookpro/Development/recova-backend-v2/docs/operations/environments.md)
- [Configuration Validation Standard](/Users/macbookpro/Development/recova-backend-v2/docs/standards/config-validation.md)
- [AI Provider Integration](/Users/macbookpro/Development/recova-backend-v2/docs/integrations/ai-provider.md)

## Source Reference

- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
- [/Users/macbookpro/Development/bisakerja-api/docs/environment.md](/Users/macbookpro/Development/bisakerja-api/docs/environment.md)
- [Google OAuth 2.0 Web Server Flow](https://developers.google.com/identity/protocols/oauth2/web-server)
- [golang-migrate PostgreSQL Driver](https://github.com/golang-migrate/migrate/tree/master/database/postgres)
