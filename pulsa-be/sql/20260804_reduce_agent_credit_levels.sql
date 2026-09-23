-- Reduce PulsaKilat agent credit levels to three active stages.
-- Level progression is based on paid loans completed on time.

INSERT INTO public.agent_credit_rank
  (code, name, description, limit_amount, min_on_time_payments, max_late_payments, sort_order, active)
VALUES
  ('start', 'Kilat Start', 'Limit awal untuk agent baru.', 500000, 0, 0, 10, TRUE),
  ('plus', 'Kilat Plus', 'Naik setelah 3 pinjaman lunas tepat waktu.', 1000000, 3, 0, 20, TRUE),
  ('elite', 'Kilat Elite', 'Limit maksimal setelah 5 pinjaman lunas tepat waktu.', 2000000, 5, 0, 30, TRUE)
ON CONFLICT (code) DO UPDATE SET
  name = EXCLUDED.name,
  description = EXCLUDED.description,
  limit_amount = EXCLUDED.limit_amount,
  min_on_time_payments = EXCLUDED.min_on_time_payments,
  max_late_payments = EXCLUDED.max_late_payments,
  sort_order = EXCLUDED.sort_order,
  active = EXCLUDED.active,
  updated_at = now();

UPDATE public.agent_credit_rank
SET active = FALSE, updated_at = now()
WHERE code IN ('pro', 'max', 'starter', 'silver', 'gold', 'platinum');
