BEGIN;

CREATE TABLE IF NOT EXISTS public.maintenance_statuspay_recovery_refund_20260606 (
  backup_at timestamptz NOT NULL DEFAULT now(),
  trx_member_id bigint NOT NULL,
  member_id bigint NOT NULL,
  ref_id text NOT NULL,
  status text NOT NULL,
  recovery_debit bigint NOT NULL,
  saldo_sebelum bigint NOT NULL,
  saldo_sesudah bigint NOT NULL,
  inserted_mutasi_dompet_id bigint
);

DO $$
DECLARE
  r record;
  saldo_before bigint;
  saldo_after bigint;
  inserted_id bigint;
BEGIN
  FOR r IN
    SELECT
      tm.id AS trx_member_id,
      tm.member_id,
      tm.ref_id,
      tm.status,
      SUM(md.jumlah)::bigint AS recovery_debit
    FROM public.transaksi_member tm
    JOIN public.mutasi_dompet md
      ON md.member_id = tm.member_id
     AND md.ref_id = tm.ref_id
     AND UPPER(COALESCE(md.arah, '')) = 'DEBIT'
     AND UPPER(COALESCE(md.alasan, '')) = 'CALLBACK_SUCCESS_RECOVERY'
    WHERE tm.ref_id IN ('meie8b68cc9e8', 'meie8b8afa258', 'junieb89b05a22')
      AND LOWER(COALESCE(tm.status, '')) = 'failed'
      AND NOT EXISTS (
        SELECT 1
        FROM public.mutasi_dompet x
        WHERE x.member_id = tm.member_id
          AND x.ref_id = tm.ref_id
          AND UPPER(COALESCE(x.arah, '')) IN ('CREDIT', 'KREDIT')
          AND UPPER(COALESCE(x.alasan, '')) = 'STATUS_PAY_RECOVERY_REFUND'
      )
    GROUP BY tm.id, tm.member_id, tm.ref_id, tm.status
    HAVING SUM(md.jumlah) > 0
    ORDER BY tm.ref_id
  LOOP
    SELECT saldo
      INTO saldo_before
      FROM public.dompet_member
     WHERE member_id = r.member_id
     FOR UPDATE;

    IF saldo_before IS NULL THEN
      RAISE EXCEPTION 'dompet member tidak ditemukan: member_id=%', r.member_id;
    END IF;

    saldo_after := saldo_before + r.recovery_debit;

    INSERT INTO public.mutasi_dompet
      (member_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, dibuat_pada)
    VALUES
      (
        r.member_id,
        r.ref_id,
        'CREDIT',
        r.recovery_debit,
        'STATUS_PAY_RECOVERY_REFUND',
        'refund ulang debit CALLBACK_SUCCESS_RECOVERY setelah perbaikan STATUS-PAY',
        saldo_before,
        saldo_after,
        now()
      )
    RETURNING id INTO inserted_id;

    UPDATE public.dompet_member
       SET saldo = saldo_after,
           diperbarui_pada = now()
     WHERE member_id = r.member_id;

    INSERT INTO public.maintenance_statuspay_recovery_refund_20260606
      (trx_member_id, member_id, ref_id, status, recovery_debit, saldo_sebelum, saldo_sesudah, inserted_mutasi_dompet_id)
    VALUES
      (r.trx_member_id, r.member_id, r.ref_id, r.status, r.recovery_debit, saldo_before, saldo_after, inserted_id);
  END LOOP;
END $$;

COMMIT;
