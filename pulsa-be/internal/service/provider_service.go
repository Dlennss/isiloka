package service

import (
	"context"
	"strings"

	"pulsa2/internal/repository"
)

type ProviderService struct {
	repo *repository.ProviderRepository
}

func NewProviderService(repo *repository.ProviderRepository) *ProviderService {
	return &ProviderService{repo: repo}
}

func (s *ProviderService) List(ctx context.Context) ([]repository.MasterSimpleRow, error) {
	return s.repo.List(ctx)
}

func (s *ProviderService) Get(ctx context.Context, id int64) (*repository.MasterSimpleRow, error) {
	return s.repo.Get(ctx, id)
}

func (s *ProviderService) Create(ctx context.Context, nama string, aktif bool) (int64, error) {
	return s.repo.Create(ctx, strings.TrimSpace(nama), aktif)
}

func (s *ProviderService) Update(ctx context.Context, id int64, nama string, aktif bool) error {
	return s.repo.Update(ctx, id, strings.TrimSpace(nama), aktif)
}

func (s *ProviderService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
