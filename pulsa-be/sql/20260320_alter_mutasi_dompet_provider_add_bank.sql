ALTER TABLE public.mutasi_dompet_provider
  ADD COLUMN IF NOT EXISTS bank_id BIGINT NULL,
  ADD COLUMN IF NOT EXISTS bank_nama TEXT NOT NULL DEFAULT ''::text;

CREATE INDEX IF NOT EXISTS idx_mutasi_dompet_provider_bank_id
  ON public.mutasi_dompet_provider (bank_id);

ALTER TABLE public.mutasi_dompet_provider
  ADD CONSTRAINT fk_mutasi_dompet_provider_bank
  FOREIGN KEY (bank_id) REFERENCES public.bank(id) ON DELETE SET NULL;
