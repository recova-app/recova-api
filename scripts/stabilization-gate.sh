#!/usr/bin/env sh
set -eu

artifact_dir="${STABILIZATION_ARTIFACT_DIR:-artifacts/stabilization}"
execution_id="${STABILIZATION_EXECUTION_ID:-$(date -u +%Y%m%dT%H%M%SZ)}"
go_bin="${GO_BIN:-go}"
openapi_script="${OPENAPI_SCRIPT:-./scripts/openapi.sh}"
e2e_script="${E2E_SCRIPT:-./scripts/e2e-critical.sh}"
performance_script="${PERFORMANCE_SCRIPT:-./scripts/performance-smoke.sh}"
run_full_regression="${RUN_FULL_REGRESSION:-true}"
run_openapi_check="${RUN_OPENAPI_CHECK:-true}"

case "$run_full_regression" in
  true|false) ;;
  *)
    echo "[stabilization-gate] RUN_FULL_REGRESSION invalid: $run_full_regression (use true|false)" >&2
    exit 1
    ;;
esac

case "$run_openapi_check" in
  true|false) ;;
  *)
    echo "[stabilization-gate] RUN_OPENAPI_CHECK invalid: $run_openapi_check (use true|false)" >&2
    exit 1
    ;;
esac

if ! command -v "$go_bin" >/dev/null 2>&1; then
  echo "[stabilization-gate] go command not found: $go_bin" >&2
  exit 1
fi

if [ ! -f "$openapi_script" ]; then
  echo "[stabilization-gate] openapi script not found: $openapi_script" >&2
  exit 1
fi

if [ ! -f "$e2e_script" ]; then
  echo "[stabilization-gate] e2e script not found: $e2e_script" >&2
  exit 1
fi

if [ ! -f "$performance_script" ]; then
  echo "[stabilization-gate] performance script not found: $performance_script" >&2
  exit 1
fi

db_url="${RECOVA_DB_INTEGRATION_URL:-}"
if [ -z "$db_url" ]; then
  echo "[stabilization-gate] RECOVA_DB_INTEGRATION_URL must be provided" >&2
  exit 1
fi
case "$db_url" in
  *_test*) ;;
  *)
    echo "[stabilization-gate] RECOVA_DB_INTEGRATION_URL must point to database *_test" >&2
    exit 1
    ;;
esac

mkdir -p "$artifact_dir"
summary_file="$artifact_dir/${execution_id}-summary.log"
e2e_report_path="$artifact_dir/${execution_id}-e2e-critical-flows.json"
perf_report_path="$artifact_dir/${execution_id}-performance-smoke.json"
stabilization_report_path="$artifact_dir/${execution_id}-stabilization-report.json"
started_at="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

log() {
  printf '%s\n' "$*" | tee -a "$summary_file"
}

run_gate() {
  gate_name="$1"
  shift
  log "[stabilization-gate] gate start: ${gate_name}"
  if "$@" >>"$summary_file" 2>&1; then
    log "[stabilization-gate] gate pass: ${gate_name}"
    return 0
  fi
  log "[stabilization-gate] gate failed: ${gate_name}"
  return 1
}

assert_report_passed() {
  report_file="$1"
  report_label="$2"
  if [ ! -f "$report_file" ]; then
    echo "[stabilization-gate] report ${report_label} not found: $report_file" >&2
    return 1
  fi
  if ! grep -E '"status"[[:space:]]*:[[:space:]]*"passed"' "$report_file" >/dev/null 2>&1; then
    echo "[stabilization-gate] report ${report_label} not in status passed: $report_file" >&2
    return 1
  fi
}

log "[stabilization-gate] execution_id=${execution_id}"
log "[stabilization-gate] artifact_dir=${artifact_dir}"

if [ "$run_full_regression" = "true" ]; then
  run_gate "full-regression-go-test" "$go_bin" test ./... || exit 1
fi

if [ "$run_openapi_check" = "true" ]; then
  run_gate "openapi-check" "$openapi_script" check || exit 1
fi

run_gate "critical-flows-e2e" env RECOVA_E2E_REPORT_PATH="$e2e_report_path" "$e2e_script" || exit 1
assert_report_passed "$e2e_report_path" "e2e-critical-flows" || exit 1

run_gate "performance-smoke" env RECOVA_PERF_REPORT_PATH="$perf_report_path" "$performance_script" || exit 1
assert_report_passed "$perf_report_path" "performance-smoke" || exit 1

finished_at="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

cat >"$stabilization_report_path" <<EOF
{
  "suite": "stabilization-gate",
  "status": "passed",
  "executionId": "${execution_id}",
  "startedAt": "${started_at}",
  "finishedAt": "${finished_at}",
  "checks": {
    "fullRegressionEnabled": ${run_full_regression},
    "openapiCheckEnabled": ${run_openapi_check},
    "e2eReportPath": "${e2e_report_path}",
    "performanceReportPath": "${perf_report_path}"
  }
}
EOF

log "[stabilization-gate] report: ${stabilization_report_path}"
log "[stabilization-gate] success"
