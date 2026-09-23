package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

func (r *AppOrderRepository) UpsertGuestRefundTicket(ctx context.Context, order *AppOrderRow, reason string) error {
	if order == nil {
		return errors.New("order not found")
	}
	if strings.TrimSpace(strings.ToLower(order.BuyerType)) != "guest" {
		return nil
	}
	if order.HargaFinal <= 0 {
		return errors.New("amount refund tidak valid")
	}

	_, err := r.db.ExecContext(ctx, `
INSERT INTO public.app_order_guest_refund
  (app_order_id, invoice_id, guest_nama, guest_email, guest_phone, amount_refund, status, reason, dibuat_pada, diubah_pada)
VALUES
  ($1,$2,$3,$4,$5,$6,'pending',NULLIF($7,''),now(),now())
ON CONFLICT (app_order_id) DO UPDATE
SET
  reason = COALESCE(public.app_order_guest_refund.reason, EXCLUDED.reason),
  diubah_pada = now()
`, order.ID, order.InvoiceID, order.GuestNama, order.GuestEmail, order.GuestPhone, order.HargaFinal, strings.TrimSpace(reason))
	return err
}

func (r *AppOrderRepository) ClaimGuestRefundToMember(ctx context.Context, memberID int64, invoiceID, guestEmail, guestPhone string) (*AppOrderGuestRefundRow, error) {
	invoiceID = strings.TrimSpace(invoiceID)
	guestEmail = strings.TrimSpace(strings.ToLower(guestEmail))
	guestPhone = normalizePhone(guestPhone)
	if memberID <= 0 {
		return nil, errors.New("member_id tidak valid")
	}
	if invoiceID == "" {
		return nil, errors.New("invoice_id wajib diisi")
	}
	if guestEmail == "" || guestPhone == "" {
		return nil, errors.New("guest_email dan guest_phone wajib diisi")
	}

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	var (
		refundID          int64
		appOrderID        int64
		dbGuestNama       sql.NullString
		dbGuestEmail      sql.NullString
		dbGuestPhone      sql.NullString
		amountRefund      int64
		status            string
		dbClaimedMemberID sql.NullInt64
		reason            sql.NullString
		notes             sql.NullString
		claimedAt         sql.NullTime
		processedAt       sql.NullTime
		dibuatPada        sql.NullTime
		diubahPada        sql.NullTime
	)
	if err := tx.QueryRowContext(ctx, `
SELECT
  id, app_order_id, guest_nama, guest_email, guest_phone, amount_refund, status, claimed_member_id,
  reason, notes, claimed_at, processed_at, dibuat_pada, diubah_pada
FROM public.app_order_guest_refund
WHERE invoice_id = $1
FOR UPDATE
`, invoiceID).Scan(
		&refundID, &appOrderID, &dbGuestNama, &dbGuestEmail, &dbGuestPhone, &amountRefund, &status, &dbClaimedMemberID,
		&reason, &notes, &claimedAt, &processedAt, &dibuatPada, &diubahPada,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("ticket refund tidak ditemukan untuk invoice ini")
		}
		return nil, err
	}

	if !dbGuestEmail.Valid || strings.TrimSpace(strings.ToLower(dbGuestEmail.String)) != guestEmail {
		return nil, errors.New("guest_email tidak cocok")
	}
	if !dbGuestPhone.Valid || normalizePhone(dbGuestPhone.String) != guestPhone {
		return nil, errors.New("guest_phone tidak cocok")
	}

	switch strings.TrimSpace(strings.ToLower(status)) {
	case "claimed_to_wallet", "paid_manual", "rejected":
		if dbClaimedMemberID.Valid && dbClaimedMemberID.Int64 == memberID {
			row := mapGuestRefundRow(refundID, appOrderID, invoiceID, dbGuestNama, dbGuestEmail, dbGuestPhone, amountRefund, status, reason, notes, dbClaimedMemberID, claimedAt, processedAt, dibuatPada, diubahPada)
			if err := tx.Commit(); err != nil {
				return nil, err
			}
			committed = true
			return row, nil
		}
		return nil, errors.New("refund sudah diproses")
	case "pending":
	default:
		return nil, errors.New("status refund tidak valid")
	}

	if amountRefund <= 0 {
		return nil, errors.New("amount_refund tidak valid")
	}

	var before int64
	if err := tx.QueryRowContext(ctx, `
SELECT saldo
FROM public.dompet_member
WHERE member_id = $1
FOR UPDATE
`, memberID).Scan(&before); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("dompet member tidak ditemukan")
		}
		return nil, err
	}
	after := before + amountRefund
	if _, err := tx.ExecContext(ctx, `
UPDATE public.dompet_member
SET saldo = $2, diperbarui_pada = now()
WHERE member_id = $1
`, memberID, after); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO public.mutasi_dompet
  (member_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, dibuat_pada)
VALUES
  ($1,$2,'CREDIT',$3,'GUEST_REFUND_CLAIM','klaim refund guest ke saldo user',$4,$5,now())
`, memberID, invoiceID, amountRefund, before, after); err != nil {
		return nil, err
	}

	now := time.Now()
	if _, err := tx.ExecContext(ctx, `
UPDATE public.app_order_guest_refund
SET
  status = 'claimed_to_wallet',
  claimed_member_id = $2,
  claimed_at = $3,
  processed_at = $3,
  diubah_pada = now()
WHERE id = $1
`, refundID, memberID, now); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `
UPDATE public.app_order
SET status = 'refunded', diubah_pada = now()
WHERE id = $1
`, appOrderID); err != nil {
		return nil, err
	}

	row := &AppOrderGuestRefundRow{
		ID:           refundID,
		AppOrderID:   appOrderID,
		InvoiceID:    invoiceID,
		AmountRefund: amountRefund,
		Status:       "claimed_to_wallet",
		ClaimedAt:    &now,
		ProcessedAt:  &now,
	}
	if dbGuestNama.Valid {
		v := dbGuestNama.String
		row.GuestNama = &v
	}
	if dbGuestEmail.Valid {
		v := dbGuestEmail.String
		row.GuestEmail = &v
	}
	if dbGuestPhone.Valid {
		v := dbGuestPhone.String
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
	claimedID := memberID
	row.ClaimedMemberID = &claimedID
	if dibuatPada.Valid {
		v := dibuatPada.Time
		row.DibuatPada = &v
	}
	row.DiubahPada = &now

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	committed = true
	return row, nil
}
