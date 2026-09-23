ALTER TABLE public.internal_finance_entry
  ADD COLUMN IF NOT EXISTS entry_type TEXT,
  ADD COLUMN IF NOT EXISTS direction TEXT,
  ADD COLUMN IF NOT EXISTS fee BIGINT NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS total_amount BIGINT,
  ADD COLUMN IF NOT EXISTS counterparty TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS occurred_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  ADD COLUMN IF NOT EXISTS meta JSONB NOT NULL DEFAULT '{}'::jsonb;

UPDATE public.internal_finance_entry
SET direction = CASE
  WHEN lower(trim(arah)) IN ('credit', 'kredit', 'in', 'masuk') THEN 'credit'
  ELSE 'debit'
END
WHERE direction IS NULL OR trim(direction) = '';

UPDATE public.internal_finance_entry
SET entry_type = CASE
  WHEN lower(trim(category)) IN (
    'purchase', 'salary', 'other_expense', 'other_income', 'bank_admin',
    'bank_interest', 'bank_interest_tax', 'rent_expense', 'audit_expense',
    'printing_expense', 'event_expense', 'tax_income_expense', 'operational_expense'
  ) THEN lower(trim(category))
  WHEN direction = 'credit' THEN 'other_income'
  ELSE 'operational_expense'
END
WHERE entry_type IS NULL OR trim(entry_type) = '';

UPDATE public.internal_finance_entry
SET total_amount = amount + COALESCE(fee, 0)
WHERE total_amount IS NULL;

UPDATE public.internal_finance_entry
SET occurred_at = created_at
WHERE occurred_at IS NULL;

UPDATE public.internal_finance_entry
SET note = ''
WHERE note IS NULL;

UPDATE public.internal_finance_entry
SET ref_id = 'FIN-MIG-' || id::text
WHERE ref_id IS NULL OR trim(ref_id) = '';

ALTER TABLE public.internal_finance_entry
  ALTER COLUMN ref_id SET NOT NULL,
  ALTER COLUMN entry_type SET NOT NULL,
  ALTER COLUMN direction SET NOT NULL,
  ALTER COLUMN total_amount SET NOT NULL,
  ALTER COLUMN note SET DEFAULT '',
  ALTER COLUMN note SET NOT NULL,
  ALTER COLUMN occurred_at SET NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS internal_finance_entry_ref_id_uidx
  ON public.internal_finance_entry (ref_id);

CREATE INDEX IF NOT EXISTS internal_finance_entry_bank_idx
  ON public.internal_finance_entry (bank_id, occurred_at DESC);

CREATE INDEX IF NOT EXISTS internal_finance_entry_type_idx
  ON public.internal_finance_entry (entry_type, category, occurred_at DESC);
