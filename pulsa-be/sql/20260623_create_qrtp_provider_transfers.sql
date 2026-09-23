CREATE TABLE IF NOT EXISTS public.qrtp_provider_transfers (
  id bigserial PRIMARY KEY,
  ref_id text NOT NULL UNIQUE,
  inquiry_order_id text NOT NULL UNIQUE,
  provider text NOT NULL,
  provider_rekening_id bigint NOT NULL REFERENCES public.provider_rekening(id),
  bank_id bigint NOT NULL REFERENCES public.bank(id),
  bank_code text NOT NULL,
  bank_name text NOT NULL,
  account_no text NOT NULL,
  account_name text NOT NULL DEFAULT '',
  amount bigint NOT NULL CHECK (amount > 0),
  admin_fee bigint NOT NULL DEFAULT 2500 CHECK (admin_fee >= 0),
  note text NOT NULL DEFAULT '',
  status text NOT NULL DEFAULT 'inquiry_success',
  inquiry_status text NOT NULL DEFAULT '',
  account_status text NOT NULL DEFAULT '',
  inquiry_public_id text NOT NULL DEFAULT '',
  inquiry_error text NOT NULL DEFAULT '',
  inquiry_raw jsonb NOT NULL DEFAULT '{}'::jsonb,
  payout_public_id text NOT NULL DEFAULT '',
  provider_transaction_id text NOT NULL DEFAULT '',
  payout_error text NOT NULL DEFAULT '',
  payout_reason text NOT NULL DEFAULT '',
  payout_raw jsonb NOT NULL DEFAULT '{}'::jsonb,
  bank_saldo_after bigint,
  provider_saldo_after bigint,
  created_by bigint REFERENCES public.member(id),
  processed_at timestamptz,
  completed_at timestamptz,
  reversed_at timestamptz,
  callback_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS qrtp_provider_transfers_provider_idx ON public.qrtp_provider_transfers (provider, id DESC);
CREATE INDEX IF NOT EXISTS qrtp_provider_transfers_status_idx ON public.qrtp_provider_transfers (status, id DESC);
CREATE INDEX IF NOT EXISTS qrtp_provider_transfers_created_idx ON public.qrtp_provider_transfers (created_at DESC);

INSERT INTO public.bank (nama, nomor_rekening, atas_nama, saldo, aktif, admin_staff_only, dibuat_pada, diubah_pada)
SELECT 'QRTP', '', 'System QRTP', 0, true, false, now(), now()
WHERE NOT EXISTS (
  SELECT 1 FROM public.bank WHERE lower(trim(nama)) = lower(trim('QRTP'))
);
