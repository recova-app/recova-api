# OpenAPI Source

Direktori ini menyimpan source contract OpenAPI yang menjadi acuan endpoint publik.

## File Utama

- `api/openapi/openapi.yaml`: source contract kanonik.
- `docs/generated/openapi.yaml`: artefak generated yang wajib sinkron.

## Perintah Standar

- `make openapi-generate`: validasi source spec dan sinkronisasi artefak generated.
- `make openapi-check`: validasi source+generated spec dan cek drift route runtime.
