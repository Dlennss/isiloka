BEGIN;

INSERT INTO public.provider (nama, aktif, dibuat_pada, diubah_pada)
SELECT 'minions', false, NOW(), NOW()
WHERE NOT EXISTS (
  SELECT 1 FROM public.provider WHERE lower(trim(nama)) = 'minions'
);

WITH desired AS (
  SELECT *
  FROM (VALUES
    ('DANA',    'HDANA',    10000::bigint,  5000000::bigint),
    ('DANA',    'HDANAB',   10000::bigint, 10000000::bigint),
    ('GOPAY',   'HGOPAY',   10000::bigint,  2000000::bigint),
    ('LINKAJA', 'HLINK',    10000::bigint,  5000000::bigint),
    ('LINKAJA', 'HLINKB',   10000::bigint, 10000000::bigint),
    ('OVO',     'HOVO',     10000::bigint,  5000000::bigint),
    ('SHOPEE',  'HSHOPEE',  10000::bigint,  5000000::bigint),
    ('SHOPEE',  'HSHOPEEB', 10000::bigint, 10000000::bigint)
  ) AS x(sku, kode_provider, minimal_nominal, maksimal_nominal)
), target AS (
  SELECT p.id AS produk_id,
         d.kode_provider,
         d.minimal_nominal,
         d.maksimal_nominal
  FROM desired d
  JOIN public.produk p
    ON upper(trim(p.sku)) = d.sku
)
UPDATE public.produk_provider_map ppm
SET kode_provider = t.kode_provider,
    minimal_nominal = t.minimal_nominal,
    maksimal_nominal = t.maksimal_nominal,
    aktif = true,
    diubah_pada = NOW()
FROM target t
WHERE ppm.produk_id = t.produk_id
  AND lower(trim(ppm.provider)) = 'minions'
  AND upper(trim(ppm.kode_provider)) = upper(trim(t.kode_provider));

WITH desired AS (
  SELECT *
  FROM (VALUES
    ('DANA',    'HDANA',    10000::bigint,  5000000::bigint),
    ('DANA',    'HDANAB',   10000::bigint, 10000000::bigint),
    ('GOPAY',   'HGOPAY',   10000::bigint,  2000000::bigint),
    ('LINKAJA', 'HLINK',    10000::bigint,  5000000::bigint),
    ('LINKAJA', 'HLINKB',   10000::bigint, 10000000::bigint),
    ('OVO',     'HOVO',     10000::bigint,  5000000::bigint),
    ('SHOPEE',  'HSHOPEE',  10000::bigint,  5000000::bigint),
    ('SHOPEE',  'HSHOPEEB', 10000::bigint, 10000000::bigint)
  ) AS x(sku, kode_provider, minimal_nominal, maksimal_nominal)
), target AS (
  SELECT p.id AS produk_id,
         d.kode_provider,
         d.minimal_nominal,
         d.maksimal_nominal
  FROM desired d
  JOIN public.produk p
    ON upper(trim(p.sku)) = d.sku
)
INSERT INTO public.produk_provider_map
  (produk_id, provider, kode_provider, minimal_nominal, maksimal_nominal, aktif, dibuat_pada, diubah_pada)
SELECT t.produk_id, 'minions', t.kode_provider, t.minimal_nominal, t.maksimal_nominal, true, NOW(), NOW()
FROM target t
WHERE NOT EXISTS (
  SELECT 1
  FROM public.produk_provider_map ppm
  WHERE ppm.produk_id = t.produk_id
    AND lower(trim(ppm.provider)) = 'minions'
    AND upper(trim(ppm.kode_provider)) = upper(trim(t.kode_provider))
);

COMMIT;
