-- Seed provider dan produk PLN untuk PulsaKilat.
-- Aman dijalankan berulang: brand, produk, dan harga di-upsert.

INSERT INTO public.kategori_fee_app
  (kategori_id, fee_master, fee_agent, fee_user, fee_non_user, aktif, created_at, updated_at)
SELECT k.id, 1000, 1000, 1000, 1500, true, now(), now()
FROM public.kategori k
LEFT JOIN public.kategori_fee_app existing ON existing.kategori_id = k.id
WHERE lower(trim(k.nama)) IN ('pln', 'listrik')
  AND existing.id IS NULL;

INSERT INTO public.brand (nama, aktif, dibuat_pada, diubah_pada)
SELECT 'PLN', true, now(), now()
WHERE NOT EXISTS (
  SELECT 1 FROM public.brand b WHERE lower(trim(b.nama)) = 'pln'
);

WITH seed(category_name, brand_name, sku, product_name, group_name, nominal, harga) AS (
  VALUES
    ('PLN','PLN','CEKIDPLN','Cek ID PLN','Cek Meter',0,0),
    ('PLN','PLN','PK-PLN-TOKEN-20000','Token Listrik 20.000','Token Listrik',20000,20500),
    ('PLN','PLN','PK-PLN-TOKEN-50000','Token Listrik 50.000','Token Listrik',50000,50500),
    ('PLN','PLN','PK-PLN-TOKEN-100000','Token Listrik 100.000','Token Listrik',100000,100500),
    ('PLN','PLN','PK-PLN-TOKEN-200000','Token Listrik 200.000','Token Listrik',200000,200800),
    ('PLN','PLN','PK-PLN-TOKEN-500000','Token Listrik 500.000','Token Listrik',500000,501000),
    ('PLN','PLN','PK-PLN-TOKEN-1000000','Token Listrik 1.000.000','Token Listrik',1000000,1001500),
    ('PLN','PLN','PLNC','Cek Tagihan PLN Pascabayar','Tagihan Listrik',0,0),
    ('PLN','PLN','PLNB','Bayar Tagihan PLN Pascabayar','Tagihan Listrik',1,2500)
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
