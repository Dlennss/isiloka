CREATE TABLE IF NOT EXISTS public.internal_finance_entry (
  id BIGSERIAL PRIMARY KEY,
  ref_id TEXT NOT NULL UNIQUE,
  entry_type TEXT NOT NULL,
  category TEXT NOT NULL,
  direction TEXT NOT NULL,
  bank_id BIGINT NOT NULL REFERENCES public.bank(id),
  amount BIGINT NOT NULL,
  fee BIGINT NOT NULL DEFAULT 0,
  total_amount BIGINT NOT NULL,
  counterparty TEXT NOT NULL DEFAULT '',
  note TEXT NOT NULL DEFAULT '',
  occurred_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_by BIGINT REFERENCES public.member(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  meta JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX IF NOT EXISTS internal_finance_entry_bank_idx
  ON public.internal_finance_entry (bank_id, occurred_at DESC);

CREATE INDEX IF NOT EXISTS internal_finance_entry_type_idx
  ON public.internal_finance_entry (entry_type, category, occurred_at DESC);

CREATE TABLE IF NOT EXISTS public.accounting_opening_balance (
  id BIGSERIAL PRIMARY KEY,
  period_month DATE NOT NULL,
  account_code TEXT NOT NULL,
  account_name TEXT NOT NULL,
  account_group TEXT NOT NULL,
  amount BIGINT NOT NULL DEFAULT 0,
  note TEXT NOT NULL DEFAULT '',
  created_by BIGINT REFERENCES public.member(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS accounting_opening_balance_period_code_uidx
  ON public.accounting_opening_balance (period_month, account_code);
