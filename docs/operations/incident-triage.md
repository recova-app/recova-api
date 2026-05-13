---
title: Recova Backend Incident Triage
description: Prosedur triage insiden untuk identifikasi cepat, klasifikasi dampak, stabilisasi layanan, dan eskalasi berbasis sinyal observability.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/operations/incident-triage.md
last_reviewed: 2026-05-08
---

# Recova Backend Incident Triage

Dokumen ini menjadi runbook triage pertama saat insiden terdeteksi.

## Triage Objectives

- konfirmasi insiden nyata vs false alert,
- ukur dampak user dan area fitur terdampak,
- stabilisasi layanan secepat mungkin,
- siapkan handoff investigasi akar masalah.

## First 15 Minutes Workflow

1. konfirmasi alert dari minimal dua sinyal (metric + log/trace),
2. tetapkan severity (`P1`, `P2`, `P3`),
3. identifikasi endpoint/dependency terdampak,
4. lakukan mitigasi awal (rollback, traffic shaping, atau isolate dependency),
5. dokumentasikan timeline tindakan.

## Severity Guide

| Severity | Dampak                                    | SLA respons awal        |
| -------- | ----------------------------------------- | ----------------------- |
| P1       | mayoritas endpoint inti tidak tersedia    | segera (<= 5 menit)     |
| P2       | fungsi inti terganggu sebagian signifikan | cepat (<= 15 menit)     |
| P3       | gangguan minor/non-kritis                 | terjadwal (<= 60 menit) |

## Signal-to-Action Mapping

| Signal                           | Tindakan awal                                 |
| -------------------------------- | --------------------------------------------- |
| readiness `503` beruntun         | cek koneksi DB, hentikan deploy aktif         |
| `5xx` naik tajam setelah release | bandingkan diff rilis, pertimbangkan rollback |
| `401`/`403` anomaly              | verifikasi auth config dan secret alignment   |
| timeout AI provider naik         | aktifkan degrade mode untuk fitur AI          |
| `429` spike                      | evaluasi abuse traffic dan tuning limiter     |

## Evidence Collection

Minimal artefak per insiden:

- rentang waktu insiden,
- request id contoh gagal,
- endpoint dan status error dominan,
- dependency yang menunjukkan degradasi,
- tindakan mitigasi dan hasil.

## Communication Rules

- gunakan satu channel insiden aktif,
- update status berkala berdasarkan severity,
- hindari spekulasi penyebab sebelum bukti cukup,
- publikasikan dampak user dalam bahasa operasional yang jelas.

## Exit Criteria for Triage

Triage selesai bila:

- layanan stabil (error rate dan latency kembali dalam batas),
- owner investigasi akar masalah ditetapkan,
- tindakan lanjut tercatat (fix, test tambahan, atau hardening).

## Related Documents

- [Failure Scenarios](/Users/macbookpro/Development/recova-backend-v2/docs/operations/failure-scenarios.md)
- [Observability](/Users/macbookpro/Development/recova-backend-v2/docs/operations/observability.md)
- [Release Gates](/Users/macbookpro/Development/recova-backend-v2/docs/operations/release-gates.md)

## Source Reference

- [Google SRE Workbook - Incident Response](https://sre.google/workbook/incident-response/)
- [OpenTelemetry Documentation](https://opentelemetry.io/docs/)
