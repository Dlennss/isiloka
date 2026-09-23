package service

import (
	"context"
	"strings"

	"pulsa2/internal/repository"
)

type MemberFeeProductService struct {
	repo *repository.MemberFeeProductRepository
}

func NewMemberFeeProductService(repo *repository.MemberFeeProductRepository) *MemberFeeProductService {
	return &MemberFeeProductService{repo: repo}
}

func (s *MemberFeeProductService) Upsert(ctx context.Context, memberID int64, produkID int64, kodeProduk string, feePersen *float64, feeRp *int64) error {
	return s.repo.Upsert(ctx, memberID, produkID, strings.ToUpper(strings.TrimSpace(kodeProduk)), feePersen, feeRp)
}

func (s *MemberFeeProductService) Delete(ctx context.Context, memberID int64, produkID int64, kodeProduk string) error {
	return s.repo.Delete(ctx, memberID, produkID, strings.ToUpper(strings.TrimSpace(kodeProduk)))
}

func (s *MemberFeeProductService) List(ctx context.Context, memberID int64) ([]repository.MemberFeeProductRow, error) {
	return s.repo.List(ctx, memberID)
}
