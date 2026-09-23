-- Seed produk paket data XL dari daftar provider.
-- Aman dijalankan berulang: produk lama XL data generik dinonaktifkan, SKU baru di-upsert.

INSERT INTO public.brand (nama, aktif, dibuat_pada, diubah_pada)
SELECT 'XL', true, now(), now()
WHERE NOT EXISTS (
  SELECT 1 FROM public.brand WHERE lower(trim(nama)) = 'xl'
);

UPDATE public.produk p
SET aktif = false, diubah_pada = now()
FROM public.kategori k, public.brand b
WHERE p.kategori_id = k.id
  AND p.brand_id = b.id
  AND lower(trim(k.nama)) = 'paket data'
  AND lower(trim(b.nama)) = 'xl'
  AND p.sku LIKE 'PK-DATA-XL-%';

WITH seed(category_name, brand_name, sku, product_name, group_name, nominal, harga) AS (
  VALUES
    ('Paket Data','XL','VXFB1','VCR XL XTRA Combo Flex S 28 Hari','VOUCHER XL COMBO FLEX BULANAN',0,34000),
    ('Paket Data','XL','VXFB2','VCR XL XTRA Combo Flex S+ 28 Hari','VOUCHER XL COMBO FLEX BULANAN',0,37750),
    ('Paket Data','XL','VXFB3','VCR XL XTRA Combo Flex M 28 Hari','VOUCHER XL COMBO FLEX BULANAN',0,50000),
    ('Paket Data','XL','VXFB4','VCR XL XTRA Combo Flex M+ 28 Hari','VOUCHER XL COMBO FLEX BULANAN',0,60600),
    ('Paket Data','XL','VXFB5','VCR XL XTRA Combo Flex L 28 Hari','VOUCHER XL COMBO FLEX BULANAN',0,67200),
    ('Paket Data','XL','VXFB6','VCR XL XTRA Combo Flex L+ 28 Hari','VOUCHER XL COMBO FLEX BULANAN',0,80200),
    ('Paket Data','XL','VXFB7','VCR XL XTRA Combo Flex XL 28 Hari','VOUCHER XL COMBO FLEX BULANAN',0,104700),
    ('Paket Data','XL','VXFB8','VCR XL XTRA Combo Flex XXL 28 Hari','VOUCHER XL COMBO FLEX BULANAN',0,138900),

    ('Paket Data','XL','VXHH1','VCR XL Harian S (2GB-3GB) 3 Hari','VOUCHER XL MINI',2000,10250),
    ('Paket Data','XL','VXHH2','VCR XL Harian M (3GB-4,5GB) 3 Hari','VOUCHER XL MINI',3000,11650),
    ('Paket Data','XL','VXHH3','VCR XL Harian L (5GB-7GB) 3 Hari','VOUCHER XL MINI',5000,14200),
    ('Paket Data','XL','VXHH4','VCR XL Harian S (2,5GB-4GB) 5 Hari','VOUCHER XL MINI',2500,13700),
    ('Paket Data','XL','VXHH5','VCR XL Harian M (3,5GB-5,5GB) 5 Hari','VOUCHER XL MINI',3500,15700),
    ('Paket Data','XL','VXHH6','VCR XL Harian L (6GB-9GB) 5 Hari','VOUCHER XL MINI',6000,18150),

    ('Paket Data','XL','VXHM1','VCR XL Hotrod 1,5GB 3 Hari','VOUCHER XL HOTROD HARIAN',1500,11250),
    ('Paket Data','XL','VXHM2','VCR XL Hotrod 2GB 7 Hari','VOUCHER XL HOTROD HARIAN',2000,14550),
    ('Paket Data','XL','VXHM3','VCR XL Hotrod 3GB 7 Hari','VOUCHER XL HOTROD HARIAN',3000,18550),
    ('Paket Data','XL','VXHM4','VCR XL Hotrod 4,5GB 7 Hari','VOUCHER XL HOTROD HARIAN',4500,23900),
    ('Paket Data','XL','VXHM6','VCR XL Hotrod 6,5GB 7 Hari','VOUCHER XL HOTROD HARIAN',6500,29825),

    ('Paket Data','XL','XD2K10','Bebas Puas 2rb (2,50GB - 4,00GB) 10 Hari','XL DATA BEBAS PUAS 6K',2500,20400),
    ('Paket Data','XL','XD2K11','Bebas Puas 2rb (2,75GB - 4,40GB) 11 Hari','XL DATA BEBAS PUAS 6K',2750,23700),
    ('Paket Data','XL','XD2K12','Bebas Puas 2rb (3,00GB - 4,80GB) 12 Hari','XL DATA BEBAS PUAS 6K',3000,25750),
    ('Paket Data','XL','XD2K13','Bebas Puas 2rb (3,25GB - 5,20GB) 13 Hari','XL DATA BEBAS PUAS 6K',3250,27670),
    ('Paket Data','XL','XD2K14','Bebas Puas 2rb (3,50GB - 5,60GB) 14 Hari','XL DATA BEBAS PUAS 6K',3500,29500),
    ('Paket Data','XL','XD2K15','Bebas Puas 2rb (3,75GB - 6,00GB) 15 Hari','XL DATA BEBAS PUAS 6K',3750,30000),
    ('Paket Data','XL','XD2K16','Bebas Puas 2rb (4,00GB - 6,40GB) 16 Hari','XL DATA BEBAS PUAS 6K',4000,34000),
    ('Paket Data','XL','XD2K17','Bebas Puas 2rb (4,25GB - 6,80GB) 17 Hari','XL DATA BEBAS PUAS 6K',4250,36000),
    ('Paket Data','XL','XD2K18','Bebas Puas 2rb (4,50GB - 7,20GB) 18 Hari','XL DATA BEBAS PUAS 6K',4500,38200),
    ('Paket Data','XL','XD2K19','Bebas Puas 2rb (4,75GB - 7,60GB) 19 Hari','XL DATA BEBAS PUAS 6K',4750,40200),
    ('Paket Data','XL','XD2K20','Bebas Puas 2rb (5,00GB - 8,00GB) 20 Hari','XL DATA BEBAS PUAS 6K',5000,42300),
    ('Paket Data','XL','XD2K21','Bebas Puas 2rb (5,25GB - 8,40GB) 21 Hari','XL DATA BEBAS PUAS 6K',5250,44300),
    ('Paket Data','XL','XD2K22','Bebas Puas 2rb (5,50GB - 8,80GB) 22 Hari','XL DATA BEBAS PUAS 6K',5500,46400),
    ('Paket Data','XL','XD2K23','Bebas Puas 2rb (5,75GB - 9,20GB) 23 Hari','XL DATA BEBAS PUAS 6K',5750,48500),
    ('Paket Data','XL','XD2K24','Bebas Puas 2rb (6,00GB - 9,60GB) 24 Hari','XL DATA BEBAS PUAS 6K',6000,50450),
    ('Paket Data','XL','XD2K25','Bebas Puas 2rb (6,25GB - 10,00GB) 25 Hari','XL DATA BEBAS PUAS 6K',6250,52500),
    ('Paket Data','XL','XD2K26','Bebas Puas 2rb (6,50GB - 10,40GB) 26 Hari','XL DATA BEBAS PUAS 6K',6500,54600),
    ('Paket Data','XL','XD2K27','Bebas Puas 2rb (6,75GB - 10,80GB) 27 Hari','XL DATA BEBAS PUAS 6K',6750,56800),
    ('Paket Data','XL','XD2K28','Bebas Puas 2rb (7,00GB - 11,20GB) 28 Hari','XL DATA BEBAS PUAS 6K',7000,58000),
    ('Paket Data','XL','XD2K29','Bebas Puas 2rb (7,25GB - 11,60GB) 29 Hari','XL DATA BEBAS PUAS 6K',7250,58200),
    ('Paket Data','XL','XD2K30','Bebas Puas 2rb (7,50GB - 12,00GB) 30 Hari','XL DATA BEBAS PUAS 6K',7500,58400),
    ('Paket Data','XL','XD2K6','Bebas Puas 2rb (1,50GB - 2,40GB) 6 Hari','XL DATA BEBAS PUAS 6K',1500,13400),
    ('Paket Data','XL','XD2K7','Bebas Puas 2rb (1,75GB - 2,80GB) 7 Hari','XL DATA BEBAS PUAS 6K',1750,14500),
    ('Paket Data','XL','XD2K8','Bebas Puas 2rb (2,00GB - 3,20GB) 8 Hari','XL DATA BEBAS PUAS 6K',2000,17500),
    ('Paket Data','XL','XD2K9','Bebas Puas 2rb (2,25GB - 3,60GB) 9 Hari','XL DATA BEBAS PUAS 6K',2250,19600),

    ('Paket Data','XL','XD3K10','Bebas Puas 3rb (5,00GB - 10,00GB) 10 Hari','XL DATA BEBAS PUAS 3K',5000,30200),
    ('Paket Data','XL','XD3K11','Bebas Puas 3rb (5,50GB - 11,00GB) 11 Hari','XL DATA BEBAS PUAS 3K',5500,35050),
    ('Paket Data','XL','XD3K12','Bebas Puas 3rb (6,00GB - 12,00GB) 12 Hari','XL DATA BEBAS PUAS 3K',6000,38150),
    ('Paket Data','XL','XD3K13','Bebas Puas 3rb (6,50GB - 13,00GB) 13 Hari','XL DATA BEBAS PUAS 3K',6500,41200),
    ('Paket Data','XL','XD3K14','Bebas Puas 3rb (7,00GB - 14,00GB) 14 Hari','XL DATA BEBAS PUAS 3K',7000,43850),
    ('Paket Data','XL','XD3K15','Bebas Puas 3rb (7,50GB - 15,00GB) 15 Hari','XL DATA BEBAS PUAS 3K',7500,44000),
    ('Paket Data','XL','XD3K16','Bebas Puas 3rb (8,00GB - 16,00GB) 16 Hari','XL DATA BEBAS PUAS 3K',8000,50500),
    ('Paket Data','XL','XD3K17','Bebas Puas 3rb (8,50GB - 17,00GB) 17 Hari','XL DATA BEBAS PUAS 3K',8500,53600),
    ('Paket Data','XL','XD3K18','Bebas Puas 3rb (9,00GB - 18,00GB) 18 Hari','XL DATA BEBAS PUAS 3K',9000,56660),
    ('Paket Data','XL','XD3K19','Bebas Puas 3rb (9,50GB - 19,00GB) 19 Hari','XL DATA BEBAS PUAS 3K',9500,59750),
    ('Paket Data','XL','XD3K20','Bebas Puas 3rb (10,00GB - 20,00GB) 20 Hari','XL DATA BEBAS PUAS 3K',10000,62850),
    ('Paket Data','XL','XD3K21','Bebas Puas 3rb (10,50GB - 21,00GB) 21 Hari','XL DATA BEBAS PUAS 3K',10500,65950),
    ('Paket Data','XL','XD3K22','Bebas Puas 3rb (11,00GB - 22,00GB) 22 Hari','XL DATA BEBAS PUAS 3K',11000,69000),
    ('Paket Data','XL','XD3K23','Bebas Puas 3rb (11,50GB - 23,00GB) 23 Hari','XL DATA BEBAS PUAS 3K',11500,72090),
    ('Paket Data','XL','XD3K24','Bebas Puas 3rb (12,00GB - 24,00GB) 24 Hari','XL DATA BEBAS PUAS 3K',12000,75200),
    ('Paket Data','XL','XD3K25','Bebas Puas 3rb (12,50GB - 25,00GB) 25 Hari','XL DATA BEBAS PUAS 3K',12500,78300),
    ('Paket Data','XL','XD3K26','Bebas Puas 3rb (13,00GB - 26,00GB) 26 Hari','XL DATA BEBAS PUAS 3K',13000,81400),
    ('Paket Data','XL','XD3K27','Bebas Puas 3rb (13,50GB - 27,00GB) 27 Hari','XL DATA BEBAS PUAS 3K',13500,84450),
    ('Paket Data','XL','XD3K28','Bebas Puas 3rb (14,00GB - 28,00GB) 28 Hari','XL DATA BEBAS PUAS 3K',14000,86500),
    ('Paket Data','XL','XD3K29','Bebas Puas 3rb (14,50GB - 29,00GB) 29 Hari','XL DATA BEBAS PUAS 3K',14500,87300),
    ('Paket Data','XL','XD3K3','Bebas Puas 3rb (1,50GB - 3,00GB) 3 Hari','XL DATA BEBAS PUAS 3K',1500,13200),
    ('Paket Data','XL','XD3K30','Bebas Puas 3rb (15,00GB - 30,00GB) 30 Hari','XL DATA BEBAS PUAS 3K',15000,87800),
    ('Paket Data','XL','XD3K4','Bebas Puas 3rb (2,00GB - 4,00GB) 4 Hari','XL DATA BEBAS PUAS 3K',2000,13400),
    ('Paket Data','XL','XD3K5','Bebas Puas 3rb (2,50GB - 5,00GB) 5 Hari','XL DATA BEBAS PUAS 3K',2500,15500),
    ('Paket Data','XL','XD3K6','Bebas Puas 3rb (3,00GB - 6,00GB) 6 Hari','XL DATA BEBAS PUAS 3K',3000,19550),
    ('Paket Data','XL','XD3K7','Bebas Puas 3rb (3,50GB - 7,00GB) 7 Hari','XL DATA BEBAS PUAS 3K',3500,21000),
    ('Paket Data','XL','XD3K8','Bebas Puas 3rb (4,00GB - 8,00GB) 8 Hari','XL DATA BEBAS PUAS 3K',4000,25750),
    ('Paket Data','XL','XD3K9','Bebas Puas 3rb (4,50GB - 9,00GB) 9 Hari','XL DATA BEBAS PUAS 3K',4500,28850),

    ('Paket Data','XL','XD5K1','Bebas Puas 5rb (1,80GB - 3,00GB) 1 Hari','XL DATA BEBAS PUAS 5K',1800,7270),
    ('Paket Data','XL','XD5K10','Bebas Puas 5rb (18,00GB - 30,00GB) 10 Hari','XL DATA BEBAS PUAS 5K',18000,52975),
    ('Paket Data','XL','XD5K11','Bebas Puas 5rb (19,80GB - 33,00GB) 11 Hari','XL DATA BEBAS PUAS 5K',19800,59025),
    ('Paket Data','XL','XD5K12','Bebas Puas 5rb (21,60GB - 36,00GB) 12 Hari','XL DATA BEBAS PUAS 5K',21600,64504),
    ('Paket Data','XL','XD5K13','Bebas Puas 5rb (23,40GB - 39,00GB) 13 Hari','XL DATA BEBAS PUAS 5K',23400,69575),
    ('Paket Data','XL','XD5K14','Bebas Puas 5rb (25,20GB - 42,00GB) 14 Hari','XL DATA BEBAS PUAS 5K',25200,74900),
    ('Paket Data','XL','XD5K15','Bebas Puas 5rb (27,00GB - 45,00GB) 15 Hari','XL DATA BEBAS PUAS 5K',27000,76980),
    ('Paket Data','XL','XD5K16','Bebas Puas 5rb (28,80GB - 48,00GB) 16 Hari','XL DATA BEBAS PUAS 5K',28800,85600),
    ('Paket Data','XL','XD5K17','Bebas Puas 5rb (30,60GB - 51,00GB) 17 Hari','XL DATA BEBAS PUAS 5K',30600,90900),
    ('Paket Data','XL','XD5K18','Bebas Puas 5rb (32,40GB - 54,00GB) 18 Hari','XL DATA BEBAS PUAS 5K',32400,96343),
    ('Paket Data','XL','XD5K19','Bebas Puas 5rb (34,20GB - 57,00GB) 19 Hari','XL DATA BEBAS PUAS 5K',34200,101629),
    ('Paket Data','XL','XD5K2','Bebas Puas 5rb (3,60GB - 6,00GB) 2 Hari','XL DATA BEBAS PUAS 5K',3600,11440),
    ('Paket Data','XL','XD5K20','Bebas Puas 5rb (36,00GB - 60,00GB) 20 Hari','XL DATA BEBAS PUAS 5K',36000,106915),
    ('Paket Data','XL','XD5K21','Bebas Puas 5rb (37,80GB - 63,00GB) 21 Hari','XL DATA BEBAS PUAS 5K',37800,111600),
    ('Paket Data','XL','XD5K22','Bebas Puas 5rb (39,60GB - 66,00GB) 22 Hari','XL DATA BEBAS PUAS 5K',39600,117000),
    ('Paket Data','XL','XD5K23','Bebas Puas 5rb (41,40GB - 69,00GB) 23 Hari','XL DATA BEBAS PUAS 5K',41400,123072),
    ('Paket Data','XL','XD5K24','Bebas Puas 5rb (43,20GB - 72,00GB) 24 Hari','XL DATA BEBAS PUAS 5K',43200,127300),
    ('Paket Data','XL','XD5K25','Bebas Puas 5rb (45,00GB - 75,00GB) 25 Hari','XL DATA BEBAS PUAS 5K',45000,133643),
    ('Paket Data','XL','XD5K26','Bebas Puas 5rb (46,80GB - 78,00GB) 26 Hari','XL DATA BEBAS PUAS 5K',46800,136600),
    ('Paket Data','XL','XD5K27','Bebas Puas 5rb (48,60GB - 81,00GB) 27 Hari','XL DATA BEBAS PUAS 5K',48600,144215),
    ('Paket Data','XL','XD5K28','Bebas Puas 5rb (50,40GB - 84,00GB) 28 Hari','XL DATA BEBAS PUAS 5K',50400,149500),
    ('Paket Data','XL','XD5K29','Bebas Puas 5rb (52,20GB - 87,00GB) 29 Hari','XL DATA BEBAS PUAS 5K',52200,154500),
    ('Paket Data','XL','XD5K3','Bebas Puas 5rb (5,40GB - 9,00GB) 3 Hari','XL DATA BEBAS PUAS 5K',5400,16825),
    ('Paket Data','XL','XD5K30','Bebas Puas 5rb (54,00GB - 90,00GB) 30 Hari','XL DATA BEBAS PUAS 5K',54000,155500),
    ('Paket Data','XL','XD5K4','Bebas Puas 5rb (7,20GB - 12,00GB) 4 Hari','XL DATA BEBAS PUAS 5K',7200,22168),
    ('Paket Data','XL','XD5K5','Bebas Puas 5rb (9,00GB - 15,00GB) 5 Hari','XL DATA BEBAS PUAS 5K',9000,27325),
    ('Paket Data','XL','XD5K6','Bebas Puas 5rb (10,80GB - 18,00GB) 6 Hari','XL DATA BEBAS PUAS 5K',10800,32765),
    ('Paket Data','XL','XD5K7','Bebas Puas 5rb (12,60GB - 21,00GB) 7 Hari','XL DATA BEBAS PUAS 5K',12600,37350),
    ('Paket Data','XL','XD5K8','Bebas Puas 5rb (14,40GB - 24,00GB) 8 Hari','XL DATA BEBAS PUAS 5K',14400,43336),
    ('Paket Data','XL','XD5K9','Bebas Puas 5rb (16,20GB - 27,00GB) 9 Hari','XL DATA BEBAS PUAS 5K',16200,48550),

    ('Paket Data','XL','XD6K1','Bebas Puas 6rb (1,00GB - 2,00GB) 1 Hari','XL DATA BEBAS PUAS 2K',1000,7575),
    ('Paket Data','XL','XD6K10','Bebas Puas 6rb (18,00GB - 30,00GB) 10 Hari','XL DATA BEBAS PUAS 2K',18000,57900),
    ('Paket Data','XL','XD6K15','Bebas Puas 6rb (27,00GB - 45,00GB) 15 Hari','XL DATA BEBAS PUAS 2K',27000,86500),
    ('Paket Data','XL','XD6K3','Bebas Puas 6rb (5,00GB - 8,00GB) 3 Hari','XL DATA BEBAS PUAS 2K',5000,18100),
    ('Paket Data','XL','XD6K30','Bebas Puas 6rb (54,00GB - 90,00GB) 30 Hari','XL DATA BEBAS PUAS 2K',54000,173000),
    ('Paket Data','XL','XD6K5','Bebas Puas 6rb (9,00GB - 15,00GB) 5 Hari','XL DATA BEBAS PUAS 2K',9000,29475),
    ('Paket Data','XL','XD6K7','Bebas Puas 6rb (12,00GB - 20,00GB) 7 Hari','XL DATA BEBAS PUAS 2K',12000,40850),

    ('Paket Data','XL','XDBLC1','XL Data Conference 15GB 7 Hari','XL',15000,11040),
    ('Paket Data','XL','XDBLE1','XL Data Edukasi 15GB 7 Hari','XL',15000,8531),
    ('Paket Data','XL','XDCM1','XL Data Combo Mini 1 GB + 500 MB Youtube + Lokal 7 Hari','XL DATA COMBO MINI',1500,11030),
    ('Paket Data','XL','XDCM2','XL Data Combo Mini 1,5 GB + 1 GB Youtube + Lokal 7 Hari','XL DATA COMBO MINI',2500,15925),
    ('Paket Data','XL','XDCM3','XL Data Combo Mini 2 GB + 2 GB Youtube + Lokal 7 Hari','XL DATA COMBO MINI',4000,19750),
    ('Paket Data','XL','XDCM4','XL Data Combo Mini 3 GB + 3 GB Youtube + Lokal 7 Hari','XL DATA COMBO MINI',6000,25550),

    ('Paket Data','XL','XDFH11','Flex Mini 3GB 1 Hari','XL DATA FLEX MINI',3000,7400),
    ('Paket Data','XL','XDFH12','Flex Mini 10GB 1 Hari','XL DATA FLEX MINI',10000,9800),
    ('Paket Data','XL','XDFH141','Flex Mini 8GB 14 Hari','XL DATA FLEX MINI',8000,27300),
    ('Paket Data','XL','XDFH142','Flex Mini 15GB 14 Hari','XL DATA FLEX MINI',15000,36950),
    ('Paket Data','XL','XDFH143','Flex Mini 30GB 14 Hari','XL DATA FLEX MINI',30000,46500),
    ('Paket Data','XL','XDFH144','Flex Mini 75GB 14 Hari','XL DATA FLEX MINI',75000,69500),
    ('Paket Data','XL','XDFH145','Flex Mini 150GB 14 Hari','XL DATA FLEX MINI',150000,88000),
    ('Paket Data','XL','XDFH31','Flex Mini 4GB 3 Hari','XL DATA FLEX MINI',4000,12430),
    ('Paket Data','XL','XDFH32','Flex Mini 6GB 3 Hari','XL DATA FLEX MINI',6000,14400),
    ('Paket Data','XL','XDFH33','Flex Mini 20GB 3 Hari','XL DATA FLEX MINI',20000,20000),
    ('Paket Data','XL','XDFH71','Flex Mini 4GB 7 Hari','XL DATA FLEX MINI',4000,15625),
    ('Paket Data','XL','XDFH72','Flex Mini 6GB 7 Hari','XL DATA FLEX MINI',6000,19200),
    ('Paket Data','XL','XDFH73','Flex Mini 10GB 7 Hari','XL DATA FLEX MINI',10000,24000),
    ('Paket Data','XL','XDFH74','Flex Mini 20GB 7 Hari','XL DATA FLEX MINI',20000,29000),
    ('Paket Data','XL','XDFH75','Flex Mini 40GB 7 Hari','XL DATA FLEX MINI',40000,37200),
    ('Paket Data','XL','XDFH76','Flex Mini 75GB 7 Hari','XL DATA FLEX MINI',75000,45000),

    ('Paket Data','XL','XDFM10','XL Data Flex Max 10GB 28 Hari','XL DATA COMBO FLEX MAX',10000,37960),
    ('Paket Data','XL','XDFM100','XL Data Flex Max 100GB 28 Hari','XL DATA COMBO FLEX MAX',100000,104950),
    ('Paket Data','XL','XDFM150','XL Data Flex Max 150GB 28 Hari','XL DATA COMBO FLEX MAX',150000,140275),
    ('Paket Data','XL','XDFM16','XL Data Flex Max 16GB 28 Hari','XL DATA COMBO FLEX MAX',16000,46250),
    ('Paket Data','XL','XDFM23','XL Data Flex Max 23GB 28 Hari','XL DATA COMBO FLEX MAX',23000,55500),
    ('Paket Data','XL','XDFM31','XL Data Flex Max 31GB 28 Hari','XL DATA COMBO FLEX MAX',31000,64250),
    ('Paket Data','XL','XDFM40','XL Data Flex Max 40GB 28 Hari','XL DATA COMBO FLEX MAX',40000,70950),
    ('Paket Data','XL','XDFM50','XL Data Flex Max 50GB 28 Hari','XL DATA COMBO FLEX MAX',50000,79738),
    ('Paket Data','XL','XDFM65','XL Data Flex Max 65GB 28 Hari','XL DATA COMBO FLEX MAX',65000,88500),
    ('Paket Data','XL','XDFM7','XL Data Flex Max 7GB 28 Hari','XL DATA COMBO FLEX MAX',7000,33100),

    ('Paket Data','XL','XDKB10','XL Data Akrab Kuota Bersama 10 GB 30 Hari','XL DATA AKRAB KUOTA BERSAMA',10000,62400),
    ('Paket Data','XL','XDKB160','XL Data Akrab Kuota Bersama 160 GB 30 Hari','XL DATA AKRAB KUOTA BERSAMA',160000,485200),
    ('Paket Data','XL','XDKB25','XL Data Akrab Kuota Bersama 25 GB 30 Hari','XL DATA AKRAB KUOTA BERSAMA',25000,119500),
    ('Paket Data','XL','XDKB45','XL Data Akrab Kuota Bersama 45 GB 30 Hari','XL DATA AKRAB KUOTA BERSAMA',45000,194500),
    ('Paket Data','XL','XDKB80','XL Data Akrab Kuota Bersama 80 GB 30 Hari','XL DATA AKRAB KUOTA BERSAMA',80000,302000),

    ('Paket Data','XL','XDM101','Paket Harian XS (2GB + Lokal) 10 Hari','XL DATA PAKET HARIAN',2000,19250),
    ('Paket Data','XL','XDM102','Paket Harian S (3GB + Lokal) 10 Hari','XL DATA PAKET HARIAN',3000,19400),
    ('Paket Data','XL','XDM103','Paket Harian M (4GB + Lokal) 10 Hari','XL DATA PAKET HARIAN',4000,24300),
    ('Paket Data','XL','XDM31','Paket Harian S (1,5GB + Lokal) 3 Hari','XL DATA PAKET HARIAN',1500,12590),
    ('Paket Data','XL','XDM32','Paket Harian M (2,5GB + Lokal) 3 Hari','XL DATA PAKET HARIAN',2500,13100),
    ('Paket Data','XL','XDM33','Paket Harian L (4,5GB + Lokal) 3 Hari','XL DATA PAKET HARIAN',4500,15500),
    ('Paket Data','XL','XDM51','Paket Harian S (2GB + Lokal) 5 Hari','XL DATA PAKET HARIAN',2000,14600),
    ('Paket Data','XL','XDM52','Paket Harian M (3GB + Lokal) 5 Hari','XL DATA PAKET HARIAN',3000,17050),
    ('Paket Data','XL','XDM53','Paket Harian L (5,5GB + Lokal) 5 Hari','XL DATA PAKET HARIAN',5500,19000),
    ('Paket Data','XL','XDM71','Paket Harian XS (2GB + Lokal) 7 Hari','XL DATA PAKET HARIAN',2000,18350),
    ('Paket Data','XL','XDM72','Paket Harian S (3GB + Lokal) 7 Hari','XL DATA PAKET HARIAN',3000,18450),
    ('Paket Data','XL','XDM73','Paket Harian M (4,5GB + Lokal) 7 Hari','XL DATA PAKET HARIAN',4500,21900)
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
