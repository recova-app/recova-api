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
assert_fail ./scripts/scalar.sh unknown-command
assert_fail ./scripts/module-consistency.sh unknown-command

# compose-smoke.sh must fail for missing compose file.
assert_fail env COMPOSE_FILE="$temp_dir/not-found.yml" ./scripts/compose-smoke.sh

# openapi generate/check must pass on repository baseline.
./scripts/openapi.sh generate >/dev/null
./scripts/openapi.sh check >/dev/null
./scripts/scalar.sh check >/dev/null
./scripts/module-consistency.sh check >/dev/null
./scripts/module-consistency.sh structure >/dev/null
./scripts/module-consistency.sh contract >/dev/null
./scripts/module-consistency.sh openapi >/dev/null
./scripts/module-consistency.sh full-check >/dev/null

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
QUOTED_ENV_CHECK="from-quoted-dotenv"
SINGLE_QUOTED_ENV_CHECK='from-single-quoted-dotenv'
ENVFILE

loaded_value="$(ENV_FILE="$env_file" ./scripts/with-env.sh sh -c 'printf "%s" "${INTEGRATION_ENV_CHECK:-}"')"
if [ "$loaded_value" != "from-dotenv" ]; then
  echo "expected with-env.sh to load env file value, got: $loaded_value" >&2
  exit 1
fi
quoted_loaded_value="$(ENV_FILE="$env_file" ./scripts/with-env.sh sh -c 'printf "%s|%s" "${QUOTED_ENV_CHECK:-}" "${SINGLE_QUOTED_ENV_CHECK:-}"')"
if [ "$quoted_loaded_value" != "from-quoted-dotenv|from-single-quoted-dotenv" ]; then
  echo "expected with-env.sh to strip surrounding env quotes, got: $quoted_loaded_value" >&2
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

DATABASE_URL='postgresql://user:pass@localhost:5432/recova?sslmode=disable' \
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
    printf '8\n'
    ;;
  *"SELECT COUNT(*) FROM daily_motivations;"*)
    printf '10\n'
    ;;
  *"SELECT COUNT(*) FROM daily_challenges;"*)
    printf '10\n'
    ;;
  *"SELECT COUNT(*) FROM achievements;"*)
    printf '10\n'
    ;;
  *"daily_motivations GROUP BY content HAVING COUNT(*) > 1"*)
    printf '0\n'
    ;;
  *"daily_challenges GROUP BY content HAVING COUNT(*) > 1"*)
    printf '0\n'
    ;;
  *"achievements GROUP BY code HAVING COUNT(*) > 1"*)
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
assert_file_contains "$fake_staging_docker_log" "SELECT COUNT(*) FROM achievements;"
assert_file_contains "$fake_staging_docker_log" "daily_motivations GROUP BY content HAVING COUNT(*) > 1"
assert_file_contains "$fake_staging_docker_log" "achievements GROUP BY code HAVING COUNT(*) > 1"

assert_file_contains "$fake_staging_migrate_log" "-path migrations -database postgres://postgres:postgres@127.0.0.1:55432/recova_stage?sslmode=disable up"
assert_file_contains "$fake_staging_migrate_log" "-path migrations -database postgres://postgres:postgres@127.0.0.1:55432/recova_stage?sslmode=disable down 1"
assert_file_contains "$fake_staging_migrate_log" "-path migrations -database postgres://postgres:postgres@127.0.0.1:55432/recova_stage?sslmode=disable version"

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
assert_file_contains "$fake_cutover_curl_log" "http://127.0.0.1:3001/health/live"
assert_file_contains "$fake_cutover_curl_log" "http://127.0.0.1:3001/health/ready"
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

# stabilization-gate.sh should orchestrate regression + e2e + performance and write evidence report.
assert_fail ./scripts/stabilization-gate.sh

fake_stabilization_go_log="$temp_dir/fake-stabilization-go.log"
cat > "$temp_dir/go-stabilization" <<'SCRIPT'
#!/usr/bin/env sh
printf '%s\n' "$*" >> "$FAKE_STABILIZATION_GO_LOG"
exit 0
SCRIPT
chmod +x "$temp_dir/go-stabilization"

fake_openapi_script="$temp_dir/openapi-ok.sh"
cat > "$fake_openapi_script" <<'SCRIPT'
#!/usr/bin/env sh
printf '%s\n' "$*" >> "$FAKE_STABILIZATION_OPENAPI_LOG"
exit 0
SCRIPT
chmod +x "$fake_openapi_script"

fake_stabilization_e2e_script="$temp_dir/e2e-stabilization-ok.sh"
cat > "$fake_stabilization_e2e_script" <<'SCRIPT'
#!/usr/bin/env sh
printf '%s\n' "${RECOVA_E2E_REPORT_PATH:-}" >> "$FAKE_STABILIZATION_E2E_LOG"
report="${RECOVA_E2E_REPORT_PATH:-}"
mkdir -p "$(dirname "$report")"
printf '{"status":"passed"}\n' > "$report"
exit 0
SCRIPT
chmod +x "$fake_stabilization_e2e_script"

fake_stabilization_perf_script="$temp_dir/perf-stabilization-ok.sh"
cat > "$fake_stabilization_perf_script" <<'SCRIPT'
#!/usr/bin/env sh
printf '%s\n' "${RECOVA_PERF_REPORT_PATH:-}" >> "$FAKE_STABILIZATION_PERF_LOG"
report="${RECOVA_PERF_REPORT_PATH:-}"
mkdir -p "$(dirname "$report")"
printf '{"status":"passed"}\n' > "$report"
exit 0
SCRIPT
chmod +x "$fake_stabilization_perf_script"

stabilization_artifact_dir="$temp_dir/stabilization-artifacts"
RECOVA_DB_INTEGRATION_URL="postgresql://postgres:postgres@localhost:5432/recova_ci_test?sslmode=disable" \
GO_BIN="$temp_dir/go-stabilization" \
OPENAPI_SCRIPT="$fake_openapi_script" \
E2E_SCRIPT="$fake_stabilization_e2e_script" \
PERFORMANCE_SCRIPT="$fake_stabilization_perf_script" \
FAKE_STABILIZATION_GO_LOG="$fake_stabilization_go_log" \
FAKE_STABILIZATION_OPENAPI_LOG="$temp_dir/fake-stabilization-openapi.log" \
FAKE_STABILIZATION_E2E_LOG="$temp_dir/fake-stabilization-e2e.log" \
FAKE_STABILIZATION_PERF_LOG="$temp_dir/fake-stabilization-perf.log" \
STABILIZATION_ARTIFACT_DIR="$stabilization_artifact_dir" \
STABILIZATION_EXECUTION_ID="stabilization-success" \
./scripts/stabilization-gate.sh >/dev/null

assert_file_contains "$fake_stabilization_go_log" "test ./..."
assert_file_contains "$temp_dir/fake-stabilization-openapi.log" "check"
assert_file_contains "$temp_dir/fake-stabilization-e2e.log" "$stabilization_artifact_dir/stabilization-success-e2e-critical-flows.json"
assert_file_contains "$temp_dir/fake-stabilization-perf.log" "$stabilization_artifact_dir/stabilization-success-performance-smoke.json"
assert_file_contains "$stabilization_artifact_dir/stabilization-success-summary.log" "gate pass: performance-smoke"
assert_file_contains "$stabilization_artifact_dir/stabilization-success-stabilization-report.json" "\"status\": \"passed\""

fake_stabilization_e2e_failed_script="$temp_dir/e2e-stabilization-failed.sh"
cat > "$fake_stabilization_e2e_failed_script" <<'SCRIPT'
#!/usr/bin/env sh
report="${RECOVA_E2E_REPORT_PATH:-}"
mkdir -p "$(dirname "$report")"
printf '{"status":"failed"}\n' > "$report"
exit 0
SCRIPT
chmod +x "$fake_stabilization_e2e_failed_script"

assert_fail env \
  RECOVA_DB_INTEGRATION_URL="postgresql://postgres:postgres@localhost:5432/recova_ci_test?sslmode=disable" \
  GO_BIN="$temp_dir/go-stabilization" \
  OPENAPI_SCRIPT="$fake_openapi_script" \
  E2E_SCRIPT="$fake_stabilization_e2e_failed_script" \
  PERFORMANCE_SCRIPT="$fake_stabilization_perf_script" \
  FAKE_STABILIZATION_GO_LOG="$fake_stabilization_go_log" \
  FAKE_STABILIZATION_OPENAPI_LOG="$temp_dir/fake-stabilization-openapi.log" \
  FAKE_STABILIZATION_E2E_LOG="$temp_dir/fake-stabilization-e2e.log" \
  FAKE_STABILIZATION_PERF_LOG="$temp_dir/fake-stabilization-perf.log" \
  STABILIZATION_ARTIFACT_DIR="$stabilization_artifact_dir" \
  STABILIZATION_EXECUTION_ID="stabilization-failed" \
  ./scripts/stabilization-gate.sh

# rollback-rehearsal.sh should force rollback execution path and write evidence report.
assert_fail env \
  RECOVA_DB_INTEGRATION_URL="postgresql://postgres:postgres@localhost:5432/recova_ci_test?sslmode=disable" \
  ./scripts/rollback-rehearsal.sh

rollback_rehearsal_artifact_dir="$temp_dir/rollback-rehearsal-artifacts"
fake_rollback_rehearsal_command_log="$temp_dir/fake-rollback-rehearsal-command.log"
RECOVA_DB_INTEGRATION_URL="postgresql://postgres:postgres@localhost:5432/recova_ci_test?sslmode=disable" \
ROLLBACK_REHEARSAL_COMMAND="printf 'rollback-command-ok\\n' >> \"$fake_rollback_rehearsal_command_log\"" \
ROLLBACK_REHEARSAL_ARTIFACT_DIR="$rollback_rehearsal_artifact_dir" \
ROLLBACK_REHEARSAL_EXECUTION_ID="rollback-rehearsal-success" \
ROLLBACK_REHEARSAL_WAVE="65" \
./scripts/rollback-rehearsal.sh >/dev/null

assert_file_contains "$fake_rollback_rehearsal_command_log" "rollback-command-ok"
assert_file_contains "$rollback_rehearsal_artifact_dir/rollback-rehearsal-success-summary.log" "success"
assert_file_contains "$rollback_rehearsal_artifact_dir/rollback-rehearsal-success-rollback-rehearsal-report.json" "\"status\": \"passed\""

# runtime-decommission.sh should enforce zero legacy traffic and rollback evidence retention checks.
assert_fail ./scripts/runtime-decommission.sh
assert_fail env \
  LEGACY_RUNTIME_TRAFFIC_COUNT="1" \
  ROLLBACK_EVIDENCE_DIR="$rollback_rehearsal_artifact_dir" \
  ./scripts/runtime-decommission.sh

runtime_truth_file="$temp_dir/runtime-source-of-truth.md"
cat > "$runtime_truth_file" <<'DOC'
---
title: Runtime Source
last_reviewed: 2026-05-08
---

Current runtime: Go Fiber
DOC

legacy_archive_dir="$temp_dir/legacy-runtime"
mkdir -p "$legacy_archive_dir"
printf 'legacy-config\n' > "$legacy_archive_dir/runtime.env"

legacy_traffic_evidence="$temp_dir/legacy-traffic.log"
printf 'legacy_public_traffic=0\n' > "$legacy_traffic_evidence"

decommission_artifact_dir="$temp_dir/decommission-artifacts"
LEGACY_RUNTIME_TRAFFIC_COUNT="0" \
LEGACY_RUNTIME_TRAFFIC_EVIDENCE_FILE="$legacy_traffic_evidence" \
LEGACY_ARCHIVE_PATHS="$legacy_archive_dir" \
ROLLBACK_EVIDENCE_DIR="$rollback_rehearsal_artifact_dir" \
ROLLBACK_RETENTION_DAYS="90" \
RUNTIME_SOURCE_OF_TRUTH_FILE="$runtime_truth_file" \
RUNTIME_SOURCE_OF_TRUTH_KEYWORD="Go" \
DECOMMISSION_ARTIFACT_DIR="$decommission_artifact_dir" \
DECOMMISSION_EXECUTION_ID="decommission-success" \
./scripts/runtime-decommission.sh >/dev/null

assert_file_contains "$decommission_artifact_dir/decommission-success-summary.log" "success"
assert_file_contains "$decommission_artifact_dir/decommission-success-decommission-report.json" "\"status\": \"passed\""
assert_file_contains "$decommission_artifact_dir/decommission-success-decommission-report.json" "\"legacyTrafficCount\": 0"
[ -f "$decommission_artifact_dir/decommission-success-express-archive.tar.gz" ] || {
  echo "expected legacy archive artifact to exist" >&2
  exit 1
}

# post-migration-maintenance.sh should require completed review checks and emit backlog with owner/priority.
assert_fail ./scripts/post-migration-maintenance.sh

maintenance_docs_root="$temp_dir/maintenance-docs"
mkdir -p "$maintenance_docs_root/operations" "$maintenance_docs_root/modules"
cat > "$maintenance_docs_root/operations/security.md" <<'DOC'
---
title: Security Ops
last_reviewed: 2020-01-01
---
DOC
cat > "$maintenance_docs_root/modules/routine.md" <<'DOC'
---
title: Routine Module
last_reviewed: 2020-01-01
---
DOC
cat > "$maintenance_docs_root/modules/users.md" <<'DOC'
---
title: Users Module
last_reviewed: 2099-01-01
---
DOC

maintenance_artifact_dir="$temp_dir/maintenance-artifacts"
ALERT_REVIEW_STATUS="done" \
SLO_REVIEW_STATUS="done" \
DEPENDENCY_CADENCE_REVIEW_STATUS="done" \
DOCS_ROOT="$maintenance_docs_root" \
MAINTENANCE_ARTIFACT_DIR="$maintenance_artifact_dir" \
MAINTENANCE_EXECUTION_ID="maintenance-success" \
DEFAULT_BACKLOG_OWNER="ops-owner" \
DEFAULT_BACKLOG_DUE_DATE="2026-06-01" \
./scripts/post-migration-maintenance.sh >/dev/null

assert_file_contains "$maintenance_artifact_dir/maintenance-success-summary.log" "success"
assert_file_contains "$maintenance_artifact_dir/maintenance-success-maintenance-report.json" "\"status\": \"passed\""
assert_file_contains "$maintenance_artifact_dir/maintenance-success-maintenance-report.json" "\"staleDocsTotal\": 2"
assert_file_contains "$maintenance_artifact_dir/maintenance-success-maintenance-backlog.md" "| MNT-001 |"
assert_file_contains "$maintenance_artifact_dir/maintenance-success-maintenance-backlog.md" "| ops-owner | high |"

# preflight should pass on current baseline repository.
./scripts/preflight.sh >/dev/null

echo "scripts integration checks passed"

# remote-deploy.sh should enforce staging APP_ENV, run migrate/check, and run deploy smoke checks.
remote_temp_dir="$(mktemp -d)"
trap 'rm -rf "$temp_dir" "$remote_temp_dir"' EXIT

remote_repo_dir="$remote_temp_dir/repo"
mkdir -p "$remote_repo_dir/scripts" "$remote_repo_dir/scripts/deploy" "$remote_repo_dir/migrations"
cp ./scripts/migrate.sh "$remote_repo_dir/scripts/migrate.sh"
cp ./scripts/deploy/remote-deploy.sh "$remote_repo_dir/scripts/deploy/remote-deploy.sh"
chmod +x "$remote_repo_dir/scripts/migrate.sh" "$remote_repo_dir/scripts/deploy/remote-deploy.sh"

touch "$remote_repo_dir/docker-compose.staging.yml"
cat > "$remote_repo_dir/.env.staging" <<'ENVFILE'
APP_ENV=staging
DATABASE_URL="postgresql://postgres:postgres@127.0.0.1:5432/recova_stage?sslmode=disable"
ENVFILE

fake_remote_log="$remote_temp_dir/fake-remote.log"

cat > "$remote_temp_dir/git" <<'SCRIPT'
#!/usr/bin/env sh
printf 'git %s\n' "$*" >> "$FAKE_REMOTE_LOG"
case "${1:-}" in
  rev-parse)
    if [ "${2:-}" = "--is-inside-work-tree" ]; then
      exit 0
    fi
    ;;
  diff)
    exit 0
    ;;
esac
exit 0
SCRIPT
chmod +x "$remote_temp_dir/git"

cat > "$remote_temp_dir/docker" <<'SCRIPT'
#!/usr/bin/env sh
printf 'docker %s\n' "$*" >> "$FAKE_REMOTE_LOG"
exit 0
SCRIPT
chmod +x "$remote_temp_dir/docker"

cat > "$remote_temp_dir/migrate" <<'SCRIPT'
#!/usr/bin/env sh
printf 'migrate %s\n' "$*" >> "$FAKE_REMOTE_LOG"
if [ "${4:-}" = "version" ]; then
  printf '43\n'
fi
exit 0
SCRIPT
chmod +x "$remote_temp_dir/migrate"

cat > "$remote_temp_dir/curl" <<'SCRIPT'
#!/usr/bin/env sh
printf 'curl %s\n' "$*" >> "$FAKE_REMOTE_LOG"
case "$*" in
  *"%{http_code}"*)
    printf '401'
    ;;
esac
exit 0
SCRIPT
chmod +x "$remote_temp_dir/curl"

PATH="$remote_temp_dir:$PATH" \
FAKE_REMOTE_LOG="$fake_remote_log" \
MIGRATE_BIN="$remote_temp_dir/migrate" \
"$remote_repo_dir/scripts/deploy/remote-deploy.sh" \
  "$remote_repo_dir" \
  "develop" \
  "ghcr.io/example/recova:sha-abc" \
  "docker-compose.staging.yml" \
  ".env.staging" \
  "staging" \
  "staging" \
  "3000" \
  "https://public.example.test" >/dev/null

assert_file_contains "$fake_remote_log" "git fetch origin develop --prune"
assert_file_contains "$fake_remote_log" "git checkout -B develop origin/develop"
assert_file_not_contains "$fake_remote_log" "git reset --hard"
assert_file_contains "$fake_remote_log" "docker compose --env-file .env.staging -f docker-compose.staging.yml pull api"
assert_file_contains "$fake_remote_log" "migrate -path migrations -database postgres://postgres:postgres@127.0.0.1:5432/recova_stage?sslmode=disable up"
assert_file_contains "$fake_remote_log" "migrate -path migrations -database postgres://postgres:postgres@127.0.0.1:5432/recova_stage?sslmode=disable version"
assert_file_contains "$fake_remote_log" "curl -fsS --connect-timeout 5 --max-time 10 --retry 6 --retry-delay 2 --retry-connrefused http://127.0.0.1:3000/health/live"
assert_file_contains "$fake_remote_log" "curl -fsS --connect-timeout 5 --max-time 10 --retry 6 --retry-delay 2 --retry-connrefused http://127.0.0.1:3000/health/ready"
assert_file_contains "$fake_remote_log" "curl -fsS --connect-timeout 5 --max-time 10 --retry 4 --retry-delay 2 http://127.0.0.1:3000/openapi.yaml"
assert_file_contains "$fake_remote_log" "curl -sS --connect-timeout 5 --max-time 10 -o /dev/null -w %{http_code} http://127.0.0.1:3000/api/v1/users/me"
assert_file_not_contains "$fake_remote_log" "https://public.example.test/health/live"

# remote-deploy.sh must fail when APP_ENV is not staging.
cat > "$remote_repo_dir/.env.bad" <<'ENVFILE'
APP_ENV=production
DATABASE_URL=postgresql://postgres:postgres@127.0.0.1:5432/recova_stage?sslmode=disable
ENVFILE

assert_fail env \
  PATH="$remote_temp_dir:$PATH" \
  FAKE_REMOTE_LOG="$fake_remote_log" \
  MIGRATE_BIN="$remote_temp_dir/migrate" \
  "$remote_repo_dir/scripts/deploy/remote-deploy.sh" \
    "$remote_repo_dir" \
    "develop" \
    "ghcr.io/example/recova:sha-abc" \
    "docker-compose.staging.yml" \
    ".env.bad" \
    "staging" \
    "staging" \
    "3000" \
    "http://127.0.0.1:3001"

# deploy-staging workflow must stay develop-only and use immutable sha tag.
assert_file_contains .github/workflows/deploy-staging.yml "branches:"
assert_file_contains .github/workflows/deploy-staging.yml "- develop"
assert_file_contains .github/workflows/deploy-staging.yml 'image_tag_sha="sha-${source_sha}"'
assert_file_contains .github/workflows/deploy-staging.yml "name: staging"
assert_file_contains .github/workflows/deploy-staging.yml 'git checkout -B "${{ env.DEPLOY_BRANCH }}" "origin/${{ env.DEPLOY_BRANCH }}"'
assert_file_contains .github/workflows/deploy-staging.yml 'cat <<'"'"'ENVEOF'"'"' > "$remote_path/$runtime_env_file"'
