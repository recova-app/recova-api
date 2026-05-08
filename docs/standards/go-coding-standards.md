---
title: Recova Backend Go Coding Standards
description: Standar penulisan kode Go untuk Recova Backend mencakup struktur package, signature lintas layer, context propagation, dan aturan kualitas code review.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/standards/go-coding-standards.md
last_reviewed: 2026-05-08
---

# Recova Backend Go Coding Standards

Dokumen ini menetapkan aturan baseline penulisan kode Go agar implementasi konsisten dan mudah direview.

## Package Naming

- gunakan nama package huruf kecil,
- hindari underscore dan nama ambigu,
- nama package harus merepresentasikan domain atau fungsi teknis,
- hindari package util generik tanpa boundary jelas.

## Layer Contract

Kontrak lintas layer:

- handler menerima request HTTP dan memetakan ke DTO,
- service memuat aturan bisnis,
- repository mengelola query persistence,
- handler tidak mengakses repository langsung.

## Function Signature Rules

- gunakan `context.Context` pada boundary service dan repository,
- context menjadi parameter pertama,
- kembalikan `error` sebagai return terakhir,
- hindari return tuple yang tidak perlu.

## Context Propagation

- request context dari HTTP harus diteruskan sampai repository,
- jangan membuat `context.Background()` baru di tengah request flow,
- timeout/cancel dikelola di boundary entrypoint.

## Data Modeling Rules

- pisahkan DTO transport dari model persistence,
- jangan expose field internal persistence langsung ke response API,
- mapper harus eksplisit dan testable.

## State and Concurrency Rules

- hindari global mutable state,
- singleton hanya untuk dependency yang memang shared (config, logger, db pool),
- akses state bersama wajib aman terhadap race,
- goroutine baru harus punya lifecycle/termination yang jelas.

## Configuration and Dependency Rules

- konfigurasi runtime wajib fail-fast,
- dependency di-inject eksplisit,
- hindari hidden dependency lewat package-level variable.

## Error Rules

- error harus dibungkus dengan konteks yang relevan,
- gunakan sentinel atau typed error pada kasus domain penting,
- mapping error ke HTTP response dilakukan terpusat.

Detail ada di [Error Handling Standard](/Users/macbookpro/Development/recova-backend-v2/docs/standards/error-handling.md).

## Formatting and Static Analysis

- format kode dengan formatter Go standar,
- lint wajib dijalankan sebelum merge,
- warning lint kritis tidak boleh diabaikan tanpa alasan terdokumentasi.

## Code Review Checklist

- naming package/fungsi jelas,
- context diteruskan konsisten,
- tidak ada layer violation,
- error handling tidak membocorkan data sensitif,
- test untuk aturan bisnis utama tersedia.

## Related Documents

- [Project Structure](/Users/macbookpro/Development/recova-backend-v2/docs/project-structure.md)
- [Import Boundaries Standard](/Users/macbookpro/Development/recova-backend-v2/docs/standards/import-boundaries.md)
- [Error Handling Standard](/Users/macbookpro/Development/recova-backend-v2/docs/standards/error-handling.md)
- [Testing Conventions](/Users/macbookpro/Development/recova-backend-v2/docs/standards/testing-conventions.md)

## Source Reference

- [How to Write Go Code](https://go.dev/doc/code)
- [Effective Go](https://go.dev/doc/effective_go)
- [Organizing a Go Module](https://go.dev/doc/modules/layout)
- [Fiber Guide: Error Handling](https://docs.gofiber.io/guide/error-handling/)
