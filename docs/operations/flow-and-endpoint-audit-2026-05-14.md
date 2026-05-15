---
title: Flow and Endpoint Audit 2026-05-14
description: Audit menyeluruh flow runtime, keamanan, potensi bug/error, serta kontrak query-response endpoint Recova Backend.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/operations/flow-and-endpoint-audit-2026-05-14.md
last_reviewed: 2026-05-14
---

# Flow and Endpoint Audit 2026-05-14

## Scope

- Runtime flow API publik (`/health/*`, `/openapi.yaml`, `/docs/api`, `/metrics`, dan `/api/v1/*`).
- Alur auth, authorization, validation, error mapping, rate limit, observability.
- Kontrak query/response endpoint dari `api/openapi/openapi.yaml` + `docs/generated/openapi.yaml`.

## Verification Evidence

- `rtk make openapi-check` -> pass.
- `rtk go test ./...` -> pass (`295 passed in 37 packages`).
- `rtk go test ./test/contract/...` -> pass (`44 passed in 1 packages`).
- `rtk make module-consistency-check` -> pass.
- `rtk make security-scan` -> pass (`govulncheck`: no direct code vulnerability).
- `rtk make test-e2e` -> not executed fully (`RECOVA_DB_INTEGRATION_URL` belum diset).

## Flow Audit Result

### 1. Auth and Trust Boundary

- Semua route produk non-public memakai `RequireAuth`.
- Token access divalidasi lewat JWT claim (`issuer`, `audience`, `exp`, `token_type`).
- Refresh token disimpan sebagai hash (`sha256`) dan dirotasi atomik.
- Refresh/logout flow idempotent dan aman terhadap token kosong/invalid.
- CORS + Helmet + request ID + recover middleware aktif.

### 2. Input Validation and Error Contract

- Handler layer tidak bypass service validation.
- Service layer konsisten mengembalikan `errs.AppError` -> dipetakan ke envelope standar.
- Error envelope konsisten: `success=false`, `error.code`, `error.details`, `error.request_id`.
- Request body/query utama memiliki guard batas ukuran/range/string length.

### 3. Abuse Protection

- Rate limiter aktif pada auth, AI, dan write-community endpoint.
- Key limiter user-aware (`userID`) bila auth tersedia, fallback ke IP.
- Endpoint reset data dibatasi environment development/local.

### 4. Observability and Data Exposure

- Structured logging aktif, atribut sensitif direduksi (`token`, `secret`, `password`, `cookie`, `prompt`, `journal`).
- Audit event direkam untuk route sensitif.
- Telemetry AI tidak menyimpan prompt/response mentah pada metrics label.

## Endpoint Query/Response Audit Result

### Coverage

- Total route runtime: `39` (sesuai `docs/generated/routes.md`).
- Route API produk utama `/api/v1/*`: tervalidasi terhadap OpenAPI contract.
- Request schema, query schema, success response, dan error response sudah terdefinisi di OpenAPI.
- Contract test memverifikasi request/response terhadap spec (`openapi3filter`).

### Fixed During This Audit

1. **AI chat role contract mismatch**

- Problem: response `/api/v1/ai/chat-history` bisa mengembalikan `role=model`, sedangkan kontrak OpenAPI mengizinkan `user|assistant`.
- Fix: normalisasi role ke `assistant` untuk `model`; role AI baru juga disimpan sebagai `assistant`.
- Files:
  - `internal/modules/ai/service.go`
  - `internal/modules/ai/service_test.go`

2. **Onboarding analysis output hardening**

- Problem: parser tidak memvalidasi enum `level` (`Low|Moderate|High`).
- Fix: tambah validasi level; invalid response kini gagal aman sebagai downstream error.
- Files:
  - `internal/modules/ai/service.go`
  - `internal/modules/ai/service_test.go`

3. **OpenAPI example drift vs runtime behavior**

- Problem: beberapa contoh query/response tidak selaras runtime.
- Fix: sinkronisasi source OpenAPI lalu regenerate artifact.
- Files:
  - `api/openapi/openapi.yaml`
  - `docs/generated/openapi.yaml`
- Drift yang diperbaiki:
  - pesan konflik onboarding,
  - detail field users settings invalid payload,
  - `activity-summary` activity type example,
  - detail field invalid reply depth,
  - fallback persona example.

## Residual Risk (Non-blocking)

- Endpoint `/metrics`, `/openapi.yaml`, dan `/docs/api` masih publik by design.
- Untuk deployment production internet-facing, disarankan proteksi network layer (allowlist/reverse proxy/auth gateway) sesuai kebijakan operasional.

## Related Documents

- [API Reference](/Users/macbookpro/Development/recova-backend-v2/docs/api-reference.md)
- [Generated OpenAPI](/Users/macbookpro/Development/recova-backend-v2/docs/generated/openapi.yaml)
- [Route Inventory](/Users/macbookpro/Development/recova-backend-v2/docs/generated/routes.md)
- [Security Operations](/Users/macbookpro/Development/recova-backend-v2/docs/operations/security.md)
