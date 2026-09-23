ALTER TABLE public.bank
  ADD COLUMN IF NOT EXISTS admin_staff_only boolean NOT NULL DEFAULT false;

UPDATE public.bank
SET admin_staff_only = false
WHERE admin_staff_only IS NULL;
