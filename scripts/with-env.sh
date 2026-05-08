#!/usr/bin/env sh
set -eu

if [ "$#" -eq 0 ]; then
  echo "usage: scripts/with-env.sh <command> [arg...]" >&2
  exit 1
fi

env_file="${ENV_FILE:-.env}"

if [ -f "$env_file" ]; then
  set -a
  # shellcheck disable=SC1090
  . "$env_file"
  set +a
fi

exec "$@"
