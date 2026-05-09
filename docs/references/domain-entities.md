---
title: Recova Backend Domain Entities Reference
description: Referensi entitas domain backend Recova beserta ownership, relasi utama, dan kebutuhan data per modul.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/references/domain-entities.md
last_reviewed: 2026-05-09
---

# Recova Backend Domain Entities Reference

Dokumen ini merinci entitas domain dan keterkaitannya dengan modul aplikasi.

## Entity Inventory

| Entity                    | Module owner       | Description                                 |
| ------------------------- | ------------------ | ------------------------------------------- |
| `User`                    | auth/users         | identitas akun pengguna                     |
| `AuthRefreshToken`        | auth               | state refresh token ter-rotasi dan revokasi |
| `Profile`                 | users              | data profil dan onboarding pengguna         |
| `CheckIn`                 | routine            | catatan check-in harian                     |
| `Streak`                  | routine/statistics | data streak aktif dan histori               |
| `JournalEntry`            | journals           | catatan jurnal pribadi pengguna             |
| `CommunityPost`           | community          | postingan publik komunitas                  |
| `CommunityComment`        | community          | komentar pada postingan komunitas           |
| `CommunityLike`           | community          | relasi like pengguna pada postingan         |
| `Achievement`             | achievements       | katalog milestone pemulihan                 |
| `UserAchievementProgress` | achievements       | progres dan status unlock achievement user  |
| `UserAIPersonaPreference` | ai                 | preferensi persona AI coach per pengguna    |
| `EducationContent`        | education          | konten edukasi aplikasi                     |
| `DailyMotivation`         | content            | konten motivasi harian                      |
| `DailyChallenge`          | content            | konten tantangan harian                     |
| `AiCoachChat`             | ai                 | histori percakapan pengguna dengan AI Coach |

## Entity to Table Mapping

| Entity                    | Table                         |
| ------------------------- | ----------------------------- |
| `User`                    | `users`                       |
| `AuthRefreshToken`        | `auth_refresh_tokens`         |
| `Profile`                 | `profiles`                    |
| `CheckIn`                 | `check_ins`                   |
| `Streak`                  | `streaks`                     |
| `JournalEntry`            | `journals`                    |
| `CommunityPost`           | `community_posts`             |
| `CommunityComment`        | `community_comments`          |
| `CommunityLike`           | `community_post_likes`        |
| `Achievement`             | `achievements`                |
| `UserAchievementProgress` | `user_achievement_progress`   |
| `UserAIPersonaPreference` | `user_ai_persona_preferences` |
| `EducationContent`        | `education_contents`          |
| `DailyMotivation`         | `daily_motivations`           |
| `DailyChallenge`          | `daily_challenges`            |
| `AiCoachChat`             | `ai_chats`                    |

## Relationship Summary

| Parent           | Child                   | Cardinality             |
| ---------------- | ----------------------- | ----------------------- |
| User             | Profile                 | `1:1`                   |
| User             | CheckIn                 | `1:N`                   |
| User             | JournalEntry            | `1:N`                   |
| User             | CommunityPost           | `1:N`                   |
| User             | CommunityComment        | `1:N`                   |
| User             | CommunityLike           | `1:N`                   |
| User             | UserAchievementProgress | `1:N`                   |
| User             | UserAIPersonaPreference | `1:1`                   |
| User             | AiCoachChat             | `1:N`                   |
| CommunityPost    | CommunityComment        | `1:N`                   |
| CommunityPost    | CommunityLike           | `1:N`                   |
| Achievement      | UserAchievementProgress | `1:N`                   |
| CommunityComment | CommunityComment        | `1:N` (reply hierarchy) |

## Module Data Dependency

| Module       | Required entities                                            |
| ------------ | ------------------------------------------------------------ |
| auth         | User, AuthRefreshToken                                       |
| users        | User, Profile                                                |
| routine      | User, CheckIn, Streak                                        |
| journals     | User, JournalEntry                                           |
| community    | User, CommunityPost, CommunityComment, CommunityLike         |
| achievements | User, Achievement, UserAchievementProgress                   |
| education    | EducationContent                                             |
| content      | DailyMotivation, DailyChallenge                              |
| ai           | User, Profile, CheckIn, AiCoachChat, UserAIPersonaPreference |

## Gap Register

- atribut field detail per entitas belum memiliki source formal,
- aturan unique constraint detail per entitas belum memiliki source formal,
- cardinality detail untuk `Streak` histori masih memerlukan keputusan desain final.

## Related Documents

- [Database](/Users/macbookpro/Development/recova-backend-v2/docs/database.md)
- [Data Sensitivity Matrix](/Users/macbookpro/Development/recova-backend-v2/docs/references/data-sensitivity-matrix.md)

## Source Reference

- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
