#!/usr/bin/env sh
set -eu

mode="${1:-watch}"

repo_root="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
watch_list_file="$repo_root/.openapi-watch-files"

generate() {
  echo "[openapi-autogen] regenerating OpenAPI artifacts"
  "$repo_root/scripts/openapi.sh" generate
}

install_hook() {
  hook_path="$repo_root/.git/hooks/pre-commit"
  cat >"$hook_path" <<'HOOK'
#!/usr/bin/env sh
set -eu

repo_root="$(git rev-parse --show-toplevel)"
"$repo_root/scripts/openapi.sh" generate

if git diff --quiet -- docs/generated/openapi.yaml docs/generated/routes.md; then
  exit 0
fi

git add docs/generated/openapi.yaml docs/generated/routes.md
HOOK
  chmod +x "$hook_path"
  echo "[openapi-autogen] installed git pre-commit hook at .git/hooks/pre-commit"
}

collect_watch_files() {
  (
    cd "$repo_root"
    {
      printf '%s\n' "api/openapi/openapi.yaml"
      rg --files internal/app/http internal/modules internal/platform/openapi cmd/tools/openapi
    } | sed '/^$/d' | sort -u
  )
}

watch_loop() {
  command -v shasum >/dev/null 2>&1 || {
    echo "[openapi-autogen] shasum command not found" >&2
    exit 1
  }

  collect_watch_files >"$watch_list_file"
  if [ ! -s "$watch_list_file" ]; then
    echo "[openapi-autogen] watch list is empty" >&2
    exit 1
  fi

  echo "[openapi-autogen] watching changes (Ctrl+C to stop)"
  checksum_prev=""
  while :; do
    checksum_now="$(
      (
        cd "$repo_root"
        xargs -I{} shasum "{}" <"$watch_list_file"
      ) | shasum | awk '{print $1}'
    )"
    if [ "$checksum_now" != "$checksum_prev" ]; then
      generate
      checksum_prev="$checksum_now"
    fi
    sleep 2
  done
}

case "$mode" in
  generate)
    generate
    ;;
  watch)
    watch_loop
    ;;
  install-hook)
    install_hook
    ;;
  *)
    echo "[openapi-autogen] unsupported mode: $mode" >&2
    echo "[openapi-autogen] use: generate | watch | install-hook" >&2
    exit 1
    ;;
esac
