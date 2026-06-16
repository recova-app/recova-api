#!/usr/bin/env bash
set -euo pipefail

base_ref="${MIGRATION_BASE_REF:-origin/main}"
head_ref="${MIGRATION_HEAD_REF:-HEAD}"
approve_destructive="${APPROVE_DESTRUCTIVE_MIGRATIONS:-false}"
backup_evidence="${BACKUP_EVIDENCE_URL:-}"

if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  echo "[migration-gate] not inside git repository" >&2
  exit 1
fi

if [ "$base_ref" = "0000000000000000000000000000000000000000" ]; then
  base_ref="HEAD^"
fi

if ! git rev-parse --verify "$base_ref" >/dev/null 2>&1; then
  echo "[migration-gate] base ref not found: $base_ref" >&2
  exit 1
fi

changed_migrations="$(git diff --name-only "$base_ref" "$head_ref" -- 'migrations/*.sql' || true)"
if [ -z "$changed_migrations" ]; then
  echo "[migration-gate] no migration changes detected"
  exit 0
fi

echo "[migration-gate] migration changes detected:"
printf '%s\n' "$changed_migrations"

if [ -z "$backup_evidence" ]; then
  echo "[migration-gate] warning: BACKUP_EVIDENCE_URL is empty; continuing for non-destructive migration"
else
  echo "[migration-gate] backup evidence present"
fi

if git diff "$base_ref" "$head_ref" -- $changed_migrations | grep -Eiq '(^|[^a-z])(drop[[:space:]]+(table|column|constraint|index|schema)|truncate[[:space:]]+table|delete[[:space:]]+from|alter[[:space:]]+table[^;]+drop)([^a-z]|$)'; then
  if [ "$approve_destructive" != "true" ]; then
    echo "[migration-gate] destructive migration pattern detected; set APPROVE_DESTRUCTIVE_MIGRATIONS=true only after review" >&2
    exit 1
  fi
  echo "[migration-gate] destructive migration explicitly approved"
else
  echo "[migration-gate] no destructive migration pattern detected"
fi
