---
title: AI Safety Operations
description: Runbook operasional untuk keamanan AI, redaksi data sensitif, guardrail input-output, retensi data, dan respons insiden AI.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/operations/ai-safety.md
last_reviewed: 2026-05-08
---

# AI Safety Operations

Dokumen ini mendefinisikan kontrol operasional minimum agar fitur AI Coach tetap aman dan privat.

## Safety Objectives

- mencegah kebocoran data sensitif pengguna,
- menahan input dan output berisiko,
- menjaga respons API tetap aman saat provider gagal,
- menyediakan jalur triase insiden AI yang dapat diaudit.

## Data Handling Rules

Aturan wajib data AI:

- jangan log prompt mentah pengguna,
- jangan log respons mentah provider pada log umum,
- jangan log konten jurnal mentah,
- simpan metadata operasional saja pada log aplikasi utama.

Metadata log minimum:

- `request_id`,
- `user_id`,
- `provider`,
- `model`,
- `latency_ms`,
- `status`.

## Input Guardrails

- batasi panjang input untuk mencegah abuse payload,
- validasi tipe dan struktur context sebelum request ke provider,
- tolak input berisi konten terlarang berat sesuai kebijakan moderasi internal,
- terapkan rate limit lebih ketat pada endpoint AI dibanding endpoint read biasa.

## Output Guardrails

- output AI harus melewati pemeriksaan safety sebelum dikirim ke client,
- jika output diklasifikasi berisiko, return error aman tanpa memantulkan teks sensitif,
- tambahkan metadata internal untuk audit keputusan safety.

## Retention and Deletion Direction

- histori chat AI dipertahankan hanya selama dibutuhkan produk,
- retensi default harus minimum dan terdokumentasi,
- penghapusan data user harus mencakup histori AI sesuai kebijakan reset/delete.

## Downstream Failure Policy

- timeout provider dipetakan ke `503 SERVICE_UNAVAILABLE`,
- invalid provider response dipetakan ke `502 DOWNSTREAM_ERROR`,
- fallback provider opsional hanya jika sudah dikonfigurasi,
- error detail internal provider tidak boleh dipaparkan ke client.

Implementasi runtime saat ini:

- fallback hanya dijalankan untuk kegagalan `timeout` atau `unavailable`,
- kegagalan validasi payload respons provider dihentikan sebagai `DOWNSTREAM_ERROR`,
- request logger hanya mencatat metadata route (tanpa body prompt/jurnal).

## Incident Triage Matrix

| Scenario                         | Signal                                          | First checks                                             |
| -------------------------------- | ----------------------------------------------- | -------------------------------------------------------- |
| lonjakan timeout provider        | kenaikan `SERVICE_UNAVAILABLE` pada endpoint AI | cek latency provider, timeout config, network            |
| kebocoran data sensitif pada log | temuan konten mentah pada log scanning          | isolasi log sink, patch redaction, rotate akses          |
| output berisiko lolos ke client  | laporan safety atau audit anomali               | cek rules output filter, matikan feature flag bila perlu |
| error mapping tidak konsisten    | status code mismatch di telemetry               | validasi mapper error dan classifier                     |

## Operational Checklist

- verifikasi env key AI tersedia dan non-empty,
- verifikasi redaction rule aktif,
- verifikasi timeout AI aktif,
- verifikasi fallback behavior sesuai kebijakan,
- verifikasi endpoint AI tercakup rate limit,
- verifikasi tidak ada prompt mentah pada log sampling.

## Related Documents

- [AI Coach Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/ai-coach.md)
- [AI Provider Integration](/Users/macbookpro/Development/recova-backend-v2/docs/integrations/ai-provider.md)
- [Data Sensitivity Matrix](/Users/macbookpro/Development/recova-backend-v2/docs/references/data-sensitivity-matrix.md)
- [Error Taxonomy](/Users/macbookpro/Development/recova-backend-v2/docs/standards/error-taxonomy.md)

## Source Reference

- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
- [OpenAI API Authentication](https://developers.openai.com/api/reference/overview#authentication)
- [Gemini API OAuth Quickstart](https://ai.google.dev/gemini-api/docs/oauth)
