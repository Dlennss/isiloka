package service

import (
	"context"
	"errors"
	"strings"

	"pulsa2/internal/repository"
)

type AppOrderRefundAdminService struct {
	orderRepo *repository.AppOrderRepository
	authRepo  *repository.AuthRepository
}

func NewAppOrderRefundAdminService(orderRepo *repository.AppOrderRepository, authRepo *repository.AuthRepository) *AppOrderRefundAdminService {
	return &AppOrderRefundAdminService{orderRepo: orderRepo, authRepo: authRepo}
}

func (s *AppOrderRefundAdminService) ListPending(ctx context.Context, limit, offset int, invoiceID string) ([]repository.AdminGuestRefundPendingRow, int64, error) {
	return s.orderRepo.ListAdminGuestRefundPending(ctx, limit, offset, invoiceID)
}

func (s *AppOrderRefundAdminService) ClaimToUser(ctx context.Context, invoiceID string, targetMemberID int64, targetEmail string) (*repository.AppOrderGuestRefundRow, *repository.AuthMeRow, error) {
	var (
		memberID int64
		target   *repository.AuthMeRow
		err      error
	)

	if targetMemberID > 0 {
		target, err = s.authRepo.GetMe(ctx, targetMemberID)
		if err != nil {
			return nil, nil, err
		}
		if target == nil {
			return nil, nil, errors.New("akun tujuan tidak ditemukan")
		}
		memberID = target.ID
	} else {
		row, getErr := s.authRepo.GetByEmail(ctx, targetEmail)
		if getErr != nil {
			return nil, nil, getErr
		}
		if row == nil {
			return nil, nil, errors.New("akun tujuan tidak ditemukan")
		}
		memberID = row.ID
		target, err = s.authRepo.GetMe(ctx, memberID)
		if err != nil {
			return nil, nil, err
		}
		if target == nil {
			return nil, nil, errors.New("akun tujuan tidak ditemukan")
		}
	}

	role := strings.TrimSpace(strings.ToLower(target.Role))
	if role != "user" {
		return nil, nil, errors.New("akun tujuan harus role user")
	}
	if !target.Aktif {
		return nil, nil, errors.New("akun tujuan tidak aktif")
	}

	refundRow, err := s.orderRepo.AdminClaimGuestRefundToMember(ctx, memberID, invoiceID)
	if err != nil {
		return nil, nil, err
	}
	return refundRow, target, nil
}
