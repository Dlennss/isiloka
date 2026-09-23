-- Seed produk top up game Mobile Legend Promo.
-- Aman dijalankan berulang: brand, produk, dan harga di-upsert berdasarkan SKU.

INSERT INTO public.kategori_fee_app
  (kategori_id, fee_master, fee_agent, fee_user, fee_non_user, aktif, created_at, updated_at)
SELECT k.id, 1000, 1000, 1000, 1500, true, now(), now()
FROM public.kategori k
LEFT JOIN public.kategori_fee_app existing ON existing.kategori_id = k.id
WHERE lower(trim(k.nama)) = 'game'
  AND existing.id IS NULL;

INSERT INTO public.brand (nama, aktif, dibuat_pada, diubah_pada)
SELECT 'Mobile Legend', true, now(), now()
WHERE NOT EXISTS (
  SELECT 1 FROM public.brand b WHERE lower(trim(b.nama)) = lower(trim('Mobile Legend'))
);

WITH seed(category_name, brand_name, sku, product_name, group_name, nominal, harga) AS (
  VALUES
    ('Game','Mobile Legend','MLY5','Mobile Legend Promo 5 Diamond','Mobile Legend Promo',5,2437),
    ('Game','Mobile Legend','MLY12','Mobile Legend Promo 12 Diamond','Mobile Legend Promo',12,4340),
    ('Game','Mobile Legend','MLY14','Mobile Legend Promo 14 Diamond','Mobile Legend Promo',14,4830),
    ('Game','Mobile Legend','MLY28','Mobile Legend Promo 28 Diamond','Mobile Legend Promo',28,8560),
    ('Game','Mobile Legend','MLY86','Mobile Legend Promo 86 Diamond','Mobile Legend Promo',86,22110),
    ('Game','Mobile Legend','MLY172','Mobile Legend Promo 172 Diamond','Mobile Legend Promo',172,45123),
    ('Game','Mobile Legend','MLY257','Mobile Legend Promo 257 Diamond','Mobile Legend Promo',257,65590),
    ('Game','Mobile Legend','MLY344','Mobile Legend Promo 344 Diamond','Mobile Legend Promo',344,88630),
    ('Game','Mobile Legend','MLY568','Mobile Legend Promo 568 Diamond','Mobile Legend Promo',568,140685),
    ('Game','Mobile Legend','MLY706','Mobile Legend Promo 706 Diamond','Mobile Legend Promo',706,174200),
    ('Game','Mobile Legend','MLY875','Mobile Legend Promo 875 Diamond','Mobile Legend Promo',875,217064),
    ('Game','Mobile Legend','MLY2010','Mobile Legend Promo 2010 Diamond','Mobile Legend Promo',2010,468737)
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
