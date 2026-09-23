ALTER TABLE public.kategori_fee_app
  DROP CONSTRAINT IF EXISTS chk_kategori_fee_app_fee_user_nonneg,
  DROP CONSTRAINT IF EXISTS chk_kategori_fee_app_fee_non_user_nonneg;
