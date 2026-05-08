---
title: ADR 0008 - CI/CD Strategy
description: Keputusan strategi CI/CD berbasis pipeline deterministik dengan gate kualitas, keamanan, migrasi, kompatibilitas kontrak, dan rollback readiness.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/decisions/adr-0008-ci-cd-strategy.md
last_reviewed: 2026-05-08
---

# ADR 0008 - CI/CD Strategy

## Status

Proposed

## Context

Layanan membutuhkan pipeline yang:

- dapat memblok perubahan berisiko sebelum deploy,
- memverifikasi migration database secara eksplisit,
- memverifikasi kompatibilitas kontrak API,
- menyediakan jalur rollback operasional.

## Decision

Gunakan pipeline CI/CD deterministik dengan urutan gate tetap:

1. static checks,
2. unit/handler tests,
3. integration + migration verification,
4. contract compatibility tests,
5. image build + security scan,
6. deploy gate,
7. post-deploy smoke and health checks.

Pipeline deployment menggunakan dependency job eksplisit dan kontrol concurrency agar satu target environment tidak dideploy paralel.

## Decision Drivers

- mengurangi risiko regresi production,
- menjaga konsistensi kontrak API,
- memastikan kegagalan migration menghentikan release,
- meningkatkan auditability keputusan release.

## Alternatives Considered

### A1 - Pipeline minimal (lint + unit test)

- plus: cepat.
- minus: risiko tinggi pada migration dan compatibility drift.
- hasil: ditolak.

### A2 - Deploy langsung dari branch tanpa release gates

- plus: sederhana.
- minus: tidak ada kontrol risiko memadai.
- hasil: ditolak.

### A3 - Pipeline deterministik multi-gate

- plus: kontrol kualitas, keamanan, dan operasional lebih kuat.
- minus: durasi pipeline lebih panjang.
- hasil: dipilih.

## Consequences

Konsekuensi positif:

- kualitas release lebih terjaga,
- drift kontrak lebih cepat terdeteksi,
- rollback decision lebih terstruktur.

Konsekuensi negatif:

- butuh investasi pemeliharaan test suite dan pipeline,
- waktu feedback lebih lama dibanding pipeline minimal.

## Guardrails

- migration failure wajib memblok deploy,
- vulnerability critical/high tanpa approval wajib memblok deploy,
- deploy ke environment yang sama wajib serial (concurrency control),
- semua override gate harus terdokumentasi.

## Related Documents

- [CI/CD Operations](/Users/macbookpro/Development/recova-backend-v2/docs/operations/ci-cd.md)
- [Release Gates](/Users/macbookpro/Development/recova-backend-v2/docs/operations/release-gates.md)
- [Testing Strategy](/Users/macbookpro/Development/recova-backend-v2/docs/operations/testing.md)

## Source Reference

- [GitHub Actions Workflow Syntax](https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax)
- [GitHub Actions Concurrency](https://docs.github.com/en/actions/concepts/workflows-and-actions/concurrency)
