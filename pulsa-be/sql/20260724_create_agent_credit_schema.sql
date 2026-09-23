-- Agent credit / credit limit workflow for PulsaKilat retail agents.
-- Covers: application, marketing approval, loan balance, repayment history,
-- and rank-based limit upgrades for agents with good payment behavior.

CREATE TABLE IF NOT EXISTS public.agent_credit_rank (
  id BIGSERIAL PRIMARY KEY,
  code TEXT NOT NULL UNIQUE,
  name TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  limit_amount BIGINT NOT NULL CHECK (limit_amount >= 0),
  min_on_time_payments INT NOT NULL DEFAULT 0 CHECK (min_on_time_payments >= 0),
  max_late_payments INT NOT NULL DEFAULT 0 CHECK (max_late_payments >= 0),
  sort_order INT NOT NULL DEFAULT 0,
  active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO public.agent_credit_rank
  (code, name, description, limit_amount, min_on_time_payments, max_late_payments, sort_order)
VALUES
  ('start', 'Kilat Start', 'Limit awal untuk agent baru.', 500000, 0, 0, 10),
  ('plus', 'Kilat Plus', 'Naik setelah 3 pinjaman lunas tepat waktu.', 1000000, 3, 0, 20),
  ('elite', 'Kilat Elite', 'Limit maksimal setelah 5 pinjaman lunas tepat waktu.', 2000000, 5, 0, 30)
ON CONFLICT (code) DO UPDATE SET
  name = EXCLUDED.name,
  description = EXCLUDED.description,
  limit_amount = EXCLUDED.limit_amount,
  min_on_time_payments = EXCLUDED.min_on_time_payments,
  max_late_payments = EXCLUDED.max_late_payments,
  sort_order = EXCLUDED.sort_order,
  updated_at = now();

CREATE TABLE IF NOT EXISTS public.agent_credit_application (
  id BIGSERIAL PRIMARY KEY,
  member_id BIGINT NOT NULL REFERENCES public.member(id) ON DELETE CASCADE,
  requested_amount BIGINT NOT NULL CHECK (requested_amount > 0),
  approved_amount BIGINT NOT NULL DEFAULT 0 CHECK (approved_amount >= 0),
  rank_id BIGINT REFERENCES public.agent_credit_rank(id) ON DELETE SET NULL,
  status TEXT NOT NULL DEFAULT 'submitted' CHECK (
    status IN ('draft', 'submitted', 'analysis_review', 'master_review', 'marketing_review', 'ready_to_disburse', 'approved', 'rejected', 'analysis_rejected', 'master_rejected', 'cancelled')
  ),
  applicant_data JSONB NOT NULL DEFAULT '{}'::jsonb,
  document_data JSONB NOT NULL DEFAULT '{}'::jsonb,
  agent_signature_data TEXT,
  agent_signature_at TIMESTAMPTZ,
  marketing_user_id BIGINT REFERENCES public.member(id) ON DELETE SET NULL,
  marketing_reviewed_at TIMESTAMPTZ,
  marketing_note TEXT NOT NULL DEFAULT '',
  analyst_user_id BIGINT REFERENCES public.member(id) ON DELETE SET NULL,
  analyst_reviewed_at TIMESTAMPTZ,
  analyst_note TEXT NOT NULL DEFAULT '',
  analyst_recommendation TEXT NOT NULL DEFAULT '',
  analyst_recommended_amount BIGINT NOT NULL DEFAULT 0 CHECK (analyst_recommended_amount >= 0),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_agent_credit_application_member
  ON public.agent_credit_application(member_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_agent_credit_application_status
  ON public.agent_credit_application(status, created_at DESC);

CREATE TABLE IF NOT EXISTS public.agent_credit_loan (
  id BIGSERIAL PRIMARY KEY,
  application_id BIGINT UNIQUE REFERENCES public.agent_credit_application(id) ON DELETE SET NULL,
  member_id BIGINT NOT NULL REFERENCES public.member(id) ON DELETE CASCADE,
  rank_id BIGINT REFERENCES public.agent_credit_rank(id) ON DELETE SET NULL,
  principal_amount BIGINT NOT NULL CHECK (principal_amount > 0),
  outstanding_amount BIGINT NOT NULL CHECK (outstanding_amount >= 0),
  status TEXT NOT NULL DEFAULT 'active' CHECK (
    status IN ('active', 'paid', 'overdue', 'defaulted', 'cancelled')
  ),
  approved_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  due_date DATE NOT NULL,
  paid_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_agent_credit_loan_member
  ON public.agent_credit_loan(member_id, status, due_date);

CREATE TABLE IF NOT EXISTS public.agent_credit_payment (
  id BIGSERIAL PRIMARY KEY,
  loan_id BIGINT NOT NULL REFERENCES public.agent_credit_loan(id) ON DELETE CASCADE,
  member_id BIGINT NOT NULL REFERENCES public.member(id) ON DELETE CASCADE,
  ref_id TEXT UNIQUE,
  amount BIGINT NOT NULL CHECK (amount > 0),
  due_date DATE NOT NULL,
  paid_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  days_late INT NOT NULL DEFAULT 0 CHECK (days_late >= 0),
  status TEXT NOT NULL DEFAULT 'on_time' CHECK (
    status IN ('on_time', 'late', 'partial')
  ),
  note TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_agent_credit_payment_member
  ON public.agent_credit_payment(member_id, paid_at DESC);

CREATE INDEX IF NOT EXISTS idx_agent_credit_payment_loan
  ON public.agent_credit_payment(loan_id, paid_at DESC);

CREATE TABLE IF NOT EXISTS public.agent_credit_rank_history (
  id BIGSERIAL PRIMARY KEY,
  member_id BIGINT NOT NULL REFERENCES public.member(id) ON DELETE CASCADE,
  old_rank_id BIGINT REFERENCES public.agent_credit_rank(id) ON DELETE SET NULL,
  new_rank_id BIGINT REFERENCES public.agent_credit_rank(id) ON DELETE SET NULL,
  reason TEXT NOT NULL DEFAULT '',
  on_time_payment_count INT NOT NULL DEFAULT 0,
  late_payment_count INT NOT NULL DEFAULT 0,
  created_by BIGINT REFERENCES public.member(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_agent_credit_rank_history_member
  ON public.agent_credit_rank_history(member_id, created_at DESC);

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
LEFT JOIN public.agent_credit_rank starter_rank ON starter_rank.code = 'starter'
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
    COALESCE(SUM(l.outstanding_amount), 0)::BIGINT AS outstanding_amount
  FROM public.agent_credit_loan l
  WHERE l.member_id = m.id
    AND l.status IN ('active', 'overdue')
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
