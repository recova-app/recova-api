#!/usr/bin/env sh
set -eu

command_name="${1:-}"
steps="${2:-1}"
migration_version="${2:-}"
migrate_bin="${MIGRATE_BIN:-migrate}"
migrations_path="${MIGRATIONS_PATH:-migrations}"
database_url="${DATABASE_URL:-}"
docker_bin="${DOCKER_BIN:-docker}"
migrate_image="${MIGRATE_IMAGE:-migrate/migrate:v4.19.0}"

usage() {
  echo "usage: scripts/migrate.sh <up|down|status|check|force> [arg]" >&2
}

if [ -z "$command_name" ]; then
  usage
  exit 1
fi

if [ -z "$database_url" ]; then
  echo "DATABASE_URL must be provided" >&2
  exit 1
fi

if [ ! -d "$migrations_path" ]; then
  echo "migrations path not found: $migrations_path" >&2
  exit 1
fi

migrate_database_url="$database_url"
case "$migrate_database_url" in
  postgresql://*)
    migrate_database_url="postgres://${migrate_database_url#postgresql://}"
    ;;
esac

runner_mode="binary"
migrations_arg="$migrations_path"
if ! command -v "$migrate_bin" >/dev/null 2>&1; then
  if [ "$migrate_bin" != "migrate" ]; then
    echo "migrate binary not found: $migrate_bin" >&2
    exit 1
  fi
  if ! command -v "$docker_bin" >/dev/null 2>&1; then
    echo "migrate binary not found and docker unavailable" >&2
    exit 1
  fi
  migrations_abs="$(cd "$migrations_path" && pwd)"
  runner_mode="docker"
  migrations_arg="/migrations"
fi

run_migrate() {
  if [ "$runner_mode" = "binary" ]; then
    "$migrate_bin" -path "$migrations_arg" -database "$migrate_database_url" "$@"
    return
  fi

  "$docker_bin" run --rm --network host \
    -v "$migrations_abs:/migrations:ro" \
    "$migrate_image" \
    -path "$migrations_arg" -database "$migrate_database_url" "$@"
}

case "$command_name" in
  up)
    run_migrate up
    ;;
  down)
    run_migrate down "$steps"
    ;;
  status)
    run_migrate version
    ;;
  check)
    set +e
    output="$(run_migrate version 2>&1)"
    status_code=$?
    set -e
    if [ "$status_code" -ne 0 ]; then
      printf '%s\n' "$output" >&2
      exit "$status_code"
    fi

    printf '%s\n' "$output"
    if printf '%s' "$output" | grep -i "dirty" >/dev/null 2>&1; then
      echo "migration status dirty; fix before continuing" >&2
      exit 1
    fi
    ;;
  force)
    if [ -z "$migration_version" ]; then
      echo "force version must be provided" >&2
      exit 1
    fi
    run_migrate force "$migration_version"
    ;;
  *)
    usage
    exit 1
    ;;
esac
