-- Speed up member dashboard statistics used by /v1/stats.
-- Run each statement outside a transaction because CREATE INDEX CONCURRENTLY requires it.

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_tm_member_status_created_stats
  ON public.transaksi_member (member_id, status, dibuat_pada DESC)
  INCLUDE (biaya_aktual, biaya_perkiraan);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_md_member_created_stats
  ON public.mutasi_dompet (member_id, dibuat_pada DESC)
  INCLUDE (arah, jumlah);
