package service

import (
	"context"

	"pulsa2/internal/repository"
)

func (s *ProviderCallbackService) tryFallbackFromSMBBankChain(ctx context.Context, row *repository.ProviderTrxRefRow, trx *repository.CallbackTrxMemberFull, smbMsg string) (bool, string, int64, error) {
	if trx == nil || row == nil {
		return false, "", 0, nil
	}
	return s.tryFallbackFromProvider(ctx, trx, "smb", "smb_callback_fallback", smbMsg)
}
