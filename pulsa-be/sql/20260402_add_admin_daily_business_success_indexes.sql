CREATE INDEX CONCURRENTLY IF NOT EXISTS app_order_success_dibuat_pada_idx
  ON public.app_order (dibuat_pada DESC, id DESC)
  WHERE lower(COALESCE(status, '')) = 'success';

CREATE INDEX CONCURRENTLY IF NOT EXISTS transaksi_member_success_dibuat_pada_idx
  ON public.transaksi_member (dibuat_pada DESC, id DESC)
  WHERE lower(COALESCE(status, '')) = 'success';
