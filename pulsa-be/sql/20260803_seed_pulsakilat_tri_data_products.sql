-- Seed produk paket data Tri dari daftar provider.
-- Aman dijalankan berulang: produk lama Tri data generik dinonaktifkan, SKU baru di-upsert.

INSERT INTO public.brand (nama, aktif, dibuat_pada, diubah_pada)
SELECT 'Tri', true, now(), now()
WHERE NOT EXISTS (
  SELECT 1 FROM public.brand WHERE lower(trim(nama)) = 'tri'
);

UPDATE public.produk p
SET aktif = false, diubah_pada = now()
FROM public.kategori k, public.brand b
WHERE p.kategori_id = k.id
  AND p.brand_id = b.id
  AND lower(trim(k.nama)) = 'paket data'
  AND lower(trim(b.nama)) = 'tri'
  AND p.sku LIKE 'PK-DATA-TRI-%';

WITH seed(category_name, brand_name, sku, product_name, group_name, nominal, harga) AS (
  VALUES
    ('Paket Data','Tri','TDBP10','Tri Data Pure 1 GB 14 Hari','TRI DATA PURE',1000,6645),
    ('Paket Data','Tri','TDBP100','Tri Data Pure 10 GB 30 Hari','TRI DATA PURE',10000,57250),
    ('Paket Data','Tri','TDBP20','Tri Data Pure 2 GB 7 Hari','TRI DATA PURE',2000,12290),
    ('Paket Data','Tri','TDBP30','Tri Data Pure 3 GB 30 Hari','TRI DATA PURE',3000,17875),
    ('Paket Data','Tri','TDBP40','Tri Data Pure 4 GB 30 Hari','TRI DATA PURE',4000,23500),
    ('Paket Data','Tri','TDBP50','Tri Data Pure 5 GB 30 Hari','TRI DATA PURE',5000,29050),
    ('Paket Data','Tri','TDBP60','Tri Data Pure 6 GB 30 Hari','TRI DATA PURE',6000,34750),
    ('Paket Data','Tri','TDBP70','Tri Data Pure 7 GB 30 Hari','TRI DATA PURE',7000,40500),
    ('Paket Data','Tri','TDBP80','Tri Data Pure 8 GB 30 Hari','TRI DATA PURE',8000,46000),
    ('Paket Data','Tri','TDBP90','Tri Data Pure 9 GB 30 Hari','TRI DATA PURE',9000,51575),

    ('Paket Data','Tri','TDPAON10','Tri Data AON 10GB + Bonus Kuota Lokal','TRI DATA AON NASIONAL',10000,57100),
    ('Paket Data','Tri','TDPAON14','Tri Data AON 14GB + Bonus Kuota Lokal','TRI DATA AON NASIONAL',14000,67475),
    ('Paket Data','Tri','TDPAON2','Tri Data AON 2,5GB + Bonus Kuota Lokal','TRI DATA AON NASIONAL',2500,24200),
    ('Paket Data','Tri','TDPAON3','Tri Data AON 3,5GB + Bonus Kuota Lokal','TRI DATA AON NASIONAL',3500,27250),
    ('Paket Data','Tri','TDPAON33','Tri Data AON 3GB + Kuota Lokal 1GB & Lokal Plus','TRI DATA AON NASIONAL',4000,24250),
    ('Paket Data','Tri','TDPAON34','Tri Data AON 4GB + Kuota Lokal 1GB & Lokal Plus','TRI DATA AON NASIONAL',5000,28250),
    ('Paket Data','Tri','TDPAON40','Tri Data AON 40GB + Bonus Kuota Lokal','TRI DATA AON NASIONAL',40000,109350),
    ('Paket Data','Tri','TDPAON9','Tri Data AON 9GB + Bonus Kuota Lokal','TRI DATA AON NASIONAL',9000,47125),

    ('Paket Data','Tri','TDPH100','Data Happy 100GB 28 Hari','TRI DATA HAPPY PROMO',100000,105400),
    ('Paket Data','Tri','TDPH109','Data Happy 9GB 10 Hari','TRI DATA HAPPY HARIAN PROMO',9000,27225),
    ('Paket Data','Tri','TDPH11','Data Happy 1,5GB 1 Hari','TRI DATA HAPPY HARIAN PROMO',1500,7670),
    ('Paket Data','Tri','TDPH110','Data Happy 10gb 1 Hari','TRI DATA HAPPY HARIAN PROMO',10000,8675),
    ('Paket Data','Tri','TDPH12','Data Happy 2,5GB 1 Hari','TRI DATA HAPPY HARIAN PROMO',2500,7680),
    ('Paket Data','Tri','TDPH13','Data Happy 3GB 1 Hari','TRI DATA HAPPY HARIAN PROMO',3000,7690),
    ('Paket Data','Tri','TDPH1410','Data Happy 10GB 14 Hari','TRI DATA HAPPY HARIAN PROMO',10000,27100),
    ('Paket Data','Tri','TDPH16','Data Happy 6GB 1 Hari','TRI DATA HAPPY HARIAN PROMO',6000,7700),
    ('Paket Data','Tri','TDPH25','Data Happy 5GB 2 hari','TRI DATA HAPPY HARIAN PROMO',5000,9985),
    ('Paket Data','Tri','TDPH27','Data Happy 7GB 2 Hari','TRI DATA HAPPY HARIAN PROMO',7000,10100),
    ('Paket Data','Tri','TDPH306','Data Happy 6GB 28 Hari','TRI DATA HAPPY PROMO',6000,31300),
    ('Paket Data','Tri','TDPH307','Data Happy 7GB 30 Hari','TRI DATA HAPPY PROMO',7000,31450),
    ('Paket Data','Tri','TDPH31','Data Happy 1GB 3 Hari','TRI DATA HAPPY HARIAN PROMO',1000,7200),
    ('Paket Data','Tri','TDPH310','Data Happy 10GB 28 Hari','TRI DATA HAPPY PROMO',10000,34950),
    ('Paket Data','Tri','TDPH311','Data Happy 11GB 28 Hari','TRI DATA HAPPY PROMO',11000,40700),
    ('Paket Data','Tri','TDPH313','Data Happy 13GB 28 Hari','TRI DATA HAPPY PROMO',13000,45320),
    ('Paket Data','Tri','TDPH314','Data Happy 14GB 28 Hari','TRI DATA HAPPY PROMO',14000,45450),
    ('Paket Data','Tri','TDPH315','Data Happy 15GB 28 Hari','TRI DATA HAPPY PROMO',15000,45600),
    ('Paket Data','Tri','TDPH323','Data Happy 23GB 14 Hari','TRI DATA HAPPY HARIAN PROMO',23000,46800),
    ('Paket Data','Tri','TDPH325','Data Happy 25GB 28 Hari','TRI DATA HAPPY PROMO',25000,55100),
    ('Paket Data','Tri','TDPH326','Data Happy 26GB 28 Hari','TRI DATA HAPPY PROMO',26000,62525),
    ('Paket Data','Tri','TDPH33','Data Happy 3,5GB 3 Hari','TRI DATA HAPPY HARIAN PROMO',3500,12700),
    ('Paket Data','Tri','TDPH330','Data Happy 30GB 28 Hari','TRI DATA HAPPY PROMO',30000,62750),
    ('Paket Data','Tri','TDPH335','Data Happy 35GB 28 Hari','TRI DATA HAPPY PROMO',35000,75250),
    ('Paket Data','Tri','TDPH340','Data Happy 40GB 28 Hari','TRI DATA HAPPY PROMO',40000,75275),
    ('Paket Data','Tri','TDPH350','Data Happy 50GB 28 Hari','TRI DATA HAPPY PROMO',50000,82500),
    ('Paket Data','Tri','TDPH355','Data Happy 55GB 28 Hari','tri',55000,87350),
    ('Paket Data','Tri','TDPH36','Data Happy 6GB 3 Hari','TRI DATA HAPPY HARIAN PROMO',6000,12950),
    ('Paket Data','Tri','TDPH365','Data Happy 65GB 28 Hari','TRI DATA HAPPY PROMO',65000,87400),
    ('Paket Data','Tri','TDPH510','Data Happy 10GB 5 Hari','TRI DATA HAPPY HARIAN PROMO',10000,17675),
    ('Paket Data','Tri','TDPH511','Data Happy 11GB 5 Hari','TRI DATA HAPPY HARIAN PROMO',11000,20590),
    ('Paket Data','Tri','TDPH52','Data Happy 2GB 5 Hari','TRI DATA HAPPY HARIAN PROMO',2000,17300),
    ('Paket Data','Tri','TDPH57','Data Happy 7GB 5 Hari','TRI DATA HAPPY HARIAN PROMO',7000,17615),
    ('Paket Data','Tri','TDPH58','Data Happy 8GB 5 Hari','TRI DATA HAPPY HARIAN PROMO',8000,17650),
    ('Paket Data','Tri','TDPH71','Data Happy 1,5GB 7 Hari','TRI DATA HAPPY HARIAN PROMO',1500,10605),
    ('Paket Data','Tri','TDPH710','Data Happy 10GB 7 Hari','TRI DATA HAPPY HARIAN PROMO',10000,23400),
    ('Paket Data','Tri','TDPH715','Data Happy 15GB 7 Hari','tri',15000,24790),
    ('Paket Data','Tri','TDPH717','Data Happy 17GB 7 Hari','TRI DATA HAPPY HARIAN PROMO',17000,28525),
    ('Paket Data','Tri','TDPH80','Data Happy 80GB 28 Hari','TRI DATA HAPPY PROMO',80000,105375),

    ('Paket Data','Tri','TDUH12','New Tri Ibadah 50GB + 1GB 12 Hari','NEW TRI IBADAH HAJI',51000,251000),
    ('Paket Data','Tri','TDUH15','New Tri Ibadah 70GB + 1GB 15 Hari','NEW TRI IBADAH HAJI',71000,311500),
    ('Paket Data','Tri','TDUH30','New Tri Ibadah 25GB + 1GB 30 Hari','NEW TRI IBADAH HAJI',26000,511000),
    ('Paket Data','Tri','TDUH45','New Tri Ibadah 30GB + 1GB 45 Hari','NEW TRI IBADAH HAJI',31000,576000),
    ('Paket Data','Tri','TDUM14','Tri Data Ibadah 14GB + 1GB 15 Hari','TRI IBADAH HAJI',15000,281000),
    ('Paket Data','Tri','TDUM19','Tri Data Ibadah 19GB + 1GB 30 Hari','TRI IBADAH HAJI',20000,521000),
    ('Paket Data','Tri','TDUM24','Tri Data Ibadah 24GB + 1GB 45 Hari','TRI IBADAH HAJI',25000,601000),
    ('Paket Data','Tri','TDUM6','Tri Data Ibadah 6GB + 1GB 12 Hari','TRI IBADAH HAJI',7000,201000),

    ('Paket Data','Tri','TDVHPA2','Tri Data Voucher Happy 5 GB 1 Hari','TRI VOUCHER HAPPY 1 HARI',5000,9300),
    ('Paket Data','Tri','TDVHPC1','Tri Voucher Happy 3.5 GB 5 Hari','TRI VOUCHER HAPPY 5 HARI',3500,13850),
    ('Paket Data','Tri','TDVHPC2','Tri Voucher Happy 6 GB 5 Hari','TRI VOUCHER HAPPY 5 HARI',6000,18125),
    ('Paket Data','Tri','TDVHPE1','Tri Voucher Happy 7 GB 30 Hari','TRI VOUCHER HAPPY 30 HARI',7000,27825),
    ('Paket Data','Tri','TDVHPE2','Tri Voucher Happy 11 GB 30 Hari','TRI VOUCHER HAPPY 30 HARI',11000,41025),
    ('Paket Data','Tri','TDVHPE3','Tri Voucher Happy 14 GB 30 Hari','TRI VOUCHER HAPPY 30 HARI',14000,53350),
    ('Paket Data','Tri','TDVHPE4','Tri Voucher Happy 18 GB 30 Hari','TRI VOUCHER HAPPY 30 HARI',18000,56450),
    ('Paket Data','Tri','TDVHPE5','Tri Voucher Happy 30 GB 30 Hari','TRI VOUCHER HAPPY 30 HARI',30000,65750),
    ('Paket Data','Tri','TDVHPE6','Tri Voucher Happy 42 GB 30 Hari','TRI VOUCHER HAPPY 30 HARI',42000,79650),
    ('Paket Data','Tri','TDVHPE7','Tri Voucher Happy 55 GB 30 Hari','TRI VOUCHER HAPPY 30 HARI',55000,106850),
    ('Paket Data','Tri','TDVHPE8','Tri Voucher Happy 100 GB 30 Hari','TRI VOUCHER HAPPY 30 HARI',100000,131100),

    ('Paket Data','Tri','TVACWJ12','Voucher Data AON 12GB','TRI DATA VOUCHER AON CWJ',12000,58050),
    ('Paket Data','Tri','TVACWJ2','Voucher Data AON 2,5GB','TRI DATA VOUCHER AON CWJ',2500,18300),
    ('Paket Data','Tri','TVACWJ3','Voucher Data AON 3,5GB','TRI DATA VOUCHER AON CWJ',3500,22700),
    ('Paket Data','Tri','TVACWJ40','Voucher Data AON 40GB','TRI DATA VOUCHER AON CWJ',40000,108450),
    ('Paket Data','Tri','TVACWJ6','Voucher Data AON 6GB','TRI DATA VOUCHER AON CWJ',6000,31150),
    ('Paket Data','Tri','TVACWJ9','Voucher Data AON 9GB','TRI DATA VOUCHER AON CWJ',9000,44450),

    ('Paket Data','Tri','TVAEJBN12','Voucher Data AON 12GB','TRI DATA VOUCHER AON EJBN',12000,59350),
    ('Paket Data','Tri','TVAEJBN2','Voucher Data AON 2,5GB','TRI DATA VOUCHER AON EJBN',2500,25350),
    ('Paket Data','Tri','TVAEJBN3','Voucher Data AON 3,5GB','TRI DATA VOUCHER AON EJBN',3500,26350),
    ('Paket Data','Tri','TVAEJBN6','Voucher Data AON 6GB','TRI DATA VOUCHER AON EJBN',6000,32750),
    ('Paket Data','Tri','TVAEJBN9','Voucher Data AON 9GB','TRI DATA VOUCHER AON EJBN',9000,47950),

    ('Paket Data','Tri','TVAJBS12','Voucher Data AON 12GB','TRI DATA VOUCHER AON JABO',12000,59850),
    ('Paket Data','Tri','TVAJBS3','Voucher Data AON 3,5GB','TRI DATA VOUCHER AON JABO',3500,26550),
    ('Paket Data','Tri','TVAJBS40','Voucher Data AON 40GB','TRI DATA VOUCHER AON JABO',40000,112420),
    ('Paket Data','Tri','TVAJBS6','Voucher Data AON 6GB','TRI DATA VOUCHER AON JABO',6000,33750),
    ('Paket Data','Tri','TVAJBS9','Voucher Data AON 9GB','TRI DATA VOUCHER AON JABO',9000,49450),

    ('Paket Data','Tri','TVAN12','Voucher Data AON 12GB','TRI DATA VOUCHER AON NASIONAL',12000,59350),
    ('Paket Data','Tri','TVAN2','Voucher Data AON 2,5GB','TRI DATA VOUCHER AON NASIONAL',2500,25350),
    ('Paket Data','Tri','TVAN3','Voucher Data AON 3,5GB','TRI DATA VOUCHER AON NASIONAL',3500,26350),
    ('Paket Data','Tri','TVAN40','Voucher Data AON 40GB','TRI DATA VOUCHER AON NASIONAL',40000,110150),
    ('Paket Data','Tri','TVAN6','Voucher Data AON 6GB','TRI DATA VOUCHER AON NASIONAL',6000,32750),
    ('Paket Data','Tri','TVAN9','Voucher Data AON 9GB','TRI DATA VOUCHER AON NASIONAL',9000,47950),

    ('Paket Data','Tri','TVDP6','Voucher Tri 6 GB + Unlimited 30 Hari','TRI VOUCHER DATA PURE',6000,60385),
    ('Paket Data','Tri','TVDP8','Voucher Tri Data 8 GB 30 Hari','TRI VOUCHER DATA PURE',8000,106510),
    ('Paket Data','Tri','TVH4K5','Tri Voucher Happy 4.5 GB 30 Hari','TRI VOUCHER HAPPY 30 HARI',4500,21020),
    ('Paket Data','Tri','TVHB1','Tri Data Voucher Happy 1 GB 3 Hari','TRI VOUCHER HAPPY 3 HARI',1000,7950),
    ('Paket Data','Tri','TVHG2','Tri Voucher Happy 2 GB 5 Hari','TRI VOUCHER HAPPY 5 HARI',2000,12903),
    ('Paket Data','Tri','TVHJ9','Tri Voucher Happy 9 GB 10 Hari','TRI VOUCHER HAPPY 7 HARI',9000,27700),
    ('Paket Data','Tri','TVHU52','Tri Voucher Happy Unlimited 52 GB 30 Hari','TRI VOUCHER HAPPY 30 HARI',52000,67800),

    ('Paket Data','Tri','VTHHB100','VCR Tri Happy 100GB + 20GB Lokal 30 Hari','VOUCHER TRI HAPPY BULANAN',120000,196850),
    ('Paket Data','Tri','VTHHB11','VCR Tri Happy 11GB + 6GB Lokal 30 Hari','VOUCHER TRI HAPPY BULANAN',17000,46250),
    ('Paket Data','Tri','VTHHB14','VCR Tri Happy 14GB + 10GB Lokal 30 Hari','VOUCHER TRI HAPPY BULANAN',24000,60250),
    ('Paket Data','Tri','VTHHB18','VCR Tri Happy 18GB + 10GB Lokal 30 Hari','VOUCHER TRI HAPPY BULANAN',28000,63350),
    ('Paket Data','Tri','VTHHB30','VCR Tri Happy 30GB + 10GB Lokal 30 Hari','VOUCHER TRI HAPPY BULANAN',40000,73350),
    ('Paket Data','Tri','VTHHB42','VCR Tri Happy 42GB + 8GB Lokal 30 Hari','VOUCHER TRI HAPPY BULANAN',50000,92050),
    ('Paket Data','Tri','VTHHB55','VCR Tri Happy 55GB + 15GB Lokal 30 Hari','VOUCHER TRI HAPPY BULANAN',70000,113800),
    ('Paket Data','Tri','VTHHB7','VCR Tri Happy 7GB + 4GB Lokal 30 Hari','VOUCHER TRI HAPPY BULANAN',11000,31075),
    ('Paket Data','Tri','VTHHM109','VCR Tri Happy 9GB + 4GB Lokal 10 Hari','VOUCHER TRI HAPPY MINI',13000,31200),
    ('Paket Data','Tri','VTHHM55','VCR Tri Happy 3,5GB + 2,5GB Lokal 5 Hari','VOUCHER TRI HAPPY MINI',6000,16520)
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
