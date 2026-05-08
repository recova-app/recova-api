---
title: ADR 0001 - Migration Strategy
description: Keputusan arsitektural untuk strategi migrasi backend dengan kompatibilitas API sebagai prioritas utama.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/decisions/adr-0001-migration-strategy.md
last_reviewed: 2026-05-08
---

# ADR 0001 - Migration Strategy

## Status

Proposed

## Context

Backend harus dipindahkan ke runtime Go tanpa memutus kontrak frontend yang sudah berjalan pada jalur `/api/v1`.

Batas utama:

- endpoint publik sudah dipakai klien aktif,
- perubahan kontrak besar selama migrasi meningkatkan risiko regresi,
- rollback harus bisa dilakukan cepat saat incident,
- migrasi data tidak boleh memaksa downtime panjang.

## Decision

Gunakan **API-compatible rewrite with controlled coexistence**:

- bangun runtime Go dengan kontrak endpoint setara,
- jalankan coexistence runtime lama dan runtime baru selama window transisi,
- lakukan cutover inkremental per domain endpoint setelah verifikasi kompatibilitas,
- pertahankan opsi rollback cepat dengan pengalihan trafik kembali ke runtime lama.

## Decision Drivers

- kestabilan kontrak frontend lebih penting daripada kecepatan cutover,
- kebutuhan rollback cepat untuk menekan blast radius,
- kebutuhan verifikasi inkremental untuk area auth, data write, dan AI endpoint,
- kebutuhan menjaga schema database tetap aman untuk dua runtime aktif.

## Alternatives Considered

### A1 - Big-bang rewrite

- plus: satu kali cutover, koordinasi sederhana.
- minus: risiko outage/regresi tinggi, rollback sulit jika schema berubah.
- hasil: ditolak.

### A2 - Strangler per endpoint penuh

- plus: risiko lebih kecil per endpoint.
- minus: kompleksitas orkestrasi tinggi bila terlalu granular.
- hasil: dipakai sebagai teknik eksekusi cutover, bukan strategi tunggal.

### A3 - Parallel service + switch instan seluruh domain

- plus: verifikasi pra-cutover bisa kuat.
- minus: blast radius tetap besar saat switch serentak.
- hasil: ditolak untuk produksi; diganti cutover inkremental.

## Consequences

Konsekuensi positif:

- frontend tidak perlu migrasi kontrak besar selama transisi,
- insiden cutover bisa dipulihkan cepat via rollback routing,
- risiko dapat diisolasi per domain endpoint.

Konsekuensi negatif:

- biaya operasi sementara naik karena dua runtime berjalan,
- governance kompatibilitas harus disiplin,
- observability dan contract testing wajib lebih ketat.

## Guardrails

- breaking change kontrak API publik tidak diizinkan pada window migrasi,
- perubahan schema harus backward-compatible terhadap dua runtime aktif,
- domain cutover hanya boleh jalan jika gate verifikasi kompatibilitas lulus,
- ownership write domain harus tunggal untuk menghindari konflik data.

## Review Triggers

ADR ini wajib ditinjau ulang jika:

- ditemukan batas infrastruktur yang menghalangi coexistence,
- kontrak klien utama berubah signifikan,
- muncul kebutuhan breaking change yang tidak dapat ditunda,
- hasil observasi menunjukkan risiko sistemik pada pendekatan inkremental.

## Related Documents

- [Migration Strategy](/Users/macbookpro/Development/recova-backend-v2/docs/roadmap/migration-strategy.md)
- [Compatibility Matrix](/Users/macbookpro/Development/recova-backend-v2/docs/roadmap/compatibility-matrix.md)
- [Current Express Baseline](/Users/macbookpro/Development/recova-backend-v2/docs/roadmap/express-baseline.md)

## Source Reference

- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
- [/Users/macbookpro/Development/bisakerja-api/docs/architecture.md](/Users/macbookpro/Development/bisakerja-api/docs/architecture.md)
- [/Users/macbookpro/Development/bisakerja-api/docs/operations/deployment.md](/Users/macbookpro/Development/bisakerja-api/docs/operations/deployment.md)
- [Fiber v3 Documentation](https://docs.gofiber.io/)
