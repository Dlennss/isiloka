-- Seed produk paket data Telkomsel dari daftar provider.
-- Aman dijalankan berulang: produk lama Telkomsel data generik dinonaktifkan, SKU baru di-upsert.

INSERT INTO public.brand (nama, aktif, dibuat_pada, diubah_pada)
SELECT 'Telkomsel', true, now(), now()
WHERE NOT EXISTS (
  SELECT 1 FROM public.brand WHERE lower(trim(nama)) = 'telkomsel'
);

UPDATE public.produk p
SET aktif = false, diubah_pada = now()
FROM public.kategori k, public.brand b
WHERE p.kategori_id = k.id
  AND p.brand_id = b.id
  AND lower(trim(k.nama)) = 'paket data'
  AND lower(trim(b.nama)) = 'telkomsel'
  AND p.sku LIKE 'PK-DATA-TELKOMSEL-%';

WITH seed(category_name, brand_name, sku, product_name, group_name, nominal, harga) AS (
  VALUES
    ('Paket Data','Telkomsel','BYROMNI','BAYAR OMNI','telkomsel',0,1300),
    ('Paket Data','Telkomsel','CEKOMNI','CEK OMNI','telkomsel',0,1000),

    ('Paket Data','Telkomsel','SDBQ10','Telkomsel Data Big Quota All Zona All Net 10GB 30 Hari','TELKOMSEL DATA BIG QUOTA',10000,52500),
    ('Paket Data','Telkomsel','SDBQ20','Telkomsel Data Big Quota All Zona All Net 20GB 30 Hari','TELKOMSEL DATA BIG QUOTA',20000,91500),
    ('Paket Data','Telkomsel','SDBQ25','Telkomsel Data Big Quota All Zona All Net 25GB 30 Hari','TELKOMSEL DATA BIG QUOTA',25000,96500),
    ('Paket Data','Telkomsel','SDBQ30','Telkomsel Data Big Quota All Zona All Net 30GB 30 Hari','TELKOMSEL DATA BIG QUOTA',30000,101500),
    ('Paket Data','Telkomsel','SDBQ35','Telkomsel Data Big Quota All Zona All Net 35GB 30 Hari','TELKOMSEL DATA BIG QUOTA',35000,121000),

    ('Paket Data','Telkomsel','SDEO100','Telkomsel Data 9,5GB + 2GB - 18GB (Zona) + 2GB OMG 30 Hari','telkomsel',9500,73500),
    ('Paket Data','Telkomsel','SDEO150','Telkomsel Data 17,5GB + 2GB - 32GB (Zona) + 2GB OMG 30 Hari','telkomsel',17500,121000),
    ('Paket Data','Telkomsel','SDEO200','Telkomsel Data 39GB + 2GB - 63GB (Zona) + 2GB OMG 30 Hari','telkomsel',39000,163000),
    ('Paket Data','Telkomsel','SDEO30','Telkomsel Data 0,7GB - 2GB (Zona) + 1GB OMG 7 Hari','telkomsel',700,27000),
    ('Paket Data','Telkomsel','SDEO50','Telkomsel Data 2,5GB + 1GB - 5,5GB (Zona) + 1GB OMG 30 Hari','telkomsel',2500,49500),
    ('Paket Data','Telkomsel','SDEO75','Telkomsel Data 3,5GB + 2GB - 8,5GB (Zona) + 2GB OMG 30 Hari','telkomsel',3500,54000),

    ('Paket Data','Telkomsel','SDFN10','Telkomsel Data Nasional AllNet 10GB 30 Hari','TELKOMSEL DATA BULANAN VIP',10000,48800),
    ('Paket Data','Telkomsel','SDFN2','Telkomsel Data Nasional AllNet 2GB 30 Hari','TELKOMSEL DATA BULANAN VIP',2000,25500),
    ('Paket Data','Telkomsel','SDFN20','Telkomsel Data Nasional AllNet 20GB 30 Hari','TELKOMSEL DATA BULANAN VIP',20000,86000),
    ('Paket Data','Telkomsel','SDFN25','Telkomsel Data Nasional AllNet 25GB 30 Hari','TELKOMSEL DATA BULANAN VIP',25000,96000),
    ('Paket Data','Telkomsel','SDFN3','Telkomsel Data Nasional AllNet 3GB 30 Hari','TELKOMSEL DATA BULANAN VIP',3000,29000),
    ('Paket Data','Telkomsel','SDFN30','Telkomsel Data Nasional AllNet 30GB 30 Hari','TELKOMSEL DATA BULANAN VIP',30000,101000),
    ('Paket Data','Telkomsel','SDFN35','Telkomsel Data Nasional AllNet 35GB 30 Hari','TELKOMSEL DATA BULANAN VIP',35000,116000),
    ('Paket Data','Telkomsel','SDFN6','Telkomsel Data Nasional AllNet 6GB 30 Hari','TELKOMSEL DATA BULANAN VIP',6000,47500),
    ('Paket Data','Telkomsel','SDFN7','Telkomsel Data Nasional AllNet 7GB 30 Hari','TELKOMSEL DATA BULANAN VIP',7000,47800),
    ('Paket Data','Telkomsel','SDFN8','Telkomsel Data Nasional AllNet 8GB 30 Hari','TELKOMSEL DATA BULANAN VIP',8000,48500),
    ('Paket Data','Telkomsel','SDFN9','Telkomsel Data Nasional AllNet 9GB 30 Hari','TELKOMSEL DATA BULANAN VIP',9000,47000),

    ('Paket Data','Telkomsel','SDFP1','Telkomsel Data 1GB 30 Hari','TELKOMSEL DATA BULANAN STANDAR',1000,12000),
    ('Paket Data','Telkomsel','SDFP10','Telkomsel Data 10GB 30 Hari','TELKOMSEL DATA BULANAN STANDAR',10000,47000),
    ('Paket Data','Telkomsel','SDFP11','Telkomsel Data 11GB 30 Hari','TELKOMSEL DATA BULANAN STANDAR',11000,63000),
    ('Paket Data','Telkomsel','SDFP12','Telkomsel Data 12GB 30 Hari','TELKOMSEL DATA BULANAN STANDAR',12000,64000),
    ('Paket Data','Telkomsel','SDFP15','Telkomsel Data 15GB 30 Hari','TELKOMSEL DATA BULANAN STANDAR',15000,66000),
    ('Paket Data','Telkomsel','SDFP2','Telkomsel Data 2GB 30 Hari','TELKOMSEL DATA BULANAN STANDAR',2000,22700),
    ('Paket Data','Telkomsel','SDFP20','Telkomsel Data 20GB 30 Hari','TELKOMSEL DATA BULANAN STANDAR',20000,81000),
    ('Paket Data','Telkomsel','SDFP25','Telkomsel Data 25GB 30 Hari','TELKOMSEL DATA BULANAN STANDAR',25000,91000),
    ('Paket Data','Telkomsel','SDFP3','Telkomsel Data 3GB 30 Hari','TELKOMSEL DATA BULANAN STANDAR',3000,28500),
    ('Paket Data','Telkomsel','SDFP30','Telkomsel Data 30GB 30 Hari','TELKOMSEL DATA BULANAN STANDAR',30000,101000),
    ('Paket Data','Telkomsel','SDFP6','Telkomsel Data 6GB 30 Hari','TELKOMSEL DATA BULANAN STANDAR',6000,38000),
    ('Paket Data','Telkomsel','SDFP7','Telkomsel Data 7GB 30 Hari','TELKOMSEL DATA BULANAN STANDAR',7000,41000),
    ('Paket Data','Telkomsel','SDFP8','Telkomsel Data 8GB 30 Hari','TELKOMSEL DATA BULANAN STANDAR',8000,46500),
    ('Paket Data','Telkomsel','SDFP9','Telkomsel Data 9GB 30 Hari','TELKOMSEL DATA BULANAN STANDAR',9000,46800),

    ('Paket Data','Telkomsel','SDHAP11','Telkomsel Data Mini 1GB 1 Hari','TELKOMSEL DATA MINI',1000,9350),
    ('Paket Data','Telkomsel','SDHAP142','Telkomsel Data Mini 2GB 14 Hari','TELKOMSEL DATA MINI',2000,20500),
    ('Paket Data','Telkomsel','SDHAP15','Telkomsel Data Mini 5GB 1 Hari','telkomsel',5000,15500),
    ('Paket Data','Telkomsel','SDHAP150','Telkomsel Data Mini 500MB 15 Hari','TELKOMSEL DATA MINI',500,7450),
    ('Paket Data','Telkomsel','SDHAP151','Telkomsel Data Mini 1GB 15 Hari','TELKOMSEL DATA MINI',1000,11800),
    ('Paket Data','Telkomsel','SDHAP152','Telkomsel Data Mini 2GB 15 Hari','TELKOMSEL DATA MINI',2000,20800),
    ('Paket Data','Telkomsel','SDHAP153','Telkomsel Data Mini 3GB 15 Hari','TELKOMSEL DATA MINI',3000,24000),
    ('Paket Data','Telkomsel','SDHAP31','Telkomsel Data Mini 1GB 3 Hari','TELKOMSEL DATA MINI',1000,10600),
    ('Paket Data','Telkomsel','SDHAP32','Telkomsel Data Mini 2GB 3 Hari','TELKOMSEL DATA MINI',2000,12300),
    ('Paket Data','Telkomsel','SDHAP325','Telkomsel Data Mini 2,5GB 3 Hari','TELKOMSEL DATA MINI',2500,12500),
    ('Paket Data','Telkomsel','SDHAP33','Telkomsel Data Mini 3GB 3 Hari','TELKOMSEL DATA MINI',3000,13400),
    ('Paket Data','Telkomsel','SDHAP52','Telkomsel Data Mini 2,5GB 5 Hari','TELKOMSEL DATA MINI',2500,12475),
    ('Paket Data','Telkomsel','SDHAP53','Telkomsel Data Mini 3GB 5 Hari','TELKOMSEL DATA MINI',3000,13800),
    ('Paket Data','Telkomsel','SDHAP69','Telkomsel Data Mini 250MB 7 Hari','TELKOMSEL DATA MINI',250,5700),
    ('Paket Data','Telkomsel','SDHAP70','Telkomsel Data Mini 750GB 7 Hari','telkomsel',750000,11150),
    ('Paket Data','Telkomsel','SDHAP71','Telkomsel Data Mini 1GB 7 Hari','TELKOMSEL DATA MINI',1000,12000),
    ('Paket Data','Telkomsel','SDHAP72','Telkomsel Data Mini 2GB 7 Hari','TELKOMSEL DATA MINI',2000,16050),
    ('Paket Data','Telkomsel','SDHAP73','Telkomsel Data Mini 3GB 7 Hari','TELKOMSEL DATA MINI',3000,19300),
    ('Paket Data','Telkomsel','SDHAP74','Telkomsel Data Mini 4GB Nasional (2GB AllNet + 2GB 5G) 7 Hari','telkomsel',4000,17000),
    ('Paket Data','Telkomsel','SDHAP77','Telkomsel Data Mini 7GB 7 Hari','TELKOMSEL DATA MINI',7000,27200),

    ('Paket Data','Telkomsel','SDOM110','Telkomsel Data 9,5GB + 2GB - 18GB (Zona) + 2GB OMG 30 Hari','TELKOMSEL DATA OMG',9500,104100),
    ('Paket Data','Telkomsel','SDOM160','Telkomsel Data 17,5GB + 2GB - 32GB (Zona) + 2GB OMG 30 Hari','TELKOMSEL DATA OMG',17500,141750),
    ('Paket Data','Telkomsel','SDOM200','Telkomsel Data 39GB + 2GB - 63GB (Zona) + 2GB OMG 30 Hari','TELKOMSEL DATA OMG',39000,189250),
    ('Paket Data','Telkomsel','SDOM30','Telkomsel Data 0,7GB - 2GB (Zona) + 1GB OMG 7 Hari','TELKOMSEL DATA OMG',700,28550),
    ('Paket Data','Telkomsel','SDOM75','Telkomsel Data 3,5GB + 2GB - 8,5GB (Zona) + 2GB OMG 30 Hari','TELKOMSEL DATA OMG',3500,54000),

    ('Paket Data','Telkomsel','SDRC14','Telkomsel Data China 14GB + 1GB Transit 3 Hari','telkomsel',15000,226000),
    ('Paket Data','Telkomsel','SDRC149','Telkomsel Data China 149GB + 1GB Transit 30 Hari','telkomsel',150000,636000),
    ('Paket Data','Telkomsel','SDRC29','Telkomsel Data China 29GB + 1GB Transit 5 Hari','telkomsel',30000,276000),
    ('Paket Data','Telkomsel','SDRC49','Telkomsel Data China 49GB + 1GB Transit 7 Hari','telkomsel',50000,326000),
    ('Paket Data','Telkomsel','SDRC5','Telkomsel Data China 5GB 1Hari','telkomsel',5000,126399),
    ('Paket Data','Telkomsel','SDRC99','Telkomsel Data China 99GB + 1GB Transit 15 Hari','telkomsel',100000,506000),

    ('Paket Data','Telkomsel','SDRH14','Telkomsel Data Hotspot Hongkong Macau 14GB + Negara Transit 1GB 15Hari','telkomsel',15000,311000),
    ('Paket Data','Telkomsel','SDRH143','Telkomsel Data Hongkong Macau 14GB + Negara Transit 1GB 3Hari','telkomsel',15000,121000),
    ('Paket Data','Telkomsel','SDRH149','Telkomsel Data Hongkong Macau 149GB + Negara Transit 1GB 30Hari','telkomsel',150000,373500),
    ('Paket Data','Telkomsel','SDRH49','Telkomsel Data Hongkong Macau 49GB + Negara Transit 1GB 7Hari','telkomsel',50000,206000),
    ('Paket Data','Telkomsel','SDRH5','Telkomsel Data Hongkong Macau 5GB 1Hari','telkomsel',5000,78500),
    ('Paket Data','Telkomsel','SDRH99','Telkomsel Data Hongkong Macau 99GB + Negara Transit 1GB 15Hari','telkomsel',100000,291000),

    ('Paket Data','Telkomsel','SDRJ124','Telkomsel Data Jepang 124GB + 1GB Transit 20 Hari','telkomsel',125000,501000),
    ('Paket Data','Telkomsel','SDRJ149','Telkomsel Data Jepang 149GB + 1GB Transit 30 hari','telkomsel',150000,551000),
    ('Paket Data','Telkomsel','SDRJ49','Telkomsel Data Jepang 49GB + 1GB Transit 7 Hari','telkomsel',50000,296000),
    ('Paket Data','Telkomsel','SDRJ5','Telkomsel Data Jepang 5GB 1Hari','telkomsel',5000,126000),

    ('Paket Data','Telkomsel','SDRK29','Telkomsel Data Korea Selatan 29GB + 1GB Transit 30 Hari','telkomsel',30000,606000),
    ('Paket Data','Telkomsel','SDRK49','Telkomsel Data Korea Selatan 49GB + 1GB Transit 7 Hari','telkomsel',50000,306000),
    ('Paket Data','Telkomsel','SDRK99','Telkomsel Data Korea Selatan 99GB + 1GB Transit 15 Hari','telkomsel',100000,463500),

    ('Paket Data','Telkomsel','SDRM150','Telkomsel Data Malaysia 150GB 7 Hari','telkomsel',150000,163500),
    ('Paket Data','Telkomsel','SDRM200','Telkomsel Data Malaysia 200GB 10 Hari','telkomsel',200000,206000),
    ('Paket Data','Telkomsel','SDRM25','Telkomsel Data Malaysia 25GB 1Hari','telkomsel',25000,76500),
    ('Paket Data','Telkomsel','SDRM250','Telkomsel Data Malaysia 250GB 15 Hari','telkomsel',250000,248500),
    ('Paket Data','Telkomsel','SDRM300','Telkomsel Data Malaysia 300GB 20 Hari','telkomsel',300000,381000),
    ('Paket Data','Telkomsel','SDRM50','Telkomsel Data Malaysia 50GB 3 Hari','telkomsel',50000,121000),
    ('Paket Data','Telkomsel','SDRM75','Telkomsel Data Malaysia 75GB 5 Hari','telkomsel',75000,142250),

    ('Paket Data','Telkomsel','SDRS150','Telkomsel Data Singapura 150GB 7 Hari','telkomsel',150000,163500),
    ('Paket Data','Telkomsel','SDRS200','Telkomsel Data Singapura 200GB 10 Hari','telkomsel',200000,206000),
    ('Paket Data','Telkomsel','SDRS25','Telkomsel Data Singapura 25GB 1Hari','telkomsel',25000,73000),
    ('Paket Data','Telkomsel','SDRS250','Telkomsel Data Singapura 250GB 15 Hari','telkomsel',250000,248500),
    ('Paket Data','Telkomsel','SDRS400','Telkomsel Data Singapura 400GB 30 Hari','telkomsel',400000,336590),
    ('Paket Data','Telkomsel','SDRS50','Telkomsel Data Singapura 50GB 3 Hari','telkomsel',50000,116000),
    ('Paket Data','Telkomsel','SDRS75','Telkomsel Data Singapura 75GB 5 Hari','telkomsel',75000,141000),

    ('Paket Data','Telkomsel','SDRT1','Telkomsel Data Turki 1GB 1Hari','telkomsel',1000,57499),
    ('Paket Data','Telkomsel','SDRT2','Telkomsel Data Turki 1,5GB + Negara Transit 1GB 3Hari','telkomsel',2500,132499),
    ('Paket Data','Telkomsel','SDRT4','Telkomsel Data Turki 4GB + Negara Transit 1GB 7Hari','telkomsel',5000,182100),
    ('Paket Data','Telkomsel','SDRT9','Telkomsel Data Turki 9GB + Negara Transit 1GB 15Hari','telkomsel',10000,372100),

    ('Paket Data','Telkomsel','SDRTH14','Telkomsel Data Thailand 14GB + 1GB Transit 3 Hari','telkomsel',15000,121000),
    ('Paket Data','Telkomsel','SDRTH149','Telkomsel Data Thailand 149GB + 1GB Transit 30 Hari','telkomsel',150000,333500),
    ('Paket Data','Telkomsel','SDRTH19','Telkomsel Data Thailand 19GB + 1GB Transit 5 Hari','telkomsel',20000,146000),
    ('Paket Data','Telkomsel','SDRTH29','Telkomsel Data Thailand 29GB + 1GB Transit 7 Hari','telkomsel',30000,186000),
    ('Paket Data','Telkomsel','SDRTH5','Telkomsel Data Thailand 5GB 1Hari','telkomsel',5000,81000),
    ('Paket Data','Telkomsel','SDRTH74','Telkomsel Data Thailand 74GB + 1GB Transit 10 Hari','telkomsel',75000,206000),
    ('Paket Data','Telkomsel','SDRTH99','Telkomsel Data Thailand 99GB + 1GB Transit 15 Hari','telkomsel',100000,248500),

    ('Paket Data','Telkomsel','SDRTW14','Telkomsel Data Taiwan 14GB + Negara Transit 1GB 7Hari','telkomsel',15000,176982),
    ('Paket Data','Telkomsel','SDRTW24','Telkomsel Data Taiwan 24GB + Negara Transit 1GB 15Hari','telkomsel',25000,271884),
    ('Paket Data','Telkomsel','SDRTW39','Telkomsel Data Taiwan 39GB + Negara Transit 1GB 30Hari','telkomsel',40000,371100),
    ('Paket Data','Telkomsel','SDRTW5','Telkomsel Data Taiwan 5GB 1Hari','telkomsel',5000,86000),
    ('Paket Data','Telkomsel','SDRTW9','Telkomsel Data Taiwan 9GB + Negara Transit 1GB 3Hari','telkomsel',10000,129530),

    ('Paket Data','Telkomsel','SDRV14','Telkomsel Data Vietnam 14GB + 1GB Transit 3 Hari','telkomsel',15000,163500),
    ('Paket Data','Telkomsel','SDRV19','Telkomsel Data Vietnam 19GB + 1GB Transit 15 Hari','telkomsel',20000,256000),
    ('Paket Data','Telkomsel','SDRV2','Telkomsel Data Vietnam 2GB 1Hari','telkomsel',2000,71500),
    ('Paket Data','Telkomsel','SDRV29','Telkomsel Data Vietnam 29GB + 1GB Transit 30 Hari','telkomsel',30000,346000),
    ('Paket Data','Telkomsel','SDRV4','Telkomsel Data Vietnam 4GB + 1GB Transit 3 Hari','telkomsel',5000,117000),

    ('Paket Data','Telkomsel','SVDIYB309','Voucher Telkomsel 9GB Jateng-DIY 30 Hari','VCR TELKOMSEL JATENG-DIY',9000,43775),
    ('Paket Data','Telkomsel','SVDIYB326','Voucher Telkomsel 26GB Jateng-DIY 30 Hari','VCR TELKOMSEL JATENG-DIY',26000,66775),
    ('Paket Data','Telkomsel','SVDIYM31','Voucher Telkomsel 1,5GB Jateng-DIY 3 Hari','VCR TELKOMSEL JATENG-DIY',1500,9650),
    ('Paket Data','Telkomsel','SVDIYM32','Voucher Telkomsel 2GB Jateng-DIY 3 Hari','VCR TELKOMSEL JATENG-DIY',2000,13825),
    ('Paket Data','Telkomsel','SVDIYM33','Voucher Telkomsel 3GB Jateng-DIY 3 Hari','VCR TELKOMSEL JATENG-DIY',3000,15250),
    ('Paket Data','Telkomsel','SVDIYM52','Voucher Telkomsel 2,5GB Jateng-DIY 5 Hari','VCR TELKOMSEL JATENG-DIY',2500,14600),
    ('Paket Data','Telkomsel','SVDIYM54','Voucher Telkomsel 4,5GB Jateng-DIY 5 Hari','VCR TELKOMSEL JATENG-DIY',4500,20400),
    ('Paket Data','Telkomsel','SVDIYM55','Voucher Telkomsel 5,5GB Jateng-DIY 5 Hari','VCR TELKOMSEL JATENG-DIY',5500,23800),
    ('Paket Data','Telkomsel','SVDIYM73','Voucher Telkomsel 3,5GB Jateng-DIY 7 Hari','VCR TELKOMSEL JATENG-DIY',3500,20550),
    ('Paket Data','Telkomsel','SVDIYM77','Voucher Telkomsel 7GB Jateng-DIY 7 Hari','VCR TELKOMSEL JATENG-DIY',7000,28155),

    ('Paket Data','Telkomsel','SVDKAL31','Voucher Telkomsel 500MB + 1GB Kalimantan (Zona 1) 1 Hari','VCR TELKOMSEL KALIMANTAN',1500,13450),
    ('Paket Data','Telkomsel','SVDKAL33','Voucher Telkomsel 0,5GB Kalimantan (Zona 3) 3 Hari','VCR TELKOMSEL KALIMANTAN',500,13450),
    ('Paket Data','Telkomsel','SVDKAL51','Voucher Telkomsel 1GB + 1GB Kalimantan (Zona 1) 5 Hari','VCR TELKOMSEL KALIMANTAN',2000,16000),
    ('Paket Data','Telkomsel','SVDKAL52','Voucher Telkomsel 1GB + 500MB Kalimantan (Zona 2) 5 Hari','VCR TELKOMSEL KALIMANTAN',1500,16200),
    ('Paket Data','Telkomsel','SVDKAL53','Voucher Telkomsel 1GB Kalimantan (Zona 3) 5 Hari','VCR TELKOMSEL KALIMANTAN',1000,15050),

    ('Paket Data','Telkomsel','SVJABB305','Voucher Telkomsel 5GB Jabodetabek 30 Hari','VCR TELKOMSEL JABODETABEK',5000,34900),
    ('Paket Data','Telkomsel','SVJABB309','Voucher Telkomsel 9GB Jabodetabek 30 Hari','VCR TELKOMSEL JABODETABEK',9000,42870),
    ('Paket Data','Telkomsel','SVJABB318','Voucher Telkomsel 18GB Jabodetabek 30 Hari','VCR TELKOMSEL JABODETABEK',18000,52835)
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
