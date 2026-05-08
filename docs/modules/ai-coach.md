---
title: AI Coach Module
description: Kontrak modul AI Coach untuk chat, histori percakapan, ringkasan check-in, analisis onboarding, serta kontrol keamanan dan privasi.
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

Modul AI Coach menyediakan dukungan percakapan, ringkasan progres, dan analisis onboarding berbasis model AI dengan kontrol privasi ketat.

## Responsibility

Modul AI Coach bertanggung jawab pada:

- menerima permintaan chat pengguna,
- mengambil histori chat pengguna,
- menghasilkan ringkasan check-in harian,
- menghasilkan analisis onboarding,
- memetakan kegagalan provider AI ke error API aman.

Modul AI Coach tidak bertanggung jawab pada:

- autentikasi pengguna,
- manajemen profil non-AI,
- manajemen konten komunitas.

## Route Prefix

```text
/api/v1/ai
```

## Endpoint Summary

| Method | Path                             | Auth   | Purpose                                        |
| ------ | -------------------------------- | ------ | ---------------------------------------------- |
| `POST` | `/api/v1/ai/ask-coach`           | Bearer | Kirim prompt pengguna dan terima respons coach |
| `GET`  | `/api/v1/ai/chat-history`        | Bearer | Ambil histori percakapan pengguna              |
| `GET`  | `/api/v1/ai/summary`             | Bearer | Ambil ringkasan progres check-in               |
| `POST` | `/api/v1/ai/onboarding-analysis` | Bearer | Analisis data onboarding untuk insight awal    |

## Request and Response Contract

Kontrak minimum `ask-coach`:

- request memuat pesan pengguna dan konteks aman minimum,
- response memuat jawaban coach, metadata request, dan indikator sumber model,
- response tidak boleh memuat credential provider atau detail internal prompt template.

Kontrak minimum `chat-history`:

- hanya mengembalikan percakapan milik pengguna terautentikasi,
- payload dapat dipaginasi,
- data sensitif dimask jika ada kebijakan redaksi tambahan.

Kontrak minimum `summary` dan `onboarding-analysis`:

- memakai data internal pengguna yang tervalidasi,
- menyajikan hasil ringkas yang aman untuk klien,
- kesalahan downstream dipetakan ke error standar.

## Ownership and Access Rules

- seluruh endpoint AI wajib bearer auth,
- `user_id` diambil dari auth context, bukan dari payload klien,
- pengguna tidak dapat membaca histori AI milik akun lain.

## Provider Abstraction Rules

- modul AI Coach harus memanggil interface provider internal, bukan SDK provider langsung dari handler,
- provider dapat diganti antar implementasi kompatibel tanpa mengubah kontrak endpoint publik,
- model, timeout, dan base URL provider dikendalikan lewat konfigurasi environment tervalidasi.

## Timeout, Retry, and Fallback

Aturan baseline:

- setiap request AI memiliki timeout eksplisit,
- retry default adalah `0` untuk mencegah duplikasi side effect percakapan,
- fallback ke provider cadangan hanya aktif jika dikonfigurasi,
- jika downstream gagal, API mengembalikan `503 SERVICE_UNAVAILABLE` atau `502 DOWNSTREAM_ERROR` sesuai klasifikasi.

## Privacy and Logging Rules

Aturan wajib:

- prompt mentah pengguna tidak dicatat pada log umum,
- konten jurnal pengguna tidak boleh dimasukkan mentah ke log AI,
- chat history disimpan dengan retensi minimum yang disetujui,
- audit log menyimpan metadata (`request_id`, `user_id`, `provider`, status), bukan isi sensitif.

## Safety Rules

- lakukan input guardrail untuk mengurangi abuse prompt,
- lakukan output filtering untuk mencegah respons berisiko,
- jika output ditahan oleh safety policy, return error aman dan tidak mengungkap detail kebijakan internal.

## Open Gaps

- durasi retensi final histori chat AI,
- kebijakan final fallback multi-provider,
- format final skor confidence/quality pada respons AI.

## Related Documents

- [AI Provider Integration](/Users/macbookpro/Development/recova-backend-v2/docs/integrations/ai-provider.md)
- [AI Safety Operations](/Users/macbookpro/Development/recova-backend-v2/docs/operations/ai-safety.md)
- [Data Sensitivity Matrix](/Users/macbookpro/Development/recova-backend-v2/docs/references/data-sensitivity-matrix.md)
- [API Response Standard](/Users/macbookpro/Development/recova-backend-v2/docs/api-response-standard.md)

## Source Reference

- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
- [/Users/macbookpro/Development/bisakerja-api/docs/integrations/model-api.md](/Users/macbookpro/Development/bisakerja-api/docs/integrations/model-api.md)
- [OpenAI API Authentication](https://developers.openai.com/api/reference/overview#authentication)
- [Gemini API OAuth Quickstart](https://ai.google.dev/gemini-api/docs/oauth)
