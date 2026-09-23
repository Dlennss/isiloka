package service

import (
	"context"
	"fmt"
	"strings"

	"pulsa2/internal/repository"
)

type AppOrderAdminDetail struct {
	Order       *repository.AppOrderRow            `json:"order"`
	Payment     *repository.AppOrderPaymentRow     `json:"payment,omitempty"`
	ProviderTrx *repository.AppOrderProviderTrxRow `json:"provider_trx,omitempty"`
}

type AppOrderAdminService struct {
	orderRepo       *repository.AppOrderRepository
	paymentRepo     *repository.AppOrderPaymentRepository
	appProviderRepo *repository.AppOrderProviderTrxRepository
}

func NewAppOrderAdminService(orderRepo *repository.AppOrderRepository, paymentRepo *repository.AppOrderPaymentRepository, appProviderRepo *repository.AppOrderProviderTrxRepository) *AppOrderAdminService {
	return &AppOrderAdminService{orderRepo: orderRepo, paymentRepo: paymentRepo, appProviderRepo: appProviderRepo}
}

func (s *AppOrderAdminService) GetDetailByInvoiceID(ctx context.Context, invoiceID string) (*AppOrderAdminDetail, error) {
	invoiceID = strings.TrimSpace(invoiceID)
	if invoiceID == "" {
		return nil, fmt.Errorf("invoice_id wajib diisi")
	}

	order, err := s.orderRepo.GetByInvoiceID(ctx, invoiceID)
	if err != nil {
		return nil, err
	}

	detail := &AppOrderAdminDetail{Order: order}

	if payment, err := s.paymentRepo.GetLatestByAppOrderID(ctx, order.ID); err == nil {
		detail.Payment = payment
	}
	if providerTrx, err := s.appProviderRepo.GetLatestByAppOrderID(ctx, order.ID); err == nil {
		detail.ProviderTrx = providerTrx
	}

	return detail, nil
}

func (s *AppOrderAdminService) ListProviderTrx(ctx context.Context, f repository.AppOrderProviderTrxListFilter) ([]repository.AppOrderProviderTrxAdminRow, error) {
	f.Provider = strings.TrimSpace(strings.ToLower(f.Provider))
	if f.Provider == "" {
		f.Provider = "yuscom"
	}
	switch f.Provider {
	case "yuscom", "javapay", "talentapay", "pulsa24jam":
	default:
		return nil, fmt.Errorf("provider tidak valid")
	}
	f.Status = strings.TrimSpace(strings.ToLower(f.Status))
	if f.Status != "" {
		switch f.Status {
		case "pending", "success", "failed":
		default:
			return nil, fmt.Errorf("status tidak valid")
		}
	}
	if f.Limit <= 0 {
		f.Limit = 20
	}
	if f.Limit > 100 {
		f.Limit = 100
	}
	if f.Offset < 0 {
		f.Offset = 0
	}
	return s.appProviderRepo.ListAdmin(ctx, f)
}
