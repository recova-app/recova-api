#!/usr/bin/env sh
set -eu

artifact_dir="${ROLLBACK_REHEARSAL_ARTIFACT_DIR:-artifacts/rollback-rehearsal}"
execution_id="${ROLLBACK_REHEARSAL_EXECUTION_ID:-$(date -u +%Y%m%dT%H%M%SZ)}"
cutover_script="${CUTOVER_SCRIPT:-./scripts/cutover-wave.sh}"
wave="${ROLLBACK_REHEARSAL_WAVE:-65}"
rollback_command="${ROLLBACK_REHEARSAL_COMMAND:-}"

if [ ! -f "$cutover_script" ]; then
  echo "[rollback-rehearsal] cutover script not found: $cutover_script" >&2
  exit 1
fi

if [ -z "$rollback_command" ]; then
  echo "[rollback-rehearsal] ROLLBACK_REHEARSAL_COMMAND must be provided" >&2
  exit 1
fi

case "$wave" in
  65|66|67|68) ;;
  *)
    echo "[rollback-rehearsal] ROLLBACK_REHEARSAL_WAVE invalid: $wave (use 65|66|67|68)" >&2
    exit 1
    ;;
esac

db_url="${RECOVA_DB_INTEGRATION_URL:-}"
if [ -z "$db_url" ]; then
  echo "[rollback-rehearsal] RECOVA_DB_INTEGRATION_URL must be provided" >&2
  exit 1
fi
case "$db_url" in
  *_test*) ;;
  *)
    echo "[rollback-rehearsal] RECOVA_DB_INTEGRATION_URL must point to database *_test" >&2
    exit 1
    ;;
esac

mkdir -p "$artifact_dir"
summary_file="$artifact_dir/${execution_id}-summary.log"
report_file="$artifact_dir/${execution_id}-rollback-rehearsal-report.json"
rollback_marker_file="$artifact_dir/${execution_id}-rollback-marker.log"
started_at="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

log() {
  printf '%s\n' "$*" | tee -a "$summary_file"
}

temp_dir="$(mktemp -d)"
cleanup() {
  rm -rf "$temp_dir"
}
trap cleanup EXIT INT TERM

fake_go="$temp_dir/go"
cat >"$fake_go" <<'SCRIPT'
#!/usr/bin/env sh
printf '%s\n' "$*" >> "$ROLLBACK_REHEARSAL_GO_LOG"
exit 0
SCRIPT
chmod +x "$fake_go"

failing_e2e="$temp_dir/e2e-fail.sh"
cat >"$failing_e2e" <<'SCRIPT'
#!/usr/bin/env sh
printf '%s|%s\n' "${RECOVA_E2E_SCOPE:-}" "${RECOVA_E2E_REPORT_PATH:-}" >> "$ROLLBACK_REHEARSAL_E2E_LOG"
report="${RECOVA_E2E_REPORT_PATH:-}"
if [ -n "$report" ]; then
  mkdir -p "$(dirname "$report")"
  printf '{"status":"failed","reason":"rollback-rehearsal-injected-failure"}\n' > "$report"
fi
exit 1
SCRIPT
chmod +x "$failing_e2e"

rollback_wrapper="$temp_dir/rollback-wrapper.sh"
cat >"$rollback_wrapper" <<'SCRIPT'
#!/usr/bin/env sh
set -eu
printf '%s\n' "rollback-invoked" >> "$ROLLBACK_REHEARSAL_MARKER_FILE"
sh -c "$ROLLBACK_REHEARSAL_COMMAND"
SCRIPT
chmod +x "$rollback_wrapper"

go_log_file="$artifact_dir/${execution_id}-go.log"
e2e_log_file="$artifact_dir/${execution_id}-e2e.log"

log "[rollback-rehearsal] execution_id=${execution_id}"
log "[rollback-rehearsal] wave=${wave}"
log "[rollback-rehearsal] start injected-failure cutover"

set +e
ROLLBACK_REHEARSAL_GO_LOG="$go_log_file" \
ROLLBACK_REHEARSAL_E2E_LOG="$e2e_log_file" \
ROLLBACK_REHEARSAL_MARKER_FILE="$rollback_marker_file" \
ROLLBACK_REHEARSAL_COMMAND="$rollback_command" \
GO_BIN="$fake_go" \
RUN_STAGING_DEPLOY="false" \
RUN_ROLLBACK_ON_FAILURE="true" \
CUTOVER_ROLLBACK_COMMAND="$rollback_wrapper" \
E2E_SCRIPT="$failing_e2e" \
CUTOVER_ARTIFACT_DIR="$artifact_dir" \
CUTOVER_EXECUTION_ID="${execution_id}-cutover" \
"$cutover_script" "$wave" >>"$summary_file" 2>&1
cutover_exit_code=$?
set -e

if [ "$cutover_exit_code" -eq 0 ]; then
  echo "[rollback-rehearsal] cutover rehearsal should fail but exited 0" >&2
  exit 1
fi

if [ ! -f "$rollback_marker_file" ]; then
  echo "[rollback-rehearsal] rollback marker not found, rollback command not invoked" >&2
  exit 1
fi

finished_at="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
cat >"$report_file" <<EOF
{
  "suite": "rollback-rehearsal",
  "status": "passed",
  "executionId": "${execution_id}",
  "startedAt": "${started_at}",
  "finishedAt": "${finished_at}",
  "wave": "${wave}",
  "cutoverExitCode": ${cutover_exit_code},
  "rollbackMarkerPath": "${rollback_marker_file}",
  "goLogPath": "${go_log_file}",
  "e2eLogPath": "${e2e_log_file}"
}
EOF

log "[rollback-rehearsal] report: ${report_file}"
log "[rollback-rehearsal] success"
