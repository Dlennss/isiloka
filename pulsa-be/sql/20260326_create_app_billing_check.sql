CREATE TABLE IF NOT EXISTS public.app_billing_check (
  id bigserial PRIMARY KEY,
  ref_id text NOT NULL UNIQUE,
  member_id bigint NULL REFERENCES public.member(id) ON DELETE SET NULL,
  guest_nama text NULL,
  guest_email text NULL,
  guest_phone text NULL,
  produk_id bigint NOT NULL REFERENCES public.produk(id),
  produk_sku_snapshot text NOT NULL,
  produk_nama_snapshot text NOT NULL,
  dest text NOT NULL,
  buyer_type text NOT NULL,
  buyer_role text NOT NULL DEFAULT '',
  provider text NOT NULL DEFAULT 'yuscom',
  harga_provider bigint NOT NULL DEFAULT 0,
  status text NOT NULL,
  kode_respon text NULL,
  pesan text NULL,
  sn text NULL,
  raw_request jsonb NULL,
  raw_callback jsonb NULL,
  dibuat_pada timestamptz NOT NULL DEFAULT now(),
  diubah_pada timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_app_billing_check_member_id ON public.app_billing_check(member_id);
CREATE INDEX IF NOT EXISTS idx_app_billing_check_status ON public.app_billing_check(status);
CREATE INDEX IF NOT EXISTS idx_app_billing_check_dibuat_pada ON public.app_billing_check(dibuat_pada DESC);

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'app_user') THEN
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE public.app_billing_check TO app_user;
    GRANT USAGE, SELECT ON SEQUENCE public.app_billing_check_id_seq TO app_user;
  END IF;
END $$;
