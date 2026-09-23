CREATE TABLE IF NOT EXISTS public.provider_rekening (
  id BIGSERIAL PRIMARY KEY,
  provider TEXT NOT NULL,
  nama TEXT NOT NULL,
  bank TEXT NOT NULL DEFAULT '',
  nomor_rekening TEXT NOT NULL,
  nomor_rekening_digits TEXT NOT NULL,
  catatan TEXT NOT NULL DEFAULT '',
  aktif BOOLEAN NOT NULL DEFAULT TRUE,
  dibuat_pada TIMESTAMPTZ NOT NULL DEFAULT now(),
  diubah_pada TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT provider_rekening_provider_not_blank CHECK (trim(provider) <> ''),
  CONSTRAINT provider_rekening_nama_not_blank CHECK (trim(nama) <> ''),
  CONSTRAINT provider_rekening_nomor_digits_not_blank CHECK (trim(nomor_rekening_digits) <> '')
);

CREATE UNIQUE INDEX IF NOT EXISTS provider_rekening_provider_nomor_uidx
  ON public.provider_rekening ((lower(trim(provider))), nomor_rekening_digits);

CREATE INDEX IF NOT EXISTS provider_rekening_nomor_digits_idx
  ON public.provider_rekening (nomor_rekening_digits);

CREATE INDEX IF NOT EXISTS provider_rekening_provider_idx
  ON public.provider_rekening ((lower(trim(provider))));

CREATE INDEX IF NOT EXISTS provider_rekening_aktif_provider_idx
  ON public.provider_rekening (aktif, (lower(trim(provider))));

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'syarif') THEN
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE public.provider_rekening TO syarif;
    GRANT USAGE, SELECT ON SEQUENCE public.provider_rekening_id_seq TO syarif;
  END IF;

  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'app_user') THEN
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE public.provider_rekening TO app_user;
    GRANT USAGE, SELECT ON SEQUENCE public.provider_rekening_id_seq TO app_user;
  END IF;
END $$;
