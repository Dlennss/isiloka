package service

import (
	"context"

	"pulsa2/internal/repository"
)

type AppBrandService struct {
	repo *repository.AppBrandRepository
}

func NewAppBrandService(repo *repository.AppBrandRepository) *AppBrandService {
	return &AppBrandService{repo: repo}
}

func (s *AppBrandService) List(ctx context.Context, kategoriID int64) ([]repository.MasterSimpleRow, error) {
	return s.repo.List(ctx, kategoriID)
}
