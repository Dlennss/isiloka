CREATE TABLE IF NOT EXISTS public.app_ads (
  id BIGSERIAL PRIMARY KEY,
  judul TEXT NOT NULL DEFAULT '',
  keterangan TEXT NOT NULL DEFAULT '',
  image_url TEXT NOT NULL,
  link_url TEXT NULL,
  urutan INTEGER NOT NULL DEFAULT 0,
  aktif BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_app_ads_active_urutan
  ON public.app_ads (aktif, urutan, id DESC);
