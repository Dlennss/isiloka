CREATE TABLE IF NOT EXISTS public.h2h_commission_ledger (
  id bigserial PRIMARY KEY,
  member_id bigint NOT NULL REFERENCES public.member(id) ON DELETE CASCADE,
  source_member_id bigint NULL REFERENCES public.member(id) ON DELETE SET NULL,
  source_trx_member_id bigint NOT NULL REFERENCES public.transaksi_member(id) ON DELETE CASCADE,
  ref_id text NOT NULL,
  level_name text NOT NULL,
  kategori_name text NOT NULL,
  amount bigint NOT NULL DEFAULT 0,
  note text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS h2h_commission_ledger_member_trx_uidx
  ON public.h2h_commission_ledger (member_id, source_trx_member_id);

CREATE INDEX IF NOT EXISTS h2h_commission_ledger_member_created_idx
  ON public.h2h_commission_ledger (member_id, created_at DESC);

CREATE TABLE IF NOT EXISTS public.h2h_withdraw_request (
  id bigserial PRIMARY KEY,
  member_id bigint NOT NULL REFERENCES public.member(id) ON DELETE CASCADE,
  amount bigint NOT NULL,
  bank_name text NOT NULL DEFAULT '',
  account_name text NOT NULL DEFAULT '',
  account_number text NOT NULL DEFAULT '',
  status text NOT NULL DEFAULT 'pending',
  note text NOT NULL DEFAULT '',
  reject_reason text NOT NULL DEFAULT '',
  ref_id text NOT NULL,
  processed_by bigint NULL REFERENCES public.member(id) ON DELETE SET NULL,
  processed_at timestamptz NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS h2h_withdraw_request_ref_id_uidx
  ON public.h2h_withdraw_request (ref_id);

CREATE INDEX IF NOT EXISTS h2h_withdraw_request_member_created_idx
  ON public.h2h_withdraw_request (member_id, created_at DESC);

CREATE INDEX IF NOT EXISTS h2h_withdraw_request_status_created_idx
  ON public.h2h_withdraw_request (status, created_at DESC);
