-- Seed produk top up game Free Fire.
-- Aman dijalankan berulang: brand, produk, dan harga di-upsert berdasarkan SKU.

INSERT INTO public.kategori_fee_app
  (kategori_id, fee_master, fee_agent, fee_user, fee_non_user, aktif, created_at, updated_at)
SELECT k.id, 1000, 1000, 1000, 1500, true, now(), now()
FROM public.kategori k
LEFT JOIN public.kategori_fee_app existing ON existing.kategori_id = k.id
WHERE lower(trim(k.nama)) = 'game'
  AND existing.id IS NULL;

INSERT INTO public.brand (nama, aktif, dibuat_pada, diubah_pada)
SELECT 'Free Fire', true, now(), now()
WHERE NOT EXISTS (
  SELECT 1 FROM public.brand b WHERE lower(trim(b.nama)) = lower(trim('Free Fire'))
);

WITH seed(category_name, brand_name, sku, product_name, group_name, nominal, harga) AS (
  VALUES
    ('Game','Free Fire','GFF5','Free Fire 5 Diamond','Free Fire Diamond',5,1765),
    ('Game','Free Fire','GFF12','Free Fire 12 Diamond','Free Fire Diamond',12,2650),
    ('Game','Free Fire','GFF50','Free Fire 50 Diamond','Free Fire Diamond',50,7165),
    ('Game','Free Fire','GFF70','Free Fire 70 Diamond','Free Fire Diamond',70,9050),
    ('Game','Free Fire','GFF100','Free Fire 100 Diamond','Free Fire Diamond',100,13120),
    ('Game','Free Fire','GFF140','Free Fire 140 Diamond','Free Fire Diamond',140,16800),
    ('Game','Free Fire','GFF210','Free Fire 210 Diamond','Free Fire Diamond',210,25050),
    ('Game','Free Fire','GFF355','Free Fire 355 Diamond','Free Fire Diamond',355,41170),
    ('Game','Free Fire','GFF500','Free Fire 500 Diamond','Free Fire Diamond',500,58550),
    ('Game','Free Fire','GFF720','Free Fire 720 Diamond','Free Fire Diamond',720,81700),
    ('Game','Free Fire','GFF1450','Free Fire 1450 Diamond','Free Fire Diamond',1450,167100),
    ('Game','Free Fire','GFF2180','Free Fire 2180 Diamond','Free Fire Diamond',2180,250550),
    ('Game','Free Fire','GFF3640','Free Fire 3640 Diamond','Free Fire Diamond',3640,417010),
    ('Game','Free Fire','VGFF5','Free Fire 5 Diamond','Free Fire Voucher Diamond',5,2325),
    ('Game','Free Fire','VGFF10','Free Fire 10 Diamond','Free Fire Voucher Diamond',10,3150),
    ('Game','Free Fire','VGFF12','Free Fire 12 Diamond','Free Fire Voucher Diamond',12,3368),
    ('Game','Free Fire','VGFF15','Free Fire 15 Diamond','Free Fire Voucher Diamond',15,3958),
    ('Game','Free Fire','VGFF20','Free Fire 20 Diamond','Free Fire Voucher Diamond',20,4801),
    ('Game','Free Fire','VGFF25','Free Fire 25 Diamond','Free Fire Voucher Diamond',25,5626),
    ('Game','Free Fire','VGFF30','Free Fire 30 Diamond','Free Fire Voucher Diamond',30,6463),
    ('Game','Free Fire','VGFF40','Free Fire 40 Diamond','Free Fire Voucher Diamond',40,8907),
    ('Game','Free Fire','VGFF50','Free Fire 50 Diamond','Free Fire Voucher Diamond',50,8880),
    ('Game','Free Fire','VGFF55','Free Fire 55 Diamond','Free Fire Voucher Diamond',55,9681),
    ('Game','Free Fire','VGFF60','Free Fire 60 Diamond','Free Fire Voucher Diamond',60,9692),
    ('Game','Free Fire','VGFF70','Free Fire 70 Diamond','Free Fire Voucher Diamond',70,11200),
    ('Game','Free Fire','VGFF75','Free Fire 75 Diamond','Free Fire Voucher Diamond',75,11400),
    ('Game','Free Fire','VGFF80','Free Fire 80 Diamond','Free Fire Voucher Diamond',80,12225),
    ('Game','Free Fire','VGFF90','Free Fire 90 Diamond','Free Fire Voucher Diamond',90,14670),
    ('Game','Free Fire','VGFF95','Free Fire 95 Diamond','Free Fire Voucher Diamond',95,14607),
    ('Game','Free Fire','VGFF100','Free Fire 100 Diamond','Free Fire Voucher Diamond',100,15549),
    ('Game','Free Fire','VGFF120','Free Fire 120 Diamond','Free Fire Voucher Diamond',120,18043),
    ('Game','Free Fire','VGFF130','Free Fire 130 Diamond','Free Fire Voucher Diamond',130,18953),
    ('Game','Free Fire','VGFF140','Free Fire 140 Diamond','Free Fire Voucher Diamond',140,20150),
    ('Game','Free Fire','VGFF145','Free Fire 145 Diamond','Free Fire Voucher Diamond',145,20591),
    ('Game','Free Fire','VGFF150','Free Fire 150 Diamond','Free Fire Voucher Diamond',150,22122),
    ('Game','Free Fire','VGFF160','Free Fire 160 Diamond','Free Fire Voucher Diamond',160,23200),
    ('Game','Free Fire','VGFF180','Free Fire 180 Diamond','Free Fire Voucher Diamond',180,26326),
    ('Game','Free Fire','VGFF190','Free Fire 190 Diamond','Free Fire Voucher Diamond',190,26549),
    ('Game','Free Fire','VGFF200','Free Fire 200 Diamond','Free Fire Voucher Diamond',200,28746),
    ('Game','Free Fire','VGFF210','Free Fire 210 Diamond','Free Fire Voucher Diamond',210,33429),
    ('Game','Free Fire','VGFF250','Free Fire 250 Diamond','Free Fire Voucher Diamond',250,35337),
    ('Game','Free Fire','VGFF280','Free Fire 280 Diamond','Free Fire Voucher Diamond',280,38049),
    ('Game','Free Fire','VGFF300','Free Fire 300 Diamond','Free Fire Voucher Diamond',300,41071),
    ('Game','Free Fire','VGFF350','Free Fire 350 Diamond','Free Fire Voucher Diamond',350,46400),
    ('Game','Free Fire','VGFF355','Free Fire 355 Diamond','Free Fire Voucher Diamond',355,47000),
    ('Game','Free Fire','VGFF375','Free Fire 375 Diamond','Free Fire Voucher Diamond',375,50424),
    ('Game','Free Fire','VGFF400','Free Fire 400 Diamond','Free Fire Voucher Diamond',400,54115),
    ('Game','Free Fire','VGFF405','Free Fire 405 Diamond','Free Fire Voucher Diamond',405,53974),
    ('Game','Free Fire','VGFF425','Free Fire 425 Diamond','Free Fire Voucher Diamond',425,56067),
    ('Game','Free Fire','VGFF475','Free Fire 475 Diamond','Free Fire Voucher Diamond',475,63049),
    ('Game','Free Fire','VGFF500','Free Fire 500 Diamond','Free Fire Voucher Diamond',500,66349),
    ('Game','Free Fire','VGFF510','Free Fire 510 Diamond','Free Fire Voucher Diamond',510,67999),
    ('Game','Free Fire','VGFF512','Free Fire 512 Diamond','Free Fire Voucher Diamond',512,68355),
    ('Game','Free Fire','VGFF545','Free Fire 545 Diamond','Free Fire Voucher Diamond',545,71631),
    ('Game','Free Fire','VGFF565','Free Fire 565 Diamond','Free Fire Voucher Diamond',565,74089),
    ('Game','Free Fire','VGFF600','Free Fire 600 Diamond','Free Fire Voucher Diamond',600,79823),
    ('Game','Free Fire','VGFF635','Free Fire 635 Diamond','Free Fire Voucher Diamond',635,83100),
    ('Game','Free Fire','VGFF700','Free Fire 700 Diamond','Free Fire Voucher Diamond',700,92111),
    ('Game','Free Fire','VGFF720','Free Fire 720 Diamond','Free Fire Voucher Diamond',720,95300),
    ('Game','Free Fire','VGFF770','Free Fire 770 Diamond','Free Fire Voucher Diamond',770,102665),
    ('Game','Free Fire','VGFF790','Free Fire 790 Diamond','Free Fire Voucher Diamond',790,106100),
    ('Game','Free Fire','VGFF800','Free Fire 800 Diamond','Free Fire Voucher Diamond',800,106761),
    ('Game','Free Fire','VGFF860','Free Fire 860 Diamond','Free Fire Voucher Diamond',860,115200),
    ('Game','Free Fire','VGFF930','Free Fire 930 Diamond','Free Fire Voucher Diamond',930,129300),
    ('Game','Free Fire','VGFF1000','Free Fire 1000 Diamond','Free Fire Voucher Diamond',1000,137156),
    ('Game','Free Fire','VGFF1050','Free Fire 1050 Diamond','Free Fire Voucher Diamond',1050,143709),
    ('Game','Free Fire','VGFF1075','Free Fire 1075 Diamond','Free Fire Voucher Diamond',1075,146167),
    ('Game','Free Fire','VGFF1080','Free Fire 1080 Diamond','Free Fire Voucher Diamond',1080,146986),
    ('Game','Free Fire','VGFF1200','Free Fire 1200 Diamond','Free Fire Voucher Diamond',1200,162551),
    ('Game','Free Fire','VGFF1215','Free Fire 1215 Diamond','Free Fire Voucher Diamond',1215,165008),
    ('Game','Free Fire','VGFF1300','Free Fire 1300 Diamond','Free Fire Voucher Diamond',1300,175658),
    ('Game','Free Fire','VGFF1440','Free Fire 1440 Diamond','Free Fire Voucher Diamond',1440,191222),
    ('Game','Free Fire','VGFF1450','Free Fire 1450 Diamond','Free Fire Voucher Diamond',1450,192516),
    ('Game','Free Fire','VGFF1490','Free Fire 1490 Diamond','Free Fire Voucher Diamond',1490,197788),
    ('Game','Free Fire','VGFF1795','Free Fire 1795 Diamond','Free Fire Voucher Diamond',1795,236278),
    ('Game','Free Fire','VGFF1800','Free Fire 1800 Diamond','Free Fire Voucher Diamond',1800,237097),
    ('Game','Free Fire','VGFF2000','Free Fire 2000 Diamond','Free Fire Voucher Diamond',2000,263311),
    ('Game','Free Fire','VGFF2140','Free Fire 2140 Diamond','Free Fire Voucher Diamond',2140,279048),
    ('Game','Free Fire','VGFF2180','Free Fire 2180 Diamond','Free Fire Voucher Diamond',2180,282972),
    ('Game','Free Fire','VGFF2210','Free Fire 2210 Diamond','Free Fire Voucher Diamond',2210,289525),
    ('Game','Free Fire','VGFF2280','Free Fire 2280 Diamond','Free Fire Voucher Diamond',2280,298536),
    ('Game','Free Fire','VGFF2355','Free Fire 2355 Diamond','Free Fire Voucher Diamond',2355,307548),
    ('Game','Free Fire','VGFF2720','Free Fire 2720 Diamond','Free Fire Voucher Diamond',2720,353422),
    ('Game','Free Fire','VGFF3640','Free Fire 3640 Diamond','Free Fire Voucher Diamond',3640,468109),
    ('Game','Free Fire','VGFF4000','Free Fire 4000 Diamond','Free Fire Voucher Diamond',4000,513984),
    ('Game','Free Fire','VGFFCEK','Free Fire Cek Nick Name','Free Fire Utility',1,1005),
    ('Game','Free Fire','VGFFMM','Free Fire Member Mingguan 50 Diamond Perhari Selama 7 Hari','Free Fire Membership',50,30100),
    ('Game','Free Fire','VGFFMB','Free Fire Member Bulanan 50 Diamond Perhari Selama 30 Hari','Free Fire Membership',50,86000)
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
