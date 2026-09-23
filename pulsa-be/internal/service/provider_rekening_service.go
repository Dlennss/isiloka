package service

import (
	"context"
	"strings"

	"pulsa2/internal/repository"
)

type ProviderRekeningService struct {
	repo *repository.ProviderRekeningRepository
}

func NewProviderRekeningService(repo *repository.ProviderRekeningRepository) *ProviderRekeningService {
	return &ProviderRekeningService{repo: repo}
}

func (s *ProviderRekeningService) List(ctx context.Context, provider, q string, aktifOnly bool, limit, offset int) ([]repository.ProviderRekeningRow, int64, error) {
	return s.repo.List(ctx, strings.TrimSpace(provider), strings.TrimSpace(q), aktifOnly, limit, offset)
}

func (s *ProviderRekeningService) Get(ctx context.Context, id int64) (*repository.ProviderRekeningRow, error) {
	return s.repo.Get(ctx, id)
}

func (s *ProviderRekeningService) Create(ctx context.Context, in repository.ProviderRekeningUpsertInput) (int64, error) {
	return s.repo.Create(ctx, in)
}

func (s *ProviderRekeningService) Update(ctx context.Context, id int64, in repository.ProviderRekeningUpsertInput) error {
	return s.repo.Update(ctx, id, in)
}

func (s *ProviderRekeningService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
