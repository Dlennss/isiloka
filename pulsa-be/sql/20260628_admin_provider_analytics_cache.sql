CREATE TABLE IF NOT EXISTS public.admin_provider_analytics_daily_cache (
  provider text NOT NULL,
  day date NOT NULL,
  month_key text NOT NULL,
  total bigint NOT NULL DEFAULT 0,
  success bigint NOT NULL DEFAULT 0,
  failed bigint NOT NULL DEFAULT 0,
  sum_qty bigint NOT NULL DEFAULT 0,
  sum_harga bigint NOT NULL DEFAULT 0,
  success_nominal bigint NOT NULL DEFAULT 0,
  deposit_amount bigint NOT NULL DEFAULT 0,
  refreshed_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (provider, day)
);

CREATE INDEX IF NOT EXISTS admin_provider_analytics_daily_cache_day_idx
  ON public.admin_provider_analytics_daily_cache (day DESC, provider);

CREATE OR REPLACE FUNCTION public.refresh_admin_provider_analytics_cache(
  p_days integer DEFAULT 93,
  p_min_age interval DEFAULT '2 minutes'::interval,
  p_include_today boolean DEFAULT false
) RETURNS integer
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = public
AS $$
DECLARE
  v_days integer;
  v_from_date date;
  v_to_date date;
  v_today date;
  v_rows integer := 0;
BEGIN
  v_days := LEAST(GREATEST(COALESCE(p_days, 93), 1), 366);
  v_today := (now() AT TIME ZONE 'Asia/Jakarta')::date;
  v_to_date := CASE WHEN COALESCE(p_include_today, false) THEN v_today + 1 ELSE v_today END;
  v_from_date := v_to_date - v_days;

  IF v_to_date <= v_from_date THEN
    RETURN 0;
  END IF;

  IF p_min_age IS NOT NULL
     AND p_min_age > interval '0 seconds'
     AND EXISTS (
       SELECT 1
       FROM public.admin_provider_analytics_daily_cache
       WHERE day >= v_from_date
         AND day < v_to_date
         AND refreshed_at >= now() - p_min_age
       LIMIT 1
     ) THEN
    RETURN 0;
  END IF;

  DELETE FROM public.admin_provider_analytics_daily_cache
  WHERE day >= v_from_date
    AND day < v_to_date;

  INSERT INTO public.admin_provider_analytics_daily_cache (
    provider,
    day,
    month_key,
    total,
    success,
    failed,
    sum_qty,
    sum_harga,
    success_nominal,
    deposit_amount,
    refreshed_at
  )
  WITH bounds AS (
    SELECT v_from_date::timestamptz AS from_ts, v_to_date::timestamptz AS to_ts
  ),
  trx AS (
    SELECT
      coalesce(nullif(trim(t.provider), ''), '-') AS provider,
      date_trunc('day', t.dibuat_pada)::date AS day,
      count(*)::bigint AS total,
      count(*) FILTER (WHERE lower(trim(coalesce(t.status, ''))) = 'success')::bigint AS success,
      count(*) FILTER (WHERE lower(trim(coalesce(t.status, ''))) <> 'success')::bigint AS failed,
      coalesce(sum(t.qty), 0)::bigint AS sum_qty,
      coalesce(sum(coalesce(t.harga, 0)), 0)::bigint AS sum_harga,
      coalesce(sum(CASE WHEN lower(trim(coalesce(t.status, ''))) = 'success' THEN coalesce(t.harga, 0) ELSE 0 END), 0)::bigint AS success_nominal
    FROM public.transaksi_provider t
    CROSS JOIN bounds b
    WHERE t.dibuat_pada >= b.from_ts
      AND t.dibuat_pada < b.to_ts
    GROUP BY 1, 2
  ),
  dep AS (
    SELECT
      coalesce(nullif(trim(mdp.provider), ''), '-') AS provider,
      date_trunc('day', mdp.dibuat_pada)::date AS day,
      coalesce(sum(mdp.jumlah), 0)::bigint AS deposit_amount
    FROM public.mutasi_dompet_provider mdp
    CROSS JOIN bounds b
    WHERE mdp.dibuat_pada >= b.from_ts
      AND mdp.dibuat_pada < b.to_ts
      AND mdp.arah = 'credit'
      AND (
        mdp.alasan = 'PROVIDER_DEPOSIT' OR
        mdp.alasan = 'BANK_TRANSFER_IN' OR
        mdp.alasan = 'PROVIDER_ADJUST_CREDIT' OR
        mdp.alasan = 'PROVIDER_SET_BALANCE'
      )
    GROUP BY 1, 2
  ),
  provider_days AS (
    SELECT provider, day FROM trx
    UNION
    SELECT provider, day FROM dep
  )
  SELECT
    d.provider,
    d.day,
    to_char(d.day, 'YYYY-MM') AS month_key,
    coalesce(t.total, 0)::bigint AS total,
    coalesce(t.success, 0)::bigint AS success,
    coalesce(t.failed, 0)::bigint AS failed,
    coalesce(t.sum_qty, 0)::bigint AS sum_qty,
    coalesce(t.sum_harga, 0)::bigint AS sum_harga,
    coalesce(t.success_nominal, 0)::bigint AS success_nominal,
    coalesce(dep.deposit_amount, 0)::bigint AS deposit_amount,
    now() AS refreshed_at
  FROM provider_days d
  LEFT JOIN trx t ON t.provider = d.provider AND t.day = d.day
  LEFT JOIN dep ON dep.provider = d.provider AND dep.day = d.day;

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
      WHERE c.oid = 'public.admin_provider_analytics_daily_cache'::regclass
        AND r.rolname <> 'syarif'
    ) THEN
      EXECUTE 'ALTER TABLE public.admin_provider_analytics_daily_cache OWNER TO syarif';
    END IF;

    IF EXISTS (
      SELECT 1
      FROM pg_proc p
      JOIN pg_roles r ON r.oid = p.proowner
      WHERE p.oid = 'public.refresh_admin_provider_analytics_cache(integer, interval, boolean)'::regprocedure
        AND r.rolname <> 'syarif'
    ) THEN
      EXECUTE 'ALTER FUNCTION public.refresh_admin_provider_analytics_cache(integer, interval, boolean) OWNER TO syarif';
    END IF;

    EXECUTE 'GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE public.admin_provider_analytics_daily_cache TO syarif';
    EXECUTE 'GRANT EXECUTE ON FUNCTION public.refresh_admin_provider_analytics_cache(integer, interval, boolean) TO syarif';
  END IF;
END $$;
