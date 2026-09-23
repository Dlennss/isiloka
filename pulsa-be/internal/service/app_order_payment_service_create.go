package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"pulsa2/internal/repository"
)

func (s *AppOrderPaymentService) CreateByInvoiceID(ctx context.Context, invoiceID string) (*repository.AppOrderPaymentRow, error) {
	order, err := s.orderRepo.GetByInvoiceID(ctx, invoiceID)
	if err != nil {
		return nil, err
	}
	if order.Status != "pending_payment" {
		return nil, fmt.Errorf("order tidak dalam status pending_payment")
	}

	existing, err := s.paymentRepo.GetByOrderID(ctx, order.InvoiceID)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	if order.HargaFinal <= 0 {
		return nil, fmt.Errorf("nominal pembayaran tidak valid")
	}

	if strings.EqualFold(strings.TrimSpace(order.BuyerType), "guest") {
		serverKey, err := midtransServerKeyFromEnv()
		if err != nil {
			return nil, err
		}

		payload := buildMidtransChargeRequest(order, order.HargaFinal)
		resp, err := chargeMidtransQRIS(serverKey, payload)
		if err != nil {
			return nil, err
		}
		rawResp, _ := json.Marshal(resp)
		payment := repository.AppOrderPaymentCreateInput{
			AppOrderID:        order.ID,
			OrderID:           order.InvoiceID,
			TransactionID:     resp.TransactionID,
			GrossAmount:       order.HargaFinal,
			PaymentType:       resp.PaymentType,
			TransactionStatus: resp.TransactionStatus,
			FraudStatus:       resp.FraudStatus,
			Acquirer:          resp.Acquirer,
			QRURL:             firstQRURL(resp.Actions),
			RawRequest:        repository.BuildPaymentRawRequest(payload, 0, order.HargaFinal, 0, order.HargaFinal, order.HargaFinal),
			RawCallback:       string(rawResp),
			ExpiredAt:         parseMidtransTime(resp.ExpiryTime),
		}
		if err := s.paymentRepo.Create(ctx, payment); err != nil {
			return nil, err
		}
		return s.paymentRepo.GetByOrderID(ctx, order.InvoiceID)
	}

	if order.BuyerType != "user" || order.MemberID == nil || *order.MemberID <= 0 {
		return nil, fmt.Errorf("tipe pembeli tidak valid")
	}
	now := time.Now()
	payment := repository.AppOrderPaymentCreateInput{
		AppOrderID:        order.ID,
		OrderID:           order.InvoiceID,
		PaymentType:       "balance",
		TransactionStatus: "settlement",
		FraudStatus:       "accept",
		PaidAt:            &now,
		SettlementTime:    &now,
	}
	if err := s.paymentRepo.CreateWithBalanceDebit(ctx, payment, *order.MemberID, order.HargaFinal); err != nil {
		return nil, err
	}
	if s.fulfillmentSvc != nil {
		if err := s.fulfillmentSvc.DispatchPaidOrderByInvoiceID(ctx, order.InvoiceID); err != nil {
			log.Printf("[app_order_payment] fulfillment Pulsa24Jam gagal invoice=%s err=%v", order.InvoiceID, err)
		}
	}
	return s.paymentRepo.GetByOrderID(ctx, order.InvoiceID)
}
