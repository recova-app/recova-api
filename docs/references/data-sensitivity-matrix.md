---
title: Recova Backend Data Sensitivity Matrix
description: Klasifikasi sensitivitas data backend Recova dan kontrol minimum untuk logging, akses, dan retensi.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/references/data-sensitivity-matrix.md
last_reviewed: 2026-05-09
---

# Recova Backend Data Sensitivity Matrix

Dokumen ini menetapkan klasifikasi sensitivitas data dan kontrol minimum operasional.

## Sensitivity Levels

| Level                 | Definition                                                                  |
| --------------------- | --------------------------------------------------------------------------- |
| `L1 Public`           | aman ditampilkan publik tanpa identitas personal                            |
| `L2 Internal`         | data operasional non-rahasia, akses terbatas internal                       |
| `L3 Sensitive`        | data pengguna personal atau perilaku, butuh kontrol akses ketat             |
| `L4 Highly Sensitive` | data sangat sensitif (mental-health journal/chat context, credential/token) |

## Data Classification Matrix

| Entity/Data                       | Level                 | Logging rule                                      | Access rule                        | Retention direction             |
| --------------------------------- | --------------------- | ------------------------------------------------- | ---------------------------------- | ------------------------------- |
| EducationContent                  | `L1 Public`           | boleh log metadata umum                           | public read                        | retensi sesuai lifecycle konten |
| DailyMotivation / DailyChallenge  | `L1 Public`           | boleh log metadata umum                           | public read                        | retensi sesuai lifecycle konten |
| CommunityPost metadata            | `L2 Internal`         | log ringkas tanpa PII berlebih                    | authenticated/public sesuai policy | retensi moderasi                |
| User profile (nickname, settings) | `L3 Sensitive`        | redaksi field sensitif                            | owner-only write/read terproteksi  | retensi akun aktif              |
| CheckIn/Streak statistics         | `L3 Sensitive`        | log agregat, tanpa detail personal                | owner-only                         | retensi sesuai kebijakan produk |
| Achievement progress              | `L3 Sensitive`        | log agregat progres tanpa payload personal mentah | owner-only                         | retensi akun aktif              |
| AI persona preference             | `L3 Sensitive`        | log metadata persona saja                         | owner-only                         | retensi akun aktif              |
| JournalEntry content              | `L4 Highly Sensitive` | dilarang log konten mentah                        | owner-only ketat                   | retensi minimum yang disetujui  |
| AiCoachChat content               | `L4 Highly Sensitive` | dilarang log prompt/chat mentah                   | owner-only ketat                   | retensi minimum yang disetujui  |
| Auth token/secret                 | `L4 Highly Sensitive` | dilarang log nilai token/secret                   | akses terbatas runtime auth        | retensi sesingkat mungkin       |

## Minimum Controls

- request/response logging wajib redaksi data sensitif.
- error message ke klien tidak menampilkan data sensitif.
- akses data `L3` dan `L4` wajib ownership check.
- backup/restore proses wajib mengikuti kontrol akses least privilege.

## Gap Register

- retensi numerik final per entitas belum memiliki source kebijakan resmi,
- baseline encryption-at-rest/in-transit detail belum ditetapkan per environment.

## Related Documents

- [Database](/Users/macbookpro/Development/recova-backend-v2/docs/database.md)
- [Domain Entities Reference](/Users/macbookpro/Development/recova-backend-v2/docs/references/domain-entities.md)
- [Error Taxonomy](/Users/macbookpro/Development/recova-backend-v2/docs/standards/error-taxonomy.md)

## Source Reference

- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
- [/Users/macbookpro/Development/bisakerja-api/docs/database.md](/Users/macbookpro/Development/bisakerja-api/docs/database.md)
