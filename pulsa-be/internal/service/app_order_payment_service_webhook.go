package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
)

func (s *AppOrderPaymentService) HandleMidtransWebhook(ctx context.Context, raw []byte, payload map[string]any) (int, map[string]any) {
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

	if err := s.applyMidtransStatus(ctx, orderID, string(raw), payload); err != nil {
		return http.StatusInternalServerError, map[string]any{"ok": false, "error": err.Error()}
	}

	updated, err := s.paymentRepo.GetByOrderID(ctx, orderID)
	if err != nil {
		return http.StatusInternalServerError, map[string]any{"ok": false, "error": err.Error()}
	}
	orderStatus := mapOrderStatusFromMidtrans(
		strings.TrimSpace(helper.ToString(payload["transaction_status"])),
		strings.TrimSpace(helper.ToString(payload["fraud_status"])),
		strings.TrimSpace(helper.ToString(payload["status_code"])),
	)
	return http.StatusOK, map[string]any{"ok": true, "item": updated, "order_status": orderStatus}
}

func (s *AppOrderPaymentService) RefreshStatusByOrderID(ctx context.Context, orderID string) error {
	orderID = strings.TrimSpace(orderID)
	if orderID == "" {
		return nil
	}

	order, err := s.orderRepo.GetByInvoiceID(ctx, orderID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}
	switch order.Status {
	case "success", "refunded", "failed", "expired", "cancelled":
		return nil
	}

	payment, err := s.paymentRepo.GetByOrderID(ctx, orderID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}
	if strings.EqualFold(derefString(payment.PaymentType), "wallet") {
		return nil
	}

	serverKey, err := midtransServerKeyFromEnv()
	if err != nil {
		return nil
	}

	client := newMidtransCoreAPIClient(serverKey)
	statusResp, midErr := client.CheckTransaction(orderID)
	if !isNilLikeError(midErr) {
		return nil
	}
	rawResp, _ := json.Marshal(statusResp)

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

	return s.applyMidtransStatus(ctx, orderID, string(rawResp), payload)
}

func (s *AppOrderPaymentService) applyMidtransStatus(ctx context.Context, orderID string, raw string, payload map[string]any) error {
	payment, err := s.paymentRepo.GetByOrderID(ctx, orderID)
	if err != nil {
		return err
	}

	transactionStatus := strings.TrimSpace(helper.ToString(payload["transaction_status"]))
	fraudStatus := strings.TrimSpace(helper.ToString(payload["fraud_status"]))
	statusCode := strings.TrimSpace(helper.ToString(payload["status_code"]))
	paymentType := strings.TrimSpace(helper.ToString(payload["payment_type"]))
	transactionID := strings.TrimSpace(helper.ToString(payload["transaction_id"]))
	acquirer := strings.TrimSpace(helper.ToString(payload["acquirer"]))
	settlementTime := parseMidtransTime(strings.TrimSpace(helper.ToString(payload["settlement_time"])))
	expiredAt := parseMidtransTime(strings.TrimSpace(helper.ToString(payload["expiry_time"])))
	transactionTime := parseMidtransTime(strings.TrimSpace(helper.ToString(payload["transaction_time"])))
	paidAt := settlementTime
	if paidAt == nil && isMidtransPaid(transactionStatus, fraudStatus, statusCode) {
		paidAt = transactionTime
	}
	qrURL := payment.QRURL
	if qrURL == nil || strings.TrimSpace(*qrURL) == "" {
		qr := firstQRURLFromPayload(payload)
		if qr != "" {
			qrURL = &qr
		}
	}

	if err := s.paymentRepo.UpdateByOrderID(ctx, repository.AppOrderPaymentUpdateInput{
		OrderID:           orderID,
		TransactionID:     transactionID,
		PaymentType:       paymentType,
		TransactionStatus: transactionStatus,
		FraudStatus:       fraudStatus,
		Acquirer:          acquirer,
		QRURL:             derefString(qrURL),
		RawCallback:       raw,
		PaidAt:            paidAt,
		ExpiredAt:         expiredAt,
		SettlementTime:    settlementTime,
	}); err != nil {
		return err
	}

	orderStatus := mapOrderStatusFromMidtrans(transactionStatus, fraudStatus, statusCode)
	if err := s.orderRepo.UpdateStatusByInvoiceID(ctx, orderID, orderStatus); err != nil {
		return err
	}
	if orderStatus == "failed" || orderStatus == "expired" || orderStatus == "cancelled" {
		if refunded, err := s.paymentRepo.RefundWalletDebitIfNeeded(ctx, orderID, "APP_ORDER_WALLET_REFUND", "refund saldo karena pembayaran qris gagal/expired"); err != nil {
			log.Printf("[app_order_payment] refund wallet gagal invoice=%s err=%v", orderID, err)
		} else if refunded {
			log.Printf("[app_order_payment] wallet refunded invoice=%s", orderID)
		}
	}
	if orderStatus == "paid" && s.bankRepo != nil {
		order, orderErr := s.orderRepo.GetByInvoiceID(ctx, orderID)
		if orderErr == nil && strings.EqualFold(strings.TrimSpace(order.BuyerType), "user") && payment.GrossAmount > 0 {
			var memberID *int64
			if order.MemberID != nil && *order.MemberID > 0 {
				memberID = order.MemberID
			}
			if _, _, bankErr := s.bankRepo.CreditSystemQrisIfNeeded(ctx, orderID, payment.GrossAmount, "RETAIL_PURCHASE_QRIS", "Pembayaran QRIS retail settled", memberID); bankErr != nil {
				return bankErr
			}
		}
	}
	if orderStatus == "paid" && s.fulfillmentSvc != nil {
		if err := s.fulfillmentSvc.DispatchPaidOrderByInvoiceID(ctx, orderID); err != nil {
			log.Printf("[app_order_payment] fulfillment dispatch gagal invoice=%s err=%v", orderID, err)
		}
	}
	return nil
}
