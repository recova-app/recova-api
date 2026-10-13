#!/usr/bin/env bash

# ==============================================================================
# Recova Backend — Neon Database Snapshot & Disaster Recovery Utility
# ==============================================================================
# Manages manual and automated Neon PostgreSQL database snapshots, restores,
# integrity validation, and retention pruning on VPS.
#
# Usage:
#   ./backup.sh backup [LABEL]       # Creates a compressed database snapshot (.sql.gz)
#   ./backup.sh restore <FILE>       # Restores database from a snapshot
#   ./backup.sh list                 # Lists available database backup snapshots
#   ./backup.sh verify <FILE>        # Validates integrity of a snapshot
#   ./backup.sh prune [KEEP_COUNT]   # Retains latest N local backups (default: 10)
#
# Environment Variables:
#   DEPLOY_DIR : Application root directory (default: current dir or /opt/recova)
#   BACKUP_DIR : Target backup directory (default: ${DEPLOY_DIR}/backups)
#   DATABASE_URL : Neon PostgreSQL connection URL
# ==============================================================================

set -euo pipefail

DEPLOY_DIR="${DEPLOY_DIR:-$(pwd)}"
BACKUP_DIR="${BACKUP_DIR:-${DEPLOY_DIR}/backups}"
DEFAULT_KEEP_COUNT=10

mkdir -p "${BACKUP_DIR}"

if [ -f "${DEPLOY_DIR}/.env" ]; then
  # shellcheck disable=SC1091
  set -a
  source "${DEPLOY_DIR}/.env"
  set +a
fi

ensure_db_url() {
  if [ -z "${DATABASE_URL:-}" ]; then
    echo "Error: DATABASE_URL environment variable is not defined." >&2
    exit 1
  fi
}

cmd_backup() {
  local label="${1:-manual}"
  local timestamp
  timestamp="$(date -u '+%Y%m%d_%H%M%S')"
  local target_file="${BACKUP_DIR}/recova_${timestamp}_${label}.sql.gz"

  ensure_db_url
  echo "Initiating database snapshot to: ${target_file}"

  docker run --rm postgres:17-alpine pg_dump \
    "${DATABASE_URL}" \
    --clean \
    --if-exists \
    --no-owner \
    --no-privileges | gzip > "${target_file}"

  if [ ! -s "${target_file}" ] || ! gzip -t "${target_file}" 2>/dev/null; then
    echo "Error: Backup failed or produced corrupted archive: ${target_file}" >&2
    rm -f "${target_file}"
    exit 1
  fi

  echo "Backup successfully created and verified ($(du -h "${target_file}" | cut -f1)): ${target_file}"
}

cmd_restore() {
  local target_file="${1:-}"

  if [ -z "${target_file}" ] || [ ! -f "${target_file}" ]; then
    echo "Error: Snapshot file does not exist: ${target_file}" >&2
    exit 1
  fi

  if ! gzip -t "${target_file}" 2>/dev/null; then
    echo "Error: Corrupted archive: ${target_file}" >&2
    exit 1
  fi

  ensure_db_url

  echo "WARNING: Restoring will overwrite existing data in the target database!"
  echo "Target file: ${target_file}"
  echo "Target URL:  ${DATABASE_URL%%@*}@..." # Obfuscate credentials

  if [ "${FORCE_RESTORE:-false}" != "true" ]; then
    read -r -p "Type 'RESTORE' to proceed: " confirmation
    if [ "${confirmation}" != "RESTORE" ]; then
      echo "Restore aborted."
      exit 1
    fi
  fi

  echo "Restoring database..."
  gunzip -c "${target_file}" | docker run -i --rm postgres:17-alpine psql "${DATABASE_URL}"

  echo "Restore completed successfully."
}

cmd_verify() {
  local target_file="${1:-}"
  if [ -z "${target_file}" ] || [ ! -f "${target_file}" ]; then
    echo "Error: File not found: ${target_file}" >&2
    exit 1
  fi

  if gzip -t "${target_file}" 2>/dev/null; then
    echo "Archive integrity valid ($(du -h "${target_file}" | cut -f1)): ${target_file}"
  else
    echo "Error: Archive integrity check failed: ${target_file}" >&2
    exit 1
  fi
}

cmd_list() {
  echo "Available snapshots in ${BACKUP_DIR}:"
  find "${BACKUP_DIR}" -maxdepth 1 -name "*.sql.gz" -printf "%T+ %s %p\n" 2>/dev/null | sort -r || {
    ls -lh "${BACKUP_DIR}"/*.sql.gz 2>/dev/null || echo "No snapshots found."
  }
}

cmd_prune() {
  local keep_count="${1:-${DEFAULT_KEEP_COUNT}}"
  echo "Pruning older snapshots (retaining newest ${keep_count})..."

  find "${BACKUP_DIR}" -maxdepth 1 -name "*.sql.gz" | sort -r | tail -n "+$((keep_count + 1))" | while IFS= read -r old_file; do
    if [ -n "${old_file}" ] && [ -f "${old_file}" ]; then
      echo "Removing: ${old_file}"
      rm -f "${old_file}"
    fi
  done
  echo "Prune complete."
}

case "${1:-}" in
  backup)
    cmd_backup "${2:-manual}"
    ;;
  restore)
    cmd_restore "${2:-}"
    ;;
  verify)
    cmd_verify "${2:-}"
    ;;
  list)
    cmd_list
    ;;
  prune)
    cmd_prune "${2:-${DEFAULT_KEEP_COUNT}}"
    ;;
  *)
    echo "Usage: ./backup.sh <backup|restore|verify|list|prune> [args]" >&2
    exit 1
    ;;
esac
