#!/usr/bin/env sh
set -eu

command_name="${1:-}"
steps="${2:-1}"
migration_version="${2:-}"
migrate_bin="${MIGRATE_BIN:-migrate}"
migrations_path="${MIGRATIONS_PATH:-migrations}"
database_url="${DATABASE_URL:-}"

usage() {
  echo "usage: scripts/migrate.sh <up|down|status|check|force> [arg]" >&2
}

if [ -z "$command_name" ]; then
  usage
  exit 1
fi

if [ -z "$database_url" ]; then
  echo "DATABASE_URL wajib diisi" >&2
  exit 1
fi

if [ ! -d "$migrations_path" ]; then
  echo "migrations path tidak ditemukan: $migrations_path" >&2
  exit 1
fi

if ! command -v "$migrate_bin" >/dev/null 2>&1; then
  echo "migrate binary tidak ditemukan: $migrate_bin" >&2
  exit 1
fi

case "$command_name" in
  up)
    exec "$migrate_bin" -path "$migrations_path" -database "$database_url" up
    ;;
  down)
    exec "$migrate_bin" -path "$migrations_path" -database "$database_url" down "$steps"
    ;;
  status)
    exec "$migrate_bin" -path "$migrations_path" -database "$database_url" version
    ;;
  check)
    set +e
    output="$("$migrate_bin" -path "$migrations_path" -database "$database_url" version 2>&1)"
    status_code=$?
    set -e
    if [ "$status_code" -ne 0 ]; then
      printf '%s\n' "$output" >&2
      exit "$status_code"
    fi

    printf '%s\n' "$output"
    if printf '%s' "$output" | grep -i "dirty" >/dev/null 2>&1; then
      echo "status migrasi dirty; perbaiki sebelum lanjut" >&2
      exit 1
    fi
    ;;
  force)
    if [ -z "$migration_version" ]; then
      echo "versi force wajib diisi" >&2
      exit 1
    fi
    exec "$migrate_bin" -path "$migrations_path" -database "$database_url" force "$migration_version"
    ;;
  *)
    usage
    exit 1
    ;;
esac
