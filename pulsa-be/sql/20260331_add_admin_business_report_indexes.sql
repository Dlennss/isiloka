CREATE INDEX CONCURRENTLY IF NOT EXISTS retail_commission_ledger_source_order_idx
  ON public.retail_commission_ledger (source_app_order_id);

CREATE INDEX CONCURRENTLY IF NOT EXISTS h2h_commission_ledger_source_trx_idx
  ON public.h2h_commission_ledger (source_trx_member_id);

CREATE INDEX CONCURRENTLY IF NOT EXISTS app_order_provider_trx_order_status_updated_idx
  ON public.app_order_provider_trx (app_order_id, status, diubah_pada DESC, dibuat_pada DESC, id DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS transaksi_provider_member_id_id_desc_idx
  ON public.transaksi_provider (transaksi_member_id, id DESC);
