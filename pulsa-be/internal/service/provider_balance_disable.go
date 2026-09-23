package service

import (
	"context"
	"strings"

	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
)

func providerBalanceDisableReason(providerName, msg string) string {
	providerName = strings.ToLower(strings.TrimSpace(providerName))
	msg = strings.Join(strings.Fields(strings.TrimSpace(msg)), " ")
	if msg == "" {
		return "auto nonaktif: saldo provider tidak cukup"
	}
	if providerName == "" {
		return "auto nonaktif: " + msg
	}
	return "auto nonaktif " + providerName + ": " + msg
}

func (h *MemberTrxService) disableProviderForOutOfBalance(ctx context.Context, providerName, msg string) {
	if h == nil || h.MemberRepo == nil || !helper.LooksLikeProviderOutOfBalance(msg) {
		return
	}
	changed, err := repository.DisableProviderForOutOfBalance(ctx, h.MemberRepo.DB(), providerName, providerBalanceDisableReason(providerName, msg))
	if err != nil {
		h.logf("PROVIDER_AUTO_DISABLE gagal provider=%s reason=out_of_balance err=%v", providerName, err)
		return
	}
	if changed > 0 {
		h.logf("PROVIDER_AUTO_DISABLE provider=%s changed_rows=%d reason=out_of_balance", providerName, changed)
	}
}

func (s *ProviderCallbackService) disableProviderForOutOfBalance(ctx context.Context, providerName, msg string) {
	if s == nil || s.repo == nil || !helper.LooksLikeProviderOutOfBalance(msg) {
		return
	}
	changed, err := repository.DisableProviderForOutOfBalance(ctx, s.repo.DB(), providerName, providerBalanceDisableReason(providerName, msg))
	if err != nil {
		helper.AppendProviderServiceLog("provider_callback_error.log", "provider auto-disable gagal provider=%s reason=out_of_balance err=%v", providerName, err)
		return
	}
	if changed > 0 {
		helper.AppendProviderServiceLog("provider_callback_service.log", "provider auto-disabled provider=%s changed_rows=%d reason=out_of_balance", providerName, changed)
	}
}
