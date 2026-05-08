#!/usr/bin/env sh
set -eu

required_cmds="go git make"

for cmd in $required_cmds; do
  if ! command -v "$cmd" >/dev/null 2>&1; then
    echo "[preflight] missing required command: $cmd" >&2
    exit 1
  fi
done

if ! command -v migrate >/dev/null 2>&1; then
  echo "[preflight] warning: 'migrate' binary not found (required for migrate-up/migrate-down)"
fi

goversion_raw="$(go env GOVERSION)"
goversion="${goversion_raw#go}"
major="$(printf '%s' "$goversion" | cut -d. -f1)"
minor="$(printf '%s' "$goversion" | cut -d. -f2)"

if [ -z "$major" ] || [ -z "$minor" ]; then
  echo "[preflight] unable to parse Go version: $goversion_raw" >&2
  exit 1
fi

if [ "$major" -lt 1 ] || { [ "$major" -eq 1 ] && [ "$minor" -lt 25 ]; }; then
  echo "[preflight] go >= 1.25 required, found $goversion_raw" >&2
  exit 1
fi

if [ ! -f go.mod ]; then
  echo "[preflight] go.mod not found" >&2
  exit 1
fi

if [ ! -f cmd/api/main.go ]; then
  echo "[preflight] cmd/api/main.go not found" >&2
  exit 1
fi

if [ ! -d migrations ]; then
  echo "[preflight] migrations directory not found" >&2
  exit 1
fi

echo "[preflight] environment looks good"
