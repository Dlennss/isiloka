ALTER TABLE public.member
  ADD COLUMN IF NOT EXISTS retail_agent_id bigint NULL,
  ADD COLUMN IF NOT EXISTS retail_master_id bigint NULL,
  ADD COLUMN IF NOT EXISTS h2h_agent_member_id bigint NULL,
  ADD COLUMN IF NOT EXISTS h2h_master_member_id bigint NULL;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'fk_member_retail_agent'
  ) THEN
    ALTER TABLE public.member
      ADD CONSTRAINT fk_member_retail_agent
      FOREIGN KEY (retail_agent_id) REFERENCES public.member(id) ON DELETE SET NULL;
  END IF;

  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'fk_member_retail_master'
  ) THEN
    ALTER TABLE public.member
      ADD CONSTRAINT fk_member_retail_master
      FOREIGN KEY (retail_master_id) REFERENCES public.member(id) ON DELETE SET NULL;
  END IF;

  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'fk_member_h2h_agent'
  ) THEN
    ALTER TABLE public.member
      ADD CONSTRAINT fk_member_h2h_agent
      FOREIGN KEY (h2h_agent_member_id) REFERENCES public.member(id) ON DELETE SET NULL;
  END IF;

  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'fk_member_h2h_master'
  ) THEN
    ALTER TABLE public.member
      ADD CONSTRAINT fk_member_h2h_master
      FOREIGN KEY (h2h_master_member_id) REFERENCES public.member(id) ON DELETE SET NULL;
  END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_member_retail_agent_id ON public.member(retail_agent_id);
CREATE INDEX IF NOT EXISTS idx_member_retail_master_id ON public.member(retail_master_id);
CREATE INDEX IF NOT EXISTS idx_member_h2h_agent_member_id ON public.member(h2h_agent_member_id);
CREATE INDEX IF NOT EXISTS idx_member_h2h_master_member_id ON public.member(h2h_master_member_id);
