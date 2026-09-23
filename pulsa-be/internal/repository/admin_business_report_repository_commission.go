package repository

import (
	"context"
	"database/sql"
	"errors"
)

func (r *AdminBusinessReportRepository) ListCommissionBySource(ctx context.Context, in AdminCommissionBySourceArgs) ([]AdminCommissionBySourceRow, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("db not initialized")
	}

	scope := normalizeAdminReportScope(in.Scope)
	limit := in.Limit
	if limit <= 0 || limit > 5000 {
		limit = 1000
	}
	offset := in.Offset
	if offset < 0 {
		offset = 0
	}

	rows, err := r.db.QueryContext(ctx, `
WITH all_rows AS (
  SELECT
    'retail'::text AS scope,
    rcl.member_id AS upline_member_id,
    COALESCE(um.email, '') AS upline_email,
    COALESCE(um.nama, '') AS upline_nama,
    COALESCE(um.role, '') AS upline_role,
    COALESCE(rcl.level_name, '') AS level,
    COALESCE(rcl.source_member_id, 0) AS source_member_id,
    COALESCE(sm.email, '') AS source_email,
    COALESCE(sm.nama, '') AS source_nama,
    COALESCE(sm.role, '') AS source_role,
    COUNT(*)::bigint AS transaction_count,
    COALESCE(SUM(rcl.amount), 0)::bigint AS total_commission,
    MAX(rcl.created_at) AS last_created_at
  FROM public.retail_commission_ledger rcl
  JOIN public.member um ON um.id = rcl.member_id
  LEFT JOIN public.member sm ON sm.id = rcl.source_member_id
  WHERE ($4 = false OR rcl.created_at >= $5)
    AND ($6 = false OR rcl.created_at < $7)
  GROUP BY
    rcl.member_id, um.email, um.nama, um.role,
    rcl.level_name,
    rcl.source_member_id, sm.email, sm.nama, sm.role

  UNION ALL

  SELECT
    'h2h'::text AS scope,
    hcl.member_id AS upline_member_id,
    COALESCE(um.email, '') AS upline_email,
    COALESCE(um.nama, '') AS upline_nama,
    COALESCE(um.role, '') AS upline_role,
    COALESCE(hcl.level, '') AS level,
    COALESCE(hcl.source_member_id, 0) AS source_member_id,
    COALESCE(sm.email, '') AS source_email,
    COALESCE(sm.nama, '') AS source_nama,
    COALESCE(sm.role, '') AS source_role,
    COUNT(*)::bigint AS transaction_count,
    COALESCE(SUM(hcl.amount), 0)::bigint AS total_commission,
    MAX(hcl.created_at) AS last_created_at
  FROM public.h2h_commission_ledger hcl
  JOIN public.member um ON um.id = hcl.member_id
  LEFT JOIN public.member sm ON sm.id = hcl.source_member_id
  WHERE ($4 = false OR hcl.created_at >= $5)
    AND ($6 = false OR hcl.created_at < $7)
  GROUP BY
    hcl.member_id, um.email, um.nama, um.role,
    hcl.level,
    hcl.source_member_id, sm.email, sm.nama, sm.role
)
SELECT
  scope,
  upline_member_id,
  upline_email,
  upline_nama,
  upline_role,
  level,
  source_member_id,
  source_email,
  source_nama,
  source_role,
  transaction_count,
  total_commission,
  last_created_at
FROM all_rows
WHERE ($1 = '' OR $1 = 'all' OR scope = $1)
ORDER BY total_commission DESC, transaction_count DESC, last_created_at DESC NULLS LAST
LIMIT $2 OFFSET $3
`, scope, limit, offset, in.HasFrom, in.From, in.HasTo, in.To)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]AdminCommissionBySourceRow, 0, limit)
	for rows.Next() {
		var item AdminCommissionBySourceRow
		var lastCreatedAt sql.NullTime
		if err := rows.Scan(
			&item.Scope,
			&item.UplineMemberID,
			&item.UplineEmail,
			&item.UplineNama,
			&item.UplineRole,
			&item.Level,
			&item.SourceMemberID,
			&item.SourceEmail,
			&item.SourceNama,
			&item.SourceRole,
			&item.TransactionCount,
			&item.TotalCommission,
			&lastCreatedAt,
		); err != nil {
			return nil, err
		}
		if lastCreatedAt.Valid {
			v := lastCreatedAt.Time
			item.LastCreatedAt = &v
		}
		out = append(out, item)
	}
	return out, rows.Err()
}
