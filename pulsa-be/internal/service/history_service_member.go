package service

import (
	"context"
	"errors"

	"pulsa2/internal/repository"
)

func (s *HistoryService) GetSaldo(ctx context.Context, memberID int64) (int64, error) {
	if memberID <= 0 {
		return 0, errors.New("unauthorized")
	}
	return s.repo.GetSaldo(ctx, memberID)
}

func (s *HistoryService) ListMutasi(ctx context.Context, memberID int64, f repository.MutasiFilter) ([]repository.MutasiRow, int64, error) {
	if memberID <= 0 {
		return nil, 0, errors.New("unauthorized")
	}
	items, err := s.repo.ListMutasi(ctx, memberID, f)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.repo.CountMutasi(ctx, memberID, f)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (s *HistoryService) GetMutasiByID(ctx context.Context, memberID, id int64) (*repository.MutasiRow, error) {
	if memberID <= 0 {
		return nil, errors.New("unauthorized")
	}
	if id <= 0 {
		return nil, errors.New("id required")
	}
	return s.repo.GetMutasiByID(ctx, memberID, id)
}

func (s *HistoryService) ListTransaksi(ctx context.Context, memberID int64, limit, offset int, search, status, fromStr, toStr string) ([]repository.TrxMemberRow, int64, error) {
	if memberID <= 0 {
		return nil, 0, errors.New("unauthorized")
	}
	items, err := s.repo.ListTransaksi(ctx, memberID, limit, offset, search, status, fromStr, toStr)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.repo.CountTransaksi(ctx, memberID, search, status, fromStr, toStr)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (s *HistoryService) AdminListMutasiByMember(
	ctx context.Context,
	memberID int64,
	limit, offset int,
	refID, arah, dateStr, fromStr, toStr string,
	walletOnly bool,
) ([]repository.AdminMutasiRow, int64, error) {
	items, err := s.repo.AdminListMutasiByMember(ctx, memberID, limit, offset, refID, arah, dateStr, fromStr, toStr, walletOnly)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.repo.AdminCountMutasiByMember(ctx, memberID, refID, arah, dateStr, fromStr, toStr, walletOnly)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (s *HistoryService) AdminListTransaksi(
	ctx context.Context,
	memberID int64,
	limit, offset int,
	search, status, kodeProduk, refID, dest, fromStr, toStr, memberRole string,
) ([]repository.AdminTrxRow, int64, error) {
	if memberID > 0 {
		items, err := s.repo.AdminListTransaksiByMember(ctx, memberID, limit, offset, search, status, kodeProduk, fromStr, toStr)
		if err != nil {
			return nil, 0, err
		}
		total, err := s.repo.AdminCountTransaksiByMember(ctx, memberID, search, status, kodeProduk, fromStr, toStr)
		if err != nil {
			return nil, 0, err
		}
		return items, total, nil
	}

	items, err := s.repo.AdminListTransaksiAll(ctx, limit, offset, search, status, kodeProduk, refID, dest, fromStr, toStr, memberRole)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.repo.AdminCountTransaksiAll(ctx, search, status, kodeProduk, refID, dest, fromStr, toStr, memberRole)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}
