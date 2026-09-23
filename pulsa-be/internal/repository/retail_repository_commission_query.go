package repository

import (
	"context"
	"database/sql"
	"strings"
)

func (r *RetailRepository) CountSuccessfulOrdersByMember(ctx context.Context, memberID int64) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM public.app_order
WHERE member_id = $1
  AND lower(COALESCE(buyer_type, '')) = 'user'
  AND lower(COALESCE(status, '')) = 'success'
`, memberID).Scan(&count)
	return count, err
}

func (r *RetailRepository) ListCommissionLedger(ctx context.Context, memberID int64, limit, offset int) ([]RetailCommissionLedgerRow, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := r.db.QueryContext(ctx, `
SELECT
  rcl.id, rcl.member_id, rcl.source_member_id, sm.nama, sm.role, rcl.source_app_order_id,
  rcl.invoice_id, rcl.level_name, rcl.amount, COALESCE(rcl.note, ''), rcl.created_at
FROM public.retail_commission_ledger rcl
LEFT JOIN public.member sm ON sm.id = rcl.source_member_id
WHERE rcl.member_id = $1
ORDER BY rcl.created_at DESC, rcl.id DESC
LIMIT $2 OFFSET $3
`, memberID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]RetailCommissionLedgerRow, 0, limit)
	for rows.Next() {
		var (
			item             RetailCommissionLedgerRow
			sourceMemberID   sql.NullInt64
			sourceMemberNama sql.NullString
			sourceMemberRole sql.NullString
			createdAt        sql.NullTime
		)
		if err := rows.Scan(
			&item.ID, &item.MemberID, &sourceMemberID, &sourceMemberNama, &sourceMemberRole,
			&item.SourceAppOrderID, &item.InvoiceID, &item.Level, &item.Amount, &item.Note, &createdAt,
		); err != nil {
			return nil, err
		}
		if sourceMemberID.Valid {
			v := sourceMemberID.Int64
			item.SourceMemberID = &v
		}
		if sourceMemberNama.Valid && strings.TrimSpace(sourceMemberNama.String) != "" {
			v := sourceMemberNama.String
			item.SourceMemberNama = &v
		}
		if sourceMemberRole.Valid && strings.TrimSpace(sourceMemberRole.String) != "" {
			v := sourceMemberRole.String
			item.SourceMemberRole = &v
		}
		if createdAt.Valid {
			v := createdAt.Time
			item.CreatedAt = &v
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *RetailRepository) GetCommissionSummary(ctx context.Context, memberID int64) (*RetailCommissionSummaryRow, error) {
	out := &RetailCommissionSummaryRow{}
	if err := r.db.QueryRowContext(ctx, `
SELECT
  COALESCE((SELECT SUM(amount) FROM public.retail_commission_ledger WHERE member_id = $1), 0),
  COALESCE((SELECT SUM(amount) FROM public.retail_withdraw_request WHERE member_id = $1 AND status = 'pending'), 0),
  COALESCE((SELECT SUM(amount) FROM public.retail_withdraw_request WHERE member_id = $1 AND status = 'approved'), 0),
  COALESCE((SELECT SUM(amount) FROM public.retail_withdraw_request WHERE member_id = $1 AND status = 'rejected'), 0),
  COALESCE((SELECT saldo FROM public.dompet_member WHERE member_id = $1), 0)
`, memberID).Scan(
		&out.TotalEarned, &out.TotalPendingWithdraw, &out.TotalApprovedWithdraw, &out.TotalRejectedWithdraw, &out.AvailableSaldo,
	); err != nil {
		return nil, err
	}
	return out, nil
}
