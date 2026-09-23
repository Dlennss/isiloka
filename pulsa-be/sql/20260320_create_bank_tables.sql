CREATE TABLE IF NOT EXISTS public.bank (
  id BIGSERIAL PRIMARY KEY,
  nama TEXT NOT NULL,
  nomor_rekening TEXT NOT NULL DEFAULT '',
  atas_nama TEXT NOT NULL DEFAULT '',
  saldo BIGINT NOT NULL DEFAULT 0,
  aktif BOOLEAN NOT NULL DEFAULT TRUE,
  dibuat_pada TIMESTAMPTZ NOT NULL DEFAULT now(),
  diubah_pada TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS bank_nama_unique_idx
  ON public.bank ((lower(trim(nama))));

CREATE TABLE IF NOT EXISTS public.mutasi_bank (
  id BIGSERIAL PRIMARY KEY,
  bank_id BIGINT NOT NULL REFERENCES public.bank(id),
  ref_id TEXT NOT NULL,
  arah TEXT NOT NULL,
  jumlah BIGINT NOT NULL,
  alasan TEXT NOT NULL,
  catatan TEXT,
  saldo_sebelum BIGINT NOT NULL,
  saldo_sesudah BIGINT NOT NULL,
  provider TEXT,
  member_id BIGINT,
  diubah_oleh BIGINT REFERENCES public.member(id),
  dibuat_pada TIMESTAMPTZ NOT NULL DEFAULT now(),
  meta JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX IF NOT EXISTS mutasi_bank_bank_id_idx
  ON public.mutasi_bank (bank_id);

CREATE INDEX IF NOT EXISTS mutasi_bank_ref_id_idx
  ON public.mutasi_bank (ref_id);

CREATE INDEX IF NOT EXISTS mutasi_bank_dibuat_pada_idx
  ON public.mutasi_bank (dibuat_pada DESC);
