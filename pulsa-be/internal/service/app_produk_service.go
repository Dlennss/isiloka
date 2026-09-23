package service

import (
	"context"
	"strings"

	"pulsa2/internal/repository"
)

type AppProdukService struct {
	repo *repository.AppProdukRepository
}

func NewAppProdukService(repo *repository.AppProdukRepository) *AppProdukService {
	return &AppProdukService{repo: repo}
}

func (s *AppProdukService) List(ctx context.Context, q string, kategoriID, brandID int64) ([]repository.AppProdukRow, error) {
	return s.repo.List(ctx, strings.TrimSpace(q), kategoriID, brandID)
}
