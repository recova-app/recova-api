---
title: AI Provider Integration
description: Kontrak integrasi provider AI untuk routing request, konfigurasi model, timeout, fallback, error mapping, dan pengamanan credential.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/integrations/ai-provider.md
last_reviewed: 2026-05-08
---

# AI Provider Integration

Dokumen ini menetapkan boundary integrasi provider AI yang dipakai oleh modul AI Coach.

## Integration Boundary

| Area                                     | Owner          |
| ---------------------------------------- | -------------- |
| autentikasi user aplikasi                | Recova Backend |
| validasi input chat dan context shaping  | Recova Backend |
| pemanggilan model AI                     | AI Provider    |
| mapping hasil AI ke response produk      | Recova Backend |
| logging dan redaction payload sensitif   | Recova Backend |
| secret management API key/provider token | Recova Backend |

Frontend tidak boleh memanggil provider AI secara langsung.

## Provider Abstraction Contract

Backend harus memakai abstraction internal dengan kontrak minimum:

- `GenerateCoachReply(context)`,
- `GenerateCheckInSummary(context)`,
- `GenerateOnboardingAnalysis(context)`.

Setiap implementasi provider wajib:

- menerima context yang sudah disanitasi,
- mengembalikan response terstruktur,
- mengembalikan error terklasifikasi (`timeout`, `unavailable`, `invalid_response`, `policy_block`).

## Supported Provider Families

Arah kompatibilitas:

- provider berbasis OpenAI-compatible API,
- provider berbasis Gemini API.

Pemilihan provider dilakukan oleh konfigurasi environment dan tidak mengubah contract endpoint publik.

## Environment Contract

Nama variabel minimum:

- `AI_PROVIDER`,
- `AI_MODEL`,
- `AI_TIMEOUT_MS`,
- `AI_API_KEY`,
- `AI_BASE_URL` (opsional bila provider non-default),
- `AI_FALLBACK_PROVIDER` (opsional),
- `AI_FALLBACK_MODEL` (opsional).

Aturan:

- variabel required harus fail-fast saat startup jika kosong/tidak valid,
- `AI_API_KEY` tidak boleh dicetak ke log,
- perubahan provider harus dapat ditelusuri melalui startup logs aman tanpa menampilkan secret.

## Payload Minimization Rules

- kirim hanya data yang diperlukan untuk inferensi,
- hindari pengiriman data sensitif yang tidak relevan,
- gunakan ringkasan data check-in/onboarding bila cukup,
- konten jurnal mentah hanya boleh dikirim jika memang diperlukan oleh use case dan sudah disetujui kebijakan privasi.

## Timeout and Fallback Strategy

- timeout request AI wajib eksplisit per request,
- jika primary provider timeout/unavailable, fallback optional dapat dipakai bila dikonfigurasi,
- fallback tidak boleh mengubah envelope response publik,
- fallback event harus dilog sebagai metadata operasional.

## Error Mapping

| Provider failure                | API code              | HTTP status |
| ------------------------------- | --------------------- | ----------- |
| koneksi ditolak                 | `SERVICE_UNAVAILABLE` | `503`       |
| timeout                         | `SERVICE_UNAVAILABLE` | `503`       |
| respons provider invalid schema | `DOWNSTREAM_ERROR`    | `502`       |
| kredensial provider invalid     | `DOWNSTREAM_ERROR`    | `502`       |
| pembatasan provider/rate limit  | `SERVICE_UNAVAILABLE` | `503`       |

## Security Rules

- API key provider hanya tersedia pada sisi backend,
- kunci tidak boleh dikirim ke frontend,
- request/response raw dari provider tidak boleh langsung dipantulkan ke klien,
- gunakan request id untuk korelasi troubleshooting.

## Operational Gaps

- kebijakan kuota dan budget lintas provider,
- strategi cache respons AI untuk permintaan identik,
- kebijakan regional routing untuk latency.

## Related Documents

- [AI Coach Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/ai-coach.md)
- [Environment Configuration](/Users/macbookpro/Development/recova-backend-v2/docs/environment.md)
- [AI Safety Operations](/Users/macbookpro/Development/recova-backend-v2/docs/operations/ai-safety.md)

## Source Reference

- [OpenAI API Authentication](https://developers.openai.com/api/reference/overview#authentication)
- [Gemini API OAuth Quickstart](https://ai.google.dev/gemini-api/docs/oauth)
- [/Users/macbookpro/Development/bisakerja-api/docs/integrations/model-api.md](/Users/macbookpro/Development/bisakerja-api/docs/integrations/model-api.md)
