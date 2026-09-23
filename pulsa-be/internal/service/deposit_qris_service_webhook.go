package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"pulsa2/internal/helper"
)

func (s *DepositService) HandleMidtransWebhook(ctx context.Context, raw []byte, payload map[string]any) (int, map[string]any) {
	serverKey, err := midtransServerKeyFromEnv()
	if err != nil {
		return http.StatusInternalServerError, map[string]any{"ok": false, "error": err.Error()}
	}

	orderID := strings.TrimSpace(helper.ToString(payload["order_id"]))
	statusCode := strings.TrimSpace(helper.ToString(payload["status_code"]))
	grossAmount := strings.TrimSpace(helper.ToString(payload["gross_amount"]))
	signatureKey := strings.TrimSpace(helper.ToString(payload["signature_key"]))
	if orderID == "" || statusCode == "" || grossAmount == "" || signatureKey == "" {
		return http.StatusBadRequest, map[string]any{"ok": false, "error": "payload midtrans tidak lengkap"}
	}
	if !verifyMidtransSignature(orderID, statusCode, grossAmount, serverKey, signatureKey) {
		return http.StatusUnauthorized, map[string]any{"ok": false, "error": "invalid signature"}
	}
	if err := s.applyMidtransQrisStatus(ctx, orderID, payload); err != nil {
		return http.StatusInternalServerError, map[string]any{"ok": false, "error": err.Error()}
	}

	row, err := s.repo.GetByRefID(ctx, orderID)
	if err != nil {
		return http.StatusInternalServerError, map[string]any{"ok": false, "error": err.Error()}
	}
	feeAdmin := calcDepositQrisFee()
	return http.StatusOK, map[string]any{
		"ok": true,
		"item": map[string]any{
			"ref_id":       row.RefID,
			"amount":       row.Amount,
			"fee_admin":    feeAdmin,
			"gross_amount": row.Amount + feeAdmin,
			"status":       row.Status,
			"qr_url":       row.BuktiURL,
		},
	}
}

func (s *DepositService) RefreshQrisStatusByRefID(ctx context.Context, refID string) error {
	refID = strings.TrimSpace(refID)
	if refID == "" {
		return errors.New("ref_id wajib diisi")
	}
	row, err := s.repo.GetByRefID(ctx, refID)
	if err != nil {
		return err
	}
	if strings.TrimSpace(strings.ToLower(row.Metode)) != depositQrisMethod {
		return errors.New("deposit qris tidak ditemukan")
	}
	switch strings.TrimSpace(strings.ToLower(row.Status)) {
	case "approved", "rejected":
		return nil
	}

	serverKey, err := midtransServerKeyFromEnv()
	if err != nil {
		return err
	}

	statusResp, err := newMidtransCoreAPIClient(serverKey).CheckTransaction(refID)
	if !isNilLikeError(err) {
		return fmt.Errorf("gagal cek status qris ke midtrans: %w", err)
	}
	payload := map[string]any{
		"order_id":           statusResp.OrderID,
		"status_code":        statusResp.StatusCode,
		"gross_amount":       statusResp.GrossAmount,
		"transaction_status": statusResp.TransactionStatus,
		"fraud_status":       statusResp.FraudStatus,
		"payment_type":       statusResp.PaymentType,
		"transaction_id":     statusResp.TransactionID,
		"acquirer":           statusResp.Acquirer,
		"settlement_time":    statusResp.SettlementTime,
		"expiry_time":        statusResp.ExpiryTime,
		"transaction_time":   statusResp.TransactionTime,
	}
	return s.applyMidtransQrisStatus(ctx, refID, payload)
}

func (s *DepositService) applyMidtransQrisStatus(ctx context.Context, refID string, payload map[string]any) error {
	refID = strings.TrimSpace(refID)
	if refID == "" {
		return errors.New("ref_id wajib diisi")
	}
	row, err := s.repo.GetByRefID(ctx, refID)
	if err != nil {
		return err
	}
	if strings.TrimSpace(strings.ToLower(row.Metode)) != depositQrisMethod {
		return errors.New("deposit qris tidak ditemukan")
	}

	transactionStatus := strings.TrimSpace(helper.ToString(payload["transaction_status"]))
	fraudStatus := strings.TrimSpace(helper.ToString(payload["fraud_status"]))
	statusCode := strings.TrimSpace(helper.ToString(payload["status_code"]))
	paymentType := strings.TrimSpace(helper.ToString(payload["payment_type"]))
	transactionID := strings.TrimSpace(helper.ToString(payload["transaction_id"]))
	acquirer := strings.TrimSpace(helper.ToString(payload["acquirer"]))
	qrURL := strings.TrimSpace(row.BuktiURL)
	if qrURL == "" {
		qrURL = firstQRURLFromPayload(payload)
	}
	note := buildDepositQrisNote(transactionStatus, paymentType, transactionID, acquirer)
	if err := s.repo.UpdateQrisPending(ctx, refID, qrURL, note); err != nil {
		return err
	}

	if isMidtransPaid(transactionStatus, fraudStatus, statusCode) {
		if err := s.repo.ApproveQrisByRefID(ctx, refID, "QRIS settled"); err != nil {
			return err
		}
		grossAmount := depositQrisPaidGrossAmount(row.Amount, payload)
		_, _, err := s.bankRepo.CreditSystemQrisIfNeeded(ctx, refID, grossAmount, "RETAIL_TOPUP_QRIS", "Topup QRIS retail settled", &row.MemberID)
		return err
	}

	switch mapOrderStatusFromMidtrans(transactionStatus, fraudStatus, statusCode) {
	case "expired":
		return s.repo.RejectQrisByRefID(ctx, refID, "QRIS expired")
	case "cancelled":
		return s.repo.RejectQrisByRefID(ctx, refID, "QRIS cancelled")
	case "failed":
		return s.repo.RejectQrisByRefID(ctx, refID, "QRIS failed")
	default:
		return nil
	}
}

func depositQrisPaidGrossAmount(rowAmount int64, payload map[string]any) int64 {
	if grossAmount := helper.ToI64(payload["gross_amount"]); grossAmount > 0 {
		return grossAmount
	}
	return rowAmount + calcDepositQrisFee()
}
