ALTER TABLE public.member
  ADD COLUMN IF NOT EXISTS profile_photo_url TEXT NOT NULL DEFAULT '';
