-- Seed produk top up game Point Blank.
-- Aman dijalankan berulang: brand, produk, dan harga di-upsert berdasarkan SKU.

INSERT INTO public.kategori_fee_app
  (kategori_id, fee_master, fee_agent, fee_user, fee_non_user, aktif, created_at, updated_at)
SELECT k.id, 1000, 1000, 1000, 1500, true, now(), now()
FROM public.kategori k
LEFT JOIN public.kategori_fee_app existing ON existing.kategori_id = k.id
WHERE lower(trim(k.nama)) = 'game'
  AND existing.id IS NULL;

INSERT INTO public.brand (nama, aktif, dibuat_pada, diubah_pada)
SELECT 'Point Blank', true, now(), now()
WHERE NOT EXISTS (
  SELECT 1 FROM public.brand b WHERE lower(trim(b.nama)) = lower(trim('Point Blank'))
);

WITH seed(category_name, brand_name, sku, product_name, group_name, nominal, harga) AS (
  VALUES
    ('Game','Point Blank','GPB2400','Point Blank 2400 Cash','Point Blank Cash',2400,18625),
    ('Game','Point Blank','GPB6000','Point Blank 6000 Cash','Point Blank Cash',6000,45000),
    ('Game','Point Blank','GPB12000','Point Blank 12000 Cash','Point Blank Cash',12000,88658),
    ('Game','Point Blank','GPB24000','Point Blank 24000 Cash','Point Blank Cash',24000,176341),
    ('Game','Point Blank','GPB36000','Point Blank 36000 Cash','Point Blank Cash',36000,257125),
    ('Game','Point Blank','GPB60000','Point Blank 60000 Cash','Point Blank Cash',60000,439140),
    ('Game','Point Blank','VGPB2400','Point Blank 2400 Cash','Point Blank Voucher Cash',2400,18737),
    ('Game','Point Blank','VGPB6000','Point Blank 6000 Cash','Point Blank Voucher Cash',6000,45167),
    ('Game','Point Blank','VGPB12000','Point Blank 12000 Cash','Point Blank Voucher Cash',12000,88950),
    ('Game','Point Blank','VGPB24000','Point Blank 24000 Cash','Point Blank Voucher Cash',24000,177100),
    ('Game','Point Blank','VGPB36000','Point Blank 36000 Cash','Point Blank Voucher Cash',36000,266150),
    ('Game','Point Blank','VGPB60000','Point Blank 60000 Cash','Point Blank Voucher Cash',60000,440250)
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
