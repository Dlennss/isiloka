CREATE TABLE IF NOT EXISTS public.provider_merchant_id (
  id BIGSERIAL PRIMARY KEY,
  provider TEXT NOT NULL,
  merchant_id TEXT NOT NULL,
  label TEXT NOT NULL DEFAULT '',
  catatan TEXT NOT NULL DEFAULT '',
  aktif BOOLEAN NOT NULL DEFAULT TRUE,
  dibuat_pada TIMESTAMPTZ NOT NULL DEFAULT now(),
  diubah_pada TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT provider_merchant_id_provider_not_blank CHECK (trim(provider) <> ''),
  CONSTRAINT provider_merchant_id_merchant_not_blank CHECK (trim(merchant_id) <> '')
);

CREATE UNIQUE INDEX IF NOT EXISTS provider_merchant_id_provider_mid_uidx
  ON public.provider_merchant_id ((lower(trim(provider))), (lower(trim(merchant_id))));

CREATE INDEX IF NOT EXISTS provider_merchant_id_provider_idx
  ON public.provider_merchant_id ((lower(trim(provider))));

CREATE INDEX IF NOT EXISTS provider_merchant_id_active_provider_idx
  ON public.provider_merchant_id (aktif, (lower(trim(provider))));

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'syarif') THEN
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE public.provider_merchant_id TO syarif;
    GRANT USAGE, SELECT ON SEQUENCE public.provider_merchant_id_id_seq TO syarif;
  END IF;

  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'app_user') THEN
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE public.provider_merchant_id TO app_user;
    GRANT USAGE, SELECT ON SEQUENCE public.provider_merchant_id_id_seq TO app_user;
  END IF;
END $$;
