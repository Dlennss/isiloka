ALTER TABLE public.deposit_request
  ADD COLUMN IF NOT EXISTS bank_id BIGINT REFERENCES public.bank(id),
  ADD COLUMN IF NOT EXISTS bank_nama TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS bank_nomor_rekening TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS bank_atas_nama TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS deposit_request_bank_id_idx
  ON public.deposit_request (bank_id);
