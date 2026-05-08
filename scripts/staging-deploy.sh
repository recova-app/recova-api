#!/usr/bin/env sh
set -eu

compose_file="${COMPOSE_FILE:-docker-compose.local.yml}"
env_file="${ENV_FILE:-.env.example}"
project_name="${COMPOSE_PROJECT_NAME:-recova-staging}"
docker_bin="${DOCKER_BIN:-docker}"
migrate_bin="${MIGRATE_BIN:-migrate}"
keep_stack="${KEEP_STACK:-false}"
run_migration_dry_run="${RUN_MIGRATION_DRY_RUN:-true}"

if ! command -v "$docker_bin" >/dev/null 2>&1; then
  echo "[staging-deploy] docker command tidak ditemukan: $docker_bin" >&2
  exit 1
fi

if ! command -v "$migrate_bin" >/dev/null 2>&1; then
  echo "[staging-deploy] migrate binary tidak ditemukan: $migrate_bin" >&2
  exit 1
fi

if ! command -v curl >/dev/null 2>&1; then
  echo "[staging-deploy] curl command tidak ditemukan" >&2
  exit 1
fi

if [ ! -f "$compose_file" ]; then
  echo "[staging-deploy] compose file tidak ditemukan: $compose_file" >&2
  exit 1
fi

if [ ! -f "$env_file" ]; then
  echo "[staging-deploy] env file tidak ditemukan: $env_file" >&2
  exit 1
fi

set -a
# shellcheck source=/dev/null
. "$env_file"
set +a

postgres_user="${POSTGRES_USER:-postgres}"
postgres_password="${POSTGRES_PASSWORD:-postgres}"
postgres_db="${POSTGRES_DB:-recova_db}"
db_port="${DB_PORT:-5432}"
app_port="${APP_PORT:-3000}"

api_base_url="${API_BASE_URL:-http://127.0.0.1:${app_port}}"
database_url="postgresql://${postgres_user}:${postgres_password}@127.0.0.1:${db_port}/${postgres_db}?sslmode=disable"

compose() {
  "$docker_bin" compose --env-file "$env_file" -f "$compose_file" -p "$project_name" "$@"
}

query_db_count() {
  query="$1"
  compose exec -T db psql -U "$postgres_user" -d "$postgres_db" -At -v ON_ERROR_STOP=1 -c "$query"
}

cleanup() {
  if [ "$keep_stack" = "true" ]; then
    echo "[staging-deploy] KEEP_STACK=true, stack tetap aktif"
    return
  fi
  compose down -v --remove-orphans >/dev/null 2>&1 || true
}
trap cleanup EXIT INT TERM

echo "[staging-deploy] validate compose config"
compose config -q

echo "[staging-deploy] reset stack"
compose down -v --remove-orphans >/dev/null 2>&1 || true

echo "[staging-deploy] start database"
compose up -d db --wait --wait-timeout 120

export DATABASE_URL="$database_url"
export DIRECT_DATABASE_URL="$database_url"
export MIGRATE_BIN="$migrate_bin"

echo "[staging-deploy] apply migrations"
./scripts/migrate.sh up
./scripts/migrate.sh check

if [ "$run_migration_dry_run" = "true" ]; then
  echo "[staging-deploy] migration dry-run (down 1 -> up)"
  ./scripts/migrate.sh down 1
  ./scripts/migrate.sh up
  ./scripts/migrate.sh check
fi

echo "[staging-deploy] run seed pass #1"
./scripts/seed.sh

education_first="$(query_db_count 'SELECT COUNT(*) FROM education_contents;')"
motivation_first="$(query_db_count 'SELECT COUNT(*) FROM daily_motivations;')"
challenge_first="$(query_db_count 'SELECT COUNT(*) FROM daily_challenges;')"

echo "[staging-deploy] run seed pass #2"
./scripts/seed.sh

education_second="$(query_db_count 'SELECT COUNT(*) FROM education_contents;')"
motivation_second="$(query_db_count 'SELECT COUNT(*) FROM daily_motivations;')"
challenge_second="$(query_db_count 'SELECT COUNT(*) FROM daily_challenges;')"

if [ "$education_first" != "$education_second" ] || [ "$motivation_first" != "$motivation_second" ] || [ "$challenge_first" != "$challenge_second" ]; then
  echo "[staging-deploy] seed idempotency gagal: before=($education_first,$motivation_first,$challenge_first) after=($education_second,$motivation_second,$challenge_second)" >&2
  exit 1
fi

if [ "$education_second" -le 0 ] || [ "$motivation_second" -le 0 ] || [ "$challenge_second" -le 0 ]; then
  echo "[staging-deploy] integrity gagal: reference content kosong" >&2
  exit 1
fi

motivation_duplicates="$(query_db_count 'SELECT COUNT(*) FROM (SELECT content FROM daily_motivations GROUP BY content HAVING COUNT(*) > 1) dup;')"
challenge_duplicates="$(query_db_count 'SELECT COUNT(*) FROM (SELECT content FROM daily_challenges GROUP BY content HAVING COUNT(*) > 1) dup;')"

if [ "$motivation_duplicates" -ne 0 ] || [ "$challenge_duplicates" -ne 0 ]; then
  echo "[staging-deploy] integrity gagal: duplicate reference content terdeteksi" >&2
  exit 1
fi

echo "[staging-deploy] start api with readiness gate"
compose up -d api --build --wait --wait-timeout 180

curl -fsS --retry 6 --retry-delay 2 --retry-connrefused "${api_base_url}/health/live" >/dev/null
curl -fsS --retry 6 --retry-delay 2 --retry-connrefused "${api_base_url}/health/ready" >/dev/null

echo "[staging-deploy] post-deploy smoke checks"
compose ps

echo "[staging-deploy] success"
