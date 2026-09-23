ALTER TABLE public.member
  ADD COLUMN IF NOT EXISTS google_sub TEXT,
  ADD COLUMN IF NOT EXISTS apple_sub TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS member_google_sub_uidx
  ON public.member(google_sub)
  WHERE google_sub IS NOT NULL AND TRIM(google_sub) <> '';

CREATE UNIQUE INDEX IF NOT EXISTS member_apple_sub_uidx
  ON public.member(apple_sub)
  WHERE apple_sub IS NOT NULL AND TRIM(apple_sub) <> '';
