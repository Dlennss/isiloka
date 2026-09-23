package service

import (
	"context"

	"pulsa2/internal/repository"
)

func (s *ProviderCallbackService) tryFallbackFromYuscom(ctx context.Context, trx *repository.CallbackTrxMemberFull, yuscomMsg string) (bool, string, int64, error) {
	return s.tryFallbackFromProvider(ctx, trx, "yuscom", "yuscom_callback_fallback", yuscomMsg)
}

func (s *ProviderCallbackService) tryFallbackFromSagara(ctx context.Context, trx *repository.CallbackTrxMemberFull, sagaraMsg string) (bool, string, int64, error) {
	return s.tryFallbackFromProvider(ctx, trx, "sagaramobile", "sagara_callback_fallback", sagaraMsg)
}

func (s *ProviderCallbackService) tryFallbackFromMinions(ctx context.Context, trx *repository.CallbackTrxMemberFull, minionsMsg string) (bool, string, int64, error) {
	return s.tryFallbackFromProvider(ctx, trx, "minions", "minions_callback_fallback", minionsMsg)
}
