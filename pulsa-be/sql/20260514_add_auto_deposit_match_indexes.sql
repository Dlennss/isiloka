CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_deposit_request_auto_bank_pending
ON public.deposit_request (bank_id, amount, dibuat_pada, id)
WHERE status = 'pending' AND bank_id IS NOT NULL;

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_mutasi_bank_auto_deposit_match
ON public.mutasi_bank (bank_id, jumlah, dibuat_pada, id)
WHERE arah = 'CREDIT'
  AND alasan = 'BANK_MANUAL_IN'
  AND member_id IS NULL
  AND provider IS NULL;
