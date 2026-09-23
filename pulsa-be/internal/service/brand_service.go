package service

import (
	"context"
	"strings"

	"pulsa2/internal/repository"
)

type BrandService struct {
	repo *repository.BrandRepository
}

func NewBrandService(repo *repository.BrandRepository) *BrandService {
	return &BrandService{repo: repo}
}

func (s *BrandService) List(ctx context.Context) ([]repository.MasterSimpleRow, error) {
	return s.repo.List(ctx)
}

func (s *BrandService) Get(ctx context.Context, id int64) (*repository.MasterSimpleRow, error) {
	return s.repo.Get(ctx, id)
}

func (s *BrandService) Create(ctx context.Context, nama string, aktif bool) (int64, error) {
	return s.repo.Create(ctx, strings.TrimSpace(nama), aktif)
}

func (s *BrandService) Update(ctx context.Context, id int64, nama string, aktif bool) error {
	return s.repo.Update(ctx, id, strings.TrimSpace(nama), aktif)
}

func (s *BrandService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
