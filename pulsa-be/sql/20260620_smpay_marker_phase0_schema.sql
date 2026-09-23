-- P24 SMPAY source marker phase-0 schema.
-- Draft only. Do not apply to production before staging/dry-run approval.
-- This file only adds marker tables. It does not change H2H commission behavior.

BEGIN;
SET LOCAL lock_timeout = '3s';
SET LOCAL statement_timeout = '2min';

CREATE TABLE IF NOT EXISTS public.smpay_ref_sources (
  id bigserial PRIMARY KEY,
  member_id bigint NOT NULL,
  ref_id text NOT NULL,
  smpay_transaction_id bigint NULL,
  smpay_website_id bigint NULL,
  smpay_division_id bigint NULL,
  source_system text NOT NULL DEFAULT 'SMPAY',
  skip_h2h_commission boolean NOT NULL DEFAULT true,
  raw_request jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT smpay_ref_sources_ref_not_blank CHECK (btrim(ref_id) <> ''),
  CONSTRAINT smpay_ref_sources_system_chk CHECK (source_system IN ('SMPAY'))
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_smpay_ref_sources_member_ref
  ON public.smpay_ref_sources (member_id, ref_id);

CREATE INDEX IF NOT EXISTS idx_smpay_ref_sources_ref
  ON public.smpay_ref_sources (ref_id);

CREATE INDEX IF NOT EXISTS idx_smpay_ref_sources_smpay_tx
  ON public.smpay_ref_sources (smpay_transaction_id)
  WHERE smpay_transaction_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS public.smpay_transaction_sources (
  transaksi_member_id bigint PRIMARY KEY,
  smpay_ref_source_id bigint NULL REFERENCES public.smpay_ref_sources(id),
  ref_id text NOT NULL,
  smpay_transaction_id bigint NULL,
  smpay_website_id bigint NULL,
  smpay_division_id bigint NULL,
  source_system text NOT NULL DEFAULT 'SMPAY',
  skip_h2h_commission boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT smpay_transaction_sources_ref_not_blank CHECK (btrim(ref_id) <> ''),
  CONSTRAINT smpay_transaction_sources_system_chk CHECK (source_system IN ('SMPAY'))
);

CREATE INDEX IF NOT EXISTS idx_smpay_transaction_sources_ref
  ON public.smpay_transaction_sources (ref_id);

CREATE INDEX IF NOT EXISTS idx_smpay_transaction_sources_smpay_tx
  ON public.smpay_transaction_sources (smpay_transaction_id)
  WHERE smpay_transaction_id IS NOT NULL;

GRANT SELECT, INSERT, UPDATE ON public.smpay_ref_sources, public.smpay_transaction_sources TO syarif;
GRANT USAGE, SELECT ON SEQUENCE public.smpay_ref_sources_id_seq TO syarif;

-- Optional FKs to existing high-traffic P24 tables should be added later in
-- a maintenance window after lock behavior is measured:
--
-- ALTER TABLE public.smpay_ref_sources
--   ADD CONSTRAINT smpay_ref_sources_member_fk
--   FOREIGN KEY (member_id) REFERENCES public.member(id) NOT VALID;
--
-- ALTER TABLE public.smpay_transaction_sources
--   ADD CONSTRAINT smpay_transaction_sources_trx_fk
--   FOREIGN KEY (transaksi_member_id) REFERENCES public.transaksi_member(id) NOT VALID;

COMMIT;
