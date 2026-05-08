---
title: Community Moderation Baseline
description: Baseline kebijakan moderasi komunitas untuk klasifikasi pelanggaran, alur penanganan abuse, dan kontrol operasional aman.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/modules/community-moderation.md
last_reviewed: 2026-05-08
---

# Community Moderation Baseline

Dokumen ini mendefinisikan baseline moderasi konten komunitas pada backend.

## Moderation Objectives

- mencegah penyebaran konten abuse,
- menjaga keamanan pengguna,
- menjaga kualitas interaksi komunitas,
- menyediakan jejak audit tindakan moderasi.

## Moderation Scope

Konten yang berada dalam cakupan moderasi:

- teks post komunitas,
- teks komentar komunitas,
- pola spam interaksi like/comment.

## Violation Categories

| Category          | Description                                             | Default action                        |
| ----------------- | ------------------------------------------------------- | ------------------------------------- |
| `spam`            | konten berulang atau promosi agresif                    | rate-limit + hide candidate           |
| `harassment`      | penghinaan, ancaman, atau serangan personal             | hide + escalate review                |
| `self-harm-risk`  | indikasi risiko keselamatan diri yang memerlukan triase | hide candidate + high-priority review |
| `illegal-content` | konten terlarang sesuai kebijakan hukum                 | hide + block + audit                  |
| `sexual-explicit` | konten seksual eksplisit                                | hide + audit                          |

## Moderation Decision States

| State          | Meaning                                      |
| -------------- | -------------------------------------------- |
| `active`       | konten aman dan tampil pada feed             |
| `hidden`       | konten disembunyikan dari feed publik        |
| `under_review` | konten ditahan sementara untuk review lanjut |
| `removed`      | konten dihapus dari tampilan pengguna        |

## Enforcement Rules

- konten yang melanggar kategori berat tidak boleh tetap `active`,
- status moderasi harus tercatat per entitas post/comment,
- perubahan status moderasi harus idempotent,
- operasi moderasi harus dapat ditelusuri melalui audit log.

## Abuse Workflow

```text
incoming content
  -> validation
  -> abuse signal check
  -> moderation decision (active/hidden/under_review/removed)
  -> feed visibility update
  -> audit event
```

## Rate-Limit Baseline

- terapkan limiter ketat pada endpoint create post/comment,
- terapkan limiter pada endpoint like untuk menahan spam interaksi,
- rate-limit key berbasis user + IP bila dibutuhkan,
- pelanggaran berulang masuk ke jalur `under_review` atau block sementara.

## Logging and Privacy Rules

- log moderasi hanya menyimpan metadata (`entity_id`, `category`, `decision`, `request_id`),
- hindari logging konten mentah untuk kategori sensitif,
- detail payload raw hanya boleh tersedia pada kanal audit dengan kontrol akses ketat.

## Operational Gaps

- threshold final untuk auto-hide vs manual review,
- SLA review untuk konten `under_review`,
- desain interface operator moderasi.

## Related Documents

- [Community Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/community.md)
- [Data Sensitivity Matrix](/Users/macbookpro/Development/recova-backend-v2/docs/references/data-sensitivity-matrix.md)
- [Error Taxonomy](/Users/macbookpro/Development/recova-backend-v2/docs/standards/error-taxonomy.md)

## Source Reference

- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
- [Fiber Limiter Middleware](https://docs.gofiber.io/middleware/limiter/)
