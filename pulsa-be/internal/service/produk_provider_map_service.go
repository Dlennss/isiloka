package service

import (
	"context"
	"fmt"
	"strings"

	"pulsa2/internal/repository"
)

type ProdukProviderMapService struct {
	repo *repository.ProdukProviderMapRepository
}

func NewProdukProviderMapService(repo *repository.ProdukProviderMapRepository) *ProdukProviderMapService {
	return &ProdukProviderMapService{repo: repo}
}

func (s *ProdukProviderMapService) List(ctx context.Context, q string) ([]repository.ProdukProviderMapRow, error) {
	return s.repo.List(ctx, strings.TrimSpace(q))
}

func (s *ProdukProviderMapService) Get(ctx context.Context, id int64) (*repository.ProdukProviderMapRow, error) {
	return s.repo.Get(ctx, id)
}

func (s *ProdukProviderMapService) Create(ctx context.Context, in repository.ProdukProviderMapUpsertInput) (int64, error) {
	normalized, err := normalizeProdukProviderMapInput(in)
	if err != nil {
		return 0, err
	}
	return s.repo.Create(ctx, normalized)
}

func (s *ProdukProviderMapService) Update(ctx context.Context, in repository.ProdukProviderMapUpsertInput) error {
	normalized, err := normalizeProdukProviderMapInput(in)
	if err != nil {
		return err
	}
	normalized.ID = in.ID
	return s.repo.Update(ctx, normalized)
}

func (s *ProdukProviderMapService) UpdateAktif(ctx context.Context, id int64, aktif bool) error {
	return s.repo.UpdateAktif(ctx, id, aktif)
}

func (s *ProdukProviderMapService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func normalizeProdukProviderMapInput(in repository.ProdukProviderMapUpsertInput) (repository.ProdukProviderMapUpsertInput, error) {
	if in.ProdukID <= 0 {
		return in, fmt.Errorf("produk_id invalid")
	}

	in.Provider = strings.ToLower(strings.TrimSpace(in.Provider))
	if in.Provider == "" {
		return in, fmt.Errorf("provider required")
	}

	in.KodeProvider = strings.ToUpper(strings.TrimSpace(in.KodeProvider))
	if in.KodeProvider == "" {
		return in, fmt.Errorf("kode_provider required")
	}
	if in.SpecialCode != nil {
		v := strings.ToUpper(strings.TrimSpace(*in.SpecialCode))
		if v == "" {
			in.SpecialCode = nil
		} else {
			in.SpecialCode = &v
		}
	}
	if in.Mode != nil {
		v := strings.ToUpper(strings.TrimSpace(*in.Mode))
		if v == "" {
			in.Mode = nil
		} else {
			switch v {
			case "ELEKTRIK", "PPOB", "WALLET_PPOB", "DIRECT":
				in.Mode = &v
			default:
				return in, fmt.Errorf("mode invalid")
			}
		}
	}

	if in.MinimalNominal != nil && *in.MinimalNominal <= 0 {
		return in, fmt.Errorf("minimal_nominal harus > 0")
	}
	if in.MaksimalNominal != nil && *in.MaksimalNominal <= 0 {
		return in, fmt.Errorf("maksimal_nominal harus > 0")
	}
	if in.MinimalNominal != nil && in.MaksimalNominal != nil && *in.MinimalNominal > *in.MaksimalNominal {
		return in, fmt.Errorf("minimal_nominal tidak boleh lebih besar dari maksimal_nominal")
	}
	if in.FeeRp < 0 {
		return in, fmt.Errorf("fee_rp invalid")
	}

	return in, nil
}
