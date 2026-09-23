ALTER TABLE public.kategori_fee_app
  ADD COLUMN IF NOT EXISTS fee_master bigint NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS fee_agent bigint NOT NULL DEFAULT 0;

UPDATE public.kategori_fee_app
SET fee_master = fee_user
WHERE fee_master = 0;

UPDATE public.kategori_fee_app
SET fee_agent = fee_user
WHERE fee_agent = 0;

UPDATE public.kategori SET nama = 'Pulsa' WHERE id = 1;
UPDATE public.kategori SET nama = 'E-Money' WHERE id = 2;
UPDATE public.kategori SET nama = 'Internet Pascabayar' WHERE id = 3;
UPDATE public.kategori SET nama = 'Game' WHERE id = 4;
UPDATE public.kategori SET nama = 'Paket Data' WHERE id = 5;
UPDATE public.kategori SET nama = 'Lainnya' WHERE id = 6;
UPDATE public.kategori SET nama = 'TV' WHERE id = 7;
UPDATE public.kategori SET nama = 'Aktivasi Perdana' WHERE id = 8;
UPDATE public.kategori SET nama = 'Masa Aktif' WHERE id = 9;
UPDATE public.kategori SET nama = 'Paket Telepon' WHERE id = 10;
UPDATE public.kategori SET nama = 'Listrik' WHERE id = 11;

UPDATE public.produk
SET kategori_id = 2
WHERE kategori_id IN (12, 13, 14, 15, 16);

UPDATE public.kategori
SET aktif = false
WHERE id IN (12, 13, 14, 15, 16);

INSERT INTO public.kategori (nama, aktif, dibuat_pada, diubah_pada)
SELECT x.nama, true, now(), now()
FROM (
  VALUES
    ('PDAM'),
    ('BPJS'),
    ('HP Pascabayar'),
    ('Gas Negara')
) AS x(nama)
WHERE NOT EXISTS (
  SELECT 1 FROM public.kategori k WHERE lower(trim(k.nama)) = lower(trim(x.nama))
);

INSERT INTO public.kategori_fee_app (kategori_id, fee_master, fee_agent, fee_user, fee_non_user, aktif, created_at, updated_at)
SELECT k.id, x.fee_user, x.fee_user, x.fee_user, x.fee_non_user, true, now(), now()
FROM (
  VALUES
    ('Pulsa', 150::bigint, 200::bigint),
    ('E-Money', 1000::bigint, 1500::bigint),
    ('Game', 1000::bigint, 1500::bigint),
    ('Paket Data', 1000::bigint, 1500::bigint),
    ('TV', 1500::bigint, 2000::bigint),
    ('Listrik', 500::bigint, 1000::bigint),
    ('Masa Aktif', 500::bigint, 1000::bigint),
    ('Aktivasi Perdana', 1000::bigint, 1500::bigint),
    ('Paket Telepon', 500::bigint, 1000::bigint),
    ('Internet Pascabayar', 1500::bigint, 2000::bigint),
    ('PDAM', 1500::bigint, 2000::bigint),
    ('BPJS', 1500::bigint, 2000::bigint),
    ('HP Pascabayar', 1500::bigint, 2000::bigint),
    ('Gas Negara', 1500::bigint, 2000::bigint),
    ('Lainnya', 1000::bigint, 1500::bigint)
) AS x(nama, fee_user, fee_non_user)
JOIN public.kategori k
  ON lower(trim(k.nama)) = lower(trim(x.nama))
LEFT JOIN public.kategori_fee_app a
  ON a.kategori_id = k.id
WHERE a.id IS NULL;
