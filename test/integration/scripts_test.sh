#!/usr/bin/env sh
set -eu

assert_fail() {
  set +e
  "$@" >/dev/null 2>&1
  code=$?
  set -e
  if [ "$code" -eq 0 ]; then
    echo "expected command to fail: $*" >&2
    exit 1
  fi
}

assert_file_contains() {
  file="$1"
  pattern="$2"
  if ! grep -F -- "$pattern" "$file" >/dev/null 2>&1; then
    echo "expected '$pattern' in $file" >&2
    echo "actual content:" >&2
    cat "$file" >&2
    exit 1
  fi
}

temp_dir="$(mktemp -d)"
trap 'rm -rf "$temp_dir"' EXIT

# migrate.sh must fail if command argument missing.
assert_fail ./scripts/migrate.sh

# Build fake migrate binary to verify argument wiring.
fake_log="$temp_dir/fake-migrate.log"
cat > "$temp_dir/migrate" <<'SCRIPT'
#!/usr/bin/env sh
printf '%s\n' "$*" >> "$FAKE_MIGRATE_LOG"
if [ -n "${FAKE_MIGRATE_STDOUT:-}" ]; then
  printf '%s\n' "$FAKE_MIGRATE_STDOUT"
fi
if [ -n "${FAKE_MIGRATE_EXIT_CODE:-}" ]; then
  exit "$FAKE_MIGRATE_EXIT_CODE"
fi
SCRIPT
chmod +x "$temp_dir/migrate"

DATABASE_URL='postgres://user:pass@localhost:5432/recova?sslmode=disable' \
MIGRATE_BIN="$temp_dir/migrate" \
FAKE_MIGRATE_LOG="$fake_log" \
./scripts/migrate.sh up

assert_file_contains "$fake_log" "-path migrations -database postgres://user:pass@localhost:5432/recova?sslmode=disable up"

DATABASE_URL='postgres://user:pass@localhost:5432/recova?sslmode=disable' \
MIGRATE_BIN="$temp_dir/migrate" \
FAKE_MIGRATE_LOG="$fake_log" \
./scripts/migrate.sh down 1

assert_file_contains "$fake_log" "-path migrations -database postgres://user:pass@localhost:5432/recova?sslmode=disable down 1"

DATABASE_URL='postgres://user:pass@localhost:5432/recova?sslmode=disable' \
MIGRATE_BIN="$temp_dir/migrate" \
FAKE_MIGRATE_LOG="$fake_log" \
./scripts/migrate.sh status

assert_file_contains "$fake_log" "-path migrations -database postgres://user:pass@localhost:5432/recova?sslmode=disable version"

DATABASE_URL='postgres://user:pass@localhost:5432/recova?sslmode=disable' \
MIGRATE_BIN="$temp_dir/migrate" \
FAKE_MIGRATE_LOG="$fake_log" \
FAKE_MIGRATE_STDOUT='43' \
./scripts/migrate.sh check >/dev/null

assert_fail env \
  DATABASE_URL='postgres://user:pass@localhost:5432/recova?sslmode=disable' \
  MIGRATE_BIN="$temp_dir/migrate" \
  FAKE_MIGRATE_LOG="$fake_log" \
  FAKE_MIGRATE_STDOUT='43 (dirty)' \
  ./scripts/migrate.sh check

# Build fake psql binary to verify seed wiring.
fake_psql_log="$temp_dir/fake-psql.log"
cat > "$temp_dir/psql" <<'SCRIPT'
#!/usr/bin/env sh
printf '%s\n' "$*" >> "$FAKE_PSQL_LOG"
SCRIPT
chmod +x "$temp_dir/psql"

seed_file="$temp_dir/seed.sql"
cat > "$seed_file" <<'SQL'
SELECT 1;
SQL

DATABASE_URL='postgres://user:pass@localhost:5432/recova?sslmode=disable' \
PSQL_BIN="$temp_dir/psql" \
FAKE_PSQL_LOG="$fake_psql_log" \
SEED_FILE="$seed_file" \
./scripts/seed.sh

assert_file_contains "$fake_psql_log" "postgres://user:pass@localhost:5432/recova?sslmode=disable -v ON_ERROR_STOP=1 -f $seed_file"

# preflight should pass on current baseline repository.
./scripts/preflight.sh >/dev/null

echo "scripts integration checks passed"
