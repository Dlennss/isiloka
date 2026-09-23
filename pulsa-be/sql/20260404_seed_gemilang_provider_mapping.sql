INSERT INTO public.provider (nama, aktif)
SELECT 'gemilang', true
WHERE NOT EXISTS (
  SELECT 1 FROM public.provider WHERE lower(trim(nama)) = 'gemilang'
);

UPDATE public.provider
SET aktif = true
WHERE lower(trim(nama)) = 'gemilang';

DELETE FROM public.produk_provider_map
WHERE lower(trim(provider)) = 'gemilang'
  AND produk_id IN (
    SELECT id
    FROM public.produk
    WHERE upper(trim(sku)) IN ('GOPAY', 'DANA', 'SHOPEE', 'OVO', 'LINKAJA')
  );

INSERT INTO public.produk_provider_map (produk_id, provider, kode_provider, aktif, minimal_nominal, maksimal_nominal, fee_rp)
SELECT p.id, 'gemilang', m.kode_provider, true, m.minimal_nominal, m.maksimal_nominal, m.fee_rp
FROM public.produk p
JOIN (
  VALUES
    ('GOPAY',   'GPO',   1000::bigint,    5000000::bigint, 860::bigint),
    ('DANA',    'DNBS',  1::bigint,       5000000::bigint, 33::bigint),
    ('SHOPEE',  'SHPB',  10000::bigint,   5000000::bigint, 105::bigint),
    ('OVO',     'OVBN',  10000::bigint,   2000000::bigint, 635::bigint),
    ('LINKAJA', 'LINKN', 10000::bigint,   2000000::bigint, 155::bigint)
) AS m(sku, kode_provider, minimal_nominal, maksimal_nominal, fee_rp)
  ON upper(trim(p.sku)) = m.sku;
