ALTER TABLE public.transaksi_provider
  ADD COLUMN IF NOT EXISTS produk_sku_snapshot text,
  ADD COLUMN IF NOT EXISTS produk_provider_map_id bigint;

CREATE INDEX IF NOT EXISTS transaksi_provider_ref_provider_map_idx
  ON public.transaksi_provider (ref_id, provider, produk_provider_map_id, id DESC);

CREATE INDEX IF NOT EXISTS transaksi_provider_ref_provider_code_idx
  ON public.transaksi_provider (ref_id, provider, kode_produk, id DESC);
