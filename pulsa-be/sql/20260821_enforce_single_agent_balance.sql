-- Model saldo agent: hanya satu dompet utama.
-- Pinjaman yang disetujui menambah dompet utama, sedangkan agent_credit_loan
-- hanya menyimpan pokok, tagihan, jatuh tempo, dan riwayat pelunasan.

CREATE TABLE IF NOT EXISTS public.agent_credit_main_balance_migration (
  loan_id BIGINT PRIMARY KEY REFERENCES public.agent_credit_loan(id) ON DELETE CASCADE,
  member_id BIGINT NOT NULL REFERENCES public.member(id) ON DELETE CASCADE,
  amount BIGINT NOT NULL CHECK (amount > 0),
  migrated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

DO $$
DECLARE
  loan_rec RECORD;
  saldo_before BIGINT;
  saldo_after BIGINT;
BEGIN
  FOR loan_rec IN
    SELECT l.id, l.application_id, l.member_id, l.available_amount
    FROM public.agent_credit_loan l
    LEFT JOIN public.agent_credit_main_balance_migration m ON m.loan_id = l.id
    WHERE m.loan_id IS NULL
      AND l.available_amount > 0
      AND l.status IN ('active', 'due', 'overdue', 'suspended')
    FOR UPDATE OF l
  LOOP
    INSERT INTO public.dompet_member (member_id, saldo)
    VALUES (loan_rec.member_id, 0)
    ON CONFLICT (member_id) DO NOTHING;

    SELECT saldo INTO saldo_before
    FROM public.dompet_member
    WHERE member_id = loan_rec.member_id
    FOR UPDATE;

    saldo_after := saldo_before + loan_rec.available_amount;
    UPDATE public.dompet_member
    SET saldo = saldo_after, diperbarui_pada = now()
    WHERE member_id = loan_rec.member_id;

    INSERT INTO public.mutasi_dompet
      (member_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, dibuat_pada)
    VALUES
      (loan_rec.member_id, 'KREDIT-MIGRASI-' || loan_rec.application_id, 'CREDIT',
       loan_rec.available_amount, 'AGENT_CREDIT_MIGRATION',
       'Saldo kredit lama dipindahkan ke saldo utama', saldo_before, saldo_after, now());

    INSERT INTO public.agent_credit_mutation
      (loan_id, application_id, member_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah)
    VALUES
      (loan_rec.id, loan_rec.application_id, loan_rec.member_id,
       'KREDIT-MIGRASI-' || loan_rec.application_id, 'DEBIT', loan_rec.available_amount,
       'CREDIT_MOVED_TO_MAIN_BALANCE', 'Saldo kredit lama dipindahkan ke saldo utama',
       loan_rec.available_amount, 0);

    UPDATE public.agent_credit_loan
    SET available_amount = 0, updated_at = now()
    WHERE id = loan_rec.id;

    INSERT INTO public.agent_credit_main_balance_migration (loan_id, member_id, amount)
    VALUES (loan_rec.id, loan_rec.member_id, loan_rec.available_amount);
  END LOOP;
END $$;

-- Jalur transaksi lama tetap ada untuk kompatibilitas skema, tetapi tidak boleh
-- lagi mendebit pinjaman saat saldo utama tidak cukup.
CREATE OR REPLACE FUNCTION public.fn_agent_credit_debit_available(
  p_member_id BIGINT,
  p_ref_id TEXT,
  p_amount BIGINT,
  p_reason TEXT,
  p_note TEXT
) RETURNS BIGINT
LANGUAGE plpgsql
AS $function$
BEGIN
  RETURN 0;
END;
$function$;

-- Refund transaksi lama juga tidak boleh membuat saldo kredit baru.
CREATE OR REPLACE FUNCTION public.fn_agent_credit_refund_by_ref(
  p_member_id BIGINT,
  p_ref_id TEXT,
  p_amount BIGINT,
  p_reason TEXT,
  p_note TEXT
) RETURNS BIGINT
LANGUAGE plpgsql
AS $function$
BEGIN
  RETURN 0;
END;
$function$;
