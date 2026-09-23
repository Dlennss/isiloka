package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

type DailyProductSuccessArgs struct {
	Q string

	From    time.Time
	HasFrom bool

	To    time.Time
	HasTo bool

	Limit  int
	Offset int
}

type DailyProductSuccessRecord struct {
	PeriodStart        time.Time `json:"period_start"`
	InternalSKU        string    `json:"internal_sku"`
	ProductName        string    `json:"product_name"`
	GroupName          string    `json:"group_name"`
	SuccessCount       int64     `json:"success_count"`
	TotalQty           int64     `json:"total_qty"`
	TotalQtyProvider   int64     `json:"total_qty_provider"`
	TotalProviderPrice int64     `json:"total_provider_price"`
	TotalMemberPrice   int64     `json:"total_member_price"`
	TotalMargin        int64     `json:"total_margin"`
	ProviderCount      int64     `json:"provider_count"`
	Providers          string    `json:"providers"`
	FirstSuccessAt     time.Time `json:"first_success_at"`
	LastSuccessAt      time.Time `json:"last_success_at"`
}

type DailyProductSuccessSummary struct {
	GroupCount         int64 `json:"group_count"`
	UniqueSKUCount     int64 `json:"unique_sku_count"`
	SuccessCount       int64 `json:"success_count"`
	TotalQty           int64 `json:"total_qty"`
	TotalQtyProvider   int64 `json:"total_qty_provider"`
	TotalProviderPrice int64 `json:"total_provider_price"`
	TotalMemberPrice   int64 `json:"total_member_price"`
	TotalMargin        int64 `json:"total_margin"`
}

type DailyProductSuccessBundle struct {
	Items   []DailyProductSuccessRecord `json:"items"`
	Total   int64                       `json:"total"`
	Summary DailyProductSuccessSummary  `json:"summary"`
}

func (r *ProviderReportingRepository) DailyProductSuccess(ctx context.Context, a DailyProductSuccessArgs) (DailyProductSuccessBundle, error) {
	if r == nil || r.db == nil {
		return DailyProductSuccessBundle{}, errors.New("db not initialized")
	}

	a.Q = strings.TrimSpace(a.Q)
	if a.Limit <= 0 || a.Limit > 1000 {
		a.Limit = 500
	}
	if a.Offset < 0 {
		a.Offset = 0
	}

	const baseCTE = `
WITH success_trx AS (
  SELECT
    date_trunc('day', tm.diperbarui_pada AT TIME ZONE 'Asia/Jakarta')::date AS period_start,
    upper(trim(coalesce(tm.kode_produk,''))) AS internal_sku,
    coalesce(p.nama,'') AS product_name,
    coalesce(p.group_name,'') AS group_name,
    coalesce(tp.provider,'') AS provider,
    tm.diperbarui_pada AS success_at,
    coalesce(tm.qty,0) AS qty,
    coalesce(tm.qty_provider, tm.qty, 0) AS qty_provider,
    coalesce(tp.harga,0) AS provider_price,
    coalesce(tm.harga_member,0) AS member_price
  FROM public.transaksi_member tm
  LEFT JOIN public.produk p ON upper(trim(p.sku)) = upper(trim(tm.kode_produk))
  LEFT JOIN LATERAL (
    SELECT provider, harga
    FROM public.transaksi_provider tp
    WHERE tp.transaksi_member_id = tm.id
      AND lower(trim(coalesce(tp.status,''))) = 'success'
    ORDER BY tp.id DESC
    LIMIT 1
  ) tp ON true
  WHERE lower(trim(coalesce(tm.status,''))) = 'success'
    AND ($1 = false OR tm.diperbarui_pada >= $2)
    AND ($3 = false OR tm.diperbarui_pada <  $4)
    AND (
      $5 = '' OR
      tm.kode_produk ILIKE '%'||$5||'%' OR
      coalesce(p.nama,'') ILIKE '%'||$5||'%' OR
      coalesce(p.group_name,'') ILIKE '%'||$5||'%'
    )
), grouped AS (
  SELECT
    period_start,
    internal_sku,
    product_name,
    group_name,
    count(*) AS success_count,
    coalesce(sum(qty),0) AS total_qty,
    coalesce(sum(qty_provider),0) AS total_qty_provider,
    coalesce(sum(provider_price),0) AS total_provider_price,
    coalesce(sum(member_price),0) AS total_member_price,
    coalesce(sum(member_price - provider_price),0) AS total_margin,
    count(DISTINCT nullif(provider,'')) AS provider_count,
    coalesce(string_agg(DISTINCT nullif(provider,''), ', ' ORDER BY nullif(provider,'')), '') AS providers,
    min(success_at) AS first_success_at,
    max(success_at) AS last_success_at
  FROM success_trx
  WHERE internal_sku <> ''
  GROUP BY period_start, internal_sku, product_name, group_name
)
`

	qList := baseCTE + `
SELECT period_start, internal_sku, product_name, group_name, success_count,
       total_qty, total_qty_provider, total_provider_price, total_member_price, total_margin,
       provider_count, providers, first_success_at, last_success_at
FROM grouped
ORDER BY period_start DESC, success_count DESC, total_qty DESC, internal_sku ASC
LIMIT $6 OFFSET $7
`

	rows, err := r.db.QueryContext(ctx, qList, a.HasFrom, a.From, a.HasTo, a.To, a.Q, a.Limit, a.Offset)
	if err != nil {
		return DailyProductSuccessBundle{}, err
	}
	defer rows.Close()

	items := make([]DailyProductSuccessRecord, 0, a.Limit)
	for rows.Next() {
		var x DailyProductSuccessRecord
		var productName, groupName, providers sql.NullString
		if err := rows.Scan(
			&x.PeriodStart,
			&x.InternalSKU,
			&productName,
			&groupName,
			&x.SuccessCount,
			&x.TotalQty,
			&x.TotalQtyProvider,
			&x.TotalProviderPrice,
			&x.TotalMemberPrice,
			&x.TotalMargin,
			&x.ProviderCount,
			&providers,
			&x.FirstSuccessAt,
			&x.LastSuccessAt,
		); err != nil {
			return DailyProductSuccessBundle{}, err
		}
		if productName.Valid {
			x.ProductName = productName.String
		}
		if groupName.Valid {
			x.GroupName = groupName.String
		}
		if providers.Valid {
			x.Providers = providers.String
		}
		items = append(items, x)
	}
	if err := rows.Err(); err != nil {
		return DailyProductSuccessBundle{}, err
	}

	qSummary := baseCTE + `
SELECT
  count(*) AS group_count,
  count(DISTINCT internal_sku) AS unique_sku_count,
  coalesce(sum(success_count),0) AS success_count,
  coalesce(sum(total_qty),0) AS total_qty,
  coalesce(sum(total_qty_provider),0) AS total_qty_provider,
  coalesce(sum(total_provider_price),0) AS total_provider_price,
  coalesce(sum(total_member_price),0) AS total_member_price,
  coalesce(sum(total_margin),0) AS total_margin
FROM grouped
`
	var summary DailyProductSuccessSummary
	if err := r.db.QueryRowContext(ctx, qSummary, a.HasFrom, a.From, a.HasTo, a.To, a.Q).Scan(
		&summary.GroupCount,
		&summary.UniqueSKUCount,
		&summary.SuccessCount,
		&summary.TotalQty,
		&summary.TotalQtyProvider,
		&summary.TotalProviderPrice,
		&summary.TotalMemberPrice,
		&summary.TotalMargin,
	); err != nil {
		return DailyProductSuccessBundle{}, err
	}

	return DailyProductSuccessBundle{Items: items, Total: summary.GroupCount, Summary: summary}, nil
}
