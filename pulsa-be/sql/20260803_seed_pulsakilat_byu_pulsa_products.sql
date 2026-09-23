-- Seed produk pulsa by.U dari daftar provider.
-- Aman dijalankan berulang: produk lama by.U generik dinonaktifkan, SKU baru di-upsert.

INSERT INTO public.brand (nama, aktif, dibuat_pada, diubah_pada)
SELECT 'by.U', true, now(), now()
WHERE NOT EXISTS (
  SELECT 1 FROM public.brand WHERE lower(regexp_replace(trim(nama), '[^a-z0-9]+', '', 'g')) = 'byu'
);

UPDATE public.produk p
SET aktif = false, diubah_pada = now()
FROM public.kategori k, public.brand b
WHERE p.kategori_id = k.id
  AND p.brand_id = b.id
  AND lower(trim(k.nama)) = 'pulsa'
  AND lower(regexp_replace(trim(b.nama), '[^a-z0-9]+', '', 'g')) = 'byu'
  AND p.sku LIKE 'PK-PULSA-BYU-%';

WITH seed(category_name, brand_name, sku, product_name, group_name, nominal, harga) AS (
  VALUES
    ('Pulsa','by.U','BRP10','Pulsa Isi Ulang By.U 10.000','BY.U REGULER PROMO',10000,11150),
    ('Pulsa','by.U','BRP100','Pulsa Isi Ulang By.U 100.000','BY.U REGULER PROMO',100000,98300),
    ('Pulsa','by.U','BRP15','Pulsa Isi Ulang By.U 15.000','BY.U REGULER PROMO',15000,15900),
    ('Pulsa','by.U','BRP150','Pulsa Isi Ulang By.U 150.000','BY.U REGULER PROMO',150000,149900),
    ('Pulsa','by.U','BRP20','Pulsa Isi Ulang By.U 20.000','BY.U REGULER PROMO',20000,20860),
    ('Pulsa','by.U','BRP200','Pulsa Isi Ulang By.U 200.000','BY.U REGULER PROMO',200000,195425),
    ('Pulsa','by.U','BRP25','Pulsa Isi Ulang By.U 25.000','BY.U REGULER PROMO',25000,25635),
    ('Pulsa','by.U','BRP30','Pulsa Isi Ulang By.U 30.000','BY.U REGULER PROMO',30000,30690),
    ('Pulsa','by.U','BRP35','Pulsa Isi Ulang By.U 35.000','BY.U REGULER PROMO',35000,35530),
    ('Pulsa','by.U','BRP40','Pulsa Isi Ulang By.U 40.000','BY.U REGULER PROMO',40000,40560),
    ('Pulsa','by.U','BRP45','Pulsa Isi Ulang By.U 45.000','BY.U REGULER PROMO',45000,45385),
    ('Pulsa','by.U','BRP5','Pulsa Isi Ulang By.U 5.000','BY.U REGULER PROMO',5000,6200),
    ('Pulsa','by.U','BRP50','Pulsa Isi Ulang By.U 50.000','BY.U REGULER PROMO',50000,50250),
    ('Pulsa','by.U','BRP55','Pulsa Isi Ulang By.U 55.000','BY.U REGULER PROMO',55000,55255),
    ('Pulsa','by.U','BRP60','Pulsa Isi Ulang By.U 60.000','BY.U REGULER PROMO',60000,60175),
    ('Pulsa','by.U','BRP65','Pulsa Isi Ulang By.U 65.000','BY.U REGULER PROMO',65000,65110),
    ('Pulsa','by.U','BRP70','Pulsa Isi Ulang By.U 70.000','BY.U REGULER PROMO',70000,70050),
    ('Pulsa','by.U','BRP75','Pulsa Isi Ulang By.U 75.000','BY.U REGULER PROMO',75000,74180),
    ('Pulsa','by.U','BRP80','Pulsa Isi Ulang By.U 80.000','BY.U REGULER PROMO',80000,79895),
    ('Pulsa','by.U','BRP85','Pulsa Isi Ulang By.U 85.000','BY.U REGULER PROMO',85000,84820),
    ('Pulsa','by.U','BRP90','Pulsa Isi Ulang By.U 90.000','BY.U REGULER PROMO',90000,89750),
    ('Pulsa','by.U','BRP95','Pulsa Isi Ulang By.U 95.000','BY.U REGULER PROMO',95000,94675),

    ('Pulsa','by.U','BRS10','Pulsa Isi Ulang By.U 10.000 Detikan','BY.U REGULER VIP',10000,11195),
    ('Pulsa','by.U','BRS100','Pulsa Isi Ulang By.U 100.000 Detikan','BY.U REGULER VIP',100000,99900),
    ('Pulsa','by.U','BRS15','Pulsa Isi Ulang By.U 15.000 Detikan','BY.U REGULER VIP',15000,16055),
    ('Pulsa','by.U','BRS150','Pulsa Isi Ulang By.U 150.000 Detikan','BY.U REGULER VIP',150000,149500),
    ('Pulsa','by.U','BRS20','Pulsa Isi Ulang By.U 20.000 Detikan','BY.U REGULER VIP',20000,20895),
    ('Pulsa','by.U','BRS200','Pulsa Isi Ulang By.U 200.000 Detikan','BY.U REGULER VIP',200000,197000),
    ('Pulsa','by.U','BRS25','Pulsa Isi Ulang By.U 25.000 Detikan','BY.U REGULER VIP',25000,25805),
    ('Pulsa','by.U','BRS30','Pulsa Isi Ulang By.U 30.000 Detikan','BY.U REGULER VIP',30000,30805),
    ('Pulsa','by.U','BRS35','Pulsa Isi Ulang By.U 35.000 Detikan','BY.U REGULER VIP',35000,35755),
    ('Pulsa','by.U','BRS40','Pulsa Isi Ulang By.U 40.000 Detikan','BY.U REGULER VIP',40000,40705),
    ('Pulsa','by.U','BRS45','Pulsa Isi Ulang By.U 45.000 Detikan','BY.U REGULER VIP',45000,45645),
    ('Pulsa','by.U','BRS5','Pulsa Isi Ulang By.U 5.000 Detikan','BY.U REGULER VIP',5000,6231),
    ('Pulsa','by.U','BRS50','Pulsa Isi Ulang By.U 50.000 Detikan','BY.U REGULER VIP',50000,50575),
    ('Pulsa','by.U','BRS55','Pulsa Isi Ulang By.U 55.000 Detikan','BY.U REGULER VIP',55000,55630),
    ('Pulsa','by.U','BRS60','Pulsa Isi Ulang By.U 60.000 Detikan','BY.U REGULER VIP',60000,60530),
    ('Pulsa','by.U','BRS65','Pulsa Isi Ulang By.U 65.000 Detikan','BY.U REGULER VIP',65000,65460),
    ('Pulsa','by.U','BRS70','Pulsa Isi Ulang By.U 70.000 Detikan','BY.U REGULER VIP',70000,70430),
    ('Pulsa','by.U','BRS75','Pulsa Isi Ulang By.U 75.000 Detikan','BY.U REGULER VIP',75000,75550),
    ('Pulsa','by.U','BRS80','Pulsa Isi Ulang By.U 80.000 Detikan','BY.U REGULER VIP',80000,80350),
    ('Pulsa','by.U','BRS85','Pulsa Isi Ulang By.U 85.000 Detikan','BY.U REGULER VIP',85000,85270),
    ('Pulsa','by.U','BRS90','Pulsa Isi Ulang By.U 90.000 Detikan','BY.U REGULER VIP',90000,90300),
    ('Pulsa','by.U','BRS95','Pulsa Isi Ulang By.U 95.000 Detikan','BY.U REGULER VIP',95000,95170)
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
