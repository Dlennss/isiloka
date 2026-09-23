package service

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"pulsa2/internal/repository"
	"pulsa2/model"
)

func isStaleLoketBayarPendingAttempt(row repository.ProviderAttemptRow) bool {
	if !strings.EqualFold(strings.TrimSpace(row.Provider), "loketbayar") {
		return false
	}
	if !strings.EqualFold(strings.TrimSpace(row.Status), "pending") {
		return false
	}
	if row.HTTPStatus == 200 {
		return false
	}
	if row.DibuatPada.IsZero() {
		return false
	}
	return time.Since(row.DibuatPada) >= loketBayarRetryMaxWindow
}

func callbackAttemptKey(provider string, mapID *int64, kodeProduk string) string {
	provider = strings.ToLower(strings.TrimSpace(provider))
	if mapID != nil && *mapID > 0 {
		return fmt.Sprintf("%s#map:%d", provider, *mapID)
	}
	return fmt.Sprintf("%s#code:%s", provider, strings.ToUpper(strings.TrimSpace(kodeProduk)))
}

func providerAttemptRefKey(row repository.ProviderAttemptRow) string {
	return callbackAttemptKey(row.Provider, row.ProdukProviderMapID, row.KodeProduk)
}

func keepCallbackFallbackCandidateOrder(in []callbackFallbackCandidate) []callbackFallbackCandidate {
	return in
}

func bankFallbackRank(failedProvider string, providerName string) int {
	failedProvider = strings.ToLower(strings.TrimSpace(failedProvider))
	providerName = strings.ToLower(strings.TrimSpace(providerName))
	switch failedProvider {
	case "smb":
		switch providerName {
		case "rajabiller":
			return 0
		case "loketbayar":
			return 1
		default:
			return -1
		}
	case "rajabiller":
		switch providerName {
		case "smb":
			return 0
		case "loketbayar":
			return 1
		default:
			return -1
		}
	default:
		switch providerName {
		case "smb", "rajabiller":
			return 0
		case "loketbayar":
			return 1
		default:
			return -1
		}
	}
}

func splitCallbackAttemptExclusions(skipCandidates map[string]bool) ([]int64, []string) {
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
		codeKeys = append(codeKeys, key)
	}
	return mapIDs, codeKeys
}

func (s *ProviderCallbackService) listTriedFallbackKeys(ctx context.Context, refID string) (map[string]bool, error) {
	out := map[string]bool{}
	if s == nil || s.repo == nil {
		return out, nil
	}
	rows, err := s.repo.ListAttemptsByRefID(ctx, refID)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		if isStaleLoketBayarPendingAttempt(row) {
			continue
		}
		out[providerAttemptRefKey(row)] = true
	}
	return out, nil
}

func (s *ProviderCallbackService) buildFallbackCandidates(ctx context.Context, trx *repository.CallbackTrxMemberFull, failedProvider string, skipCandidates map[string]bool) ([]callbackFallbackCandidate, error) {
	if trx == nil || s.repo == nil {
		return nil, nil
	}
	reqQty := trx.QtyProvider
	if reqQty <= 0 {
		reqQty = trx.Qty
	}
	if reqQty <= 0 {
		return nil, fmt.Errorf("qty provider invalid")
	}
	failedProvider = strings.ToLower(strings.TrimSpace(failedProvider))
	mapIDs, codeKeys := splitCallbackAttemptExclusions(skipCandidates)
	candidates, err := repository.ListRouteCandidates(ctx, s.repo.DB(), trx.KodeProduk, reqQty, mapIDs, codeKeys, true)
	if err != nil {
		return nil, err
	}
	isBank := s.isBankH2HProduct(ctx, trx.KodeProduk)
	type rankedFallbackCandidate struct {
		candidate callbackFallbackCandidate
		rank      int
		order     int
	}
	ranked := make([]rankedFallbackCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		providerName := strings.ToLower(strings.TrimSpace(candidate.Provider))
		rank := 0
		if isBank {
			rank = bankFallbackRank(failedProvider, providerName)
			if rank < 0 {
				continue
			}
		}
		need, fee, src := s.fallbackNeed(ctx, providerName, trx.KodeProduk, reqQty, candidate.ProdukProviderMapID, candidate.KodeProvider)
		if !s.fallbackEnoughSaldo(ctx, providerName, need) {
			continue
		}
		ranked = append(ranked, rankedFallbackCandidate{
			rank:  rank,
			order: len(ranked),
			candidate: callbackFallbackCandidate{
				Provider:            providerName,
				ProdukSKUSnapshot:   candidate.ProdukSKUSnapshot,
				ProdukProviderMapID: candidate.ProdukProviderMapID,
				KodeProduk:          candidate.KodeProvider,
				SpecialCode:         candidate.SpecialCode,
				Mode:                candidate.Mode,
				Need:                need,
				Fee:                 fee,
				Source:              src,
			},
		})
	}
	if isBank {
		sort.SliceStable(ranked, func(i, j int) bool {
			if ranked[i].rank != ranked[j].rank {
				return ranked[i].rank < ranked[j].rank
			}
			return ranked[i].order < ranked[j].order
		})
	}
	out := make([]callbackFallbackCandidate, 0, len(ranked))
	for _, item := range ranked {
		out = append(out, item.candidate)
	}
	return keepCallbackFallbackCandidateOrder(out), nil
}

func hasOtherProviderAttempt(excludedProvider string, used map[string]bool) (bool, string) {
	excludedProvider = strings.ToLower(strings.TrimSpace(excludedProvider))
	for provider := range used {
		if provider != "" && provider != excludedProvider {
			return true, provider
		}
	}
	return false, ""
}

func (s *ProviderCallbackService) hasOtherProviderAttemptForRef(ctx context.Context, trx *repository.CallbackTrxMemberFull, excludedProvider string) (bool, string) {
	if trx == nil || s.repo == nil {
		return false, ""
	}

	rows, err := s.repo.ListAttemptsByRefID(ctx, trx.RefID)
	if err != nil {
		return false, ""
	}

	for _, row := range rows {
		provider := strings.ToLower(strings.TrimSpace(row.Provider))
		if provider == "" || provider == strings.ToLower(strings.TrimSpace(excludedProvider)) {
			continue
		}

		jRow := &model.JavapayTrxRow{
			ID: row.ID, Provider: row.Provider, ProdukProviderMapID: row.ProdukProviderMapID,
			KodeProduk: row.KodeProduk, KodeRespon: row.KodeRespon, Pesan: row.Pesan,
			NoReferensi: row.NoReferensi, Harga: row.Harga, Status: row.Status,
		}
		label := routeAttemptLabelFromProviderRow(jRow)

		if failed, _, _, _, _, _ := providerRowFailureState(provider, jRow); failed {
			continue
		}
		if ok, _, _, _, _ := providerRowSuccessState(provider, jRow); ok {
			return true, label
		}
		if ok, _, _ := providerRowPendingState(provider, jRow); ok {
			if isStaleLoketBayarPendingAttempt(row) {
				continue
			}
			return true, label
		}
		return true, label
	}
	return false, ""
}
