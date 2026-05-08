---
title: AI Coach Module
description: Kontrak modul AI Coach untuk chat pendampingan, riwayat percakapan, ringkasan progres, dan analisis onboarding dengan kontrol keamanan data.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/modules/ai-coach.md
last_reviewed: 2026-05-08
---

# AI Coach Module

## Responsibility

- menerima prompt pengguna,
- memproses permintaan ke provider AI melalui abstraction layer,
- menyajikan chat history dan ringkasan,
- menjaga safety dan privasi data AI.

## API Contract

Route prefix:

```text
/api/v1/ai
```

| Method | Path                             | Auth class | Purpose                          |
| ------ | -------------------------------- | ---------- | -------------------------------- |
| `POST` | `/api/v1/ai/ask-coach`           | Bearer     | kirim pertanyaan ke AI coach     |
| `GET`  | `/api/v1/ai/chat-history`        | Bearer     | ambil riwayat percakapan         |
| `GET`  | `/api/v1/ai/summary`             | Bearer     | ambil ringkasan progres pengguna |
| `POST` | `/api/v1/ai/onboarding-analysis` | Bearer     | analisis onboarding pengguna     |

## Database Model

Entitas utama:

- `ai_chat_messages` atau penyimpanan histori sejenis,
- metadata request AI (`provider`, `model`, `latency`, `status`),
- relasi riwayat AI ke `user_id`.

Constraint minimum:

- data AI terikat ke pemilik akun,
- retensi histori mengikuti kebijakan privasi.

## Authentication and Authorization

- seluruh endpoint AI wajib bearer auth,
- data AI hanya untuk user pemilik,
- user tidak boleh membaca histori user lain.

## Service and Business Rules

- handler wajib memanggil abstraction provider, bukan SDK vendor langsung,
- timeout request AI wajib eksplisit,
- retry default konservatif untuk hindari duplikasi side effect,
- safety filtering dijalankan sebelum respons dikirim.

## Validation Rules

- prompt tidak boleh kosong,
- batas panjang prompt wajib ditegakkan,
- parameter mode/opsi harus whitelist,
- input invalid dipetakan ke `VALIDATION_ERROR`.

## Error Contract

| Condition              | HTTP        | Error code         |
| ---------------------- | ----------- | ------------------ |
| auth invalid/missing   | `401`       | `UNAUTHENTICATED`  |
| payload invalid        | `422`       | `VALIDATION_ERROR` |
| provider timeout/error | `502`/`503` | `DOWNSTREAM_ERROR` |
| akses tidak diizinkan  | `403`       | `FORBIDDEN`        |
| kegagalan internal     | `500`       | `INTERNAL_ERROR`   |

## Observability Contract

Log field minimum:

- `request_id`,
- `user_id`,
- `provider`,
- `model`,
- `ai_action`,
- `status_code`.

Metrik minimum:

- request success rate,
- provider error rate,
- timeout rate,
- p95 latency AI endpoint.

Prompt mentah dan data sensitif tidak boleh dicatat di log umum.

## Testing Requirements

- unit test validator prompt,
- unit test mapping error provider,
- handler test auth + ownership,
- integration test provider abstraction dengan mock,
- contract test response AI dan error envelope.

## Open Gaps

- retensi final histori chat,
- aturan final fallback multi-provider,
- format final metadata confidence output.

## Related Documents

- [AI Provider Integration](/Users/macbookpro/Development/recova-backend-v2/docs/integrations/ai-provider.md)
- [AI Safety Operations](/Users/macbookpro/Development/recova-backend-v2/docs/operations/ai-safety.md)
- [Data Sensitivity Matrix](/Users/macbookpro/Development/recova-backend-v2/docs/references/data-sensitivity-matrix.md)

## Source Reference

- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
- [OpenAI API Overview](https://platform.openai.com/docs/overview)
- [Gemini API Overview](https://ai.google.dev/gemini-api/docs)
