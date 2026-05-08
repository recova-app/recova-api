---
title: Recova Backend Compatibility Matrix
description: Matriks target kompatibilitas API dan perilaku runtime untuk menjaga kontrak klien tetap stabil selama migrasi backend.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/roadmap/compatibility-matrix.md
last_reviewed: 2026-05-08
---

# Recova Backend Compatibility Matrix

Dokumen ini mendefinisikan target kompatibilitas kontrak API publik agar migrasi backend tidak memutus integrasi klien.

## Compatibility Levels

| Level                       | Definisi                                                           | Status penggunaan                       |
| --------------------------- | ------------------------------------------------------------------ | --------------------------------------- |
| `C0 - Strict Parity`        | path, method, status code utama, dan struktur respons tetap setara | default untuk semua endpoint publik     |
| `C1 - Compatible Extension` | penambahan field non-breaking, optional, atau metadata tambahan    | boleh jika tidak mengubah perilaku lama |
| `C2 - Breaking`             | perubahan yang menuntut perubahan klien                            | tidak diizinkan pada window migrasi     |

## Global Contract Invariants

Semua domain endpoint publik wajib menjaga:

- prefix `/api/v1`,
- metode HTTP per endpoint,
- requirement autentikasi/otorisasi,
- struktur response envelope sukses dan gagal,
- format identifier utama dan pagination contract bila ada.

## Domain Compatibility Matrix

| Domain          | Path group          | Target level                                             | Acceptance signal                                                       |
| --------------- | ------------------- | -------------------------------------------------------- | ----------------------------------------------------------------------- |
| Authentication  | `/api/v1/auth`      | `C0`                                                     | login, onboarding, token validation, dan error auth utama identik       |
| Users & profile | `/api/v1/users`     | `C0`                                                     | profil baca/tulis dan validasi payload konsisten                        |
| AI coach        | `/api/v1/ai`        | `C0` untuk kontrak API, `C1` untuk metadata non-breaking | bentuk respons inti stabil, tambahan field opsional tidak merusak klien |
| Daily routine   | `/api/v1/routine`   | `C0`                                                     | check-in, streak, dan state harian sesuai baseline                      |
| Journals        | `/api/v1/journals`  | `C0`                                                     | CRUD jurnal dan error handling utama setara                             |
| Community       | `/api/v1/community` | `C0`                                                     | post, komentar, dan like behavior konsisten                             |
| Education       | `/api/v1/education` | `C0`                                                     | daftar/detail konten dan filter utama konsisten                         |
| Daily content   | `/api/v1/content`   | `C0`                                                     | motivasi/tantangan harian dan fallback behavior setara                  |

## Compatibility Verification Matrix

| Area verifikasi       | Rule                                      | Bukti lulus                                |
| --------------------- | ----------------------------------------- | ------------------------------------------ |
| Endpoint existence    | endpoint lama tersedia di runtime baru    | contract test endpoint map lulus           |
| Method parity         | method tidak berubah untuk path yang sama | assertion method-path matrix lulus         |
| Status code parity    | skenario sukses/gagal utama tetap         | golden test status code lulus              |
| Response shape parity | field wajib tetap ada dan tipe stabil     | schema/JSON snapshot parity lulus          |
| Auth parity           | kebutuhan token/akses setara              | test unauthorized/forbidden behavior lulus |
| Error envelope parity | format error tetap kompatibel             | snapshot error envelope lulus              |

## Breaking Change Guardrail

Perubahan berikut dikategorikan `C2` dan harus ditolak selama migrasi:

- mengganti path atau method endpoint publik,
- menghapus field response yang sebelumnya wajib,
- mengubah tipe field yang sudah dipakai klien,
- mengubah status code utama pada skenario bisnis inti,
- mengubah requirement autentikasi tanpa kontrak pengganti yang kompatibel.

## Controlled Extension Rule

Perubahan `C1` boleh dilakukan jika:

- field baru bersifat optional,
- nilai default aman untuk klien lama,
- tidak mengubah semantics field lama,
- dokumentasi endpoint diperbarui bersamaan.

## Cutover Readiness Signals Per Domain

Satu domain baru boleh cutover jika:

- seluruh verifikasi matriks domain `pass`,
- tidak ada mismatch mayor dalam observasi pra-produksi,
- rollback switch untuk domain tersebut sudah diuji.

## Related Documents

- [Migration Strategy](/Users/macbookpro/Development/recova-backend-v2/docs/roadmap/migration-strategy.md)
- [Current Express Baseline](/Users/macbookpro/Development/recova-backend-v2/docs/roadmap/express-baseline.md)
- [Feature Inventory](/Users/macbookpro/Development/recova-backend-v2/docs/roadmap/feature-inventory.md)

## Source Reference

- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
- [/Users/macbookpro/Development/bisakerja-api/docs/architecture.md](/Users/macbookpro/Development/bisakerja-api/docs/architecture.md)
- [Fiber v3 Documentation](https://docs.gofiber.io/)
