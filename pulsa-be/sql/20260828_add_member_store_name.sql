ALTER TABLE public.member
  ADD COLUMN IF NOT EXISTS store_name TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_member_store_name_lower
  ON public.member (LOWER(store_name));
