package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	trxmemberdto "pulsa2/internal/dto/trx_member"
	"pulsa2/internal/repository"
	"pulsa2/model"
)

func providerSupportsOpenAmountRule(rule *repository.ProviderOpenAmountRule, billingNominal int64) bool {
	if billingNominal <= 0 {
		return false
	}
	if rule == nil {
		return true
	}
	if rule.MinimalNominal != nil && *rule.MinimalNominal > 0 && billingNominal < *rule.MinimalNominal {
		return false
	}
	if rule.MaksimalNominal != nil && *rule.MaksimalNominal > 0 && billingNominal > *rule.MaksimalNominal {
		return false
	}
	return true
}

func routeAttemptKey(provider string, mapID *int64, kodeProduk string) string {
	provider = strings.ToLower(strings.TrimSpace(provider))
	if mapID != nil && *mapID > 0 {
		return fmt.Sprintf("%s#map:%d", provider, *mapID)
	}
	return fmt.Sprintf("%s#code:%s", provider, strings.ToUpper(strings.TrimSpace(kodeProduk)))
}

func providerAttemptTriedKeys(rows []*model.JavapayTrxRow) map[string]bool {
	out := map[string]bool{}
	for _, row := range rows {
		if row == nil {
			continue
		}
		out[routeAttemptKey(row.Provider, row.ProdukProviderMapID, row.KodeProduk)] = true
	}
	return out
}

func routeAttemptKeyFromProviderRow(row *model.JavapayTrxRow) string {
	if row == nil {
		return ""
	}
	return routeAttemptKey(row.Provider, row.ProdukProviderMapID, row.KodeProduk)
}

func routeAttemptLabelFromProviderRow(row *model.JavapayTrxRow) string {
	if row == nil {
		return ""
	}
	key := routeAttemptKeyFromProviderRow(row)
	if key == "" {
		return strings.ToLower(strings.TrimSpace(row.Provider))
	}
	return key
}

func (h *MemberTrxService) allExistingRouteAttemptsFinalFailed(ctx context.Context, trxID int64, refID string) (bool, string) {
	if h == nil || h.JPRepo == nil {
		return false, "provider repo tidak siap"
	}
	refID = strings.TrimSpace(refID)
	if refID == "" {
		return false, "refid kosong"
	}
	rows, err := h.JPRepo.ListByRefID(ctx, refID)
	if err != nil {
		return false, fmt.Sprintf("gagal cek attempt refid=%s: %v", refID, err)
	}
	failedCount := 0
	for _, row := range rows {
		if row == nil {
			continue
		}
		if trxID > 0 && row.TransaksiMemberID > 0 && row.TransaksiMemberID != trxID {
			continue
		}
		provider := strings.ToLower(strings.TrimSpace(row.Provider))
		label := routeAttemptLabelFromProviderRow(row)
		if ok, _, _, _, _ := providerRowSuccessState(provider, row); ok {
			return false, fmt.Sprintf("mapping %s sudah success", label)
		}
		if ok, _, _ := providerRowPendingState(provider, row); ok {
			return false, fmt.Sprintf("mapping %s masih pending", label)
		}
		if !providerRowDefinitelyFailed(provider, row) {
			return false, fmt.Sprintf("mapping %s belum final", label)
		}
		failedCount++
	}
	if failedCount == 0 {
		return false, "belum ada mapping provider final failed"
	}
	return true, fmt.Sprintf("semua %d mapping yang dicoba sudah final failed", failedCount)
}

func splitRouteAttemptExclusions(skipCandidates map[string]bool) ([]int64, []string) {
	if len(skipCandidates) == 0 {
		return nil, nil
	}
	mapIDs := make([]int64, 0, len(skipCandidates))
	codeKeys := make([]string, 0, len(skipCandidates))
	for key, skip := range skipCandidates {
		if !skip {
			continue
		}
		key = strings.TrimSpace(strings.ToLower(key))
		if key == "" {
			continue
		}
		if strings.Contains(key, "#map:") {
			parts := strings.SplitN(key, "#map:", 2)
			if len(parts) == 2 {
				if id, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64); err == nil && id > 0 {
					mapIDs = append(mapIDs, id)
					continue
				}
			}
		}
		codeKeys = append(codeKeys, strings.ToLower(strings.TrimSpace(key)))
	}
	return mapIDs, codeKeys
}

func ptrString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func (h *MemberTrxService) buildProviderAttempts(
	ctx context.Context,
	in trxmemberdto.TrxRequest,
	billingNominal int64,
	withWalletCheck bool,
	skipCandidates map[string]bool,
) []providerRouteAttempt {
	excludeMapIDs, excludeCodeKeys := splitRouteAttemptExclusions(skipCandidates)
	isBank := h.isBankH2HProduct(ctx, in.Product)
	candidates, err := repository.ListRouteCandidates(ctx, h.MemberRepo.DB(), in.Product, billingNominal, excludeMapIDs, excludeCodeKeys, isBank)
	attempts := make([]providerRouteAttempt, 0, len(candidates))
	if err != nil {
		h.logf("ROUTING skip refid=%s produk=%s alasan=lookup_candidate_random_gagal err=%v", in.RefID, in.Product, err)
	}
	if len(candidates) == 0 && err == nil {
		h.logf("ROUTING info refid=%s produk=%s alasan=kandidat_db_tidak_ada", in.RefID, in.Product)
	}

	for _, candidate := range candidates {
		p := strings.ToLower(strings.TrimSpace(candidate.Provider))
		if p != "pulsa24jam" {
			h.logf("ROUTING skip candidate provider=%s refid=%s kode=%s alasan=only_pulsa24jam_enabled", p, in.RefID, candidate.KodeProvider)
			continue
		}
		need, fee, src := h.computeProviderNeed(ctx, p, in.Product, billingNominal, candidate.ProdukProviderMapID, candidate.KodeProvider)
		if withWalletCheck && h.ProviderWallet != nil {
			bal := h.providerAvailableBalance(ctx, p)
			if bal < need {
				h.logf("ROUTING skip candidate provider=%s refid=%s kode=%s alasan=saldo_kurang saldo=%d butuh=%d fee=%d sumber_fee=%s", p, in.RefID, candidate.KodeProvider, bal, need, fee, src)
				continue
			}
		}
		attempts = append(attempts, providerRouteAttempt{
			Name:                p,
			Need:                need,
			Fee:                 fee,
			Src:                 src,
			ProdukSKUSnapshot:   candidate.ProdukSKUSnapshot,
			ProdukProviderMapID: candidate.ProdukProviderMapID,
			KodeProduk:          candidate.KodeProvider,
			SpecialCode:         strings.ToUpper(strings.TrimSpace(ptrString(candidate.SpecialCode))),
			Mode:                strings.ToUpper(strings.TrimSpace(ptrString(candidate.Mode))),
		})
	}

	return attempts
}
