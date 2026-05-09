#!/usr/bin/env sh
set -eu

artifact_dir="${MAINTENANCE_ARTIFACT_DIR:-artifacts/maintenance}"
execution_id="${MAINTENANCE_EXECUTION_ID:-$(date -u +%Y%m%dT%H%M%SZ)}"
docs_root="${DOCS_ROOT:-docs}"
security_cadence_days="${SECURITY_DOC_CADENCE_DAYS:-30}"
general_cadence_days="${GENERAL_DOC_CADENCE_DAYS:-90}"
alert_review_status="${ALERT_REVIEW_STATUS:-pending}"
slo_review_status="${SLO_REVIEW_STATUS:-pending}"
dependency_cadence_review_status="${DEPENDENCY_CADENCE_REVIEW_STATUS:-pending}"
default_owner="${DEFAULT_BACKLOG_OWNER:-backend-owner}"
default_due_date="${DEFAULT_BACKLOG_DUE_DATE:-$(date -u +%Y-%m-%d)}"

case "$security_cadence_days" in
  ''|*[!0-9]*)
    echo "[post-migration-maintenance] SECURITY_DOC_CADENCE_DAYS must be numeric: $security_cadence_days" >&2
    exit 1
    ;;
esac

case "$general_cadence_days" in
  ''|*[!0-9]*)
    echo "[post-migration-maintenance] GENERAL_DOC_CADENCE_DAYS must be numeric: $general_cadence_days" >&2
    exit 1
    ;;
esac

if [ ! -d "$docs_root" ]; then
  echo "[post-migration-maintenance] docs root not found: $docs_root" >&2
  exit 1
fi

date_days_ago() {
  days="$1"
  if date -u -d "1970-01-02" +%Y-%m-%d >/dev/null 2>&1; then
    date -u -d "$days days ago" +%Y-%m-%d
    return
  fi
  date -u -v-"${days}"d +%Y-%m-%d
}

validate_review_status() {
  label="$1"
  value="$2"
  case "$value" in
    done) ;;
    *)
      echo "[post-migration-maintenance] ${label} must be 'done' (actual: ${value})" >&2
      exit 1
      ;;
  esac
}

validate_review_status "ALERT_REVIEW_STATUS" "$alert_review_status"
validate_review_status "SLO_REVIEW_STATUS" "$slo_review_status"
validate_review_status "DEPENDENCY_CADENCE_REVIEW_STATUS" "$dependency_cadence_review_status"

mkdir -p "$artifact_dir"
summary_file="$artifact_dir/${execution_id}-summary.log"
freshness_report_file="$artifact_dir/${execution_id}-docs-freshness-report.tsv"
backlog_file="$artifact_dir/${execution_id}-maintenance-backlog.md"
report_file="$artifact_dir/${execution_id}-maintenance-report.json"
started_at="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
security_cutoff="$(date_days_ago "$security_cadence_days")"
general_cutoff="$(date_days_ago "$general_cadence_days")"

log() {
  printf '%s\n' "$*" | tee -a "$summary_file"
}

is_high_risk_doc() {
  path="$1"
  case "$path" in
    */security*|*/auth*|*/privacy*|*/deployment*)
      return 0
      ;;
    *)
      return 1
      ;;
  esac
}

extract_last_reviewed() {
  file="$1"
  value="$(awk '/^last_reviewed:[[:space:]]*/{print $2; exit}' "$file" | tr -d '"')"
  printf '%s' "$value"
}

extract_doc_title() {
  file="$1"
  title="$(awk '/^title:[[:space:]]*/{sub(/^title:[[:space:]]*/, "", $0); print; exit}' "$file" | sed 's/^"//; s/"$//')"
  if [ -z "$title" ]; then
    title="$(basename "$file" .md)"
  fi
  printf '%s' "$title"
}

printf 'path\tlast_reviewed\tcadence_days\tcutoff_date\tstale\n' >"$freshness_report_file"

stale_total=0
stale_high_risk=0
stale_general=0

for file in $(find "$docs_root" -type f -name '*.md' | sort); do
  if [ "$file" = "$docs_root/generated/release-confidence-report.md" ]; then
    continue
  fi

  last_reviewed="$(extract_last_reviewed "$file")"
  if [ -z "$last_reviewed" ]; then
    continue
  fi

  cadence_days="$general_cadence_days"
  cutoff_date="$general_cutoff"
  if is_high_risk_doc "$file"; then
    cadence_days="$security_cadence_days"
    cutoff_date="$security_cutoff"
  fi

  stale="false"
  if [ "$last_reviewed" \< "$cutoff_date" ]; then
    stale="true"
    stale_total=$((stale_total + 1))
    if is_high_risk_doc "$file"; then
      stale_high_risk=$((stale_high_risk + 1))
    else
      stale_general=$((stale_general + 1))
    fi
  fi

  printf '%s\t%s\t%s\t%s\t%s\n' "$file" "$last_reviewed" "$cadence_days" "$cutoff_date" "$stale" >>"$freshness_report_file"
done

cat >"$backlog_file" <<EOF
# Maintenance Backlog

| ID | Item | Owner | Priority | Status | Due Date |
| --- | --- | --- | --- | --- | --- |
| MNT-001 | Review and update stale high-risk documents (${stale_high_risk} item) | ${default_owner} | high | todo | ${default_due_date} |
| MNT-002 | Review and update stale general-category documents (${stale_general} item) | ${default_owner} | medium | todo | ${default_due_date} |
| MNT-003 | Review observability alert thresholds after cutover | ${default_owner} | high | todo | ${default_due_date} |
| MNT-004 | Review SLO targets and error budget after cutover | ${default_owner} | high | todo | ${default_due_date} |
| MNT-005 | Review dependency update cadence and security patching | ${default_owner} | medium | todo | ${default_due_date} |
| MNT-006 | Prioritize follow-up optimizations based on stabilization evidence | ${default_owner} | medium | todo | ${default_due_date} |

## Evidence

- Freshness report: \`${freshness_report_file}\`
- Alert review status: \`${alert_review_status}\`
- SLO review status: \`${slo_review_status}\`
- Dependency cadence review status: \`${dependency_cadence_review_status}\`
EOF

finished_at="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
cat >"$report_file" <<EOF
{
  "suite": "post-migration-maintenance",
  "status": "passed",
  "executionId": "${execution_id}",
  "startedAt": "${started_at}",
  "finishedAt": "${finished_at}",
  "checks": {
    "docsRoot": "${docs_root}",
    "staleDocsTotal": ${stale_total},
    "staleHighRiskDocs": ${stale_high_risk},
    "staleGeneralDocs": ${stale_general},
    "securityDocCadenceDays": ${security_cadence_days},
    "generalDocCadenceDays": ${general_cadence_days},
    "alertReviewStatus": "${alert_review_status}",
    "sloReviewStatus": "${slo_review_status}",
    "dependencyCadenceReviewStatus": "${dependency_cadence_review_status}"
  },
  "artifacts": {
    "freshnessReportPath": "${freshness_report_file}",
    "maintenanceBacklogPath": "${backlog_file}"
  }
}
EOF

log "[post-migration-maintenance] stale_docs_total=${stale_total}"
log "[post-migration-maintenance] stale_high_risk=${stale_high_risk} stale_general=${stale_general}"
log "[post-migration-maintenance] freshness_report=${freshness_report_file}"
log "[post-migration-maintenance] backlog=${backlog_file}"
log "[post-migration-maintenance] report=${report_file}"
log "[post-migration-maintenance] success"
