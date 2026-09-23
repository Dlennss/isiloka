CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_provider_saldo_snapshot_provider_trx_latest
ON public.provider_saldo_snapshot (
  provider,
  transaksi_provider_id DESC NULLS LAST,
  dibuat_pada DESC,
  id DESC
);
