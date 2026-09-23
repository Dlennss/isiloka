BEGIN;

INSERT INTO public.provider (nama, aktif, dibuat_pada, diubah_pada)
SELECT 'sagaramobile', false, NOW(), NOW()
WHERE NOT EXISTS (
  SELECT 1 FROM public.provider WHERE lower(trim(nama)) = 'sagaramobile'
);

WITH provider_row AS (
  SELECT id
  FROM public.provider
  WHERE lower(trim(nama)) = 'sagaramobile'
  ORDER BY id DESC
  LIMIT 1
), desired AS (
  SELECT *
  FROM (VALUES
    ('DANA',     'DANARP',  1000::bigint, 5000000::bigint,   30::bigint),
    ('SHOPEE',   'SHPRP',  20000::bigint, 1000000::bigint,  450::bigint),
    ('LINKAJA',  'LARP',   20000::bigint, 1000000::bigint,  675::bigint),
    ('OVO',      'OVONRP', 10000::bigint, 2000000::bigint,  850::bigint),
    ('GOPAY',    'SGORP',  20000::bigint, 1000000::bigint, 1100::bigint)
  ) AS x(sku, kode_provider, minimal_nominal, maksimal_nominal, fee_rp)
), target AS (
  SELECT p.id AS produk_id,
         d.sku,
         d.kode_provider,
         d.minimal_nominal,
         d.maksimal_nominal,
         d.fee_rp,
         pr.id AS provider_id
  FROM desired d
  JOIN public.produk p
    ON upper(trim(p.sku)) = d.sku
  CROSS JOIN provider_row pr
)
UPDATE public.produk_provider_map ppm
SET kode_provider = t.kode_provider,
    minimal_nominal = t.minimal_nominal,
    maksimal_nominal = t.maksimal_nominal,
    fee_rp = t.fee_rp,
    aktif = true,
    diubah_pada = NOW()
FROM target t
WHERE ppm.produk_id = t.produk_id
  AND lower(trim(ppm.provider)) = 'sagaramobile';

WITH provider_row AS (
  SELECT id
  FROM public.provider
  WHERE lower(trim(nama)) = 'sagaramobile'
  ORDER BY id DESC
  LIMIT 1
), desired AS (
  SELECT *
  FROM (VALUES
    ('DANA',     'DANARP',  1000::bigint, 5000000::bigint,   30::bigint),
    ('SHOPEE',   'SHPRP',  20000::bigint, 1000000::bigint,  450::bigint),
    ('LINKAJA',  'LARP',   20000::bigint, 1000000::bigint,  675::bigint),
    ('OVO',      'OVONRP', 10000::bigint, 2000000::bigint,  850::bigint),
    ('GOPAY',    'SGORP',  20000::bigint, 1000000::bigint, 1100::bigint)
  ) AS x(sku, kode_provider, minimal_nominal, maksimal_nominal, fee_rp)
), target AS (
  SELECT p.id AS produk_id,
         d.sku,
         d.kode_provider,
         d.minimal_nominal,
         d.maksimal_nominal,
         d.fee_rp,
         pr.id AS provider_id
  FROM desired d
  JOIN public.produk p
    ON upper(trim(p.sku)) = d.sku
  CROSS JOIN provider_row pr
)
INSERT INTO public.produk_provider_map
  (produk_id, provider, kode_provider, minimal_nominal, maksimal_nominal, fee_rp, aktif, dibuat_pada, diubah_pada)
SELECT t.produk_id, 'sagaramobile', t.kode_provider, t.minimal_nominal, t.maksimal_nominal, t.fee_rp, true, NOW(), NOW()
FROM target t
WHERE NOT EXISTS (
  SELECT 1
  FROM public.produk_provider_map ppm
  WHERE ppm.produk_id = t.produk_id
    AND lower(trim(ppm.provider)) = 'sagaramobile'
);

COMMIT;
