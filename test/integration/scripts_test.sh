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

assert_file_not_contains() {
  file="$1"
  pattern="$2"
  if grep -F -- "$pattern" "$file" >/dev/null 2>&1; then
    echo "did not expect '$pattern' in $file" >&2
    echo "actual content:" >&2
    cat "$file" >&2
    exit 1
  fi
}

temp_dir="$(mktemp -d)"
trap 'rm -rf "$temp_dir"' EXIT

# with-env.sh must fail when command missing.
assert_fail ./scripts/with-env.sh

# e2e/performance scripts must fail without integration DB url.
assert_fail ./scripts/e2e-critical.sh
assert_fail ./scripts/performance-smoke.sh
assert_fail env RECOVA_DB_INTEGRATION_URL='postgresql://postgres:postgres@localhost:5432/recova_ci_test?sslmode=disable' RECOVA_E2E_SCOPE='invalid-scope' ./scripts/e2e-critical.sh

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

# staging-deploy.sh must orchestrate compose + migrate + seed + readiness checks.
staging_compose_file="$temp_dir/staging-compose.yml"
cat > "$staging_compose_file" <<'YAML'
services:
  db:
    image: postgres:17-alpine
  api:
    image: recova-backend-v2:test
YAML

staging_env_file="$temp_dir/staging.env"
cat > "$staging_env_file" <<'ENVFILE'
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=recova_stage
DB_PORT=55432
APP_PORT=33000
ENVFILE

fake_staging_docker_log="$temp_dir/fake-staging-docker.log"
cat > "$temp_dir/docker-staging" <<'SCRIPT'
#!/usr/bin/env sh
printf '%s\n' "$*" >> "$FAKE_STAGING_DOCKER_LOG"

case "$*" in
  *"SELECT COUNT(*) FROM education_contents;"*)
    printf '2\n'
    ;;
  *"SELECT COUNT(*) FROM daily_motivations;"*)
    printf '2\n'
    ;;
  *"SELECT COUNT(*) FROM daily_challenges;"*)
    printf '2\n'
    ;;
  *"daily_motivations GROUP BY content HAVING COUNT(*) > 1"*)
    printf '0\n'
    ;;
  *"daily_challenges GROUP BY content HAVING COUNT(*) > 1"*)
    printf '0\n'
    ;;
esac

exit 0
SCRIPT
chmod +x "$temp_dir/docker-staging"

fake_staging_migrate_log="$temp_dir/fake-staging-migrate.log"
cat > "$temp_dir/migrate-staging" <<'SCRIPT'
#!/usr/bin/env sh
printf '%s\n' "$*" >> "$FAKE_STAGING_MIGRATE_LOG"
if [ "${1:-}" = "-path" ] && [ "${4:-}" = "version" ]; then
  printf '43\n'
fi
SCRIPT
chmod +x "$temp_dir/migrate-staging"

fake_staging_psql_log="$temp_dir/fake-staging-psql.log"
cat > "$temp_dir/psql-staging" <<'SCRIPT'
#!/usr/bin/env sh
printf '%s\n' "$*" >> "$FAKE_STAGING_PSQL_LOG"
SCRIPT
chmod +x "$temp_dir/psql-staging"

fake_staging_curl_log="$temp_dir/fake-staging-curl.log"
cat > "$temp_dir/curl" <<'SCRIPT'
#!/usr/bin/env sh
printf '%s\n' "$*" >> "$FAKE_STAGING_CURL_LOG"
exit 0
SCRIPT
chmod +x "$temp_dir/curl"

PATH="$temp_dir:$PATH" \
DOCKER_BIN="$temp_dir/docker-staging" \
MIGRATE_BIN="$temp_dir/migrate-staging" \
PSQL_BIN="$temp_dir/psql-staging" \
FAKE_STAGING_DOCKER_LOG="$fake_staging_docker_log" \
FAKE_STAGING_MIGRATE_LOG="$fake_staging_migrate_log" \
FAKE_STAGING_PSQL_LOG="$fake_staging_psql_log" \
FAKE_STAGING_CURL_LOG="$fake_staging_curl_log" \
COMPOSE_FILE="$staging_compose_file" \
ENV_FILE="$staging_env_file" \
COMPOSE_PROJECT_NAME="recova-staging-test" \
./scripts/staging-deploy.sh >/dev/null

assert_file_contains "$fake_staging_docker_log" "compose --env-file $staging_env_file -f $staging_compose_file -p recova-staging-test config -q"
assert_file_contains "$fake_staging_docker_log" "compose --env-file $staging_env_file -f $staging_compose_file -p recova-staging-test up -d db --wait --wait-timeout 120"
assert_file_contains "$fake_staging_docker_log" "compose --env-file $staging_env_file -f $staging_compose_file -p recova-staging-test up -d api --build --wait --wait-timeout 180"
assert_file_contains "$fake_staging_docker_log" "compose --env-file $staging_env_file -f $staging_compose_file -p recova-staging-test ps"
assert_file_contains "$fake_staging_docker_log" "compose --env-file $staging_env_file -f $staging_compose_file -p recova-staging-test down -v --remove-orphans"

assert_file_contains "$fake_staging_migrate_log" "-path migrations -database postgresql://postgres:postgres@127.0.0.1:55432/recova_stage?sslmode=disable up"
assert_file_contains "$fake_staging_migrate_log" "-path migrations -database postgresql://postgres:postgres@127.0.0.1:55432/recova_stage?sslmode=disable down 1"
assert_file_contains "$fake_staging_migrate_log" "-path migrations -database postgresql://postgres:postgres@127.0.0.1:55432/recova_stage?sslmode=disable version"

assert_file_contains "$fake_staging_psql_log" "postgresql://postgres:postgres@127.0.0.1:55432/recova_stage?sslmode=disable -v ON_ERROR_STOP=1 -f migrations/seeds/000001_baseline_seed.sql"
assert_file_contains "$fake_staging_curl_log" "http://127.0.0.1:33000/health/live"
assert_file_contains "$fake_staging_curl_log" "http://127.0.0.1:33000/health/ready"

# cutover-wave.sh should orchestrate waves serially and stop on failure.
assert_fail ./scripts/cutover-wave.sh invalid-wave

fake_cutover_go_log="$temp_dir/fake-cutover-go.log"
cat > "$temp_dir/go-cutover" <<'SCRIPT'
#!/usr/bin/env sh
printf '%s|%s|%s\n' "$*" "${RECOVA_E2E_SCOPE:-}" "${RECOVA_E2E_REPORT_PATH:-}" >> "$FAKE_CUTOVER_GO_LOG"
exit 0
SCRIPT
chmod +x "$temp_dir/go-cutover"

fake_cutover_curl_log="$temp_dir/fake-cutover-curl.log"
cat > "$temp_dir/curl-cutover" <<'SCRIPT'
#!/usr/bin/env sh
printf '%s\n' "$*" >> "$FAKE_CUTOVER_CURL_LOG"
exit 0
SCRIPT
chmod +x "$temp_dir/curl-cutover"

fake_cutover_staging_log="$temp_dir/fake-cutover-staging.log"
cat > "$temp_dir/staging-ok.sh" <<'SCRIPT'
#!/usr/bin/env sh
printf 'staging-ok\n' >> "$FAKE_CUTOVER_STAGING_LOG"
exit 0
SCRIPT
chmod +x "$temp_dir/staging-ok.sh"

fake_cutover_e2e_log="$temp_dir/fake-cutover-e2e.log"
cat > "$temp_dir/e2e-ok.sh" <<'SCRIPT'
#!/usr/bin/env sh
printf '%s|%s\n' "${RECOVA_E2E_SCOPE:-}" "${RECOVA_E2E_REPORT_PATH:-}" >> "$FAKE_CUTOVER_E2E_LOG"
report="${RECOVA_E2E_REPORT_PATH:-}"
if [ -n "$report" ]; then
  mkdir -p "$(dirname "$report")"
  printf '{"status":"passed"}\n' > "$report"
fi
exit 0
SCRIPT
chmod +x "$temp_dir/e2e-ok.sh"

cutover_artifact_dir="$temp_dir/cutover-artifacts"
RECOVA_DB_INTEGRATION_URL='postgresql://postgres:postgres@localhost:5432/recova_ci_test?sslmode=disable' \
GO_BIN="$temp_dir/go-cutover" \
CURL_BIN="$temp_dir/curl-cutover" \
STAGING_DEPLOY_SCRIPT="$temp_dir/staging-ok.sh" \
E2E_SCRIPT="$temp_dir/e2e-ok.sh" \
FAKE_CUTOVER_GO_LOG="$fake_cutover_go_log" \
FAKE_CUTOVER_CURL_LOG="$fake_cutover_curl_log" \
FAKE_CUTOVER_STAGING_LOG="$fake_cutover_staging_log" \
FAKE_CUTOVER_E2E_LOG="$fake_cutover_e2e_log" \
CUTOVER_ARTIFACT_DIR="$cutover_artifact_dir" \
CUTOVER_EXECUTION_ID="exec-success" \
./scripts/cutover-wave.sh all >/dev/null

assert_file_contains "$fake_cutover_staging_log" "staging-ok"
assert_file_contains "$fake_cutover_curl_log" "http://127.0.0.1:3000/health/live"
assert_file_contains "$fake_cutover_curl_log" "http://127.0.0.1:3000/health/ready"
assert_file_contains "$fake_cutover_go_log" "test -count=1 ./test/contract -run ^TestContract_HealthResponses_ValidAgainstOpenAPI$||"
assert_file_contains "$fake_cutover_go_log" "test -count=1 ./test/contract -run ^TestContract_AuthRouteParity_ValidAgainstOpenAPI$||"
assert_file_contains "$fake_cutover_e2e_log" "wave65|$cutover_artifact_dir/exec-success-wave-65-e2e.json"
assert_file_contains "$fake_cutover_e2e_log" "wave66|$cutover_artifact_dir/exec-success-wave-66-e2e.json"
assert_file_contains "$fake_cutover_e2e_log" "wave67|$cutover_artifact_dir/exec-success-wave-67-e2e.json"
assert_file_contains "$fake_cutover_e2e_log" "wave68|$cutover_artifact_dir/exec-success-wave-68-e2e.json"
assert_file_contains "$cutover_artifact_dir/exec-success-summary.log" "all requested waves passed"

fake_cutover_fail_e2e_log="$temp_dir/fake-cutover-fail-e2e.log"
cat > "$temp_dir/e2e-fail.sh" <<'SCRIPT'
#!/usr/bin/env sh
printf '%s|%s\n' "${RECOVA_E2E_SCOPE:-}" "${RECOVA_E2E_REPORT_PATH:-}" >> "$FAKE_CUTOVER_FAIL_E2E_LOG"
if [ "${RECOVA_E2E_SCOPE:-}" = "wave66" ]; then
  exit 1
fi
exit 0
SCRIPT
chmod +x "$temp_dir/e2e-fail.sh"

fake_cutover_rollback_log="$temp_dir/fake-cutover-rollback.log"
cat > "$temp_dir/rollback.sh" <<'SCRIPT'
#!/usr/bin/env sh
printf 'rollback-called\n' >> "$FAKE_CUTOVER_ROLLBACK_LOG"
exit 0
SCRIPT
chmod +x "$temp_dir/rollback.sh"

assert_fail env \
  RECOVA_DB_INTEGRATION_URL='postgresql://postgres:postgres@localhost:5432/recova_ci_test?sslmode=disable' \
  GO_BIN="$temp_dir/go-cutover" \
  CURL_BIN="$temp_dir/curl-cutover" \
  STAGING_DEPLOY_SCRIPT="$temp_dir/staging-ok.sh" \
  E2E_SCRIPT="$temp_dir/e2e-fail.sh" \
  FAKE_CUTOVER_GO_LOG="$fake_cutover_go_log" \
  FAKE_CUTOVER_CURL_LOG="$fake_cutover_curl_log" \
  FAKE_CUTOVER_STAGING_LOG="$fake_cutover_staging_log" \
  FAKE_CUTOVER_FAIL_E2E_LOG="$fake_cutover_fail_e2e_log" \
  FAKE_CUTOVER_ROLLBACK_LOG="$fake_cutover_rollback_log" \
  RUN_ROLLBACK_ON_FAILURE="true" \
  CUTOVER_ROLLBACK_COMMAND="$temp_dir/rollback.sh" \
  CUTOVER_ARTIFACT_DIR="$cutover_artifact_dir" \
  CUTOVER_EXECUTION_ID="exec-failed" \
  ./scripts/cutover-wave.sh all

assert_file_contains "$fake_cutover_fail_e2e_log" "wave65|$cutover_artifact_dir/exec-failed-wave-65-e2e.json"
assert_file_contains "$fake_cutover_fail_e2e_log" "wave66|$cutover_artifact_dir/exec-failed-wave-66-e2e.json"
assert_file_not_contains "$fake_cutover_fail_e2e_log" "wave67|$cutover_artifact_dir/exec-failed-wave-67-e2e.json"
assert_file_contains "$fake_cutover_rollback_log" "rollback-called"

# e2e-critical.sh and performance-smoke.sh should invoke go test with report env.
fake_go_log="$temp_dir/fake-go.log"
cat > "$temp_dir/go" <<'SCRIPT'
#!/usr/bin/env sh
printf '%s|%s|%s\n' "$*" "${RECOVA_E2E_REPORT_PATH:-}" "${RECOVA_PERF_REPORT_PATH:-}" >> "$FAKE_GO_LOG"
exit 0
SCRIPT
chmod +x "$temp_dir/go"

PATH="$temp_dir:$PATH" \
FAKE_GO_LOG="$fake_go_log" \
RECOVA_DB_INTEGRATION_URL="postgresql://postgres:postgres@localhost:5432/recova_ci_test?sslmode=disable" \
RECOVA_E2E_REPORT_PATH="$temp_dir/e2e-report.json" \
./scripts/e2e-critical.sh >/dev/null

assert_file_contains "$fake_go_log" "test -count=1 ./test/e2e -run TestE2E_CriticalFlows|$temp_dir/e2e-report.json|"

PATH="$temp_dir:$PATH" \
FAKE_GO_LOG="$fake_go_log" \
RECOVA_DB_INTEGRATION_URL="postgresql://postgres:postgres@localhost:5432/recova_ci_test?sslmode=disable" \
RECOVA_PERF_REPORT_PATH="$temp_dir/perf-report.json" \
./scripts/performance-smoke.sh >/dev/null

assert_file_contains "$fake_go_log" "test -count=1 ./test/performance -run TestPerformance_LoadSmoke||$temp_dir/perf-report.json"

# preflight should pass on current baseline repository.
./scripts/preflight.sh >/dev/null

echo "scripts integration checks passed"
