ALTER TABLE public.app_ads
  ADD COLUMN IF NOT EXISTS judul TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS keterangan TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS urutan INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

UPDATE public.app_ads
SET
  judul = COALESCE(NULLIF(TRIM(judul), ''), title, ''),
  urutan = CASE
    WHEN COALESCE(urutan, 0) <> 0 THEN urutan
    ELSE COALESCE(sort_order, 0)
  END,
  created_at = COALESCE(created_at, dibuat_pada, now()),
  updated_at = COALESCE(updated_at, diubah_pada, now())
WHERE COALESCE(NULLIF(TRIM(judul), ''), '') = ''
   OR COALESCE(urutan, 0) = 0
   OR created_at IS NULL
   OR updated_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_app_ads_active_urutan
  ON public.app_ads (aktif, urutan, id DESC);
