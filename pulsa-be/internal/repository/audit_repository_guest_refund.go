package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

func (r *AuditRepository) AdminListGuestRefundMissing(
	ctx context.Context,
	limit, offset int,
	fromStr, toStr, invoiceID string,
) ([]AdminGuestRefundMissingRow, int64, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}

	loc, _ := time.LoadLocation("Asia/Jakarta")
	var (
		args   []any
		wheres []string
	)

	wheres = append(wheres, "LOWER(TRIM(ao.buyer_type)) = 'guest'")
	wheres = append(wheres, "LOWER(TRIM(ao.status)) = 'failed'")
	wheres = append(wheres, "ao.harga_final > 0")
	wheres = append(wheres, "gr.id IS NULL")

	fromStr = strings.TrimSpace(fromStr)
	if fromStr != "" {
		t, err := time.ParseInLocation("2006-01-02", fromStr, loc)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid from (expected YYYY-MM-DD): %w", err)
		}
		args = append(args, t)
		wheres = append(wheres, fmt.Sprintf("ao.dibuat_pada >= $%d", len(args)))
	}
	toStr = strings.TrimSpace(toStr)
	if toStr != "" {
		t, err := time.ParseInLocation("2006-01-02", toStr, loc)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid to (expected YYYY-MM-DD): %w", err)
		}
		args = append(args, t.AddDate(0, 0, 1))
		wheres = append(wheres, fmt.Sprintf("ao.dibuat_pada < $%d", len(args)))
	}
	invoiceID = strings.TrimSpace(invoiceID)
	if invoiceID != "" {
		args = append(args, strings.ToUpper(invoiceID))
		wheres = append(wheres, fmt.Sprintf("UPPER(TRIM(ao.invoice_id)) = $%d", len(args)))
	}

	whereSQL := strings.Join(wheres, " AND ")
	baseFrom := `
FROM public.app_order ao
LEFT JOIN public.app_order_guest_refund gr ON gr.app_order_id = ao.id
LEFT JOIN LATERAL (
  SELECT apt.id, apt.provider, apt.status
  FROM public.app_order_provider_trx apt
  WHERE apt.app_order_id = ao.id
  ORDER BY apt.dibuat_pada DESC, apt.id DESC
  LIMIT 1
) apt ON true
`

	listArgs := append([]any{}, args...)
	listArgs = append(listArgs, limit, offset)
	limitPos := len(args) + 1
	offsetPos := len(args) + 2

	qList := fmt.Sprintf(`
SELECT
  ao.id,
  ao.invoice_id,
  ao.buyer_type,
  ao.status,
  ao.member_id,
  ao.guest_nama,
  ao.guest_email,
  ao.guest_phone,
  ao.harga_final,
  ao.dibuat_pada,
  ao.diubah_pada,
  apt.id AS provider_trx_id,
  CASE
    WHEN apt.id IS NULL THEN NULL
    ELSE CONCAT(COALESCE(apt.provider,''), ':', COALESCE(apt.status,''))
  END AS provider_info
%s
WHERE %s
ORDER BY ao.dibuat_pada DESC, ao.id DESC
LIMIT $%d OFFSET $%d
`, baseFrom, whereSQL, limitPos, offsetPos)

	rows, err := r.db.QueryContext(ctx, qList, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]AdminGuestRefundMissingRow, 0, limit)
	for rows.Next() {
		var (
			x          AdminGuestRefundMissingRow
			memberID   sql.NullInt64
			guestNama  sql.NullString
			guestEmail sql.NullString
			guestPhone sql.NullString
			providerID sql.NullInt64
			provInfo   sql.NullString
		)
		if err := rows.Scan(
			&x.AppOrderID, &x.InvoiceID, &x.BuyerType, &x.Status, &memberID, &guestNama, &guestEmail, &guestPhone,
			&x.HargaFinal, &x.DibuatPada, &x.DiubahPada, &providerID, &provInfo,
		); err != nil {
			return nil, 0, err
		}
		if memberID.Valid {
			v := memberID.Int64
			x.MemberID = &v
		}
		if guestNama.Valid {
			v := strings.TrimSpace(guestNama.String)
			x.GuestNama = &v
		}
		if guestEmail.Valid {
			v := strings.TrimSpace(guestEmail.String)
			x.GuestEmail = &v
		}
		if guestPhone.Valid {
			v := strings.TrimSpace(guestPhone.String)
			x.GuestPhone = &v
		}
		if providerID.Valid {
			v := providerID.Int64
			x.ProviderTrx = &v
		}
		if provInfo.Valid {
			v := strings.TrimSpace(provInfo.String)
			x.ProviderInfo = &v
		}
		out = append(out, x)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	qCount := fmt.Sprintf(`SELECT COUNT(1) %s WHERE %s`, baseFrom, whereSQL)
	var total int64
	if err := r.db.QueryRowContext(ctx, qCount, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}
