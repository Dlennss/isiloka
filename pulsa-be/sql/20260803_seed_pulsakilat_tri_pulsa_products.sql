-- Seed produk pulsa Tri dari daftar provider.
-- Aman dijalankan berulang: produk lama Tri generik dinonaktifkan, SKU baru di-upsert.

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
  AND lower(trim(k.nama)) = 'pulsa'
  AND lower(trim(b.nama)) = 'tri'
  AND p.sku LIKE 'PK-PULSA-TRI-%';

WITH seed(category_name, brand_name, sku, product_name, group_name, nominal, harga) AS (
  VALUES
    ('Pulsa','Tri','TRP10','Tri Reguler Standar 10.000','TRI REGULER STANDAR',10000,11115),
    ('Pulsa','Tri','TRP100','Tri Reguler Standar 100.000','TRI REGULER STANDAR',100000,94600),
    ('Pulsa','Tri','TRP15','Tri Reguler Standar 15.000','TRI REGULER STANDAR',15000,15620),
    ('Pulsa','Tri','TRP150','Tri Reguler Standar 150.000','TRI REGULER STANDAR',150000,141750),
    ('Pulsa','Tri','TRP200','Tri Reguler Standar 200.000','TRI REGULER STANDAR',200000,190000),
    ('Pulsa','Tri','TRP25','Tri Reguler Standar 25.000','TRI REGULER STANDAR',25000,25535),
    ('Pulsa','Tri','TRP30','Tri Reguler Standar 30.000','TRI REGULER STANDAR',30000,30280),
    ('Pulsa','Tri','TRP300','Tri Reguler Standar 300.000','TRI REGULER STANDAR',300000,293600),
    ('Pulsa','Tri','TRP40','Tri Reguler Standar 40.000','TRI REGULER STANDAR',40000,40150),
    ('Pulsa','Tri','TRP5','Tri Reguler Standar 5.000','TRI REGULER STANDAR',5000,6127),
    ('Pulsa','Tri','TRP50','Tri Reguler Standar 50.000','TRI REGULER STANDAR',50000,48940),
    ('Pulsa','Tri','TRP500','Tri Reguler Standar 500.000','TRI REGULER STANDAR',500000,474000),
    ('Pulsa','Tri','TRP60','Tri Reguler Standar 60.000','TRI REGULER STANDAR',60000,58480),
    ('Pulsa','Tri','TRP70','Tri Reguler Standar 70.000','TRI REGULER STANDAR',70000,68060),
    ('Pulsa','Tri','TRP75','Tri Reguler Standar 75.000','TRI REGULER STANDAR',75000,73300),
    ('Pulsa','Tri','TRP80','Tri Reguler Standar 80.000','TRI REGULER STANDAR',80000,77700),
    ('Pulsa','Tri','TRP90','Tri Reguler Standar 90.000','TRI REGULER STANDAR',90000,88950),

    ('Pulsa','Tri','TRSP10','TRI REGULER PROMO 10K','TRI REGULER PROMO',10000,11000),
    ('Pulsa','Tri','TRSP100','TRI REGULER PROMO 100K','TRI REGULER PROMO',100000,94590),
    ('Pulsa','Tri','TRSP15','TRI REGULER PROMO 15K','TRI REGULER PROMO',15000,15550),
    ('Pulsa','Tri','TRSP150','TRI REGULER PROMO 150K','TRI REGULER PROMO',150000,141725),
    ('Pulsa','Tri','TRSP20','TRI REGULER PROMO 20K','TRI REGULER PROMO',20000,20300),
    ('Pulsa','Tri','TRSP200','TRI REGULER PROMO 200K','TRI REGULER PROMO',200000,189900),
    ('Pulsa','Tri','TRSP25','TRI REGULER PROMO 25K','TRI REGULER PROMO',25000,25200),
    ('Pulsa','Tri','TRSP30','TRI REGULER PROMO 30K','TRI REGULER PROMO',30000,30050),
    ('Pulsa','Tri','TRSP300','TRI REGULER PROMO 300K','TRI REGULER PROMO',300000,293550),
    ('Pulsa','Tri','TRSP40','TRI REGULER PROMO 40K','TRI REGULER PROMO',40000,40140),
    ('Pulsa','Tri','TRSP5','TRI REGULER PROMO 5K','TRI REGULER PROMO',5000,6125),
    ('Pulsa','Tri','TRSP50','TRI REGULER PROMO 50K','TRI REGULER PROMO',50000,48930),
    ('Pulsa','Tri','TRSP500','TRI REGULER PROMO 500K','TRI REGULER PROMO',500000,473900),
    ('Pulsa','Tri','TRSP60','TRI REGULER PROMO 60K','TRI REGULER PROMO',60000,58450),
    ('Pulsa','Tri','TRSP70','TRI REGULER PROMO 70K','TRI REGULER PROMO',70000,68045),
    ('Pulsa','Tri','TRSP75','TRI REGULER PROMO 75K','TRI REGULER PROMO',75000,73250),
    ('Pulsa','Tri','TRSP80','TRI REGULER PROMO 80K','TRI REGULER PROMO',80000,77640),
    ('Pulsa','Tri','TRSP90','TRI REGULER PROMO 90K','TRI REGULER PROMO',90000,88940),

    ('Pulsa','Tri','TRV10','Tri Reguler V-Tri 10.000','TRI REGULER VIP',10000,11205),
    ('Pulsa','Tri','TRV100','Tri Reguler V-Tri 100.000','TRI REGULER VIP',100000,98755),
    ('Pulsa','Tri','TRV15','Tri Reguler V-Tri 15.000','TRI REGULER VIP',15000,15752),
    ('Pulsa','Tri','TRV150','Tri Reguler V-Tri 150.000','TRI REGULER VIP',150000,147410),
    ('Pulsa','Tri','TRV20','Tri Reguler V-Tri 20.000','TRI REGULER VIP',20000,20770),
    ('Pulsa','Tri','TRV200','Tri Reguler V-Tri 200.000','TRI REGULER VIP',200000,196005),
    ('Pulsa','Tri','TRV25','Tri Reguler V-Tri 25.000','TRI REGULER VIP',25000,25705),
    ('Pulsa','Tri','TRV30','Tri Reguler V-Tri 30.000','TRI REGULER VIP',30000,30320),
    ('Pulsa','Tri','TRV300','Tri Reguler V-Tri 300.000','TRI REGULER VIP',300000,297300),
    ('Pulsa','Tri','TRV40','Tri Reguler V-Tri 40.000','TRI REGULER VIP',40000,40200),
    ('Pulsa','Tri','TRV5','Tri Reguler V-Tri 5.000','TRI REGULER VIP',5000,6195),
    ('Pulsa','Tri','TRV50','Tri Reguler V-Tri 50.000','TRI REGULER VIP',50000,50175),
    ('Pulsa','Tri','TRV500','Tri Reguler V-Tri 500.000','TRI REGULER VIP',500000,487735),
    ('Pulsa','Tri','TRV60','Tri Reguler V-Tri 60.000','TRI REGULER VIP',60000,59975),
    ('Pulsa','Tri','TRV70','Tri Reguler V-Tri 70.000','TRI REGULER VIP',70000,69575),
    ('Pulsa','Tri','TRV75','Tri Reguler V-Tri 75.000','TRI REGULER VIP',75000,74425),
    ('Pulsa','Tri','TRV80','Tri Reguler V-Tri 80.000','TRI REGULER VIP',80000,79575),
    ('Pulsa','Tri','TRV90','Tri Reguler V-Tri 90.000','TRI REGULER VIP',90000,89000),

    ('Pulsa','Tri','YHT10','Tri Transfer 10.000','TRI TRANSFER PROMO',10000,10900),
    ('Pulsa','Tri','YHT100','Tri Transfer 100.000','TRI TRANSFER PROMO',100000,94600),
    ('Pulsa','Tri','YHT125','Tri Transfer 125.000','TRI TRANSFER PROMO',125000,117100),
    ('Pulsa','Tri','YHT15','Tri Transfer 15.000','TRI TRANSFER PROMO',15000,15400),
    ('Pulsa','Tri','YHT150','Tri Transfer 150.000','TRI TRANSFER PROMO',150000,139600),
    ('Pulsa','Tri','YHT175','Tri Transfer 175.000','TRI TRANSFER PROMO',175000,162100),
    ('Pulsa','Tri','YHT20','Tri Transfer 20.000','TRI TRANSFER PROMO',20000,19900),
    ('Pulsa','Tri','YHT200','Tri Transfer 200.000','TRI TRANSFER PROMO',200000,184600),
    ('Pulsa','Tri','YHT25','Tri Transfer 25.000','TRI TRANSFER PROMO',25000,24850),
    ('Pulsa','Tri','YHT30','Tri Transfer 30.000','TRI TRANSFER PROMO',30000,29350),
    ('Pulsa','Tri','YHT40','Tri Transfer 40.000','TRI TRANSFER PROMO',40000,38350),
    ('Pulsa','Tri','YHT45','Tri Transfer 45.000','TRI TRANSFER PROMO',45000,42850),
    ('Pulsa','Tri','YHT5','Tri Transfer 5.000','TRI TRANSFER PROMO',5000,6400),
    ('Pulsa','Tri','YHT50','Tri Transfer 50.000','TRI TRANSFER PROMO',50000,47800),
    ('Pulsa','Tri','YHT55','Tri Transfer 55.000','TRI TRANSFER PROMO',55000,52300),
    ('Pulsa','Tri','YHT60','Tri Transfer 60.000','TRI TRANSFER PROMO',60000,56800),
    ('Pulsa','Tri','YHT65','Tri Transfer 65.000','TRI TRANSFER PROMO',65000,61300),
    ('Pulsa','Tri','YHT70','Tri Transfer 70.000','TRI TRANSFER PROMO',70000,65800),
    ('Pulsa','Tri','YHT75','Tri Transfer 75.000','TRI TRANSFER PROMO',75000,70300),
    ('Pulsa','Tri','YHT80','Tri Transfer 80.000','TRI TRANSFER PROMO',80000,74800),
    ('Pulsa','Tri','YHT85','Tri Transfer 85.000','TRI TRANSFER PROMO',85000,79300),
    ('Pulsa','Tri','YHT90','Tri Transfer 90.000','TRI TRANSFER PROMO',90000,83800),
    ('Pulsa','Tri','YHT95','Tri Transfer 95.000','TRI TRANSFER PROMO',95000,88300)
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
