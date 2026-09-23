package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

func (r *AppOrderRepository) ListAdminGuestRefundPending(ctx context.Context, limit, offset int, invoiceID string) ([]AdminGuestRefundPendingRow, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	args := []any{"pending"}
	wheres := []string{"LOWER(TRIM(gr.status)) = $1"}

	invoiceID = strings.TrimSpace(invoiceID)
	if invoiceID != "" {
		args = append(args, strings.ToUpper(invoiceID))
		wheres = append(wheres, fmt.Sprintf("UPPER(TRIM(gr.invoice_id)) = $%d", len(args)))
	}

	whereSQL := strings.Join(wheres, " AND ")
	baseFrom := `
FROM public.app_order_guest_refund gr
LEFT JOIN public.app_order ao ON ao.id = gr.app_order_id
`

	var total int64
	qCount := fmt.Sprintf(`SELECT COUNT(*) %s WHERE %s`, baseFrom, whereSQL)
	if err := r.db.QueryRowContext(ctx, qCount, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	listArgs := append([]any{}, args...)
	listArgs = append(listArgs, limit, offset)
	limitPos := len(args) + 1
	offsetPos := len(args) + 2

	qList := fmt.Sprintf(`
SELECT
  gr.id,
  gr.app_order_id,
  gr.invoice_id,
  gr.guest_nama,
  gr.guest_email,
  gr.guest_phone,
  gr.amount_refund,
  gr.status,
  gr.reason,
  gr.claimed_member_id,
  ao.produk_nama_snapshot,
  ao.dest,
  ao.status,
  gr.dibuat_pada,
  gr.diubah_pada
%s
WHERE %s
ORDER BY gr.dibuat_pada DESC, gr.id DESC
LIMIT $%d OFFSET $%d
`, baseFrom, whereSQL, limitPos, offsetPos)

	rows, err := r.db.QueryContext(ctx, qList, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]AdminGuestRefundPendingRow, 0, limit)
	for rows.Next() {
		var (
			row             AdminGuestRefundPendingRow
			guestNama       sql.NullString
			guestEmail      sql.NullString
			guestPhone      sql.NullString
			reason          sql.NullString
			claimedMemberID sql.NullInt64
			produkNama      sql.NullString
			dest            sql.NullString
			orderStatus     sql.NullString
			dibuatPada      sql.NullTime
			diubahPada      sql.NullTime
		)
		if err := rows.Scan(
			&row.ID, &row.AppOrderID, &row.InvoiceID, &guestNama, &guestEmail, &guestPhone, &row.AmountRefund, &row.Status,
			&reason, &claimedMemberID, &produkNama, &dest, &orderStatus, &dibuatPada, &diubahPada,
		); err != nil {
			return nil, 0, err
		}
		if guestNama.Valid {
			v := strings.TrimSpace(guestNama.String)
			row.GuestNama = &v
		}
		if guestEmail.Valid {
			v := strings.TrimSpace(guestEmail.String)
			row.GuestEmail = &v
		}
		if guestPhone.Valid {
			v := strings.TrimSpace(guestPhone.String)
			row.GuestPhone = &v
		}
		if reason.Valid {
			v := strings.TrimSpace(reason.String)
			row.Reason = &v
		}
		if claimedMemberID.Valid {
			v := claimedMemberID.Int64
			row.ClaimedMemberID = &v
		}
		if produkNama.Valid {
			v := strings.TrimSpace(produkNama.String)
			row.ProdukNamaSnapshot = &v
		}
		if dest.Valid {
			v := strings.TrimSpace(dest.String)
			row.Dest = &v
		}
		if orderStatus.Valid {
			v := strings.TrimSpace(orderStatus.String)
			row.OrderStatus = &v
		}
		if dibuatPada.Valid {
			v := dibuatPada.Time
			row.DibuatPada = &v
		}
		if diubahPada.Valid {
			v := diubahPada.Time
			row.DiubahPada = &v
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func (r *AppOrderRepository) AdminClaimGuestRefundToMember(ctx context.Context, memberID int64, invoiceID string) (*AppOrderGuestRefundRow, error) {
	invoiceID = strings.TrimSpace(invoiceID)
	if memberID <= 0 {
		return nil, errors.New("member_id tidak valid")
	}
	if invoiceID == "" {
		return nil, errors.New("invoice_id wajib diisi")
	}

	var guestEmail, guestPhone sql.NullString
	if err := r.db.QueryRowContext(ctx, `
SELECT guest_email, guest_phone
FROM public.app_order_guest_refund
WHERE invoice_id = $1
LIMIT 1
`, invoiceID).Scan(&guestEmail, &guestPhone); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("ticket refund tidak ditemukan untuk invoice ini")
		}
		return nil, err
	}
	if !guestEmail.Valid || !guestPhone.Valid {
		return nil, errors.New("data guest refund tidak lengkap")
	}

	return r.ClaimGuestRefundToMember(ctx, memberID, invoiceID, guestEmail.String, guestPhone.String)
}

func mapGuestRefundRow(
	refundID int64,
	appOrderID int64,
	invoiceID string,
	guestNama sql.NullString,
	guestEmail sql.NullString,
	guestPhone sql.NullString,
	amountRefund int64,
	status string,
	reason sql.NullString,
	notes sql.NullString,
	claimedMemberID sql.NullInt64,
	claimedAt sql.NullTime,
	processedAt sql.NullTime,
	dibuatPada sql.NullTime,
	diubahPada sql.NullTime,
) *AppOrderGuestRefundRow {
	row := &AppOrderGuestRefundRow{
		ID:           refundID,
		AppOrderID:   appOrderID,
		InvoiceID:    invoiceID,
		AmountRefund: amountRefund,
		Status:       status,
	}
	if guestNama.Valid {
		v := guestNama.String
		row.GuestNama = &v
	}
	if guestEmail.Valid {
		v := guestEmail.String
		row.GuestEmail = &v
	}
	if guestPhone.Valid {
		v := guestPhone.String
		row.GuestPhone = &v
	}
	if reason.Valid {
		v := reason.String
		row.Reason = &v
	}
	if notes.Valid {
		v := notes.String
		row.Notes = &v
	}
	if claimedMemberID.Valid {
		v := claimedMemberID.Int64
		row.ClaimedMemberID = &v
	}
	if claimedAt.Valid {
		v := claimedAt.Time
		row.ClaimedAt = &v
	}
	if processedAt.Valid {
		v := processedAt.Time
		row.ProcessedAt = &v
	}
	if dibuatPada.Valid {
		v := dibuatPada.Time
		row.DibuatPada = &v
	}
	if diubahPada.Valid {
		v := diubahPada.Time
		row.DiubahPada = &v
	}
	return row
}
