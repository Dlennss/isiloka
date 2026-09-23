-- Seed tambahan provider Game sesuai daftar PulsaKilat.
-- Aman dijalankan berulang.

INSERT INTO public.brand (nama, aktif, dibuat_pada, diubah_pada)
SELECT x.nama, true, now(), now()
FROM (
  VALUES
    ('Magic Chess Go Go'),
    ('Free Fire MAX'),
    ('Call of Duty Mobile'),
    ('Hago'),
    ('Genshin Impact'),
    ('ZEPETO'),
    ('Blood Strike'),
) AS x(nama)
WHERE NOT EXISTS (
  SELECT 1 FROM public.brand b WHERE lower(trim(b.nama)) = lower(trim(x.nama))
);

WITH seed(category_name, brand_name, sku, product_name, group_name, nominal, harga) AS (
  VALUES
    ('Game','Magic Chess Go Go','PK-GAME-MCGG-50','Magic Chess Go Go 50 Diamonds','Diamond',50,13500),
    ('Game','Magic Chess Go Go','PK-GAME-MCGG-100','Magic Chess Go Go 100 Diamonds','Diamond',100,25500),
    ('Game','Magic Chess Go Go','PK-GAME-MCGG-250','Magic Chess Go Go 250 Diamonds','Diamond',250,61500),
    ('Game','Magic Chess Go Go','PK-GAME-MCGG-500','Magic Chess Go Go 500 Diamonds','Diamond',500,121500),

    ('Game','Free Fire MAX','PK-GAME-FFMAX-70','Free Fire MAX 70 Diamonds','Diamond',70,10500),
    ('Game','Free Fire MAX','PK-GAME-FFMAX-140','Free Fire MAX 140 Diamonds','Diamond',140,20500),
    ('Game','Free Fire MAX','PK-GAME-FFMAX-355','Free Fire MAX 355 Diamonds','Diamond',355,50500),
    ('Game','Free Fire MAX','PK-GAME-FFMAX-720','Free Fire MAX 720 Diamonds','Diamond',720,100500),

    ('Game','Call of Duty Mobile','PK-GAME-CODM-31','Call of Duty Mobile 31 CP','CP',31,6500),
    ('Game','Call of Duty Mobile','PK-GAME-CODM-62','Call of Duty Mobile 62 CP','CP',62,12500),
    ('Game','Call of Duty Mobile','PK-GAME-CODM-127','Call of Duty Mobile 127 CP','CP',127,24500),
    ('Game','Call of Duty Mobile','PK-GAME-CODM-320','Call of Duty Mobile 320 CP','CP',320,60500),

    ('Game','Hago','PK-GAME-HAGO-60','Hago 60 Diamonds','Diamond',60,10500),
    ('Game','Hago','PK-GAME-HAGO-120','Hago 120 Diamonds','Diamond',120,20500),
    ('Game','Hago','PK-GAME-HAGO-300','Hago 300 Diamonds','Diamond',300,50500),
    ('Game','Hago','PK-GAME-HAGO-600','Hago 600 Diamonds','Diamond',600,99500),

    ('Game','Genshin Impact','PK-GAME-GENSHIN-60','Genshin Impact 60 Genesis Crystals','Genesis Crystal',60,15500),
    ('Game','Genshin Impact','PK-GAME-GENSHIN-300','Genshin Impact 300 Genesis Crystals','Genesis Crystal',300,75500),
    ('Game','Genshin Impact','PK-GAME-GENSHIN-980','Genshin Impact 980 Genesis Crystals','Genesis Crystal',980,229000),
    ('Game','Genshin Impact','PK-GAME-GENSHIN-WELKIN','Genshin Impact Blessing of Welkin Moon','Pass',1,74500),

    ('Game','ZEPETO','PK-GAME-ZEPETO-14','ZEPETO 14 ZEM','ZEM',14,15500),
    ('Game','ZEPETO','PK-GAME-ZEPETO-29','ZEPETO 29 ZEM','ZEM',29,30500),
    ('Game','ZEPETO','PK-GAME-ZEPETO-60','ZEPETO 60 ZEM','ZEM',60,60500),
    ('Game','ZEPETO','PK-GAME-ZEPETO-125','ZEPETO 125 ZEM','ZEM',125,121000),

    ('Game','Blood Strike','PK-GAME-BS-100','Blood Strike 100 Gold','Gold',100,15500),
    ('Game','Blood Strike','PK-GAME-BS-300','Blood Strike 300 Gold','Gold',300,45500),
    ('Game','Blood Strike','PK-GAME-BS-500','Blood Strike 500 Gold','Gold',500,73500),
    ('Game','Blood Strike','PK-GAME-BS-1000','Blood Strike 1000 Gold','Gold',1000,145000),
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
