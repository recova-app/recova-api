---
title: Recova Backend Database
description: Baseline model domain data Recova, relasi entitas, ownership data, sensitivitas, dan query pattern utama sebelum pemodelan ORM detail.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/database.md
last_reviewed: 2026-05-08
---

# Recova Backend Database

Dokumen ini mendefinisikan baseline domain data backend Recova sebagai fondasi desain persistence.

## Domain Entity Coverage

Entitas domain yang sudah teridentifikasi:

- users
- profiles
- check-ins
- streaks
- journals
- community posts
- comments
- likes
- education content
- ai coach chats
- daily motivations
- daily challenges

## Ownership Boundary

| Data area                   | Owner       | Notes                                      |
| --------------------------- | ----------- | ------------------------------------------ |
| User identity/profile       | backend API | data inti pengguna                         |
| Routine, journal, statistik | backend API | data personal pengguna                     |
| Community interactions      | backend API | data sosial aplikasi                       |
| Education/daily content     | backend API | data konten aplikasi                       |
| AI coach conversation       | backend API | data sensitif, butuh kontrol privasi ketat |

## Relationship Baseline

```text
User
  -> Profile (1:1)
  -> CheckIn (1:N)
  -> Streak (1:1 atau 1:N histori)
  -> Journal (1:N)
  -> CommunityPost (1:N)
  -> Comment (1:N)
  -> Like (1:N)
  -> AiCoachChat (1:N)

CommunityPost
  -> Comment (1:N)
  -> Like (1:N)

Content (Education / Daily Motivation / Daily Challenge)
  -> disajikan ke User melalui API read flows
```

## Data Sensitivity Baseline

Klasifikasi detail ada di:

- [Domain Entities Reference](/Users/macbookpro/Development/recova-backend-v2/docs/references/domain-entities.md)
- [Data Sensitivity Matrix](/Users/macbookpro/Development/recova-backend-v2/docs/references/data-sensitivity-matrix.md)

## Retention and Privacy Direction

- data jurnal dan chat AI diperlakukan sebagai data sensitif.
- konten sensitif tidak boleh masuk log aplikasi.
- endpoint reset data hanya boleh untuk lingkungan development.
- aturan retensi final perlu ditetapkan per entitas sebelum implementasi produksi.

## Query Pattern Baseline

| Module            | Query pattern utama                                    |
| ----------------- | ------------------------------------------------------ |
| auth/users        | lookup by identity, update profile/settings            |
| routine           | insert check-in harian, aggregate statistik dan streak |
| journals          | list by user + create entry                            |
| community         | list post, create post, add comment, add like          |
| education/content | list/read konten aktif                                 |
| ai                | insert/read chat history dan summary terkait pengguna  |

## Gap Register

Area yang belum punya source rinci dan butuh pendalaman:

- skema kolom per entitas,
- kebijakan retensi kuantitatif per entitas,
- strategi indexing per query hot path,
- aturan data lifecycle untuk konten dinamis.

## Related Documents

- [API Reference](/Users/macbookpro/Development/recova-backend-v2/docs/api-reference.md)
- [Tech Stack](/Users/macbookpro/Development/recova-backend-v2/docs/tech-stack.md)
- [ADR 0004 ORM GORM PostgreSQL](/Users/macbookpro/Development/recova-backend-v2/docs/decisions/adr-0004-orm-gorm-postgresql.md)
- [ADR 0005 Database Migration Tool](/Users/macbookpro/Development/recova-backend-v2/docs/decisions/adr-0005-database-migration-tool.md)

## Source Reference

- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
- [/Users/macbookpro/Development/bisakerja-api/docs/database.md](/Users/macbookpro/Development/bisakerja-api/docs/database.md)
- [PostgreSQL Current Documentation](https://www.postgresql.org/docs/current/index.html)
