package service

import (
	"context"
	"errors"
	"fmt"

	"pulsa2/internal/repository"
)

type ProdukAppPricingService struct {
	repo *repository.ProdukAppPricingRepository
}

func NewProdukAppPricingService(repo *repository.ProdukAppPricingRepository) *ProdukAppPricingService {
	return &ProdukAppPricingService{repo: repo}
}

func (s *ProdukAppPricingService) List(ctx context.Context, q string, aktif *bool, limit, offset int) ([]repository.ProdukAppPricingRow, int64, error) {
	return s.repo.List(ctx, q, aktif, limit, offset)
}

func (s *ProdukAppPricingService) Get(ctx context.Context, id int64) (*repository.ProdukAppPricingRow, error) {
	return s.repo.Get(ctx, id)
}

func (s *ProdukAppPricingService) Create(ctx context.Context, in repository.ProdukAppPricingUpsertInput) (int64, error) {
	normalized, err := normalizeProdukAppPricingInput(in)
	if err != nil {
		return 0, err
	}
	exist, err := s.repo.ExistsByProdukDefaultProvider(ctx, normalized.ProdukID)
	if err != nil {
		return 0, err
	}
	if exist {
		return 0, errors.New("produk_app_pricing sudah ada untuk produk ini")
	}
	return s.repo.Create(ctx, normalized)
}

func (s *ProdukAppPricingService) Update(ctx context.Context, in repository.ProdukAppPricingUpsertInput) error {
	normalized, err := normalizeProdukAppPricingInput(in)
	if err != nil {
		return err
	}
	normalized.ID = in.ID
	return s.repo.Update(ctx, normalized)
}

func (s *ProdukAppPricingService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func normalizeProdukAppPricingInput(in repository.ProdukAppPricingUpsertInput) (repository.ProdukAppPricingUpsertInput, error) {
	if in.ProdukID <= 0 {
		return repository.ProdukAppPricingUpsertInput{}, fmt.Errorf("produk_id invalid")
	}
	return in, nil
}
