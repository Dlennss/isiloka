package repository

import (
	"context"
	"database/sql"
	"strings"
)

func (r *H2HRepository) CountSuccessfulTransactionsByMember(ctx context.Context, memberID int64) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM public.transaksi_member
WHERE member_id = $1
  AND lower(COALESCE(status, '')) = 'success'
`, memberID).Scan(&count)
	return count, err
}

func (r *H2HRepository) ListCommissionLedger(ctx context.Context, memberID int64, limit, offset int) ([]H2HCommissionLedgerRow, error) {
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
  hcl.id, hcl.member_id, hcl.source_member_id, sm.nama, sm.role, hcl.source_trx_member_id,
  hcl.ref_id, hcl.level_name, hcl.amount, hcl.kategori_name, COALESCE(hcl.note, ''), hcl.created_at
FROM public.h2h_commission_ledger hcl
LEFT JOIN public.member sm ON sm.id = hcl.source_member_id
WHERE hcl.member_id = $1
ORDER BY hcl.created_at DESC, hcl.id DESC
LIMIT $2 OFFSET $3
`, memberID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]H2HCommissionLedgerRow, 0, limit)
	for rows.Next() {
		var (
			item             H2HCommissionLedgerRow
			sourceMemberID   sql.NullInt64
			sourceMemberNama sql.NullString
			sourceMemberRole sql.NullString
			createdAt        sql.NullTime
		)
		if err := rows.Scan(
			&item.ID, &item.MemberID, &sourceMemberID, &sourceMemberNama, &sourceMemberRole,
			&item.SourceTrxMemberID, &item.RefID, &item.Level, &item.Amount, &item.Kategori, &item.Note, &createdAt,
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

func (r *H2HRepository) GetCommissionSummary(ctx context.Context, memberID int64) (*H2HCommissionSummaryRow, error) {
	out := &H2HCommissionSummaryRow{}
	if err := r.db.QueryRowContext(ctx, `
SELECT
  COALESCE((SELECT SUM(amount) FROM public.h2h_commission_ledger WHERE member_id = $1), 0),
  COALESCE((SELECT SUM(amount) FROM public.h2h_withdraw_request WHERE member_id = $1 AND status = 'pending'), 0),
  COALESCE((SELECT SUM(amount) FROM public.h2h_withdraw_request WHERE member_id = $1 AND status = 'approved'), 0),
  COALESCE((SELECT SUM(amount) FROM public.h2h_withdraw_request WHERE member_id = $1 AND status = 'rejected'), 0),
  COALESCE((SELECT saldo FROM public.dompet_member WHERE member_id = $1), 0)
`, memberID).Scan(
		&out.TotalEarned, &out.TotalPendingWithdraw, &out.TotalApprovedWithdraw, &out.TotalRejectedWithdraw, &out.AvailableSaldo,
	); err != nil {
		return nil, err
	}
	return out, nil
}
