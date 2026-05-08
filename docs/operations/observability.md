---
title: Recova Backend Observability
description: Baseline observability untuk request diagnostics, structured logs, metrics, traces, dan health signals agar kegagalan layanan dapat dideteksi dan ditangani cepat.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/operations/observability.md
last_reviewed: 2026-05-08
---

# Recova Backend Observability

Dokumen ini menetapkan model observability minimum untuk operasi harian dan respons insiden.

## Observability Goals

- korelasi setiap request dan error menggunakan request id,
- visibilitas latency, error rate, dan dependency health,
- triage kegagalan tanpa membuka data sensitif,
- sinyal readiness yang konsisten untuk gate deployment.

## Core Signals

| Signal           | Wajib | Tujuan                                    |
| ---------------- | ----- | ----------------------------------------- |
| Request ID       | ya    | korelasi log, error response, dan tracing |
| Structured logs  | ya    | investigasi cepat dan query log konsisten |
| Metrics          | ya    | deteksi anomali latency/error/traffic     |
| Traces           | ya    | analisis bottleneck lintas dependency     |
| Health endpoints | ya    | readiness/liveness gate                   |
| Audit events     | ya    | jejak event penting keamanan dan auth     |

## Request ID Contract

- header default: `X-Request-ID`,
- gunakan request id dari upstream jika valid,
- generate request id bila tidak ada,
- wajib return request id pada error response.

## Structured Logging Baseline

Field minimum log request summary:

- timestamp,
- level,
- service,
- env,
- requestId,
- method,
- route,
- statusCode,
- durationMs,
- errorCode (jika gagal).

Larangan log:

- raw password/token/cookie,
- raw body endpoint sensitif,
- raw chat AI/jurnal.

## Metrics Baseline

### HTTP Metrics

- request count per route + status class,
- request latency histogram,
- 4xx/5xx rate,
- rate-limit hit count.

### Dependency Metrics

- PostgreSQL latency + error rate,
- AI provider latency + timeout rate,
- outbound integration error rate.

### Business-Critical Metrics

- auth success/failure ratio,
- check-in write success ratio,
- AI request completion ratio.

## Tracing Baseline

- span root per request HTTP,
- span dependency call (DB/AI/downstream),
- propagate context antar layer,
- tandai error pada span status dengan aman.

## Health Signals

- `GET /health/live`: hanya cek proses hidup,
- `GET /health/ready`: cek dependency kritis (minimum database),
- readiness gagal harus memicu `503` dengan payload aman.

## Alerting Baseline

Alert minimum:

- readiness failure berulang,
- lonjakan `5xx`,
- lonjakan auth failure,
- lonjakan timeout AI provider,
- lonjakan latency endpoint kritikal.

## Data Safety in Observability

- redaksi kunci sensitif wajib aktif,
- sampling trace tidak boleh meloloskan payload sensitif,
- export observability harus tunduk ke kebijakan retensi data.

## Related Documents

- [Failure Scenarios](/Users/macbookpro/Development/recova-backend-v2/docs/operations/failure-scenarios.md)
- [Incident Triage](/Users/macbookpro/Development/recova-backend-v2/docs/operations/incident-triage.md)
- [Health and Readiness](/Users/macbookpro/Development/recova-backend-v2/docs/operations/health-readiness.md)

## Source Reference

- [Fiber RequestID Middleware](https://docs.gofiber.io/middleware/requestid/)
- [Fiber Logger Middleware](https://docs.gofiber.io/middleware/logger/)
- [OpenTelemetry Go Instrumentation](https://opentelemetry.io/docs/languages/go/instrumentation/)
