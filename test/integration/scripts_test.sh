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

# with-env.sh must fail when command missing.
assert_fail ./scripts/with-env.sh

# openapi.sh must fail for unsupported command.
assert_fail ./scripts/openapi.sh unknown-command

# compose-smoke.sh must fail for missing compose file.
assert_fail env COMPOSE_FILE="$temp_dir/not-found.yml" ./scripts/compose-smoke.sh

# openapi generate/check must pass on repository baseline.
./scripts/openapi.sh generate >/dev/null
./scripts/openapi.sh check >/dev/null

# security-scan.sh must call provided govulncheck binary with target argument.
fake_vuln_log="$temp_dir/fake-govulncheck.log"
cat > "$temp_dir/govulncheck" <<'SCRIPT'
#!/usr/bin/env sh
printf '%s\n' "$*" >> "$FAKE_VULN_LOG"
SCRIPT
chmod +x "$temp_dir/govulncheck"

GOVULNCHECK_BIN="$temp_dir/govulncheck" \
FAKE_VULN_LOG="$fake_vuln_log" \
./scripts/security-scan.sh ./internal/... >/dev/null

assert_file_contains "$fake_vuln_log" "./internal/..."

# compose-smoke.sh must call docker compose up/ps/down with expected flags.
fake_docker_log="$temp_dir/fake-docker.log"
cat > "$temp_dir/docker" <<'SCRIPT'
#!/usr/bin/env sh
printf '%s\n' "$*" >> "$FAKE_DOCKER_LOG"
exit 0
SCRIPT
chmod +x "$temp_dir/docker"

DOCKER_BIN="$temp_dir/docker" \
FAKE_DOCKER_LOG="$fake_docker_log" \
ENV_FILE=".env.example" \
COMPOSE_PROJECT_NAME="recova-script-test" \
./scripts/compose-smoke.sh >/dev/null

assert_file_contains "$fake_docker_log" "compose --env-file .env.example -f docker-compose.local.yml -p recova-script-test up --build --wait --wait-timeout 120"
assert_file_contains "$fake_docker_log" "compose --env-file .env.example -f docker-compose.local.yml -p recova-script-test ps"
assert_file_contains "$fake_docker_log" "compose --env-file .env.example -f docker-compose.local.yml -p recova-script-test down -v"

# with-env.sh must load env file before running command.
env_file="$temp_dir/local.env"
cat > "$env_file" <<'ENVFILE'
INTEGRATION_ENV_CHECK=from-dotenv
ENVFILE

loaded_value="$(ENV_FILE="$env_file" ./scripts/with-env.sh sh -c 'printf "%s" "${INTEGRATION_ENV_CHECK:-}"')"
if [ "$loaded_value" != "from-dotenv" ]; then
  echo "expected with-env.sh to load env file value, got: $loaded_value" >&2
  exit 1
fi

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
