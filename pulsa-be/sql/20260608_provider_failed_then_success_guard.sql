ALTER TABLE public.provider
  ADD COLUMN IF NOT EXISTS keterangan TEXT NOT NULL DEFAULT '';

CREATE TABLE IF NOT EXISTS public.provider_failed_then_success_guard (
  id BIGSERIAL PRIMARY KEY,
  provider TEXT NOT NULL,
  transaksi_provider_id BIGINT NOT NULL,
  ref_id TEXT NOT NULL DEFAULT '',
  occurred_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  occurred_date_jakarta DATE NOT NULL,
  CONSTRAINT uq_provider_failed_then_success_guard_trx UNIQUE (transaksi_provider_id)
);

CREATE INDEX IF NOT EXISTS idx_provider_failed_then_success_guard_provider_day
  ON public.provider_failed_then_success_guard (provider, occurred_date_jakarta);

CREATE OR REPLACE FUNCTION public.fn_provider_failed_then_success_guard()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
  v_provider TEXT;
  v_day DATE;
  v_count BIGINT;
BEGIN
  IF TG_OP <> 'UPDATE' THEN
    RETURN NEW;
  END IF;

  IF lower(trim(coalesce(OLD.status, ''))) <> 'failed'
     OR lower(trim(coalesce(NEW.status, ''))) <> 'success' THEN
    RETURN NEW;
  END IF;

  v_provider := lower(trim(coalesce(NEW.provider, '')));
  IF v_provider = '' THEN
    RETURN NEW;
  END IF;

  v_day := (now() AT TIME ZONE 'Asia/Jakarta')::date;

  INSERT INTO public.provider_failed_then_success_guard (
    provider,
    transaksi_provider_id,
    ref_id,
    occurred_at,
    occurred_date_jakarta
  ) VALUES (
    v_provider,
    NEW.id,
    coalesce(NEW.ref_id, ''),
    now(),
    v_day
  ) ON CONFLICT (transaksi_provider_id) DO NOTHING;

  SELECT count(*)
  INTO v_count
  FROM public.provider_failed_then_success_guard
  WHERE provider = v_provider
    AND occurred_date_jakarta = v_day;

  IF v_count >= 5 THEN
    UPDATE public.provider
    SET aktif = false,
        keterangan = '5 kali gagal lalu sukses',
        diubah_pada = now()
    WHERE lower(trim(nama)) = v_provider
      AND aktif = true;
  END IF;

  RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_provider_failed_then_success_guard ON public.transaksi_provider;

CREATE TRIGGER trg_provider_failed_then_success_guard
AFTER UPDATE OF status ON public.transaksi_provider
FOR EACH ROW
EXECUTE FUNCTION public.fn_provider_failed_then_success_guard();
