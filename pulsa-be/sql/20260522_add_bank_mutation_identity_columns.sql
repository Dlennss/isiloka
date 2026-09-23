ALTER TABLE public.mutasi_bank
  ADD COLUMN IF NOT EXISTS waktu_mutasi_bank TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS pengirim TEXT,
  ADD COLUMN IF NOT EXISTS penerima TEXT;

CREATE INDEX CONCURRENTLY IF NOT EXISTS mutasi_bank_bank_waktu_mutasi_idx
  ON public.mutasi_bank (bank_id, waktu_mutasi_bank DESC, id DESC)
  WHERE waktu_mutasi_bank IS NOT NULL;

CREATE INDEX CONCURRENTLY IF NOT EXISTS mutasi_bank_bank_identity_idx
  ON public.mutasi_bank (bank_id, arah, jumlah, waktu_mutasi_bank)
  WHERE waktu_mutasi_bank IS NOT NULL;
