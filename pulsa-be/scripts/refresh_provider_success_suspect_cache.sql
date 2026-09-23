SET lock_timeout = '5s';
SET statement_timeout = '45s';
SET client_min_messages = warning;

CREATE TABLE IF NOT EXISTS public.provider_success_suspect_cache (
  transaksi_provider_id BIGINT PRIMARY KEY,
  candidate_at TIMESTAMPTZ NOT NULL,
  source TEXT NOT NULL DEFAULT 'manual_refresh',
  refreshed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_provider_success_suspect_cache_candidate
ON public.provider_success_suspect_cache (candidate_at DESC, transaksi_provider_id DESC);

WITH params AS (
  SELECT
    now() - make_interval(days => GREATEST(1, :days_back)::int) AS from_at,
    now() + interval '1 day' AS to_at
),
candidate_text AS (
  SELECT DISTINCT
    tp0.id,
    tp0.dibuat_pada AS candidate_at,
    'text'::text AS source
  FROM public.transaksi_provider tp0
  JOIN public.transaksi_member tm0 ON tm0.id = tp0.transaksi_member_id
  CROSS JOIN params p
  WHERE tp0.dibuat_pada >= p.from_at
    AND tp0.dibuat_pada < p.to_at
    AND LOWER(TRIM(COALESCE(tm0.status, ''))) = 'success'
    AND (
      LOWER(TRIM(COALESCE(tp0.status, ''))) = 'success'
      OR (
        LOWER(TRIM(COALESCE(tp0.status, ''))) = 'failed'
        AND (
          LOWER(COALESCE(tp0.pesan, '')) LIKE '%refund%'
          OR LOWER(COALESCE(tp0.pesan, '')) LIKE '%saldo dikembalikan%'
          OR LOWER(COALESCE(tp0.pesan, '')) LIKE '%dikembalikan%'
          OR LOWER(COALESCE(tp0.pesan, '')) LIKE '%stok pulsa kembali%'
          OR LOWER(COALESCE(tp0.pesan, '')) LIKE '%pulsa kembali%'
        )
      )
    )
    AND (
      LOWER(COALESCE(tp0.pesan, '')) LIKE '%pernah selesai%'
      OR LOWER(COALESCE(tp0.pesan, '')) LIKE '%refund%'
      OR LOWER(COALESCE(tp0.pesan, '')) LIKE '%saldo dikembalikan%'
      OR LOWER(COALESCE(tp0.pesan, '')) LIKE '%stok pulsa kembali%'
      OR LOWER(COALESCE(tp0.pesan, '')) LIKE '%pulsa kembali%'
      OR LOWER(COALESCE(tp0.pesan, '')) LIKE '%sempat sukses%'
    )
),
candidate_all_failed AS (
  SELECT DISTINCT
    tp0.id,
    tp0.dibuat_pada AS candidate_at,
    'all_failed'::text AS source
  FROM public.transaksi_provider tp0
  JOIN public.transaksi_member tm0 ON tm0.id = tp0.transaksi_member_id
  CROSS JOIN params p
  WHERE tp0.dibuat_pada >= p.from_at
    AND tp0.dibuat_pada < p.to_at
    AND LOWER(TRIM(COALESCE(tp0.status, ''))) = 'failed'
    AND LOWER(TRIM(COALESCE(tm0.status, ''))) = 'success'
    AND NOT EXISTS (
      SELECT 1
      FROM public.transaksi_provider tp_other
      WHERE tp_other.transaksi_member_id = tp0.transaksi_member_id
        AND LOWER(TRIM(COALESCE(tp_other.status, ''))) <> 'failed'
    )
    AND NOT EXISTS (
      SELECT 1
      FROM public.transaksi_provider tp_newer
      WHERE tp_newer.transaksi_member_id = tp0.transaksi_member_id
        AND tp_newer.id > tp0.id
    )
),
candidate_anomaly AS (
  SELECT DISTINCT
    tp0.id,
    a0.dibuat_pada AS candidate_at,
    'anomaly'::text AS source
  FROM public.transaksi_anomasi_provider a0
  JOIN public.transaksi_provider tp0
    ON tp0.ref_id = a0.ref_id
   AND LOWER(TRIM(COALESCE(tp0.provider, ''))) = LOWER(TRIM(COALESCE(a0.provider, '')))
   AND a0.dibuat_pada >= tp0.dibuat_pada
  JOIN public.transaksi_member tm0 ON tm0.id = tp0.transaksi_member_id
  CROSS JOIN params p
  WHERE a0.dibuat_pada >= p.from_at
    AND a0.dibuat_pada < p.to_at
    AND LOWER(TRIM(COALESCE(tm0.status, ''))) = 'success'
    AND LOWER(TRIM(COALESCE(tp0.status, ''))) IN ('success', 'failed')
    AND (
      LOWER(COALESCE(a0.pesan, '')) LIKE '%gagal%'
      OR LOWER(COALESCE(a0.pesan, '')) LIKE '%failed%'
      OR LOWER(COALESCE(a0.pesan, '')) LIKE '%batal%'
      OR LOWER(COALESCE(a0.pesan, '')) LIKE '%ditolak%'
      OR (
        COALESCE(BTRIM(a0.kode_respon), '') <> ''
        AND LOWER(TRIM(COALESCE(a0.kode_respon, ''))) NOT IN ('00', '0', '9', '28', '68')
      )
    )
),
candidate_loket_prior AS (
  SELECT DISTINCT
    tp0.id,
    tp0.dibuat_pada AS candidate_at,
    'loket_prior'::text AS source
  FROM public.transaksi_provider tp0
  JOIN public.transaksi_member tm0 ON tm0.id = tp0.transaksi_member_id
  CROSS JOIN params p
  JOIN LATERAL (
    SELECT tp_prior.id
    FROM public.transaksi_provider tp_prior
    WHERE tp_prior.transaksi_member_id = tp0.transaksi_member_id
      AND tp_prior.id < tp0.id
      AND LOWER(TRIM(COALESCE(tp_prior.status, ''))) = 'failed'
      AND (
        LOWER(COALESCE(tp_prior.pesan, '')) LIKE '%pernah selesai%'
        OR LOWER(COALESCE(tp_prior.pesan, '')) LIKE '%refund%'
        OR LOWER(COALESCE(tp_prior.pesan, '')) LIKE '%saldo dikembalikan%'
        OR LOWER(COALESCE(tp_prior.pesan, '')) LIKE '%stok pulsa kembali%'
        OR LOWER(COALESCE(tp_prior.pesan, '')) LIKE '%pulsa kembali%'
        OR LOWER(COALESCE(tp_prior.pesan, '')) LIKE '%sempat sukses%'
      )
    ORDER BY tp_prior.id DESC
    LIMIT 1
  ) prior0 ON true
  WHERE tp0.dibuat_pada >= p.from_at
    AND tp0.dibuat_pada < p.to_at
    AND LOWER(TRIM(COALESCE(tp0.provider, ''))) = 'loketbayar'
    AND LOWER(TRIM(COALESCE(tp0.status, ''))) = 'failed'
    AND LOWER(TRIM(COALESCE(tm0.status, ''))) = 'success'
    AND COALESCE(BTRIM(tp0.pesan), '') <> ''
    AND NOT EXISTS (
      SELECT 1
      FROM public.transaksi_provider newer
      WHERE newer.transaksi_member_id = tp0.transaksi_member_id
        AND newer.id > tp0.id
    )
),
all_candidates AS (
  SELECT id, candidate_at, source FROM candidate_text
  UNION ALL
  SELECT id, candidate_at, source FROM candidate_all_failed
  UNION ALL
  SELECT id, candidate_at, source FROM candidate_anomaly
  UNION ALL
  SELECT id, candidate_at, source FROM candidate_loket_prior
),
merged_candidates AS (
  SELECT
    id,
    MAX(candidate_at) AS candidate_at,
    STRING_AGG(DISTINCT source, ',' ORDER BY source) AS source
  FROM all_candidates
  GROUP BY id
)
INSERT INTO public.provider_success_suspect_cache
  (transaksi_provider_id, candidate_at, source, refreshed_at)
SELECT id, candidate_at, source, now()
FROM merged_candidates
ON CONFLICT (transaksi_provider_id) DO UPDATE
SET candidate_at = EXCLUDED.candidate_at,
    source = EXCLUDED.source,
    refreshed_at = now();

DELETE FROM public.provider_success_suspect_cache
WHERE refreshed_at < now() - interval '45 days';

SELECT COUNT(*), MIN(candidate_at), MAX(candidate_at), MAX(refreshed_at)
FROM public.provider_success_suspect_cache;
