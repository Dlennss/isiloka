-- Performance indexes for high-traffic transaction and catalog queries.
-- Apply manually with CREATE INDEX CONCURRENTLY in production, one statement per transaction.

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_tp_ref_provider_status_id_desc
  ON public.transaksi_provider (ref_id, provider, status, id DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_tp_ref_status_id_desc
  ON public.transaksi_provider (ref_id, status, id DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_tm_member_updated_id_desc
  ON public.transaksi_member (member_id, diperbarui_pada DESC, id DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_tm_member_status_updated_id_desc
  ON public.transaksi_member (member_id, status, diperbarui_pada DESC, id DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_md_member_id_asc
  ON public.mutasi_dompet (member_id, id ASC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_md_member_alasan_id_asc
  ON public.mutasi_dompet (member_id, alasan, id ASC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_yuscom_snapshot_provider_norm_sku
  ON public.yuscom_produk_snapshot (provider, UPPER(TRIM(COALESCE(sku, ''::text))));

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_produk_app_pricing_provider_norm_yuscom_sku
  ON public.produk_app_pricing (provider, UPPER(TRIM(COALESCE(yuscom_sku, ''::text))));
