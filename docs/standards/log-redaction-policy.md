---
title: Recova Backend Log Redaction Policy
description: Standar redaksi log untuk mencegah kebocoran data sensitif melalui application logs, audit logs, dan telemetry pipeline.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/standards/log-redaction-policy.md
last_reviewed: 2026-05-08
---

# Recova Backend Log Redaction Policy

Dokumen ini menetapkan aturan wajib redaksi data pada seluruh jalur logging.

## Policy Objectives

- mencegah data sensitif tampil pada log mentah,
- menjaga log tetap berguna untuk debugging dan incident response,
- menyelaraskan payload log dengan model observability terstruktur.

## Data Never Logged

Data berikut tidak boleh muncul dalam bentuk mentah:

- password, hash password, reset token, OTP,
- access token, refresh token, session credential,
- secret key, API key, private key, connection string penuh,
- isi jurnal mentah dan isi percakapan AI mentah,
- informasi identitas sensitif yang tidak diperlukan untuk diagnosis.

## Redaction Strategy Matrix

| Tipe data                | Metode                | Contoh output                  |
| ------------------------ | --------------------- | ------------------------------ |
| Secret/token             | drop penuh            | `[REDACTED_SECRET]`            |
| Identifier sensitif      | hashing stabil        | `user_hash=sha256:...`         |
| Email/nomor kontak       | masking parsial       | `u***@domain.com`              |
| Konten teks pengguna     | truncation + masking  | `content=[REDACTED_USER_TEXT]` |
| Payload request sensitif | metadata-only logging | method, route, status, latency |

## Structured Logging Contract

Field minimum yang aman untuk observability:

- `timestamp`,
- `level`,
- `requestId`,
- `route`,
- `method`,
- `statusCode`,
- `latencyMs`,
- `errorCode` (jika gagal),
- `dependency` (jika error dari downstream).

Konten request/response body hanya boleh dicatat jika endpoint tidak memproses data sensitif dan sudah disetujui owner.

## Enforcement Points

Redaksi harus diterapkan di tiga titik:

1. application logger sebelum event keluar dari service,
2. log collector/forwarder sebelum sink eksternal,
3. SIEM/query layer untuk mencegah kebocoran data historis.

## Route Sensitivity Rules

- auth routes: log metadata-only, tanpa body,
- journal dan AI coach routes: log metadata-only, tanpa prompt/isi,
- community routes: hindari logging konten mentah kecuali untuk moderasi terkontrol,
- admin/internal routes: wajib redaksi ketat karena akses elevated.

## Testing and Verification

Uji otomatis wajib mencakup:

- assertion bahwa secret/token tidak pernah muncul di log output,
- assertion redaksi untuk error path,
- assertion format log tetap valid JSON/structured format setelah redaksi.

Uji manual rutin:

- sampling log produksi/staging berbasis query kata kunci sensitif,
- review dashboard observability untuk field yang berpotensi bocor.

## Incident Handling for Redaction Failure

Jika kebocoran log ditemukan:

1. hentikan stream log yang bocor,
2. rotasi credential terdampak,
3. hapus atau isolasi data log sensitif pada sink,
4. patch redaction rule dan validasi ulang,
5. dokumentasikan RCA dan prevention rule.

## Related Documents

- [Data Privacy Operations](/Users/macbookpro/Development/recova-backend-v2/docs/operations/data-privacy.md)
- [Security Operations Baseline](/Users/macbookpro/Development/recova-backend-v2/docs/operations/security.md)
- [Observability](/Users/macbookpro/Development/recova-backend-v2/docs/operations/observability.md)
- [Error Handling Standard](/Users/macbookpro/Development/recova-backend-v2/docs/standards/error-handling.md)

## Source Reference

- [OWASP Logging Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Logging_Cheat_Sheet.html)
- [OpenTelemetry Logs Data Model](https://opentelemetry.io/docs/specs/otel/logs/data-model/)
- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
