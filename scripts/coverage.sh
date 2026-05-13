#!/usr/bin/env sh
set -eu

# Local coverage report helper.
# Default: focuses on production packages (`./internal/...`) and uses `-short` to avoid DB-backed E2E suite.

out_dir="${COVERAGE_DIR:-artifacts/coverage}"
pkg="${COVERAGE_PACKAGES:-./internal/...}"
mode="${COVERAGE_MODE:-atomic}"
short="${COVERAGE_SHORT:-1}"

mkdir -p "$out_dir"

profile="$out_dir/internal.out"
func_out="$out_dir/internal.func.txt"
by_area_json="$out_dir/by-area.json"
by_area_md="$out_dir/by-area.md"

echo "[coverage] packages=$pkg"
echo "[coverage] out_dir=$out_dir"

if [ "$short" = "1" ]; then
  echo "[coverage] go test -short -coverprofile=$profile"
  go test -short "$pkg" -covermode="$mode" -coverprofile="$profile"
else
  echo "[coverage] go test -coverprofile=$profile"
  go test "$pkg" -covermode="$mode" -coverprofile="$profile"
fi

echo "[coverage] go tool cover -func"
go tool cover -func="$profile" > "$func_out"

echo "[coverage] aggregate by area"
go run ./cmd/tools/coverage -coverprofile "$profile" -out-json "$by_area_json" -out-md "$by_area_md" >/dev/null

tail -n 1 "$func_out" | sed 's/^total:/[coverage] total:/' || true
echo "[coverage] wrote:"
echo "  - $profile"
echo "  - $func_out"
echo "  - $by_area_json"
echo "  - $by_area_md"

