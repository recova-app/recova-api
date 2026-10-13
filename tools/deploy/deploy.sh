#!/usr/bin/env bash

# ==============================================================================
# Recova Backend — Production Standalone Deployment Script
# ==============================================================================
# Executes minimal-downtime release (~2-3s container recreate) on single-server VPS
# running Docker Compose behind host Nginx reverse proxy with Neon PostgreSQL.
#
# Usage:
#   ./deploy.sh <IMAGE_TAG>
#   IMAGE_TAG=sha-a1b2c3d ./deploy.sh
#   ALLOW_MOVING_TAG=true ./deploy.sh main
#
# Arguments:
#   $1 (required)    : Immutable container image tag (e.g. sha-4c9f1a2).
#                      Refuses moving tag 'main' unless ALLOW_MOVING_TAG=true is set.
#
# Environment Variables:
#   DEPLOY_DIR        : Application root directory (default: /opt/recova)
#   DEPLOY_LOCK_FILE  : Concurrency lockfile path (default: /tmp/recova-deploy.lock)
#   IMAGE_TAG         : Immutable container image tag (e.g. sha-4c9f1a2)
#   ALLOW_MOVING_TAG  : Allow unpinned 'main' or 'latest' tag during emergencies (default: false)
#   FORCE_DB_SNAPSHOT : Force pre-migration database snapshot even if no pending migrations (default: false)
#   FORCE_DB_MIGRATE  : Force database migration execution even if counts match (default: false)
#   ENVIRONMENT       : Environment descriptor (default: production)
# ==============================================================================

set -euo pipefail

DEPLOY_DIR="${DEPLOY_DIR:-/opt/recova}"
IMAGE_TAG="${1:-${IMAGE_TAG:-}}"
ALLOW_MOVING_TAG="${ALLOW_MOVING_TAG:-false}"
ENVIRONMENT="${ENVIRONMENT:-production}"

if [ "${DEPLOY_SOURCE_ONLY:-false}" != "true" ]; then
  if [ -z "${IMAGE_TAG}" ]; then
    echo "Error: Missing required container image tag." >&2
    echo "Usage: ./deploy.sh <IMAGE_TAG> (e.g. ./deploy.sh sha-a1b2c3d)" >&2
    echo "To explicitly allow deploying moving tags ('main' or 'latest') during manual emergencies, run:" >&2
    echo "  ALLOW_MOVING_TAG=true ./deploy.sh main" >&2
    exit 1
  fi

  if { [ "${IMAGE_TAG}" = "main" ] || [ "${IMAGE_TAG}" = "latest" ]; } && [ "${ALLOW_MOVING_TAG}" != "true" ]; then
    echo "Error: Refusing deployment with moving tag '${IMAGE_TAG}'. Production deployments must specify an immutable SHA tag (e.g. sha-a1b2c3d)." >&2
    echo "To explicitly allow moving tag during manual bootstrap or emergency, execute with ALLOW_MOVING_TAG=true:" >&2
    echo "  ALLOW_MOVING_TAG=true ./deploy.sh ${IMAGE_TAG}" >&2
    exit 1
  fi

  # Sanitize image tag format to prevent path traversal or malformed filenames
  if ! [[ "${IMAGE_TAG}" =~ ^[a-zA-Z0-9_][a-zA-Z0-9_.-]*$ ]]; then
    echo "Error: Invalid image tag format '${IMAGE_TAG}'. Image tags must start with an alphanumeric character or underscore and contain only alphanumeric, underscores, dots, or dashes." >&2
    exit 1
  fi

  # Concurrency Lock: Prevents overlapping automated and manual deployment executions on single VPS
  LOCK_FILE="${DEPLOY_LOCK_FILE:-/tmp/recova-deploy.lock}"
  if command -v flock >/dev/null 2>&1; then
    exec 9>"${LOCK_FILE}"
    if ! flock -n 9; then
      echo "Error: Another deployment process is currently running (lock held on ${LOCK_FILE})." >&2
      exit 1
    fi
  fi

  export IMAGE_TAG
  export COMPOSE_FILE="docker-compose.yml"

  echo "========================================================================"
  echo "Recova Backend Deployment Initiated"
  echo "Target Directory: ${DEPLOY_DIR}"
  echo "Target Image Tag: ${IMAGE_TAG}"
  echo "Timestamp:        $(date -u '+%Y-%m-%d %H:%M:%S UTC')"
  echo "========================================================================"

  if [ ! -d "${DEPLOY_DIR}" ]; then
    echo "Error: Deployment directory '${DEPLOY_DIR}' does not exist." >&2
    exit 1
  fi

  cd "${DEPLOY_DIR}"

  if [ ! -f "docker-compose.yml" ]; then
    echo "Error: docker-compose.yml not found in '${DEPLOY_DIR}'." >&2
    exit 1
  fi

  if [ -f ".env" ]; then
    # shellcheck disable=SC1091
    set -a
    source .env
    set +a
  else
    echo "Warning: .env file not found in '${DEPLOY_DIR}'. Ensure environment variables are loaded." >&2
  fi

  # ------------------------------------------------------------------------------
  # 0. Detect Active Deployment State & Register Rollback Trap
  # ------------------------------------------------------------------------------
  PREVIOUS_IMAGE_TAG=""
  if docker inspect recova-api >/dev/null 2>&1; then
    CURRENT_API_IMAGE="$(docker inspect --format '{{.Config.Image}}' recova-api 2>/dev/null || true)"
    IMAGE_WITHOUT_DIGEST="${CURRENT_API_IMAGE%@*}"
    if [[ "${IMAGE_WITHOUT_DIGEST}" == *:* ]]; then
      CANDIDATE_TAG="${IMAGE_WITHOUT_DIGEST##*:}"
      if [[ "${CANDIDATE_TAG}" =~ ^[a-zA-Z0-9_][a-zA-Z0-9_.-]*$ ]]; then
        PREVIOUS_IMAGE_TAG="${CANDIDATE_TAG}"
        echo "Detected existing active container image tag: '${PREVIOUS_IMAGE_TAG}'"
      fi
    fi
  fi
fi

DEPLOYMENT_SUCCESS="false"
SNAPSHOT_FILE=""

rollback_on_failure() {
  local exit_code=$?
  trap - EXIT INT TERM

  if [ "${DEPLOYMENT_SUCCESS}" = "true" ]; then
    return 0
  fi

  if [ "${exit_code}" -eq 0 ]; then
    exit_code=1
  fi

  echo "========================================================================" >&2
  echo "CRITICAL: Deployment failed with exit code ${exit_code}." >&2
  echo "Initiating automated recovery procedure..." >&2
  echo "========================================================================" >&2

  if [ -n "${PREVIOUS_IMAGE_TAG}" ] && [ "${PREVIOUS_IMAGE_TAG}" != "${IMAGE_TAG}" ]; then
    echo "[ROLLBACK] Reverting application container to previous image tag '${PREVIOUS_IMAGE_TAG}'..." >&2
    IMAGE_TAG="${PREVIOUS_IMAGE_TAG}" docker compose up -d --remove-orphans api || {
      echo "[ROLLBACK ERROR] Failed to recreate container with previous image tag '${PREVIOUS_IMAGE_TAG}'." >&2
    }

    echo "[ROLLBACK] Waiting for previous application container to recover..." >&2
    ROLLBACK_ATTEMPTS=20
    ROLLBACK_ATTEMPT=0
    ROLLBACK_HEALTHY="false"

    until [ "${ROLLBACK_ATTEMPT}" -ge "${ROLLBACK_ATTEMPTS}" ]; do
      ROLLBACK_ATTEMPT=$((ROLLBACK_ATTEMPT + 1))
      if curl -fsS "http://127.0.0.1:${PORT:-3001}/health/live" >/dev/null 2>&1; then
        ROLLBACK_HEALTHY="true"
        break
      fi
      sleep 2
    done

    if [ "${ROLLBACK_HEALTHY}" = "true" ]; then
      echo "[ROLLBACK SUCCESS] Restored and verified previous stable container (tag: '${PREVIOUS_IMAGE_TAG}')." >&2
    else
      echo "[ROLLBACK WARNING] Application health probe still failing after ${ROLLBACK_ATTEMPTS} attempts during rollback." >&2
    fi
  elif [ "${PREVIOUS_IMAGE_TAG}" = "${IMAGE_TAG}" ]; then
    echo "[ROLLBACK] Previous image tag matches target tag ('${IMAGE_TAG}'). Skipping container revert." >&2
  else
    echo "[ROLLBACK] No previous image tag available for automated container rollback." >&2
  fi

  if [ -n "${SNAPSHOT_FILE}" ] && [ -s "${SNAPSHOT_FILE}" ] && gzip -t "${SNAPSHOT_FILE}" 2>/dev/null; then
    echo "[ROLLBACK NOTICE] Verified database pre-migration snapshot is preserved at: ${SNAPSHOT_FILE}" >&2
    echo "[ROLLBACK NOTICE] To restore database snapshot manually if needed:" >&2
    echo "  gunzip -c ${SNAPSHOT_FILE} | docker run -i --rm postgres:17-alpine psql \"\${DATABASE_URL}\"" >&2
  elif [ -n "${SNAPSHOT_FILE}" ] && [ -f "${SNAPSHOT_FILE}" ]; then
    echo "[ROLLBACK NOTICE] Removing invalid or empty pre-migration snapshot: ${SNAPSHOT_FILE}" >&2
    rm -f "${SNAPSHOT_FILE}"
  fi

  exit "${exit_code}"
}

if [ "${DEPLOY_SOURCE_ONLY:-false}" = "true" ]; then
  return 0 2>/dev/null || exit 0
fi

trap rollback_on_failure EXIT INT TERM

# ------------------------------------------------------------------------------
# 1. Pull Target Application Image
# ------------------------------------------------------------------------------
echo "[DEPLOY 1/7] Pulling prebuilt application image for tag '${IMAGE_TAG}'..."
docker compose pull api

# ------------------------------------------------------------------------------
# 2. Verify Neon Database Connectivity
# ------------------------------------------------------------------------------
echo "[DEPLOY 2/7] Verifying Neon PostgreSQL database connectivity..."
if [ -z "${DATABASE_URL:-}" ]; then
  echo "Error: DATABASE_URL environment variable is not set." >&2
  exit 1
fi

DB_MAX_ATTEMPTS=15
DB_ATTEMPT=0
DB_READY="false"

until docker run --rm postgres:17-alpine pg_isready -d "${DATABASE_URL}" -q 2>/dev/null; do
  DB_ATTEMPT=$((DB_ATTEMPT + 1))
  if [ "${DB_ATTEMPT}" -ge "${DB_MAX_ATTEMPTS}" ]; then
    echo "Error: Neon database failed to respond within timeout." >&2
    exit 1
  fi
  echo "Waiting for Neon database connectivity... (attempt ${DB_ATTEMPT}/${DB_MAX_ATTEMPTS})"
  sleep 2
done
echo "Neon database is reachable and accepting connections."

# ------------------------------------------------------------------------------
# 3. Check Pending Database Migrations & Conditional Pre-Migration Snapshot
# ------------------------------------------------------------------------------
MIGRATIONS_DIR="${DEPLOY_DIR}/migrations"
APPLIED_VERSION=0
LATEST_FILE_VERSION=0
NEED_MIGRATION="false"

# Query applied version from schema_migrations table (golang-migrate standard)
APPLIED_VERSION="$(docker run --rm postgres:17-alpine psql "${DATABASE_URL}" -t -A -c "SELECT COALESCE(max(version), 0) FROM schema_migrations WHERE dirty = false;" 2>/dev/null || echo "0")"
APPLIED_VERSION="$(echo "${APPLIED_VERSION}" | tr -d '[:space:]')"
if [ -z "${APPLIED_VERSION}" ]; then
  APPLIED_VERSION=0
fi

# Find highest version in migrations/*.up.sql
if [ -d "${MIGRATIONS_DIR}" ]; then
  LATEST_FILE_VERSION="$(ls -1 "${MIGRATIONS_DIR}"/*.up.sql 2>/dev/null | sed -E 's/.*\/([0-9]+)_.*/\1/' | sort -n | tail -n 1 || echo "0")"
  LATEST_FILE_VERSION="$(echo "${LATEST_FILE_VERSION}" | tr -d '[:space:]')"
fi
if [ -z "${LATEST_FILE_VERSION}" ]; then
  LATEST_FILE_VERSION=0
fi

echo "Database migration state: applied=${APPLIED_VERSION}, latest_file=${LATEST_FILE_VERSION}"

if [ "${FORCE_DB_MIGRATE:-false}" = "true" ] || [ "${FORCE_DB_SNAPSHOT:-false}" = "true" ]; then
  NEED_MIGRATION="true"
  echo "Database migration forced via environment flag (FORCE_DB_MIGRATE / FORCE_DB_SNAPSHOT)."
elif [ "${LATEST_FILE_VERSION}" -gt "${APPLIED_VERSION}" ]; then
  NEED_MIGRATION="true"
  echo "Detected pending database migration(s): current=${APPLIED_VERSION}, target=${LATEST_FILE_VERSION}."
else
  echo "Database schema is up to date. Skipping pre-migration snapshot and migration step."
fi

if [ "${NEED_MIGRATION}" = "true" ]; then
  BACKUP_DIR="${BACKUP_DIR:-${DEPLOY_DIR}/backups}"
  mkdir -p "${BACKUP_DIR}"
  SNAPSHOT_FILE="${BACKUP_DIR}/recova_pre_migrate_$(date -u '+%Y%m%d_%H%M%S')_${IMAGE_TAG}.sql.gz"

  echo "[DEPLOY 3/7] Creating automated pre-migration database snapshot at '${SNAPSHOT_FILE}'..."
  docker run --rm postgres:17-alpine pg_dump \
    "${DATABASE_URL}" \
    --clean \
    --if-exists \
    --no-owner \
    --no-privileges | gzip > "${SNAPSHOT_FILE}"

  if [ ! -s "${SNAPSHOT_FILE}" ] || ! gzip -t "${SNAPSHOT_FILE}" 2>/dev/null; then
    echo "Error: Pre-migration database snapshot failed or produced an invalid archive: ${SNAPSHOT_FILE}" >&2
    rm -f "${SNAPSHOT_FILE}"
    exit 1
  fi

  echo "Pre-migration database snapshot verified successfully ($(du -h "${SNAPSHOT_FILE}" | cut -f1))."

  # Retain last 10 pre-migration snapshots to protect disk space
  find "${BACKUP_DIR}" -maxdepth 1 -type f -name "recova_pre_migrate_*.sql.gz" | sort -r | tail -n +11 | while IFS= read -r old_snap; do
    if [ -n "${old_snap}" ] && [ -f "${old_snap}" ]; then
      rm -f "${old_snap}"
    fi
  done || true

  # ----------------------------------------------------------------------------
  # 4. Execute Database Migrations (Pre-Deployment Release Step)
  # ----------------------------------------------------------------------------
  echo "[DEPLOY 4/7] Applying database migrations..."
  MIGRATE_URL="${DATABASE_URL}"
  case "${MIGRATE_URL}" in
    postgresql://*)
      MIGRATE_URL="postgres://${MIGRATE_URL#postgresql://}"
      ;;
  esac

  docker run --rm \
    -v "${MIGRATIONS_DIR}:/migrations:ro" \
    migrate/migrate:v4.19.1 \
    -path /migrations \
    -database "${MIGRATE_URL}" \
    up

  echo "Database migrations applied successfully."
else
  echo "[DEPLOY 3/7] Pre-migration database snapshot skipped (no pending schema changes)."
  echo "[DEPLOY 4/7] Database migrations skipped (schema already at latest version)."
fi

# ------------------------------------------------------------------------------
# 5. Recreate & Start Application Container
# ------------------------------------------------------------------------------
echo "[DEPLOY 5/7] Recreating application container with image tag '${IMAGE_TAG}'..."
docker compose up -d --remove-orphans api

# ------------------------------------------------------------------------------
# 6. Post-Deployment Verification & Multi-Route Smoke Probes
# ------------------------------------------------------------------------------
echo "[DEPLOY 6/7] Running post-deployment health verification and smoke probes..."
APP_PORT="${PORT:-3001}"
HEALTH_MAX_ATTEMPTS=20
HEALTH_ATTEMPT=0
HEALTH_URL="http://127.0.0.1:${APP_PORT}/health/live"

until curl -fsS "${HEALTH_URL}" >/dev/null 2>&1; do
  HEALTH_ATTEMPT=$((HEALTH_ATTEMPT + 1))
  if [ "${HEALTH_ATTEMPT}" -ge "${HEALTH_MAX_ATTEMPTS}" ]; then
    echo "Error: Application liveness probe failed at ${HEALTH_URL} after ${HEALTH_MAX_ATTEMPTS} attempts." >&2
    echo "Dumping container statuses:" >&2
    docker compose ps >&2
    echo "Dumping recent API container logs:" >&2
    docker compose logs --tail 30 api >&2
    exit 1
  fi
  echo "Waiting for application liveness probe response... (attempt ${HEALTH_ATTEMPT}/${HEALTH_MAX_ATTEMPTS})"
  sleep 2
done
echo "Liveness probe responded HTTP 200 OK."

# Probe readiness (verifies DB connection pool inside Fiber app)
READY_URL="http://127.0.0.1:${APP_PORT}/health/ready"
echo "Verifying application readiness probe (${READY_URL})..."
if ! curl -fsS "${READY_URL}" >/dev/null 2>&1; then
  echo "Error: Application readiness probe failed at ${READY_URL}." >&2
  docker compose logs --tail 30 api >&2
  exit 1
fi
echo "Readiness probe responded HTTP 200 OK (Neon DB connection operational)."

# Probe OpenAPI documentation artifact
OPENAPI_URL="http://127.0.0.1:${APP_PORT}/openapi.yaml"
echo "Verifying OpenAPI schema availability (${OPENAPI_URL})..."
if ! curl -fsS "${OPENAPI_URL}" >/dev/null 2>&1; then
  echo "Error: OpenAPI schema endpoint failed at ${OPENAPI_URL}." >&2
  exit 1
fi
echo "OpenAPI schema probe responded HTTP 200 OK."

# Smoke test auth boundary: protected endpoint must reject unauthenticated request
AUTH_PROBE_URL="http://127.0.0.1:${APP_PORT}/api/v1/users/me"
echo "Verifying protected auth boundary (${AUTH_PROBE_URL})..."
AUTH_STATUS="$(curl -sS -o /dev/null -w '%{http_code}' "${AUTH_PROBE_URL}" || true)"
case "${AUTH_STATUS}" in
  401|403)
    echo "Auth boundary passed: protected route rejected unauthorized request with HTTP ${AUTH_STATUS}."
    ;;
  *)
    echo "Error: Protected route returned unexpected HTTP status ${AUTH_STATUS} (expected 401 or 403)." >&2
    exit 1
    ;;
esac

# Optional public host ingress verification if PUBLIC_APP_URL is set
PUBLIC_INGRESS_URL="${PUBLIC_APP_URL:-}"
if [ -n "${PUBLIC_INGRESS_URL}" ]; then
  echo "Verifying public ingress on ${PUBLIC_INGRESS_URL}..."
  if curl -fsS --max-time 10 "${PUBLIC_INGRESS_URL}/health/live" >/dev/null 2>&1; then
    echo "Public ingress probe responded HTTP 200 OK on ${PUBLIC_INGRESS_URL}."
  else
    echo "Notice: Public ingress probe to ${PUBLIC_INGRESS_URL}/health/live was unreachable from host (verify DNS/SSL/Cloudflare Tunnel routing)."
  fi
fi

# ------------------------------------------------------------------------------
# 7. Cleanup Dangling Images
# ------------------------------------------------------------------------------
echo "[DEPLOY 7/7] Cleaning up dangling images..."
docker image prune -f >/dev/null 2>&1 || true

DEPLOYMENT_SUCCESS="true"

echo "========================================================================"
echo "Recova Backend Deployment Completed Successfully"
echo "Active Image Tag: ${IMAGE_TAG}"
echo "Timestamp:        $(date -u '+%Y-%m-%d %H:%M:%S UTC')"
echo "========================================================================"
