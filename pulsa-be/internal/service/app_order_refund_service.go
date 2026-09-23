package service

import (
	"context"
	"fmt"
	"strings"

	"pulsa2/internal/repository"
)

type AppOrderRefundService struct {
	orderRepo *repository.AppOrderRepository
}

func NewAppOrderRefundService(orderRepo *repository.AppOrderRepository) *AppOrderRefundService {
	return &AppOrderRefundService{orderRepo: orderRepo}
}

func (s *AppOrderRefundService) ClaimGuestRefund(ctx context.Context, memberID int64, invoiceID, guestEmail, guestPhone string) (*repository.AppOrderGuestRefundRow, error) {
	if memberID <= 0 {
		return nil, fmt.Errorf("user tidak valid")
	}
	invoiceID = strings.TrimSpace(invoiceID)
	guestEmail = normalizeGuestEmail(guestEmail)
	guestPhone = normalizeGuestPhone(guestPhone)

	if invoiceID == "" {
		return nil, fmt.Errorf("invoice_id wajib diisi")
	}
	if guestEmail == "" || guestPhone == "" {
		return nil, fmt.Errorf("guest_email dan guest_phone wajib diisi")
	}
	return s.orderRepo.ClaimGuestRefundToMember(ctx, memberID, invoiceID, guestEmail, guestPhone)
}
