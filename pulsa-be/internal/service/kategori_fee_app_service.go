package service

import (
	"context"
	"fmt"

	"pulsa2/internal/repository"
)

type KategoriFeeAppService struct {
	repo *repository.KategoriFeeAppRepository
}

func NewKategoriFeeAppService(repo *repository.KategoriFeeAppRepository) *KategoriFeeAppService {
	return &KategoriFeeAppService{repo: repo}
}

func (s *KategoriFeeAppService) List(ctx context.Context, q string) ([]repository.KategoriFeeAppRow, error) {
	return s.repo.List(ctx, q)
}

func (s *KategoriFeeAppService) Get(ctx context.Context, id int64) (*repository.KategoriFeeAppRow, error) {
	return s.repo.Get(ctx, id)
}

func (s *KategoriFeeAppService) Create(ctx context.Context, in repository.KategoriFeeAppUpsertInput) (int64, error) {
	normalized, err := normalizeKategoriFeeAppInput(in)
	if err != nil {
		return 0, err
	}
	return s.repo.Create(ctx, normalized)
}

func (s *KategoriFeeAppService) Update(ctx context.Context, in repository.KategoriFeeAppUpsertInput) error {
	normalized, err := normalizeKategoriFeeAppInput(in)
	if err != nil {
		return err
	}
	normalized.ID = in.ID
	return s.repo.Update(ctx, normalized)
}

func (s *KategoriFeeAppService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func normalizeKategoriFeeAppInput(in repository.KategoriFeeAppUpsertInput) (repository.KategoriFeeAppUpsertInput, error) {
	if in.KategoriID <= 0 {
		return repository.KategoriFeeAppUpsertInput{}, fmt.Errorf("kategori_id invalid")
	}
	return in, nil
}
