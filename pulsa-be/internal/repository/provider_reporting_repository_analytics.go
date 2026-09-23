package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type ProviderAnalyticsArgs struct {
	From    time.Time
	HasFrom bool
	To      time.Time
	HasTo   bool
}

type ProviderAnalyticsRecord struct {
	Provider       string `json:"provider"`
	Total          int64  `json:"total"`
	Success        int64  `json:"success"`
	Failed         int64  `json:"failed"`
	SumQty         int64  `json:"sum_qty"`
	SumHarga       int64  `json:"sum_harga"`
	SuccessNominal int64  `json:"success_nominal"`
	DepositAmount  int64  `json:"deposit_amount"`
}

type ProviderAnalyticsPeriodRecord struct {
	Provider       string    `json:"provider"`
	PeriodStart    time.Time `json:"period_start"`
	SuccessCount   int64     `json:"success_count"`
	SuccessNominal int64     `json:"success_nominal"`
	DepositAmount  int64     `json:"deposit_amount"`
}

type ProviderAnalyticsBundle struct {
	Items        []ProviderAnalyticsRecord       `json:"items"`
	DailyItems   []ProviderAnalyticsPeriodRecord `json:"daily_items"`
	MonthlyItems []ProviderAnalyticsPeriodRecord `json:"monthly_items"`
}

type ProviderAnalyticsCacheRefreshResult struct {
	Days          int       `json:"days"`
	RefreshedRows int64     `json:"refreshed_rows"`
	RefreshedAt   time.Time `json:"refreshed_at"`
}

func (r *ProviderReportingRepository) Analytics(ctx context.Context, a ProviderAnalyticsArgs) (ProviderAnalyticsBundle, error) {
	if r == nil || r.db == nil {
		return ProviderAnalyticsBundle{}, errors.New("db not initialized")
	}

	if !a.HasTo {
		a.To = time.Now().In(time.UTC).Add(24 * time.Hour)
		a.HasTo = true
	}
	if !a.HasFrom {
		start := time.Date(a.To.Year(), a.To.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, -2, 0)
		a.From = start
		a.HasFrom = true
	}

	const successExpr = `(lower(trim(coalesce(t.status,''))) = 'success')`
	const providerExpr = `coalesce(nullif(trim(t.provider), ''), '-')`
	const depositProviderExpr = `coalesce(nullif(trim(mdp.provider), ''), '-')`

	const q = `
WITH bounds AS (
  SELECT
    $1::timestamptz AS from_ts,
    $2::timestamptz AS to_ts,
    (now() AT TIME ZONE 'Asia/Jakarta')::date AS today_date
),
cached_daily AS (
  SELECT
    c.provider,
    c.day,
    c.total,
    c.success,
    c.failed,
    c.sum_qty,
    c.sum_harga,
    c.success_nominal,
    c.deposit_amount
  FROM public.admin_provider_analytics_daily_cache c
  CROSS JOIN bounds b
  WHERE c.day >= b.from_ts::date
    AND c.day < LEAST(b.to_ts::date, b.today_date)
),
today_requested AS (
  SELECT
    GREATEST(b.from_ts, b.today_date::timestamp) AS from_ts,
    b.to_ts AS to_ts
  FROM bounds b
  WHERE b.to_ts > b.today_date::timestamp
    AND b.from_ts < b.to_ts
),
today_trx AS (
  SELECT
    ` + providerExpr + ` AS provider,
    date_trunc('day', t.dibuat_pada)::date AS day,
    count(*)::bigint AS total,
    count(*) FILTER (WHERE ` + successExpr + `)::bigint AS success,
    count(*) FILTER (WHERE NOT ` + successExpr + `)::bigint AS failed,
    coalesce(sum(t.qty),0)::bigint AS sum_qty,
    coalesce(sum(coalesce(t.harga,0)),0)::bigint AS sum_harga,
    coalesce(sum(CASE WHEN ` + successExpr + ` THEN coalesce(t.harga,0) ELSE 0 END),0)::bigint AS success_nominal
  FROM public.transaksi_provider t
  CROSS JOIN today_requested tr
  WHERE t.dibuat_pada >= tr.from_ts
    AND t.dibuat_pada <  tr.to_ts
  GROUP BY 1, 2
),
today_dep AS (
  SELECT
    ` + depositProviderExpr + ` AS provider,
    date_trunc('day', mdp.dibuat_pada)::date AS day,
    coalesce(sum(mdp.jumlah),0)::bigint AS deposit_amount
  FROM public.mutasi_dompet_provider mdp
  CROSS JOIN today_requested tr
  WHERE mdp.dibuat_pada >= tr.from_ts
    AND mdp.dibuat_pada <  tr.to_ts
    AND mdp.arah = 'credit'
    AND (
      mdp.alasan = 'PROVIDER_DEPOSIT' OR
      mdp.alasan = 'BANK_TRANSFER_IN' OR
      mdp.alasan = 'PROVIDER_ADJUST_CREDIT' OR
      mdp.alasan = 'PROVIDER_SET_BALANCE'
    )
  GROUP BY 1, 2
),
today_days AS (
  SELECT provider, day FROM today_trx
  UNION
  SELECT provider, day FROM today_dep
),
today_daily AS (
  SELECT
    d.provider,
    d.day,
    coalesce(t.total,0)::bigint AS total,
    coalesce(t.success,0)::bigint AS success,
    coalesce(t.failed,0)::bigint AS failed,
    coalesce(t.sum_qty,0)::bigint AS sum_qty,
    coalesce(t.sum_harga,0)::bigint AS sum_harga,
    coalesce(t.success_nominal,0)::bigint AS success_nominal,
    coalesce(dep.deposit_amount,0)::bigint AS deposit_amount
  FROM today_days d
  LEFT JOIN today_trx t ON t.provider = d.provider AND t.day = d.day
  LEFT JOIN today_dep dep ON dep.provider = d.provider AND dep.day = d.day
),
all_daily AS MATERIALIZED (
  SELECT * FROM cached_daily
  UNION ALL
  SELECT * FROM today_daily
)
SELECT
  row_kind,
  provider,
  period_start,
  total,
  success,
  failed,
  sum_qty,
  sum_harga,
  success_nominal,
  deposit_amount
FROM (
  SELECT
    'summary'::text AS row_kind,
    provider,
    NULL::date AS period_start,
    sum(total)::bigint AS total,
    sum(success)::bigint AS success,
    sum(failed)::bigint AS failed,
    sum(sum_qty)::bigint AS sum_qty,
    sum(sum_harga)::bigint AS sum_harga,
    sum(success_nominal)::bigint AS success_nominal,
    sum(deposit_amount)::bigint AS deposit_amount
  FROM all_daily
  GROUP BY provider
  UNION ALL
  SELECT
    'daily'::text AS row_kind,
    provider,
    day AS period_start,
    0::bigint AS total,
    sum(success)::bigint AS success,
    0::bigint AS failed,
    0::bigint AS sum_qty,
    0::bigint AS sum_harga,
    sum(success_nominal)::bigint AS success_nominal,
    sum(deposit_amount)::bigint AS deposit_amount
  FROM all_daily
  GROUP BY provider, day
  UNION ALL
  SELECT
    'monthly'::text AS row_kind,
    provider,
    date_trunc('month', day::timestamp)::date AS period_start,
    0::bigint AS total,
    sum(success)::bigint AS success,
    0::bigint AS failed,
    0::bigint AS sum_qty,
    0::bigint AS sum_harga,
    sum(success_nominal)::bigint AS success_nominal,
    sum(deposit_amount)::bigint AS deposit_amount
  FROM all_daily
  GROUP BY provider, date_trunc('month', day::timestamp)::date
) x
ORDER BY
  CASE row_kind WHEN 'summary' THEN 0 WHEN 'daily' THEN 1 ELSE 2 END,
  CASE WHEN row_kind = 'summary' THEN success_nominal ELSE NULL END DESC NULLS LAST,
  period_start DESC NULLS LAST,
  provider ASC
`
	rows, err := r.db.QueryContext(ctx, q, a.From, a.To)
	if err != nil {
		return ProviderAnalyticsBundle{}, err
	}
	defer rows.Close()

	items := make([]ProviderAnalyticsRecord, 0, 16)
	dailyItems := make([]ProviderAnalyticsPeriodRecord, 0, 128)
	monthlyItems := make([]ProviderAnalyticsPeriodRecord, 0, 32)
	for rows.Next() {
		var rowKind string
		var periodStart sql.NullTime
		var provider string
		var total, success, failed, sumQty, sumHarga, successNominal, depositAmount int64
		if err := rows.Scan(
			&rowKind,
			&provider,
			&periodStart,
			&total,
			&success,
			&failed,
			&sumQty,
			&sumHarga,
			&successNominal,
			&depositAmount,
		); err != nil {
			return ProviderAnalyticsBundle{}, err
		}

		switch rowKind {
		case "summary":
			items = append(items, ProviderAnalyticsRecord{
				Provider:       provider,
				Total:          total,
				Success:        success,
				Failed:         failed,
				SumQty:         sumQty,
				SumHarga:       sumHarga,
				SuccessNominal: successNominal,
				DepositAmount:  depositAmount,
			})
		case "daily":
			if periodStart.Valid {
				dailyItems = append(dailyItems, ProviderAnalyticsPeriodRecord{
					Provider:       provider,
					PeriodStart:    periodStart.Time,
					SuccessCount:   success,
					SuccessNominal: successNominal,
					DepositAmount:  depositAmount,
				})
			}
		case "monthly":
			if periodStart.Valid {
				monthlyItems = append(monthlyItems, ProviderAnalyticsPeriodRecord{
					Provider:       provider,
					PeriodStart:    periodStart.Time,
					SuccessCount:   success,
					SuccessNominal: successNominal,
					DepositAmount:  depositAmount,
				})
			}
		}
	}
	if err := rows.Err(); err != nil {
		return ProviderAnalyticsBundle{}, err
	}

	return ProviderAnalyticsBundle{
		Items:        items,
		DailyItems:   dailyItems,
		MonthlyItems: monthlyItems,
	}, nil
}

func (r *ProviderReportingRepository) RefreshAnalyticsCache(ctx context.Context, days int) (*ProviderAnalyticsCacheRefreshResult, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("db not initialized")
	}
	if days <= 0 || days > 366 {
		days = 93
	}

	out := &ProviderAnalyticsCacheRefreshResult{Days: days}
	err := r.db.QueryRowContext(ctx, `
SELECT
  public.refresh_admin_provider_analytics_cache($1, '0 seconds'::interval, false)::bigint AS refreshed_rows,
  now() AS refreshed_at
`, days).Scan(&out.RefreshedRows, &out.RefreshedAt)
	if err != nil {
		return nil, err
	}
	return out, nil
}
