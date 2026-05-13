---
title: Recova Backend Import Boundaries Standard
description: Aturan import dan boundary dependensi antar-layer agar coupling tetap rendah dan arsitektur tetap testable.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/standards/import-boundaries.md
last_reviewed: 2026-05-08
---

# Recova Backend Import Boundaries Standard

Dokumen ini menetapkan batas import agar dependency graph backend tetap terkontrol.

## Import Policy

- import antar-layer mengikuti arah dependensi resmi.
- modul domain tidak boleh mengimpor package private modul lain.
- adapter platform boleh dipakai service melalui interface yang jelas.
- package util bersama harus netral dari domain spesifik.

## Allowed Dependency Matrix

| From                            | Allowed to import                                                      |
| ------------------------------- | ---------------------------------------------------------------------- |
| `cmd/*`                         | `internal/app`, `internal/platform/config`, `internal/platform/logger` |
| `internal/app`                  | `internal/modules/*`, `internal/platform/*`, `internal/shared/*`       |
| `internal/modules/*/route`      | `handler`, middleware, DTO/schema                                      |
| `internal/modules/*/handler`    | `service`, DTO/schema, shared response                                 |
| `internal/modules/*/service`    | `repository`, platform adapter interface, shared errors                |
| `internal/modules/*/repository` | `internal/platform/database`, model/mapper internal modul              |
| `internal/platform/*`           | package eksternal + `internal/shared/*` yang relevan                   |

## Disallowed Imports

- `repository -> handler`
- `repository -> route`
- `handler -> platform/database` langsung
- `cmd -> internal/modules/*/repository` langsung
- `module A` import file private `module B` selain API publik yang disepakati

## Interface Boundary Rule

- dependency ke layanan eksternal dan komponen infra dipetakan lewat interface pada service layer.
- implementasi interface berada di `internal/platform/*`.
- test service menggunakan mock dari interface, bukan mock driver infra langsung.

## Circular Dependency Prevention

- shared type lintas modul dipindah ke `internal/shared/types`.
- error codes bersama dipindah ke `internal/shared/errs`.
- mapper lintas modul dilarang; gunakan DTO netral pada boundary service.

## Validation and Enforcement

Validasi minimum saat review:

- tidak ada import yang menembus boundary terlarang,
- tidak ada domain logic di `cmd/*`,
- package baru punya owner layer yang jelas.

## Related Documents

- [Project Structure](/Users/macbookpro/Development/recova-backend-v2/docs/project-structure.md)
- [Architecture](/Users/macbookpro/Development/recova-backend-v2/docs/architecture.md)

## Source Reference

- [Go Module Layout](https://go.dev/doc/modules/layout)
- [golang-standards/project-layout](https://github.com/golang-standards/project-layout)
