CREATE TABLE IF NOT EXISTS public.admin_daily_business_cache (
  scope text NOT NULL,
  day date NOT NULL,
  month_key text NOT NULL,
  transaction_count bigint NOT NULL DEFAULT 0,
  transaction_amount bigint NOT NULL DEFAULT 0,
  provider_payment_amount bigint NOT NULL DEFAULT 0,
  margin_amount bigint NOT NULL DEFAULT 0,
  commission_amount bigint NOT NULL DEFAULT 0,
  transaction_expense_amount bigint NOT NULL DEFAULT 0,
  member_deposit_amount bigint NOT NULL DEFAULT 0,
  provider_deposit_amount bigint NOT NULL DEFAULT 0,
  deposit_gap_amount bigint NOT NULL DEFAULT 0,
  profit_amount bigint NOT NULL DEFAULT 0,
  refreshed_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (scope, day),
  CONSTRAINT admin_daily_business_cache_scope_chk CHECK (scope IN ('all', 'retail', 'h2h'))
);

CREATE INDEX IF NOT EXISTS admin_daily_business_cache_day_scope_idx
  ON public.admin_daily_business_cache (day DESC, scope);

CREATE OR REPLACE FUNCTION public.refresh_admin_daily_business_cache(
  p_days integer DEFAULT 7,
  p_min_age interval DEFAULT '2 minutes'::interval
) RETURNS integer
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = public
AS $$
DECLARE
  v_days integer;
  v_from_date date;
  v_to_date date;
  v_rows integer := 0;
BEGIN
  v_days := LEAST(GREATEST(COALESCE(p_days, 7), 1), 366);
  v_from_date := ((now() AT TIME ZONE 'Asia/Jakarta')::date - v_days);
  v_to_date := (now() AT TIME ZONE 'Asia/Jakarta')::date;

  IF p_min_age IS NOT NULL
     AND p_min_age > interval '0 seconds'
     AND EXISTS (
       SELECT 1
       FROM public.admin_daily_business_cache
       WHERE day >= v_from_date
         AND refreshed_at >= now() - p_min_age
       LIMIT 1
     ) THEN
    RETURN 0;
  END IF;

  INSERT INTO public.admin_daily_business_cache (
    scope,
    day,
    month_key,
    transaction_count,
    transaction_amount,
    provider_payment_amount,
    margin_amount,
    commission_amount,
    transaction_expense_amount,
    member_deposit_amount,
    provider_deposit_amount,
    deposit_gap_amount,
    profit_amount,
    refreshed_at
  )
  WITH bounds AS (
    SELECT v_from_date::timestamp AS from_ts, v_to_date::timestamp AS to_ts
  ),
  retail_sales AS (
    SELECT
      date_trunc('day', o.dibuat_pada)::date AS day,
      COUNT(*)::bigint AS transaction_count,
      COALESCE(SUM(COALESCE(o.harga_final, 0)), 0)::bigint AS transaction_amount
    FROM public.app_order o
    CROSS JOIN bounds b
    WHERE lower(COALESCE(o.status, '')) = 'success'
      AND o.dibuat_pada >= b.from_ts
      AND o.dibuat_pada < b.to_ts
    GROUP BY 1
  ),
  retail_provider AS (
    SELECT
      date_trunc('day', apt.dibuat_pada)::date AS day,
      COALESCE(SUM(CASE WHEN COALESCE(apt.harga_provider, 0) > 0 THEN apt.harga_provider ELSE 0 END), 0)::bigint AS provider_payment_amount
    FROM public.app_order_provider_trx apt
    CROSS JOIN bounds b
    WHERE lower(COALESCE(apt.status, '')) = 'success'
      AND apt.dibuat_pada >= b.from_ts
      AND apt.dibuat_pada < b.to_ts
    GROUP BY 1
  ),
  retail_commission AS (
    SELECT
      date_trunc('day', rcl.created_at)::date AS day,
      COALESCE(SUM(rcl.amount), 0)::bigint AS commission_amount
    FROM public.retail_commission_ledger rcl
    CROSS JOIN bounds b
    WHERE rcl.created_at >= b.from_ts
      AND rcl.created_at < b.to_ts
    GROUP BY 1
  ),
  retail_days AS (
    SELECT day FROM retail_sales
    UNION
    SELECT day FROM retail_provider
    UNION
    SELECT day FROM retail_commission
  ),
  retail_rows AS (
    SELECT
      'retail'::text AS scope,
      d.day,
      COALESCE(rs.transaction_count, 0)::bigint AS transaction_count,
      COALESCE(rs.transaction_amount, 0)::bigint AS transaction_amount,
      COALESCE(rp.provider_payment_amount, 0)::bigint AS provider_payment_amount,
      (COALESCE(rs.transaction_amount, 0) - COALESCE(rp.provider_payment_amount, 0))::bigint AS margin_amount,
      COALESCE(rc.commission_amount, 0)::bigint AS commission_amount,
      0::bigint AS transaction_expense_amount
    FROM retail_days d
    LEFT JOIN retail_sales rs ON rs.day = d.day
    LEFT JOIN retail_provider rp ON rp.day = d.day
    LEFT JOIN retail_commission rc ON rc.day = d.day
  ),
  h2h_sales AS (
    SELECT
      date_trunc('day', tm.dibuat_pada)::date AS day,
      COUNT(*)::bigint AS transaction_count,
      COALESCE(SUM(COALESCE(NULLIF(tm.biaya_aktual, 0), tm.harga_member, 0)), 0)::bigint AS transaction_amount
    FROM public.transaksi_member tm
    CROSS JOIN bounds b
    WHERE lower(COALESCE(tm.status, '')) = 'success'
      AND tm.dibuat_pada >= b.from_ts
      AND tm.dibuat_pada < b.to_ts
    GROUP BY 1
  ),
  h2h_provider AS (
    SELECT
      date_trunc('day', tp.dibuat_pada)::date AS day,
      COALESCE(SUM(CASE WHEN COALESCE(tp.harga, 0) > 0 THEN tp.harga ELSE 0 END), 0)::bigint AS provider_payment_amount
    FROM public.transaksi_provider tp
    CROSS JOIN bounds b
    WHERE lower(COALESCE(tp.status, '')) = 'success'
      AND tp.dibuat_pada >= b.from_ts
      AND tp.dibuat_pada < b.to_ts
    GROUP BY 1
  ),
  h2h_commission AS (
    SELECT
      date_trunc('day', hcl.created_at)::date AS day,
      COALESCE(SUM(hcl.amount), 0)::bigint AS commission_amount
    FROM public.h2h_commission_ledger hcl
    CROSS JOIN bounds b
    WHERE hcl.created_at >= b.from_ts
      AND hcl.created_at < b.to_ts
    GROUP BY 1
  ),
  h2h_transaction_expense AS (
    SELECT
      date_trunc('day', COALESCE(mb.waktu_mutasi_bank, mb.dibuat_pada) AT TIME ZONE 'Asia/Jakarta')::date AS day,
      COALESCE(SUM(CASE WHEN mb.arah = 'DEBIT' THEN mb.jumlah ELSE 0 END), 0)::bigint AS transaction_expense_amount
    FROM public.mutasi_bank mb
    CROSS JOIN bounds b
    WHERE mb.arah = 'DEBIT'
      AND mb.alasan = 'TRANSAKSI_SUSPECT_SELESAI'
      AND COALESCE(mb.waktu_mutasi_bank, mb.dibuat_pada) >= b.from_ts
      AND COALESCE(mb.waktu_mutasi_bank, mb.dibuat_pada) < b.to_ts
    GROUP BY 1
  ),
  h2h_days AS (
    SELECT day FROM h2h_sales
    UNION
    SELECT day FROM h2h_provider
    UNION
    SELECT day FROM h2h_commission
    UNION
    SELECT day FROM h2h_transaction_expense
  ),
  h2h_rows AS (
    SELECT
      'h2h'::text AS scope,
      d.day,
      COALESCE(hs.transaction_count, 0)::bigint AS transaction_count,
      COALESCE(hs.transaction_amount, 0)::bigint AS transaction_amount,
      COALESCE(hp.provider_payment_amount, 0)::bigint AS provider_payment_amount,
      (COALESCE(hs.transaction_amount, 0) - COALESCE(hp.provider_payment_amount, 0))::bigint AS margin_amount,
      COALESCE(hc.commission_amount, 0)::bigint AS commission_amount,
      COALESCE(he.transaction_expense_amount, 0)::bigint AS transaction_expense_amount
    FROM h2h_days d
    LEFT JOIN h2h_sales hs ON hs.day = d.day
    LEFT JOIN h2h_provider hp ON hp.day = d.day
    LEFT JOIN h2h_commission hc ON hc.day = d.day
    LEFT JOIN h2h_transaction_expense he ON he.day = d.day
  ),
  member_deposit_rows AS (
    SELECT
      CASE
        WHEN lower(COALESCE(m.role, '')) IN ('user', 'agent', 'master') THEN 'retail'
        WHEN lower(COALESCE(m.role, '')) IN ('member', 'agent_member', 'master_member') THEN 'h2h'
        ELSE 'other'
      END AS scope,
      date_trunc('day', dr.dibuat_pada)::date AS day,
      COALESCE(SUM(COALESCE(dr.approved_amount, dr.amount)), 0)::bigint AS member_deposit_amount
    FROM public.deposit_request dr
    CROSS JOIN bounds b
    LEFT JOIN public.member m ON m.id = dr.member_id
    WHERE lower(COALESCE(dr.status, '')) = 'approved'
      AND dr.dibuat_pada >= b.from_ts
      AND dr.dibuat_pada < b.to_ts
    GROUP BY 1, 2
  ),
  provider_deposit_rows AS (
    SELECT
      date_trunc('day', COALESCE(mb.waktu_mutasi_bank, mb.dibuat_pada) AT TIME ZONE 'Asia/Jakarta')::date AS day,
      COALESCE(SUM(mdp.jumlah), 0)::bigint AS provider_deposit_amount
    FROM public.mutasi_dompet_provider mdp
    CROSS JOIN bounds b
    JOIN public.mutasi_bank mb
      ON mb.ref_id = mdp.ref_id
     AND mb.jumlah = mdp.jumlah
     AND lower(trim(COALESCE(mb.provider, ''))) = lower(trim(COALESCE(mdp.provider, '')))
     AND mb.arah = 'DEBIT'
     AND mb.alasan = 'BANK_TRANSFER_TO_PROVIDER'
    WHERE lower(COALESCE(mdp.arah, '')) = 'credit'
      AND COALESCE(mdp.alasan, '') = 'BANK_TRANSFER_IN'
      AND COALESCE(mb.waktu_mutasi_bank, mb.dibuat_pada) >= b.from_ts
      AND COALESCE(mb.waktu_mutasi_bank, mb.dibuat_pada) < b.to_ts
    GROUP BY 1
  ),
  base_rows AS (
    SELECT * FROM retail_rows
    UNION ALL
    SELECT * FROM h2h_rows
  ),
  scoped_rows AS (
    SELECT * FROM base_rows
    UNION ALL
    SELECT
      'all'::text AS scope,
      day,
      SUM(transaction_count)::bigint AS transaction_count,
      SUM(transaction_amount)::bigint AS transaction_amount,
      SUM(provider_payment_amount)::bigint AS provider_payment_amount,
      SUM(margin_amount)::bigint AS margin_amount,
      SUM(commission_amount)::bigint AS commission_amount,
      SUM(transaction_expense_amount)::bigint AS transaction_expense_amount
    FROM base_rows
    GROUP BY day
  ),
  member_deposits AS (
    SELECT
      scope,
      day,
      SUM(member_deposit_amount)::bigint AS member_deposit_amount
    FROM member_deposit_rows
    WHERE scope IN ('retail', 'h2h')
    GROUP BY scope, day
    UNION ALL
    SELECT
      'all'::text AS scope,
      day,
      SUM(member_deposit_amount)::bigint AS member_deposit_amount
    FROM member_deposit_rows
    WHERE scope IN ('retail', 'h2h')
    GROUP BY day
  )
  SELECT
    sr.scope,
    sr.day,
    TO_CHAR(sr.day, 'YYYY-MM') AS month_key,
    sr.transaction_count,
    sr.transaction_amount,
    sr.provider_payment_amount,
    sr.margin_amount,
    sr.commission_amount,
    sr.transaction_expense_amount,
    COALESCE(md.member_deposit_amount, 0)::bigint AS member_deposit_amount,
    COALESCE(pd.provider_deposit_amount, 0)::bigint AS provider_deposit_amount,
    (COALESCE(md.member_deposit_amount, 0) - COALESCE(pd.provider_deposit_amount, 0))::bigint AS deposit_gap_amount,
    (sr.margin_amount - sr.commission_amount - sr.transaction_expense_amount)::bigint AS profit_amount,
    now() AS refreshed_at
  FROM scoped_rows sr
  LEFT JOIN member_deposits md ON md.scope = sr.scope AND md.day = sr.day
  LEFT JOIN provider_deposit_rows pd ON pd.day = sr.day
  ON CONFLICT (scope, day) DO UPDATE SET
    month_key = EXCLUDED.month_key,
    transaction_count = EXCLUDED.transaction_count,
    transaction_amount = EXCLUDED.transaction_amount,
    provider_payment_amount = EXCLUDED.provider_payment_amount,
    margin_amount = EXCLUDED.margin_amount,
    commission_amount = EXCLUDED.commission_amount,
    transaction_expense_amount = EXCLUDED.transaction_expense_amount,
    member_deposit_amount = EXCLUDED.member_deposit_amount,
    provider_deposit_amount = EXCLUDED.provider_deposit_amount,
    deposit_gap_amount = EXCLUDED.deposit_gap_amount,
    profit_amount = EXCLUDED.profit_amount,
    refreshed_at = EXCLUDED.refreshed_at;

  GET DIAGNOSTICS v_rows = ROW_COUNT;
  RETURN v_rows;
END;
$$;

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'syarif') THEN
    IF EXISTS (
      SELECT 1
      FROM pg_class c
      JOIN pg_roles r ON r.oid = c.relowner
      WHERE c.oid = 'public.admin_daily_business_cache'::regclass
        AND r.rolname <> 'syarif'
    ) THEN
      EXECUTE 'ALTER TABLE public.admin_daily_business_cache OWNER TO syarif';
    END IF;

    IF EXISTS (
      SELECT 1
      FROM pg_proc p
      JOIN pg_roles r ON r.oid = p.proowner
      WHERE p.oid = 'public.refresh_admin_daily_business_cache(integer, interval)'::regprocedure
        AND r.rolname <> 'syarif'
    ) THEN
      EXECUTE 'ALTER FUNCTION public.refresh_admin_daily_business_cache(integer, interval) OWNER TO syarif';
    END IF;

    EXECUTE 'GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE public.admin_daily_business_cache TO syarif';
    EXECUTE 'GRANT EXECUTE ON FUNCTION public.refresh_admin_daily_business_cache(integer, interval) TO syarif';
  END IF;
END $$;
