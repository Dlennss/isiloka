\set ON_ERROR_STOP on
BEGIN;

-- Never run the fresh-install path against a populated or unrelated database.
DO $$ BEGIN
  IF current_database() <> 'isiloka' OR EXISTS (
    SELECT 1 FROM pg_tables WHERE schemaname = 'public'
  ) THEN
    RAISE EXCEPTION 'Bootstrap requires the empty isiloka database';
  END IF;
END $$;

\ir ../pulsa-be/sql/00000000_local_bootstrap_full_schema.sql
\ir ../pulsa-be/sql/20260424_alter_produk_provider_map_add_mode.sql
\ir ../pulsa-be/sql/20260715_add_produk_app_pricing_compat_columns.sql
\ir ../pulsa-be/sql/20260717_fix_app_order_guest_refund_schema.sql
\ir ../pulsa-be/sql/20260718_fix_internal_finance_schema.sql
\ir ../pulsa-be/sql/20260718_fix_local_schema_compat.sql
\ir ../pulsa-be/sql/20260718_fix_provider_rekening_schema.sql
\ir ../pulsa-be/sql/20260724_add_member_phone.sql
\ir ../pulsa-be/sql/20260724_create_agent_credit_schema.sql
\ir ../pulsa-be/sql/20260727_add_agent_credit_analyst_workflow.sql
\ir ../pulsa-be/sql/20260804_agent_credit_available_balance.sql
\ir ../pulsa-be/sql/20260806_align_agent_credit_workflow.sql
\ir ../pulsa-be/sql/20260821_enforce_single_agent_balance.sql
\ir ../pulsa-be/sql/20260821_add_marketing_agent_relation.sql
\ir ../pulsa-be/sql/20260824_agent_credit_flexible_limit.sql

-- Remove only the reference seeds inserted above, before any application runs.
DELETE FROM public.kategori;
DELETE FROM public.brand;
DELETE FROM public.provider;
DELETE FROM public.agent_credit_rank;

DO $$
DECLARE
  item RECORD;
  populated BOOLEAN;
BEGIN
  FOR item IN SELECT tablename FROM pg_tables WHERE schemaname = 'public' LOOP
    EXECUTE format('SELECT EXISTS (SELECT 1 FROM public.%I)', item.tablename) INTO populated;
    IF populated THEN
      RAISE EXCEPTION 'Fresh database unexpectedly contains rows in %', item.tablename;
    END IF;
  END LOOP;
END $$;
COMMIT;
