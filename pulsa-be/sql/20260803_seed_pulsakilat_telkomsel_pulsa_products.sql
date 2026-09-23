-- Seed produk pulsa Telkomsel dari daftar provider.
-- Aman dijalankan berulang: produk lama Telkomsel generik dinonaktifkan, SKU baru di-upsert.

INSERT INTO public.brand (nama, aktif, dibuat_pada, diubah_pada)
SELECT 'Telkomsel', true, now(), now()
WHERE NOT EXISTS (
  SELECT 1 FROM public.brand WHERE lower(trim(nama)) = 'telkomsel'
);

UPDATE public.produk p
SET aktif = false, diubah_pada = now()
FROM public.kategori k, public.brand b
WHERE p.kategori_id = k.id
  AND p.brand_id = b.id
  AND lower(trim(k.nama)) = 'pulsa'
  AND lower(trim(b.nama)) = 'telkomsel'
  AND p.sku LIKE 'PK-PULSA-TELKOMSEL-%';

WITH seed(category_name, brand_name, sku, product_name, group_name, nominal, harga) AS (
  VALUES
    ('Pulsa','Telkomsel','SRM10','Telkomsel Reguler VIP 10.000','TELKOMSEL REGULER VIP',10000,11225),
    ('Pulsa','Telkomsel','SRM100','Telkomsel Reguler VIP 100.000','TELKOMSEL REGULER VIP',100000,98200),
    ('Pulsa','Telkomsel','SRM1000','Telkomsel Reguler VIP 1.000.000','TELKOMSEL REGULER VIP',1000000,993450),
    ('Pulsa','Telkomsel','SRM15','Telkomsel Reguler VIP 15.000','TELKOMSEL REGULER VIP',15000,15825),
    ('Pulsa','Telkomsel','SRM150','Telkomsel Reguler VIP 150.000','TELKOMSEL REGULER VIP',150000,149825),
    ('Pulsa','Telkomsel','SRM2','Telkomsel Reguler VIP 2.000','TELKOMSEL REGULER VIP',2000,3925),
    ('Pulsa','Telkomsel','SRM20','Telkomsel Reguler VIP 20.000','TELKOMSEL REGULER VIP',20000,20850),
    ('Pulsa','Telkomsel','SRM200','Telkomsel Reguler VIP 200.000','TELKOMSEL REGULER VIP',200000,197500),
    ('Pulsa','Telkomsel','SRM25','Telkomsel Reguler VIP 25.000','TELKOMSEL REGULER VIP',25000,25650),
    ('Pulsa','Telkomsel','SRM3','Telkomsel Reguler VIP 3.000','TELKOMSEL REGULER VIP',3000,4925),
    ('Pulsa','Telkomsel','SRM30','Telkomsel Reguler VIP 30.000','TELKOMSEL REGULER VIP',30000,30530),
    ('Pulsa','Telkomsel','SRM300','Telkomsel Reguler VIP 300.000','TELKOMSEL REGULER VIP',300000,298150),
    ('Pulsa','Telkomsel','SRM35','Telkomsel Reguler VIP 35.000','TELKOMSEL REGULER VIP',35000,35510),
    ('Pulsa','Telkomsel','SRM4','Telkomsel Reguler VIP 4.000','TELKOMSEL REGULER VIP',4000,5925),
    ('Pulsa','Telkomsel','SRM40','Telkomsel Reguler VIP 40.000','TELKOMSEL REGULER VIP',40000,40200),
    ('Pulsa','Telkomsel','SRM45','Telkomsel Reguler VIP 45.000','TELKOMSEL REGULER VIP',45000,45360),
    ('Pulsa','Telkomsel','SRM5','Telkomsel Reguler VIP 5.000','TELKOMSEL REGULER VIP',5000,6265),
    ('Pulsa','Telkomsel','SRM50','Telkomsel Reguler VIP 50.000','TELKOMSEL REGULER VIP',50000,50275),
    ('Pulsa','Telkomsel','SRM500','Telkomsel Reguler VIP 500.000','TELKOMSEL REGULER VIP',500000,497600),
    ('Pulsa','Telkomsel','SRM55','Telkomsel Reguler VIP 55.000','TELKOMSEL REGULER VIP',55000,55206),
    ('Pulsa','Telkomsel','SRM6','Telkomsel Reguler VIP 6.000','TELKOMSEL REGULER VIP',6000,7925),
    ('Pulsa','Telkomsel','SRM60','Telkomsel Reguler VIP 60.000','TELKOMSEL REGULER VIP',60000,60128),
    ('Pulsa','Telkomsel','SRM65','Telkomsel Reguler VIP 65.000','TELKOMSEL REGULER VIP',65000,65050),
    ('Pulsa','Telkomsel','SRM7','Telkomsel Reguler VIP 7.000','TELKOMSEL REGULER VIP',7000,8925),
    ('Pulsa','Telkomsel','SRM70','Telkomsel Reguler VIP 70.000','TELKOMSEL REGULER VIP',70000,69973),
    ('Pulsa','Telkomsel','SRM75','Telkomsel Reguler VIP 75.000','TELKOMSEL REGULER VIP',75000,74000),
    ('Pulsa','Telkomsel','SRM8','Telkomsel Reguler VIP 8.000','TELKOMSEL REGULER VIP',8000,9925),
    ('Pulsa','Telkomsel','SRM80','Telkomsel Reguler VIP 80.000','TELKOMSEL REGULER VIP',80000,79818),
    ('Pulsa','Telkomsel','SRM85','Telkomsel Reguler VIP 85.000','TELKOMSEL REGULER VIP',85000,84740),
    ('Pulsa','Telkomsel','SRM9','Telkomsel Reguler VIP 9.000','TELKOMSEL REGULER VIP',9000,10925),
    ('Pulsa','Telkomsel','SRM90','Telkomsel Reguler VIP 90.000','TELKOMSEL REGULER VIP',90000,89663),
    ('Pulsa','Telkomsel','SRM95','Telkomsel Reguler VIP 95.000','TELKOMSEL REGULER VIP',95000,94586),

    ('Pulsa','Telkomsel','SRP10','Telkomsel Reguler Promo 10.000','TELKOMSEL REGULER PROMO',10000,11120),
    ('Pulsa','Telkomsel','SRP15','Telkomsel Reguler Promo 15.000','TELKOMSEL REGULER PROMO',15000,15740),
    ('Pulsa','Telkomsel','SRP150','Telkomsel Reguler Promo 150.000','TELKOMSEL REGULER PROMO',150000,146525),
    ('Pulsa','Telkomsel','SRP200','Telkomsel Reguler Promo 200.000','TELKOMSEL REGULER PROMO',200000,194125),
    ('Pulsa','Telkomsel','SRP25','Telkomsel Reguler Promo 25.000','TELKOMSEL REGULER PROMO',25000,25590),
    ('Pulsa','Telkomsel','SRP30','Telkomsel Reguler Promo 30.000','TELKOMSEL REGULER PROMO',30000,30430),
    ('Pulsa','Telkomsel','SRP40','Telkomsel Reguler Promo 40.000','TELKOMSEL REGULER PROMO',40000,40110),
    ('Pulsa','Telkomsel','SRP5','Telkomsel Reguler Promo 5.000','TELKOMSEL REGULER PROMO',5000,6183),
    ('Pulsa','Telkomsel','SRP50','Telkomsel Reguler Promo 50.000','TELKOMSEL REGULER PROMO',50000,50165),
    ('Pulsa','Telkomsel','SRP75','Telkomsel Reguler Promo 75.000','TELKOMSEL REGULER PROMO',75000,73790),

    ('Pulsa','Telkomsel','SRVP10','Telkomsel Reguler 10.000 Standar','TELKOMSEL REGULER STANDAR',10000,11140),
    ('Pulsa','Telkomsel','SRVP100','Telkomsel Reguler 100.000 Standar','TELKOMSEL REGULER STANDAR',100000,97815),
    ('Pulsa','Telkomsel','SRVP15','Telkomsel Reguler 15.000 Standar','TELKOMSEL REGULER STANDAR',15000,15765),
    ('Pulsa','Telkomsel','SRVP150','Telkomsel Reguler 150.000 Standar','TELKOMSEL REGULER STANDAR',150000,149050),
    ('Pulsa','Telkomsel','SRVP20','Telkomsel Reguler 20.000 Standar','TELKOMSEL REGULER STANDAR',20000,20780),
    ('Pulsa','Telkomsel','SRVP200','Telkomsel Reguler 200.000 Standar','TELKOMSEL REGULER STANDAR',200000,194450),
    ('Pulsa','Telkomsel','SRVP25','Telkomsel Reguler 25.000 Standar','TELKOMSEL REGULER STANDAR',25000,25610),
    ('Pulsa','Telkomsel','SRVP30','Telkomsel Reguler 30.000 Standar','TELKOMSEL REGULER STANDAR',30000,30470),
    ('Pulsa','Telkomsel','SRVP35','Telkomsel Reguler 35.000 Standar','TELKOMSEL REGULER STANDAR',35000,35457),
    ('Pulsa','Telkomsel','SRVP40','Telkomsel Reguler 40.000 Standar','TELKOMSEL REGULER STANDAR',40000,40130),
    ('Pulsa','Telkomsel','SRVP45','Telkomsel Reguler 45.000 Standar','TELKOMSEL REGULER STANDAR',45000,45305),
    ('Pulsa','Telkomsel','SRVP5','Telkomsel Reguler 5.000 Standar','TELKOMSEL REGULER STANDAR',5000,6190),
    ('Pulsa','Telkomsel','SRVP50','Telkomsel Reguler 50.000 Standar','TELKOMSEL REGULER STANDAR',50000,50185),
    ('Pulsa','Telkomsel','SRVP55','Telkomsel Reguler 55.000 Standar','TELKOMSEL REGULER STANDAR',55000,55150),
    ('Pulsa','Telkomsel','SRVP60','Telkomsel Reguler 60.000 Standar','TELKOMSEL REGULER STANDAR',60000,60070),
    ('Pulsa','Telkomsel','SRVP65','Telkomsel Reguler 65.000 Standar','TELKOMSEL REGULER STANDAR',65000,64990),
    ('Pulsa','Telkomsel','SRVP70','Telkomsel Reguler 70.000 Standar','TELKOMSEL REGULER STANDAR',70000,69915),
    ('Pulsa','Telkomsel','SRVP75','Telkomsel Reguler 75.000 Standar','TELKOMSEL REGULER STANDAR',75000,73830),
    ('Pulsa','Telkomsel','SRVP80','Telkomsel Reguler 80.000 Standar','TELKOMSEL REGULER STANDAR',80000,79760),
    ('Pulsa','Telkomsel','SRVP85','Telkomsel Reguler 85.000 Standar','TELKOMSEL REGULER STANDAR',85000,84675),
    ('Pulsa','Telkomsel','SRVP90','Telkomsel Reguler 90.000 Standar','TELKOMSEL REGULER STANDAR',90000,89600),
    ('Pulsa','Telkomsel','SRVP95','Telkomsel Reguler 95.000 Standar','TELKOMSEL REGULER STANDAR',95000,94515),

    ('Pulsa','Telkomsel','YTT10','Telkomsel Transfer 10.000','TELKOMSEL PULSA TRANSFER',10000,12372),
    ('Pulsa','Telkomsel','YTT100','Telkomsel Transfer 100.000','TELKOMSEL PULSA TRANSFER',100000,97023),
    ('Pulsa','Telkomsel','YTT105','Telkomsel Transfer 105.000','TELKOMSEL PULSA TRANSFER',105000,101448),
    ('Pulsa','Telkomsel','YTT110','Telkomsel Transfer 110.000','TELKOMSEL PULSA TRANSFER',110000,105873),
    ('Pulsa','Telkomsel','YTT115','Telkomsel Transfer 115.000','TELKOMSEL PULSA TRANSFER',115000,110298),
    ('Pulsa','Telkomsel','YTT120','Telkomsel Transfer 120.000','TELKOMSEL PULSA TRANSFER',120000,114723),
    ('Pulsa','Telkomsel','YTT125','Telkomsel Transfer 125.000','TELKOMSEL PULSA TRANSFER',125000,119148),
    ('Pulsa','Telkomsel','YTT130','Telkomsel Transfer 130.000','TELKOMSEL PULSA TRANSFER',130000,123573),
    ('Pulsa','Telkomsel','YTT135','Telkomsel Transfer 135.000','TELKOMSEL PULSA TRANSFER',135000,127998),
    ('Pulsa','Telkomsel','YTT140','Telkomsel Transfer 140.000','TELKOMSEL PULSA TRANSFER',140000,132423),
    ('Pulsa','Telkomsel','YTT145','Telkomsel Transfer 145.000','TELKOMSEL PULSA TRANSFER',145000,136848),
    ('Pulsa','Telkomsel','YTT15','Telkomsel Transfer 15.000','TELKOMSEL PULSA TRANSFER',15000,16797),
    ('Pulsa','Telkomsel','YTT150','Telkomsel Transfer 150.000','TELKOMSEL PULSA TRANSFER',150000,141273),
    ('Pulsa','Telkomsel','YTT155','Telkomsel Transfer 155.000','TELKOMSEL PULSA TRANSFER',155000,145698),
    ('Pulsa','Telkomsel','YTT160','Telkomsel Transfer 160.000','TELKOMSEL PULSA TRANSFER',160000,150123),
    ('Pulsa','Telkomsel','YTT165','Telkomsel Transfer 165.000','TELKOMSEL PULSA TRANSFER',165000,154548),
    ('Pulsa','Telkomsel','YTT170','Telkomsel Transfer 170.000','TELKOMSEL PULSA TRANSFER',170000,158973),
    ('Pulsa','Telkomsel','YTT175','Telkomsel Transfer 175.000','TELKOMSEL PULSA TRANSFER',175000,163398),
    ('Pulsa','Telkomsel','YTT180','Telkomsel Transfer 180.000','TELKOMSEL PULSA TRANSFER',180000,167823),
    ('Pulsa','Telkomsel','YTT185','Telkomsel Transfer 185.000','TELKOMSEL PULSA TRANSFER',185000,172248),
    ('Pulsa','Telkomsel','YTT190','Telkomsel Transfer 190.000','TELKOMSEL PULSA TRANSFER',190000,176673),
    ('Pulsa','Telkomsel','YTT195','Telkomsel Transfer 195.000','TELKOMSEL PULSA TRANSFER',195000,181098),
    ('Pulsa','Telkomsel','YTT20','Telkomsel Transfer 20.000','TELKOMSEL PULSA TRANSFER',20000,22240),
    ('Pulsa','Telkomsel','YTT200','Telkomsel Transfer 200.000','TELKOMSEL PULSA TRANSFER',200000,189948),
    ('Pulsa','Telkomsel','YTT25','Telkomsel Transfer 25.000','TELKOMSEL PULSA TRANSFER',25000,26665),
    ('Pulsa','Telkomsel','YTT30','Telkomsel Transfer 30.000','TELKOMSEL PULSA TRANSFER',30000,31090),
    ('Pulsa','Telkomsel','YTT35','Telkomsel Transfer 35.000','TELKOMSEL PULSA TRANSFER',35000,35515),
    ('Pulsa','Telkomsel','YTT40','Telkomsel Transfer 40.000','TELKOMSEL PULSA TRANSFER',40000,39940),
    ('Pulsa','Telkomsel','YTT45','Telkomsel Transfer 45.000','TELKOMSEL PULSA TRANSFER',45000,44365),
    ('Pulsa','Telkomsel','YTT5','Telkomsel Transfer 5.000','TELKOMSEL PULSA TRANSFER',5000,7947),
    ('Pulsa','Telkomsel','YTT50','Telkomsel Transfer 50.000','TELKOMSEL PULSA TRANSFER',50000,49675),
    ('Pulsa','Telkomsel','YTT55','Telkomsel Transfer 55.000','TELKOMSEL PULSA TRANSFER',55000,54100),
    ('Pulsa','Telkomsel','YTT60','Telkomsel Transfer 60.000','TELKOMSEL PULSA TRANSFER',60000,58525),
    ('Pulsa','Telkomsel','YTT65','Telkomsel Transfer 65.000','TELKOMSEL PULSA TRANSFER',65000,62950),
    ('Pulsa','Telkomsel','YTT70','Telkomsel Transfer 70.000','TELKOMSEL PULSA TRANSFER',70000,67375),
    ('Pulsa','Telkomsel','YTT75','Telkomsel Transfer 75.000','TELKOMSEL PULSA TRANSFER',75000,71800),
    ('Pulsa','Telkomsel','YTT80','Telkomsel Transfer 80.000','TELKOMSEL PULSA TRANSFER',80000,76225),
    ('Pulsa','Telkomsel','YTT85','Telkomsel Transfer 85.000','TELKOMSEL PULSA TRANSFER',85000,80650),
    ('Pulsa','Telkomsel','YTT90','Telkomsel Transfer 90.000','TELKOMSEL PULSA TRANSFER',90000,85075),
    ('Pulsa','Telkomsel','YTT95','Telkomsel Transfer 95.000','TELKOMSEL PULSA TRANSFER',95000,89500)
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
