---
title: Achievements Module
description: Kontrak modul achievements untuk milestone pemulihan berbasis streak, konsistensi check-in/jurnal, recovery relapse, partisipasi komunitas, dan onboarding completion.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/modules/achievements.md
last_reviewed: 2026-05-09
---

# Achievements Module

## Responsibility

- mendefinisikan katalog achievement dan rule pencapaiannya,
- melacak progres user pada setiap achievement,
- memberikan status unlock dan timestamp yang konsisten,
- menyediakan endpoint read untuk klien agar bisa menampilkan progres milestone.

## API Contract

Route prefix:

```text
/api/v1/achievements
```

| Method | Path                            | Auth class | Purpose                                   |
| ------ | ------------------------------- | ---------- | ----------------------------------------- |
| `GET`  | `/api/v1/achievements/catalog`  | Bearer     | ambil katalog achievement aktif           |
| `GET`  | `/api/v1/achievements/progress` | Bearer     | ambil progres achievement user saat ini   |
| `GET`  | `/api/v1/achievements/unlocked` | Bearer     | ambil daftar achievement user yang unlock |

## Database Model

Entitas utama:

- `achievements`,
- `user_achievement_progress`.

Constraint minimum:

- `achievements.code` unique global,
- `user_achievement_progress` unique per `(user_id, achievement_id)`,
- nilai progres tidak boleh negatif,
- `unlocked_at` null jika belum unlock.

## Achievement Categories

Kategori minimum:

- `streak_milestone`,
- `checkin_consistency`,
- `journal_consistency`,
- `relapse_recovery`,
- `community_participation`,
- `onboarding_completion`.

## Service and Business Rules

- evaluasi achievement berjalan idempotent,
- unlock tidak boleh mundur (sekali unlock tetap unlocked),
- progress update wajib atomic untuk menghindari double-unlock event,
- fallback aman: jika rule evaluator gagal, endpoint read tetap mengembalikan state persisted terakhir.

## Validation Rules

- query filter category harus whitelist,
- payload internal update progres harus valid terhadap target threshold,
- data invalid dipetakan ke `VALIDATION_ERROR`.

## Error Contract

| Condition             | HTTP  | Error code         |
| --------------------- | ----- | ------------------ |
| auth invalid/missing  | `401` | `UNAUTHENTICATED`  |
| payload/query invalid | `422` | `VALIDATION_ERROR` |
| akses tidak diizinkan | `403` | `FORBIDDEN`        |
| kegagalan internal    | `500` | `INTERNAL_ERROR`   |

## Observability Contract

Log field minimum:

- `request_id`,
- `user_id`,
- `achievement_code`,
- `achievement_action`,
- `status_code`.

Metrik minimum:

- unlock rate per category,
- failed evaluation count,
- p95 latency endpoint achievements.

## Testing Requirements

- unit test evaluator per kategori achievement,
- unit test idempotensi unlock,
- integration test unique constraint progress,
- handler test ownership/auth,
- contract test response achievements.

## Related Documents

- [Statistics Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/statistics.md)
- [Routine Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/routine.md)
- [Community Module](/Users/macbookpro/Development/recova-backend-v2/docs/modules/community.md)
- [Data Sensitivity Matrix](/Users/macbookpro/Development/recova-backend-v2/docs/references/data-sensitivity-matrix.md)

## Source Reference

- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
