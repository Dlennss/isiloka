-- Seed produk paket data Smartfren dari daftar provider.
-- Aman dijalankan berulang: produk lama Smartfren data generik dinonaktifkan, SKU baru di-upsert.

INSERT INTO public.brand (nama, aktif, dibuat_pada, diubah_pada)
SELECT 'Smartfren', true, now(), now()
WHERE NOT EXISTS (
  SELECT 1 FROM public.brand WHERE lower(trim(nama)) = 'smartfren'
);

UPDATE public.produk p
SET aktif = false, diubah_pada = now()
FROM public.kategori k, public.brand b
WHERE p.kategori_id = k.id
  AND p.brand_id = b.id
  AND lower(trim(k.nama)) = 'paket data'
  AND lower(trim(b.nama)) = 'smartfren'
  AND p.sku LIKE 'PK-DATA-SMARTFREN-%';

WITH seed(category_name, brand_name, sku, product_name, group_name, nominal, harga) AS (
  VALUES
    ('Paket Data','Smartfren','FDHNS110','Smart Data 10 GB + Unlimited NonStop 24 Jam 1 Hari','smartfren',10000,8725),
    ('Paket Data','Smartfren','FDHNS1412','Smart Data 12 GB + Unlimited NonStop 24 Jam 14 Hari','smartfren',12000,28500),
    ('Paket Data','Smartfren','FDHNS145','Smart Data 5 GB + Unlimited NonStop 24 Jam 14 Hari','smartfren',5000,20100),
    ('Paket Data','Smartfren','FDHNS18','Smart Data 8 GB + Unlimited NonStop 24 Jam 1 Hari','smartfren',8000,8700),
    ('Paket Data','Smartfren','FDHNS26','Smart Data 6 GB + Unlimited NonStop 24 Jam 2 Hari','smartfren',6000,9975),
    ('Paket Data','Smartfren','FDHNS310','Smart Data 10 GB + Unlimited NonStop 24 Jam 3 Hari','smartfren',10000,15125),
    ('Paket Data','Smartfren','FDHNS320','Smart Data 20 GB + Unlimited NonStop 24 Jam 3 Hari','smartfren',20000,20050),
    ('Paket Data','Smartfren','FDHNS36','Smart Data 6 GB + Unlimited NonStop 24 Jam 3 Hari','smartfren',6000,12250),
    ('Paket Data','Smartfren','FDHNS710','Smart Data 10 GB + Unlimited NonStop 24 Jam 7 Hari','smartfren',10000,19225),
    ('Paket Data','Smartfren','FDHNS740','Smart Data 40 GB + Unlimited NonStop 24 Jam 7 Hari','smartfren',40000,38700),

    ('Paket Data','Smartfren','FDHU11','Smartfren Data Unlimited 1 GB/Hari 1 Hari','smartfren',1000,7375),
    ('Paket Data','Smartfren','FDHU12','Smartfren Data Unlimited 2 GB/Hari 1 Hari','smartfren',2000,11126),
    ('Paket Data','Smartfren','FDHU13','Smartfren Data Unlimited 3 GB/Hari 1 Hari','smartfren',3000,13073),
    ('Paket Data','Smartfren','FDHU15','Smartfren Data Unlimited 5 GB/Hari 1 Hari','smartfren',5000,18100),
    ('Paket Data','Smartfren','FDHU31','Smartfren Data Unlimited 1 GB/Hari 3 Hari','smartfren',1000,12455),
    ('Paket Data','Smartfren','FDHU32','Smartfren Data Unlimited 2 GB/Hari 3 Hari','smartfren',2000,12800),
    ('Paket Data','Smartfren','FDHU33','Smartfren Data Unlimited 3 GB/Hari 3 Hari','smartfren',3000,15350),
    ('Paket Data','Smartfren','FDHU35','Smartfren Data Unlimited 5 GB/Hari 3 Hari','smartfren',5000,20300),
    ('Paket Data','Smartfren','FDHU72','Smartfren Data Unlimited 2 GB/Hari 7 Hari','smartfren',2000,24300),
    ('Paket Data','Smartfren','FDHU73','Smartfren Data Unlimited 3 GB/Hari 7 Hari','smartfren',3000,30300),
    ('Paket Data','Smartfren','FDHU75','Smartfren Data Unlimited 5 GB/Hari 7 Hari','smartfren',5000,39000),

    ('Paket Data','Smartfren','FDM1','Smartfren Data 1GB 3 Hari','SMART DATA MINI',1000,7000),
    ('Paket Data','Smartfren','FDM2','Smartfren Data 2GB 3 Hari','SMART DATA MINI',2000,9947),
    ('Paket Data','Smartfren','FDM3','Smartfren Data 3GB 5 Hari','SMART DATA MINI',3000,14480),
    ('Paket Data','Smartfren','FDM4','Smartfren Data 4GB 14 Hari','SMART DATA MINI',4000,20050),
    ('Paket Data','Smartfren','FDM5','Smartfren Data 2GB (1GB + 1GB CHAT) 7 Hari','SMART DATA MINI',2000,11350),
    ('Paket Data','Smartfren','FDM6','Smartfren Data 2,5GB (1,5GB + 1GB CHAT) 7 Hari','SMART DATA MINI',2500,14500),

    ('Paket Data','Smartfren','FDPNS10','Smart Data 20 GB + Unlimited NonStop 24 Jam 30 Hari','SMARTFREN DATA NONSTOP',20000,56920),
    ('Paket Data','Smartfren','FDPNS16','Smart Data 16 GB + Unlimited NonStop 24 Jam 30 Hari','SMARTFREN DATA NONSTOP',16000,44825),
    ('Paket Data','Smartfren','FDPNS2','Smart Data 2 GB + Unlimited NonStop 24 Jam 10 Hari','SMARTFREN DATA NONSTOP',2000,13400),
    ('Paket Data','Smartfren','FDPNS24','Smart Data 24 GB + Unlimited NonStop 24 Jam 28 Hari','smartfren',24000,56960),
    ('Paket Data','Smartfren','FDPNS3','Smart Data 3 GB + Unlimited NonStop 24 Jam 14 Hari','SMARTFREN DATA NONSTOP',3000,19625),
    ('Paket Data','Smartfren','FDPNS30','Smart Data 35 GB + Unlimited NonStop 24 Jam 30 Hari','SMARTFREN DATA NONSTOP',35000,74000),
    ('Paket Data','Smartfren','FDPNS4','Smart Data 4 GB + Unlimited NonStop 24 Jam 14 Hari','SMARTFREN DATA NONSTOP',4000,19680),
    ('Paket Data','Smartfren','FDPNS45','Smart Data 75 GB + Unlimited NonStop 24 Jam 30 Hari','SMARTFREN DATA NONSTOP',75000,116500),
    ('Paket Data','Smartfren','FDPNS6','Smart Data 10 GB + Unlimited NonStop 24 Jam 30 Hari','SMARTFREN DATA NONSTOP',10000,39800),
    ('Paket Data','Smartfren','FDPNS60','Smart Data 100 GB + Unlimited NonStop 24 Jam 30 Hari','SMARTFREN DATA NONSTOP',100000,120800),

    ('Paket Data','Smartfren','FDPU1','Smart Data Unlimited FUP 1GB/Hari 7 Hari','SMARTFREN DATA UNLIMITED',1000,17600),
    ('Paket Data','Smartfren','FDPU2','Smart Data Unlimited FUP 1GB/Hari 14 Hari','SMARTFREN DATA UNLIMITED',1000,32600),
    ('Paket Data','Smartfren','FDPU3','Smart Data Unlimited FUP 500MB/Hari 28 Hari','SMARTFREN DATA UNLIMITED',500,63800),
    ('Paket Data','Smartfren','FDPU4','Smart Data Unlimited FUP 700MB/Hari 28 Hari','SMARTFREN DATA UNLIMITED',700,63850),
    ('Paket Data','Smartfren','FDPU5','Smart Data Unlimited FUP 2GB/Hari 28 Hari','SMARTFREN DATA UNLIMITED',2000,90450),
    ('Paket Data','Smartfren','FDPU6','Smart Data Unlimited FUP 3GB/Hari 28 Hari','SMARTFREN DATA UNLIMITED',3000,117800),
    ('Paket Data','Smartfren','FDPU7','Smart Data Unlimited FUP 5GB/Hari 28 Hari','SMARTFREN DATA UNLIMITED',5000,143000),
    ('Paket Data','Smartfren','FDPU8','Smart Data Unlimited FUP 1GB/Hari 28 Hari','SMARTFREN DATA UNLIMITED',1000,63900),

    ('Paket Data','Smartfren','FDPVB2','Smart Data Volume Base 500MB + 1.5GB Midnite + 1 GB Chat 3 Hari','SMARTFREN DATA VOLUME BASE',3000,10200),
    ('Paket Data','Smartfren','FDPVB4','Smart Data Volume Base 2GB + 2GB + 2GB Chat 7 Hari','SMARTFREN DATA VOLUME BASE',6000,15800),

    ('Paket Data','Smartfren','FDVU14','Smartfren Voucher Unlimited FUP 1 GB/Hari 14 Hari','SMARTFREN VOUCHER DATA',1000,49300),
    ('Paket Data','Smartfren','FDVU28','Smartfren Voucher Unlimited FUP 0.5 GB/Hari 28 Hari','SMARTFREN VOUCHER DATA',500,69300),
    ('Paket Data','Smartfren','FDVU7','Voucher Data Unlimited FUP 1GB 7 Hari','SMARTFREN VOUCHER DATA',1000,22700),
    ('Paket Data','Smartfren','FDVU700','Voucher Data Unlimited FUP 700MB 28 Hari','SMARTFREN VOUCHER DATA',700,66750),
    ('Paket Data','Smartfren','FDVUNL1','Voucher Data Unlimited FUP 1GB 28 Hari','SMARTFREN VOUCHER DATA',1000,66760),
    ('Paket Data','Smartfren','FDVUNL2','Voucher Data Unlimited FUP 2GB 28 Hari','SMARTFREN VOUCHER DATA',2000,92250),

    ('Paket Data','Smartfren','VSUF1','VCR Smart Unli (FUP 1GB/Hari) 7 Hari','VOUCHER SMART UNLIMITED FUP',1000,27000),
    ('Paket Data','Smartfren','VSUN1','VCR Smart 2GB + Unli Nonstop 10 Hari','VOUCHER SMART UNLI NONSTOP',2000,14000),
    ('Paket Data','Smartfren','VSUN2','VCR Smart 4GB + Unli Nonstop 14 Hari','VOUCHER SMART UNLI NONSTOP',4000,19000),
    ('Paket Data','Smartfren','VSUN3','VCR Smart 10GB + Unli Nonstop 30 Hari','VOUCHER SMART UNLI NONSTOP',10000,41500),
    ('Paket Data','Smartfren','VSUN4','VCR Smart 15GB + Unli Nonstop 30 Hari','VOUCHER SMART UNLI NONSTOP',15000,58000),
    ('Paket Data','Smartfren','VSUN5','VCR Smart 35GB + Unli Nonstop 30 Hari','VOUCHER SMART UNLI NONSTOP',35000,80800),
    ('Paket Data','Smartfren','VSUN6','VCR Smart 75GB + Unli Nonstop 30 Hari','VOUCHER SMART UNLI NONSTOP',75000,118500),
    ('Paket Data','Smartfren','VSUN7','VCR Smart 100GB + Unli Nonstop 30 Hari','VOUCHER SMART UNLI NONSTOP',100000,157000)
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
