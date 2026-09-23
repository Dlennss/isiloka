ALTER TABLE public.member
  ADD COLUMN IF NOT EXISTS phone text;

CREATE INDEX IF NOT EXISTS idx_member_phone_digits
  ON public.member ((regexp_replace(COALESCE(phone, ''), '[^0-9]', '', 'g')));
