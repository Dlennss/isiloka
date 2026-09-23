CREATE TABLE IF NOT EXISTS public.loketbayar_provider_transfers (
  id bigserial PRIMARY KEY,
  ref_id text NOT NULL UNIQUE,
  source_provider text NOT NULL DEFAULT 'loketbayar',
  provider text NOT NULL,
  provider_rekening_id bigint NOT NULL REFERENCES public.provider_rekening(id),
  bank_code text NOT NULL,
  bank_name text NOT NULL,
  account_no text NOT NULL,
  account_name text NOT NULL DEFAULT '',
  amount bigint NOT NULL CHECK (amount > 0),
  admin_fee bigint NOT NULL DEFAULT 0 CHECK (admin_fee >= 0),
  note text NOT NULL DEFAULT '',
  status text NOT NULL DEFAULT 'ready',
  request_raw jsonb NOT NULL DEFAULT '{}'::jsonb,
  response_raw jsonb NOT NULL DEFAULT '{}'::jsonb,
  provider_transaction_id text NOT NULL DEFAULT '',
  response_error text NOT NULL DEFAULT '',
  response_reason text NOT NULL DEFAULT '',
  source_saldo_after bigint,
  source_snapshot_after bigint,
  provider_saldo_after bigint,
  created_by bigint REFERENCES public.member(id),
  processed_at timestamptz,
  completed_at timestamptz,
  reversed_at timestamptz,
  callback_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS loketbayar_provider_transfers_provider_idx ON public.loketbayar_provider_transfers (provider, id DESC);
CREATE INDEX IF NOT EXISTS loketbayar_provider_transfers_status_idx ON public.loketbayar_provider_transfers (status, id DESC);
CREATE INDEX IF NOT EXISTS loketbayar_provider_transfers_created_idx ON public.loketbayar_provider_transfers (created_at DESC);
