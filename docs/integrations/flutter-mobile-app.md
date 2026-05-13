---
title: Recova Backend Flutter Mobile App Integration
description: Kontrak integrasi API Recova Backend untuk aplikasi mobile Flutter meliputi stabilitas schema, autentikasi, bahasa response, dan batas kompatibilitas perubahan.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/integrations/flutter-mobile-app.md
last_reviewed: 2026-05-08
---

# Recova Backend Flutter Mobile App Integration

Dokumen ini mendefinisikan kontrak integrasi antara backend API dan aplikasi mobile Flutter.

## Integration Objectives

- parsing response di Flutter stabil dan deterministik,
- penanganan error konsisten untuk UI mobile,
- perubahan API tidak memaksa update klien mendadak.

## API Contract Rules for Flutter Client

- gunakan envelope response yang konsisten untuk semua endpoint,
- pertahankan nama field JSON yang sudah dipublikasikan,
- hindari perubahan tipe field pada kontrak existing,
- tambahan field baru harus optional dan backward-compatible.

## Language Contract

- message human-readable pada response/error harus bahasa Indonesia,
- machine-readable fields (`error.code`, identifier teknis) tetap English.

## Authentication Contract

- endpoint terproteksi menerima header `Authorization` sesuai standar bearer token,
- failure auth wajib konsisten (`UNAUTHENTICATED`/`FORBIDDEN`) untuk memudahkan mapping UI state.

## Mobile Network Reliability Considerations

- response payload dijaga ringkas agar cocok untuk jaringan mobile,
- timeout dan retry policy harus didukung dengan error code yang stabil,
- pagination diprioritaskan pada list endpoint untuk membatasi ukuran payload.

## Serialization and Compatibility

- format JSON harus konsisten dan mudah dipetakan ke model Dart,
- timestamp wajib konsisten lintas endpoint,
- hindari nilai ambigu (`""` vs `null`) untuk field optional.

## Change Management

Perubahan kontrak yang berdampak ke Flutter wajib:

- memperbarui dokumen API terkait,
- memperbarui test contract,
- mencantumkan compatibility note sebelum rilis.

## Related Documents

- [API Response Standard](/Users/macbookpro/Development/recova-backend-v2/docs/api-response-standard.md)
- [API Reference](/Users/macbookpro/Development/recova-backend-v2/docs/api-reference.md)
- [Compatibility Matrix](/Users/macbookpro/Development/recova-backend-v2/docs/roadmap/compatibility-matrix.md)

## Source Reference

- [Flutter Networking Cookbook](https://docs.flutter.dev/cookbook/networking)
- [Flutter JSON and Serialization](https://docs.flutter.dev/data-and-backend/serialization/json)
- [Flutter Internationalization](https://docs.flutter.dev/ui/internationalization)
- [Flutter Authenticated Requests](https://docs.flutter.dev/cookbook/networking/authenticated-requests)
