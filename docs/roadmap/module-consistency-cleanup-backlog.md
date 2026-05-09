---
title: Module Consistency Cleanup Backlog
description: Backlog cleanup konsistensi module berbasis batch risiko agar eksekusi refactor tetap aman, terukur, dan tidak mengubah kontrak API tanpa approval khusus.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/roadmap/module-consistency-cleanup-backlog.md
last_reviewed: 2026-05-09
---

# Module Consistency Cleanup Backlog

Dokumen ini membagi cleanup konsistensi module menjadi batch risiko kecil agar implementasi tetap behavior-preserving secara default.

## Risk Batch Classification

| Batch | Kriteria                                                                                 | Verifikasi minimum                                                                         |
| ----- | ---------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------ |
| B1    | perubahan naming-only (identifier internal, komentar, naming helper) tanpa ubah behavior | `make module-consistency-check`, `go test ./...`                                           |
| B2    | perubahan test-only (penambahan/pengetatan test tanpa ubah runtime)                      | `go test ./...`, `go vet ./...`, suite test target modul terkait                           |
| B3    | refactor behavior-preserving (struktur internal, wiring helper, guard check)             | `go test ./...`, `go vet ./...`, `make module-consistency-full-check`                      |
| B4    | perubahan berpotensi mengubah kontrak API/behavior domain                                | harus dipisah ke change set khusus dengan approval eksplisit sebelum implementasi runtime. |

## Active Backlog

| ID     | Area                          | Ringkasan item                                                                                   | Batch | Status         | Owner              | Exit Criteria                                                                   |
| ------ | ----------------------------- | ------------------------------------------------------------------------------------------------ | ----- | -------------- | ------------------ | ------------------------------------------------------------------------------- |
| MC-001 | Pagination envelope parity    | Seragamkan metadata pagination list endpoint ke `meta.pagination` lintas modul read-list.        | B4    | needs-planning | api-contract-owner | contract update disetujui + OpenAPI update + contract tests lulus.              |
| MC-002 | Repository timeout policy     | Standarisasi timeout query repository lintas modul dengan guard test konsistensi.                | B3    | backlog        | platform-owner     | policy tertulis + test guard baru + `make module-consistency-full-check` lulus. |
| MC-003 | Read-only validator exception | Evaluasi ulang exception `validator.go` pada modul read-only saat write payload baru ditambah.   | B3    | monitor        | content-owner      | exception tetap valid atau validator + test companion ditambahkan.              |
| MC-004 | Content repo companion test   | Tambah repository-level companion test khusus content bila jalur persistence bertambah kompleks. | B2    | monitor        | content-owner      | test baru tersedia atau exception diperbarui dengan bukti coverage setara.      |

## Enforcement Checklist

Checklist ini wajib dipakai pada setiap cleanup konsistensi module:

- jalankan `make module-consistency-check` untuk anatomy, companion tests, route registration, dan auth guard.
- jalankan `make openapi-check` untuk deteksi drift route runtime versus OpenAPI.
- gunakan `make module-consistency-full-check` sebagai jalur gabungan sebelum merge cleanup batch B3.
- jika item masuk batch B4, pecah ke change set terpisah dan tahan implementasi runtime sampai approval.

## Related Documents

- [Module Structure Consistency Standard](/Users/macbookpro/Development/recova-backend-v2/docs/standards/module-structure-consistency.md)
- [Module Contract Consistency Standard](/Users/macbookpro/Development/recova-backend-v2/docs/standards/module-contract-consistency.md)
- [Internal Modules Consistency Audit](/Users/macbookpro/Development/recova-backend-v2/docs/roadmap/internal-modules-consistency-audit.md)
- [Testing Strategy](/Users/macbookpro/Development/recova-backend-v2/docs/operations/testing.md)

## Source Reference

- [Go Testing Package](https://go.dev/pkg/testing/?m=old)
- [Fiber Grouping](https://docs.gofiber.io/guide/grouping)
- [OpenAPI Specification](https://spec.openapis.org/oas/latest)
