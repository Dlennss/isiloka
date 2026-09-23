-- Seed produk pulsa Indosat dari daftar provider.
-- Aman dijalankan berulang: produk lama Indosat generik dinonaktifkan, SKU baru di-upsert.

INSERT INTO public.brand (nama, aktif, dibuat_pada, diubah_pada)
SELECT 'Indosat', true, now(), now()
WHERE NOT EXISTS (
  SELECT 1 FROM public.brand WHERE lower(trim(nama)) = 'indosat'
);

UPDATE public.produk p
SET aktif = false, diubah_pada = now()
FROM public.kategori k, public.brand b
WHERE p.kategori_id = k.id
  AND p.brand_id = b.id
  AND lower(trim(k.nama)) = 'pulsa'
  AND lower(trim(b.nama)) = 'indosat'
  AND p.sku LIKE 'PK-PULSA-INDOSAT-%';

WITH seed(category_name, brand_name, sku, product_name, group_name, nominal, harga) AS (
  VALUES
    ('Pulsa','Indosat','IRK100','Indosat Reguler Promo 100.000','INDOSAT REGULER PROMO',100000,95365),
    ('Pulsa','Indosat','IRK125','Indosat Reguler Promo 125.000','INDOSAT REGULER PROMO',125000,119975),
    ('Pulsa','Indosat','IRK150','Indosat Reguler Promo 150.000','INDOSAT REGULER PROMO',150000,142175),
    ('Pulsa','Indosat','IRK175','Indosat Reguler Promo 175.000','INDOSAT REGULER PROMO',175000,174025),
    ('Pulsa','Indosat','IRK200','Indosat Reguler Promo 200.000','INDOSAT REGULER PROMO',200000,190695),
    ('Pulsa','Indosat','IRK40','Indosat Reguler Promo 40.000','INDOSAT REGULER PROMO',40000,40615),
    ('Pulsa','Indosat','IRK50','Indosat Reguler Promo 50.000','INDOSAT REGULER PROMO',50000,49540),
    ('Pulsa','Indosat','IRK60','Indosat Reguler Promo 60.000','INDOSAT REGULER PROMO',60000,59050),
    ('Pulsa','Indosat','IRK70','Indosat Reguler Promo 70.000','INDOSAT REGULER PROMO',70000,68620),
    ('Pulsa','Indosat','IRK80','Indosat Reguler Promo 80.000','INDOSAT REGULER PROMO',80000,78200),
    ('Pulsa','Indosat','IRK90','Indosat Reguler Promo 90.000','INDOSAT REGULER PROMO',90000,89520),

    ('Pulsa','Indosat','IRP10','Indosat Reguler VIP 10.000','INDOSAT REGULER VIP',10000,12520),
    ('Pulsa','Indosat','IRP100','Indosat Reguler VIP 100.000','INDOSAT REGULER VIP',100000,95600),
    ('Pulsa','Indosat','IRP1000','Indosat Reguler VIP 1.000.000','INDOSAT REGULER VIP',1000000,987400),
    ('Pulsa','Indosat','IRP12','Indosat Reguler VIP 12.000','INDOSAT REGULER VIP',12000,13850),
    ('Pulsa','Indosat','IRP125','Indosat Reguler VIP 125.000','INDOSAT REGULER VIP',125000,120550),
    ('Pulsa','Indosat','IRP15','Indosat Reguler VIP 15.000','INDOSAT REGULER VIP',15000,16690),
    ('Pulsa','Indosat','IRP150','Indosat Reguler VIP 150.000','INDOSAT REGULER VIP',150000,144100),
    ('Pulsa','Indosat','IRP175','Indosat Reguler VIP 175.000','INDOSAT REGULER VIP',175000,174175),
    ('Pulsa','Indosat','IRP20','Indosat Reguler VIP 20.000','INDOSAT REGULER VIP',20000,21580),
    ('Pulsa','Indosat','IRP200','Indosat Reguler VIP 200.000','INDOSAT REGULER VIP',200000,191345),
    ('Pulsa','Indosat','IRP25','Indosat Reguler VIP 25.000','INDOSAT REGULER VIP',25000,26465),
    ('Pulsa','Indosat','IRP250','Indosat Reguler VIP 250.000','INDOSAT REGULER VIP',250000,248100),
    ('Pulsa','Indosat','IRP30','Indosat Reguler VIP 30.000','INDOSAT REGULER VIP',30000,31355),
    ('Pulsa','Indosat','IRP300','Indosat Reguler VIP 300.000','INDOSAT REGULER VIP',300000,297140),
    ('Pulsa','Indosat','IRP40','Indosat Reguler VIP 40.000','INDOSAT REGULER VIP',40000,40635),
    ('Pulsa','Indosat','IRP400','Indosat Reguler VIP 400.000','INDOSAT REGULER VIP',400000,394940),
    ('Pulsa','Indosat','IRP5','Indosat Reguler VIP 5.000','INDOSAT REGULER VIP',5000,7620),
    ('Pulsa','Indosat','IRP50','Indosat Reguler VIP 50.000','INDOSAT REGULER VIP',50000,49830),
    ('Pulsa','Indosat','IRP500','Indosat Reguler VIP 500.000','INDOSAT REGULER VIP',500000,479880),
    ('Pulsa','Indosat','IRP60','Indosat Reguler VIP 60.000','INDOSAT REGULER VIP',60000,59480),
    ('Pulsa','Indosat','IRP70','Indosat Reguler VIP 70.000','INDOSAT REGULER VIP',70000,68850),
    ('Pulsa','Indosat','IRP80','Indosat Reguler VIP 80.000','INDOSAT REGULER VIP',80000,78350),
    ('Pulsa','Indosat','IRP90','Indosat Reguler VIP 90.000','INDOSAT REGULER VIP',90000,89585),

    ('Pulsa','Indosat','IVR10','INDOSAT REGULER STANDAR 10K','INDOSAT REGULER STANDAR',10000,12500),
    ('Pulsa','Indosat','IVR100','INDOSAT REGULER STANDAR 100K','INDOSAT REGULER STANDAR',100000,95375),
    ('Pulsa','Indosat','IVR12','INDOSAT REGULER STANDAR 12K','INDOSAT REGULER STANDAR',12000,13820),
    ('Pulsa','Indosat','IVR125','INDOSAT REGULER STANDAR 125K','INDOSAT REGULER STANDAR',125000,120000),
    ('Pulsa','Indosat','IVR15','INDOSAT REGULER STANDAR 15K','INDOSAT REGULER STANDAR',15000,16650),
    ('Pulsa','Indosat','IVR150','INDOSAT REGULER STANDAR 150K','INDOSAT REGULER STANDAR',150000,142200),
    ('Pulsa','Indosat','IVR175','INDOSAT REGULER STANDAR 175K','INDOSAT REGULER STANDAR',175000,174100),
    ('Pulsa','Indosat','IVR20','INDOSAT REGULER STANDAR 20K','INDOSAT REGULER STANDAR',20000,21550),
    ('Pulsa','Indosat','IVR200','INDOSAT REGULER STANDAR 200K','INDOSAT REGULER STANDAR',200000,190700),
    ('Pulsa','Indosat','IVR25','INDOSAT REGULER STANDAR 25K','INDOSAT REGULER STANDAR',25000,26150),
    ('Pulsa','Indosat','IVR30','INDOSAT REGULER STANDAR 30K','INDOSAT REGULER STANDAR',30000,31325),
    ('Pulsa','Indosat','IVR40','INDOSAT REGULER STANDAR 40K','INDOSAT REGULER STANDAR',40000,40619),
    ('Pulsa','Indosat','IVR5','INDOSAT REGULER STANDAR 5K','INDOSAT REGULER STANDAR',5000,7607),
    ('Pulsa','Indosat','IVR50','INDOSAT REGULER STANDAR 50K','INDOSAT REGULER STANDAR',50000,49550),
    ('Pulsa','Indosat','IVR60','INDOSAT REGULER STANDAR 60K','INDOSAT REGULER STANDAR',60000,59060),
    ('Pulsa','Indosat','IVR70','INDOSAT REGULER STANDAR 70K','INDOSAT REGULER STANDAR',70000,68630),
    ('Pulsa','Indosat','IVR80','INDOSAT REGULER STANDAR 80K','INDOSAT REGULER STANDAR',80000,78210),
    ('Pulsa','Indosat','IVR90','INDOSAT REGULER STANDAR 90K','INDOSAT REGULER STANDAR',90000,89530),

    ('Pulsa','Indosat','YIT10','Pulsa Transfer Indosat 10K','INDOSAT TRANSFER',10000,11450),
    ('Pulsa','Indosat','YIT100','Pulsa Transfer Indosat 100K','INDOSAT TRANSFER',100000,99800),
    ('Pulsa','Indosat','YIT15','Pulsa Transfer Indosat 15K','INDOSAT TRANSFER',15000,16200),
    ('Pulsa','Indosat','YIT150','Pulsa Transfer Indosat 150K','INDOSAT TRANSFER',150000,147300),
    ('Pulsa','Indosat','YIT20','Pulsa Transfer Indosat 20K','INDOSAT TRANSFER',20000,20950),
    ('Pulsa','Indosat','YIT200','Pulsa Transfer Indosat 200K','INDOSAT TRANSFER',200000,194800),
    ('Pulsa','Indosat','YIT25','Pulsa Transfer Indosat 25K','INDOSAT TRANSFER',25000,26175),
    ('Pulsa','Indosat','YIT30','Pulsa Transfer Indosat 30K','INDOSAT TRANSFER',30000,30925),
    ('Pulsa','Indosat','YIT35','Pulsa Transfer Indosat 35K','INDOSAT TRANSFER',35000,35675),
    ('Pulsa','Indosat','YIT40','Pulsa Transfer Indosat 40K','INDOSAT TRANSFER',40000,40425),
    ('Pulsa','Indosat','YIT45','Pulsa Transfer Indosat 45K','INDOSAT TRANSFER',45000,45175),
    ('Pulsa','Indosat','YIT5','Pulsa Transfer Indosat 5K','INDOSAT TRANSFER',5000,6700),
    ('Pulsa','Indosat','YIT50','Pulsa Transfer Indosat 50K','INDOSAT TRANSFER',50000,50400),
    ('Pulsa','Indosat','YIT55','Pulsa Transfer Indosat 55K','INDOSAT TRANSFER',55000,55150),
    ('Pulsa','Indosat','YIT60','Pulsa Transfer Indosat 60K','INDOSAT TRANSFER',60000,59900),
    ('Pulsa','Indosat','YIT65','Pulsa Transfer Indosat 65K','INDOSAT TRANSFER',65000,64650),
    ('Pulsa','Indosat','YIT70','Pulsa Transfer Indosat 70K','INDOSAT TRANSFER',70000,69400),
    ('Pulsa','Indosat','YIT75','Pulsa Transfer Indosat 75K','INDOSAT TRANSFER',75000,74150),
    ('Pulsa','Indosat','YIT80','Pulsa Transfer Indosat 80K','INDOSAT TRANSFER',80000,78900),
    ('Pulsa','Indosat','YIT85','Pulsa Transfer Indosat 85K','INDOSAT TRANSFER',85000,83650),
    ('Pulsa','Indosat','YIT90','Pulsa Transfer Indosat 90K','INDOSAT TRANSFER',90000,88400),
    ('Pulsa','Indosat','YIT95','Pulsa Transfer Indosat 95K','INDOSAT TRANSFER',95000,93150)
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
