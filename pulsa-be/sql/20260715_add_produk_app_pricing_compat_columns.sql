ALTER TABLE public.produk_app_pricing
  ADD COLUMN IF NOT EXISTS provider TEXT NOT NULL DEFAULT 'yuscom',
  ADD COLUMN IF NOT EXISTS harga BIGINT NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS yuscom_group TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS yuscom_category TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS yuscom_subcategory TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS yuscom_sku TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS yuscom_name TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS yuscom_status TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS yuscom_display_brand TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS fetched_at TIMESTAMPTZ NULL,
  ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

UPDATE public.produk_app_pricing
SET
  provider = COALESCE(NULLIF(TRIM(provider), ''), 'yuscom'),
  harga = CASE
    WHEN COALESCE(harga, 0) > 0 THEN harga
    ELSE COALESCE(harga_dasar, 0)
  END,
  created_at = COALESCE(created_at, dibuat_pada, now()),
  updated_at = COALESCE(updated_at, diubah_pada, now())
WHERE provider IS NULL
   OR TRIM(provider) = ''
   OR COALESCE(harga, 0) = 0
   OR created_at IS NULL
   OR updated_at IS NULL;

CREATE INDEX IF NOT EXISTS produk_app_pricing_provider_norm_idx
  ON public.produk_app_pricing ((LOWER(TRIM(provider))));

CREATE INDEX IF NOT EXISTS produk_app_pricing_yuscom_group_idx
  ON public.produk_app_pricing (yuscom_group);
