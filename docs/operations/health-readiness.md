---
title: Recova Backend Health and Readiness
description: Kontrak endpoint health, dependency check, dan runbook respons insiden untuk liveness dan readiness backend.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/operations/health-readiness.md
last_reviewed: 2026-05-08
---

# Recova Backend Health and Readiness

Dokumen ini mendefinisikan kontrak endpoint kesehatan layanan untuk observability dan deployment gate.

## Endpoint Contract

| Method | Path            | Tujuan                                          | Dependency check                     |
| ------ | --------------- | ----------------------------------------------- | ------------------------------------ |
| `GET`  | `/health/live`  | memastikan proses aktif dan bisa merespons HTTP | tidak memeriksa dependency eksternal |
| `GET`  | `/health/ready` | memastikan layanan siap menerima trafik         | wajib memeriksa koneksi database     |

## Response Behavior

### Liveness Success

- status `200`
- body ringkas yang menyatakan service hidup

### Readiness Success

- status `200`
- dependency utama berstatus sehat

### Readiness Failure

- status `503`
- body error aman dan tidak membocorkan detail sensitif
- reason operasional dicatat di log internal dengan request id

## Implementation Rules

- endpoint health tidak boleh bergantung pada route bisnis.
- liveness harus cepat dan deterministik.
- readiness harus memiliki timeout check terikat agar tidak hang.
- failure readiness tidak boleh menyebabkan process crash.

## Operational Gate

Sebelum route menerima trafik:

- `/health/live` pass,
- `/health/ready` pass,
- error rate startup dalam batas aman,
- log startup menunjukkan dependency check stabil.

## Incident Handling

Jika `/health/ready` gagal:

1. verifikasi koneksi database dan kredensial runtime.
2. cek pool saturation, timeout, dan resource limit.
3. cek lonjakan error aplikasi pada log request id terkait.
4. aktifkan rollback routing bila layanan tidak pulih pada window operasional.

## Monitoring Signals

- readiness success ratio,
- latency endpoint health,
- count kegagalan dependency check,
- correlation antara spike readiness failure dan error API.

## Related Documents

- [Architecture](/Users/macbookpro/Development/recova-backend-v2/docs/architecture.md)
- [Request Lifecycle](/Users/macbookpro/Development/recova-backend-v2/docs/overview/request-lifecycle.md)
- [API Response Standard](/Users/macbookpro/Development/recova-backend-v2/docs/api-response-standard.md)

## Source Reference

- [Fiber v3 Documentation](https://docs.gofiber.io/)
- [PostgreSQL Current Documentation](https://www.postgresql.org/docs/current/index.html)
- [OpenTelemetry Go Documentation](https://opentelemetry.io/docs/languages/go/)
