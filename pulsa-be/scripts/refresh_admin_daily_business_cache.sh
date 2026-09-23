#!/usr/bin/env bash
set -euo pipefail

APP_DIR="${P24_APP_DIR:-/home/syarif/app/releases/pulsa-be/current}"
LOCK_FILE="${P24_ADMIN_BUSINESS_CACHE_LOCK:-/tmp/p24_admin_daily_business_cache_refresh.lock}"
DAYS_BACK="${P24_ADMIN_BUSINESS_CACHE_DAYS_BACK:-1}"

if ! [[ "$DAYS_BACK" =~ ^[0-9]+$ ]] || (( DAYS_BACK < 1 || DAYS_BACK > 366 )); then
  echo "P24_ADMIN_BUSINESS_CACHE_DAYS_BACK must be an integer from 1 to 366" >&2
  exit 1
fi

cd "$APP_DIR"
set -a
. ./.env
set +a

if [[ -z "${DATABASE_URL:-}" ]]; then
  echo "DATABASE_URL is empty" >&2
  exit 1
fi

SQL="SET lock_timeout = '5s'; SET statement_timeout = '180s'; SELECT public.refresh_admin_daily_business_cache(${DAYS_BACK}::int, '0 seconds'::interval);"

exec flock -n "$LOCK_FILE" psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -qAt -c "$SQL"
