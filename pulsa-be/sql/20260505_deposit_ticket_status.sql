ALTER TABLE public.deposit_request
  ADD COLUMN IF NOT EXISTS requested_amount bigint,
  ADD COLUMN IF NOT EXISTS unique_code bigint,
  ADD COLUMN IF NOT EXISTS approved_amount bigint;

UPDATE public.deposit_request
SET approved_amount = amount
WHERE status = 'approved'
  AND approved_amount IS NULL;

ALTER TABLE public.deposit_request
  DROP CONSTRAINT IF EXISTS deposit_request_status_check;

ALTER TABLE public.deposit_request
  ADD CONSTRAINT deposit_request_status_check
  CHECK (status = ANY (ARRAY['ticket'::text, 'pending'::text, 'approved'::text, 'rejected'::text, 'cancelled'::text]));
