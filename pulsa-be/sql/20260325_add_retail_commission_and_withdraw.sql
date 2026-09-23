ALTER TABLE public.app_order
  ADD COLUMN IF NOT EXISTS buyer_role TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS fee_user_snapshot BIGINT NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS fee_agent_snapshot BIGINT NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS fee_master_snapshot BIGINT NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS retail_agent_id_snapshot BIGINT NULL,
  ADD COLUMN IF NOT EXISTS retail_master_id_snapshot BIGINT NULL;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint
    WHERE conname = 'app_order_retail_agent_snapshot_fkey'
  ) THEN
    ALTER TABLE public.app_order
      ADD CONSTRAINT app_order_retail_agent_snapshot_fkey
      FOREIGN KEY (retail_agent_id_snapshot) REFERENCES public.member(id) ON DELETE SET NULL;
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint
    WHERE conname = 'app_order_retail_master_snapshot_fkey'
  ) THEN
    ALTER TABLE public.app_order
      ADD CONSTRAINT app_order_retail_master_snapshot_fkey
      FOREIGN KEY (retail_master_id_snapshot) REFERENCES public.member(id) ON DELETE SET NULL;
  END IF;
END $$;

CREATE INDEX IF NOT EXISTS app_order_retail_agent_snapshot_idx
  ON public.app_order (retail_agent_id_snapshot);

CREATE INDEX IF NOT EXISTS app_order_retail_master_snapshot_idx
  ON public.app_order (retail_master_id_snapshot);

CREATE TABLE IF NOT EXISTS public.retail_commission_ledger (
  id BIGSERIAL PRIMARY KEY,
  member_id BIGINT NOT NULL REFERENCES public.member(id) ON DELETE CASCADE,
  source_member_id BIGINT NULL REFERENCES public.member(id) ON DELETE SET NULL,
  source_app_order_id BIGINT NOT NULL REFERENCES public.app_order(id) ON DELETE CASCADE,
  invoice_id TEXT NOT NULL DEFAULT '',
  level_name TEXT NOT NULL DEFAULT '',
  amount BIGINT NOT NULL DEFAULT 0,
  note TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS retail_commission_ledger_member_order_uidx
  ON public.retail_commission_ledger (member_id, source_app_order_id);

CREATE INDEX IF NOT EXISTS retail_commission_ledger_member_created_idx
  ON public.retail_commission_ledger (member_id, created_at DESC);

CREATE TABLE IF NOT EXISTS public.retail_withdraw_request (
  id BIGSERIAL PRIMARY KEY,
  member_id BIGINT NOT NULL REFERENCES public.member(id) ON DELETE CASCADE,
  amount BIGINT NOT NULL DEFAULT 0,
  bank_name TEXT NOT NULL DEFAULT '',
  account_name TEXT NOT NULL DEFAULT '',
  account_number TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'pending',
  note TEXT NOT NULL DEFAULT '',
  reject_reason TEXT NOT NULL DEFAULT '',
  ref_id TEXT NOT NULL DEFAULT '',
  processed_by BIGINT NULL REFERENCES public.member(id) ON DELETE SET NULL,
  processed_at TIMESTAMPTZ NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS retail_withdraw_request_ref_id_uidx
  ON public.retail_withdraw_request (ref_id);

CREATE INDEX IF NOT EXISTS retail_withdraw_request_member_created_idx
  ON public.retail_withdraw_request (member_id, created_at DESC);

CREATE INDEX IF NOT EXISTS retail_withdraw_request_status_created_idx
  ON public.retail_withdraw_request (status, created_at DESC);
