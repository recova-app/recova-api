---
title: ADR 0002 - Go Project Layout
description: Keputusan layout repository Go berbasis internal modules dan pemisahan tegas antara runtime entrypoint, domain, dan platform adapters.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/decisions/adr-0002-go-project-layout.md
last_reviewed: 2026-05-08
---

# ADR 0002 - Go Project Layout

## Status

Proposed

## Context

Migrasi backend membutuhkan layout repo Go yang:

- mudah dipahami lintas tim,
- menjaga domain logic tetap terisolasi,
- mencegah coupling berlebihan antar-modul,
- mendukung testability dan evolusi layanan.

## Decision

Gunakan layout berbasis:

- `cmd/` untuk entrypoint runtime,
- `internal/modules/` untuk domain modules,
- `internal/platform/` untuk adapter infrastruktur,
- `internal/shared/` untuk komponen lintas-modul yang netral.

Semua domain logic ditempatkan di `internal/modules/*` dan tidak berada di `cmd/*`.

## Decision Drivers

- boundary domain dan infra harus jelas,
- import graph harus mudah diverifikasi,
- kebutuhan scale modul fitur di masa depan,
- keselarasan dengan praktik layout Go modern.

## Alternatives Considered

### A1 - Flat package layout

- plus: sederhana di awal.
- minus: cepat menjadi coupled saat modul bertambah.
- hasil: ditolak.

### A2 - `pkg/` heavy public layout

- plus: reusable package jelas untuk konsumsi eksternal.
- minus: backend ini fokus layanan internal; ekspos package publik berisiko memperlebar API surface internal.
- hasil: tidak dipilih sebagai default.

### A3 - Internal-first modular layout

- plus: boundary kuat, refactor internal lebih aman, cocok untuk service backend.
- minus: butuh disiplin boundary import.
- hasil: dipilih.

## Consequences

Konsekuensi positif:

- batas tanggung jawab antar-layer jelas,
- domain logic tidak tercampur dengan bootstrap runtime,
- lebih mudah membuat guardrail terhadap import terlarang.

Konsekuensi negatif:

- perlu aturan review import yang konsisten,
- jumlah package meningkat dibanding layout flat.

## Guardrails

- `cmd/*` dilarang memuat domain rules atau query persistence.
- import antar-modul domain hanya lewat contract yang disepakati.
- repository tidak boleh bergantung pada handler/route.

## Related Documents

- [Project Structure](/Users/macbookpro/Development/recova-backend-v2/docs/project-structure.md)
- [Import Boundaries Standard](/Users/macbookpro/Development/recova-backend-v2/docs/standards/import-boundaries.md)
- [Architecture](/Users/macbookpro/Development/recova-backend-v2/docs/architecture.md)

## Source Reference

- [Go Module Layout](https://go.dev/doc/modules/layout)
- [golang-standards/project-layout](https://github.com/golang-standards/project-layout)
