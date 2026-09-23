package service

import (
	"context"
	"strings"

	"pulsa2/internal/helper"
)

func (s *ProviderCallbackService) fallbackEnoughSaldo(ctx context.Context, provider string, need int64) bool {
	if need <= 0 {
		return false
	}
	bal, err := s.repo.GetProviderSaldo(ctx, provider)
	if err != nil {
		return false
	}
	return bal >= need
}

func (s *ProviderCallbackService) fallbackNeed(ctx context.Context, provider, product string, reqQty int64, produkProviderMapID *int64, kodeProvider string) (int64, int64, string) {
	need := reqQty
	fee := int64(0)
	source := "qty_provider"
	if reqQty <= 0 || s == nil || s.repo == nil {
		return need, fee, source
	}

	if strings.EqualFold(strings.TrimSpace(provider), "minions") {
		if pf := minionsFeeByCodeForCallback(kodeProvider); pf > 0 {
			need = reqQty + pf
			fee = pf
			source = "minions_code_fee"
		}
		return need, fee, source
	}

	if pf, src, err := s.repo.GetProviderFeeByProduct(ctx, provider, product, produkProviderMapID, kodeProvider); err == nil && pf > 0 {
		need = reqQty + pf
		fee = pf
		source = src
	} else if err != nil {
		helper.AppendProviderServiceLog("provider_callback_service.log", "fallback fee lookup gagal provider=%s produk=%s qty=%d err=%v", provider, product, reqQty, err)
	}
	return need, fee, source
}
