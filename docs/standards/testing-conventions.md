---
title: Recova Backend Testing Conventions
description: Konvensi implementasi test Go untuk Recova Backend meliputi struktur file test, table-driven tests, subtests, integration tests, dan assertion perilaku kontrak.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: draft
source_repo: recova-backend-v2
source_path: docs/standards/testing-conventions.md
last_reviewed: 2026-05-08
---

# Recova Backend Testing Conventions

Dokumen ini menetapkan pola test agar hasil verifikasi stabil, terbaca, dan mudah dirawat.

## Default Test Requirement (Mandatory)

- setiap pembuatan file kode baru default **wajib** disertai test,
- setiap pembuatan atau perubahan file konfigurasi yang memengaruhi runtime/build/deploy default **wajib** disertai test atau verification check otomatis,
- pengecualian hanya boleh jika memang tidak relevan secara teknis, dan harus mencantumkan alasan eksplisit pada PR/task.

Contoh pengecualian yang diperbolehkan:

- perubahan komentar atau typo tanpa dampak perilaku,
- perubahan dokumen non-eksekusi murni editorial,
- file konfigurasi yang tidak dieksekusi dan tidak memengaruhi runtime/CI.

Untuk semua pengecualian, tetap wajib ada bukti verifikasi minimal bahwa perubahan tidak mengubah perilaku sistem.

## Test File Structure

- unit test diletakkan berdampingan dengan package target (`*_test.go`),
- integration test dipisahkan jelas dari unit test,
- helper test jangan memakai prefix `test_` pada nama fungsi.

## Naming Conventions

- nama test harus deskriptif terhadap perilaku,
- gunakan pola `Test<Subject>_<Condition>_<Expectation>` bila relevan,
- gunakan `t.Run` untuk membagi skenario.

## Table-Driven Test Rules

Gunakan table-driven tests untuk logika bercabang.

Aturan:

- setiap case punya nama unik,
- input/expected disimpan eksplisit dalam struct case,
- jalankan per-case sebagai subtest.

## Subtests and Parallelism

- gunakan `t.Run` untuk isolasi skenario,
- `t.Parallel()` hanya pada test yang benar-benar tidak berbagi state,
- hindari parallel test jika memakai resource global yang sama.

## Assertions and Error Checks

- validasi `error` secara eksplisit,
- untuk sentinel/typed error gunakan `errors.Is`/`errors.As`,
- validasi output domain utama, bukan hanya non-nil check.

## HTTP Handler Test Rules

- verifikasi status code,
- verifikasi envelope success/error,
- verifikasi mapping request invalid ke error code yang tepat,
- sertakan skenario auth missing/invalid pada endpoint terproteksi.

## Repository and Database Test Rules

- gunakan database terisolasi untuk integration test,
- migration harus diterapkan sebelum test berjalan,
- data fixture harus deterministik,
- jangan hardcode kredensial sensitif.

## Contract and Regression Test Rules

- endpoint kritis harus punya regression tests,
- perubahan kontrak response wajib update test contract,
- bug yang sudah diperbaiki harus memiliki test pencegah regresi.

## CI Expectations

- seluruh test wajib lulus sebelum merge,
- failure pada migration/integration test memblokir release,
- test harus bisa dijalankan tanpa secret production.
- perubahan file/config baru tanpa test companion dianggap belum siap merge kecuali ada exception rationale yang disetujui reviewer.

## Related Documents

- [Testing Strategy](/Users/macbookpro/Development/recova-backend-v2/docs/operations/testing.md)
- [Verification Matrix](/Users/macbookpro/Development/recova-backend-v2/docs/operations/verification-matrix.md)
- [Error Handling Standard](/Users/macbookpro/Development/recova-backend-v2/docs/standards/error-handling.md)

## Source Reference

- [Go `testing` Package](https://go.dev/pkg/testing/?m=old)
- [Using Subtests and Sub-benchmarks](https://go.dev/blog/subtests)
- [Go Command Documentation](https://pkg.go.dev/cmd/go)
