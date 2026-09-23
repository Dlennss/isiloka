package service

import (
	"context"
	"errors"
	"math"
	"strings"

	"pulsa2/internal/repository"
)

func (s *HistoryService) AdminCancelPendingTransaksiBulk(ctx context.Context, adminID int64, trxIDs []int64, reason string, allowSuccessCancel bool) (resolved []*repository.AdminCancelTrxResult, failed []map[string]any, err error) {
	if adminID <= 0 {
		return nil, nil, errors.New("admin only")
	}
	if len(trxIDs) == 0 {
		return nil, nil, errors.New("trx_ids required")
	}
	if strings.TrimSpace(reason) == "" {
		return nil, nil, errors.New("reason required")
	}
	for _, trxID := range trxIDs {
		item, callErr := s.repo.AdminCancelPendingTransaksi(ctx, adminID, trxID, reason, allowSuccessCancel)
		if callErr != nil {
			failed = append(failed, map[string]any{
				"trx_id": trxID,
				"error":  callErr.Error(),
			})
			continue
		}
		resolved = append(resolved, item)
	}
	return resolved, failed, nil
}

func (s *HistoryService) AdminCompletePendingTransaksi(ctx context.Context, adminID, trxID int64, reason string) (*repository.AdminCompleteTrxResult, error) {
	if adminID <= 0 {
		return nil, errors.New("admin only")
	}
	if trxID <= 0 {
		return nil, errors.New("trx_id required")
	}
	return s.repo.AdminCompletePendingTransaksi(ctx, adminID, trxID, reason)
}

func (s *HistoryService) AdminCompletePendingTransaksiBulk(ctx context.Context, adminID int64, trxIDs []int64, reason string) (resolved []*repository.AdminCompleteTrxResult, failed []map[string]any, err error) {
	if adminID <= 0 {
		return nil, nil, errors.New("admin only")
	}
	if len(trxIDs) == 0 {
		return nil, nil, errors.New("trx_ids required")
	}
	for _, trxID := range trxIDs {
		item, callErr := s.repo.AdminCompletePendingTransaksi(ctx, adminID, trxID, reason)
		if callErr != nil {
			failed = append(failed, map[string]any{
				"trx_id": trxID,
				"error":  callErr.Error(),
			})
			continue
		}
		resolved = append(resolved, item)
	}
	return resolved, failed, nil
}

func (s *HistoryService) AdminListTransaksiStatusLogsManual(ctx context.Context, f repository.TrxMemberStatusLogFilter) ([]repository.TrxMemberStatusLogRow, int64, int, error) {
	f.ManualOnly = true
	return s.AdminListTransaksiStatusLogs(ctx, f)
}

func (s *HistoryService) AdminListTransaksiStatusLogs(ctx context.Context, f repository.TrxMemberStatusLogFilter) ([]repository.TrxMemberStatusLogRow, int64, int, error) {
	items, err := s.repo.ListTrxMemberStatusLogs(ctx, f)
	if err != nil {
		return nil, 0, 0, err
	}
	total, err := s.repo.CountTrxMemberStatusLogs(ctx, f)
	if err != nil {
		return nil, 0, 0, err
	}

	limit := f.Limit
	if limit <= 0 {
		limit = 10
	}
	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(limit)))
	}
	return items, total, totalPages, nil
}
