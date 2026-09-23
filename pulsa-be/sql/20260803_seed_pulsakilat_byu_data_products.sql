-- Seed produk paket data by.U dari daftar provider.
-- Aman dijalankan berulang: produk lama by.U data generik dinonaktifkan, SKU baru di-upsert.

INSERT INTO public.brand (nama, aktif, dibuat_pada, diubah_pada)
SELECT 'by.U', true, now(), now()
WHERE NOT EXISTS (
  SELECT 1 FROM public.brand WHERE lower(trim(nama)) = 'by.u'
);

UPDATE public.produk p
SET aktif = false, diubah_pada = now()
FROM public.kategori k, public.brand b
WHERE p.kategori_id = k.id
  AND p.brand_id = b.id
  AND lower(trim(k.nama)) = 'paket data'
  AND lower(trim(b.nama)) = 'by.u'
  AND p.sku LIKE 'PK-DATA-BYU-%';

WITH seed(category_name, brand_name, sku, product_name, group_name, nominal, harga) AS (
  VALUES
    ('Paket Data','by.U','BDPB1','Data BYU 1 GB 30 Hari','by.u',1000,34560),
    ('Paket Data','by.U','BDPB10','Data BYU 10 GB 30 Hari','BY.U DATA BULANAN',10000,43700),
    ('Paket Data','by.U','BDPB11','Data BYU 11 GB 30 Hari','BY.U DATA BULANAN',11000,43750),
    ('Paket Data','by.U','BDPB2','Data BYU 2 GB 30 Hari','by.u',2000,34580),
    ('Paket Data','by.U','BDPB20','Data BYU 20 GB 30 Hari','BY.U DATA BULANAN',20000,69150),
    ('Paket Data','by.U','BDPB3','Data BYU 3 GB 30 Hari','by.u',3000,34600),
    ('Paket Data','by.U','BDPB33','Data BYU 33 GB 30 Hari','BY.U DATA BULANAN',33000,69900),
    ('Paket Data','by.U','BDPB4','Data BYU 4 GB 30 Hari','BY.U DATA BULANAN',4000,34620),
    ('Paket Data','by.U','BDPB42','Data BYU 42 GB 30 Hari','BY.U DATA BULANAN',42000,92100),
    ('Paket Data','by.U','BDPB5','Data BYU 5 GB 30 Hari','BY.U DATA BULANAN',5000,34640),
    ('Paket Data','by.U','BDPB50','Data BYU 50 GB 30 Hari','BY.U DATA BULANAN',50000,92300),
    ('Paket Data','by.U','BDPB6','Data BYU 6 GB 30 Hari','BY.U DATA BULANAN',6000,34650),
    ('Paket Data','by.U','BDPB65','Data BYU 65 GB 30 Hari','BY.U DATA BULANAN',65000,96500),
    ('Paket Data','by.U','BDPB7','Data BYU 7 GB 30 Hari','BY.U DATA BULANAN',7000,34680),
    ('Paket Data','by.U','BDPB8','Data BYU 8 GB 30 Hari','BY.U DATA BULANAN',8000,43600),
    ('Paket Data','by.U','BDPB9','Data BYU 9 GB 30 Hari','BY.U DATA BULANAN',9000,43650),

    ('Paket Data','by.U','BDPH110','Data BYU 10 GB 1 Hari','BY.U DATA HARIAN',10000,10550),
    ('Paket Data','by.U','BDPH12','Data BYU 2 GB 1 Hari','BY.U DATA HARIAN',2000,7450),
    ('Paket Data','by.U','BDPH143','Data BYU 3 GB 14 Hari','BY.U DATA HARIAN',3000,17250),
    ('Paket Data','by.U','BDPH36','Data BYU 3 GB 6 Hari','BY.U DATA HARIAN',3000,12600),
    ('Paket Data','by.U','BDPH57','Data BYU 7,5 GB 5 Hari','BY.U DATA HARIAN',7500,15300),
    ('Paket Data','by.U','BDPH73','Data BYU 3 GB 7 Hari','BY.U DATA HARIAN',3000,12800),
    ('Paket Data','by.U','BDPH74','Data BYU 4 GB 7 Hari','BY.U DATA HARIAN',4000,12850),
    ('Paket Data','by.U','BDPH75','Data BYU 5 GB 7 Hari','BY.U DATA HARIAN',5000,14645)
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
