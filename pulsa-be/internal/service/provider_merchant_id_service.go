package service

import (
	"context"
	"strings"

	"pulsa2/internal/repository"
)

type ProviderMerchantIDService struct {
	repo *repository.ProviderMerchantIDRepository
}

func NewProviderMerchantIDService(repo *repository.ProviderMerchantIDRepository) *ProviderMerchantIDService {
	return &ProviderMerchantIDService{repo: repo}
}

func (s *ProviderMerchantIDService) List(ctx context.Context, provider, q string, aktifOnly bool, limit, offset int) ([]repository.ProviderMerchantIDRow, int64, error) {
	return s.repo.List(ctx, strings.TrimSpace(provider), strings.TrimSpace(q), aktifOnly, limit, offset)
}

func (s *ProviderMerchantIDService) Get(ctx context.Context, id int64) (*repository.ProviderMerchantIDRow, error) {
	return s.repo.Get(ctx, id)
}

func (s *ProviderMerchantIDService) Create(ctx context.Context, in repository.ProviderMerchantIDUpsertInput) (int64, error) {
	return s.repo.Create(ctx, in)
}

func (s *ProviderMerchantIDService) Update(ctx context.Context, id int64, in repository.ProviderMerchantIDUpsertInput) error {
	return s.repo.Update(ctx, id, in)
}

func (s *ProviderMerchantIDService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
