-- Seed kartu produk Pulsa & Paket Data untuk tampilan PulsaKilat.
-- Aman dijalankan berulang: SKU dan pricing akan di-upsert.

INSERT INTO public.kategori_fee_app
  (kategori_id, fee_master, fee_agent, fee_user, fee_non_user, aktif, created_at, updated_at)
SELECT k.id, x.fee_user, x.fee_user, x.fee_user, x.fee_guest, true, now(), now()
FROM (
  VALUES
    ('Pulsa', 150::bigint, 150::bigint, 150::bigint, 200::bigint),
    ('Paket Data', 1000::bigint, 1000::bigint, 1000::bigint, 1500::bigint)
) AS x(nama, fee_master, fee_agent, fee_user, fee_guest)
JOIN public.kategori k ON lower(trim(k.nama)) = lower(trim(x.nama))
LEFT JOIN public.kategori_fee_app existing ON existing.kategori_id = k.id
WHERE existing.id IS NULL;

INSERT INTO public.brand (nama, aktif, dibuat_pada, diubah_pada)
SELECT 'AXIS', true, now(), now()
WHERE NOT EXISTS (
  SELECT 1 FROM public.brand WHERE lower(trim(nama)) = lower('AXIS')
);

WITH seed(category_name, brand_name, sku, product_name, group_name, nominal, harga) AS (
  VALUES
    -- Pulsa Telkomsel
    ('Pulsa','Telkomsel','PK-PULSA-TELKOMSEL-5000','Telkomsel 5rb','Pulsa Reguler',5000,5200),
    ('Pulsa','Telkomsel','PK-PULSA-TELKOMSEL-10000','Telkomsel 10rb','Pulsa Reguler',10000,10200),
    ('Pulsa','Telkomsel','PK-PULSA-TELKOMSEL-12000','Telkomsel 12rb','Pulsa Reguler',12000,12200),
    ('Pulsa','Telkomsel','PK-PULSA-TELKOMSEL-15000','Telkomsel 15rb','Pulsa Reguler',15000,15250),
    ('Pulsa','Telkomsel','PK-PULSA-TELKOMSEL-20000','Telkomsel 20rb','Pulsa Reguler',20000,20300),
    ('Pulsa','Telkomsel','PK-PULSA-TELKOMSEL-25000','Telkomsel 25rb','Pulsa Reguler',25000,25300),
    ('Pulsa','Telkomsel','PK-PULSA-TELKOMSEL-30000','Telkomsel 30rb','Pulsa Reguler',30000,30350),
    ('Pulsa','Telkomsel','PK-PULSA-TELKOMSEL-40000','Telkomsel 40rb','Pulsa Reguler',40000,40400),
    ('Pulsa','Telkomsel','PK-PULSA-TELKOMSEL-50000','Telkomsel 50rb','Pulsa Reguler',50000,50450),
    ('Pulsa','Telkomsel','PK-PULSA-TELKOMSEL-100000','Telkomsel 100rb','Pulsa Reguler',100000,100900),

    -- Pulsa Indosat
    ('Pulsa','Indosat','PK-PULSA-INDOSAT-5000','Indosat 5rb','Pulsa Reguler',5000,5150),
    ('Pulsa','Indosat','PK-PULSA-INDOSAT-10000','Indosat 10rb','Pulsa Reguler',10000,10150),
    ('Pulsa','Indosat','PK-PULSA-INDOSAT-12000','Indosat 12rb','Pulsa Reguler',12000,12150),
    ('Pulsa','Indosat','PK-PULSA-INDOSAT-15000','Indosat 15rb','Pulsa Reguler',15000,15200),
    ('Pulsa','Indosat','PK-PULSA-INDOSAT-20000','Indosat 20rb','Pulsa Reguler',20000,20250),
    ('Pulsa','Indosat','PK-PULSA-INDOSAT-25000','Indosat 25rb','Pulsa Reguler',25000,25250),
    ('Pulsa','Indosat','PK-PULSA-INDOSAT-30000','Indosat 30rb','Pulsa Reguler',30000,30300),
    ('Pulsa','Indosat','PK-PULSA-INDOSAT-40000','Indosat 40rb','Pulsa Reguler',40000,40350),
    ('Pulsa','Indosat','PK-PULSA-INDOSAT-50000','Indosat 50rb','Pulsa Reguler',50000,50400),
    ('Pulsa','Indosat','PK-PULSA-INDOSAT-100000','Indosat 100rb','Pulsa Reguler',100000,100800),

    -- Pulsa XL
    ('Pulsa','XL','PK-PULSA-XL-5000','XL 5rb','Pulsa Reguler',5000,5150),
    ('Pulsa','XL','PK-PULSA-XL-10000','XL 10rb','Pulsa Reguler',10000,10150),
    ('Pulsa','XL','PK-PULSA-XL-12000','XL 12rb','Pulsa Reguler',12000,12150),
    ('Pulsa','XL','PK-PULSA-XL-15000','XL 15rb','Pulsa Reguler',15000,15200),
    ('Pulsa','XL','PK-PULSA-XL-20000','XL 20rb','Pulsa Reguler',20000,20250),
    ('Pulsa','XL','PK-PULSA-XL-25000','XL 25rb','Pulsa Reguler',25000,25250),
    ('Pulsa','XL','PK-PULSA-XL-30000','XL 30rb','Pulsa Reguler',30000,30300),
    ('Pulsa','XL','PK-PULSA-XL-40000','XL 40rb','Pulsa Reguler',40000,40350),
    ('Pulsa','XL','PK-PULSA-XL-50000','XL 50rb','Pulsa Reguler',50000,50400),
    ('Pulsa','XL','PK-PULSA-XL-100000','XL 100rb','Pulsa Reguler',100000,100800),

    -- Pulsa AXIS
    ('Pulsa','AXIS','PK-PULSA-AXIS-5000','AXIS 5rb','Pulsa Reguler',5000,5100),
    ('Pulsa','AXIS','PK-PULSA-AXIS-10000','AXIS 10rb','Pulsa Reguler',10000,10100),
    ('Pulsa','AXIS','PK-PULSA-AXIS-12000','AXIS 12rb','Pulsa Reguler',12000,12100),
    ('Pulsa','AXIS','PK-PULSA-AXIS-15000','AXIS 15rb','Pulsa Reguler',15000,15150),
    ('Pulsa','AXIS','PK-PULSA-AXIS-20000','AXIS 20rb','Pulsa Reguler',20000,20200),
    ('Pulsa','AXIS','PK-PULSA-AXIS-25000','AXIS 25rb','Pulsa Reguler',25000,25200),
    ('Pulsa','AXIS','PK-PULSA-AXIS-30000','AXIS 30rb','Pulsa Reguler',30000,30250),
    ('Pulsa','AXIS','PK-PULSA-AXIS-40000','AXIS 40rb','Pulsa Reguler',40000,40300),
    ('Pulsa','AXIS','PK-PULSA-AXIS-50000','AXIS 50rb','Pulsa Reguler',50000,50350),
    ('Pulsa','AXIS','PK-PULSA-AXIS-100000','AXIS 100rb','Pulsa Reguler',100000,100700),

    -- Pulsa Smartfren
    ('Pulsa','Smartfren','PK-PULSA-SMARTFREN-5000','Smartfren 5rb','Pulsa Reguler',5000,5100),
    ('Pulsa','Smartfren','PK-PULSA-SMARTFREN-10000','Smartfren 10rb','Pulsa Reguler',10000,10100),
    ('Pulsa','Smartfren','PK-PULSA-SMARTFREN-12000','Smartfren 12rb','Pulsa Reguler',12000,12100),
    ('Pulsa','Smartfren','PK-PULSA-SMARTFREN-15000','Smartfren 15rb','Pulsa Reguler',15000,15150),
    ('Pulsa','Smartfren','PK-PULSA-SMARTFREN-20000','Smartfren 20rb','Pulsa Reguler',20000,20200),
    ('Pulsa','Smartfren','PK-PULSA-SMARTFREN-25000','Smartfren 25rb','Pulsa Reguler',25000,25200),
    ('Pulsa','Smartfren','PK-PULSA-SMARTFREN-30000','Smartfren 30rb','Pulsa Reguler',30000,30250),
    ('Pulsa','Smartfren','PK-PULSA-SMARTFREN-40000','Smartfren 40rb','Pulsa Reguler',40000,40300),
    ('Pulsa','Smartfren','PK-PULSA-SMARTFREN-50000','Smartfren 50rb','Pulsa Reguler',50000,50350),
    ('Pulsa','Smartfren','PK-PULSA-SMARTFREN-100000','Smartfren 100rb','Pulsa Reguler',100000,100700),

    -- Pulsa Tri
    ('Pulsa','Tri','PK-PULSA-TRI-5000','Tri 5rb','Pulsa Reguler',5000,5100),
    ('Pulsa','Tri','PK-PULSA-TRI-10000','Tri 10rb','Pulsa Reguler',10000,10100),
    ('Pulsa','Tri','PK-PULSA-TRI-12000','Tri 12rb','Pulsa Reguler',12000,12100),
    ('Pulsa','Tri','PK-PULSA-TRI-15000','Tri 15rb','Pulsa Reguler',15000,15150),
    ('Pulsa','Tri','PK-PULSA-TRI-20000','Tri 20rb','Pulsa Reguler',20000,20200),
    ('Pulsa','Tri','PK-PULSA-TRI-25000','Tri 25rb','Pulsa Reguler',25000,25200),
    ('Pulsa','Tri','PK-PULSA-TRI-30000','Tri 30rb','Pulsa Reguler',30000,30250),
    ('Pulsa','Tri','PK-PULSA-TRI-40000','Tri 40rb','Pulsa Reguler',40000,40300),
    ('Pulsa','Tri','PK-PULSA-TRI-50000','Tri 50rb','Pulsa Reguler',50000,50350),
    ('Pulsa','Tri','PK-PULSA-TRI-100000','Tri 100rb','Pulsa Reguler',100000,100700),

    -- Paket Data semua operator
    ('Paket Data','Telkomsel','PK-DATA-TELKOMSEL-1GB','Telkomsel 1GB 7 Hari','Paket Data',1000,11200),
    ('Paket Data','Telkomsel','PK-DATA-TELKOMSEL-2GB','Telkomsel 2GB 15 Hari','Paket Data',2000,18200),
    ('Paket Data','Telkomsel','PK-DATA-TELKOMSEL-3GB','Telkomsel 3GB 30 Hari','Paket Data',3000,25200),
    ('Paket Data','Telkomsel','PK-DATA-TELKOMSEL-5GB','Telkomsel 5GB 30 Hari','Paket Data',5000,41200),
    ('Paket Data','Telkomsel','PK-DATA-TELKOMSEL-8GB','Telkomsel 8GB 30 Hari','Paket Data',8000,62200),
    ('Paket Data','Telkomsel','PK-DATA-TELKOMSEL-10GB','Telkomsel 10GB 30 Hari','Paket Data',10000,75200),
    ('Paket Data','Telkomsel','PK-DATA-TELKOMSEL-15GB','Telkomsel 15GB 30 Hari','Paket Data',15000,98200),
    ('Paket Data','Telkomsel','PK-DATA-TELKOMSEL-20GB','Telkomsel 20GB 30 Hari','Paket Data',20000,124200),

    ('Paket Data','Indosat','PK-DATA-INDOSAT-1GB','Indosat 1GB 7 Hari','Paket Data',1000,10200),
    ('Paket Data','Indosat','PK-DATA-INDOSAT-2GB','Indosat 2GB 15 Hari','Paket Data',2000,16200),
    ('Paket Data','Indosat','PK-DATA-INDOSAT-3GB','Indosat 3GB 30 Hari','Paket Data',3000,22200),
    ('Paket Data','Indosat','PK-DATA-INDOSAT-5GB','Indosat 5GB 30 Hari','Paket Data',5000,35200),
    ('Paket Data','Indosat','PK-DATA-INDOSAT-8GB','Indosat 8GB 30 Hari','Paket Data',8000,50200),
    ('Paket Data','Indosat','PK-DATA-INDOSAT-10GB','Indosat 10GB 30 Hari','Paket Data',10000,63200),
    ('Paket Data','Indosat','PK-DATA-INDOSAT-15GB','Indosat 15GB 30 Hari','Paket Data',15000,82200),
    ('Paket Data','Indosat','PK-DATA-INDOSAT-20GB','Indosat 20GB 30 Hari','Paket Data',20000,104200),

    ('Paket Data','XL','PK-DATA-XL-1GB','XL 1GB 7 Hari','Paket Data',1000,10200),
    ('Paket Data','XL','PK-DATA-XL-2GB','XL 2GB 15 Hari','Paket Data',2000,16200),
    ('Paket Data','XL','PK-DATA-XL-3GB','XL 3GB 30 Hari','Paket Data',3000,22200),
    ('Paket Data','XL','PK-DATA-XL-5GB','XL 5GB 30 Hari','Paket Data',5000,35200),
    ('Paket Data','XL','PK-DATA-XL-8GB','XL 8GB 30 Hari','Paket Data',8000,50200),
    ('Paket Data','XL','PK-DATA-XL-10GB','XL 10GB 30 Hari','Paket Data',10000,63200),
    ('Paket Data','XL','PK-DATA-XL-15GB','XL 15GB 30 Hari','Paket Data',15000,82200),
    ('Paket Data','XL','PK-DATA-XL-20GB','XL 20GB 30 Hari','Paket Data',20000,104200),

    ('Paket Data','AXIS','PK-DATA-AXIS-1GB','AXIS 1GB 7 Hari','Paket Data',1000,9200),
    ('Paket Data','AXIS','PK-DATA-AXIS-2GB','AXIS 2GB 15 Hari','Paket Data',2000,14200),
    ('Paket Data','AXIS','PK-DATA-AXIS-3GB','AXIS 3GB 30 Hari','Paket Data',3000,20200),
    ('Paket Data','AXIS','PK-DATA-AXIS-5GB','AXIS 5GB 30 Hari','Paket Data',5000,32200),
    ('Paket Data','AXIS','PK-DATA-AXIS-8GB','AXIS 8GB 30 Hari','Paket Data',8000,45200),
    ('Paket Data','AXIS','PK-DATA-AXIS-10GB','AXIS 10GB 30 Hari','Paket Data',10000,58200),
    ('Paket Data','AXIS','PK-DATA-AXIS-15GB','AXIS 15GB 30 Hari','Paket Data',15000,76200),
    ('Paket Data','AXIS','PK-DATA-AXIS-20GB','AXIS 20GB 30 Hari','Paket Data',20000,98200),

    ('Paket Data','Smartfren','PK-DATA-SMARTFREN-1GB','Smartfren 1GB 7 Hari','Paket Data',1000,9200),
    ('Paket Data','Smartfren','PK-DATA-SMARTFREN-2GB','Smartfren 2GB 15 Hari','Paket Data',2000,14200),
    ('Paket Data','Smartfren','PK-DATA-SMARTFREN-3GB','Smartfren 3GB 30 Hari','Paket Data',3000,20200),
    ('Paket Data','Smartfren','PK-DATA-SMARTFREN-5GB','Smartfren 5GB 30 Hari','Paket Data',5000,32200),
    ('Paket Data','Smartfren','PK-DATA-SMARTFREN-8GB','Smartfren 8GB 30 Hari','Paket Data',8000,45200),
    ('Paket Data','Smartfren','PK-DATA-SMARTFREN-10GB','Smartfren 10GB 30 Hari','Paket Data',10000,58200),
    ('Paket Data','Smartfren','PK-DATA-SMARTFREN-15GB','Smartfren 15GB 30 Hari','Paket Data',15000,76200),
    ('Paket Data','Smartfren','PK-DATA-SMARTFREN-20GB','Smartfren 20GB 30 Hari','Paket Data',20000,98200),

    ('Paket Data','Tri','PK-DATA-TRI-1GB','Tri 1GB 7 Hari','Paket Data',1000,9200),
    ('Paket Data','Tri','PK-DATA-TRI-2GB','Tri 2GB 15 Hari','Paket Data',2000,14200),
    ('Paket Data','Tri','PK-DATA-TRI-3GB','Tri 3GB 30 Hari','Paket Data',3000,20200),
    ('Paket Data','Tri','PK-DATA-TRI-5GB','Tri 5GB 30 Hari','Paket Data',5000,32200),
    ('Paket Data','Tri','PK-DATA-TRI-8GB','Tri 8GB 30 Hari','Paket Data',8000,45200),
    ('Paket Data','Tri','PK-DATA-TRI-10GB','Tri 10GB 30 Hari','Paket Data',10000,58200),
    ('Paket Data','Tri','PK-DATA-TRI-15GB','Tri 15GB 30 Hari','Paket Data',15000,76200),
    ('Paket Data','Tri','PK-DATA-TRI-20GB','Tri 20GB 30 Hari','Paket Data',20000,98200)
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
