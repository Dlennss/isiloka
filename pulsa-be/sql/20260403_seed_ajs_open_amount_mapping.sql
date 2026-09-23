BEGIN;

-- Seed provider AJS in disabled state first.
INSERT INTO public.provider (nama, aktif, dibuat_pada, diubah_pada)
SELECT 'ajs', false, NOW(), NOW()
WHERE NOT EXISTS (
  SELECT 1 FROM public.provider WHERE lower(trim(nama)) = 'ajs'
);

UPDATE public.provider
SET aktif = false,
    diubah_pada = NOW()
WHERE lower(trim(nama)) = 'ajs';

WITH desired AS (
  SELECT *
  FROM (VALUES
    ('DANA',     'DND',     1000::bigint,    10000000::bigint,   90::bigint),
    ('SHOPEE',   'SHPPAY', 10000::bigint,     1000000::bigint,  300::bigint),
    ('GOPAY',    'GPY',    10000::bigint,     1000000::bigint,  925::bigint),
    ('LINKAJA',  'LINKAJA',10000::bigint,     1000000::bigint,  828::bigint),
    ('OVO',      'OVO',    10000::bigint,     1000000::bigint,  665::bigint)
  ) AS x(sku, kode_provider, minimal_nominal, maksimal_nominal, fee_rp)
), target AS (
  SELECT p.id AS produk_id,
         d.sku,
         d.kode_provider,
         d.minimal_nominal,
         d.maksimal_nominal,
         d.fee_rp
  FROM desired d
  JOIN public.produk p
    ON upper(trim(p.sku)) = d.sku
)
UPDATE public.produk_provider_map ppm
SET kode_provider = t.kode_provider,
    minimal_nominal = t.minimal_nominal,
    maksimal_nominal = t.maksimal_nominal,
    fee_rp = t.fee_rp,
    aktif = false,
    diubah_pada = NOW()
FROM target t
WHERE ppm.produk_id = t.produk_id
  AND lower(trim(ppm.provider)) = 'ajs'
  AND upper(trim(ppm.kode_provider)) = upper(trim(t.kode_provider));

WITH desired AS (
  SELECT *
  FROM (VALUES
    ('DANA',     'DND',     1000::bigint,    10000000::bigint,   90::bigint),
    ('SHOPEE',   'SHPPAY', 10000::bigint,     1000000::bigint,  300::bigint),
    ('GOPAY',    'GPY',    10000::bigint,     1000000::bigint,  925::bigint),
    ('LINKAJA',  'LINKAJA',10000::bigint,     1000000::bigint,  828::bigint),
    ('OVO',      'OVO',    10000::bigint,     1000000::bigint,  665::bigint)
  ) AS x(sku, kode_provider, minimal_nominal, maksimal_nominal, fee_rp)
), target AS (
  SELECT p.id AS produk_id,
         d.sku,
         d.kode_provider,
         d.minimal_nominal,
         d.maksimal_nominal,
         d.fee_rp
  FROM desired d
  JOIN public.produk p
    ON upper(trim(p.sku)) = d.sku
)
INSERT INTO public.produk_provider_map
  (produk_id, provider, kode_provider, minimal_nominal, maksimal_nominal, fee_rp, aktif, dibuat_pada, diubah_pada)
SELECT t.produk_id, 'ajs', t.kode_provider, t.minimal_nominal, t.maksimal_nominal, t.fee_rp, false, NOW(), NOW()
FROM target t
WHERE NOT EXISTS (
  SELECT 1
  FROM public.produk_provider_map ppm
  WHERE ppm.produk_id = t.produk_id
    AND lower(trim(ppm.provider)) = 'ajs'
    AND upper(trim(ppm.kode_provider)) = upper(trim(t.kode_provider))
);

COMMIT;
