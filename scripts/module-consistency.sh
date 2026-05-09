#!/usr/bin/env sh
set -eu

command="${1:-check}"

case "$command" in
  check)
    go test -count=1 ./test/contract -run '^TestContract_Module.*Consistency'
    ;;
  *)
    echo "[module-consistency] command not supported: $command" >&2
    echo "[module-consistency] use: check" >&2
    exit 1
    ;;
esac
