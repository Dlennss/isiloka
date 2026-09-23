INSERT INTO public.kategori (nama, aktif, dibuat_pada, diubah_pada)
SELECT 'Bebas Nominal', false, now(), now()
WHERE NOT EXISTS (
  SELECT 1 FROM public.kategori WHERE LOWER(nama) = LOWER('Bebas Nominal')
);

INSERT INTO public.kategori_fee_app
  (kategori_id, fee_master, fee_agent, fee_user, fee_non_user, aktif, created_at, updated_at)
SELECT
  k_special.id,
  COALESCE(kfa_emoney.fee_master, 0),
  COALESCE(kfa_emoney.fee_agent, 0),
  COALESCE(kfa_emoney.fee_user, 0),
  COALESCE(kfa_emoney.fee_non_user, 0),
  true,
  now(),
  now()
FROM public.kategori k_special
LEFT JOIN public.kategori k_emoney
  ON LOWER(k_emoney.nama) = LOWER('E-Money')
LEFT JOIN public.kategori_fee_app kfa_emoney
  ON kfa_emoney.kategori_id = k_emoney.id
 AND kfa_emoney.aktif = true
WHERE LOWER(k_special.nama) = LOWER('Bebas Nominal')
  AND NOT EXISTS (
    SELECT 1
    FROM public.kategori_fee_app existing
    WHERE existing.kategori_id = k_special.id
  );
