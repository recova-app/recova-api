#!/usr/bin/env sh
set -eu

command="${1:-check}"

case "$command" in
  check)
    ./scripts/openapi.sh check
    go run ./cmd/tools/scalar check
    ;;
  preview)
    ./scripts/openapi.sh check
    echo "[scalar] mulai preview runtime di /docs/api (OpenAPI: /openapi.yaml)"
    ./scripts/with-env.sh go run ./cmd/api
    ;;
  *)
    echo "[scalar] command not supported: $command" >&2
    echo "[scalar] use: check | preview" >&2
    exit 1
    ;;
esac
