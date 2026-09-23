ALTER TABLE public.produk_provider_map
ADD COLUMN IF NOT EXISTS mode text;

ALTER TABLE public.produk_provider_map
ADD COLUMN IF NOT EXISTS special_code text;

UPDATE public.produk_provider_map ppm
SET mode = CASE
  WHEN LOWER(TRIM(ppm.provider)) <> 'smb' THEN ppm.mode
  WHEN UPPER(TRIM(p.sku)) IN ('BCA', 'BRI', 'BNI', 'MANDIRI', 'BSI', 'CIMB', 'DANAMON', 'PERMATA', 'BTN', 'OCBC', 'MAYBANK', 'PANIN', 'MEGA', 'BUKOPIN', 'SINARMAS', 'COMMONWEALTH', 'UOB', 'HSBC', 'BTPN', 'MUAMALAT', 'JAGO', 'NEOCOMMERCE', 'SEABANK', 'ALLOBANK', 'BJB') THEN 'DIRECT'
  WHEN UPPER(TRIM(ppm.kode_provider)) IN ('ELDN', 'DANAOPEN', 'GPYOPEN', 'SHPOPEN', 'ELOV', 'ELLI', 'OVOOPEN', 'BIFASTOPEN') THEN 'DIRECT'
  WHEN UPPER(TRIM(ppm.kode_provider)) IN ('DANA', 'OVO', 'GOPAY', 'SHOPEEPAY', 'LINKAJA') THEN 'WALLET_PPOB'
  ELSE 'PPOB'
END
FROM public.produk p
WHERE p.id = ppm.produk_id
  AND ppm.mode IS NULL;

UPDATE public.produk_provider_map ppm
SET special_code = 'BIFASTOPEN'
FROM public.produk p
WHERE p.id = ppm.produk_id
  AND LOWER(TRIM(ppm.provider)) = 'smb'
  AND UPPER(TRIM(p.sku)) IN ('BCA', 'BRI', 'BNI', 'MANDIRI', 'BSI', 'CIMB', 'DANAMON', 'PERMATA', 'BTN', 'OCBC', 'MAYBANK', 'PANIN', 'MEGA', 'BUKOPIN', 'SINARMAS', 'COMMONWEALTH', 'UOB', 'HSBC', 'BTPN', 'MUAMALAT', 'JAGO', 'NEOCOMMERCE', 'SEABANK', 'ALLOBANK', 'BJB')
  AND ppm.special_code IS NULL;
