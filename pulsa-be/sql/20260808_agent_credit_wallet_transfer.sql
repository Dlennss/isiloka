-- Saldo kredit adalah sumber pencairan pinjaman, bukan sumber pembayaran produk.
-- Transaksi produk harus menggunakan saldo utama setelah agent melakukan mutasi.

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

-- Pinjaman yang disetujui menjadi tagihan penuh. Pemindahan dana hanya mengubah
-- available_amount dan tidak menambah atau mengurangi nilai tagihan.
UPDATE public.agent_credit_loan
SET outstanding_amount = principal_amount,
    updated_at = now()
WHERE status IN ('active', 'due', 'overdue', 'suspended')
  AND outstanding_amount <> principal_amount;

UPDATE public.agent_credit_loan
SET available_amount = 0,
    updated_at = now()
WHERE status = 'paid'
  AND available_amount <> 0;

CREATE OR REPLACE FUNCTION public.fn_agent_credit_refund_by_ref(
  p_member_id BIGINT,
  p_ref_id TEXT,
  p_amount BIGINT,
  p_reason TEXT,
  p_note TEXT
) RETURNS BIGINT
LANGUAGE plpgsql
AS $function$
DECLARE
  refundable BIGINT := COALESCE(p_amount, 0);
  used_amount BIGINT;
  already_refunded BIGINT;
  remaining BIGINT;
  refund_amount BIGINT;
  debit_rec RECORD;
  before_amount BIGINT;
BEGIN
  IF refundable <= 0 THEN
    RETURN 0;
  END IF;

  SELECT COALESCE(SUM(jumlah), 0)
  INTO used_amount
  FROM public.agent_credit_mutation
  WHERE member_id = p_member_id
    AND ref_id = p_ref_id
    AND arah = 'DEBIT';

  SELECT COALESCE(SUM(jumlah), 0)
  INTO already_refunded
  FROM public.agent_credit_mutation
  WHERE member_id = p_member_id
    AND ref_id = p_ref_id
    AND arah = 'CREDIT'
    AND alasan = p_reason;

  remaining := LEAST(refundable, GREATEST(used_amount - already_refunded, 0));
  IF remaining <= 0 THEN
    RETURN 0;
  END IF;

  FOR debit_rec IN
    SELECT m.loan_id, m.application_id, m.member_id, SUM(m.jumlah) AS debit_amount
    FROM public.agent_credit_mutation m
    WHERE m.member_id = p_member_id
      AND m.ref_id = p_ref_id
      AND m.arah = 'DEBIT'
    GROUP BY m.loan_id, m.application_id, m.member_id
    ORDER BY MIN(m.id) DESC
  LOOP
    EXIT WHEN remaining <= 0;
    refund_amount := LEAST(remaining, debit_rec.debit_amount);

    SELECT available_amount
    INTO before_amount
    FROM public.agent_credit_loan
    WHERE id = debit_rec.loan_id
    FOR UPDATE;

    UPDATE public.agent_credit_loan
    SET available_amount = LEAST(principal_amount, available_amount + refund_amount),
        updated_at = now()
    WHERE id = debit_rec.loan_id;

    INSERT INTO public.agent_credit_mutation
      (loan_id, application_id, member_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah)
    VALUES
      (debit_rec.loan_id, debit_rec.application_id, debit_rec.member_id, p_ref_id, 'CREDIT', refund_amount, p_reason, COALESCE(p_note, ''), before_amount, LEAST(before_amount + refund_amount, (SELECT principal_amount FROM public.agent_credit_loan WHERE id = debit_rec.loan_id)))
    ON CONFLICT (loan_id, member_id, ref_id, arah, alasan) DO NOTHING;

    remaining := remaining - refund_amount;
  END LOOP;

  RETURN LEAST(refundable, GREATEST(used_amount - already_refunded, 0)) - remaining;
END;
$function$;
