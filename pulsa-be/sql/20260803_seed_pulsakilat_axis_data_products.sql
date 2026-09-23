-- Seed produk paket data AXIS dari daftar provider.
-- Aman dijalankan berulang: produk lama AXIS data generik dinonaktifkan, SKU baru di-upsert.

INSERT INTO public.brand (nama, aktif, dibuat_pada, diubah_pada)
SELECT 'AXIS', true, now(), now()
WHERE NOT EXISTS (
  SELECT 1 FROM public.brand WHERE lower(trim(nama)) = 'axis'
);

UPDATE public.produk p
SET aktif = false, diubah_pada = now()
FROM public.kategori k, public.brand b
WHERE p.kategori_id = k.id
  AND p.brand_id = b.id
  AND lower(trim(k.nama)) = 'paket data'
  AND lower(trim(b.nama)) = 'axis'
  AND p.sku LIKE 'PK-DATA-AXIS-%';

WITH seed(category_name, brand_name, sku, product_name, group_name, nominal, harga) AS (
  VALUES
    ('Paket Data','AXIS','ADBV1','BRONET 24Jam Kuota utama 1GB 30 hari','axis',1000,20000),
    ('Paket Data','AXIS','ADPAS10','Data Bronet 1,5GB + Kuota di Kotamu 1 Hari','AXIS DATA AIGO SS',1500,7300),
    ('Paket Data','AXIS','ADPAS11','Data Bronet 1GB + Kuota di Kotamu 1 Hari','AXIS DATA AIGO SS',1000,7285),
    ('Paket Data','AXIS','ADPAS154','Data Bronet 4 GB + Kuota di Kotamu 14 Hari','AXIS DATA AIGO SS',4000,22450),
    ('Paket Data','AXIS','ADPAS156','Data Bronet 6 GB + Kuota di Kotamu 14 Hari','AXIS DATA AIGO SS',6000,27775),
    ('Paket Data','AXIS','ADPAS19','Data Bronet Unlimited 1 Hari','AXIS DATA AIGO SS',0,24850),
    ('Paket Data','AXIS','ADPAS25','Data Bronet 5GB + Kuota di Kotamu 2 Hari','AXIS DATA AIGO SS',5000,9200),
    ('Paket Data','AXIS','ADPAS30','Data Bronet 1GB + Kuota di Kotamu 3 Hari','AXIS DATA AIGO SS',1000,10150),
    ('Paket Data','AXIS','ADPAS31','Data Bronet 1,5GB + Kuota di Kotamu 3 Hari','AXIS DATA AIGO SS',1500,10170),
    ('Paket Data','AXIS','ADPAS32','Data Bronet 2 GB + Kuota di Kotamu 3 Hari','AXIS DATA AIGO SS',2000,10200),
    ('Paket Data','AXIS','ADPAS33','Data Bronet 3GB + Kuota di Kotamu 3 Hari','AXIS DATA AIGO SS',3000,10220),
    ('Paket Data','AXIS','ADPAS36','Data Bronet 6 GB + Kuota di Kotamu 3 Hari','AXIS DATA AIGO SS',6000,13950),
    ('Paket Data','AXIS','ADPAS39','Data Bronet 30GB 3 Hari','AXIS DATA AIGO SS',30000,21070),
    ('Paket Data','AXIS','ADPAS50','Data Bronet 1GB + Kuota di Kotamu 5 Hari','AXIS DATA AIGO SS',1000,13400),
    ('Paket Data','AXIS','ADPAS51','Data Bronet 1,5GB + Kuota di Kotamu 5 Hari','AXIS DATA AIGO SS',1500,13420),
    ('Paket Data','AXIS','ADPAS52','Data Bronet 2,5 GB + Kuota di Kotamu 5 Hari','AXIS DATA AIGO SS',2500,13450),
    ('Paket Data','AXIS','ADPAS53','Data Bronet 3 GB + Kuota di Kotamu 5 Hari','AXIS DATA AIGO SS',3000,13600),
    ('Paket Data','AXIS','ADPAS54','Data Bronet 3,5GB + Kuota di Kotamu 5 Hari','axis',3500,13650),
    ('Paket Data','AXIS','ADPAS55','Data Bronet 5 GB + Kuota di Kotamu 5 Hari','AXIS DATA AIGO SS',5000,15750),
    ('Paket Data','AXIS','ADPAS60','Data Bronet 10 GB + Kuota di Kotamu 5 Hari','AXIS DATA AIGO SS',10000,23200),
    ('Paket Data','AXIS','ADPAS70','Data Bronet 10GB + Kuota di Kotamu 7 Hari','AXIS DATA AIGO SS',10000,23500),
    ('Paket Data','AXIS','ADPAS73','Data Bronet 3GB + Kuota di Kotamu 7 Hari','AXIS DATA AIGO SS',3000,14100),
    ('Paket Data','AXIS','ADPAS75','Data Bronet 15 GB + Kuota di Kotamu 7 Hari','AXIS DATA AIGO SS',15000,27000),

    ('Paket Data','AXIS','ADPB1','Data 1GB 30 Hari','AXIS DATA BRONET',1000,8945),
    ('Paket Data','AXIS','ADPB10','Data Bronet 10GB + Kuota di Kotamu 28 Hari','AXIS DATA BRONET',10000,35100),
    ('Paket Data','AXIS','ADPB12','Data Bronet 12GB + Kuota di Kotamu 28 Hari','axis',12000,46150),
    ('Paket Data','AXIS','ADPB14','Data Bronet 14GB + Kuota di Kotamu 28 Hari','AXIS DATA BRONET',14000,46300),
    ('Paket Data','AXIS','ADPB16','Data Bronet 16GB + Kuota di Kotamu 28 Hari','AXIS DATA BRONET',16000,46400),
    ('Paket Data','AXIS','ADPB2','Data 2GB 30 Hari','AXIS DATA BRONET',2000,15000),
    ('Paket Data','AXIS','ADPB20','Data Bronet 20GB + Kuota di Kotamu 28 Hari','AXIS DATA BRONET',20000,59000),
    ('Paket Data','AXIS','ADPB22','Data Bronet 22 GB + Kuota di Kotamu 28 Hari','AXIS DATA BRONET',22000,58800),
    ('Paket Data','AXIS','ADPB24','Data Bronet 24 GB + Kuota di Kotamu 28 Hari','AXIS DATA BRONET',24000,58825),
    ('Paket Data','AXIS','ADPB3','Data Bronet 3GB + Kuota di Kotamu 28 Hari','AXIS DATA BRONET',3000,19200),
    ('Paket Data','AXIS','ADPB30','Data Bronet 30GB + Kuota di Kotamu 28 Hari','AXIS DATA BRONET',30000,72800),
    ('Paket Data','AXIS','ADPB35','Data Bronet 35 GB + Kuota di Kotamu 28 Hari','AXIS DATA BRONET',35000,72900),
    ('Paket Data','AXIS','ADPB5','Data Bronet 5GB + Kuota di Kotamu 28 Hari','AXIS DATA BRONET',5000,28000),
    ('Paket Data','AXIS','ADPB6','Data Bronet 6GB + Kuota di Kotamu 28 Hari','AXIS DATA BRONET',6000,30650),
    ('Paket Data','AXIS','ADPB7','Data Bronet 7GB + Kuota di Kotamu 28 Hari','AXIS DATA BRONET',7000,30700),
    ('Paket Data','AXIS','ADPB8','Data Bronet 8GB + Kuota di Kotamu 28 Hari','AXIS DATA BRONET',8000,30750),
    ('Paket Data','AXIS','ADPBM1','Axis Data Mini 1GB + Kuota di Kotamu 5 Hari','AXIS DATA BRONET',1000,9000),
    ('Paket Data','AXIS','ADPBM2','Axis Data Mini 2GB + Kuota di Kotamu 7 Hari','AXIS DATA BRONET',2000,15390),
    ('Paket Data','AXIS','ADPBM3','Axis Data Mini 3GB + Kuota di Kotamu 15 Hari','AXIS DATA BRONET',3000,19000),
    ('Paket Data','AXIS','ADPBM5','Axis Data Mini 5GB + Kuota di Kotamu 15 Hari','AXIS DATA BRONET',5000,23000),

    ('Paket Data','AXIS','ADPD1','Axis Data Pure 1GB 30 Hari','AXIS DATA PURE',1000,8800),
    ('Paket Data','AXIS','ADPD10','Axis Data Pure 10 GB 30 Hari','AXIS DATA PURE',10000,78500),
    ('Paket Data','AXIS','ADPD12','Axis Data Pure 12 GB 30 Hari','AXIS DATA PURE',12000,47000),
    ('Paket Data','AXIS','ADPD2','Axis Data Pure 2 GB 30 Hari','AXIS DATA PURE',2000,16600),
    ('Paket Data','AXIS','ADPD3','Axis Data Pure 3 GB 30 Hari','AXIS DATA PURE',3000,24400),
    ('Paket Data','AXIS','ADPD4','Axis Data Pure 4 GB 30 Hari','AXIS DATA PURE',4000,32200),
    ('Paket Data','AXIS','ADPD5','Axis Data Pure 5 GB 30 Hari','AXIS DATA PURE',5000,40000),
    ('Paket Data','AXIS','ADPD8','Axis Data Pure 8 GB 30 Hari','AXIS DATA PURE',8000,63000),

    ('Paket Data','AXIS','ADPF10','Data Bronet 10 GB 60 Hari','AXIS DATA BRONET 60 HARI',10000,85000),
    ('Paket Data','AXIS','ADPF12','Data Bronet 12 GB 60 Hari','AXIS DATA BRONET 60 HARI',12000,94425),
    ('Paket Data','AXIS','ADPF16','Data Bronet 16 GB + Kuota di Kotamu 60 Hari','AXIS DATA BRONET 60 HARI',16000,106400),
    ('Paket Data','AXIS','ADPF2','Data Bronet 2 GB 60 Hari','AXIS DATA BRONET 60 HARI',2000,29075),
    ('Paket Data','AXIS','ADPF3','Data Bronet 3 GB 60 Hari','AXIS DATA BRONET 60 HARI',3000,36300),
    ('Paket Data','AXIS','ADPF35','Data Bronet 35 GB + Kuota di Kotamu 60 Hari','AXIS DATA BRONET 60 HARI',35000,106800),
    ('Paket Data','AXIS','ADPF5','Data Bronet 5 GB 60 Hari','AXIS DATA BRONET 60 HARI',5000,53350),
    ('Paket Data','AXIS','ADPF75','Data Bronet 75 GB + Kuota di Kotamu 60 Hari','AXIS DATA BRONET 60 HARI',75000,168500),
    ('Paket Data','AXIS','ADPF8','Data Bronet 8 GB 60 Hari','AXIS DATA BRONET 60 HARI',8000,73450),

    ('Paket Data','AXIS','ADPMVA1','Voucher Aigo Mini 1GB + Lokal 5 Hari','AXIS VOUCHER AIGO MINI',1000,14600),
    ('Paket Data','AXIS','ADPMVA3','Voucher Aigo Mini 3,5GB + Lokal 7 Hari','AXIS VOUCHER AIGO MINI',3500,25000),
    ('Paket Data','AXIS','ADPMVA4','Voucher Aigo Mini 4GB + Lokal 15 Hari','AXIS VOUCHER AIGO MINI',4000,29500),
    ('Paket Data','AXIS','ADPMVA8','Voucher Aigo Mini 8GB + Lokal 15 Hari','AXIS VOUCHER AIGO MINI',8000,31000),
    ('Paket Data','AXIS','ADPVA2','Voucher Axis Aigo 2GB + Kuota di Kotamu 28 Hari','AXIS VOUCHER AIGO',2000,32900),
    ('Paket Data','AXIS','ADPVA3','Voucher Axis Aigo 3GB + Kuota di Kotamu 28 Hari','AXIS VOUCHER AIGO',3000,33000),
    ('Paket Data','AXIS','ADPVA30','Voucher Axis Aigo 30GB + Kuota di Kotamu 28 Hari','AXIS VOUCHER AIGO',30000,73000),
    ('Paket Data','AXIS','ADPVA6','Voucher Axis Aigo 6GB + Kuota di Kotamu 28 Hari','AXIS VOUCHER AIGO',6000,33100),
    ('Paket Data','AXIS','ADPVA75','Voucher Axis Aigo 75GB + Kuota di Kotamu 28 Hari','AXIS VOUCHER AIGO',75000,120000),
    ('Paket Data','AXIS','ADPVA8','Voucher Axis Aigo 8GB + Kuota di Kotamu 28 Hari','AXIS VOUCHER AIGO',8000,38700),
    ('Paket Data','AXIS','ADUW1','Data Warnet Unlimited 1 Jam','axis',0,2600)
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
