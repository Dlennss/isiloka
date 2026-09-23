-- Seed produk paket data Indosat dari daftar provider.
-- Aman dijalankan berulang: produk lama Indosat data generik dinonaktifkan, SKU baru di-upsert.

INSERT INTO public.brand (nama, aktif, dibuat_pada, diubah_pada)
SELECT 'Indosat', true, now(), now()
WHERE NOT EXISTS (
  SELECT 1 FROM public.brand WHERE lower(trim(nama)) = 'indosat'
);

UPDATE public.produk p
SET aktif = false, diubah_pada = now()
FROM public.kategori k, public.brand b
WHERE p.kategori_id = k.id
  AND p.brand_id = b.id
  AND lower(trim(k.nama)) = 'paket data'
  AND lower(trim(b.nama)) = 'indosat'
  AND p.sku LIKE 'PK-DATA-INDOSAT-%';

WITH seed(category_name, brand_name, sku, product_name, group_name, nominal, harga) AS (
  VALUES
    ('Paket Data','Indosat','AVICSUM1','AKT. VCR ISAT FREEDOM 1,5GB 1HR LOKAL SUMATERA TENGAH','AKT VCR ISAT SUMATERA TENGAH',1500,7900),
    ('Paket Data','Indosat','AVICSUM2','AKT. VCR ISAT FREEDOM 5GB 2HR LOKAL SUMATERA TENGAH','AKT VCR ISAT SUMATERA TENGAH',5000,9650),
    ('Paket Data','Indosat','AVICSUM3','AKT. VCR ISAT FREEDOM 3GB 3HR LOKAL SUMATERA TENGAH','AKT VCR ISAT SUMATERA TENGAH',3000,13075),
    ('Paket Data','Indosat','AVICSUM4','AKT. VCR ISAT FREEDOM 5GB 3HR LOKAL SUMATERA TENGAH','AKT VCR ISAT SUMATERA TENGAH',5000,14350),
    ('Paket Data','Indosat','AVICSUM6','AKT. VCR ISAT FREEDOM 4GB 5HR LOKAL SUMATERA TENGAH','AKT VCR ISAT SUMATERA TENGAH',4000,15350),
    ('Paket Data','Indosat','AVICSUM7','AKT. VCR ISAT FREEDOM 6GB 5HR LOKAL SUMATERA TENGAH','AKT VCR ISAT SUMATERA TENGAH',6000,18650),
    ('Paket Data','Indosat','AVICSUM8','AKT. VCR ISAT FREEDOM 7GB 7HR LOKAL SUMATERA TENGAH','AKT VCR ISAT SUMATERA TENGAH',7000,24300),
    ('Paket Data','Indosat','AVICSUM9','AKT. VCR ISAT FREEDOM 15GB 7HR LOKAL SUMATERA TENGAH','AKT VCR ISAT SUMATERA TENGAH',15000,30100),
    ('Paket Data','Indosat','AVIJBK1','AKT. VCR ISAT FREEDOM 1,5GB 1HR LOKAL JABODETABEK DSK','AKT VCR ISAT JABODETABEK DSK',1500,8000),
    ('Paket Data','Indosat','AVIJBK2','AKT. VCR ISAT FREEDOM 3GB 3HR LOKAL JABODETABEK DSK','AKT VCR ISAT JABODETABEK DSK',3000,13000),
    ('Paket Data','Indosat','AVIJBK3','AKT. VCR ISAT FREEDOM 2.5GB 5HR LOKAL JABODETABEK DSK','AKT VCR ISAT JABODETABEK DSK',2500,18000),
    ('Paket Data','Indosat','AVIJBK4','AKT. VCR ISAT FREEDOM 4GB 5HR LOKAL JABODETABEK DSK','AKT VCR ISAT JABODETABEK DSK',4000,14100),
    ('Paket Data','Indosat','AVIJBK5','AKT. VCR ISAT FREEDOM 6GB 5HR LOKAL JABODETABEK DSK','AKT VCR ISAT JABODETABEK DSK',6000,15800),
    ('Paket Data','Indosat','AVIJBK6','AKT. VCR ISAT FREEDOM 7GB 7HR LOKAL JABODETABEK DSK','AKT VCR ISAT JABODETABEK DSK',7000,23800),
    ('Paket Data','Indosat','AVIJBK7','AKT. VCR ISAT FREEDOM 15GB 7HR LOKAL JABODETABEK DSK','AKT VCR ISAT JABODETABEK DSK',15000,29000),
    ('Paket Data','Indosat','AVIJBK8','AKT. VCR ISAT FREEDOM 5GB 2HR LOKAL JABODETABEK DSK','AKT VCR ISAT JABODETABEK DSK',5000,10800),
    ('Paket Data','Indosat','AVINSUM1','AKT. VCR ISAT FREEDOM 1,5GB 1HR LOKAL SUMATERA UTARA','AKT VCR ISAT SUMATERA UTARA',1500,8100),
    ('Paket Data','Indosat','AVINSUM2','AKT. VCR ISAT FREEDOM 5GB 2HR LOKAL SUMATERA UTARA','AKT VCR ISAT SUMATERA UTARA',5000,10550),
    ('Paket Data','Indosat','AVINSUM3','AKT. VCR ISAT FREEDOM 3GB 3HR LOKAL SUMATERA UTARA','AKT VCR ISAT SUMATERA UTARA',3000,13200),
    ('Paket Data','Indosat','AVINSUM4','AKT. VCR ISAT FREEDOM 5GB 3HR LOKAL SUMATERA UTARA','AKT VCR ISAT SUMATERA UTARA',5000,14350),
    ('Paket Data','Indosat','AVINSUM6','AKT. VCR ISAT FREEDOM 3,5GB 5HR LOKAL SUMATERA UTARA','AKT VCR ISAT SUMATERA UTARA',3500,15720),
    ('Paket Data','Indosat','AVINSUM7','AKT. VCR ISAT FREEDOM 5GB 5HR LOKAL SUMATERA UTARA','AKT VCR ISAT SUMATERA UTARA',5000,18800),
    ('Paket Data','Indosat','AVINSUM8','AKT. VCR ISAT FREEDOM 7GB 7HR LOKAL SUMATERA UTARA','AKT VCR ISAT SUMATERA UTARA',7000,24200),
    ('Paket Data','Indosat','AVINSUM9','AKT. VCR ISAT FREEDOM 15GB 7HR LOKAL SUMATERA UTARA','AKT VCR ISAT SUMATERA UTARA',15000,30050),
    ('Paket Data','Indosat','AVISSUM1','AKT. VCR ISAT FREEDOM 1,5GB 1HR LOKAL SUMATERA SELATAN','AKT VCR ISAT SUMATERA SELATAN',1500,6770),
    ('Paket Data','Indosat','AVISSUM2','AKT. VCR ISAT FREEDOM 5GB 2HR LOKAL SUMATERA SELATAN','AKT VCR ISAT SUMATERA SELATAN',5000,9630),
    ('Paket Data','Indosat','AVISSUM3','AKT. VCR ISAT FREEDOM 3GB 3HR LOKAL SUMATERA SELATAN','AKT VCR ISAT SUMATERA SELATAN',3000,13100),
    ('Paket Data','Indosat','AVISSUM4','AKT. VCR ISAT FREEDOM 5GB 3HR LOKAL SUMATERA SELATAN','AKT VCR ISAT SUMATERA SELATAN',5000,14325),
    ('Paket Data','Indosat','AVISSUM5','AKT. VCR ISAT FREEDOM 2,5GB 5HR LOKAL SUMATERA SELATAN','AKT VCR ISAT SUMATERA SELATAN',2500,14800),
    ('Paket Data','Indosat','AVISSUM6','AKT. VCR ISAT FREEDOM 4GB 5HR LOKAL SUMATERA SELATAN','AKT VCR ISAT SUMATERA SELATAN',4000,15275),
    ('Paket Data','Indosat','AVISSUM7','AKT. VCR ISAT FREEDOM 6GB 5HR LOKAL SUMATERA SELATAN','AKT VCR ISAT SUMATERA SELATAN',6000,18550),
    ('Paket Data','Indosat','AVISSUM8','AKT. VCR ISAT FREEDOM 7GB 7HR LOKAL SUMATERA SELATAN','AKT VCR ISAT SUMATERA SELATAN',7000,24250),
    ('Paket Data','Indosat','AVISSUM9','AKT. VCR ISAT FREEDOM 15GB 7HR LOKAL SUMATERA SELATAN','AKT VCR ISAT SUMATERA SELATAN',15000,29975),

    ('Paket Data','Indosat','IDFAC1','Indosat Data 1,5GB (1GB Utama + 500MB Apps) 1 Hari','INDOSAT DATA MINI',1500,7750),
    ('Paket Data','Indosat','IDFAC13','Indosat Data 13GB (7GB Utama + 6GB Apps) 30 Hari','INDOSAT DATA MINI',13000,43700),
    ('Paket Data','Indosat','IDFAC6','Indosat Data 6GB (4GB Utama + 2GB Apps) 7 Hari','INDOSAT DATA MINI',6000,19900),
    ('Paket Data','Indosat','IDFIC14','Indosat Data Freedom 14GB 14 Hari','INDOSAT DATA FREEDOM HARIAN',14000,47250),
    ('Paket Data','Indosat','IDFIC15','Indosat Data Freedom 5GB 1 Hari','INDOSAT DATA FREEDOM HARIAN',5000,7550),
    ('Paket Data','Indosat','IDFIC21','Indosat Data Freedom 1GB 2 Hari','INDOSAT DATA FREEDOM HARIAN',1000,8600),
    ('Paket Data','Indosat','IDFIC22','Indosat Data Freedom 22GB 14 Hari','INDOSAT DATA FREEDOM HARIAN',22000,48000),
    ('Paket Data','Indosat','IDFIC28','Indosat Data Freedom 30GB 30 Hari (1GB/Hari)','INDOSAT DATA FREEDOM HARIAN',30000,69000),
    ('Paket Data','Indosat','IDFIC33','Indosat Data Freedom 3,5GB 3 Hari','INDOSAT DATA FREEDOM HARIAN',3500,12710),
    ('Paket Data','Indosat','IDFIC36','Indosat Data Freedom 6GB 3 Hari','INDOSAT DATA FREEDOM HARIAN',6000,13570),
    ('Paket Data','Indosat','IDFIC54','Indosat Data Freedom 4GB 5 Hari','INDOSAT DATA FREEDOM HARIAN',4000,13870),
    ('Paket Data','Indosat','IDFIC58','Indosat Data Freedom 8GB 5 Hari','INDOSAT DATA FREEDOM HARIAN',8000,17635),
    ('Paket Data','Indosat','IDFIC59','Indosat Data Freedom 9GB 5 Hari','INDOSAT DATA FREEDOM HARIAN',9000,17640),
    ('Paket Data','Indosat','IDFIC7','Indosat Data Freedom 7GB 14 Hari','INDOSAT DATA FREEDOM HARIAN',7000,21960),
    ('Paket Data','Indosat','IDFIC717','Indosat Data Freedom 17GB 7 Hari','INDOSAT DATA FREEDOM HARIAN',17000,28400),
    ('Paket Data','Indosat','IDFIC718','Indosat Data Freedom 18GB 7 Hari','INDOSAT DATA FREEDOM HARIAN',18000,34000),
    ('Paket Data','Indosat','IDFIC79','Indosat Data Freedom 9GB 7 Hari','INDOSAT DATA FREEDOM HARIAN',9000,23285),
    ('Paket Data','Indosat','IDFMU1','Indosat Data Freedom Unlimited 1 GB + 4,5 GB Apps 30 Hari','INDOSAT DATA FREEDOM UNL VIP',5500,36500),
    ('Paket Data','Indosat','IDFMU10','Indosat Data Freedom Unlimited 10 GB + 25 GB Apps 30 Hari','INDOSAT DATA FREEDOM UNL VIP',35000,113650),
    ('Paket Data','Indosat','IDFMU2','Indosat Data Freedom Unlimited 2 GB + 7 GB Apps 30 Hari','INDOSAT DATA FREEDOM UNL VIP',9000,57950),
    ('Paket Data','Indosat','IDFMU3','Indosat Data Freedom Unlimited 3 GB + 17 GB Apps 30 Hari','INDOSAT DATA FREEDOM UNL VIP',20000,80650),
    ('Paket Data','Indosat','IDFMU7','Indosat Data Freedom Unlimited 7 GB + 20 GB Apps 30 Hari','INDOSAT DATA FREEDOM UNL VIP',27000,102850),
    ('Paket Data','Indosat','IDFP10','Indosat Data Freedom 10 GB 28 Hari','INDOSAT DATA FREEDOM',10000,35900),
    ('Paket Data','Indosat','IDFP12','Indosat Data Freedom 12 GB 28 Hari','INDOSAT DATA FREEDOM',12000,44750),
    ('Paket Data','Indosat','IDFP14','Indosat Data Freedom 14 GB 28 Hari','INDOSAT DATA FREEDOM',14000,44775),
    ('Paket Data','Indosat','IDFP150','Indosat Data Freedom 150 GB 28 Hari','INDOSAT DATA FREEDOM',150000,120200),
    ('Paket Data','Indosat','IDFP18','Indosat Data Freedom 18 GB 28 hari','INDOSAT DATA FREEDOM',18000,51600),
    ('Paket Data','Indosat','IDFP20','Indosat Data Freedom 20 GB 28 Hari','INDOSAT DATA FREEDOM',20000,54425),
    ('Paket Data','Indosat','IDFP22','Indosat Data Freedom 22 GB 28 Hari','INDOSAT DATA FREEDOM',22000,54475),
    ('Paket Data','Indosat','IDFP25','Indosat Data Freedom 25 GB 28 Hari','INDOSAT DATA FREEDOM',25000,62600),
    ('Paket Data','Indosat','IDFP28','Indosat Data Freedom 28 GB 28 Hari','INDOSAT DATA FREEDOM',28000,62650),
    ('Paket Data','Indosat','IDFP30','Indosat Data Freedom 30 GB 28 Hari','INDOSAT DATA FREEDOM',30000,62700),
    ('Paket Data','Indosat','IDFP35','Indosat Data Freedom 35 GB 28 Hari','INDOSAT DATA FREEDOM',35000,69400),
    ('Paket Data','Indosat','IDFP4','Indosat Data Freedom 4 GB 20 Hari','INDOSAT DATA FREEDOM',4000,31025),
    ('Paket Data','Indosat','IDFP5','Indosat Data Freedom 5,5 GB 28 Hari','INDOSAT DATA FREEDOM',5500,31100),
    ('Paket Data','Indosat','IDFP50','Indosat Data Freedom 50 GB 28 Hari','INDOSAT DATA FREEDOM',50000,88000),
    ('Paket Data','Indosat','IDFP6','Indosat Data Freedom 6,5 GB 28 Hari','INDOSAT DATA FREEDOM',6500,31125),
    ('Paket Data','Indosat','IDFP7','Indosat Data Freedom 7 GB 28 Hari','INDOSAT DATA FREEDOM',7000,31150),
    ('Paket Data','Indosat','IDFP8','Indosat Data Freedom 8 GB 28 Hari','INDOSAT DATA FREEDOM',8000,31175),
    ('Paket Data','Indosat','IDFP80','Indosat Data Freedom 80 GB 28 Hari','INDOSAT DATA FREEDOM',80000,106025),
    ('Paket Data','Indosat','IDFP9','Indosat Data Freedom 9 GB 28 Hari','INDOSAT DATA FREEDOM',9000,33950),
    ('Paket Data','Indosat','IDFR150','Freedom internet Ramadhan 150 GB ( 5GB/Hari ) 30Hari','INDOSAT DATA FREEDOM RAMADHAN',150000,128350),
    ('Paket Data','Indosat','IDFR60','Freedom internet Ramadhan 60 GB ( 2GB/Hari ) 30Hari','INDOSAT DATA FREEDOM RAMADHAN',60000,83800),
    ('Paket Data','Indosat','IDFUN1','Indosat Data Freedom Unlimited 1 GB + 4,5 GB Apps 30 Hari','INDOSAT DATA FREEDOM UNL',5500,33150),
    ('Paket Data','Indosat','IDFUN10','Indosat Data Freedom Unlimited 10 GB + 25 GB Apps 30 Hari','INDOSAT DATA FREEDOM UNL',35000,112400),
    ('Paket Data','Indosat','IDFUN16','Indosat Data Freedom Unlimited 1 GB + 2 GB Apps 6 Hari','INDOSAT DATA FREEDOM UNL',3000,16950),
    ('Paket Data','Indosat','IDFUN2','Indosat Data Freedom Unlimited 2 GB + 8 GB Apps 30 Hari','INDOSAT DATA FREEDOM UNL',10000,52400),
    ('Paket Data','Indosat','IDFUN3','Indosat Data Freedom Unlimited 3 GB + 17 GB Apps 30 Hari','INDOSAT DATA FREEDOM UNL',20000,75100),
    ('Paket Data','Indosat','IDFUN7','Indosat Data Freedom Unlimited 7 GB + 28 GB Apps 30 Hari','INDOSAT DATA FREEDOM UNL',35000,101800),
    ('Paket Data','Indosat','IDHF125','Indosat Modem HIFI Air 125GB 30 Hari','INDOSAT DATA HIFI',125000,88800),
    ('Paket Data','Indosat','IDHF20','Indosat HIFI Booster 20GB','INDOSAT DATA HIFI',20000,22350),
    ('Paket Data','Indosat','IDHF200','Indosat Modem HIFI Air 200GB 30 Hari','INDOSAT DATA HIFI',200000,133350),
    ('Paket Data','Indosat','IDHF50','Indosat HIFI Booster 50GB','INDOSAT DATA HIFI',50000,44700),
    ('Paket Data','Indosat','IDHF500','Indosat Modem HIFI Air 500GB 30 Hari','INDOSAT DATA HIFI',500000,222100),
    ('Paket Data','Indosat','IDHF75','Indosat Modem HIFI Air 75GB 30 Hari','INDOSAT DATA HIFI',75000,66800),
    ('Paket Data','Indosat','IDPB100','Indosat Data Pure 10GB 30 Hari','INDOSAT DATA PURE',10000,57650),
    ('Paket Data','Indosat','IDPB40','Indosat Data Pure 4GB 30 Hari','INDOSAT DATA PURE',4000,23650),
    ('Paket Data','Indosat','IDPB50','Indosat Data Pure 5GB 30 Hari','INDOSAT DATA PURE',5000,29350),
    ('Paket Data','Indosat','IDPB60','Indosat Data Pure 6GB 30 Hari','INDOSAT DATA PURE',6000,35000),
    ('Paket Data','Indosat','IDPB70','Indosat Data Pure 7GB 30 Hari','INDOSAT DATA PURE',7000,40650),
    ('Paket Data','Indosat','IDPB80','Indosat Data Pure 8GB 30 Hari','INDOSAT DATA PURE',8000,46300),
    ('Paket Data','Indosat','IDPB90','Indosat Data Pure 9GB 30 Hari','INDOSAT DATA PURE',9000,52000),
    ('Paket Data','Indosat','IDSATS150','Indosat Paket SATSPAM+ 150GB 28Hari','indosat',150000,128200),
    ('Paket Data','Indosat','IDSATS20','Indosat Paket SATSPAM+ 20GB 3Hari','indosat',20000,17750),
    ('Paket Data','Indosat','IDSATS30','Indosat Paket SATSPAM+ 30GB 5Hari','indosat',30000,26750),
    ('Paket Data','Indosat','IDSATS300','Indosat Paket SATSPAM+ 300GB 28Hari','indosat',300000,161200),
    ('Paket Data','Indosat','IDSATS75','Indosat Paket SATSPAM+ 75GB 7Hari','indosat',75000,44500),
    ('Paket Data','Indosat','IDT15','Indosat Data Freedom 5GB 1 Hari','INDOSAT DATA SUPER PROMO',5000,7600),
    ('Paket Data','Indosat','IDT17','Indosat Data Freedom 7GB 1 Hari','INDOSAT DATA SUPER PROMO',7000,8700),
    ('Paket Data','Indosat','IDT26','Indosat Data Freedom 6GB 2 Hari','INDOSAT DATA SUPER PROMO',6000,9900),
    ('Paket Data','Indosat','IDT36','Indosat Data Freedom 6GB 3 Hari','indosat',6000,13375),
    ('Paket Data','Indosat','IDT59','Indosat Data Freedom 9GB 5 Hari','INDOSAT DATA SUPER PROMO',9000,17400),
    ('Paket Data','Indosat','IDUH12','New Umroh Haji 50GB + 1GB 12 Hari','NEW INDOSAT DATA UMROH HAJI',51000,323500),
    ('Paket Data','Indosat','IDUH15','New Umroh Haji 70GB + 1GB 15 Hari','NEW INDOSAT DATA UMROH HAJI',71000,402000),
    ('Paket Data','Indosat','IDUH30','New Umroh Haji 25GB + 1GB 30 Hari','NEW INDOSAT DATA UMROH HAJI',26000,669000),
    ('Paket Data','Indosat','IDUH45','New Umroh Haji 30GB + 1GB 45 Hari','NEW INDOSAT DATA UMROH HAJI',31000,747500),
    ('Paket Data','Indosat','IDUM14','Indosat Data Umroh Haji 14GB + 1GB 15 hari','INDOSAT DATA UMROH HAJI',15000,276500),
    ('Paket Data','Indosat','IDUM19','Indosat Data Umroh Haji 19GB + 1GB 30 hari','INDOSAT DATA UMROH HAJI',20000,512000),
    ('Paket Data','Indosat','IDUM24','Indosat Data Umroh Haji 24GB + 1GB 45 hari','INDOSAT DATA UMROH HAJI',25000,590500),
    ('Paket Data','Indosat','IDUM6','Indosat Data Umroh Haji 6GB + 1GB 12 hari','INDOSAT DATA UMROH HAJI',7000,198000),
    ('Paket Data','Indosat','IDY1','Indosat Data 1 GB 1 Hari','INDOSAT DATA YELLOW',1000,6625),
    ('Paket Data','Indosat','IDY15','Indosat Data 1 GB 15 Hari','INDOSAT DATA YELLOW',1000,9500),
    ('Paket Data','Indosat','IDY3','Indosat Data 1 GB 3 Hari','INDOSAT DATA YELLOW',1000,7000),
    ('Paket Data','Indosat','IDY7','Indosat Data 1 GB 7 Hari','INDOSAT DATA YELLOW',1000,8700)
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
