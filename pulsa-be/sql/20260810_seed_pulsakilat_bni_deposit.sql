INSERT INTO public.bank
  (nama, nomor_rekening, atas_nama, saldo, aktif, admin_staff_only, dibuat_pada, diubah_pada)
VALUES
  ('BNI', '1955637480', 'M Sansan Irfanda', 0, true, false, now(), now())
ON CONFLICT ((lower(trim(nama)))) DO UPDATE
SET nomor_rekening = EXCLUDED.nomor_rekening,
    atas_nama = EXCLUDED.atas_nama,
    aktif = true,
    admin_staff_only = false,
    diubah_pada = now();
