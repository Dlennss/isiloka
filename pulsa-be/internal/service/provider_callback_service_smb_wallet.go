package service

import (
	"context"

	"pulsa2/internal/repository"
)

// Debit dompet provider di-handle oleh DB trigger (trg_provider_wallet_on_success).
// Fungsi ini dipertahankan supaya caller tidak break, tapi tidak melakukan apa-apa.

func (s *ProviderCallbackService) syncSMBWalletSuccess(_ context.Context, _ string, _ *repository.ProviderTrxRefRow, _ int64, _ string) bool {
	return true
}

func (s *ProviderCallbackService) syncJavapayWalletSuccess(_ context.Context, _ string, _ *repository.ProviderTrxRefRow, _ int64, _ string) bool {
	return true
}
