package service

import (
	"context"
	"strings"

	"pulsa2/internal/helper"
)

func rajabillerRouteLooksLikeBank(internalProduct, mode, kodeProduk, specialCode string) bool {
	internalProduct = strings.ToUpper(strings.TrimSpace(internalProduct))
	mode = strings.ToUpper(strings.TrimSpace(mode))
	kodeProduk = strings.ToUpper(strings.TrimSpace(kodeProduk))
	providerProduct := strings.ToUpper(buildProviderProduct(kodeProduk, specialCode))

	return strings.Contains(mode, "BANK") ||
		strings.Contains(mode, "TRANSFER") ||
		strings.Contains(mode, "TRF") ||
		strings.HasPrefix(kodeProduk, "BLTRF") ||
		strings.HasPrefix(providerProduct, "BLTRF") ||
		strings.Contains(providerProduct, ":BLTRF") ||
		strings.HasPrefix(internalProduct, "BANK")
}

func (h *MemberTrxService) rajabillerMerchantIDForAttempt(ctx context.Context, internalProduct string, attempt providerRouteAttempt, fallback string) string {
	fallback = strings.TrimSpace(fallback)
	if h == nil || h.ProviderMerchantIDs == nil || !isRajabillerProvider(attempt.Name) {
		return fallback
	}
	if !h.isBankH2HProduct(ctx, internalProduct) && !rajabillerRouteLooksLikeBank(internalProduct, attempt.Mode, attempt.KodeProduk, attempt.SpecialCode) {
		return fallback
	}
	merchantID, ok, err := h.ProviderMerchantIDs.RandomActive(ctx, "rajabiller")
	if err != nil {
		h.logf("RAJABILLER merchant_id random gagal ref_produk=%s err=%v", internalProduct, err)
		return fallback
	}
	if !ok {
		h.logf("RAJABILLER merchant_id aktif tidak ada ref_produk=%s fallback_env_or_request=%t", internalProduct, fallback != "")
		return fallback
	}
	return merchantID
}

func (s *ProviderCallbackService) rajabillerMerchantIDForFallback(ctx context.Context, internalProduct string, candidate callbackFallbackCandidate, fallback string) string {
	fallback = strings.TrimSpace(fallback)
	if s == nil || s.providerMerchantIDs == nil || !isRajabillerProvider(candidate.Provider) {
		return fallback
	}
	if !s.isBankH2HProduct(ctx, internalProduct) && !rajabillerRouteLooksLikeBank(internalProduct, ptrString(candidate.Mode), candidate.KodeProduk, ptrString(candidate.SpecialCode)) {
		return fallback
	}
	merchantID, ok, err := s.providerMerchantIDs.RandomActive(ctx, "rajabiller")
	if err != nil {
		helperLogProviderMerchantIDError(candidate.Provider, internalProduct, err)
		return fallback
	}
	if !ok {
		helperLogProviderMerchantIDMissing(candidate.Provider, internalProduct, fallback != "")
		return fallback
	}
	return merchantID
}

func helperLogProviderMerchantIDError(providerName, product string, err error) {
	// Kept as a tiny wrapper so provider_callback_service_core.go does not own this selection detail.
	helper.AppendProviderServiceLog("provider_callback_error.log", "RAJABILLER merchant_id random gagal provider=%s ref_produk=%s err=%v", providerName, product, err)
}

func helperLogProviderMerchantIDMissing(providerName, product string, hasFallback bool) {
	helper.AppendProviderServiceLog("provider_callback_service.log", "RAJABILLER merchant_id aktif tidak ada provider=%s ref_produk=%s fallback_env_or_request=%t", providerName, product, hasFallback)
}
