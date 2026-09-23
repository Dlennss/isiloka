DO $$
DECLARE
  v_bank_id bigint;
  v_removed_amount bigint := 0;
BEGIN
  SELECT id
  INTO v_bank_id
  FROM public.bank
  WHERE trim(nomor_rekening) = '3432738881'
  ORDER BY id
  LIMIT 1
  FOR UPDATE;

  IF v_bank_id IS NULL THEN
    RETURN;
  END IF;

  SELECT COALESCE(SUM(jumlah), 0)
  INTO v_removed_amount
  FROM public.mutasi_bank
  WHERE bank_id = v_bank_id
    AND arah = 'CREDIT'
    AND alasan = 'BANK_OPENING_BALANCE'
    AND catatan = 'Saldo awal rekening operasional'
    AND ref_id LIKE 'BOPEN-OPR-%';

  DELETE FROM public.mutasi_bank
  WHERE bank_id = v_bank_id
    AND arah = 'CREDIT'
    AND alasan = 'BANK_OPENING_BALANCE'
    AND catatan = 'Saldo awal rekening operasional'
    AND ref_id LIKE 'BOPEN-OPR-%';

  IF v_removed_amount > 0 THEN
    UPDATE public.bank
    SET saldo = GREATEST(0, saldo - v_removed_amount),
        diubah_pada = now()
    WHERE id = v_bank_id;
  END IF;
END $$;
