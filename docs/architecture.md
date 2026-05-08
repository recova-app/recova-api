---
title: Recova Backend Architecture
description: Kontrak arsitektur runtime, service boundary, request lifecycle, dan kontrol lintas-layanan untuk Recova Backend.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/architecture.md
last_reviewed: 2026-05-08
---

# Recova Backend Architecture

Dokumen ini mendefinisikan kontrak arsitektur backend Recova sebagai layanan API yang menjaga kestabilan kontrak `/api/v1` untuk aplikasi klien.

## Runtime Topology

```text
Client App
  -> Recova Backend API
      -> PostgreSQL
      -> External OAuth Provider
      -> AI Provider
```

Aturan utama:

- aplikasi klien hanya berkomunikasi dengan backend API,
- backend API adalah titik tunggal untuk authn/authz, validasi request, aturan bisnis, dan formatting response,
- PostgreSQL menyimpan state aplikasi dan histori domain,
- provider eksternal tidak boleh menjadi sumber otorisasi pengguna aplikasi.

## Service Boundaries

| Boundary              | Owner                  | Responsibility                                               |
| --------------------- | ---------------------- | ------------------------------------------------------------ |
| Client -> API         | Backend API            | validasi input, authn/authz, rules bisnis, response contract |
| API -> Database       | Backend API            | transaksi data, constraint domain, query access pattern      |
| API -> OAuth provider | Backend API + provider | verifikasi identitas federasi                                |
| API -> AI provider    | Backend API + provider | orkestrasi request AI, timeout/fallback, sanitasi output     |

## Layered Architecture

```text
HTTP Router + Middleware
  -> Handler/Controller
      -> Service (domain rules)
          -> Repository (data access)
              -> PostgreSQL
```

### Layer Rules

| Layer      | Wajib                                                              | Dilarang                                        |
| ---------- | ------------------------------------------------------------------ | ----------------------------------------------- |
| Router     | definisi method/path/middleware                                    | domain logic                                    |
| Middleware | request id, CORS, recover, rate limit, auth guard, validation gate | query domain kompleks                           |
| Handler    | mapping HTTP ke service call                                       | akses DB langsung                               |
| Service    | aturan bisnis, otorisasi berbasis konteks domain, transaksi        | ketergantungan langsung ke objek HTTP framework |
| Repository | query persistence dan mapping data                                 | formatting response client                      |

## Cross-Cutting Controls

- **Error handling**: error terpusat, response aman, tanpa stack trace ke klien.
- **Observability**: request id wajib, log terstruktur, health/readiness endpoint.
- **Security headers**: baseline hardening wajib aktif.
- **Rate limiting**: kebijakan limit berbeda untuk auth, AI, dan read endpoint.
- **Validation**: request params/query/body tervalidasi sebelum service dieksekusi.

## Failure Containment

- Kegagalan provider AI tidak boleh memutus alur non-AI.
- Kegagalan readiness dependency harus memicu `503` pada endpoint readiness.
- Kegagalan validasi harus berhenti sebelum menyentuh service/repository.
- Kegagalan tak terduga dipetakan ke error internal yang aman.

## Health and Readiness Contract

- `GET /health/live`: proses hidup dan menerima request HTTP.
- `GET /health/ready`: layanan siap menerima trafik, termasuk pemeriksaan koneksi database.

Rincian runbook ada di [Health and Readiness](/Users/macbookpro/Development/recova-backend-v2/docs/operations/health-readiness.md).

## Related Documents

- [Request Lifecycle](/Users/macbookpro/Development/recova-backend-v2/docs/overview/request-lifecycle.md)
- [API Reference](/Users/macbookpro/Development/recova-backend-v2/docs/api-reference.md)
- [API Response Standard](/Users/macbookpro/Development/recova-backend-v2/docs/api-response-standard.md)
- [Database](/Users/macbookpro/Development/recova-backend-v2/docs/database.md)
- [Decisions](/Users/macbookpro/Development/recova-backend-v2/docs/decisions/index.md)

## Source Reference

- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
- [/Users/macbookpro/Development/bisakerja-api/docs/architecture.md](/Users/macbookpro/Development/bisakerja-api/docs/architecture.md)
- [Fiber v3 Documentation](https://docs.gofiber.io/)
- [Fiber Error Handling Guide](https://docs.gofiber.io/guide/error-handling)
- [OpenTelemetry Go Documentation](https://opentelemetry.io/docs/languages/go/)
