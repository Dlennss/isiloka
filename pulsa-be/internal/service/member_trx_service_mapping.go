package service

import (
	"context"
	"strings"

	"pulsa2/internal/helper"
)

func (h *MemberTrxService) resolveMappedProductForProvider(ctx context.Context, provider string, internalProduct string) (string, bool) {
	p := strings.TrimSpace(strings.ToUpper(internalProduct))
	provider = strings.ToLower(strings.TrimSpace(provider))
	if p == "" {
		return "", false
	}
	if h.MemberRepo == nil {
		return "", false
	}
	mapped, err := h.MemberRepo.ResolveProviderProductCode(ctx, provider, p)
	if err != nil {
		h.logf("gagal lookup mapping ketat provider=%s internal=%s err=%v", provider, p, err)
		return "", false
	}
	mapped = strings.TrimSpace(strings.ToUpper(mapped))
	if mapped == "" {
		return "", false
	}
	return mapped, true
}

func (h *MemberTrxService) logf(format string, args ...any) {
	helper.AppendMemberTrxLog(format, args...)
}

func isTalentaProductCode(code string) bool {
	return strings.HasPrefix(strings.ToUpper(strings.TrimSpace(code)), "LINKAJA")
}

func minionsFeeByCode(code string) int64 {
	switch strings.ToUpper(strings.TrimSpace(code)) {
	case "HDANA":
		return 55
	case "HDANAB":
		return 100
	case "HGOPAY":
		return 975
	case "HLINK":
		return 625
	case "HLINKB":
		return 700
	case "HOVO":
		return 580
	case "HSHOPEE":
		return 300
	case "HSHOPEEB":
		return 500
	default:
		return 0
	}
}

func (h *MemberTrxService) computeProviderNeed(ctx context.Context, provider string, product string, billingNominal int64, produkProviderMapID *int64, kodeProvider string) (need int64, fee int64, source string) {
	need = billingNominal
	fee = 0
	source = "billing_nominal"

	if h.MemberRepo == nil {
		return need, fee, source
	}

	if strings.EqualFold(strings.TrimSpace(provider), "minions") {
		if pf := minionsFeeByCode(kodeProvider); pf > 0 {
			fee = pf
			need = billingNominal + pf
			source = "minions_code_fee"
		}
		return need, fee, source
	}

	pf, src, err := h.MemberRepo.GetProviderFeeByProduct(ctx, provider, product, produkProviderMapID, kodeProvider)
	if err != nil {
		h.logf("gagal lookup fee provider provider=%s produk=%s err=%v", provider, product, err)
		return need, fee, source
	}
	if src != "" && pf > 0 {
		fee = pf
		need = billingNominal + pf
		source = src
	}
	return need, fee, source
}

func effectiveMemberSellingPrice(hargaMember, biayaPerkiraan int64) int64 {
	if hargaMember > 0 {
		return hargaMember
	}
	if biayaPerkiraan > 0 {
		return biayaPerkiraan
	}
	return 0
}
