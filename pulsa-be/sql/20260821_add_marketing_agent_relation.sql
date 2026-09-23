ALTER TABLE public.member
  ADD COLUMN IF NOT EXISTS marketing_id BIGINT REFERENCES public.member(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_member_marketing_id
  ON public.member(marketing_id);
