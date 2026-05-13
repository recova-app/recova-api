---
title: Recova Backend Migration Strategy
description: Strategi migrasi backend ke Go dengan kompatibilitas API terjaga, kebijakan freeze kontrak, rollback direction, dan kriteria cutover.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/roadmap/migration-strategy.md
last_reviewed: 2026-05-08
---

# Recova Backend Migration Strategy

Dokumen ini menetapkan strategi migrasi backend Recova dari runtime saat ini ke Go dengan target menjaga kontrak frontend tetap stabil pada jalur `/api/v1`.

## Strategy Decision Summary

Pendekatan yang dipilih adalah **API-compatible rewrite with controlled coexistence**:

- layanan Go dibangun sebagai implementation baru untuk kontrak API yang sama,
- runtime lama dan runtime baru berjalan berdampingan dalam window transisi,
- perpindahan trafik dilakukan inkremental per domain endpoint,
- rollback dilakukan dengan mengembalikan routing ke runtime lama tanpa mengubah kontrak klien.

Pendekatan ini dipilih untuk menekan risiko putus kontrak frontend dibanding cutover sekaligus.

## Option Tradeoff

| Option                                 | Kelebihan                                | Risiko utama                                             | Keputusan                                         |
| -------------------------------------- | ---------------------------------------- | -------------------------------------------------------- | ------------------------------------------------- |
| Big-bang rewrite                       | satu waktu cutover, koordinasi sederhana | blast radius tinggi, rollback sulit jika DB ikut berubah | tidak dipilih                                     |
| Strangler per endpoint                 | risiko lebih kecil per area              | butuh kontrol ownership endpoint ketat                   | dipakai sebagian sebagai pola eksekusi cutover    |
| Parallel service penuh + switch instan | validasi parity kuat sebelum switch      | biaya infrastruktur dan observability naik               | dipilih sebagai basis, dengan cutover inkremental |
| API-compatible rewrite + coexistence   | kontrak klien stabil, rollback cepat     | butuh governance kompatibilitas disiplin                 | dipilih                                           |

## Compatibility Target

Semua endpoint publik pada `/api/v1` wajib menjaga kompatibilitas:

- path dan HTTP method tetap,
- status code untuk skenario sukses/gagal utama tetap,
- struktur response envelope tetap kompatibel,
- requirement autentikasi/otorisasi tidak melemah,
- perilaku bisnis inti tetap konsisten terhadap baseline.

Rincian per domain ada di [Compatibility Matrix](/Users/macbookpro/Development/recova-backend-v2/docs/roadmap/compatibility-matrix.md).

## Freeze Policy

Selama window migrasi:

- perubahan breaking pada kontrak API publik dibekukan,
- perubahan behavior di runtime lama hanya boleh untuk security, bug kritis, atau compliance,
- setiap perubahan runtime lama yang menyentuh kontrak API wajib direfleksikan ke runtime baru dalam siklus yang sama,
- perubahan schema database hanya boleh jika backward-compatible untuk kedua runtime aktif.

## Coexistence and Ownership Rule

Untuk mencegah konflik data saat dua runtime aktif:

- ownership write per domain harus tunggal pada satu waktu,
- endpoint yang belum cutover tetap dilayani runtime lama,
- endpoint yang sudah cutover dilayani runtime baru dengan guard kompatibilitas aktif,
- dual-write tanpa kontrak idempotensi eksplisit tidak diperbolehkan.

## Rollout Flow

1. Tetapkan baseline kontrak dari inventory endpoint saat ini.
2. Implementasikan endpoint setara di runtime Go untuk satu domain.
3. Jalankan contract test dan snapshot comparison terhadap baseline.
4. Aktifkan cutover inkremental domain tersebut.
5. Pantau error rate, latensi, dan mismatch response.
6. Lanjut ke domain berikutnya hanya bila gate terpenuhi.

## Rollback Direction

Rollback operasional:

- hentikan routing ke runtime Go untuk domain terdampak,
- kembalikan trafik ke runtime lama,
- lakukan investigasi mismatch kontrak atau issue runtime,
- re-enable cutover hanya setelah gate verifikasi ulang lulus.

Rollback data:

- perubahan schema harus punya rencana forward-fix dan rollback direction tertulis,
- migrasi destruktif tidak boleh dieksekusi tanpa window kompatibilitas yang aman untuk kedua runtime.

## Go / No-Go Gate Before Domain Cutover

Semua item wajib `pass` sebelum domain endpoint dipindahkan:

- contract test domain lulus,
- tidak ada mismatch pada response envelope dan error envelope utama,
- auth dan permission parity tervalidasi,
- observability minimum aktif: request id, structured log, health/readiness,
- rollback switch tervalidasi di environment pra-produksi.

## Exit Criteria

Migrasi dianggap selesai bila:

- seluruh domain `/api/v1` sudah dipindahkan ke runtime Go,
- tidak ada dependency runtime lama pada jalur request publik,
- SLO operasional pasca-cutover stabil pada periode observasi yang disepakati,
- runtime lama dapat dinonaktifkan tanpa dampak kontrak API.

## Risk Register

| Risiko                                        | Dampak                             | Mitigasi                                                      |
| --------------------------------------------- | ---------------------------------- | ------------------------------------------------------------- |
| Drift kontrak response antar runtime          | bug frontend dan regresi UX        | contract test otomatis + snapshot parity                      |
| Perubahan schema tidak kompatibel dua runtime | write failure atau data corruption | kebijakan backward-compatible migration + rollout inkremental |
| Auth behavior tidak setara                    | akses tidak sah atau false reject  | parity test authn/authz per domain                            |
| Observability kurang saat cutover             | incident lambat terdeteksi         | wajib request id, log terstruktur, health/readiness gate      |

## Related Documents

- [Current Express Baseline](/Users/macbookpro/Development/recova-backend-v2/docs/roadmap/express-baseline.md)
- [Feature Inventory](/Users/macbookpro/Development/recova-backend-v2/docs/roadmap/feature-inventory.md)
- [Current Runtime Inventory](/Users/macbookpro/Development/recova-backend-v2/docs/roadmap/current-runtime-inventory.md)
- [Compatibility Matrix](/Users/macbookpro/Development/recova-backend-v2/docs/roadmap/compatibility-matrix.md)
- [ADR 0001 Migration Strategy](/Users/macbookpro/Development/recova-backend-v2/docs/decisions/adr-0001-migration-strategy.md)

## Source Reference

- [references/README.md](/Users/macbookpro/Development/recova-backend-v2/references/README.md)
- [/Users/macbookpro/Development/bisakerja-api/docs/architecture.md](/Users/macbookpro/Development/bisakerja-api/docs/architecture.md)
- [/Users/macbookpro/Development/bisakerja-api/docs/operations/deployment.md](/Users/macbookpro/Development/bisakerja-api/docs/operations/deployment.md)
- [Fiber v3 Documentation](https://docs.gofiber.io/)
- [GORM Documentation](https://gorm.io/docs/)
- [golang-migrate](https://github.com/golang-migrate/migrate)
- [PostgreSQL Current Documentation](https://www.postgresql.org/docs/current/index.html)
- [Go Module Layout](https://go.dev/doc/modules/layout)
- [OpenTelemetry Go Documentation](https://opentelemetry.io/docs/languages/go/)
