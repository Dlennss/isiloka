package repository

import (
	"context"
	"database/sql"
	"encoding/json"
)

type AppOrderPaymentRepository struct {
	db *sql.DB
}

func NewAppOrderPaymentRepository(db *sql.DB) *AppOrderPaymentRepository {
	return &AppOrderPaymentRepository{db: db}
}

func (r *AppOrderPaymentRepository) Create(ctx context.Context, in AppOrderPaymentCreateInput) error {
	return r.db.QueryRowContext(ctx, `
INSERT INTO public.app_order_payment
  (app_order_id, order_id, transaction_id, gross_amount, payment_type, transaction_status, fraud_status, acquirer, qr_url, raw_request, raw_callback, paid_at, expired_at, settlement_time, dibuat_pada, diubah_pada)
VALUES
  ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10::jsonb,$11::jsonb,$12,$13,$14,now(),now())
RETURNING id
`, in.AppOrderID, in.OrderID, nullableStringValue(in.TransactionID), in.GrossAmount, nullableStringValue(in.PaymentType), nullableStringValue(in.TransactionStatus), nullableStringValue(in.FraudStatus), nullableStringValue(in.Acquirer), nullableStringValue(in.QRURL), nullableJSON(in.RawRequest), nullableJSON(in.RawCallback), in.PaidAt, in.ExpiredAt, in.SettlementTime).Scan(&in.ID)
}

func (r *AppOrderPaymentRepository) UpdateByOrderID(ctx context.Context, in AppOrderPaymentUpdateInput) error {
	res, err := r.db.ExecContext(ctx, `
UPDATE public.app_order_payment
SET transaction_id = COALESCE($2, transaction_id),
    payment_type = COALESCE($3, payment_type),
    transaction_status = COALESCE($4, transaction_status),
    fraud_status = COALESCE($5, fraud_status),
    acquirer = COALESCE($6, acquirer),
    qr_url = COALESCE($7, qr_url),
    raw_callback = COALESCE($8::jsonb, raw_callback),
    paid_at = COALESCE($9, paid_at),
    expired_at = COALESCE($10, expired_at),
    settlement_time = COALESCE($11, settlement_time),
    diubah_pada = now()
WHERE order_id = $1
`, in.OrderID, nullableStringValue(in.TransactionID), nullableStringValue(in.PaymentType), nullableStringValue(in.TransactionStatus), nullableStringValue(in.FraudStatus), nullableStringValue(in.Acquirer), nullableStringValue(in.QRURL), nullableJSON(in.RawCallback), in.PaidAt, in.ExpiredAt, in.SettlementTime)
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
	return nil
}
func nullableJSON(v string) any {
	if v == "" {
		return nil
	}
	return v
}

func nullableStringValue(v string) any {
	if v == "" {
		return nil
	}
	return v
}
func BuildPaymentRawRequest(midtransReq any, walletDebit int64, qrisAmount int64, qrisFeeAdmin int64, orderAmount int64, payableAmount int64) string {
	payload := map[string]any{
		"midtrans":       midtransReq,
		"wallet_debit":   walletDebit,
		"qris_amount":    qrisAmount,
		"qris_fee_admin": qrisFeeAdmin,
		"order_amount":   orderAmount,
		"payable_amount": payableAmount,
		"total_amount":   payableAmount,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "{}"
	}
	return string(raw)
}

func BuildBalancePaymentRawRequest(walletDebit, creditDebit, orderAmount int64) string {
	payload := map[string]any{
		"provider":       "pulsa24jam",
		"wallet_debit":   walletDebit,
		"credit_debit":   creditDebit,
		"qris_amount":    0,
		"qris_fee_admin": 0,
		"order_amount":   orderAmount,
		"payable_amount": orderAmount,
		"total_amount":   orderAmount,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "{}"
	}
	return string(raw)
}
