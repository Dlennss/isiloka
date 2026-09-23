package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

func (r *AppOrderPaymentRepository) CreateWithWalletDebit(ctx context.Context, in AppOrderPaymentCreateInput, memberID int64, walletDebit int64, reason string, note string, setOrderPaid bool) error {
	if walletDebit < 0 {
		return errors.New("wallet_debit tidak valid")
	}
	if walletDebit > 0 && memberID <= 0 {
		return errors.New("member_id tidak valid")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if walletDebit > 0 {
		var before int64
		if err := tx.QueryRowContext(ctx, `
SELECT saldo
FROM public.dompet_member
WHERE member_id = $1
FOR UPDATE
`, memberID).Scan(&before); err != nil {
			return err
		}
		if before < walletDebit {
			return errors.New("saldo tidak cukup")
		}
		after := before - walletDebit

		if _, err := tx.ExecContext(ctx, `
UPDATE public.dompet_member
SET saldo = $2, diperbarui_pada = now()
WHERE member_id = $1
`, memberID, after); err != nil {
			return err
		}

		if _, err := tx.ExecContext(ctx, `
INSERT INTO public.mutasi_dompet
  (member_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, dibuat_pada)
VALUES
  ($1,$2,'DEBIT',$3,$4,NULLIF($5,''),$6,$7,now())
`, memberID, in.OrderID, walletDebit, reason, note, before, after); err != nil {
			return err
		}
	}

	if err := tx.QueryRowContext(ctx, `
INSERT INTO public.app_order_payment
  (app_order_id, order_id, transaction_id, gross_amount, payment_type, transaction_status, fraud_status, acquirer, qr_url, raw_request, raw_callback, paid_at, expired_at, settlement_time, dibuat_pada, diubah_pada)
VALUES
  ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10::jsonb,$11::jsonb,$12,$13,$14,now(),now())
RETURNING id
`, in.AppOrderID, in.OrderID, nullableStringValue(in.TransactionID), in.GrossAmount, nullableStringValue(in.PaymentType), nullableStringValue(in.TransactionStatus), nullableStringValue(in.FraudStatus), nullableStringValue(in.Acquirer), nullableStringValue(in.QRURL), nullableJSON(in.RawRequest), nullableJSON(in.RawCallback), in.PaidAt, in.ExpiredAt, in.SettlementTime).Scan(&in.ID); err != nil {
		return err
	}

	if setOrderPaid {
		res, err := tx.ExecContext(ctx, `
UPDATE public.app_order
SET status = 'paid', diubah_pada = now()
WHERE id = $1
`, in.AppOrderID)
		if err != nil {
			return err
		}
		aff, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if aff == 0 {
			return sql.ErrNoRows
		}
	}

	return tx.Commit()
}

func (r *AppOrderPaymentRepository) CreateWithBalanceDebit(ctx context.Context, in AppOrderPaymentCreateInput, memberID, amount int64) error {
	if memberID <= 0 || amount <= 0 {
		return errors.New("member dan nominal pembayaran tidak valid")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var walletBefore int64
	if err := tx.QueryRowContext(ctx, `SELECT saldo FROM public.dompet_member WHERE member_id = $1 FOR UPDATE`, memberID).Scan(&walletBefore); err != nil {
		return err
	}
	if walletBefore < amount {
		return errors.New("saldo utama tidak cukup")
	}
	walletDebit := amount
	if walletDebit > 0 {
		walletAfter := walletBefore - walletDebit
		if _, err := tx.ExecContext(ctx, `UPDATE public.dompet_member SET saldo=$2, diperbarui_pada=now() WHERE member_id=$1`, memberID, walletAfter); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO public.mutasi_dompet
  (member_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, dibuat_pada)
VALUES ($1,$2,'DEBIT',$3,'APP_ORDER_WALLET_DEBIT','pembayaran produk melalui Pulsa24Jam',$4,$5,now())
`, memberID, in.OrderID, walletDebit, walletBefore, walletAfter); err != nil {
			return err
		}
	}

	in.RawRequest = BuildBalancePaymentRawRequest(walletDebit, 0, amount)
	if err := tx.QueryRowContext(ctx, `
INSERT INTO public.app_order_payment
  (app_order_id, order_id, transaction_id, gross_amount, payment_type, transaction_status, fraud_status, acquirer, qr_url, raw_request, raw_callback, paid_at, expired_at, settlement_time, dibuat_pada, diubah_pada)
VALUES
  ($1,$2,NULL,0,'balance','settlement','accept','pulsa24jam',NULL,$3::jsonb,'{}'::jsonb,$4,NULL,$4,now(),now())
RETURNING id
`, in.AppOrderID, in.OrderID, in.RawRequest, in.PaidAt).Scan(&in.ID); err != nil {
		return err
	}
	res, err := tx.ExecContext(ctx, `UPDATE public.app_order SET status='paid', diubah_pada=now() WHERE id=$1 AND status='pending_payment'`, in.AppOrderID)
	if err != nil {
		return err
	}
	if affected, err := res.RowsAffected(); err != nil || affected != 1 {
		if err != nil {
			return err
		}
		return errors.New("status order sudah berubah")
	}
	return tx.Commit()
}

func (r *AppOrderPaymentRepository) RefundWalletDebitIfNeeded(ctx context.Context, orderID string, reason string, note string) (bool, error) {
	orderID = strings.TrimSpace(orderID)
	if orderID == "" {
		return false, errors.New("order_id wajib diisi")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	var (
		memberID    sql.NullInt64
		walletDebit int64
	)
	if err := tx.QueryRowContext(ctx, `
SELECT
  ao.member_id,
  COALESCE((ap.raw_request ->> 'wallet_debit')::bigint, 0) AS wallet_debit
FROM public.app_order_payment ap
JOIN public.app_order ao ON ao.id = ap.app_order_id
WHERE ap.order_id = $1
FOR UPDATE
`, orderID).Scan(&memberID, &walletDebit); err != nil {
		return false, err
	}
	if !memberID.Valid || memberID.Int64 <= 0 || walletDebit <= 0 {
		return false, nil
	}

	var exists bool
	if err := tx.QueryRowContext(ctx, `
SELECT EXISTS(
  SELECT 1
  FROM public.mutasi_dompet
  WHERE member_id = $1
    AND ref_id = $2
    AND alasan = $3
)
`, memberID.Int64, orderID, reason).Scan(&exists); err != nil {
		return false, err
	}
	if exists {
		return false, tx.Commit()
	}

	var before int64
	if err := tx.QueryRowContext(ctx, `
SELECT saldo
FROM public.dompet_member
WHERE member_id = $1
FOR UPDATE
`, memberID.Int64).Scan(&before); err != nil {
		return false, err
	}
	after := before + walletDebit

	if _, err := tx.ExecContext(ctx, `
UPDATE public.dompet_member
SET saldo = $2, diperbarui_pada = now()
WHERE member_id = $1
`, memberID.Int64, after); err != nil {
		return false, err
	}

	if _, err := tx.ExecContext(ctx, `
INSERT INTO public.mutasi_dompet
  (member_id, ref_id, arah, jumlah, alasan, catatan, saldo_sebelum, saldo_sesudah, dibuat_pada)
VALUES
  ($1,$2,'CREDIT',$3,$4,NULLIF($5,''),$6,$7,now())
`, memberID.Int64, orderID, walletDebit, reason, note, before, after); err != nil {
		return false, err
	}

	return true, tx.Commit()
}
