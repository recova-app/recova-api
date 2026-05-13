---
title: Health Module
description: Kontrak modul health untuk endpoint liveness dan readiness sebagai sinyal status runtime dan dependency layanan.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/modules/health.md
last_reviewed: 2026-05-08
---

# Health Module

## Responsibility

- menyediakan endpoint liveness,
- menyediakan endpoint readiness,
- memberi sinyal status service untuk kebutuhan orchestration dan monitoring.

## API Contract

Route prefix:

```text
/health
```

| Method | Path            | Auth class      | Purpose                          |
| ------ | --------------- | --------------- | -------------------------------- |
| `GET`  | `/health/live`  | Public/internal | cek proses HTTP hidup            |
| `GET`  | `/health/ready` | Public/internal | cek service siap melayani trafik |

## Database Model

Modul health tidak memiliki tabel domain sendiri.

Dependensi utama readiness:

- koneksi database,
- dependensi wajib runtime lain yang ditetapkan sebagai readiness dependency.
- dependency placeholder hanya dipakai bila dependency kontrak sudah ditentukan tetapi integrasi runtime belum aktif.

## Authentication and Authorization

- endpoint health dapat dibuka untuk kebutuhan infrastruktur,
- jika dibatasi jaringan, pembatasan dilakukan di layer gateway/network,
- response tidak boleh membocorkan detail sensitif dependency.

## Service and Business Rules

- liveness hanya memverifikasi proses dapat merespons,
- readiness memverifikasi dependency kritis bertipe `required`,
- dependency bertipe `placeholder` dilaporkan di ringkasan checks tanpa memblokir readiness sukses,
- koneksi database berada pada mode `required` sebagai gate readiness utama runtime,
- timeout readiness harus terukur dan konsisten,
- status readiness gagal memblokir promote release.

## Validation Rules

- endpoint health tidak menerima payload body,
- query parameter tak dikenal diabaikan atau ditolak sesuai kebijakan global,
- response schema harus stabil untuk tooling observability.

## Error Contract

| Condition                             | HTTP  | Error code            |
| ------------------------------------- | ----- | --------------------- |
| liveness gagal                        | `503` | `SERVICE_UNAVAILABLE` |
| readiness dependency `required` gagal | `503` | `SERVICE_UNAVAILABLE` |
| kegagalan internal handler            | `500` | `INTERNAL_ERROR`      |

## Observability Contract

Log field minimum:

- `request_id`,
- `health_type` (`live`/`ready`),
- `status_code`,
- `dependency_state` (ringkas).

Metrik minimum:

- health request rate,
- readiness success ratio,
- readiness check duration.

## Testing Requirements

- unit test readiness evaluator,
- handler test status code live/ready,
- integration test readiness saat database down,
- smoke test endpoint health di pipeline deploy.

## Open Gaps

- daftar final dependency readiness wajib,
- format final payload details dependency untuk setiap status (`ok`, `placeholder`, `down`),
- strategi pembatasan akses health endpoint di production.

## Related Documents

- [Health and Readiness](/Users/macbookpro/Development/recova-backend-v2/docs/operations/health-readiness.md)
- [Post-Deploy Checks](/Users/macbookpro/Development/recova-backend-v2/docs/operations/post-deploy-checks.md)
- [Observability](/Users/macbookpro/Development/recova-backend-v2/docs/operations/observability.md)

## Source Reference

- [Fiber Health Check Middleware](https://docs.gofiber.io/middleware/healthcheck/)
- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
