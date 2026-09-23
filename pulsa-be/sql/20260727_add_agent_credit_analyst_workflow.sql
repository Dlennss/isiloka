ALTER TABLE public.agent_credit_application
  DROP CONSTRAINT IF EXISTS agent_credit_application_status_check;

ALTER TABLE public.agent_credit_application
  ADD CONSTRAINT agent_credit_application_status_check CHECK (
    status IN (
      'draft',
      'submitted',
      'analysis_review',
      'master_review',
      'marketing_review',
      'ready_to_disburse',
      'approved',
      'rejected',
      'analysis_rejected',
      'master_rejected',
      'cancelled'
    )
  );

ALTER TABLE public.agent_credit_application
  ADD COLUMN IF NOT EXISTS analyst_user_id BIGINT REFERENCES public.member(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS analyst_reviewed_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS analyst_note TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS analyst_recommendation TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS analyst_recommended_amount BIGINT NOT NULL DEFAULT 0 CHECK (analyst_recommended_amount >= 0);

CREATE INDEX IF NOT EXISTS idx_agent_credit_application_analyst_status
  ON public.agent_credit_application(status, analyst_reviewed_at DESC);
