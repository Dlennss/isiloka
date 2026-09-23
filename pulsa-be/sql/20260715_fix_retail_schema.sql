ALTER TABLE public.retail_commission_ledger
  ADD COLUMN IF NOT EXISTS level_name TEXT NOT NULL DEFAULT '';

UPDATE public.retail_commission_ledger
SET level_name = COALESCE(NULLIF(TRIM(level_name), ''), level, '')
WHERE COALESCE(NULLIF(TRIM(level_name), ''), '') = '';

CREATE UNIQUE INDEX IF NOT EXISTS retail_commission_ledger_member_order_uidx
  ON public.retail_commission_ledger (member_id, source_app_order_id);

CREATE INDEX IF NOT EXISTS retail_commission_ledger_member_created_idx
  ON public.retail_commission_ledger (member_id, created_at DESC);

ALTER TABLE public.retail_withdraw_request
  ADD COLUMN IF NOT EXISTS account_name TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS account_number TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS reject_reason TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS ref_id TEXT NOT NULL DEFAULT '';

UPDATE public.retail_withdraw_request
SET
  account_name = COALESCE(NULLIF(TRIM(account_name), ''), bank_account_name, ''),
  account_number = COALESCE(NULLIF(TRIM(account_number), ''), bank_account_no, '')
WHERE COALESCE(NULLIF(TRIM(account_name), ''), '') = ''
   OR COALESCE(NULLIF(TRIM(account_number), ''), '') = '';

CREATE UNIQUE INDEX IF NOT EXISTS retail_withdraw_request_ref_id_uidx
  ON public.retail_withdraw_request (ref_id)
  WHERE TRIM(ref_id) <> '';

CREATE INDEX IF NOT EXISTS retail_withdraw_request_member_created_idx
  ON public.retail_withdraw_request (member_id, created_at DESC);

CREATE INDEX IF NOT EXISTS retail_withdraw_request_status_created_idx
  ON public.retail_withdraw_request (status, created_at DESC);
