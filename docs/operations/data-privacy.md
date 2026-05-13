---
title: Recova Backend Data Privacy Operations
description: Kebijakan operasional perlindungan data pengguna mencakup klasifikasi data sensitif, kontrol akses, penggunaan data, dan proses penghapusan.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/operations/data-privacy.md
last_reviewed: 2026-05-08
---

# Recova Backend Data Privacy Operations

Dokumen ini mendefinisikan aturan operasional perlindungan data pengguna untuk seluruh domain API.

## Privacy Objectives

- melindungi data sensitif dari akses tidak sah,
- memastikan data diproses hanya untuk tujuan produk yang terdokumentasi,
- menekan paparan data pada log, metrics, dan audit trail,
- menjaga proses reset dan delete tetap konsisten serta terverifikasi.

## Data Classification Matrix

| Domain data        | Contoh field                                | Klasifikasi      | Aturan minimum                                                                         |
| ------------------ | ------------------------------------------- | ---------------- | -------------------------------------------------------------------------------------- |
| OAuth identity     | `google_sub`, email, nama profil            | confidential     | enkripsi in-transit, kontrol akses berbasis user ownership, tidak muncul di log mentah |
| Recovery profile   | alasan pemulihan, target check-in, timezone | restricted       | akses hanya user terkait dan service internal yang berwenang                           |
| Daily check-in     | mood, catatan check-in                      | restricted       | tidak dicetak ke log aplikasi sebagai payload penuh                                    |
| Journal            | judul/isi jurnal                            | highly-sensitive | tidak boleh ada logging body mentah, export hanya atas permintaan user                 |
| Community content  | post, comment, like                         | internal         | tampil publik sesuai aturan produk, audit moderation wajib                             |
| AI coach chat      | prompt user, response model                 | highly-sensitive | redaksi ketat pada log dan observability event                                         |
| Security/audit log | auth event, status route, request id        | internal         | simpan metadata operasional, hindari konten pribadi mentah                             |

## Privacy-by-Default Controls

- minimisasi data: simpan hanya field yang diperlukan untuk fungsi produk,
- minimisasi akses: setiap query write/read harus berbasis user ownership,
- minimisasi observability: log berisi metadata teknis, bukan konten pribadi mentah,
- minimisasi retensi: data tidak disimpan lebih lama dari kebutuhan operasional.

## Access Control Rules

- endpoint user hanya boleh membaca/menulis data milik user yang terautentikasi,
- service-to-service credential tidak boleh dipakai sebagai identitas end-user,
- akses data sensitif oleh operator harus berbasis tiket insiden dan dicatat,
- data highly-sensitive tidak boleh diekspor massal tanpa persetujuan owner.

## Data Processing Boundaries

- input pengguna diproses di backend API sebelum diteruskan ke dependency eksternal,
- payload ke provider eksternal dibatasi ke konteks minimum yang dibutuhkan,
- data pribadi tidak boleh dipindahkan ke analytics stream tanpa anonimisasi/pseudonimisasi.

## Reset and Delete Behavior

Untuk operasi `reset` dan `delete` user:

- definisikan scope data yang dihapus per modul (profil, check-in, streak, jurnal, komunitas, chat),
- catat event penghapusan sebagai audit metadata tanpa menyimpan isi data yang dihapus,
- pastikan API response tidak menampilkan kembali data yang sudah dihapus,
- proses backup purge mengikuti kebijakan retensi backup pada dokumen retensi.

## Privacy Incident Handling

Jika ada indikasi kebocoran data:

1. isolasi jalur yang terpapar,
2. nonaktifkan log atau collector yang menyebabkan kebocoran,
3. rotasi credential jika ada indikasi secret exposure,
4. jalankan investigasi berbasis request id dan event metadata,
5. dokumentasikan corrective action pada runbook insiden.

## Verification Checklist

- endpoint sensitif tidak mencetak body mentah di log,
- data owner enforcement tervalidasi lewat negative test unauthorized/forbidden,
- proses reset/delete mencakup semua domain data pengguna,
- log pipeline menerapkan redaksi sesuai [Log Redaction Policy](/Users/macbookpro/Development/recova-backend-v2/docs/standards/log-redaction-policy.md).

## Related Documents

- [Data Retention Operations](/Users/macbookpro/Development/recova-backend-v2/docs/operations/data-retention.md)
- [Security Operations Baseline](/Users/macbookpro/Development/recova-backend-v2/docs/operations/security.md)
- [Secure Coding Standard](/Users/macbookpro/Development/recova-backend-v2/docs/standards/secure-coding.md)
- [Incident Triage](/Users/macbookpro/Development/recova-backend-v2/docs/operations/incident-triage.md)

## Source Reference

- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
- [/Users/macbookpro/Development/bisakerja-api/docs/operations/security.md](/Users/macbookpro/Development/bisakerja-api/docs/operations/security.md)
- [OWASP Logging Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Logging_Cheat_Sheet.html)
