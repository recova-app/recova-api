#!/usr/bin/env sh
set -eu

seed_file="${SEED_FILE:-migrations/seeds/000001_baseline_seed.sql}"
database_url="${DATABASE_URL:-}"
psql_bin="${PSQL_BIN:-psql}"

if [ -z "$database_url" ]; then
  echo "DATABASE_URL must be provided" >&2
  exit 1
fi

if [ ! -f "$seed_file" ]; then
  echo "seed file not found: $seed_file" >&2
  exit 1
fi

if ! command -v "$psql_bin" >/dev/null 2>&1; then
  echo "psql binary not found: $psql_bin" >&2
  exit 1
fi

exec "$psql_bin" "$database_url" -v ON_ERROR_STOP=1 -f "$seed_file"
