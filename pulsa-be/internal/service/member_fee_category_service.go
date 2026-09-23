package service

import (
	"context"

	"pulsa2/internal/repository"
)

type MemberFeeCategoryService struct {
	repo *repository.MemberFeeCategoryRepository
}

func NewMemberFeeCategoryService(repo *repository.MemberFeeCategoryRepository) *MemberFeeCategoryService {
	return &MemberFeeCategoryService{repo: repo}
}

func (s *MemberFeeCategoryService) Upsert(ctx context.Context, memberID int64, feeCode string, kategoriID, feeRp int64, aktif bool) error {
	return s.repo.Upsert(ctx, memberID, feeCode, kategoriID, feeRp, aktif)
}

func (s *MemberFeeCategoryService) Delete(ctx context.Context, memberID int64, feeCode string, kategoriID int64) error {
	return s.repo.Delete(ctx, memberID, feeCode, kategoriID)
}

func (s *MemberFeeCategoryService) List(ctx context.Context, memberID int64) ([]repository.MemberFeeCategoryRow, error) {
	return s.repo.List(ctx, memberID)
}
