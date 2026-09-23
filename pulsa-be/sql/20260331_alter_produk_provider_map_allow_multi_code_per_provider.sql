BEGIN;

ALTER TABLE public.produk_provider_map
  DROP CONSTRAINT IF EXISTS uq_produk_provider_map_produk_provider;

DROP INDEX IF EXISTS public.uq_produk_provider_map_produk_provider;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'uq_produk_provider_map_produk_provider_kode'
      AND conrelid = 'public.produk_provider_map'::regclass
  ) THEN
    ALTER TABLE public.produk_provider_map
      ADD CONSTRAINT uq_produk_provider_map_produk_provider_kode
      UNIQUE (produk_id, provider, kode_provider);
  END IF;
END $$;

COMMIT;
