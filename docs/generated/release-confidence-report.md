---
title: Recova Backend Release Confidence Report
description: Ringkasan evidence verifikasi release candidate dari critical flow E2E dan performance smoke baseline.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/generated/release-confidence-report.md
last_reviewed: 2026-05-09
generated_by: test-e2e-and-performance-smoke
---

# Recova Backend Release Confidence Report

Dokumen ini mendefinisikan format evidence release confidence berbasis hasil verifikasi otomatis.

## Evidence Inputs

- `artifacts/release-confidence/e2e-critical-flows.json`
- `artifacts/release-confidence/performance-smoke.json`

## Pass/Fail Rules

- critical flow E2E harus `passed`,
- performance smoke harus `passed`,
- threshold performance minimum:
  - `errorRate <= 0.01`,
  - `p95Ms <= 300`.

## Reporting Template

| Item                       | Status   | Catatan                                                                                                     |
| -------------------------- | -------- | ----------------------------------------------------------------------------------------------------------- |
| E2E critical flows         | `passed` | `generatedAt=2026-05-09T13:43:34Z`, scope `all`, termasuk stats summary/threaded reply/achievements/persona |
| Performance smoke          | `passed` | `generatedAt=2026-05-09T13:43:36Z`, semua scenario `errorRate=0`, `p95Ms` jauh di bawah threshold           |
| Release confidence overall | `passed` | seluruh item lulus pada candidate run yang sama                                                             |

## Review Checklist

- [x] report E2E terbarui dari run candidate yang sama,
- [x] report performance terbarui dari run candidate yang sama,
- [x] tidak ada payload sensitif dalam artefak report,
- [x] hasil sinkron dengan gate CI release candidate.

## Related Documents

- [Testing Strategy](/Users/macbookpro/Development/recova-backend-v2/docs/operations/testing.md)
- [Release Gates](/Users/macbookpro/Development/recova-backend-v2/docs/operations/release-gates.md)
- [CI/CD Operations](/Users/macbookpro/Development/recova-backend-v2/docs/operations/ci-cd.md)

## Source Reference

- [Go Testing Package](https://pkg.go.dev/testing)
- [go Command Documentation](https://pkg.go.dev/cmd/go)
- [Fiber App API](https://docs.gofiber.io/api/app/)
- [k6 Thresholds](https://grafana.com/docs/k6/latest/using-k6/thresholds/)
- [Store and Share Data with Workflow Artifacts](https://docs.github.com/en/actions/tutorials/store-and-share-data)
