#!/usr/bin/env bash
set -euo pipefail

APP_DIR="${P24_APP_DIR:-/home/syarif/app/releases/pulsa-be/current}"
SQL_FILE="${P24_SUSPECT_CACHE_SQL:-$APP_DIR/scripts/refresh_provider_success_suspect_cache.sql}"
LOCK_FILE="${P24_SUSPECT_CACHE_LOCK:-/tmp/p24_suspect_cache_refresh.lock}"
DAYS_BACK="${P24_SUSPECT_CACHE_DAYS_BACK:-3}"

cd "$APP_DIR"
set -a
. ./.env
set +a

if [[ -z "${DATABASE_URL:-}" ]]; then
  echo "DATABASE_URL is empty" >&2
  exit 1
fi

exec flock -n "$LOCK_FILE" psql "$DATABASE_URL" -v "days_back=$DAYS_BACK" -qAt -f "$SQL_FILE"
