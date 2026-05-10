---
title: AI Coach Module
description: Kontrak modul AI Coach untuk chat pendampingan, riwayat percakapan, ringkasan progres, analisis onboarding, dan preferensi persona respons.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/modules/ai-coach.md
last_reviewed: 2026-05-09
---

# AI Coach Module

## Responsibility

- menerima prompt pengguna,
- memproses permintaan ke provider AI melalui abstraction layer,
- menyajikan chat history dan ringkasan,
- menyimpan preferensi persona AI per pengguna,
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
| `GET`  | `/api/v1/ai/persona-preferences` | Bearer     | ambil preferensi persona user    |
| `PUT`  | `/api/v1/ai/persona-preferences` | Bearer     | ubah preferensi persona user     |

## Database Model

Entitas utama:

- `ai_chats` untuk histori percakapan (`role=user|model`, `content`, `created_at`),
- `user_ai_persona_preferences` untuk konfigurasi persona user,
- metadata request AI (`provider`, `model`, `latency`, `status`),
- relasi riwayat AI ke `user_id`.

Constraint minimum:

- data AI terikat ke pemilik akun,
- preferensi persona terikat unik per `user_id`,
- retensi histori mengikuti kebijakan privasi.

## Authentication and Authorization

- seluruh endpoint AI wajib bearer auth,
- data AI hanya untuk user pemilik,
- user tidak boleh membaca histori user lain.

## Service and Business Rules

- handler wajib memanggil abstraction provider, bukan SDK vendor langsung,
- timeout request AI wajib eksplisit,
- fallback provider optional hanya dipakai untuk failure `timeout/unavailable`,
- safety filtering dijalankan sebelum respons dikirim.
- jika preferensi persona kosong, fallback ke persona default aman,
- persona aktif harus dipakai saat membangun system instruction AI,
- respons `ask-coach` menyertakan `persona_used` agar audit troubleshooting mudah.

## Validation Rules

- prompt tidak boleh kosong,
- batas panjang prompt wajib ditegakkan,
- enum persona wajib whitelist,
- parameter mode/opsi harus whitelist,
- input invalid dipetakan ke `VALIDATION_ERROR`.

Whitelist persona minimum:

- `supportive`,
- `friendly`,
- `concise`,
- `direct`.

## Error Contract

| Condition              | HTTP  | Error code            |
| ---------------------- | ----- | --------------------- |
| auth invalid/missing   | `401` | `UNAUTHENTICATED`     |
| payload invalid        | `422` | `VALIDATION_ERROR`    |
| provider invalid/error | `502` | `DOWNSTREAM_ERROR`    |
| provider timeout/down  | `503` | `SERVICE_UNAVAILABLE` |
| akses tidak diizinkan  | `403` | `FORBIDDEN`           |
| melebihi rate limit    | `429` | `RATE_LIMITED`        |
| kegagalan internal     | `500` | `INTERNAL_ERROR`      |

## Observability Contract

Log field minimum:

- `request_id`,
- `user_id`,
- `provider`,
- `model`,
- `persona`,
- `ai_action`,
- `status_code`.

Metrik minimum:

- request success rate,
- provider error rate,
- timeout rate,
- persona usage distribution,
- p95 latency AI endpoint.

Metrik persona usage direkomendasikan:

- `recova_ai_persona_usage_total{action,persona,status}`

Prompt mentah dan data sensitif tidak boleh dicatat di log umum.

## Testing Requirements

- unit test validator prompt,
- unit test validator persona preference,
- unit test mapping error provider,
- handler test auth + ownership,
- integration test provider abstraction dengan mock,
- contract test response AI dan error envelope.

## Open Gaps

- retensi final histori chat lintas environment,
- daftar final persona tambahan di luar baseline.

## Related Documents

- [AI Provider Integration](/Users/macbookpro/Development/recova-backend-v2/docs/integrations/ai-provider.md)
- [AI Safety Operations](/Users/macbookpro/Development/recova-backend-v2/docs/operations/ai-safety.md)
- [Data Sensitivity Matrix](/Users/macbookpro/Development/recova-backend-v2/docs/references/data-sensitivity-matrix.md)

## Source Reference

- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
- [OpenAI API Authentication](https://platform.openai.com/docs/api-reference/authentication?api-mode=responses)
- [OpenAI Chat Completions Endpoint](https://platform.openai.com/docs/api-reference/chat/create?lang=curl)
- [Gemini API Reference](https://ai.google.dev/api)
- [Gemini Generate Content](https://ai.google.dev/api/generate-content)
