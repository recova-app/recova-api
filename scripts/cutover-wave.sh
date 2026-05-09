#!/usr/bin/env sh
set -eu

wave_input="${1:-${CUTOVER_WAVE:-all}}"
artifact_dir="${CUTOVER_ARTIFACT_DIR:-artifacts/cutover}"
execution_id="${CUTOVER_EXECUTION_ID:-$(date -u +%Y%m%dT%H%M%SZ)}"
api_base_url="${API_BASE_URL:-http://127.0.0.1:${APP_PORT:-3000}}"
go_bin="${GO_BIN:-go}"
curl_bin="${CURL_BIN:-curl}"
run_staging_deploy="${RUN_STAGING_DEPLOY:-auto}"
run_rollback_on_failure="${RUN_ROLLBACK_ON_FAILURE:-false}"
rollback_command="${CUTOVER_ROLLBACK_COMMAND:-}"
staging_deploy_script="${STAGING_DEPLOY_SCRIPT:-./scripts/staging-deploy.sh}"
e2e_script="${E2E_SCRIPT:-./scripts/e2e-critical.sh}"

case "$wave_input" in
  all)
    waves="64 65 66 67 68"
    ;;
  64|65|66|67|68)
    waves="$wave_input"
    ;;
  *)
    echo "[cutover-wave] wave invalid: $wave_input" >&2
    echo "[cutover-wave] use: 64|65|66|67|68|all" >&2
    exit 1
    ;;
esac

case "$run_staging_deploy" in
  true|false|auto) ;;
  *)
    echo "[cutover-wave] RUN_STAGING_DEPLOY invalid: $run_staging_deploy (use true|false|auto)" >&2
    exit 1
    ;;
esac

case "$run_rollback_on_failure" in
  true|false) ;;
  *)
    echo "[cutover-wave] RUN_ROLLBACK_ON_FAILURE invalid: $run_rollback_on_failure (use true|false)" >&2
    exit 1
    ;;
esac

if [ "$run_rollback_on_failure" = "true" ] && [ -z "$rollback_command" ]; then
  echo "[cutover-wave] CUTOVER_ROLLBACK_COMMAND must be provided when RUN_ROLLBACK_ON_FAILURE=true" >&2
  exit 1
fi

if ! command -v "$go_bin" >/dev/null 2>&1; then
  echo "[cutover-wave] go command not found: $go_bin" >&2
  exit 1
fi

if ! command -v "$curl_bin" >/dev/null 2>&1; then
  echo "[cutover-wave] curl command not found: $curl_bin" >&2
  exit 1
fi

if [ ! -f "$staging_deploy_script" ]; then
  echo "[cutover-wave] staging deploy script not found: $staging_deploy_script" >&2
  exit 1
fi

if [ ! -f "$e2e_script" ]; then
  echo "[cutover-wave] e2e script not found: $e2e_script" >&2
  exit 1
fi

mkdir -p "$artifact_dir"
summary_file="$artifact_dir/${execution_id}-summary.log"
wave_report_file() {
  wave="$1"
  printf '%s/%s-wave-%s-e2e.json' "$artifact_dir" "$execution_id" "$wave"
}

log() {
  printf '%s\n' "$*" | tee -a "$summary_file"
}

require_e2e_db_url() {
  db_url="${RECOVA_DB_INTEGRATION_URL:-}"
  if [ -z "$db_url" ]; then
    echo "[cutover-wave] RECOVA_DB_INTEGRATION_URL must be provided for wave domain" >&2
    exit 1
  fi
  case "$db_url" in
    *_test*) ;;
    *)
      echo "[cutover-wave] RECOVA_DB_INTEGRATION_URL must point to database *_test" >&2
      exit 1
      ;;
  esac
}

run_wave_logic() {
  wave="$1"
  echo "[cutover-wave] wave $wave start"
  case "$wave" in
    64)
      if [ "$run_staging_deploy" = "true" ] || [ "$run_staging_deploy" = "auto" ]; then
        "$staging_deploy_script" || return 1
      fi
      "$curl_bin" -fsS --retry 6 --retry-delay 2 --retry-connrefused "${api_base_url}/health/live" >/dev/null || return 1
      "$curl_bin" -fsS --retry 6 --retry-delay 2 --retry-connrefused "${api_base_url}/health/ready" >/dev/null || return 1
      "$go_bin" test -count=1 ./test/contract -run '^TestContract_HealthResponses_ValidAgainstOpenAPI$' || return 1
      ;;
    65)
      require_e2e_db_url
      "$go_bin" test -count=1 ./test/contract -run '^TestContract_AuthRouteParity_ValidAgainstOpenAPI$' || return 1
      RECOVA_E2E_SCOPE="wave65" RECOVA_E2E_REPORT_PATH="$(wave_report_file "$wave")" "$e2e_script" || return 1
      ;;
    66)
      require_e2e_db_url
      RECOVA_E2E_SCOPE="wave66" RECOVA_E2E_REPORT_PATH="$(wave_report_file "$wave")" "$e2e_script" || return 1
      ;;
    67)
      require_e2e_db_url
      RECOVA_E2E_SCOPE="wave67" RECOVA_E2E_REPORT_PATH="$(wave_report_file "$wave")" "$e2e_script" || return 1
      ;;
    68)
      require_e2e_db_url
      RECOVA_E2E_SCOPE="wave68" RECOVA_E2E_REPORT_PATH="$(wave_report_file "$wave")" "$e2e_script" || return 1
      ;;
    *)
      echo "[cutover-wave] unknown wave: $wave" >&2
      exit 1
      ;;
  esac
  echo "[cutover-wave] wave $wave pass"
}

execute_rollback_if_needed() {
  failed_wave="$1"
  if [ "$run_rollback_on_failure" != "true" ]; then
    return
  fi
  log "[cutover-wave] rollback trigger for wave ${failed_wave}"
  if sh -c "$rollback_command" >>"$summary_file" 2>&1; then
    log "[cutover-wave] rollback command success"
  else
    log "[cutover-wave] rollback command failed"
  fi
}

log "[cutover-wave] execution_id=$execution_id wave_input=$wave_input"

for wave in $waves; do
  wave_log="$artifact_dir/${execution_id}-wave-${wave}.log"
  if run_wave_logic "$wave" >"$wave_log" 2>&1; then
    log "[cutover-wave] wave ${wave} pass log=${wave_log}"
  else
    log "[cutover-wave] wave ${wave} failed log=${wave_log}"
    execute_rollback_if_needed "$wave"
    cat "$wave_log" >&2
    exit 1
  fi
done

log "[cutover-wave] all requested waves passed"
