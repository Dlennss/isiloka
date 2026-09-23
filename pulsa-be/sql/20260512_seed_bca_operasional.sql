DO $$
DECLARE
  v_bank_id bigint;
  v_before bigint;
  v_after bigint := 1000000;
  v_delta bigint;
  v_direction text;
  v_reason text;
BEGIN
  SELECT id, saldo
  INTO v_bank_id, v_before
  FROM public.bank
  WHERE trim(nomor_rekening) = '3432738881'
  ORDER BY id
  LIMIT 1
  FOR UPDATE;

  IF v_bank_id IS NULL THEN
    PERFORM setval(
      pg_get_serial_sequence('public.bank', 'id'),
      GREATEST((SELECT COALESCE(MAX(id), 0) FROM public.bank), 1),
      true
    );

    INSERT INTO public.bank
      (nama, nomor_rekening, atas_nama, saldo, aktif, admin_staff_only, dibuat_pada, diubah_pada)
    VALUES
      ('BCA OPERASIONAL', '3432738881', 'PULSA MITRA NASIONAL', v_after, true, true, now(), now())
    RETURNING id INTO v_bank_id;

    INSERT INTO public.mutasi_bank
      (bank_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, diubah_oleh, dibuat_pada)
    VALUES
      (v_bank_id, 'BOPEN-OPR-' || to_char(now(), 'YYYYMMDDHH24MISS'), 'CREDIT', v_after, 'BANK_OPENING_BALANCE', 'Saldo awal rekening operasional', 0, v_after, NULL, now());
  ELSE
    UPDATE public.bank
    SET nama = 'BCA OPERASIONAL',
        nomor_rekening = '3432738881',
        atas_nama = 'PULSA MITRA NASIONAL',
        saldo = v_after,
        aktif = true,
        admin_staff_only = true,
        diubah_pada = now()
    WHERE id = v_bank_id;

    IF COALESCE(v_before, 0) <> v_after THEN
      v_delta := abs(v_after - COALESCE(v_before, 0));
      v_direction := CASE WHEN v_after > COALESCE(v_before, 0) THEN 'CREDIT' ELSE 'DEBIT' END;
      v_reason := CASE WHEN v_direction = 'CREDIT' THEN 'BANK_ADJUST_CREDIT' ELSE 'BANK_ADJUST_DEBIT' END;

      INSERT INTO public.mutasi_bank
        (bank_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, diubah_oleh, dibuat_pada)
      VALUES
        (v_bank_id, 'BADJ-OPR-' || to_char(now(), 'YYYYMMDDHH24MISS'), v_direction, v_delta, v_reason, 'Set saldo rekening operasional', COALESCE(v_before, 0), v_after, NULL, now());
    END IF;
  END IF;
END $$;
