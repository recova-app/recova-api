---
title: Recova Backend Containerization Strategy
description: Strategi container image untuk runtime backend yang aman dan ringan, meliputi multi-stage build, non-root execution, base image tradeoff, dan aturan secret handling.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/operations/containerization.md
last_reviewed: 2026-05-08
---

# Recova Backend Containerization Strategy

Dokumen ini mendefinisikan baseline container image untuk deployment backend.

## Objectives

- image production kecil dan reproducible,
- minim attack surface,
- runtime non-root,
- secret tidak pernah dibake ke image.

## Build Model

Gunakan multi-stage Dockerfile:

1. **builder stage**: compile binary + dependency build tools,
2. **runtime stage**: hanya bawa binary dan aset runtime minimum.

Aturan:

- hanya file runtime yang dicopy ke final stage,
- pisahkan dependency build dari image runtime,
- image final harus deterministic berdasarkan lock dan source.

## Runtime User Policy

- proses aplikasi wajib berjalan sebagai non-root,
- gunakan instruction `USER` pada final image,
- hindari capability tambahan kecuali diperlukan dan terdokumentasi.

## Base Image Tradeoff

| Option     | Kelebihan                    | Kekurangan                               | Pemakaian                                 |
| ---------- | ---------------------------- | ---------------------------------------- | ----------------------------------------- |
| Distroless | attack surface sangat kecil  | debugging lebih sulit                    | production default bila tooling mendukung |
| Alpine     | ukuran kecil, shell tersedia | potensi perbedaan libc dan paket runtime | cocok untuk debugging ringan              |

## Static Binary Guidance

- evaluasi build static binary untuk mengurangi dependency runtime,
- verifikasi behavior TLS/CA cert pada base image yang dipilih,
- dokumentasikan keputusan bila memilih dynamic runtime.

## Secret Handling Rules

- secret hanya di-inject saat runtime,
- dilarang copy `.env` ke image,
- dilarang hardcode token/API key di Dockerfile,
- gunakan secret manager atau env injection platform.

## Health Integration

- image/runtime wajib expose endpoint health standar,
- healthcheck container mengarah ke `/health/live`,
- readiness diverifikasi oleh orchestrator/deploy gate.

## Verification Checklist

- image berjalan non-root,
- ukuran image sesuai target efisiensi,
- tidak ada secret pada layer image,
- healthcheck pass,
- graceful shutdown terverifikasi.

## Related Documents

- [Local Development Runtime](/Users/macbookpro/Development/recova-backend-v2/docs/operations/local-development.md)
- [Deployment Operations](/Users/macbookpro/Development/recova-backend-v2/docs/operations/ci-cd.md)
- [Health and Readiness](/Users/macbookpro/Development/recova-backend-v2/docs/operations/health-readiness.md)

## Source Reference

- [Docker Multi-stage Builds](https://docs.docker.com/build/building/multi-stage/)
- [Dockerfile Reference](https://docs.docker.com/reference/dockerfile)
