CREATE OR REPLACE FUNCTION public.fn_enforce_saldo_on_status_change()
 RETURNS trigger
 LANGUAGE plpgsql
AS $function$
DECLARE
  has_debit BOOLEAN := FALSE;
  has_refund BOOLEAN := FALSE;
  has_redebit BOOLEAN := FALSE;
  saldo_skrg BIGINT;
  biaya BIGINT;
BEGIN
  IF TG_OP = 'UPDATE' AND OLD.status = 'success' AND NEW.status = 'failed' THEN
    IF COALESCE(current_setting('p24.admin_manual_cancel_success', TRUE), '') <> '1'
      OR COALESCE(NEW.dibatalkan_oleh_admin_id, 0) <= 0
      OR NULLIF(TRIM(COALESCE(NEW.alasan_batal_admin, '')), '') IS NULL
      OR LOWER(TRIM(COALESCE(NEW.keterangan, ''))) NOT LIKE 'dibatalkan admin:%'
    THEN
      RAISE EXCEPTION 'transaksi_member success immutable: ref_id=% old_status=% new_status=%', NEW.ref_id, OLD.status, NEW.status;
    END IF;
  END IF;

  biaya := COALESCE(NEW.biaya_perkiraan, 0);
  IF biaya <= 0 THEN
    RETURN NEW;
  END IF;

  IF NEW.status = 'pending' THEN
    SELECT saldo
      INTO saldo_skrg
      FROM public.dompet_member
     WHERE member_id = NEW.member_id
     FOR UPDATE;

    IF saldo_skrg IS NULL THEN
      RAISE EXCEPTION 'dompet member tidak ditemukan: member_id=%', NEW.member_id;
    END IF;

    SELECT EXISTS(
      SELECT 1
        FROM public.mutasi_dompet
       WHERE member_id = NEW.member_id
         AND ref_id = NEW.ref_id
         AND UPPER(COALESCE(arah, '')) = 'DEBIT'
         AND UPPER(COALESCE(alasan, '')) = 'TRX_HOLD'
    ) INTO has_debit;

    IF NOT has_debit THEN
      IF saldo_skrg < biaya THEN
        RAISE EXCEPTION 'saldo tidak cukup: saldo=% biaya=%', saldo_skrg, biaya;
      END IF;

      INSERT INTO public.mutasi_dompet
        (member_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, dibuat_pada)
      VALUES
        (NEW.member_id, NEW.ref_id, 'DEBIT', biaya, 'TRX_HOLD', 'hold', saldo_skrg, saldo_skrg - biaya, NOW());

      UPDATE public.dompet_member
         SET saldo = saldo - biaya
       WHERE member_id = NEW.member_id;
    END IF;

    RETURN NEW;
  END IF;

  IF NEW.status = 'failed' AND TG_OP = 'UPDATE' AND OLD.status <> 'failed' THEN
    SELECT EXISTS(
      SELECT 1
        FROM public.mutasi_dompet
       WHERE member_id = NEW.member_id
         AND ref_id = NEW.ref_id
         AND UPPER(COALESCE(arah, '')) = 'DEBIT'
         AND UPPER(COALESCE(alasan, '')) = 'TRX_HOLD'
    ) INTO has_debit;

    IF NOT has_debit THEN
      RETURN NEW;
    END IF;

    SELECT EXISTS(
      SELECT 1
        FROM public.mutasi_dompet
       WHERE member_id = NEW.member_id
         AND ref_id = NEW.ref_id
         AND UPPER(COALESCE(arah, '')) IN ('CREDIT', 'KREDIT')
         AND UPPER(COALESCE(alasan, '')) = 'REFUND'
    ) INTO has_refund;

    IF has_refund THEN
      RETURN NEW;
    END IF;

    SELECT saldo
      INTO saldo_skrg
      FROM public.dompet_member
     WHERE member_id = NEW.member_id
     FOR UPDATE;

    INSERT INTO public.mutasi_dompet
      (member_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, dibuat_pada)
    VALUES
      (NEW.member_id, NEW.ref_id, 'CREDIT', biaya, 'REFUND', 'refund', saldo_skrg, saldo_skrg + biaya, NOW());

    UPDATE public.dompet_member
       SET saldo = saldo + biaya
     WHERE member_id = NEW.member_id;

    RETURN NEW;
  END IF;

  IF NEW.status = 'success' AND TG_OP = 'UPDATE' AND OLD.status <> 'success' THEN
    SELECT EXISTS(
      SELECT 1
        FROM public.mutasi_dompet
       WHERE member_id = NEW.member_id
         AND ref_id = NEW.ref_id
         AND UPPER(COALESCE(arah, '')) IN ('CREDIT', 'KREDIT')
         AND UPPER(COALESCE(alasan, '')) = 'REFUND'
    ) INTO has_refund;

    IF has_refund THEN
      SELECT EXISTS(
        SELECT 1
          FROM public.mutasi_dompet
         WHERE member_id = NEW.member_id
           AND ref_id = NEW.ref_id
           AND UPPER(COALESCE(arah, '')) = 'DEBIT'
           AND UPPER(COALESCE(alasan, '')) IN ('TRX_SETTLE', 'CALLBACK_SUCCESS_RECOVERY')
      ) INTO has_redebit;

      IF NOT has_redebit THEN
        SELECT saldo
          INTO saldo_skrg
          FROM public.dompet_member
         WHERE member_id = NEW.member_id
         FOR UPDATE;

        IF saldo_skrg >= biaya THEN
          INSERT INTO public.mutasi_dompet
            (member_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, dibuat_pada)
          VALUES
            (NEW.member_id, NEW.ref_id, 'DEBIT', biaya, 'TRX_SETTLE', 'potong ulang setelah refund - transaksi sukses', saldo_skrg, saldo_skrg - biaya, NOW());

          UPDATE public.dompet_member
             SET saldo = saldo - biaya
           WHERE member_id = NEW.member_id;
        ELSE
          RAISE WARNING 'saldo tidak cukup untuk re-debit: ref_id=% saldo=% biaya=%', NEW.ref_id, saldo_skrg, biaya;
        END IF;
      END IF;
    END IF;
  END IF;

  RETURN NEW;
END;
$function$;
