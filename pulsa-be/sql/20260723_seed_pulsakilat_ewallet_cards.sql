-- Seed provider dan produk E-Wallet untuk PulsaKilat.
-- Aman dijalankan berulang: brand, produk, dan harga di-upsert.

INSERT INTO public.kategori_fee_app
  (kategori_id, fee_master, fee_agent, fee_user, fee_non_user, aktif, created_at, updated_at)
SELECT k.id, 500, 500, 500, 1000, true, now(), now()
FROM public.kategori k
LEFT JOIN public.kategori_fee_app existing ON existing.kategori_id = k.id
WHERE lower(trim(k.nama)) IN ('e-money', 'e-wallet')
  AND existing.id IS NULL;

INSERT INTO public.brand (nama, aktif, dibuat_pada, diubah_pada)
SELECT x.nama, true, now(), now()
FROM (
  VALUES
    ('DANA'),
    ('GoPay'),
    ('OVO'),
    ('ShopeePay'),
    ('LinkAja'),
    ('AstraPay'),
    ('i.saku')
) AS x(nama)
WHERE NOT EXISTS (
  SELECT 1 FROM public.brand b WHERE lower(trim(b.nama)) = lower(trim(x.nama))
);

WITH seed(category_name, brand_name, sku, product_name, group_name, nominal, harga) AS (
  VALUES
    ('E-Wallet','DANA','PK-EWALLET-DANA-10000','DANA 10.000','Reguler',10000,10500),
    ('E-Wallet','DANA','PK-EWALLET-DANA-20000','DANA 20.000','Reguler',20000,20500),
    ('E-Wallet','DANA','PK-EWALLET-DANA-25000','DANA 25.000','Reguler',25000,25500),
    ('E-Wallet','DANA','PK-EWALLET-DANA-50000','DANA 50.000','Reguler',50000,50500),
    ('E-Wallet','DANA','PK-EWALLET-DANA-100000','DANA 100.000','Reguler',100000,100500),
    ('E-Wallet','DANA','PK-EWALLET-DANA-200000','DANA 200.000','Reguler',200000,200700),

    ('E-Wallet','GoPay','PK-EWALLET-GOPAY-10000','GoPay 10.000','Reguler',10000,10600),
    ('E-Wallet','GoPay','PK-EWALLET-GOPAY-20000','GoPay 20.000','Reguler',20000,20600),
    ('E-Wallet','GoPay','PK-EWALLET-GOPAY-25000','GoPay 25.000','Reguler',25000,25600),
    ('E-Wallet','GoPay','PK-EWALLET-GOPAY-50000','GoPay 50.000','Reguler',50000,50600),
    ('E-Wallet','GoPay','PK-EWALLET-GOPAY-100000','GoPay 100.000','Reguler',100000,100600),
    ('E-Wallet','GoPay','PK-EWALLET-GOPAY-200000','GoPay 200.000','Reguler',200000,200800),

    ('E-Wallet','OVO','PK-EWALLET-OVO-10000','OVO 10.000','Reguler',10000,10650),
    ('E-Wallet','OVO','PK-EWALLET-OVO-20000','OVO 20.000','Reguler',20000,20650),
    ('E-Wallet','OVO','PK-EWALLET-OVO-25000','OVO 25.000','Reguler',25000,25650),
    ('E-Wallet','OVO','PK-EWALLET-OVO-50000','OVO 50.000','Reguler',50000,50650),
    ('E-Wallet','OVO','PK-EWALLET-OVO-100000','OVO 100.000','Reguler',100000,100650),
    ('E-Wallet','OVO','PK-EWALLET-OVO-200000','OVO 200.000','Reguler',200000,200900),

    ('E-Wallet','ShopeePay','PK-EWALLET-SHOPEEPAY-10000','ShopeePay 10.000','Reguler',10000,10550),
    ('E-Wallet','ShopeePay','PK-EWALLET-SHOPEEPAY-20000','ShopeePay 20.000','Reguler',20000,20550),
    ('E-Wallet','ShopeePay','PK-EWALLET-SHOPEEPAY-25000','ShopeePay 25.000','Reguler',25000,25550),
    ('E-Wallet','ShopeePay','PK-EWALLET-SHOPEEPAY-50000','ShopeePay 50.000','Reguler',50000,50550),
    ('E-Wallet','ShopeePay','PK-EWALLET-SHOPEEPAY-100000','ShopeePay 100.000','Reguler',100000,100550),
    ('E-Wallet','ShopeePay','PK-EWALLET-SHOPEEPAY-200000','ShopeePay 200.000','Reguler',200000,200800),

    ('E-Wallet','LinkAja','PK-EWALLET-LINKAJA-10000','LinkAja 10.000','Reguler',10000,10600),
    ('E-Wallet','LinkAja','PK-EWALLET-LINKAJA-20000','LinkAja 20.000','Reguler',20000,20600),
    ('E-Wallet','LinkAja','PK-EWALLET-LINKAJA-25000','LinkAja 25.000','Reguler',25000,25600),
    ('E-Wallet','LinkAja','PK-EWALLET-LINKAJA-50000','LinkAja 50.000','Reguler',50000,50600),
    ('E-Wallet','LinkAja','PK-EWALLET-LINKAJA-100000','LinkAja 100.000','Reguler',100000,100600),
    ('E-Wallet','LinkAja','PK-EWALLET-LINKAJA-200000','LinkAja 200.000','Reguler',200000,200850),

    ('E-Wallet','AstraPay','PK-EWALLET-ASTRAPAY-10000','AstraPay 10.000','Reguler',10000,10600),
    ('E-Wallet','AstraPay','PK-EWALLET-ASTRAPAY-20000','AstraPay 20.000','Reguler',20000,20600),
    ('E-Wallet','AstraPay','PK-EWALLET-ASTRAPAY-25000','AstraPay 25.000','Reguler',25000,25600),
    ('E-Wallet','AstraPay','PK-EWALLET-ASTRAPAY-50000','AstraPay 50.000','Reguler',50000,50600),
    ('E-Wallet','AstraPay','PK-EWALLET-ASTRAPAY-100000','AstraPay 100.000','Reguler',100000,100600),
    ('E-Wallet','AstraPay','PK-EWALLET-ASTRAPAY-200000','AstraPay 200.000','Reguler',200000,200850),

    ('E-Wallet','i.saku','PK-EWALLET-ISAKU-10000','i.saku 10.000','Reguler',10000,10600),
    ('E-Wallet','i.saku','PK-EWALLET-ISAKU-20000','i.saku 20.000','Reguler',20000,20600),
    ('E-Wallet','i.saku','PK-EWALLET-ISAKU-25000','i.saku 25.000','Reguler',25000,25600),
    ('E-Wallet','i.saku','PK-EWALLET-ISAKU-50000','i.saku 50.000','Reguler',50000,50600),
    ('E-Wallet','i.saku','PK-EWALLET-ISAKU-100000','i.saku 100.000','Reguler',100000,100600),
    ('E-Wallet','i.saku','PK-EWALLET-ISAKU-200000','i.saku 200.000','Reguler',200000,200850)
),
upsert_products AS (
  INSERT INTO public.produk
    (sku, nama, group_name, kategori_id, brand_id, tipe_harga, nominal, aktif, dibuat_pada, diubah_pada)
  SELECT
    seed.sku,
    seed.product_name,
    seed.group_name,
    k.id,
    b.id,
    'FIXED',
    seed.nominal,
    true,
    now(),
    now()
  FROM seed
  JOIN public.kategori k ON lower(trim(k.nama)) = lower(trim(seed.category_name))
  JOIN public.brand b ON lower(trim(b.nama)) = lower(trim(seed.brand_name))
  ON CONFLICT (sku) DO UPDATE SET
    nama = EXCLUDED.nama,
    group_name = EXCLUDED.group_name,
    kategori_id = EXCLUDED.kategori_id,
    brand_id = EXCLUDED.brand_id,
    tipe_harga = EXCLUDED.tipe_harga,
    nominal = EXCLUDED.nominal,
    aktif = true,
    diubah_pada = now()
  RETURNING id, sku
)
INSERT INTO public.produk_app_pricing
  (produk_id, provider, harga, harga_dasar, fee_user, fee_agent, fee_master, harga_user, harga_agent, harga_master, aktif, created_at, updated_at, dibuat_pada, diubah_pada)
SELECT
  p.id,
  'yuscom',
  seed.harga,
  seed.harga,
  0,
  0,
  0,
  seed.harga,
  seed.harga,
  seed.harga,
  true,
  now(),
  now(),
  now(),
  now()
FROM seed
JOIN upsert_products p ON p.sku = seed.sku
ON CONFLICT (produk_id) DO UPDATE SET
  provider = EXCLUDED.provider,
  harga = EXCLUDED.harga,
  harga_dasar = EXCLUDED.harga_dasar,
  harga_user = EXCLUDED.harga_user,
  harga_agent = EXCLUDED.harga_agent,
  harga_master = EXCLUDED.harga_master,
  aktif = true,
  updated_at = now(),
  diubah_pada = now();
