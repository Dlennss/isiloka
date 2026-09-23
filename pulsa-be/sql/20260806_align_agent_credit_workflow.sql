-- Align PulsaKilat agent credit workflow with the Admin-Marketing-Agent model.
-- This migration is additive/safe: it keeps legacy statuses but prepares the
-- simpler business states used by the new credit flow.

ALTER TABLE public.agent_credit_loan
  DROP CONSTRAINT IF EXISTS agent_credit_loan_status_check;

ALTER TABLE public.agent_credit_loan
  ADD CONSTRAINT agent_credit_loan_status_check CHECK (
    status IN ('active', 'due', 'overdue', 'suspended', 'paid', 'defaulted', 'cancelled')
  );

UPDATE public.agent_credit_loan
SET available_amount = principal_amount,
    updated_at = now()
WHERE outstanding_amount = 0
  AND status IN ('active', 'due', 'overdue');

UPDATE public.agent_credit_loan
SET status = 'overdue',
    updated_at = now()
WHERE status = 'active'
  AND outstanding_amount > 0
  AND due_date < CURRENT_DATE;

CREATE TABLE IF NOT EXISTS public.audit_log (
  id BIGSERIAL PRIMARY KEY,
  actor_id BIGINT REFERENCES public.member(id) ON DELETE SET NULL,
  actor_role TEXT NOT NULL DEFAULT '',
  action TEXT NOT NULL,
  entity_type TEXT NOT NULL,
  entity_id TEXT NOT NULL DEFAULT '',
  before_data JSONB NOT NULL DEFAULT '{}'::jsonb,
  after_data JSONB NOT NULL DEFAULT '{}'::jsonb,
  reason TEXT NOT NULL DEFAULT '',
  ip_address TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_audit_log_actor
  ON public.audit_log(actor_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_audit_log_entity
  ON public.audit_log(entity_type, entity_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_audit_log_action
  ON public.audit_log(action, created_at DESC);

DROP VIEW IF EXISTS public.agent_credit_member_summary;

CREATE OR REPLACE VIEW public.agent_credit_member_summary AS
SELECT
  m.id AS member_id,
  m.email,
  m.nama,
  m.role,
  COALESCE(current_rank.id, starter_rank.id) AS current_rank_id,
  COALESCE(current_rank.code, starter_rank.code) AS current_rank_code,
  COALESCE(current_rank.name, starter_rank.name) AS current_rank_name,
  COALESCE(current_rank.limit_amount, starter_rank.limit_amount, 0) AS current_limit_amount,
  COALESCE(pay_stats.on_time_payment_count, 0) AS on_time_payment_count,
  COALESCE(pay_stats.late_payment_count, 0) AS late_payment_count,
  COALESCE(active_loan.active_loan_count, 0) AS active_loan_count,
  COALESCE(active_loan.outstanding_amount, 0) AS outstanding_amount,
  COALESCE(active_loan.available_amount, 0) AS credit_available_amount,
  next_rank.id AS next_rank_id,
  next_rank.code AS next_rank_code,
  next_rank.name AS next_rank_name,
  next_rank.limit_amount AS next_limit_amount
FROM public.member m
LEFT JOIN LATERAL (
  SELECT h.new_rank_id
  FROM public.agent_credit_rank_history h
  WHERE h.member_id = m.id
  ORDER BY h.created_at DESC, h.id DESC
  LIMIT 1
) last_rank ON TRUE
LEFT JOIN public.agent_credit_rank current_rank ON current_rank.id = last_rank.new_rank_id
LEFT JOIN public.agent_credit_rank starter_rank ON starter_rank.code = 'start'
LEFT JOIN LATERAL (
  SELECT
    COUNT(*) FILTER (WHERE p.status = 'on_time')::INT AS on_time_payment_count,
    COUNT(*) FILTER (WHERE p.status = 'late')::INT AS late_payment_count
  FROM public.agent_credit_payment p
  WHERE p.member_id = m.id
) pay_stats ON TRUE
LEFT JOIN LATERAL (
  SELECT
    COUNT(*)::INT AS active_loan_count,
    COALESCE(SUM(l.outstanding_amount), 0)::BIGINT AS outstanding_amount,
    COALESCE(SUM(l.available_amount), 0)::BIGINT AS available_amount
  FROM public.agent_credit_loan l
  WHERE l.member_id = m.id
    AND l.status IN ('active', 'due', 'overdue', 'suspended')
) active_loan ON TRUE
LEFT JOIN LATERAL (
  SELECT r.*
  FROM public.agent_credit_rank r
  WHERE r.active = TRUE
    AND r.sort_order > COALESCE(current_rank.sort_order, starter_rank.sort_order, 0)
    AND COALESCE(pay_stats.on_time_payment_count, 0) >= r.min_on_time_payments
    AND COALESCE(pay_stats.late_payment_count, 0) <= r.max_late_payments
  ORDER BY r.sort_order ASC
  LIMIT 1
) next_rank ON TRUE
WHERE lower(COALESCE(m.role, '')) IN ('agent', 'master');
