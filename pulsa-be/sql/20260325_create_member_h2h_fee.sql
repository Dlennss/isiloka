CREATE TABLE IF NOT EXISTS public.member_h2h_fee (
  id bigserial PRIMARY KEY,
  member_id bigint NOT NULL REFERENCES public.member(id) ON DELETE CASCADE,
  fee_code text NOT NULL,
  fee_rp bigint NOT NULL DEFAULT 0,
  aktif boolean NOT NULL DEFAULT true,
  dibuat_pada timestamptz NOT NULL DEFAULT now(),
  diubah_pada timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT member_h2h_fee_code_check CHECK (
    UPPER(TRIM(fee_code)) IN ('DANA', 'GOPAY', 'OVO', 'LINKAJA', 'SHOPEEPAY', 'LAINNYA')
  ),
  CONSTRAINT member_h2h_fee_unique UNIQUE (member_id, fee_code)
);

CREATE INDEX IF NOT EXISTS member_h2h_fee_member_idx
  ON public.member_h2h_fee (member_id);

INSERT INTO public.member_h2h_fee
  (member_id, fee_code, fee_rp, aktif, dibuat_pada, diubah_pada)
SELECT
  mfk.member_id,
  UPPER(TRIM(k.nama)) AS fee_code,
  COALESCE(mfk.fee_rp, 0) AS fee_rp,
  COALESCE(mfk.aktif, true) AS aktif,
  COALESCE(mfk.dibuat_pada, now()) AS dibuat_pada,
  COALESCE(mfk.diubah_pada, now()) AS diubah_pada
FROM public.member_fee_kategori mfk
JOIN public.kategori k ON k.id = mfk.kategori_id
WHERE UPPER(TRIM(k.nama)) IN ('DANA', 'GOPAY', 'OVO', 'LINKAJA', 'SHOPEEPAY', 'LAINNYA')
ON CONFLICT (member_id, fee_code)
DO UPDATE SET
  fee_rp = EXCLUDED.fee_rp,
  aktif = EXCLUDED.aktif,
  diubah_pada = now();

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'syarif') THEN
    GRANT SELECT, INSERT, UPDATE, DELETE
    ON TABLE public.member_h2h_fee
    TO syarif;

    GRANT USAGE, SELECT
    ON SEQUENCE public.member_h2h_fee_id_seq
    TO syarif;
  END IF;

  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'app_user') THEN
    GRANT SELECT, INSERT, UPDATE, DELETE
    ON TABLE public.member_h2h_fee
    TO app_user;

    GRANT USAGE, SELECT
    ON SEQUENCE public.member_h2h_fee_id_seq
    TO app_user;
  END IF;
END $$;
