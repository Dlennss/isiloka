package service

import (
	"context"
	"errors"
	"strings"
)

func (s *HistoryService) resolveProviderResendTarget(ctx context.Context, providerTrxID, trxID int64) (int64, error) {
	if providerTrxID > 0 {
		return providerTrxID, nil
	}
	if trxID <= 0 {
		return 0, errors.New("provider_trx_id required")
	}
	if s.repo == nil || s.trxSvc == nil || s.trxSvc.JPRepo == nil {
		return 0, errors.New("trx service belum siap")
	}

	trx, err := s.repo.GetAdminTrxCallbackTarget(ctx, trxID)
	if err != nil {
		return 0, err
	}
	if trx == nil {
		return 0, errors.New("transaksi tidak ditemukan")
	}
	rows, err := s.trxSvc.JPRepo.ListByRefID(ctx, trx.RefID)
	if err != nil {
		return 0, err
	}
	for _, row := range rows {
		if row == nil {
			continue
		}
		if row.TransaksiMemberID != trx.ID {
			continue
		}
		if !strings.EqualFold(strings.TrimSpace(row.Perintah), "PAY") {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(row.Status), "pending") {
			return row.ID, nil
		}
	}
	return 0, errors.New("provider pending row tidak ditemukan")
}

func (s *HistoryService) AdminResendPendingTransaksi(ctx context.Context, adminID, providerTrxID, trxID int64) (map[string]any, error) {
	if adminID <= 0 {
		return nil, errors.New("admin only")
	}
	if s.trxSvc == nil {
		return nil, errors.New("trx service belum siap")
	}

	resolvedProviderTrxID, err := s.resolveProviderResendTarget(ctx, providerTrxID, trxID)
	if err != nil {
		return nil, err
	}

	provider, newProviderRowID, err := s.trxSvc.ResendPendingProviderRow(ctx, resolvedProviderTrxID)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"provider_trx_id":     resolvedProviderTrxID,
		"status":              "pending",
		"provider":            provider,
		"new_provider_trx_id": newProviderRowID,
		"processed_by":        adminID,
	}, nil
}

func (s *HistoryService) AdminResendRefundNoSuccessTransaksi(ctx context.Context, adminID, providerTrxID int64) (map[string]any, error) {
	if adminID <= 0 {
		return nil, errors.New("admin only")
	}
	if providerTrxID <= 0 {
		return nil, errors.New("provider_trx_id required")
	}
	if s.trxSvc == nil {
		return nil, errors.New("trx service belum siap")
	}

	provider, newProviderRowID, err := s.trxSvc.ResendRefundProviderRowNoSuccess(ctx, providerTrxID)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"provider_trx_id":     providerTrxID,
		"status":              "pending",
		"provider":            provider,
		"new_provider_trx_id": newProviderRowID,
		"processed_by":        adminID,
		"mode":                "refund_no_success",
	}, nil
}
