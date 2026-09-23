-- Separate agent credit balance from the member's real wallet balance.
-- principal_amount = credit limit, outstanding_amount = used credit debt,
-- available_amount = remaining credit that can be spent.

ALTER TABLE public.agent_credit_loan
  ADD COLUMN IF NOT EXISTS available_amount BIGINT NOT NULL DEFAULT 0 CHECK (available_amount >= 0);

UPDATE public.agent_credit_loan
SET available_amount = GREATEST(principal_amount - outstanding_amount, 0)
WHERE available_amount = 0
  AND status IN ('active', 'overdue')
  AND principal_amount > 0;

CREATE TABLE IF NOT EXISTS public.agent_credit_mutation (
  id BIGSERIAL PRIMARY KEY,
  loan_id BIGINT NOT NULL REFERENCES public.agent_credit_loan(id) ON DELETE CASCADE,
  application_id BIGINT REFERENCES public.agent_credit_application(id) ON DELETE SET NULL,
  member_id BIGINT NOT NULL REFERENCES public.member(id) ON DELETE CASCADE,
  ref_id TEXT NOT NULL,
  arah TEXT NOT NULL CHECK (arah IN ('CREDIT', 'DEBIT')),
  jumlah BIGINT NOT NULL CHECK (jumlah > 0),
  alasan TEXT NOT NULL DEFAULT '',
  catatan TEXT NOT NULL DEFAULT '',
  saldo_sebelum BIGINT NOT NULL DEFAULT 0,
  saldo_sesudah BIGINT NOT NULL DEFAULT 0,
  dibuat_pada TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS agent_credit_mutation_unique_reason
  ON public.agent_credit_mutation(loan_id, member_id, ref_id, arah, alasan);

CREATE INDEX IF NOT EXISTS idx_agent_credit_mutation_member
  ON public.agent_credit_mutation(member_id, dibuat_pada DESC);

CREATE INDEX IF NOT EXISTS idx_agent_credit_mutation_loan
  ON public.agent_credit_mutation(loan_id, dibuat_pada DESC);

UPDATE public.agent_credit_loan l
SET outstanding_amount = 0,
    available_amount = principal_amount,
    status = 'active',
    updated_at = now()
WHERE l.status IN ('active', 'overdue')
  AND l.principal_amount > 0
  AND l.outstanding_amount = l.principal_amount
  AND l.available_amount = l.principal_amount
  AND NOT EXISTS (
    SELECT 1 FROM public.agent_credit_mutation m WHERE m.loan_id = l.id
  )
  AND NOT EXISTS (
    SELECT 1 FROM public.agent_credit_payment p WHERE p.loan_id = l.id
  );

CREATE OR REPLACE FUNCTION public.fn_agent_credit_debit_available(
  p_member_id BIGINT,
  p_ref_id TEXT,
  p_amount BIGINT,
  p_reason TEXT,
  p_note TEXT
) RETURNS BIGINT
LANGUAGE plpgsql
AS $function$
DECLARE
  remaining BIGINT := COALESCE(p_amount, 0);
  use_amount BIGINT;
  loan_rec RECORD;
BEGIN
  IF remaining <= 0 THEN
    RETURN 0;
  END IF;

  FOR loan_rec IN
    SELECT id, application_id, member_id, available_amount, outstanding_amount
    FROM public.agent_credit_loan
    WHERE member_id = p_member_id
      AND status IN ('active', 'overdue')
      AND available_amount > 0
    ORDER BY due_date ASC, id ASC
    FOR UPDATE
  LOOP
    EXIT WHEN remaining <= 0;
    use_amount := LEAST(remaining, loan_rec.available_amount);

    UPDATE public.agent_credit_loan
    SET available_amount = available_amount - use_amount,
        outstanding_amount = outstanding_amount + use_amount,
        status = 'active',
        paid_at = NULL,
        updated_at = now()
    WHERE id = loan_rec.id;

    INSERT INTO public.agent_credit_mutation
      (loan_id, application_id, member_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah)
    VALUES
      (loan_rec.id, loan_rec.application_id, loan_rec.member_id, p_ref_id, 'DEBIT', use_amount, p_reason, COALESCE(p_note, ''), loan_rec.available_amount, loan_rec.available_amount - use_amount)
    ON CONFLICT (loan_id, member_id, ref_id, arah, alasan) DO NOTHING;

    remaining := remaining - use_amount;
  END LOOP;

  RETURN COALESCE(p_amount, 0) - remaining;
END;
$function$;

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
        outstanding_amount = GREATEST(outstanding_amount - refund_amount, 0),
        updated_at = now()
    WHERE id = debit_rec.loan_id;

    INSERT INTO public.agent_credit_mutation
      (loan_id, application_id, member_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah)
    VALUES
      (debit_rec.loan_id, debit_rec.application_id, debit_rec.member_id, p_ref_id, 'CREDIT', refund_amount, p_reason, COALESCE(p_note, ''), before_amount, before_amount + refund_amount)
    ON CONFLICT (loan_id, member_id, ref_id, arah, alasan) DO NOTHING;

    remaining := remaining - refund_amount;
  END LOOP;

  RETURN LEAST(refundable, GREATEST(used_amount - already_refunded, 0)) - remaining;
END;
$function$;

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
  wallet_debit BIGINT;
  credit_debit BIGINT;
  refunded_wallet BIGINT;
  refunded_credit BIGINT;
  redebit_wallet BIGINT;
  redebit_credit BIGINT;
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
      wallet_debit := LEAST(saldo_skrg, biaya);
      credit_debit := 0;

      IF wallet_debit < biaya THEN
        credit_debit := public.fn_agent_credit_debit_available(
          NEW.member_id,
          NEW.ref_id,
          biaya - wallet_debit,
          'TRX_HOLD',
          'hold dari saldo pinjaman agent'
        );
      END IF;

      IF wallet_debit + credit_debit < biaya THEN
        RAISE EXCEPTION 'saldo tidak cukup: saldo=% kredit=% biaya=%', saldo_skrg, credit_debit, biaya;
      END IF;

      IF wallet_debit > 0 THEN
        INSERT INTO public.mutasi_dompet
          (member_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, dibuat_pada)
        VALUES
          (NEW.member_id, NEW.ref_id, 'DEBIT', wallet_debit, 'TRX_HOLD', 'hold saldo asli', saldo_skrg, saldo_skrg - wallet_debit, NOW());

        UPDATE public.dompet_member
           SET saldo = saldo - wallet_debit
         WHERE member_id = NEW.member_id;
      END IF;
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
      SELECT EXISTS(
        SELECT 1
          FROM public.agent_credit_mutation
         WHERE member_id = NEW.member_id
           AND ref_id = NEW.ref_id
           AND arah = 'DEBIT'
           AND alasan = 'TRX_HOLD'
      ) INTO has_debit;
    END IF;

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

    SELECT COALESCE(SUM(jumlah), 0)
    INTO wallet_debit
    FROM public.mutasi_dompet
    WHERE member_id = NEW.member_id
      AND ref_id = NEW.ref_id
      AND UPPER(COALESCE(arah, '')) = 'DEBIT'
      AND UPPER(COALESCE(alasan, '')) = 'TRX_HOLD';

    IF wallet_debit > 0 THEN
      INSERT INTO public.mutasi_dompet
        (member_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, dibuat_pada)
      VALUES
        (NEW.member_id, NEW.ref_id, 'CREDIT', wallet_debit, 'REFUND', 'refund saldo asli', saldo_skrg, saldo_skrg + wallet_debit, NOW());

      UPDATE public.dompet_member
         SET saldo = saldo + wallet_debit
       WHERE member_id = NEW.member_id;
    END IF;

    refunded_credit := public.fn_agent_credit_refund_by_ref(
      NEW.member_id,
      NEW.ref_id,
      biaya,
      'REFUND',
      'refund saldo pinjaman agent'
    );

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

    IF NOT has_refund THEN
      SELECT EXISTS(
        SELECT 1
          FROM public.agent_credit_mutation
         WHERE member_id = NEW.member_id
           AND ref_id = NEW.ref_id
           AND arah = 'CREDIT'
           AND alasan = 'REFUND'
      ) INTO has_refund;
    END IF;

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

        SELECT COALESCE(SUM(jumlah), 0)
        INTO refunded_wallet
        FROM public.mutasi_dompet
        WHERE member_id = NEW.member_id
          AND ref_id = NEW.ref_id
          AND UPPER(COALESCE(arah, '')) IN ('CREDIT', 'KREDIT')
          AND UPPER(COALESCE(alasan, '')) = 'REFUND';

        redebit_wallet := LEAST(saldo_skrg, LEAST(refunded_wallet, biaya));
        redebit_credit := 0;

        IF redebit_wallet < biaya THEN
          redebit_credit := public.fn_agent_credit_debit_available(
            NEW.member_id,
            NEW.ref_id,
            biaya - redebit_wallet,
            'TRX_SETTLE',
            'potong ulang saldo pinjaman setelah refund - transaksi sukses'
          );
        END IF;

        IF redebit_wallet + redebit_credit >= biaya THEN
          IF redebit_wallet > 0 THEN
            INSERT INTO public.mutasi_dompet
              (member_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, dibuat_pada)
            VALUES
              (NEW.member_id, NEW.ref_id, 'DEBIT', redebit_wallet, 'TRX_SETTLE', 'potong ulang saldo asli setelah refund - transaksi sukses', saldo_skrg, saldo_skrg - redebit_wallet, NOW());

            UPDATE public.dompet_member
               SET saldo = saldo - redebit_wallet
             WHERE member_id = NEW.member_id;
          END IF;
        ELSE
          RAISE WARNING 'saldo tidak cukup untuk re-debit: ref_id=% saldo=% kredit=% biaya=%', NEW.ref_id, saldo_skrg, redebit_credit, biaya;
        END IF;
      END IF;
    END IF;
  END IF;

  RETURN NEW;
END;
$function$;
