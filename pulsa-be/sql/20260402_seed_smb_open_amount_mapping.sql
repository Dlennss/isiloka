BEGIN;

INSERT INTO public.provider (nama, aktif, dibuat_pada, diubah_pada)
SELECT 'smb', false, NOW(), NOW()
WHERE NOT EXISTS (
  SELECT 1 FROM public.provider WHERE lower(trim(nama)) = 'smb'
);

WITH desired AS (
  SELECT *
  FROM (VALUES
    ('DANA',    'DANA',      1::bigint,      NULL::bigint),
    ('OVO',     'OVO',       10000::bigint,  2000000::bigint),
    ('GOPAY',   'GOPAY',     10000::bigint,  NULL::bigint),
    ('SHOPEE',  'SHOPEEPAY', 10000::bigint,  2000000::bigint),
    ('LINKAJA', 'LINKAJA',   10000::bigint, 10000000::bigint),
    ('DANA',    'ELDN',      1::bigint,      NULL::bigint),
    ('GOPAY',   'GPYOPEN',   10000::bigint,  NULL::bigint),
    ('SHOPEE',  'SHPOPEN',   10000::bigint, 10000000::bigint),
    ('OVO',     'ELOV',      10000::bigint,  2000000::bigint),
    ('LINKAJA', 'ELLI',      10000::bigint, 10000000::bigint)
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
SELECT t.produk_id, 'smb', t.kode_provider, t.minimal_nominal, t.maksimal_nominal, false, NOW(), NOW()
FROM target t
WHERE NOT EXISTS (
  SELECT 1
  FROM public.produk_provider_map ppm
  WHERE ppm.produk_id = t.produk_id
    AND lower(trim(ppm.provider)) = 'smb'
    AND upper(trim(ppm.kode_provider)) = upper(trim(t.kode_provider))
);

COMMIT;
