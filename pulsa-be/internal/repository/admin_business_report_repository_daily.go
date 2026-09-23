package repository

import (
	"context"
	"errors"
	"strings"
	"time"
)

func (r *AdminBusinessReportRepository) ListDailyBusiness(ctx context.Context, in AdminDailyBusinessArgs) ([]AdminDailyBusinessRow, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("db not initialized")
	}
	if err := r.ensureDailyBusinessSupport(ctx); err != nil {
		return nil, err
	}

	scope := normalizeAdminReportScope(in.Scope)
	months := in.Months
	if months <= 0 || months > 12 {
		months = 3
	}

	if out, err := r.listDailyBusinessCached(ctx, scope, months, in); err == nil {
		return out, nil
	}

	rows, err := r.db.QueryContext(ctx, `
WITH retail_sales AS (
  SELECT
    date_trunc('day', o.dibuat_pada)::date AS day,
    COUNT(*)::bigint AS transaction_count,
    COALESCE(SUM(COALESCE(o.harga_final, 0)), 0)::bigint AS transaction_amount
  FROM public.app_order o
  WHERE lower(COALESCE(o.status, '')) = 'success'
    AND o.dibuat_pada >= CASE WHEN $3 THEN $4 ELSE date_trunc('day', now() - make_interval(months => $2)) END
    AND ($5 = false OR o.dibuat_pada < $6)
  GROUP BY 1
),
retail_provider AS (
  SELECT
    date_trunc('day', apt.dibuat_pada)::date AS day,
    COALESCE(SUM(CASE WHEN COALESCE(apt.harga_provider, 0) > 0 THEN apt.harga_provider ELSE 0 END), 0)::bigint AS provider_payment_amount
  FROM public.app_order_provider_trx apt
  WHERE lower(COALESCE(apt.status, '')) = 'success'
    AND apt.dibuat_pada >= CASE WHEN $3 THEN $4 ELSE date_trunc('day', now() - make_interval(months => $2)) END
    AND ($5 = false OR apt.dibuat_pada < $6)
  GROUP BY 1
),
retail_commission AS (
  SELECT
    date_trunc('day', rcl.created_at)::date AS day,
    COALESCE(SUM(rcl.amount), 0)::bigint AS commission_amount
  FROM public.retail_commission_ledger rcl
  WHERE rcl.created_at >= CASE WHEN $3 THEN $4 ELSE date_trunc('day', now() - make_interval(months => $2)) END
    AND ($5 = false OR rcl.created_at < $6)
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
  WHERE lower(COALESCE(tm.status, '')) = 'success'
    AND tm.dibuat_pada >= CASE WHEN $3 THEN $4 ELSE date_trunc('day', now() - make_interval(months => $2)) END
    AND ($5 = false OR tm.dibuat_pada < $6)
  GROUP BY 1
),
h2h_provider AS (
  SELECT
    date_trunc('day', tp.dibuat_pada)::date AS day,
    COALESCE(SUM(CASE WHEN COALESCE(tp.harga, 0) > 0 THEN tp.harga ELSE 0 END), 0)::bigint AS provider_payment_amount
  FROM public.transaksi_provider tp
  WHERE lower(COALESCE(tp.status, '')) = 'success'
    AND tp.dibuat_pada >= CASE WHEN $3 THEN $4 ELSE date_trunc('day', now() - make_interval(months => $2)) END
    AND ($5 = false OR tp.dibuat_pada < $6)
  GROUP BY 1
),
h2h_commission AS (
  SELECT
    date_trunc('day', hcl.created_at)::date AS day,
    COALESCE(SUM(hcl.amount), 0)::bigint AS commission_amount
  FROM public.h2h_commission_ledger hcl
  WHERE hcl.created_at >= CASE WHEN $3 THEN $4 ELSE date_trunc('day', now() - make_interval(months => $2)) END
    AND ($5 = false OR hcl.created_at < $6)
  GROUP BY 1
),
h2h_transaction_expense AS (
  SELECT
    date_trunc('day', COALESCE(mb.waktu_mutasi_bank, mb.dibuat_pada) AT TIME ZONE 'Asia/Jakarta')::date AS day,
    COALESCE(SUM(CASE WHEN mb.arah = 'DEBIT' THEN mb.jumlah ELSE 0 END), 0)::bigint AS transaction_expense_amount
  FROM public.mutasi_bank mb
  WHERE mb.arah = 'DEBIT'
    AND mb.alasan = 'TRANSAKSI_SUSPECT_SELESAI'
    AND COALESCE(mb.waktu_mutasi_bank, mb.dibuat_pada) >= CASE WHEN $3 THEN $4 ELSE date_trunc('day', now() - make_interval(months => $2)) END
    AND ($5 = false OR COALESCE(mb.waktu_mutasi_bank, mb.dibuat_pada) < $6)
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
  LEFT JOIN public.member m ON m.id = dr.member_id
  WHERE lower(COALESCE(dr.status, '')) = 'approved'
    AND dr.dibuat_pada >= CASE WHEN $3 THEN $4 ELSE date_trunc('day', now() - make_interval(months => $2)) END
    AND ($5 = false OR dr.dibuat_pada < $6)
  GROUP BY 1, 2
),
provider_deposit_rows AS (
  SELECT
    date_trunc('day', COALESCE(mb.waktu_mutasi_bank, mb.dibuat_pada) AT TIME ZONE 'Asia/Jakarta')::date AS day,
    COALESCE(SUM(mdp.jumlah), 0)::bigint AS provider_deposit_amount
  FROM public.mutasi_dompet_provider mdp
  JOIN public.mutasi_bank mb
    ON mb.ref_id = mdp.ref_id
   AND mb.jumlah = mdp.jumlah
   AND lower(trim(COALESCE(mb.provider, ''))) = lower(trim(COALESCE(mdp.provider, '')))
   AND mb.arah = 'DEBIT'
   AND mb.alasan = 'BANK_TRANSFER_TO_PROVIDER'
  WHERE lower(COALESCE(mdp.arah, '')) = 'credit'
    AND COALESCE(mdp.alasan, '') = 'BANK_TRANSFER_IN'
    AND COALESCE(mb.waktu_mutasi_bank, mb.dibuat_pada) >= CASE WHEN $3 THEN $4 ELSE date_trunc('day', now() - make_interval(months => $2)) END
    AND ($5 = false OR COALESCE(mb.waktu_mutasi_bank, mb.dibuat_pada) < $6)
  GROUP BY 1
),
all_rows AS (
  SELECT * FROM retail_rows
  UNION ALL
  SELECT * FROM h2h_rows
),
grouped AS (
  SELECT
    CASE WHEN $1 = '' OR $1 = 'all' THEN 'all' ELSE scope END AS scope,
    day,
    TO_CHAR(day, 'YYYY-MM') AS month_key,
    SUM(transaction_count)::bigint AS transaction_count,
    SUM(transaction_amount)::bigint AS transaction_amount,
    SUM(provider_payment_amount)::bigint AS provider_payment_amount,
    SUM(margin_amount)::bigint AS margin_amount,
    SUM(commission_amount)::bigint AS commission_amount,
    SUM(transaction_expense_amount)::bigint AS transaction_expense_amount
  FROM all_rows
  WHERE ($1 = '' OR $1 = 'all' OR scope = $1)
  GROUP BY CASE WHEN $1 = '' OR $1 = 'all' THEN 'all' ELSE scope END, day
),
member_deposits AS (
  SELECT
    CASE WHEN $1 = '' OR $1 = 'all' THEN 'all' ELSE scope END AS scope,
    day,
    SUM(member_deposit_amount)::bigint AS member_deposit_amount
  FROM member_deposit_rows
  WHERE scope IN ('retail', 'h2h')
    AND ($1 = '' OR $1 = 'all' OR scope = $1)
  GROUP BY CASE WHEN $1 = '' OR $1 = 'all' THEN 'all' ELSE scope END, day
)
SELECT
  g.scope,
  g.day,
  g.month_key,
  g.transaction_count,
  g.transaction_amount,
  g.provider_payment_amount,
  g.margin_amount,
  g.commission_amount,
  g.transaction_expense_amount,
  COALESCE(md.member_deposit_amount, 0)::bigint AS member_deposit_amount,
  COALESCE(pd.provider_deposit_amount, 0)::bigint AS provider_deposit_amount,
  (COALESCE(md.member_deposit_amount, 0) - COALESCE(pd.provider_deposit_amount, 0))::bigint AS deposit_gap_amount,
  (g.margin_amount - g.commission_amount - g.transaction_expense_amount)::bigint AS profit_amount
FROM grouped g
LEFT JOIN member_deposits md ON md.scope = g.scope AND md.day = g.day
LEFT JOIN provider_deposit_rows pd ON pd.day = g.day
ORDER BY day DESC
`, scope, months, in.HasFrom, in.From, in.HasTo, in.To)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]AdminDailyBusinessRow, 0, 120)
	for rows.Next() {
		var item AdminDailyBusinessRow
		if err := rows.Scan(
			&item.Scope,
			&item.Day,
			&item.MonthKey,
			&item.TransactionCount,
			&item.TransactionAmount,
			&item.ProviderPaymentAmount,
			&item.MarginAmount,
			&item.CommissionAmount,
			&item.TransactionExpense,
			&item.MemberDepositAmount,
			&item.ProviderDepositAmount,
			&item.DepositGapAmount,
			&item.ProfitAmount,
		); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *AdminBusinessReportRepository) listDailyBusinessCached(ctx context.Context, scope string, months int, in AdminDailyBusinessArgs) ([]AdminDailyBusinessRow, error) {
	rows, err := r.db.QueryContext(ctx, `
WITH bounds AS (
  SELECT
    (CASE WHEN $3 THEN $4::timestamp ELSE date_trunc('day', now() - make_interval(months => $2)) END)::date AS from_date,
    (CASE WHEN $5 THEN $6::timestamp ELSE NULL::timestamp END)::date AS to_date,
    (now() AT TIME ZONE 'Asia/Jakarta')::date AS today_date
),
today_requested AS (
  SELECT
    today_date,
    today_date::timestamp AS from_ts,
    (today_date + 1)::timestamp AS to_ts
  FROM bounds
  WHERE today_date >= from_date
    AND ($5 = false OR today_date < to_date)
),
cached_rows AS (
  SELECT
    c.scope,
    c.day,
    c.month_key,
    c.transaction_count,
    c.transaction_amount,
    c.provider_payment_amount,
    c.margin_amount,
    c.commission_amount,
    c.transaction_expense_amount,
    c.member_deposit_amount,
    c.provider_deposit_amount,
    c.deposit_gap_amount,
    c.profit_amount
  FROM public.admin_daily_business_cache c
  CROSS JOIN bounds b
  WHERE c.scope = CASE WHEN $1 = '' OR $1 = 'all' THEN 'all' ELSE $1 END
    AND c.day >= b.from_date
    AND ($5 = false OR c.day < b.to_date)
    AND c.day < b.today_date
),
retail_sales AS (
  SELECT
    date_trunc('day', o.dibuat_pada)::date AS day,
    COUNT(*)::bigint AS transaction_count,
    COALESCE(SUM(COALESCE(o.harga_final, 0)), 0)::bigint AS transaction_amount
  FROM public.app_order o
  CROSS JOIN today_requested tr
  WHERE lower(COALESCE(o.status, '')) = 'success'
    AND o.dibuat_pada >= tr.from_ts
    AND o.dibuat_pada < tr.to_ts
  GROUP BY 1
),
retail_provider AS (
  SELECT
    date_trunc('day', apt.dibuat_pada)::date AS day,
    COALESCE(SUM(CASE WHEN COALESCE(apt.harga_provider, 0) > 0 THEN apt.harga_provider ELSE 0 END), 0)::bigint AS provider_payment_amount
  FROM public.app_order_provider_trx apt
  CROSS JOIN today_requested tr
  WHERE lower(COALESCE(apt.status, '')) = 'success'
    AND apt.dibuat_pada >= tr.from_ts
    AND apt.dibuat_pada < tr.to_ts
  GROUP BY 1
),
retail_commission AS (
  SELECT
    date_trunc('day', rcl.created_at)::date AS day,
    COALESCE(SUM(rcl.amount), 0)::bigint AS commission_amount
  FROM public.retail_commission_ledger rcl
  CROSS JOIN today_requested tr
  WHERE rcl.created_at >= tr.from_ts
    AND rcl.created_at < tr.to_ts
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
  CROSS JOIN today_requested tr
  WHERE lower(COALESCE(tm.status, '')) = 'success'
    AND tm.dibuat_pada >= tr.from_ts
    AND tm.dibuat_pada < tr.to_ts
  GROUP BY 1
),
h2h_provider AS (
  SELECT
    date_trunc('day', tp.dibuat_pada)::date AS day,
    COALESCE(SUM(CASE WHEN COALESCE(tp.harga, 0) > 0 THEN tp.harga ELSE 0 END), 0)::bigint AS provider_payment_amount
  FROM public.transaksi_provider tp
  CROSS JOIN today_requested tr
  WHERE lower(COALESCE(tp.status, '')) = 'success'
    AND tp.dibuat_pada >= tr.from_ts
    AND tp.dibuat_pada < tr.to_ts
  GROUP BY 1
),
h2h_commission AS (
  SELECT
    date_trunc('day', hcl.created_at)::date AS day,
    COALESCE(SUM(hcl.amount), 0)::bigint AS commission_amount
  FROM public.h2h_commission_ledger hcl
  CROSS JOIN today_requested tr
  WHERE hcl.created_at >= tr.from_ts
    AND hcl.created_at < tr.to_ts
  GROUP BY 1
),
h2h_transaction_expense AS (
  SELECT
    date_trunc('day', COALESCE(mb.waktu_mutasi_bank, mb.dibuat_pada) AT TIME ZONE 'Asia/Jakarta')::date AS day,
    COALESCE(SUM(CASE WHEN mb.arah = 'DEBIT' THEN mb.jumlah ELSE 0 END), 0)::bigint AS transaction_expense_amount
  FROM public.mutasi_bank mb
  CROSS JOIN today_requested tr
  WHERE mb.arah = 'DEBIT'
    AND mb.alasan = 'TRANSAKSI_SUSPECT_SELESAI'
    AND COALESCE(mb.waktu_mutasi_bank, mb.dibuat_pada) >= tr.from_ts
    AND COALESCE(mb.waktu_mutasi_bank, mb.dibuat_pada) < tr.to_ts
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
  CROSS JOIN today_requested tr
  LEFT JOIN public.member m ON m.id = dr.member_id
  WHERE lower(COALESCE(dr.status, '')) = 'approved'
    AND dr.dibuat_pada >= tr.from_ts
    AND dr.dibuat_pada < tr.to_ts
  GROUP BY 1, 2
),
provider_deposit_rows AS (
  SELECT
    date_trunc('day', COALESCE(mb.waktu_mutasi_bank, mb.dibuat_pada) AT TIME ZONE 'Asia/Jakarta')::date AS day,
    COALESCE(SUM(mdp.jumlah), 0)::bigint AS provider_deposit_amount
  FROM public.mutasi_dompet_provider mdp
  CROSS JOIN today_requested tr
  JOIN public.mutasi_bank mb
    ON mb.ref_id = mdp.ref_id
   AND mb.jumlah = mdp.jumlah
   AND lower(trim(COALESCE(mb.provider, ''))) = lower(trim(COALESCE(mdp.provider, '')))
   AND mb.arah = 'DEBIT'
   AND mb.alasan = 'BANK_TRANSFER_TO_PROVIDER'
  WHERE lower(COALESCE(mdp.arah, '')) = 'credit'
    AND COALESCE(mdp.alasan, '') = 'BANK_TRANSFER_IN'
    AND COALESCE(mb.waktu_mutasi_bank, mb.dibuat_pada) >= tr.from_ts
    AND COALESCE(mb.waktu_mutasi_bank, mb.dibuat_pada) < tr.to_ts
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
),
today_rows AS (
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
    (sr.margin_amount - sr.commission_amount - sr.transaction_expense_amount)::bigint AS profit_amount
  FROM scoped_rows sr
  LEFT JOIN member_deposits md ON md.scope = sr.scope AND md.day = sr.day
  LEFT JOIN provider_deposit_rows pd ON pd.day = sr.day
  WHERE sr.scope = CASE WHEN $1 = '' OR $1 = 'all' THEN 'all' ELSE $1 END
)
SELECT
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
  profit_amount
FROM cached_rows
UNION ALL
SELECT
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
  profit_amount
FROM today_rows
ORDER BY day DESC
`, scope, months, in.HasFrom, in.From, in.HasTo, in.To)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]AdminDailyBusinessRow, 0, 120)
	for rows.Next() {
		var item AdminDailyBusinessRow
		if err := rows.Scan(
			&item.Scope,
			&item.Day,
			&item.MonthKey,
			&item.TransactionCount,
			&item.TransactionAmount,
			&item.ProviderPaymentAmount,
			&item.MarginAmount,
			&item.CommissionAmount,
			&item.TransactionExpense,
			&item.MemberDepositAmount,
			&item.ProviderDepositAmount,
			&item.DepositGapAmount,
			&item.ProfitAmount,
		); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *AdminBusinessReportRepository) RefreshDailyBusinessCache(ctx context.Context, days int) (*AdminDailyBusinessCacheRefreshResult, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("db not initialized")
	}
	if err := r.ensureDailyBusinessSupport(ctx); err != nil {
		return nil, err
	}
	if days <= 0 || days > 366 {
		days = 93
	}

	out := &AdminDailyBusinessCacheRefreshResult{Days: days}
	err := r.db.QueryRowContext(ctx, `
SELECT
  public.refresh_admin_daily_business_cache($1, '0 seconds'::interval)::bigint AS refreshed_rows,
  now() AS refreshed_at
`, days).Scan(&out.RefreshedRows, &out.RefreshedAt)
	if err != nil {
		if isMissingDailyBusinessRefreshFunction(err) {
			out.RefreshedRows = 0
			out.RefreshedAt = time.Now()
			return out, nil
		}
		return nil, err
	}
	return out, nil
}

func isMissingDailyBusinessRefreshFunction(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "refresh_admin_daily_business_cache") &&
		(strings.Contains(msg, "does not exist") || strings.Contains(msg, "undefined_function"))
}

func (r *AdminBusinessReportRepository) ensureDailyBusinessSupport(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS public.retail_commission_ledger (
  id BIGSERIAL PRIMARY KEY,
  member_id BIGINT,
  source_member_id BIGINT,
  source_app_order_id BIGINT,
  level TEXT NOT NULL DEFAULT '',
  amount BIGINT NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS public.h2h_commission_ledger (
  id BIGSERIAL PRIMARY KEY,
  member_id BIGINT,
  source_member_id BIGINT,
  source_transaksi_member_id BIGINT,
  level TEXT NOT NULL DEFAULT '',
  amount BIGINT NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS public.admin_daily_business_cache (
  scope TEXT NOT NULL,
  day DATE NOT NULL,
  month_key TEXT NOT NULL,
  transaction_count BIGINT NOT NULL DEFAULT 0,
  transaction_amount BIGINT NOT NULL DEFAULT 0,
  provider_payment_amount BIGINT NOT NULL DEFAULT 0,
  margin_amount BIGINT NOT NULL DEFAULT 0,
  commission_amount BIGINT NOT NULL DEFAULT 0,
  transaction_expense_amount BIGINT NOT NULL DEFAULT 0,
  member_deposit_amount BIGINT NOT NULL DEFAULT 0,
  provider_deposit_amount BIGINT NOT NULL DEFAULT 0,
  deposit_gap_amount BIGINT NOT NULL DEFAULT 0,
  profit_amount BIGINT NOT NULL DEFAULT 0,
  refreshed_at TIMESTAMP NOT NULL DEFAULT now(),
  PRIMARY KEY (scope, day)
);
CREATE INDEX IF NOT EXISTS admin_daily_business_cache_day_scope_idx
  ON public.admin_daily_business_cache (day DESC, scope);
`)
	return err
}
