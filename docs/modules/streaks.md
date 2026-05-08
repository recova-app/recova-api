---
title: Streaks Module
description: Aturan perhitungan streak pengguna berdasarkan data check-in harian, boundary tanggal lokal, dan penanganan gap hari.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/modules/streaks.md
last_reviewed: 2026-05-08
---

# Streaks Module

Dokumen ini mendefinisikan aturan streak untuk menjaga konsistensi hasil statistik pengguna.

## Streak Definitions

- `current_streak`: jumlah hari beruntun hingga hari aktif terakhir,
- `longest_streak`: nilai streak tertinggi sepanjang histori,
- `streak_anchor_date`: tanggal lokal terakhir yang menjadi acuan streak.

## Core Calculation Rules

- check-in hari pertama memulai `current_streak = 1`,
- check-in pada hari lokal berikutnya meningkatkan streak (`+1`),
- gap lebih dari satu hari lokal mereset `current_streak` ke `1`,
- `longest_streak` diperbarui jika `current_streak` melampaui nilai lama.

## Timezone and Date Boundary

- streak dihitung dari `local_date` pengguna,
- perubahan timezone tidak boleh menyebabkan loncatan streak tak terjelaskan,
- aturan migrasi timezone harus terdokumentasi bila fitur timezone change diaktifkan.

## Duplicate and Late Event Handling

- duplicate check-in hari yang sama tidak boleh menambah streak dua kali,
- event terlambat harus diproses dengan aturan deterministik berdasarkan `local_date` final,
- reprocess data historis harus menjaga konsistensi `longest_streak`.

## Consistency and Concurrency

- update streak harus atomik dengan insert check-in,
- gunakan lock/transaksi untuk mencegah race pada submit paralel,
- nilai streak harus bisa dihitung ulang dari histori untuk audit/recovery.

## Related Documents

- [Routine Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/routine.md)
- [Check-Ins Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/check-ins.md)
- [Statistics Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/statistics.md)

## Source Reference

- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
