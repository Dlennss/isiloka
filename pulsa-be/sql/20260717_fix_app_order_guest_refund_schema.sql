CREATE TABLE IF NOT EXISTS public.app_order_guest_refund (
  id BIGSERIAL PRIMARY KEY,
  app_order_id BIGINT NOT NULL REFERENCES public.app_order(id) ON DELETE CASCADE,
  invoice_id TEXT NOT NULL,
  guest_nama TEXT,
  guest_email TEXT NOT NULL DEFAULT '',
  guest_phone TEXT NOT NULL DEFAULT '',
  amount_refund BIGINT NOT NULL DEFAULT 0,
  status TEXT NOT NULL DEFAULT 'pending',
  reason TEXT,
  notes TEXT,
  claimed_member_id BIGINT,
  claimed_at TIMESTAMPTZ,
  processed_at TIMESTAMPTZ,
  dibuat_pada TIMESTAMPTZ NOT NULL DEFAULT now(),
  diubah_pada TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(app_order_id)
);

ALTER TABLE public.app_order_guest_refund
  ADD COLUMN IF NOT EXISTS guest_nama TEXT,
  ADD COLUMN IF NOT EXISTS amount_refund BIGINT NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS reason TEXT,
  ADD COLUMN IF NOT EXISTS notes TEXT,
  ADD COLUMN IF NOT EXISTS claimed_member_id BIGINT,
  ADD COLUMN IF NOT EXISTS claimed_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS processed_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS dibuat_pada TIMESTAMPTZ NOT NULL DEFAULT now(),
  ADD COLUMN IF NOT EXISTS diubah_pada TIMESTAMPTZ NOT NULL DEFAULT now();

DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_schema = 'public'
      AND table_name = 'app_order_guest_refund'
      AND column_name = 'amount'
  ) THEN
    EXECUTE 'UPDATE public.app_order_guest_refund SET amount_refund = amount WHERE COALESCE(amount_refund, 0) = 0 AND amount IS NOT NULL';
  END IF;

  IF EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_schema = 'public'
      AND table_name = 'app_order_guest_refund'
      AND column_name = 'note'
  ) THEN
    EXECUTE 'UPDATE public.app_order_guest_refund SET notes = note WHERE notes IS NULL AND note IS NOT NULL';
  END IF;

  IF EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_schema = 'public'
      AND table_name = 'app_order_guest_refund'
      AND column_name = 'created_at'
  ) THEN
    EXECUTE 'UPDATE public.app_order_guest_refund SET dibuat_pada = created_at WHERE created_at IS NOT NULL';
  END IF;

  IF EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_schema = 'public'
      AND table_name = 'app_order_guest_refund'
      AND column_name = 'updated_at'
  ) THEN
    EXECUTE 'UPDATE public.app_order_guest_refund SET diubah_pada = updated_at WHERE updated_at IS NOT NULL';
  END IF;
END $$;

UPDATE public.app_order_guest_refund gr
SET guest_nama = ao.guest_nama
FROM public.app_order ao
WHERE ao.id = gr.app_order_id
  AND gr.guest_nama IS NULL
  AND ao.guest_nama IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS app_order_guest_refund_app_order_uidx
  ON public.app_order_guest_refund (app_order_id);

CREATE INDEX IF NOT EXISTS app_order_guest_refund_status_created_idx
  ON public.app_order_guest_refund (status, dibuat_pada DESC);

CREATE INDEX IF NOT EXISTS app_order_guest_refund_invoice_idx
  ON public.app_order_guest_refund (invoice_id);
