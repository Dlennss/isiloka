-- Seed provider dan produk Game untuk PulsaKilat.
-- Aman dijalankan berulang: brand, produk, dan harga di-upsert.

INSERT INTO public.kategori_fee_app
  (kategori_id, fee_master, fee_agent, fee_user, fee_non_user, aktif, created_at, updated_at)
SELECT k.id, 1000, 1000, 1000, 1500, true, now(), now()
FROM public.kategori k
LEFT JOIN public.kategori_fee_app existing ON existing.kategori_id = k.id
WHERE lower(trim(k.nama)) = 'game'
  AND existing.id IS NULL;

INSERT INTO public.brand (nama, aktif, dibuat_pada, diubah_pada)
SELECT x.nama, true, now(), now()
FROM (
  VALUES
    ('Free Fire'),
    ('PUBG Mobile'),
    ('Mobile Legend'),
    ('Roblox'),
    ('Point Blank')
) AS x(nama)
WHERE NOT EXISTS (
  SELECT 1 FROM public.brand b WHERE lower(trim(b.nama)) = lower(trim(x.nama))
);

WITH seed(category_name, brand_name, sku, product_name, group_name, nominal, harga) AS (
  VALUES
    ('Game','Free Fire','PK-GAME-FF-70','Free Fire 70 Diamonds','Diamond',70,10500),
    ('Game','Free Fire','PK-GAME-FF-140','Free Fire 140 Diamonds','Diamond',140,20500),
    ('Game','Free Fire','PK-GAME-FF-355','Free Fire 355 Diamonds','Diamond',355,50500),
    ('Game','Free Fire','PK-GAME-FF-720','Free Fire 720 Diamonds','Diamond',720,100500),
    ('Game','Free Fire','PK-GAME-FF-MINGGUAN','Free Fire Member Mingguan','Membership',1,31000),
    ('Game','Free Fire','PK-GAME-FF-BULANAN','Free Fire Member Bulanan','Membership',2,90500),

    ('Game','PUBG Mobile','PK-GAME-PUBG-60','PUBG Mobile 60 UC','UC',60,14500),
    ('Game','PUBG Mobile','PK-GAME-PUBG-325','PUBG Mobile 325 UC','UC',325,69500),
    ('Game','PUBG Mobile','PK-GAME-PUBG-660','PUBG Mobile 660 UC','UC',660,139000),
    ('Game','PUBG Mobile','PK-GAME-PUBG-1800','PUBG Mobile 1800 UC','UC',1800,365000),
    ('Game','PUBG Mobile','PK-GAME-PUBG-3850','PUBG Mobile 3850 UC','UC',3850,725000),
    ('Game','PUBG Mobile','PK-GAME-PUBG-8100','PUBG Mobile 8100 UC','UC',8100,1450000),

    ('Game','Mobile Legend','PK-GAME-ML-86','Mobile Legend 86 Diamonds','Diamond',86,22500),
    ('Game','Mobile Legend','PK-GAME-ML-172','Mobile Legend 172 Diamonds','Diamond',172,43500),
    ('Game','Mobile Legend','PK-GAME-ML-257','Mobile Legend 257 Diamonds','Diamond',257,63500),
    ('Game','Mobile Legend','PK-GAME-ML-344','Mobile Legend 344 Diamonds','Diamond',344,84500),
    ('Game','Mobile Legend','PK-GAME-ML-429','Mobile Legend 429 Diamonds','Diamond',429,105000),
    ('Game','Mobile Legend','PK-GAME-ML-WEEKLY','Mobile Legend Weekly Diamond Pass','Pass',1,31500),

    ('Game','Roblox','PK-GAME-ROBLOX-80','Roblox 80 Robux','Robux',80,17500),
    ('Game','Roblox','PK-GAME-ROBLOX-400','Roblox 400 Robux','Robux',400,76500),
    ('Game','Roblox','PK-GAME-ROBLOX-800','Roblox 800 Robux','Robux',800,151000),
    ('Game','Roblox','PK-GAME-ROBLOX-1700','Roblox 1700 Robux','Robux',1700,315000),
    ('Game','Roblox','PK-GAME-ROBLOX-4500','Roblox 4500 Robux','Robux',4500,765000),
    ('Game','Roblox','PK-GAME-ROBLOX-10000','Roblox 10000 Robux','Robux',10000,1510000),

    ('Game','Point Blank','PK-GAME-PB-1200','Point Blank 1200 Cash','Cash',1200,11500),
    ('Game','Point Blank','PK-GAME-PB-2400','Point Blank 2400 Cash','Cash',2400,21500),
    ('Game','Point Blank','PK-GAME-PB-6000','Point Blank 6000 Cash','Cash',6000,50500),
    ('Game','Point Blank','PK-GAME-PB-12000','Point Blank 12000 Cash','Cash',12000,100500),
    ('Game','Point Blank','PK-GAME-PB-24000','Point Blank 24000 Cash','Cash',24000,199000),
    ('Game','Point Blank','PK-GAME-PB-36000','Point Blank 36000 Cash','Cash',36000,295000)
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
