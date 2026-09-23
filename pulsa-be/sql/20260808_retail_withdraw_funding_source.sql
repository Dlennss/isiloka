ALTER TABLE public.retail_withdraw_request
  ADD COLUMN IF NOT EXISTS source_type TEXT NOT NULL DEFAULT 'main_balance',
  ADD COLUMN IF NOT EXISTS credit_loan_id BIGINT REFERENCES public.agent_credit_loan(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS credit_application_id BIGINT REFERENCES public.agent_credit_application(id) ON DELETE SET NULL;

UPDATE public.retail_withdraw_request
SET source_type = 'main_balance'
WHERE source_type IS NULL OR source_type NOT IN ('main_balance', 'credit');

ALTER TABLE public.retail_withdraw_request
  DROP CONSTRAINT IF EXISTS retail_withdraw_request_source_type_check;

ALTER TABLE public.retail_withdraw_request
  ADD CONSTRAINT retail_withdraw_request_source_type_check
  CHECK (source_type IN ('main_balance', 'credit'));

CREATE INDEX IF NOT EXISTS retail_withdraw_request_source_status_idx
  ON public.retail_withdraw_request(source_type, status, created_at DESC);
