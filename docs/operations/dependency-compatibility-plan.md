---
title: Recova Backend Dependency Compatibility Plan
description: Rencana kompatibilitas pembaruan keamanan dependensi, pinning versi tooling CI/CD, dan mitigasi risiko supply chain.
owner: backend-owner
reviewers:
  - engineering-lead
  - platform-docs-maintainer
doc_status: active
source_repo: recova-backend-v2
source_path: docs/operations/dependency-compatibility-plan.md
last_reviewed: 2026-05-18
---

# Recova Backend Dependency Compatibility Plan

Dokumen ini mencatat rencana kompatibilitas pembaruan dependensi keamanan, status penanganan kerentanan reachable, kebijakan pinning tooling, dan verifikasi non-breaking untuk runtime backend Recova.

## Prinsip Pembaruan Dependensi

1. **Bertahap dan Terukur**: Pembaruan dilakukan per komponen dengan verifikasi regresi langsung, bukan upgrade massal yang tidak terkendali.
2. **Prioritas Keamanan Reachable**: Fokus utama pada dependensi yang memiliki jalur eksekusi aktif dari kode aplikasi atau harness pengujian.
3. **Reproducibility Tooling**: Seluruh binary tooling CI dan build image menggunakan versi ter-pin, tanpa flag `@latest` bebas.
4. **Preservasi Kontrak API**: Pembaruan dependensi runtime tidak boleh mengubah kontrak publik OpenAPI, routing, atau schema response.

## Matriks Evaluasi Kerentanan dan Keputusan Pembaruan

| Dependensi                    | Versi Asal | Target Versi | Advisory / Kerentanan Terkait                                                                                     | Keputusan  | Bukti Kompatibilitas                                                                                                |
| ----------------------------- | ---------- | ------------ | ----------------------------------------------------------------------------------------------------------------- | ---------- | ------------------------------------------------------------------------------------------------------------------- |
| `github.com/jackc/pgx/v5`     | `v5.6.0`   | `v5.9.2`     | `GO-2026-5004`: SQL injection via dollar-quoted string placeholders                                               | **Update** | Kompatibel dengan GORM Postgres driver `v1.6.0`; seluruh test database dan repository lulus tanpa modifikasi query. |
| `github.com/gofiber/fiber/v3` | `v3.1.0`   | `v3.4.0`     | `GO-2026-5912`: Helmet middleware HSTS protocol header check                                                      | **Update** | Seluruh route handler, middleware stack, dan OpenAPI drift check lulus tanpa perubahan routing.                     |
| `google.golang.org/grpc`      | `v1.74.2`  | `v1.83.1`    | `GO-2026-6348` (HTTP/2 OOM) & `GO-2026-6061` (xDS RBAC/HTTP2)                                                     | **Update** | Diperbarui melalui `google.golang.org/api` `v0.264.0`; verifikasi token Google OAuth tetap valid dan stabil.        |
| `go.opentelemetry.io/otel`    | `v1.37.0`  | `v1.44.0`    | `GO-2026-5506`: Baggage header extraction memory allocation                                                       | **Update** | Diperbarui secara harmonis bersama sub-ekosistem OTel; metrics dan trace interceptor berfungsi normal.              |
| `golang.org/x/text`           | `v0.37.0`  | `v0.42.0`    | `GO-2026-6629` (precis panic) & `GO-2026-5970` (infinite loop)                                                    | **Update** | Pemrosesan karakter Unicode dan normalisasi teks pada database client stabil.                                       |
| `golang.org/x/net`            | `v0.55.0`  | `v0.60.0`    | `GO-2026-6617`, `GO-2026-6612`, `GO-2026-6611`, `GO-2026-6610`, `GO-2026-6603` (HTTP/2 transport & frame parsing) | **Update** | Kompatibel dengan Go 1.26 runtime; client AI HTTP dan unit test lulus.                                              |
| `golang.org/x/crypto`         | `v0.51.0`  | `v0.57.0`    | Ketergantungan harmonis dari pembaruan x/net dan x/text                                                           | **Update** | Password hashing (bcrypt) dan helper kriptografi tetap kompatibel.                                                  |

## Tooling dan Supply Chain Version Pinning

Untuk memastikan build pipeline deterministik dan mencegah silent drift pada CI/CD:

| Tool / Artefak                | Mekanisme Sebelumnya      | Konfigurasi Ter-pin                                                                    | Alasan dan Lokasi                                                                                                       |
| ----------------------------- | ------------------------- | -------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------- |
| `golang-migrate` CLI          | `go install ...@latest`   | `go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.19.1` | Pinned di `.github/workflows/ci.yml` dan `.github/workflows/cutover-waves.yml` agar identik dengan library di `go.mod`. |
| `golang-migrate` Docker Image | `migrate/migrate:v4.19.0` | `migrate/migrate:v4.19.1`                                                              | Pinned di `scripts/migrate.sh`, `README.md`, dan `.github/workflows/deploy-staging.yml`.                                |
| `govulncheck` CLI             | `go install ...@latest`   | `go install golang.org/x/vuln/cmd/govulncheck@v1.9.0`                                  | Pinned di `scripts/security-scan.sh` untuk stabilitas hasil scan lokal dan pipeline.                                    |
| Go Builder Image              | `golang:1.25.11-alpine`   | `golang:1.26.6-alpine`                                                                 | Pinned di `Dockerfile` untuk mendukung module directive `go 1.26.0` dan menyertakan patch keamanan Go 1.26 runtime.     |

## Residual Risk dan Mitigasi Runtime

1. **Standard Library Toolchain Scanner**:
   - `govulncheck` memeriksa implementasi standard library Go pada mesin yang menjalankan scan.
   - Bila workstation pengembang menggunakan toolchain lokal Go minor sebelum patch (misalnya `go1.27.1`), scanner dapat menampilkan peringatan standard library internal hingga toolchain lokal diperbarui ke patch rilis (misalnya `go1.27.2` atau Go security patch terkait).
   - Pada runtime container produksi, builder image dipatok ke `golang:1.26.6-alpine`, dan CI menggunakan `go.mod` version file (`go 1.26.0`), menjamin konsistensi build artefak.
2. **Ketergantungan Antar Package**:
   - Pembaruan `golang.org/x/net` ke `v0.60.0` mensyaratkan `go 1.26.0` pada `go.mod` dan `golang.org/x/text` minimal `v0.42.0`.
   - Seluruh kombinasi versi telah diverifikasi lewat `go mod tidy` dan pengujian penuh lintas paket.

## Rangkaian Verifikasi

Setiap pembaruan dependensi diverifikasi melalui gerbang berikut:

1. `make test` — seluruh unit dan handler test lulus.
2. `make lint` (`go vet ./...`) — analisis statis lulus tanpa peringatan.
3. `make openapi-check` — verifikasi kontrak OpenAPI tidak mengalami drift.
4. `make module-consistency-check` — validasi konsistensi modul dan arsitektur route.
5. `make test-integration` — verifikasi runner skrip integrasi operasional.
6. `make security-scan` — verifikasi seluruh kerentanan modul third-party telah teratasi.
7. Docker multi-stage build — verifikasi kompilasi kontainer berhasil bersih.
