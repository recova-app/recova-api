---
title: Recova Backend Data Retention Operations
description: Kebijakan retensi data layanan Recova Backend meliputi masa simpan, aturan purge, dan jejak audit untuk setiap kategori data.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/operations/data-retention.md
last_reviewed: 2026-05-08
---

# Recova Backend Data Retention Operations

Dokumen ini menetapkan masa simpan data, strategi purge, dan pembuktian audit untuk menjaga kepatuhan operasional dan minimisasi data.

## Retention Principles

- retensi berbasis kebutuhan operasional dan keamanan,
- hindari penyimpanan permanen untuk data yang tidak lagi dibutuhkan,
- penghapusan harus dapat diaudit,
- backup retention dipisahkan dari retention data aplikasi utama.

## Retention Schedule

| Kategori data            | Contoh                                | Masa simpan target                          | Mekanisme hapus                                          |
| ------------------------ | ------------------------------------- | ------------------------------------------- | -------------------------------------------------------- |
| OAuth/account identity   | identitas user, mapping auth          | selama akun aktif                           | hapus saat akun dihapus permanen                         |
| Recovery profile         | alasan pemulihan, preferensi check-in | selama akun aktif                           | hapus saat reset data atau account deletion              |
| Check-in and streak data | mood, statistik rutin                 | selama akun aktif                           | hapus saat reset data atau account deletion              |
| Journal content          | entri jurnal                          | selama akun aktif                           | hapus saat reset data atau account deletion              |
| Community content        | post/comment/like                     | selama konten aktif sesuai kebijakan produk | soft-delete lalu hard-delete sesuai keputusan moderation |
| AI coach chat context    | percakapan user dengan AI coach       | retensi terbatas sesuai kebutuhan fitur     | purge berkala otomatis + purge saat account deletion     |
| Access and audit logs    | auth event, request metadata          | retensi terbatas operasional                | purge terjadwal sesuai window observability              |
| Incident evidence logs   | event insiden terpilih                | sampai insiden ditutup + review selesai     | purge setelah masa simpan forensik selesai               |

## Deletion Modes

- `soft delete`: item disembunyikan dari API namun masih dapat dipulihkan dalam window terbatas,
- `hard delete`: data dihapus permanen dari storage utama,
- `anonymize`: identifier personal dihapus/diganti agar analitik tetap bisa dipakai tanpa identitas pengguna.

Setiap domain wajib mendokumentasikan mode delete yang dipakai dan alasan pemilihannya.

## Trigger Conditions

Penghapusan data dapat dipicu oleh:

- permintaan reset data pengguna,
- penghapusan akun permanen,
- jadwal purge berkala berbasis retention window,
- keputusan moderator atau kebijakan konten.

## Backup Retention Boundary

- backup database dipertahankan sesuai kebijakan disaster recovery,
- data yang terhapus di primary dapat masih berada di backup sampai window backup selesai,
- restore backup harus mempertahankan kebijakan redaksi log dan akses data sensitif.

## Retention Execution Controls

- purge job berjalan idempotent dan memiliki dry-run mode untuk verifikasi,
- purge wajib mencatat ringkasan hasil (jumlah item, domain, timestamp) tanpa payload sensitif,
- kegagalan purge harus masuk alert operasional.

## Evidence and Audit

Bukti minimal untuk setiap siklus purge:

- waktu eksekusi,
- cakupan domain data,
- jumlah record yang diproses,
- status sukses/gagal,
- owner penanggung jawab.

## Verification Checklist

- retention schedule terdokumentasi untuk semua domain data utama,
- operasi reset/delete memicu alur hapus yang sesuai,
- backup retention dan primary retention tidak saling bertentangan,
- audit bukti purge tersedia dan dapat ditinjau.

## Related Documents

- [Data Privacy Operations](/Users/macbookpro/Development/recova-backend-v2/docs/operations/data-privacy.md)
- [Database Migrations](/Users/macbookpro/Development/recova-backend-v2/docs/operations/database-migrations.md)
- [Database Seeding](/Users/macbookpro/Development/recova-backend-v2/docs/operations/database-seeding.md)
- [Incident Triage](/Users/macbookpro/Development/recova-backend-v2/docs/operations/incident-triage.md)

## Source Reference

- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
- [/Users/macbookpro/Development/bisakerja-api/docs/operations/security.md](/Users/macbookpro/Development/bisakerja-api/docs/operations/security.md)
- [OWASP Logging Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Logging_Cheat_Sheet.html)
