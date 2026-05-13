#!/usr/bin/env sh
set -eu

if [ "$#" -eq 0 ]; then
  echo "usage: scripts/with-env.sh <command> [arg...]" >&2
  exit 1
fi

env_file="${ENV_FILE:-.env}"

strip_env_quotes() {
  value="$1"
  case "$value" in
    \"*\")
      printf '%s' "${value#\"}" | sed 's/"$//'
      ;;
    \'*\')
      printf '%s' "${value#\'}" | sed "s/'$//"
      ;;
    *)
      printf '%s' "$value"
      ;;
  esac
}

if [ -f "$env_file" ]; then
  while IFS= read -r line || [ -n "$line" ]; do
    line="$(printf '%s' "$line" | sed 's/\r$//')"
    case "$line" in
      ''|'#'*)
        continue
        ;;
    esac

    case "$line" in
      *=*)
        key="${line%%=*}"
        value="${line#*=}"
        ;;
      *)
        continue
        ;;
    esac

    key="$(printf '%s' "$key" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')"
    value="$(printf '%s' "$value" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')"
    value="$(strip_env_quotes "$value")"

    case "$key" in
      ''|*[!A-Za-z0-9_]*)
        continue
        ;;
    esac

    eval "is_set=\${$key+x}"
    if [ -n "${is_set:-}" ]; then
      continue
    fi

    export "$key=$value"
  done < "$env_file"
fi

exec "$@"
