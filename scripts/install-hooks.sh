#!/usr/bin/env sh
set -eu

repo_root="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
hook_src="$repo_root/.githooks/pre-commit"

if [ ! -f "$hook_src" ]; then
  echo "[hooks] error: pre-commit hook not found at $hook_src" >&2
  exit 1
fi

chmod +x "$hook_src"

git config core.hooksPath .githooks

git_hooks_dir="$repo_root/.git/hooks"
if [ -d "$git_hooks_dir" ]; then
  cp "$hook_src" "$git_hooks_dir/pre-commit"
  chmod +x "$git_hooks_dir/pre-commit"
fi

echo "[hooks] pre-commit hook installed successfully (core.hooksPath = .githooks)"
