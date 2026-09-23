package repository

import (
	"context"
	"database/sql"
	"errors"
)

func (r *AppOrderPaymentRepository) GetByOrderID(ctx context.Context, orderID string) (*AppOrderPaymentRow, error) {
	var (
		row            AppOrderPaymentRow
		transactionID  sql.NullString
		paymentType    sql.NullString
		transactionSt  sql.NullString
		fraudStatus    sql.NullString
		acquirer       sql.NullString
		qrURL          sql.NullString
		rawRequest     sql.NullString
		rawCallback    sql.NullString
		paidAt         sql.NullTime
		expiredAt      sql.NullTime
		settlementTime sql.NullTime
		dibuatPada     sql.NullTime
		diubahPada     sql.NullTime
	)
	err := r.db.QueryRowContext(ctx, `
SELECT
  id, app_order_id, order_id, transaction_id, gross_amount, payment_type,
  transaction_status, fraud_status, acquirer, qr_url,
  raw_request::text, raw_callback::text,
  paid_at, expired_at, settlement_time, dibuat_pada, diubah_pada
FROM public.app_order_payment
WHERE order_id = $1
LIMIT 1
`, orderID).Scan(
		&row.ID,
		&row.AppOrderID,
		&row.OrderID,
		&transactionID,
		&row.GrossAmount,
		&paymentType,
		&transactionSt,
		&fraudStatus,
		&acquirer,
		&qrURL,
		&rawRequest,
		&rawCallback,
		&paidAt,
		&expiredAt,
		&settlementTime,
		&dibuatPada,
		&diubahPada,
	)
	if err != nil {
		return nil, err
	}
	if transactionID.Valid {
		v := transactionID.String
		row.TransactionID = &v
	}
	if paymentType.Valid {
		v := paymentType.String
		row.PaymentType = &v
	}
	if transactionSt.Valid {
		v := transactionSt.String
		row.TransactionStatus = &v
	}
	if fraudStatus.Valid {
		v := fraudStatus.String
		row.FraudStatus = &v
	}
	if acquirer.Valid {
		v := acquirer.String
		row.Acquirer = &v
	}
	if qrURL.Valid {
		v := qrURL.String
		row.QRURL = &v
	}
	if rawRequest.Valid {
		v := rawRequest.String
		row.RawRequest = &v
	}
	if rawCallback.Valid {
		v := rawCallback.String
		row.RawCallback = &v
	}
	if paidAt.Valid {
		v := paidAt.Time
		row.PaidAt = &v
	}
	if expiredAt.Valid {
		v := expiredAt.Time
		row.ExpiredAt = &v
	}
	if settlementTime.Valid {
		v := settlementTime.Time
		row.SettlementTime = &v
	}
	if dibuatPada.Valid {
		v := dibuatPada.Time
		row.DibuatPada = &v
	}
	if diubahPada.Valid {
		v := diubahPada.Time
		row.DiubahPada = &v
	}
	return &row, nil
}

func (r *AppOrderPaymentRepository) GetLatestByAppOrderID(ctx context.Context, appOrderID int64) (*AppOrderPaymentRow, error) {
	var (
		row            AppOrderPaymentRow
		transactionID  sql.NullString
		paymentType    sql.NullString
		transactionSt  sql.NullString
		fraudStatus    sql.NullString
		acquirer       sql.NullString
		qrURL          sql.NullString
		rawRequest     sql.NullString
		rawCallback    sql.NullString
		paidAt         sql.NullTime
		expiredAt      sql.NullTime
		settlementTime sql.NullTime
		dibuatPada     sql.NullTime
		diubahPada     sql.NullTime
	)
	err := r.db.QueryRowContext(ctx, `
SELECT
  id, app_order_id, order_id, transaction_id, gross_amount, payment_type,
  transaction_status, fraud_status, acquirer, qr_url,
  raw_request::text, raw_callback::text,
  paid_at, expired_at, settlement_time, dibuat_pada, diubah_pada
FROM public.app_order_payment
WHERE app_order_id = $1
ORDER BY id DESC
LIMIT 1
`, appOrderID).Scan(
		&row.ID,
		&row.AppOrderID,
		&row.OrderID,
		&transactionID,
		&row.GrossAmount,
		&paymentType,
		&transactionSt,
		&fraudStatus,
		&acquirer,
		&qrURL,
		&rawRequest,
		&rawCallback,
		&paidAt,
		&expiredAt,
		&settlementTime,
		&dibuatPada,
		&diubahPada,
	)
	if err != nil {
		return nil, err
	}
	if transactionID.Valid {
		v := transactionID.String
		row.TransactionID = &v
	}
	if paymentType.Valid {
		v := paymentType.String
		row.PaymentType = &v
	}
	if transactionSt.Valid {
		v := transactionSt.String
		row.TransactionStatus = &v
	}
	if fraudStatus.Valid {
		v := fraudStatus.String
		row.FraudStatus = &v
	}
	if acquirer.Valid {
		v := acquirer.String
		row.Acquirer = &v
	}
	if qrURL.Valid {
		v := qrURL.String
		row.QRURL = &v
	}
	if rawRequest.Valid {
		v := rawRequest.String
		row.RawRequest = &v
	}
	if rawCallback.Valid {
		v := rawCallback.String
		row.RawCallback = &v
	}
	if paidAt.Valid {
		v := paidAt.Time
		row.PaidAt = &v
	}
	if expiredAt.Valid {
		v := expiredAt.Time
		row.ExpiredAt = &v
	}
	if settlementTime.Valid {
		v := settlementTime.Time
		row.SettlementTime = &v
	}
	if dibuatPada.Valid {
		v := dibuatPada.Time
		row.DibuatPada = &v
	}
	if diubahPada.Valid {
		v := diubahPada.Time
		row.DiubahPada = &v
	}
	return &row, nil
}

func (r *AppOrderPaymentRepository) GetMemberSaldo(ctx context.Context, memberID int64) (int64, error) {
	if memberID <= 0 {
		return 0, errors.New("member_id tidak valid")
	}
	var saldo int64
	if err := r.db.QueryRowContext(ctx, `
SELECT saldo
FROM public.dompet_member
WHERE member_id = $1
LIMIT 1
`, memberID).Scan(&saldo); err != nil {
		return 0, err
	}
	return saldo, nil
}

func (r *AppOrderPaymentRepository) GetMemberFundingBalance(ctx context.Context, memberID int64) (wallet, credit int64, err error) {
	if memberID <= 0 {
		return 0, 0, errors.New("member_id tidak valid")
	}
	// Kredit yang disetujui sudah dicairkan ke dompet utama. Parameter credit
	// dipertahankan untuk kompatibilitas pemanggil lama, tetapi selalu nol agar
	// transaksi tidak pernah memakai saldo kredit sebagai sumber pembayaran.
	err = r.db.QueryRowContext(ctx, `
SELECT COALESCE(saldo, 0)
FROM public.dompet_member
WHERE member_id = $1
LIMIT 1
`, memberID).Scan(&wallet)
	credit = 0
	return wallet, credit, err
}
