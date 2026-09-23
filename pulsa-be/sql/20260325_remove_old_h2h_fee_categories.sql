DELETE FROM public.kategori
WHERE nama IN (
  'Gopay Bebas Nominal',
  'OVO Bebas Nominal',
  'ShopeePay Bebas Nominal',
  'Linkaja Bebas Nominal',
  'Dana Bebas Nominal'
)
AND NOT EXISTS (
  SELECT 1
  FROM public.member_fee_kategori mfk
  WHERE mfk.kategori_id = public.kategori.id
);
