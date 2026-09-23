-- Seed produk top up game Roblox dan Roblox Gift Card.
-- Aman dijalankan berulang: brand, produk, dan harga di-upsert berdasarkan SKU.

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
    ('Roblox'),
    ('Roblox Gift Card')
) AS x(nama)
WHERE NOT EXISTS (
  SELECT 1 FROM public.brand b WHERE lower(trim(b.nama)) = lower(trim(x.nama))
);

WITH seed(category_name, brand_name, sku, product_name, group_name, nominal, harga) AS (
  VALUES
    ('Game','Roblox','GRB100','Voucher Roblox 100 Rbx','Roblox Rbx',100,66957),
    ('Game','Roblox','GRB200','Voucher Roblox 200 Rbx','Roblox Rbx',200,87147),
    ('Game','Roblox','GRB400','Voucher Roblox 400 Rbx','Roblox Rbx',400,102074),
    ('Game','Roblox','GRB800','Voucher Roblox 800 Rbx','Roblox Rbx',800,155262),
    ('Game','Roblox','GRB2000','Voucher Roblox 2000 Rbx','Roblox Rbx',2000,381766),
    ('Game','Roblox','GRB4500','Voucher Roblox 4500 Rbx','Roblox Rbx',4500,848450),
    ('Game','Roblox','GRB10000','Voucher Roblox 10000 Rbx','Roblox Rbx',10000,1753516),
    ('Game','Roblox Gift Card','GRGC50','Voucher Roblox Gift Card 50K','Roblox Gift Card',50000,49800),
    ('Game','Roblox Gift Card','GRGC65','Voucher Roblox Gift Card 65K','Roblox Gift Card',65000,63650),
    ('Game','Roblox Gift Card','GRGC100','Voucher Roblox Gift Card 100K','Roblox Gift Card',100000,96800),
    ('Game','Roblox Gift Card','GRGC200','Voucher Roblox Gift Card 200K','Roblox Gift Card',200000,194000),
    ('Game','Roblox Gift Card','GRGC300','Voucher Roblox Gift Card 300K','Roblox Gift Card',300000,291084),
    ('Game','Roblox Gift Card','GRGC500','Voucher Roblox Gift Card 500K','Roblox Gift Card',500000,479379),
    ('Game','Roblox','VGRB100','Voucher Roblox 100 Rbx','Roblox Voucher Rbx',100,79850),
    ('Game','Roblox','VGRB200','Voucher Roblox 200 Rbx','Roblox Voucher Rbx',200,90050),
    ('Game','Roblox','VGRB400','Voucher Roblox 400 Rbx','Roblox Voucher Rbx',400,104050),
    ('Game','Roblox','VGRB800','Voucher Roblox 800 Rbx','Roblox Voucher Rbx',800,157356),
    ('Game','Roblox','VGRB2000','Voucher Roblox 2000 Rbx','Roblox Voucher Rbx',2000,386300),
    ('Game','Roblox','VGRB4500','Voucher Roblox 4500 Rbx','Roblox Voucher Rbx',4500,786150),
    ('Game','Roblox','VGRB10000','Voucher Roblox 10000 Rbx','Roblox Voucher Rbx',10000,1597020)
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
