#!/usr/bin/env sh
set -eu

command="${1:-check}"

case "$command" in
  generate|check)
    ;;
  *)
    echo "[openapi] command tidak didukung: $command" >&2
    echo "[openapi] gunakan: generate | check" >&2
    exit 1
    ;;
esac

go run ./cmd/tools/openapi "$command"
