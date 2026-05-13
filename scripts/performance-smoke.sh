#!/usr/bin/env sh
set -eu

db_url="${RECOVA_DB_INTEGRATION_URL:-}"
if [ -z "$db_url" ]; then
  echo "[performance-smoke] RECOVA_DB_INTEGRATION_URL must be provided" >&2
  exit 1
fi
case "$db_url" in
  *_test*) ;;
  *)
    echo "[performance-smoke] RECOVA_DB_INTEGRATION_URL must point to database *_test" >&2
    exit 1
    ;;
esac

report_path="${RECOVA_PERF_REPORT_PATH:-artifacts/release-confidence/performance-smoke.json}"
mkdir -p "$(dirname "$report_path")"

RECOVA_PERF_REPORT_PATH="$report_path" go test -count=1 ./test/performance -run TestPerformance_LoadSmoke

echo "[performance-smoke] report: $report_path"
