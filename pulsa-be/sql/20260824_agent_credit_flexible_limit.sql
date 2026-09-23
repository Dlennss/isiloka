ALTER TABLE public.agent_credit_rank_history
  ADD COLUMN IF NOT EXISTS custom_limit_amount BIGINT,
  ADD COLUMN IF NOT EXISTS custom_limit_name TEXT NOT NULL DEFAULT '';

ALTER TABLE public.agent_credit_rank_history
  DROP CONSTRAINT IF EXISTS agent_credit_rank_history_custom_limit_check;

ALTER TABLE public.agent_credit_rank_history
  ADD CONSTRAINT agent_credit_rank_history_custom_limit_check
  CHECK (custom_limit_amount IS NULL OR custom_limit_amount > 0);
