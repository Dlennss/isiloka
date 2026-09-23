package service

import (
	"context"

	"pulsa2/internal/repository"
)

func (s *ProviderCallbackService) tryFallbackFromJavapay(ctx context.Context, trx *repository.CallbackTrxMemberFull, javapayMsg string) (bool, string, int64, error) {
	return s.tryFallbackFromProvider(ctx, trx, "javapay", "javapay_callback_fallback", javapayMsg)
}

func (s *ProviderCallbackService) tryFallbackFromTalenta(ctx context.Context, trx *repository.CallbackTrxMemberFull, talentaMsg string) (bool, string, int64, error) {
	return s.tryFallbackFromProvider(ctx, trx, "talentapay", "talenta_callback_fallback", talentaMsg)
}

func (s *ProviderCallbackService) tryFallbackFromMultikom(ctx context.Context, trx *repository.CallbackTrxMemberFull, multikomMsg string) (bool, string, int64, error) {
	return s.tryFallbackFromProvider(ctx, trx, "multikom", "multikom_callback_fallback", multikomMsg)
}
