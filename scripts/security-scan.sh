#!/usr/bin/env sh
set -eu

target="${1:-./...}"

if [ -n "${GOVULNCHECK_BIN:-}" ]; then
  if [ ! -x "$GOVULNCHECK_BIN" ]; then
    echo "[security-scan] GOVULNCHECK_BIN tidak executable: $GOVULNCHECK_BIN" >&2
    exit 1
  fi
  scan_bin="$GOVULNCHECK_BIN"
elif command -v govulncheck >/dev/null 2>&1; then
  scan_bin="$(command -v govulncheck)"
else
  go_bin="$(go env GOPATH)/bin/govulncheck"
  if [ ! -x "$go_bin" ]; then
    echo "[security-scan] govulncheck tidak ditemukan, install otomatis..." >&2
    go install golang.org/x/vuln/cmd/govulncheck@latest
  fi
  scan_bin="$go_bin"
fi

echo "[security-scan] using $scan_bin"
"$scan_bin" "$target"
