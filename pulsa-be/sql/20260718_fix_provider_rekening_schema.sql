ALTER TABLE public.provider_rekening
  ADD COLUMN IF NOT EXISTS nama TEXT,
  ADD COLUMN IF NOT EXISTS bank TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS nomor_rekening TEXT,
  ADD COLUMN IF NOT EXISTS nomor_rekening_digits TEXT,
  ADD COLUMN IF NOT EXISTS catatan TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS dibuat_pada TIMESTAMPTZ NOT NULL DEFAULT now(),
  ADD COLUMN IF NOT EXISTS diubah_pada TIMESTAMPTZ NOT NULL DEFAULT now();

UPDATE public.provider_rekening
SET nama = COALESCE(NULLIF(trim(nama), ''), NULLIF(trim(account_name), ''), provider)
WHERE nama IS NULL OR trim(nama) = '';

UPDATE public.provider_rekening
SET bank = COALESCE(NULLIF(trim(bank), ''), NULLIF(trim(bank_name), ''), '')
WHERE trim(bank) = '';

UPDATE public.provider_rekening
SET nomor_rekening = COALESCE(NULLIF(trim(nomor_rekening), ''), NULLIF(trim(account_no), ''), '')
WHERE nomor_rekening IS NULL OR trim(nomor_rekening) = '';

UPDATE public.provider_rekening
SET nomor_rekening_digits = regexp_replace(COALESCE(nomor_rekening, account_no, ''), '[^0-9]', '', 'g')
WHERE nomor_rekening_digits IS NULL OR trim(nomor_rekening_digits) = '';

UPDATE public.provider_rekening
SET dibuat_pada = created_at
WHERE created_at IS NOT NULL
  AND dibuat_pada IS NULL;

UPDATE public.provider_rekening
SET diubah_pada = updated_at
WHERE updated_at IS NOT NULL
  AND diubah_pada IS NULL;

ALTER TABLE public.provider_rekening
  ALTER COLUMN nama SET NOT NULL,
  ALTER COLUMN nomor_rekening SET NOT NULL,
  ALTER COLUMN nomor_rekening_digits SET NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS provider_rekening_provider_nomor_uidx
  ON public.provider_rekening ((lower(trim(provider))), nomor_rekening_digits);

CREATE INDEX IF NOT EXISTS provider_rekening_nomor_digits_idx
  ON public.provider_rekening (nomor_rekening_digits);

CREATE INDEX IF NOT EXISTS provider_rekening_provider_idx
  ON public.provider_rekening ((lower(trim(provider))));

CREATE INDEX IF NOT EXISTS provider_rekening_aktif_provider_idx
  ON public.provider_rekening (aktif, (lower(trim(provider))));
