package service

import (
	"context"
	"strings"

	"pulsa2/internal/repository"
)

type KategoriService struct {
	repo *repository.KategoriRepository
}

func NewKategoriService(repo *repository.KategoriRepository) *KategoriService {
	return &KategoriService{repo: repo}
}

func (s *KategoriService) List(ctx context.Context) ([]repository.MasterSimpleRow, error) {
	return s.repo.List(ctx)
}

func (s *KategoriService) Get(ctx context.Context, id int64) (*repository.MasterSimpleRow, error) {
	return s.repo.Get(ctx, id)
}

func (s *KategoriService) Create(ctx context.Context, nama string, aktif bool) (int64, error) {
	return s.repo.Create(ctx, strings.TrimSpace(nama), aktif)
}

func (s *KategoriService) Update(ctx context.Context, id int64, nama string, aktif bool) error {
	return s.repo.Update(ctx, id, strings.TrimSpace(nama), aktif)
}

func (s *KategoriService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
