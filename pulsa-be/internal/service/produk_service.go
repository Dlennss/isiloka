package service

import (
	"context"
	"fmt"
	"strings"

	"pulsa2/internal/repository"
)

type ProdukService struct {
	repo *repository.ProdukRepository
}

func NewProdukService(repo *repository.ProdukRepository) *ProdukService {
	return &ProdukService{repo: repo}
}

func (s *ProdukService) List(ctx context.Context, q, sku, groupName string, kategoriID, brandID int64, limit, offset int) ([]repository.ProdukRow, int64, error) {
	return s.repo.List(ctx, strings.TrimSpace(q), strings.TrimSpace(sku), strings.TrimSpace(groupName), kategoriID, brandID, limit, offset)
}

func (s *ProdukService) Get(ctx context.Context, id int64) (*repository.ProdukRow, error) {
	return s.repo.Get(ctx, id)
}

func (s *ProdukService) ListGroupNames(ctx context.Context) ([]string, error) {
	return s.repo.ListGroupNames(ctx)
}

func (s *ProdukService) Create(ctx context.Context, in repository.ProdukUpsertInput) (int64, error) {
	normalized, err := normalizeTipeHarga(in.TipeHarga)
	if err != nil {
		return 0, err
	}
	in.TipeHarga = normalized
	switch in.TipeHarga {
	case "FIXED":
		in.MaksimalNominal = nil
	case "OPEN_AMOUNT":
		in.Nominal = nil
	}
	return s.repo.Create(ctx, in)
}

func (s *ProdukService) Update(ctx context.Context, in repository.ProdukUpsertInput) error {
	normalized, err := normalizeTipeHarga(in.TipeHarga)
	if err != nil {
		return err
	}
	in.TipeHarga = normalized
	switch in.TipeHarga {
	case "FIXED":
		in.MaksimalNominal = nil
	case "OPEN_AMOUNT":
		in.Nominal = nil
	}
	return s.repo.Update(ctx, in)
}

func (s *ProdukService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func normalizeTipeHarga(v string) (string, error) {
	switch strings.ToUpper(strings.TrimSpace(v)) {
	case "FIXED":
		return "FIXED", nil
	case "OPEN_AMOUNT":
		return "OPEN_AMOUNT", nil
	default:
		return "", fmt.Errorf("tipe_harga invalid: gunakan FIXED atau OPEN_AMOUNT")
	}
}
