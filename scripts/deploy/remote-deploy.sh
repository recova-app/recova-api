#!/usr/bin/env sh
set -eu

app_dir="${1:?APP_DIR is required}"
deploy_branch="${2:?DEPLOY_BRANCH is required}"
image_ref="${3:?IMAGE_REF is required}"
compose_file="${4:-docker-compose.staging.yml}"
runtime_env_file="${5:-.env.staging}"
deploy_target="${6:-staging}"
expected_app_env="${7:-staging}"
app_port="${8:-3001}"
public_base_url="${9:-}"
migrate_bin="${MIGRATE_BIN:-migrate}"
curl_connect_timeout="${CURL_CONNECT_TIMEOUT:-5}"
curl_max_time="${CURL_MAX_TIME:-10}"
smoke_base_url="${SMOKE_BASE_URL:-}"
run_public_smoke="${RUN_PUBLIC_SMOKE:-false}"

if [ -z "$smoke_base_url" ]; then
  smoke_base_url="http://127.0.0.1:${app_port}"
fi

log() {
  printf '[remote-deploy] %s\n' "$1"
}

require_command() {
  if ! command -v "$1" >/dev/null 2>&1; then
    printf '[remote-deploy] missing command: %s\n' "$1" >&2
    exit 1
  fi
}

require_file() {
  if [ ! -f "$1" ]; then
    printf '[remote-deploy] missing file: %s\n' "$1" >&2
    exit 1
  fi
}

read_env_value() {
  key="$1"
  awk -v lookup_key="$key" '
    BEGIN { single_quote=sprintf("%c", 39) }
    /^[[:space:]]*#/ { next }
    {
      line=$0
      sub(/\r$/, "", line)
      eq=index(line, "=")
      if (eq == 0) {
        next
      }
      env_key=substr(line, 1, eq-1)
      gsub(/^[[:space:]]+|[[:space:]]+$/, "", env_key)
      if (env_key != lookup_key) {
        next
      }
      env_value=substr(line, eq+1)
      gsub(/^[[:space:]]+|[[:space:]]+$/, "", env_value)
      first=substr(env_value, 1, 1)
      last=substr(env_value, length(env_value), 1)
      if ((first == "\"" && last == "\"") || (first == single_quote && last == single_quote)) {
        env_value=substr(env_value, 2, length(env_value)-2)
      }
      print env_value
      exit
    }
  ' "$runtime_env_file"
}

compose() {
  docker compose --env-file "$runtime_env_file" -f "$compose_file" "$@"
}

show_diagnostics() {
  set +e
  log "diagnostics: compose ps"
  compose ps || true
  log "diagnostics: compose logs tail"
  compose logs --tail=120 api || true
  log "diagnostics: migrate status"
  DATABASE_URL="$database_url" MIGRATE_BIN="$migrate_bin" ./scripts/migrate.sh status || true
  log "diagnostics: health/live"
  curl -fsS --connect-timeout "$curl_connect_timeout" --max-time "$curl_max_time" "$smoke_base_url/health/live" || true
  log "diagnostics: health/ready"
  curl -fsS --connect-timeout "$curl_connect_timeout" --max-time "$curl_max_time" "$smoke_base_url/health/ready" || true
  if [ -n "$public_base_url" ]; then
    log "diagnostics: public health/live"
    curl -fsS --connect-timeout "$curl_connect_timeout" --max-time "$curl_max_time" "$public_base_url/health/live" || true
  fi
}

require_command git
require_command docker
require_command curl

cd "$app_dir"

if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  printf '[remote-deploy] target is not git repository: %s\n' "$app_dir" >&2
  exit 1
fi

require_file "$runtime_env_file"
require_file "$compose_file"

if ! git diff --quiet || ! git diff --cached --quiet; then
  printf '[remote-deploy] working tree not clean; abort deploy\n' >&2
  git status --short >&2
  exit 1
fi

declared_app_env="$(read_env_value APP_ENV)"
if [ -z "$declared_app_env" ]; then
  printf '[remote-deploy] APP_ENV missing in %s\n' "$runtime_env_file" >&2
  exit 1
fi
if [ "$declared_app_env" != "$expected_app_env" ]; then
  printf '[remote-deploy] APP_ENV mismatch. expected=%s actual=%s\n' "$expected_app_env" "$declared_app_env" >&2
  exit 1
fi

database_url="$(read_env_value DATABASE_URL)"
if [ -z "$database_url" ]; then
  printf '[remote-deploy] DATABASE_URL missing in %s\n' "$runtime_env_file" >&2
  exit 1
fi

previous_app_image="$(read_env_value APP_IMAGE)"

trap 'show_diagnostics' EXIT

log "sync branch ${deploy_branch}"
git fetch origin "$deploy_branch" --prune
git checkout -B "$deploy_branch" "origin/$deploy_branch"

log "set APP_IMAGE in env file"
if grep -q '^APP_IMAGE=' "$runtime_env_file"; then
  sed -i.bak "s|^APP_IMAGE=.*$|APP_IMAGE=${image_ref}|" "$runtime_env_file"
else
  printf '\nAPP_IMAGE=%s\n' "$image_ref" >> "$runtime_env_file"
fi
rm -f "$runtime_env_file.bak"
chmod 600 "$runtime_env_file"

export DATABASE_URL="$database_url"
export DIRECT_DATABASE_URL="$database_url"
export MIGRATE_BIN="$migrate_bin"
export RUNTIME_ENV_FILE="$runtime_env_file"

log "pull image ${image_ref}"
compose pull api

log "run migrate up"
./scripts/migrate.sh up
./scripts/migrate.sh check

log "start api"
compose up -d --wait --wait-timeout 180 api

log "smoke health"
curl -fsS --connect-timeout "$curl_connect_timeout" --max-time "$curl_max_time" --retry 6 --retry-delay 2 --retry-connrefused "$smoke_base_url/health/live" >/dev/null
curl -fsS --connect-timeout "$curl_connect_timeout" --max-time "$curl_max_time" --retry 6 --retry-delay 2 --retry-connrefused "$smoke_base_url/health/ready" >/dev/null

log "smoke openapi"
curl -fsS --connect-timeout "$curl_connect_timeout" --max-time "$curl_max_time" --retry 4 --retry-delay 2 "$smoke_base_url/openapi.yaml" >/dev/null

log "smoke protected route unauthorized"
status_code="$(curl -sS --connect-timeout "$curl_connect_timeout" --max-time "$curl_max_time" -o /dev/null -w '%{http_code}' "$smoke_base_url/api/v1/users/me")"
if [ "$status_code" != "401" ] && [ "$status_code" != "403" ]; then
  printf '[remote-deploy] expected protected route reject (401/403), got %s\n' "$status_code" >&2
  exit 1
fi

if [ "$run_public_smoke" = "true" ] && [ -n "$public_base_url" ]; then
  log "smoke public health"
  curl -fsS --connect-timeout "$curl_connect_timeout" --max-time "$curl_max_time" --retry 2 --retry-delay 2 "$public_base_url/health/live" >/dev/null
  curl -fsS --connect-timeout "$curl_connect_timeout" --max-time "$curl_max_time" --retry 2 --retry-delay 2 "$public_base_url/health/ready" >/dev/null
fi

log "deploy success target=${deploy_target} image=${image_ref}"
trap - EXIT

if [ -n "$previous_app_image" ] && [ "$previous_app_image" != "$image_ref" ]; then
  log "previous image for rollback: ${previous_app_image}"
fi
