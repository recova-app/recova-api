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
last_reviewed: 2026-05-14
---

# Streaks Module

Dokumen ini mendefinisikan aturan streak untuk menjaga konsistensi hasil statistik pengguna.

## Streak Definitions

- `current_streak`: jumlah hari beruntun hingga hari aktif terakhir,
- `longest_streak`: nilai streak tertinggi sepanjang histori,
- `streak_anchor_date`: tanggal lokal terakhir yang menjadi acuan streak,
- `streak_goal_comparison`: pembanding streak terhadap target `porn_free_goal` user.

## Core Calculation Rules

- hari hanya dihitung sebagai streak bila ada check-in sukses **dan** tidak ada relapse pada tanggal UTC yang sama,
- check-in hari pertama memulai `current_streak = 1` jika hari tersebut bebas relapse,
- check-in pada hari UTC berikutnya meningkatkan streak (`+1`) jika hari itu bebas relapse,
- gap lebih dari satu hari lokal mereset `current_streak` ke `1`,
- `longest_streak` diperbarui jika `current_streak` melampaui nilai lama.
- relapse pada hari berjalan menutup streak aktif hari itu dan `current_streak` menjadi `0`.

## Timezone and Date Boundary

- streak dihitung dari `check_in_date` berbasis UTC,
- perubahan timezone client tidak mengubah boundary streak karena sumber kebenaran waktu berada di UTC.

## Duplicate and Late Event Handling

- duplicate check-in hari yang sama tidak boleh menambah streak dua kali,
- check-in sukses dan relapse pada hari UTC yang sama harus memperlakukan hari tersebut sebagai hari relapse (bukan hari streak),
- relapse dicatat lebih dulu lalu check-in sukses di hari yang sama tetap tidak membuka streak untuk hari itu,
- event terlambat harus diproses dengan aturan deterministik berdasarkan `local_date` final,
- reprocess data historis harus menjaga konsistensi `longest_streak`.

## Consistency and Concurrency

- update streak harus atomik dengan insert check-in,
- gunakan lock/transaksi untuk mencegah race pada submit paralel,
- nilai streak harus bisa dihitung ulang dari histori untuk audit/recovery.
- field pembanding goal (`goal_reached`, `remaining_days`, `progress_rate`) harus dihitung dari `current_streak` versus `porn_free_goal`.
- sumber relapse legacy (`check_ins.is_successful=false`) dan tabel `relapses` tidak boleh menyebabkan double count event relapse pada tanggal UTC yang sama.

## Related Documents

- [Routine Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/routine.md)
- [Check-Ins Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/check-ins.md)
- [Statistics Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/statistics.md)

## Source Reference

- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
