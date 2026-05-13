#!/usr/bin/env sh
set -eu

artifact_dir="${DECOMMISSION_ARTIFACT_DIR:-artifacts/decommission}"
execution_id="${DECOMMISSION_EXECUTION_ID:-$(date -u +%Y%m%dT%H%M%SZ)}"
legacy_runtime_label="${LEGACY_RUNTIME_LABEL:-express}"
legacy_traffic_count="${LEGACY_RUNTIME_TRAFFIC_COUNT:-}"
legacy_traffic_evidence_file="${LEGACY_RUNTIME_TRAFFIC_EVIDENCE_FILE:-}"
legacy_archive_paths="${LEGACY_ARCHIVE_PATHS:-references}"
rollback_evidence_dir="${ROLLBACK_EVIDENCE_DIR:-artifacts/rollback-rehearsal}"
rollback_retention_days="${ROLLBACK_RETENTION_DAYS:-90}"
rollback_min_reports="${ROLLBACK_MIN_REPORTS:-1}"
runtime_source_of_truth_file="${RUNTIME_SOURCE_OF_TRUTH_FILE:-docs/roadmap/current-runtime-inventory.md}"
runtime_keyword="${RUNTIME_SOURCE_OF_TRUTH_KEYWORD:-Go}"

if [ -z "$legacy_traffic_count" ]; then
  echo "[runtime-decommission] LEGACY_RUNTIME_TRAFFIC_COUNT must be provided" >&2
  exit 1
fi

case "$legacy_traffic_count" in
  *[!0-9]*)
    echo "[runtime-decommission] LEGACY_RUNTIME_TRAFFIC_COUNT must be numeric: $legacy_traffic_count" >&2
    exit 1
    ;;
esac

case "$rollback_retention_days" in
  ''|*[!0-9]*)
    echo "[runtime-decommission] ROLLBACK_RETENTION_DAYS must be numeric: $rollback_retention_days" >&2
    exit 1
    ;;
esac

case "$rollback_min_reports" in
  ''|*[!0-9]*)
    echo "[runtime-decommission] ROLLBACK_MIN_REPORTS must be numeric: $rollback_min_reports" >&2
    exit 1
    ;;
esac

if [ "$rollback_min_reports" -lt 1 ]; then
  echo "[runtime-decommission] ROLLBACK_MIN_REPORTS minimal 1" >&2
  exit 1
fi

if [ "$rollback_retention_days" -lt 1 ]; then
  echo "[runtime-decommission] ROLLBACK_RETENTION_DAYS minimal 1" >&2
  exit 1
fi

if [ "$legacy_traffic_count" -ne 0 ]; then
  echo "[runtime-decommission] traffic runtime legacy not zero yet: $legacy_traffic_count" >&2
  exit 1
fi

if [ -n "$legacy_traffic_evidence_file" ] && [ ! -f "$legacy_traffic_evidence_file" ]; then
  echo "[runtime-decommission] file evidence traffic not found: $legacy_traffic_evidence_file" >&2
  exit 1
fi

if [ ! -d "$rollback_evidence_dir" ]; then
  echo "[runtime-decommission] directory rollback evidence not found: $rollback_evidence_dir" >&2
  exit 1
fi

if [ ! -f "$runtime_source_of_truth_file" ]; then
  echo "[runtime-decommission] source of truth runtime not found: $runtime_source_of_truth_file" >&2
  exit 1
fi

if ! grep -F -- "$runtime_keyword" "$runtime_source_of_truth_file" >/dev/null 2>&1; then
  echo "[runtime-decommission] source of truth runtime does not include yet keyword '$runtime_keyword'" >&2
  exit 1
fi

mkdir -p "$artifact_dir"
summary_file="$artifact_dir/${execution_id}-summary.log"
report_file="$artifact_dir/${execution_id}-decommission-report.json"
archive_file="$artifact_dir/${execution_id}-${legacy_runtime_label}-archive.tar.gz"
started_at="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

log() {
  printf '%s\n' "$*" | tee -a "$summary_file"
}

rollback_reports_count="$(find "$rollback_evidence_dir" -maxdepth 1 -type f -name '*-rollback-rehearsal-report.json' | wc -l | tr -d '[:space:]')"
if [ "$rollback_reports_count" -lt "$rollback_min_reports" ]; then
  echo "[runtime-decommission] rollback evidence insufficient: minimum $rollback_min_reports, actual $rollback_reports_count" >&2
  exit 1
fi

for report in "$rollback_evidence_dir"/*-rollback-rehearsal-report.json; do
  if [ ! -f "$report" ]; then
    continue
  fi
  if ! grep -E '"status"[[:space:]]*:[[:space:]]*"passed"' "$report" >/dev/null 2>&1; then
    echo "[runtime-decommission] rollback report not in status passed: $report" >&2
    exit 1
  fi
done

recent_rollback_reports_count="$(find "$rollback_evidence_dir" -maxdepth 1 -type f -name '*-rollback-rehearsal-report.json' -mtime "-$rollback_retention_days" | wc -l | tr -d '[:space:]')"
if [ "$recent_rollback_reports_count" -lt 1 ]; then
  echo "[runtime-decommission] no rollback evidence within retention window $rollback_retention_days days" >&2
  exit 1
fi

archive_candidates=""
for candidate in $legacy_archive_paths; do
  if [ ! -e "$candidate" ]; then
    echo "[runtime-decommission] archive path not found: $candidate" >&2
    exit 1
  fi
  archive_candidates="$archive_candidates $candidate"
done

if ! tar -czf "$archive_file" $archive_candidates >/dev/null 2>&1; then
  echo "[runtime-decommission] failed to create archive legacy runtime" >&2
  exit 1
fi

traffic_evidence_copy=""
if [ -n "$legacy_traffic_evidence_file" ]; then
  traffic_evidence_copy="$artifact_dir/${execution_id}-legacy-traffic-evidence.log"
  cp "$legacy_traffic_evidence_file" "$traffic_evidence_copy"
fi

finished_at="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

cat >"$report_file" <<EOF
{
  "suite": "runtime-decommission",
  "status": "passed",
  "executionId": "${execution_id}",
  "legacyRuntime": "${legacy_runtime_label}",
  "startedAt": "${started_at}",
  "finishedAt": "${finished_at}",
  "checks": {
    "legacyTrafficCount": ${legacy_traffic_count},
    "rollbackEvidenceDir": "${rollback_evidence_dir}",
    "rollbackEvidenceCount": ${rollback_reports_count},
    "recentRollbackEvidenceCount": ${recent_rollback_reports_count},
    "rollbackRetentionDays": ${rollback_retention_days},
    "runtimeSourceOfTruthFile": "${runtime_source_of_truth_file}",
    "runtimeSourceOfTruthKeyword": "${runtime_keyword}"
  },
  "artifacts": {
    "legacyArchivePath": "${archive_file}",
    "legacyTrafficEvidencePath": "${traffic_evidence_copy}"
  }
}
EOF

log "[runtime-decommission] legacy_runtime=$legacy_runtime_label"
log "[runtime-decommission] rollback_evidence_count=$rollback_reports_count recent_within_${rollback_retention_days}d=$recent_rollback_reports_count"
log "[runtime-decommission] archive=$archive_file"
if [ -n "$traffic_evidence_copy" ]; then
  log "[runtime-decommission] traffic_evidence_copy=$traffic_evidence_copy"
fi
log "[runtime-decommission] report=$report_file"
log "[runtime-decommission] success"
