#!/usr/bin/env sh
set -eu

db_url="${RECOVA_DB_INTEGRATION_URL:-}"
if [ -z "$db_url" ]; then
  echo "[e2e-critical] RECOVA_DB_INTEGRATION_URL must be provided" >&2
  exit 1
fi
case "$db_url" in
  *_test*) ;;
  *)
    echo "[e2e-critical] RECOVA_DB_INTEGRATION_URL must point to database *_test" >&2
    exit 1
    ;;
esac

scope="${RECOVA_E2E_SCOPE:-all}"
case "$scope" in
  all|wave64|wave65|wave66|wave67|wave68) ;;
  *)
    echo "[e2e-critical] RECOVA_E2E_SCOPE invalid: $scope" >&2
    exit 1
    ;;
esac

report_path="${RECOVA_E2E_REPORT_PATH:-artifacts/release-confidence/e2e-critical-flows.json}"
mkdir -p "$(dirname "$report_path")"

RECOVA_E2E_REPORT_PATH="$report_path" go test -count=1 ./test/e2e -run TestE2E_CriticalFlows

echo "[e2e-critical] scope: $scope"
echo "[e2e-critical] report: $report_path"
