ALTER TABLE public.produk_provider_map
  ADD COLUMN IF NOT EXISTS minimal_nominal BIGINT NULL;

ALTER TABLE public.produk_provider_map
  ADD COLUMN IF NOT EXISTS maksimal_nominal BIGINT NULL;
