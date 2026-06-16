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
  echo "[staging-deploy] docker command not found: $docker_bin" >&2
  exit 1
fi

if ! command -v "$migrate_bin" >/dev/null 2>&1; then
  echo "[staging-deploy] migrate binary not found: $migrate_bin" >&2
  exit 1
fi

if ! command -v curl >/dev/null 2>&1; then
  echo "[staging-deploy] curl command not found" >&2
  exit 1
fi

if [ ! -f "$compose_file" ]; then
  echo "[staging-deploy] compose file not found: $compose_file" >&2
  exit 1
fi

if [ ! -f "$env_file" ]; then
  echo "[staging-deploy] env file not found: $env_file" >&2
  exit 1
fi

case "$env_file" in
  /*|*/*) env_source="$env_file" ;;
  *) env_source="./$env_file" ;;
esac

set -a
# shellcheck source=/dev/null
. "$env_source"
set +a

postgres_user="${POSTGRES_USER:-postgres}"
postgres_password="${POSTGRES_PASSWORD:-postgres}"
postgres_db="${POSTGRES_DB:-recova_db}"
db_port="${DB_PORT:-5432}"
app_port="${APP_PORT:-3001}"

api_base_url="${API_BASE_URL:-http://127.0.0.1:${app_port}}"
database_url="postgresql://${postgres_user}:${postgres_password}@127.0.0.1:${db_port}/${postgres_db}?sslmode=disable"

compose() {
  "$docker_bin" compose --env-file "$env_file" -f "$compose_file" -p "$project_name" "$@"
}

query_db_count() {
  query="$1"
  compose exec -T db psql -U "$postgres_user" -d "$postgres_db" -At -v ON_ERROR_STOP=1 -c "$query"
}

seed_tables="
users
profiles
streaks
check_ins
journals
community_posts
community_comments
community_post_likes
education_contents
daily_motivations
daily_challenges
achievements
user_achievement_progress
user_ai_persona_preferences
ai_chats
"

seed_min_count() {
  table="$1"
  case "$table" in
    users) echo 6 ;;
    profiles) echo 6 ;;
    streaks) echo 11 ;;
    check_ins) echo 84 ;;
    journals) echo 84 ;;
    community_posts) echo 12 ;;
    community_comments) echo 25 ;;
    community_post_likes) echo 20 ;;
    education_contents) echo 23 ;;
    daily_motivations) echo 35 ;;
    daily_challenges) echo 35 ;;
    achievements) echo 15 ;;
    user_achievement_progress) echo 24 ;;
    user_ai_persona_preferences) echo 6 ;;
    ai_chats) echo 18 ;;
    *)
      echo "unknown seed table: $table" >&2
      exit 1
      ;;
  esac
}

capture_seed_counts() {
  prefix="$1"
  for table in $seed_tables; do
    count="$(query_db_count "SELECT COUNT(*) FROM $table;")"
    eval "${prefix}_${table}=${count}"
  done
}

cleanup() {
  if [ "$keep_stack" = "true" ]; then
    echo "[staging-deploy] KEEP_STACK=true, stack remains running"
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

capture_seed_counts first

echo "[staging-deploy] run seed pass #2"
./scripts/seed.sh

capture_seed_counts second

for table in $seed_tables; do
  eval "first_count=\${first_${table}}"
  eval "second_count=\${second_${table}}"
  if [ "$first_count" != "$second_count" ]; then
    echo "[staging-deploy] seed idempotency failed on $table: before=$first_count after=$second_count" >&2
    exit 1
  fi
  min_count="$(seed_min_count "$table")"
  if [ "$second_count" -lt "$min_count" ]; then
    echo "[staging-deploy] integrity failed on $table: got=$second_count min=$min_count" >&2
    exit 1
  fi
done

motivation_duplicates="$(query_db_count 'SELECT COUNT(*) FROM (SELECT content FROM daily_motivations GROUP BY content HAVING COUNT(*) > 1) dup;')"
challenge_duplicates="$(query_db_count 'SELECT COUNT(*) FROM (SELECT content FROM daily_challenges GROUP BY content HAVING COUNT(*) > 1) dup;')"
achievement_code_duplicates="$(query_db_count 'SELECT COUNT(*) FROM (SELECT code FROM achievements GROUP BY code HAVING COUNT(*) > 1) dup;')"
users_google_duplicates="$(query_db_count 'SELECT COUNT(*) FROM (SELECT google_id FROM users WHERE google_id IS NOT NULL GROUP BY google_id HAVING COUNT(*) > 1) dup;')"
users_email_duplicates="$(query_db_count 'SELECT COUNT(*) FROM (SELECT email FROM users GROUP BY email HAVING COUNT(*) > 1) dup;')"
profiles_user_duplicates="$(query_db_count 'SELECT COUNT(*) FROM (SELECT user_id FROM profiles GROUP BY user_id HAVING COUNT(*) > 1) dup;')"
checkins_unique_conflicts="$(query_db_count 'SELECT COUNT(*) FROM (SELECT user_id, check_in_date FROM check_ins GROUP BY user_id, check_in_date HAVING COUNT(*) > 1) dup;')"
journals_checkin_duplicates="$(query_db_count 'SELECT COUNT(*) FROM (SELECT check_in_id FROM journals WHERE check_in_id IS NOT NULL GROUP BY check_in_id HAVING COUNT(*) > 1) dup;')"
achievement_progress_duplicates="$(query_db_count 'SELECT COUNT(*) FROM (SELECT user_id, achievement_id FROM user_achievement_progress GROUP BY user_id, achievement_id HAVING COUNT(*) > 1) dup;')"

if [ "$motivation_duplicates" -ne 0 ] || [ "$challenge_duplicates" -ne 0 ] || [ "$achievement_code_duplicates" -ne 0 ] || [ "$users_google_duplicates" -ne 0 ] || [ "$users_email_duplicates" -ne 0 ] || [ "$profiles_user_duplicates" -ne 0 ] || [ "$checkins_unique_conflicts" -ne 0 ] || [ "$journals_checkin_duplicates" -ne 0 ] || [ "$achievement_progress_duplicates" -ne 0 ]; then
  echo "[staging-deploy] integrity failed: duplicate reference content detected" >&2
  exit 1
fi

echo "[staging-deploy] start api with readiness gate"
compose up -d api --build --wait --wait-timeout 180

curl -fsS --retry 6 --retry-delay 2 --retry-connrefused "${api_base_url}/health/live" >/dev/null
curl -fsS --retry 6 --retry-delay 2 --retry-connrefused "${api_base_url}/health/ready" >/dev/null

echo "[staging-deploy] post-deploy smoke checks"
compose ps

echo "[staging-deploy] success"
