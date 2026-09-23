DO $$
DECLARE
  v_bank_id bigint;
BEGIN
  SELECT id
    INTO v_bank_id
  FROM public.bank
  WHERE regexp_replace(COALESCE(nomor_rekening, ''), '[^0-9]', '', 'g') = '8761518267'
  ORDER BY id ASC
  LIMIT 1;

  IF v_bank_id IS NULL THEN
    INSERT INTO public.bank
      (nama, nomor_rekening, atas_nama, saldo, aktif, admin_staff_only, dibuat_pada, diubah_pada)
    VALUES
      ('BCA H2H', '8761518267', 'LISA OKTARIA', 0, true, true, now(), now());
  ELSE
    UPDATE public.bank
    SET nama = 'BCA H2H',
        nomor_rekening = '8761518267',
        atas_nama = 'LISA OKTARIA',
        aktif = true,
        admin_staff_only = true,
        diubah_pada = now()
    WHERE id = v_bank_id;
  END IF;
END $$;
