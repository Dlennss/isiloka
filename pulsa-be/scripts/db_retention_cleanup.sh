#!/usr/bin/env bash
set -euo pipefail

MODE="${1:-dry-run}"
RETENTION_INTERVAL="${RETENTION_INTERVAL:-3 months}"
BATCH_SIZE="${BATCH_SIZE:-50000}"
APP_DIR="${APP_DIR:-/home/syarif/app/pulsa-be}"
ARCHIVE_ROOT="${ARCHIVE_ROOT:-/home/syarif/app/data/db-retention-archive}"
LOG_FILE="${LOG_FILE:-/home/syarif/app/logs/db-retention-cleanup.log}"

if [[ "$MODE" != "dry-run" && "$MODE" != "run" ]]; then
  echo "usage: $0 [dry-run|run]" >&2
  exit 2
fi

mkdir -p "$(dirname "$LOG_FILE")"

log() {
  printf '[%s] %s\n' "$(date '+%Y-%m-%d %H:%M:%S%z')" "$*" | tee -a "$LOG_FILE"
}

if [[ ! -f "$APP_DIR/.env" ]]; then
  log "missing env file: $APP_DIR/.env"
  exit 1
fi

set -a
# shellcheck source=/dev/null
. "$APP_DIR/.env"
set +a

if [[ -z "${DATABASE_URL:-}" ]]; then
  log "DATABASE_URL is empty"
  exit 1
fi

PSQL=(psql "$DATABASE_URL" -X -v ON_ERROR_STOP=1)
RUN_ID="$(date '+%Y%m%d-%H%M%S')"
ARCHIVE_DIR="$ARCHIVE_ROOT/$RUN_ID"

scalar() {
  "${PSQL[@]}" -Atc "$1"
}

CUTOFF_VALUE="$(scalar "select now() - interval '$RETENTION_INTERVAL'")"
CUTOFF_VALUE="${CUTOFF_VALUE//$'\r'/}"
CUTOFF_VALUE="${CUTOFF_VALUE//$'\n'/}"
CUTOFF_SQL_LITERAL="${CUTOFF_VALUE//\'/\'\'}"
CUTOFF_EXPR="timestamp with time zone '$CUTOFF_SQL_LITERAL'"

archive_query() {
  local label="$1"
  local select_sql="$2"
  local count="$3"

  if [[ "$MODE" != "run" || "$count" == "0" ]]; then
    return
  fi

  mkdir -p "$ARCHIVE_DIR"
  local file="$ARCHIVE_DIR/${label}.csv.gz"
  log "archive $label count=$count file=$file"
  "${PSQL[@]}" -c "\\copy ($select_sql) TO STDOUT WITH CSV HEADER" | gzip -c > "$file"
}

delete_batches() {
  local label="$1"
  local delete_template="$2"

  if [[ "$MODE" != "run" ]]; then
    return
  fi

  local total=0
  while true; do
    local deleted
    deleted="$(scalar "${delete_template//__BATCH_SIZE__/$BATCH_SIZE}")"
    deleted="${deleted//$'\r'/}"
    deleted="${deleted//$'\n'/}"
    [[ -z "$deleted" ]] && deleted=0
    total=$((total + deleted))
    log "delete batch $label deleted=$deleted total=$total"
    [[ "$deleted" == "0" ]] && break
    sleep 0.2
  done
}

run_item() {
  local label="$1"
  local count_sql="$2"
  local select_sql="$3"
  local delete_template="$4"

  local count
  count="$(scalar "$count_sql")"
  count="${count//$'\r'/}"
  count="${count//$'\n'/}"
  [[ -z "$count" ]] && count=0
  log "$MODE $label candidates=$count retention='$RETENTION_INTERVAL'"
  archive_query "$label" "$select_sql" "$count"
  delete_batches "$label" "$delete_template"
}

log "start mode=$MODE retention='$RETENTION_INTERVAL' cutoff='$CUTOFF_VALUE' batch_size=$BATCH_SIZE"

if [[ "$(scalar "select pg_try_advisory_lock(hashtext('p24_db_retention_cleanup'))")" != "t" ]]; then
  log "another retention cleanup is already running"
  exit 0
fi

trap 'scalar "select pg_advisory_unlock(hashtext('\''p24_db_retention_cleanup'\''))" >/dev/null 2>&1 || true' EXIT

# Child rows and logs first.
run_item "audit_provider_wallet_missing_debit_ignore" \
  "SELECT count(*) FROM public.audit_provider_wallet_missing_debit_ignore a WHERE a.dibuat_pada < $CUTOFF_EXPR OR EXISTS (SELECT 1 FROM public.transaksi_provider tp WHERE tp.id = a.transaksi_provider_id AND tp.dibuat_pada < $CUTOFF_EXPR)" \
  "SELECT a.* FROM public.audit_provider_wallet_missing_debit_ignore a WHERE a.dibuat_pada < $CUTOFF_EXPR OR EXISTS (SELECT 1 FROM public.transaksi_provider tp WHERE tp.id = a.transaksi_provider_id AND tp.dibuat_pada < $CUTOFF_EXPR) ORDER BY a.id" \
  "WITH doomed AS (SELECT a.id FROM public.audit_provider_wallet_missing_debit_ignore a WHERE a.dibuat_pada < $CUTOFF_EXPR OR EXISTS (SELECT 1 FROM public.transaksi_provider tp WHERE tp.id = a.transaksi_provider_id AND tp.dibuat_pada < $CUTOFF_EXPR) ORDER BY a.id LIMIT __BATCH_SIZE__), deleted AS (DELETE FROM public.audit_provider_wallet_missing_debit_ignore t USING doomed d WHERE t.id=d.id RETURNING 1) SELECT count(*) FROM deleted"

run_item "transaksi_member_status_log" \
  "SELECT count(*) FROM public.transaksi_member_status_log l WHERE l.dibuat_pada < $CUTOFF_EXPR OR EXISTS (SELECT 1 FROM public.transaksi_member tm WHERE tm.id = l.transaksi_member_id AND tm.dibuat_pada < $CUTOFF_EXPR)" \
  "SELECT l.* FROM public.transaksi_member_status_log l WHERE l.dibuat_pada < $CUTOFF_EXPR OR EXISTS (SELECT 1 FROM public.transaksi_member tm WHERE tm.id = l.transaksi_member_id AND tm.dibuat_pada < $CUTOFF_EXPR) ORDER BY l.id" \
  "WITH doomed AS (SELECT l.id FROM public.transaksi_member_status_log l WHERE l.dibuat_pada < $CUTOFF_EXPR OR EXISTS (SELECT 1 FROM public.transaksi_member tm WHERE tm.id = l.transaksi_member_id AND tm.dibuat_pada < $CUTOFF_EXPR) ORDER BY l.id LIMIT __BATCH_SIZE__), deleted AS (DELETE FROM public.transaksi_member_status_log t USING doomed d WHERE t.id=d.id RETURNING 1) SELECT count(*) FROM deleted"

run_item "h2h_commission_ledger" \
  "SELECT count(*) FROM public.h2h_commission_ledger h WHERE h.created_at < $CUTOFF_EXPR OR EXISTS (SELECT 1 FROM public.transaksi_member tm WHERE tm.id = h.source_trx_member_id AND tm.dibuat_pada < $CUTOFF_EXPR)" \
  "SELECT h.* FROM public.h2h_commission_ledger h WHERE h.created_at < $CUTOFF_EXPR OR EXISTS (SELECT 1 FROM public.transaksi_member tm WHERE tm.id = h.source_trx_member_id AND tm.dibuat_pada < $CUTOFF_EXPR) ORDER BY h.id" \
  "WITH doomed AS (SELECT h.id FROM public.h2h_commission_ledger h WHERE h.created_at < $CUTOFF_EXPR OR EXISTS (SELECT 1 FROM public.transaksi_member tm WHERE tm.id = h.source_trx_member_id AND tm.dibuat_pada < $CUTOFF_EXPR) ORDER BY h.id LIMIT __BATCH_SIZE__), deleted AS (DELETE FROM public.h2h_commission_ledger t USING doomed d WHERE t.id=d.id RETURNING 1) SELECT count(*) FROM deleted"

run_item "retail_commission_ledger" \
  "SELECT count(*) FROM public.retail_commission_ledger r WHERE r.created_at < $CUTOFF_EXPR OR EXISTS (SELECT 1 FROM public.app_order o WHERE o.id = r.source_app_order_id AND o.dibuat_pada < $CUTOFF_EXPR)" \
  "SELECT r.* FROM public.retail_commission_ledger r WHERE r.created_at < $CUTOFF_EXPR OR EXISTS (SELECT 1 FROM public.app_order o WHERE o.id = r.source_app_order_id AND o.dibuat_pada < $CUTOFF_EXPR) ORDER BY r.id" \
  "WITH doomed AS (SELECT r.id FROM public.retail_commission_ledger r WHERE r.created_at < $CUTOFF_EXPR OR EXISTS (SELECT 1 FROM public.app_order o WHERE o.id = r.source_app_order_id AND o.dibuat_pada < $CUTOFF_EXPR) ORDER BY r.id LIMIT __BATCH_SIZE__), deleted AS (DELETE FROM public.retail_commission_ledger t USING doomed d WHERE t.id=d.id RETURNING 1) SELECT count(*) FROM deleted"

run_item "app_order_guest_refund" \
  "SELECT count(*) FROM public.app_order_guest_refund g WHERE g.dibuat_pada < $CUTOFF_EXPR OR EXISTS (SELECT 1 FROM public.app_order o WHERE o.id = g.app_order_id AND o.dibuat_pada < $CUTOFF_EXPR)" \
  "SELECT g.* FROM public.app_order_guest_refund g WHERE g.dibuat_pada < $CUTOFF_EXPR OR EXISTS (SELECT 1 FROM public.app_order o WHERE o.id = g.app_order_id AND o.dibuat_pada < $CUTOFF_EXPR) ORDER BY g.id" \
  "WITH doomed AS (SELECT g.id FROM public.app_order_guest_refund g WHERE g.dibuat_pada < $CUTOFF_EXPR OR EXISTS (SELECT 1 FROM public.app_order o WHERE o.id = g.app_order_id AND o.dibuat_pada < $CUTOFF_EXPR) ORDER BY g.id LIMIT __BATCH_SIZE__), deleted AS (DELETE FROM public.app_order_guest_refund t USING doomed d WHERE t.id=d.id RETURNING 1) SELECT count(*) FROM deleted"

run_item "app_order_payment" \
  "SELECT count(*) FROM public.app_order_payment p WHERE p.dibuat_pada < $CUTOFF_EXPR OR EXISTS (SELECT 1 FROM public.app_order o WHERE o.id = p.app_order_id AND o.dibuat_pada < $CUTOFF_EXPR)" \
  "SELECT p.* FROM public.app_order_payment p WHERE p.dibuat_pada < $CUTOFF_EXPR OR EXISTS (SELECT 1 FROM public.app_order o WHERE o.id = p.app_order_id AND o.dibuat_pada < $CUTOFF_EXPR) ORDER BY p.id" \
  "WITH doomed AS (SELECT p.id FROM public.app_order_payment p WHERE p.dibuat_pada < $CUTOFF_EXPR OR EXISTS (SELECT 1 FROM public.app_order o WHERE o.id = p.app_order_id AND o.dibuat_pada < $CUTOFF_EXPR) ORDER BY p.id LIMIT __BATCH_SIZE__), deleted AS (DELETE FROM public.app_order_payment t USING doomed d WHERE t.id=d.id RETURNING 1) SELECT count(*) FROM deleted"

# Wallet/provider child rows before app/provider transaction parents.
run_item "mutasi_dompet_provider" \
  "SELECT count(*) FROM public.mutasi_dompet_provider m WHERE m.dibuat_pada < $CUTOFF_EXPR OR EXISTS (SELECT 1 FROM public.app_order_provider_trx apt WHERE apt.id = m.app_order_provider_trx_id AND apt.dibuat_pada < $CUTOFF_EXPR)" \
  "SELECT m.* FROM public.mutasi_dompet_provider m WHERE m.dibuat_pada < $CUTOFF_EXPR OR EXISTS (SELECT 1 FROM public.app_order_provider_trx apt WHERE apt.id = m.app_order_provider_trx_id AND apt.dibuat_pada < $CUTOFF_EXPR) ORDER BY m.id" \
  "WITH doomed AS (SELECT m.id FROM public.mutasi_dompet_provider m WHERE m.dibuat_pada < $CUTOFF_EXPR OR EXISTS (SELECT 1 FROM public.app_order_provider_trx apt WHERE apt.id = m.app_order_provider_trx_id AND apt.dibuat_pada < $CUTOFF_EXPR) ORDER BY m.id LIMIT __BATCH_SIZE__), deleted AS (DELETE FROM public.mutasi_dompet_provider t USING doomed d WHERE t.id=d.id RETURNING 1) SELECT count(*) FROM deleted"

run_item "mutasi_dompet" \
  "SELECT count(*) FROM public.mutasi_dompet m WHERE m.dibuat_pada < $CUTOFF_EXPR" \
  "SELECT m.* FROM public.mutasi_dompet m WHERE m.dibuat_pada < $CUTOFF_EXPR ORDER BY m.id" \
  "WITH doomed AS (SELECT m.id FROM public.mutasi_dompet m WHERE m.dibuat_pada < $CUTOFF_EXPR ORDER BY m.id LIMIT __BATCH_SIZE__), deleted AS (DELETE FROM public.mutasi_dompet t USING doomed d WHERE t.id=d.id RETURNING 1) SELECT count(*) FROM deleted"

run_item "provider_saldo_snapshot" \
  "SELECT count(*) FROM public.provider_saldo_snapshot s WHERE s.dibuat_pada < $CUTOFF_EXPR" \
  "SELECT s.* FROM public.provider_saldo_snapshot s WHERE s.dibuat_pada < $CUTOFF_EXPR ORDER BY s.id" \
  "WITH doomed AS (SELECT s.id FROM public.provider_saldo_snapshot s WHERE s.dibuat_pada < $CUTOFF_EXPR ORDER BY s.id LIMIT __BATCH_SIZE__), deleted AS (DELETE FROM public.provider_saldo_snapshot t USING doomed d WHERE t.id=d.id RETURNING 1) SELECT count(*) FROM deleted"

# Retail app transactions.
run_item "app_order_provider_trx" \
  "SELECT count(*) FROM public.app_order_provider_trx apt WHERE apt.dibuat_pada < $CUTOFF_EXPR OR EXISTS (SELECT 1 FROM public.app_order o WHERE o.id = apt.app_order_id AND o.dibuat_pada < $CUTOFF_EXPR)" \
  "SELECT apt.* FROM public.app_order_provider_trx apt WHERE apt.dibuat_pada < $CUTOFF_EXPR OR EXISTS (SELECT 1 FROM public.app_order o WHERE o.id = apt.app_order_id AND o.dibuat_pada < $CUTOFF_EXPR) ORDER BY apt.id" \
  "WITH doomed AS (SELECT apt.id FROM public.app_order_provider_trx apt WHERE apt.dibuat_pada < $CUTOFF_EXPR OR EXISTS (SELECT 1 FROM public.app_order o WHERE o.id = apt.app_order_id AND o.dibuat_pada < $CUTOFF_EXPR) ORDER BY apt.id LIMIT __BATCH_SIZE__), deleted AS (DELETE FROM public.app_order_provider_trx t USING doomed d WHERE t.id=d.id RETURNING 1) SELECT count(*) FROM deleted"

run_item "app_order" \
  "SELECT count(*) FROM public.app_order o WHERE o.dibuat_pada < $CUTOFF_EXPR" \
  "SELECT o.* FROM public.app_order o WHERE o.dibuat_pada < $CUTOFF_EXPR ORDER BY o.id" \
  "WITH doomed AS (SELECT o.id FROM public.app_order o WHERE o.dibuat_pada < $CUTOFF_EXPR ORDER BY o.id LIMIT __BATCH_SIZE__), deleted AS (DELETE FROM public.app_order t USING doomed d WHERE t.id=d.id RETURNING 1) SELECT count(*) FROM deleted"

# H2H transactions.
run_item "transaksi_provider" \
  "SELECT count(*) FROM public.transaksi_provider tp WHERE tp.dibuat_pada < $CUTOFF_EXPR OR EXISTS (SELECT 1 FROM public.transaksi_member tm WHERE tm.id = tp.transaksi_member_id AND tm.dibuat_pada < $CUTOFF_EXPR)" \
  "SELECT tp.* FROM public.transaksi_provider tp WHERE tp.dibuat_pada < $CUTOFF_EXPR OR EXISTS (SELECT 1 FROM public.transaksi_member tm WHERE tm.id = tp.transaksi_member_id AND tm.dibuat_pada < $CUTOFF_EXPR) ORDER BY tp.id" \
  "WITH doomed AS (SELECT tp.id FROM public.transaksi_provider tp WHERE tp.dibuat_pada < $CUTOFF_EXPR OR EXISTS (SELECT 1 FROM public.transaksi_member tm WHERE tm.id = tp.transaksi_member_id AND tm.dibuat_pada < $CUTOFF_EXPR) ORDER BY tp.id LIMIT __BATCH_SIZE__), deleted AS (DELETE FROM public.transaksi_provider t USING doomed d WHERE t.id=d.id RETURNING 1) SELECT count(*) FROM deleted"

run_item "transaksi_member" \
  "SELECT count(*) FROM public.transaksi_member tm WHERE tm.dibuat_pada < $CUTOFF_EXPR" \
  "SELECT tm.* FROM public.transaksi_member tm WHERE tm.dibuat_pada < $CUTOFF_EXPR ORDER BY tm.id" \
  "WITH doomed AS (SELECT tm.id FROM public.transaksi_member tm WHERE tm.dibuat_pada < $CUTOFF_EXPR ORDER BY tm.id LIMIT __BATCH_SIZE__), deleted AS (DELETE FROM public.transaksi_member t USING doomed d WHERE t.id=d.id RETURNING 1) SELECT count(*) FROM deleted"

# Bank/deposit/mutation operational history.
run_item "deposit_request" \
  "SELECT count(*) FROM public.deposit_request d WHERE d.dibuat_pada < $CUTOFF_EXPR" \
  "SELECT d.* FROM public.deposit_request d WHERE d.dibuat_pada < $CUTOFF_EXPR ORDER BY d.id" \
  "WITH doomed AS (SELECT d.id FROM public.deposit_request d WHERE d.dibuat_pada < $CUTOFF_EXPR ORDER BY d.id LIMIT __BATCH_SIZE__), deleted AS (DELETE FROM public.deposit_request t USING doomed d WHERE t.id=d.id RETURNING 1) SELECT count(*) FROM deleted"

run_item "mutasi_bank" \
  "SELECT count(*) FROM public.mutasi_bank b WHERE COALESCE(b.waktu_mutasi_bank, b.dibuat_pada) < $CUTOFF_EXPR" \
  "SELECT b.* FROM public.mutasi_bank b WHERE COALESCE(b.waktu_mutasi_bank, b.dibuat_pada) < $CUTOFF_EXPR ORDER BY b.id" \
  "WITH doomed AS (SELECT b.id FROM public.mutasi_bank b WHERE COALESCE(b.waktu_mutasi_bank, b.dibuat_pada) < $CUTOFF_EXPR ORDER BY b.id LIMIT __BATCH_SIZE__), deleted AS (DELETE FROM public.mutasi_bank t USING doomed d WHERE t.id=d.id RETURNING 1) SELECT count(*) FROM deleted"

run_item "transaksi_anomasi_provider" \
  "SELECT count(*) FROM public.transaksi_anomasi_provider a WHERE a.dibuat_pada < $CUTOFF_EXPR" \
  "SELECT a.* FROM public.transaksi_anomasi_provider a WHERE a.dibuat_pada < $CUTOFF_EXPR ORDER BY a.id" \
  "WITH doomed AS (SELECT a.id FROM public.transaksi_anomasi_provider a WHERE a.dibuat_pada < $CUTOFF_EXPR ORDER BY a.id LIMIT __BATCH_SIZE__), deleted AS (DELETE FROM public.transaksi_anomasi_provider t USING doomed d WHERE t.id=d.id RETURNING 1) SELECT count(*) FROM deleted"

log "done mode=$MODE archive_dir=${ARCHIVE_DIR:-}"
