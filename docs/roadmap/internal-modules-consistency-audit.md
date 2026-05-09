---
title: Internal Modules Consistency Audit
description: Laporan audit konsistensi struktur dan kontrak runtime lintas module internal Recova Backend berbasis bukti file implementasi aktual.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/roadmap/internal-modules-consistency-audit.md
last_reviewed: 2026-05-09
---

# Internal Modules Consistency Audit

Dokumen ini mencatat hasil audit anatomi module dan konsistensi kontrak runtime dari implementasi aktual di `internal/modules`.

## Scope and Method

Audit mencakup:

- inventory semua subfolder langsung `internal/modules`,
- keberadaan anatomy file baseline,
- companion test per layer,
- boundary handler/service/repository,
- pola route registration dan auth guard,
- sinkronisasi contract test terhadap OpenAPI/runtime.

Metode:

- pemeriksaan file tree,
- pembacaan source route/handler/service/repository,
- eksekusi gate otomatis `make module-consistency-check` dan `make openapi-check`.

## Module Inventory Result

| Module       | Anatomy baseline  | Companion test baseline | Status              |
| ------------ | ----------------- | ----------------------- | ------------------- |
| achievements | pass              | pass                    | pass                |
| ai           | pass              | pass                    | pass                |
| auth         | pass              | pass                    | pass                |
| community    | pass              | pass                    | pass                |
| content      | deviation tracked | deviation tracked       | pass-with-exception |
| education    | deviation tracked | pass                    | pass-with-exception |
| journals     | pass              | pass                    | pass                |
| routine      | pass              | pass                    | pass                |
| users        | pass              | pass                    | pass                |

## Gap Register

| Type           | Module       | Gap                                                    | Status               | Owner                | Resolution Track                                           |
| -------------- | ------------ | ------------------------------------------------------ | -------------------- | -------------------- | ---------------------------------------------------------- |
| structure      | content      | `validator.go` belum tersedia                          | `accepted-exception` | `content-owner`      | tetap dipantau sampai ada write payload baru               |
| structure      | education    | `validator.go` belum tersedia                          | `accepted-exception` | `education-owner`    | tetap dipantau sampai ada write payload baru               |
| companion-test | content      | repository test khusus module belum tersedia           | `accepted-exception` | `content-owner`      | saat ini dicakup oleh seed integration lintas jalur konten |
| contract       | cross-module | pagination metadata belum seragam di `meta.pagination` | `needs-follow-up`    | `api-contract-owner` | butuh perubahan kontrak terpisah agar tidak breaking       |
| contract       | cross-module | timeout policy repository belum seragam eksplisit      | `needs-follow-up`    | `platform-owner`     | butuh standar timeout lintas repository                    |

## Boundary Consistency Result

Temuan yang lolos:

- handler tidak mengakses adapter database langsung,
- service tidak bergantung ke objek HTTP context Fiber,
- repository tidak memformat response envelope,
- route non-auth memakai auth guard untuk endpoint protected.

Exception sah:

- `auth` dan `users` memakai nama fungsi registrasi route khusus domain.

## Drift Summary

- route runtime dan OpenAPI saat ini sinkron (mengacu gate `openapi-check` + contract tests).
- tidak ada rekomendasi refactor behavior-changing pada dokumen ini.
- item `needs-follow-up` dipisah sebagai backlog kontrak agar aman dari perubahan mendadak.

## Related Documents

- [Module Structure Consistency Standard](/Users/macbookpro/Development/recova-backend-v2/docs/standards/module-structure-consistency.md)
- [Module Contract Consistency Standard](/Users/macbookpro/Development/recova-backend-v2/docs/standards/module-contract-consistency.md)
- [Module Consistency Cleanup Backlog](/Users/macbookpro/Development/recova-backend-v2/docs/roadmap/module-consistency-cleanup-backlog.md)
- [Testing Strategy](/Users/macbookpro/Development/recova-backend-v2/docs/operations/testing.md)
- [Verification Matrix](/Users/macbookpro/Development/recova-backend-v2/docs/operations/verification-matrix.md)

## Source Reference

- [internal/modules](/Users/macbookpro/Development/recova-backend-v2/internal/modules)
- [Route Contract Tests](/Users/macbookpro/Development/recova-backend-v2/test/contract)
