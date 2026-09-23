-- Read-only preflight checks for P24 SMPAY source marker.
-- Safe to run against production. This file must not contain writes.

SELECT 'smpay_ref_sources_exists' AS check_name,
       COALESCE(to_regclass('public.smpay_ref_sources')::text, '') AS value;

SELECT 'smpay_transaction_sources_exists' AS check_name,
       COALESCE(to_regclass('public.smpay_transaction_sources')::text, '') AS value;

SELECT 'transaksi_member_smpay_columns' AS check_name,
       count(*)::text AS value
FROM information_schema.columns
WHERE table_schema = 'public'
  AND table_name = 'transaksi_member'
  AND column_name LIKE 'smpay%';

SELECT 'h2h_commission_ledger_rows' AS check_name,
       count(*)::text AS value
FROM public.h2h_commission_ledger;

SELECT 'h2h_commission_recent_7d' AS check_name,
       count(*) || '|amount=' || COALESCE(sum(amount), 0) AS value
FROM public.h2h_commission_ledger
WHERE created_at >= now() - interval '7 days';

