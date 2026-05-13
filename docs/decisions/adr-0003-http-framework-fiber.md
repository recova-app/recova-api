---
title: ADR 0003 - HTTP Framework Fiber
description: Keputusan penggunaan Fiber v3 sebagai framework HTTP utama untuk backend Go Recova.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/decisions/adr-0003-http-framework-fiber.md
last_reviewed: 2026-05-08
---

# ADR 0003 - HTTP Framework Fiber

## Status

Proposed

## Context

Backend target membutuhkan framework HTTP yang:

- cepat dan ringan,
- punya middleware ekosistem matang,
- mendukung route grouping dan error handling terpusat,
- cocok untuk kontrak API JSON ber-volume menengah hingga tinggi.

## Decision

Gunakan **Fiber v3** sebagai framework HTTP utama.

## Decision Drivers

- model routing dan middleware sesuai kebutuhan API modular,
- performa tinggi dengan overhead rendah,
- dukungan pattern recover middleware dan centralized error handling,
- dokumentasi resmi aktif dan terkini.

## Alternatives Considered

### A1 - net/http + router minimal

- plus: dependency kecil.
- minus: boilerplate middleware dan ergonomi lebih berat.
- hasil: tidak dipilih untuk baseline.

### A2 - Gin/Echo

- plus: komunitas besar.
- minus: tidak dipilih karena arah arsitektur sudah mengutamakan Fiber di kontrak target.
- hasil: ditolak.

### A3 - Fiber v3

- plus: sesuai arah stack target dan kebutuhan performa.
- minus: perlu disiplin penggunaan context agar tidak menyimpan nilai yang tidak aman lintas request.
- hasil: dipilih.

## Consequences

Konsekuensi positif:

- bootstrap middleware konsisten,
- route grouping API mudah dikelola,
- integrasi observability dan error handling lebih terstruktur.

Konsekuensi negatif:

- tim perlu mengadopsi pola context handling Fiber dengan benar,
- perubahan major version framework perlu evaluasi kompatibilitas khusus.

## Guardrails

- semua handler mengembalikan error untuk diproses mapper terpusat,
- recover middleware wajib aktif,
- data context tidak disimpan keluar scope request tanpa copy aman,
- contract API tetap sumber kebenaran, bukan implementasi framework.

## Related Documents

- [Tech Stack](/Users/macbookpro/Development/recova-backend-v2/docs/tech-stack.md)
- [Architecture](/Users/macbookpro/Development/recova-backend-v2/docs/architecture.md)
- [Request Lifecycle](/Users/macbookpro/Development/recova-backend-v2/docs/overview/request-lifecycle.md)

## Source Reference

- [Fiber v3 Documentation](https://docs.gofiber.io/)
- [Fiber Error Handling Guide](https://docs.gofiber.io/guide/error-handling)
