ALTER TABLE public.produk_app_pricing
  ADD COLUMN IF NOT EXISTS yuscom_display_brand text NOT NULL DEFAULT '';

ALTER TABLE public.yuscom_produk_snapshot
  ADD COLUMN IF NOT EXISTS display_brand text NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS yuscom_produk_snapshot_display_brand_idx
  ON public.yuscom_produk_snapshot (display_brand);
