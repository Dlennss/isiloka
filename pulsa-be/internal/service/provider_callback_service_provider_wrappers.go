package service

import (
	"context"

	"pulsa2/internal/repository"
)

func (s *ProviderCallbackService) tryFallbackFromTrionik(ctx context.Context, trx *repository.CallbackTrxMemberFull, trionikMsg string) (bool, string, int64, error) {
	return s.tryFallbackFromProvider(ctx, trx, "trionik", "trionik_callback_fallback", trionikMsg)
}

func (s *ProviderCallbackService) tryFallbackFromAJS(ctx context.Context, trx *repository.CallbackTrxMemberFull, ajsMsg string) (bool, string, int64, error) {
	return s.tryFallbackFromProvider(ctx, trx, "ajs", "ajs_callback_fallback", ajsMsg)
}

func (s *ProviderCallbackService) tryFallbackFromGemilang(ctx context.Context, trx *repository.CallbackTrxMemberFull, gemilangMsg string) (bool, string, int64, error) {
	return s.tryFallbackFromProvider(ctx, trx, "gemilang", "gemilang_callback_fallback", gemilangMsg)
}

func (s *ProviderCallbackService) tryFallbackFromChytron(ctx context.Context, trx *repository.CallbackTrxMemberFull, chytronMsg string) (bool, string, int64, error) {
	return s.tryFallbackFromProvider(ctx, trx, "chytron", "chytron_callback_fallback", chytronMsg)
}
