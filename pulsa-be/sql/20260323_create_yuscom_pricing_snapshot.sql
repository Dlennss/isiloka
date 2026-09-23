ALTER TABLE public.produk_app_pricing
  ADD COLUMN IF NOT EXISTS yuscom_group text NOT NULL DEFAULT '';

ALTER TABLE public.produk_app_pricing
  ADD COLUMN IF NOT EXISTS yuscom_category text NOT NULL DEFAULT '';

ALTER TABLE public.produk_app_pricing
  ADD COLUMN IF NOT EXISTS yuscom_subcategory text NOT NULL DEFAULT '';

ALTER TABLE public.produk_app_pricing
  ADD COLUMN IF NOT EXISTS yuscom_sku text NOT NULL DEFAULT '';

ALTER TABLE public.produk_app_pricing
  ADD COLUMN IF NOT EXISTS yuscom_name text NOT NULL DEFAULT '';

ALTER TABLE public.produk_app_pricing
  ADD COLUMN IF NOT EXISTS yuscom_status text NOT NULL DEFAULT '';

CREATE TABLE IF NOT EXISTS public.yuscom_produk_snapshot (
  id bigserial PRIMARY KEY,
  provider text NOT NULL DEFAULT 'yuscom',
  sku text NOT NULL,
  nama text NOT NULL DEFAULT '',
  kelompok text NOT NULL DEFAULT '',
  kategori_yuscom text NOT NULL DEFAULT '',
  subkategori_yuscom text NOT NULL DEFAULT '',
  harga bigint NOT NULL DEFAULT 0,
  harga_text text NOT NULL DEFAULT '',
  status text NOT NULL DEFAULT '',
  aktif boolean NOT NULL DEFAULT true,
  produk_id bigint NULL REFERENCES public.produk(id) ON DELETE SET NULL,
  fetched_at timestamptz NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT yuscom_produk_snapshot_provider_sku_key UNIQUE (provider, sku)
);

CREATE INDEX IF NOT EXISTS yuscom_produk_snapshot_kelompok_idx
  ON public.yuscom_produk_snapshot (kelompok);

CREATE INDEX IF NOT EXISTS yuscom_produk_snapshot_kategori_idx
  ON public.yuscom_produk_snapshot (kategori_yuscom);

CREATE INDEX IF NOT EXISTS yuscom_produk_snapshot_subkategori_idx
  ON public.yuscom_produk_snapshot (subkategori_yuscom);

CREATE INDEX IF NOT EXISTS yuscom_produk_snapshot_produk_id_idx
  ON public.yuscom_produk_snapshot (produk_id);

CREATE INDEX IF NOT EXISTS produk_app_pricing_yuscom_group_idx
  ON public.produk_app_pricing (yuscom_group);
