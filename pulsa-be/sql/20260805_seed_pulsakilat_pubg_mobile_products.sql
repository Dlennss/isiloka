-- Seed produk top up game PUBG Mobile.
-- Aman dijalankan berulang: brand, produk, dan harga di-upsert berdasarkan SKU.

INSERT INTO public.kategori_fee_app
  (kategori_id, fee_master, fee_agent, fee_user, fee_non_user, aktif, created_at, updated_at)
SELECT k.id, 1000, 1000, 1000, 1500, true, now(), now()
FROM public.kategori k
LEFT JOIN public.kategori_fee_app existing ON existing.kategori_id = k.id
WHERE lower(trim(k.nama)) = 'game'
  AND existing.id IS NULL;

INSERT INTO public.brand (nama, aktif, dibuat_pada, diubah_pada)
SELECT 'PUBG Mobile', true, now(), now()
WHERE NOT EXISTS (
  SELECT 1 FROM public.brand b WHERE lower(trim(b.nama)) = lower(trim('PUBG Mobile'))
);

WITH seed(category_name, brand_name, sku, product_name, group_name, nominal, harga) AS (
  VALUES
    ('Game','PUBG Mobile','GPM60','PUBG Mobile 60 UC','PUBG Mobile UC',60,17120),
    ('Game','PUBG Mobile','GPM120','PUBG Mobile 120 UC','PUBG Mobile UC',120,33150),
    ('Game','PUBG Mobile','GPM180','PUBG Mobile 180 UC','PUBG Mobile UC',180,49576),
    ('Game','PUBG Mobile','GPM240','PUBG Mobile 240 UC','PUBG Mobile UC',240,65050),
    ('Game','PUBG Mobile','GPM325','PUBG Mobile 325 UC','PUBG Mobile UC',325,82275),
    ('Game','PUBG Mobile','GPM385','PUBG Mobile 385 UC','PUBG Mobile UC',385,97975),
    ('Game','PUBG Mobile','GPM445','PUBG Mobile 445 UC','PUBG Mobile UC',445,113025),
    ('Game','PUBG Mobile','GPM505','PUBG Mobile 505 UC','PUBG Mobile UC',505,131235),
    ('Game','PUBG Mobile','GPM565','PUBG Mobile 250 UC','PUBG Mobile UC',250,146050),
    ('Game','PUBG Mobile','GPM660','PUBG Mobile 263 UC','PUBG Mobile UC',263,159875),
    ('Game','PUBG Mobile','GPM900','PUBG Mobile 300 UC','PUBG Mobile UC',300,224427),
    ('Game','PUBG Mobile','GPM1320','PUBG Mobile 350 UC','PUBG Mobile UC',350,323829),
    ('Game','PUBG Mobile','GPM1800','PUBG Mobile 500 UC','PUBG Mobile UC',500,404887),
    ('Game','PUBG Mobile','GPM3850','PUBG Mobile 600 UC','PUBG Mobile UC',600,818924),
    ('Game','PUBG Mobile','VGPM60','PUBG Mobile 60 UC','PUBG Mobile Voucher UC',60,17999),
    ('Game','PUBG Mobile','VGPM120','PUBG Mobile 120 UC','PUBG Mobile Voucher UC',120,34498),
    ('Game','PUBG Mobile','VGPM180','PUBG Mobile 180 UC','PUBG Mobile Voucher UC',180,50965),
    ('Game','PUBG Mobile','VGPM240','PUBG Mobile 240 UC','PUBG Mobile Voucher UC',240,67453),
    ('Game','PUBG Mobile','VGPM325','PUBG Mobile 325 UC','PUBG Mobile Voucher UC',325,84370),
    ('Game','PUBG Mobile','VGPM385','PUBG Mobile 385 UC','PUBG Mobile Voucher UC',385,100869),
    ('Game','PUBG Mobile','VGPM445','PUBG Mobile 445 UC','PUBG Mobile Voucher UC',445,117868),
    ('Game','PUBG Mobile','VGPM505','PUBG Mobile 505 UC','PUBG Mobile Voucher UC',505,134366),
    ('Game','PUBG Mobile','VGPM565','PUBG Mobile 565 UC','PUBG Mobile Voucher UC',565,150865),
    ('Game','PUBG Mobile','VGPM660','PUBG Mobile 660 UC','PUBG Mobile Voucher UC',660,167740),
    ('Game','PUBG Mobile','VGPM720','PUBG Mobile 720 UC','PUBG Mobile Voucher UC',720,184239),
    ('Game','PUBG Mobile','VGPM780','PUBG Mobile 780 UC','PUBG Mobile Voucher UC',780,200611),
    ('Game','PUBG Mobile','VGPM840','PUBG Mobile 840 UC','PUBG Mobile Voucher UC',840,217236),
    ('Game','PUBG Mobile','VGPM900','PUBG Mobile 900 UC','PUBG Mobile Voucher UC',900,234735),
    ('Game','PUBG Mobile','VGPM985','PUBG Mobile 985 UC','PUBG Mobile Voucher UC',985,251452),
    ('Game','PUBG Mobile','VGPM1105','PUBG Mobile 1105 UC','PUBG Mobile Voucher UC',1105,252691),
    ('Game','PUBG Mobile','VGPM1165','PUBG Mobile 1165 UC','PUBG Mobile Voucher UC',1165,301107),
    ('Game','PUBG Mobile','VGPM1320','PUBG Mobile 1320 UC','PUBG Mobile Voucher UC',1320,334269),
    ('Game','PUBG Mobile','VGPM1440','PUBG Mobile 1440 UC','PUBG Mobile Voucher UC',1440,364935),
    ('Game','PUBG Mobile','VGPM1500','PUBG Mobile 1500 UC','PUBG Mobile Voucher UC',1500,384734),
    ('Game','PUBG Mobile','VGPM1800','PUBG Mobile 1800 UC','PUBG Mobile Voucher UC',1800,418466),
    ('Game','PUBG Mobile','VGPM1920','PUBG Mobile 1920 UC','PUBG Mobile Voucher UC',1920,451442),
    ('Game','PUBG Mobile','VGPM1980','PUBG Mobile 1980 UC','PUBG Mobile Voucher UC',1980,468931),
    ('Game','PUBG Mobile','VGPM2125','PUBG Mobile 2125 UC','PUBG Mobile Voucher UC',2125,502283),
    ('Game','PUBG Mobile','VGPM2460','PUBG Mobile 2460 UC','PUBG Mobile Voucher UC',2460,586101),
    ('Game','PUBG Mobile','VGPM2785','PUBG Mobile 2785 UC','PUBG Mobile Voucher UC',2785,668918),
    ('Game','PUBG Mobile','VGPM3120','PUBG Mobile 3120 UC','PUBG Mobile Voucher UC',3120,752735),
    ('Game','PUBG Mobile','VGPM3850','PUBG Mobile 3850 UC','PUBG Mobile Voucher UC',3850,836932),
    ('Game','PUBG Mobile','VGPM4030','PUBG Mobile 4030 UC','PUBG Mobile Voucher UC',4030,886397),
    ('Game','PUBG Mobile','VGPM4175','PUBG Mobile 4175 UC','PUBG Mobile Voucher UC',4175,920749),
    ('Game','PUBG Mobile','VGPM4510','PUBG Mobile 4510 UC','PUBG Mobile Voucher UC',4510,1003567),
    ('Game','PUBG Mobile','VGPM4835','PUBG Mobile 4835 UC','PUBG Mobile Voucher UC',4835,1087384),
    ('Game','PUBG Mobile','VGPM5170','PUBG Mobile 5170 UC','PUBG Mobile Voucher UC',5170,1171201),
    ('Game','PUBG Mobile','VGPM5650','PUBG Mobile 5650 UC','PUBG Mobile Voucher UC',5650,1259398),
    ('Game','PUBG Mobile','VGPM5975','PUBG Mobile 5975 UC','PUBG Mobile Voucher UC',5975,1347215),
    ('Game','PUBG Mobile','VGPM6310','PUBG Mobile 6310 UC','PUBG Mobile Voucher UC',6310,1430032),
    ('Game','PUBG Mobile','VGPM6635','PUBG Mobile 6635 UC','PUBG Mobile Voucher UC',6635,1512850),
    ('Game','PUBG Mobile','VGPM6970','PUBG Mobile 6970 UC','PUBG Mobile Voucher UC',6970,1595667),
    ('Game','PUBG Mobile','VGPM8100','PUBG Mobile 8100 UC','PUBG Mobile Voucher UC',8100,1680108),
    ('Game','PUBG Mobile','VGVPGM60','Voucher PUBG Mobile 60 UC','PUBG Mobile Voucher UC',60,14650),
    ('Game','PUBG Mobile','VGVPGM325','Voucher PUBG Mobile 325 UC','PUBG Mobile Voucher UC',325,68550),
    ('Game','PUBG Mobile','VGVPGM660','Voucher PUBG Mobile 660 UC','PUBG Mobile Voucher UC',660,136050),
    ('Game','PUBG Mobile','VGVPGM1800','Voucher PUBG Mobile 1800 UC','PUBG Mobile Voucher UC',1800,338550),
    ('Game','PUBG Mobile','VGVPGM3850','Voucher PUBG Mobile 3850 UC','PUBG Mobile Voucher UC',3850,675850),
    ('Game','PUBG Mobile','VGVPGM8100','Voucher PUBG Mobile 8100 UC','PUBG Mobile Voucher UC',8100,1350550)
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
