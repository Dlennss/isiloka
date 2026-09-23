ALTER TABLE public.produk_fee_provider
  ADD COLUMN IF NOT EXISTS produk_provider_map_id bigint NULL;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'fk_produk_fee_provider_map'
      AND conrelid = 'public.produk_fee_provider'::regclass
  ) THEN
    ALTER TABLE public.produk_fee_provider
      ADD CONSTRAINT fk_produk_fee_provider_map
      FOREIGN KEY (produk_provider_map_id) REFERENCES public.produk_provider_map(id) ON DELETE CASCADE;
  END IF;
END $$;

ALTER TABLE public.produk_fee_provider
  DROP CONSTRAINT IF EXISTS uq_produk_fee_provider_produk_provider;

CREATE UNIQUE INDEX IF NOT EXISTS uq_produk_fee_provider_map_active
  ON public.produk_fee_provider (produk_provider_map_id)
  WHERE produk_provider_map_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_produk_fee_provider_produk_provider_legacy
  ON public.produk_fee_provider (produk_id, provider_id)
  WHERE produk_provider_map_id IS NULL;

INSERT INTO public.produk_fee_provider
  (produk_id, provider_id, produk_provider_map_id, fee_rp, aktif, dibuat_pada, diubah_pada)
SELECT
  f.produk_id,
  f.provider_id,
  ppm.id,
  f.fee_rp,
  f.aktif,
  now(),
  now()
FROM public.produk_fee_provider f
JOIN public.provider pr
  ON pr.id = f.provider_id
JOIN public.produk_provider_map ppm
  ON ppm.produk_id = f.produk_id
 AND LOWER(TRIM(ppm.provider)) = LOWER(TRIM(pr.nama))
WHERE LOWER(TRIM(pr.nama)) = 'smb'
  AND f.produk_provider_map_id IS NULL
  AND NOT EXISTS (
    SELECT 1
    FROM public.produk_fee_provider fx
    WHERE fx.produk_provider_map_id = ppm.id
  );

DELETE FROM public.produk_fee_provider f
USING public.provider pr
WHERE pr.id = f.provider_id
  AND LOWER(TRIM(pr.nama)) = 'smb'
  AND f.produk_provider_map_id IS NULL;
