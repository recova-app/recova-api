---
title: Recova Backend Current Runtime Inventory
description: Inventaris kontrak runtime backend saat ini yang mencakup konfigurasi environment, script operasional, database workflow, dan workflow container.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/roadmap/current-runtime-inventory.md
last_reviewed: 2026-05-08
---

# Recova Backend Current Runtime Inventory

Dokumen ini mencatat kontrak runtime backend yang saat ini sudah tersedia dari sumber resmi layanan, agar operasi dan verifikasi perilaku dapat dilakukan secara konsisten.

## Environment Variable Inventory

### Application

| Variable   | Tujuan                                    |
| ---------- | ----------------------------------------- |
| `PORT`     | Menentukan port layanan HTTP.             |
| `NODE_ENV` | Menentukan mode runtime aplikasi.         |
| `DOCS_URL` | Menentukan URL dokumentasi API eksternal. |

### Authentication

| Variable           | Tujuan                                         |
| ------------------ | ---------------------------------------------- |
| `JWT_SECRET`       | Kunci rahasia untuk penandatanganan token JWT. |
| `GOOGLE_CLIENT_ID` | Client ID untuk login Google OAuth.            |

### AI Integration

| Variable          | Tujuan                                        |
| ----------------- | --------------------------------------------- |
| `GEMINI_API_KEY`  | Kunci API untuk layanan Gemini.               |
| `GEMINI_MODEL`    | Nama model Gemini yang dipakai.               |
| `OPENAI_API_KEY`  | Kunci API untuk layanan OpenAI (opsional).    |
| `OPENAI_MODEL`    | Nama model OpenAI yang dipakai.               |
| `OPENAI_BASE_URL` | Base URL API OpenAI atau endpoint kompatibel. |

### Database

| Variable            | Tujuan                                        |
| ------------------- | --------------------------------------------- |
| `DATABASE_USER`     | Username PostgreSQL untuk workflow container. |
| `DATABASE_PASSWORD` | Password PostgreSQL.                          |
| `DATABASE_NAME`     | Nama database utama layanan.                  |
| `DATABASE_URL`      | URL koneksi PostgreSQL layanan.               |

## Runtime Execution Modes

| Mode        | Command                          | Catatan                                           |
| ----------- | -------------------------------- | ------------------------------------------------- |
| Development | `npm run dev`                    | Menjalankan server development dengan hot reload. |
| Production  | `npm run build` lalu `npm start` | Menjalankan hasil build untuk runtime production. |

## Database Workflow Inventory

| Workflow              | Command              | Tujuan                                                  |
| --------------------- | -------------------- | ------------------------------------------------------- |
| Migrate (development) | `npm run db:migrate` | Menjalankan migrasi schema pada konteks development.    |
| Deploy migration      | `npm run db:deploy`  | Menerapkan migrasi pada konteks deployment/production.  |
| Reset database        | `npm run db:reset`   | Reset database lalu migrasi ulang.                      |
| Push schema           | `npm run db:push`    | Sinkronisasi schema ORM ke database tanpa migrasi file. |
| Seed data             | `npm run db:seed`    | Mengisi data awal untuk development/testing.            |
| DB studio             | `npm run db:studio`  | Membuka studio manajemen data ORM.                      |

## Engineering Script Inventory

| Area                    | Command               |
| ----------------------- | --------------------- |
| Build                   | `npm run build`       |
| Lint                    | `npm run lint`        |
| Lint fix                | `npm run lint:fix`    |
| Format                  | `npm run format`      |
| Post-install generation | `npm run postinstall` |

## Container Workflow Inventory

| Environment | Command                                                  | Catatan                                             |
| ----------- | -------------------------------------------------------- | --------------------------------------------------- |
| Development | `docker-compose -f docker-compose.dev.yml up -d --build` | Menggunakan konfigurasi development dan hot reload. |
| Production  | `docker-compose -f docker-compose.yml up -d --build`     | Menggunakan konfigurasi production.                 |

Artefak container yang disebutkan pada sumber:

- `Dockerfile`
- `Dockerfile.dev`
- `docker-compose.yml`
- `docker-compose.dev.yml`

## Operational Baseline

- Aplikasi mendefinisikan endpoint di bawah prefix `/api/v1`.
- Data seeding tersedia untuk mendukung development dan testing.
- Workflow container tersedia untuk development dan production.

## Known Gaps

- kebijakan secret management pada environment production,
- kontrak health/readiness endpoint,
- kebijakan logging terstruktur dan retensi log,
- kebijakan metrics dan tracing,
- SLA startup/shutdown serta timeout service,
- kontrak backup/restore database,
- kontrol keamanan runtime (CORS policy, rate limit policy, abuse control).

## Related Documents

- [Current Express Baseline](/Users/macbookpro/Development/recova-backend-v2/docs/roadmap/express-baseline.md)
- [Feature Inventory](/Users/macbookpro/Development/recova-backend-v2/docs/roadmap/feature-inventory.md)
- [Recova Backend Documentation Overview](/Users/macbookpro/Development/recova-backend-v2/docs/overview.md)

## Source Reference

- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
