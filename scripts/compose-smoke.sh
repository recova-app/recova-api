#!/usr/bin/env sh
set -eu

compose_file="${COMPOSE_FILE:-docker-compose.local.yml}"
env_file="${ENV_FILE:-.env.example}"
project_name="${COMPOSE_PROJECT_NAME:-recova-smoke}"
docker_bin="${DOCKER_BIN:-docker}"

if ! command -v "$docker_bin" >/dev/null 2>&1; then
  echo "[compose-smoke] docker command not found: $docker_bin" >&2
  exit 1
fi

if [ ! -f "$compose_file" ]; then
  echo "[compose-smoke] compose file not found: $compose_file" >&2
  exit 1
fi

if [ ! -f "$env_file" ]; then
  echo "[compose-smoke] env file not found: $env_file" >&2
  exit 1
fi

cleanup() {
  "$docker_bin" compose --env-file "$env_file" -f "$compose_file" -p "$project_name" down -v >/dev/null 2>&1 || true
}
trap cleanup EXIT

"$docker_bin" compose --env-file "$env_file" -f "$compose_file" -p "$project_name" up --build --wait --wait-timeout 120

"$docker_bin" compose --env-file "$env_file" -f "$compose_file" -p "$project_name" ps
