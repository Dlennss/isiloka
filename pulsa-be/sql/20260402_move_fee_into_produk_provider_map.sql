BEGIN;

ALTER TABLE public.produk_provider_map
  ADD COLUMN IF NOT EXISTS fee_rp bigint NOT NULL DEFAULT 0;

WITH map_fee AS (
  SELECT DISTINCT ON (f.produk_provider_map_id)
    f.produk_provider_map_id AS map_id,
    COALESCE(f.fee_rp, 0) AS fee_rp
  FROM public.produk_fee_provider f
  WHERE f.produk_provider_map_id IS NOT NULL
    AND f.aktif = true
  ORDER BY f.produk_provider_map_id, f.id DESC
)
UPDATE public.produk_provider_map ppm
SET fee_rp = map_fee.fee_rp,
    diubah_pada = NOW()
FROM map_fee
WHERE ppm.id = map_fee.map_id;

WITH legacy_fee AS (
  SELECT DISTINCT ON (ppm.id)
    ppm.id AS map_id,
    COALESCE(f.fee_rp, 0) AS fee_rp
  FROM public.produk_provider_map ppm
  JOIN public.produk_fee_provider f
    ON f.produk_id = ppm.produk_id
   AND f.produk_provider_map_id IS NULL
   AND f.aktif = true
  JOIN public.provider pr
    ON pr.id = f.provider_id
   AND LOWER(TRIM(pr.nama)) = LOWER(TRIM(ppm.provider))
  ORDER BY ppm.id, f.id DESC
)
UPDATE public.produk_provider_map ppm
SET fee_rp = legacy_fee.fee_rp,
    diubah_pada = NOW()
FROM legacy_fee
WHERE ppm.id = legacy_fee.map_id
  AND COALESCE(ppm.fee_rp, 0) = 0;

DROP TABLE IF EXISTS public.produk_fee_provider;

COMMIT;
