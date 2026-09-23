-- Seed produk pulsa Smartfren dari daftar provider.
-- Aman dijalankan berulang: produk lama Smartfren generik dinonaktifkan, SKU baru di-upsert.

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
  AND lower(trim(k.nama)) = 'pulsa'
  AND lower(trim(b.nama)) = 'smartfren'
  AND p.sku LIKE 'PK-PULSA-SMARTFREN-%';

WITH seed(category_name, brand_name, sku, product_name, group_name, nominal, harga) AS (
  VALUES
    ('Pulsa','Smartfren','FRP10','Smartfren Reguler Standar 10.000','SMARTFREN REGULER STANDAR',10000,11005),
    ('Pulsa','Smartfren','FRP100','Smartfren Reguler Standar 100.000','SMARTFREN REGULER STANDAR',100000,100975),
    ('Pulsa','Smartfren','FRP1000','Smartfren Reguler Standar 1.000.000','SMARTFREN REGULER STANDAR',1000000,1002000),
    ('Pulsa','Smartfren','FRP12','Smartfren Reguler Standar 12.000','SMARTFREN REGULER STANDAR',12000,13005),
    ('Pulsa','Smartfren','FRP125','Smartfren Reguler Standar 125.000','SMARTFREN REGULER STANDAR',125000,126100),
    ('Pulsa','Smartfren','FRP15','Smartfren Reguler Standar 15.000','SMARTFREN REGULER STANDAR',15000,16005),
    ('Pulsa','Smartfren','FRP150','Smartfren Reguler Standar 150.000','SMARTFREN REGULER STANDAR',150000,151100),
    ('Pulsa','Smartfren','FRP20','Smartfren Reguler Standar 20.000','SMARTFREN REGULER STANDAR',20000,21005),
    ('Pulsa','Smartfren','FRP200','Smartfren Reguler Standar 200.000','SMARTFREN REGULER STANDAR',200000,201100),
    ('Pulsa','Smartfren','FRP25','Smartfren Reguler Standar 25.000','SMARTFREN REGULER STANDAR',25000,26005),
    ('Pulsa','Smartfren','FRP250','Smartfren Reguler Standar 250.000','SMARTFREN REGULER STANDAR',250000,251200),
    ('Pulsa','Smartfren','FRP30','Smartfren Reguler Standar 30.000','SMARTFREN REGULER STANDAR',30000,31005),
    ('Pulsa','Smartfren','FRP300','Smartfren Reguler Standar 300.000','SMARTFREN REGULER STANDAR',300000,300950),
    ('Pulsa','Smartfren','FRP35','Smartfren Reguler Standar 35.000','SMARTFREN REGULER STANDAR',35000,36005),
    ('Pulsa','Smartfren','FRP40','Smartfren Reguler Standar 40.000','SMARTFREN REGULER STANDAR',40000,41005),
    ('Pulsa','Smartfren','FRP45','Smartfren Reguler Standar 45.000','SMARTFREN REGULER STANDAR',45000,46005),
    ('Pulsa','Smartfren','FRP5','Smartfren Reguler Standar 5.000','SMARTFREN REGULER STANDAR',5000,6005),
    ('Pulsa','Smartfren','FRP50','Smartfren Reguler Standar 50.000','SMARTFREN REGULER STANDAR',50000,51005),
    ('Pulsa','Smartfren','FRP500','Smartfren Reguler Standar 500.000','SMARTFREN REGULER STANDAR',500000,500950),
    ('Pulsa','Smartfren','FRP55','Smartfren Reguler Standar 55.000','SMARTFREN REGULER STANDAR',55000,56005),
    ('Pulsa','Smartfren','FRP60','Smartfren Reguler Standar 60.000','SMARTFREN REGULER STANDAR',60000,61005),
    ('Pulsa','Smartfren','FRP65','Smartfren Reguler Standar 65.000','SMARTFREN REGULER STANDAR',65000,66005),
    ('Pulsa','Smartfren','FRP70','Smartfren Reguler Standar 70.000','SMARTFREN REGULER STANDAR',70000,71005),
    ('Pulsa','Smartfren','FRP75','Smartfren Reguler Standar 75.000','SMARTFREN REGULER STANDAR',75000,76005),

    ('Pulsa','Smartfren','FVR12','SMARTFREN REGULER VIP 12K','SMARTFREN REGULER VIP',12000,13015),
    ('Pulsa','Smartfren','FVR125','SMARTFREN REGULER VIP 125K','SMARTFREN REGULER VIP',125000,126150),
    ('Pulsa','Smartfren','FVR15','SMARTFREN REGULER VIP 15K','SMARTFREN REGULER VIP',15000,16010),
    ('Pulsa','Smartfren','FVR150','SMARTFREN REGULER VIP 150K','SMARTFREN REGULER VIP',150000,151150),
    ('Pulsa','Smartfren','FVR20','SMARTFREN REGULER VIP 20K','SMARTFREN REGULER VIP',20000,21010),
    ('Pulsa','Smartfren','FVR200','SMARTFREN REGULER VIP 200K','SMARTFREN REGULER VIP',200000,201200),
    ('Pulsa','Smartfren','FVR25','SMARTFREN REGULER VIP 25K','SMARTFREN REGULER VIP',25000,26010),
    ('Pulsa','Smartfren','FVR30','SMARTFREN REGULER VIP 30K','SMARTFREN REGULER VIP',30000,31010),
    ('Pulsa','Smartfren','FVR300','SMARTFREN REGULER VIP 300K','SMARTFREN REGULER VIP',300000,301300),
    ('Pulsa','Smartfren','FVR35','SMARTFREN REGULER VIP 35K','SMARTFREN REGULER VIP',35000,36010),
    ('Pulsa','Smartfren','FVR40','SMARTFREN REGULER VIP 40K','SMARTFREN REGULER VIP',40000,41010),
    ('Pulsa','Smartfren','FVR45','SMARTFREN REGULER VIP 45K','SMARTFREN REGULER VIP',45000,46010),
    ('Pulsa','Smartfren','FVR5','SMARTFREN REGULER VIP 5K','SMARTFREN REGULER VIP',5000,6010),
    ('Pulsa','Smartfren','FVR50','SMARTFREN REGULER VIP 50K','SMARTFREN REGULER VIP',50000,51010),
    ('Pulsa','Smartfren','FVR500','SMARTFREN REGULER VIP 500K','SMARTFREN REGULER VIP',500000,501500),
    ('Pulsa','Smartfren','FVR55','SMARTFREN REGULER VIP 55K','SMARTFREN REGULER VIP',55000,56010),
    ('Pulsa','Smartfren','FVR60','SMARTFREN REGULER VIP 60K','SMARTFREN REGULER VIP',60000,61010),
    ('Pulsa','Smartfren','FVR65','SMARTFREN REGULER VIP 65K','SMARTFREN REGULER VIP',65000,66010),
    ('Pulsa','Smartfren','FVR70','SMARTFREN REGULER VIP 70K','SMARTFREN REGULER VIP',70000,71010),
    ('Pulsa','Smartfren','FVR75','SMARTFREN REGULER VIP 75K','SMARTFREN REGULER VIP',75000,76010),
    ('Pulsa','Smartfren','FVR80','SMARTFREN REGULER VIP 80K','SMARTFREN REGULER VIP',80000,81010),
    ('Pulsa','Smartfren','FVR85','SMARTFREN REGULER VIP 85K','SMARTFREN REGULER VIP',85000,86010),
    ('Pulsa','Smartfren','FVR90','SMARTFREN REGULER VIP 90K','SMARTFREN REGULER VIP',90000,91010),
    ('Pulsa','Smartfren','FVR95','SMARTFREN REGULER VIP 95K','SMARTFREN REGULER VIP',95000,96010)
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
