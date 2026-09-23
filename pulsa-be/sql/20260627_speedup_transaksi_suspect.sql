-- Run outside a transaction block. These indexes support
-- /dashboard/*/transaksi/transaksi-suspect fast pagination.

CREATE TABLE IF NOT EXISTS public.provider_success_suspect_cache (
  transaksi_provider_id BIGINT PRIMARY KEY,
  candidate_at TIMESTAMPTZ NOT NULL,
  source TEXT NOT NULL DEFAULT 'manual_refresh',
  refreshed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_provider_success_suspect_cache_candidate
ON public.provider_success_suspect_cache (candidate_at DESC, transaksi_provider_id DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_tp_status_norm_created_id
ON public.transaksi_provider (
  (lower(TRIM(BOTH FROM COALESCE(status, ''::character varying)))),
  dibuat_pada DESC,
  id DESC
);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_tp_provider_status_norm_created_id
ON public.transaksi_provider (
  (lower(TRIM(BOTH FROM COALESCE(provider, ''::text)))),
  (lower(TRIM(BOTH FROM COALESCE(status, ''::character varying)))),
  dibuat_pada DESC,
  id DESC
);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_tp_smb_bifastopen2_success_member
ON public.transaksi_provider (transaksi_member_id, id DESC)
WHERE lower(TRIM(BOTH FROM COALESCE(provider, ''::text))) = 'smb'
  AND lower(TRIM(BOTH FROM COALESCE(status, ''::character varying))) = 'success'
  AND upper(TRIM(BOTH FROM COALESCE(kode_produk, ''::text))) LIKE 'BIFASTOPEN2:%';
