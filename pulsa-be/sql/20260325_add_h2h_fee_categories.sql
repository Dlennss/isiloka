INSERT INTO public.kategori (nama, aktif)
SELECT v.nama, false
FROM (
  VALUES
    ('DANA'),
    ('GOPAY'),
    ('OVO'),
    ('LINKAJA'),
    ('SHOPEEPAY'),
    ('LAINNYA')
) AS v(nama)
WHERE NOT EXISTS (
  SELECT 1
  FROM public.kategori k
  WHERE UPPER(TRIM(k.nama)) = UPPER(TRIM(v.nama))
);
