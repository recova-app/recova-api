#!/usr/bin/env sh
set -eu

command="${1:-check}"

case "$command" in
  structure)
    go test -count=1 ./test/contract -run '^TestContract_ModuleStructureConsistency_Baseline$'
    ;;
  contract)
    go test -count=1 ./test/contract -run '^TestContract_ModuleLayerAndRouteConsistency_Baseline$'
    ;;
  openapi)
    ./scripts/openapi.sh check
    ;;
  check)
    go test -count=1 ./test/contract -run '^TestContract_Module.*Consistency'
    ;;
  full-check)
    go test -count=1 ./test/contract -run '^TestContract_Module.*Consistency'
    ./scripts/openapi.sh check
    ;;
  *)
    echo "[module-consistency] command not supported: $command" >&2
    echo "[module-consistency] use: structure|contract|openapi|check|full-check" >&2
    exit 1
    ;;
esac
