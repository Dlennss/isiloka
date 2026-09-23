package service

import (
	"context"
	"fmt"
	"strings"

	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
	"pulsa2/smb"
)

func isBankH2HProduct(product string) bool {
	return helper.ClassifyH2HFeeCategory(product) == helper.H2HFeeCategoryBank
}

func (h *MemberTrxService) isBankH2HProduct(ctx context.Context, product string) bool {
	if h != nil && h.MemberRepo != nil {
		categoryName, err := h.MemberRepo.ClassifyH2HFeeCategoryBySKU(ctx, product)
		if err == nil && strings.TrimSpace(categoryName) != "" {
			return categoryName == helper.H2HFeeCategoryBank
		}
		if err != nil {
			h.logf("kategori produk fallback ke helper produk=%s err=%v", product, err)
		}
	}
	return isBankH2HProduct(product)
}

func (s *ProviderCallbackService) isBankH2HProduct(ctx context.Context, product string) bool {
	if s != nil && s.repo != nil {
		categoryName, err := repository.ClassifyH2HFeeCategoryBySKU(ctx, s.repo.DB(), product)
		if err == nil && strings.TrimSpace(categoryName) != "" {
			return categoryName == helper.H2HFeeCategoryBank
		}
		if err != nil {
			helper.AppendProviderServiceLog("provider_callback_error.log", "kategori produk fallback ke helper produk=%s err=%v", product, err)
		}
	}
	return isBankH2HProduct(product)
}

func isDigitsOnly(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func deriveBankCodeFromSMBRoute(mode string, kodeProduk string) (string, error) {
	_, code, bankPrefix, err := smb.ParseMappedCodeTargetWithMode(mode, kodeProduk)
	if err != nil {
		return "", err
	}
	if bankPrefix != "" {
		return strings.TrimSpace(bankPrefix), nil
	}
	if isDigitsOnly(code) {
		return strings.TrimSpace(code), nil
	}
	return "", fmt.Errorf("kode bank smb tidak ditemukan dari %s", strings.TrimSpace(kodeProduk))
}

func deriveBankCodeFromRajabillerRoute(kodeProduk string) (string, error) {
	kodeProduk = strings.ToUpper(strings.TrimSpace(kodeProduk))
	if kodeProduk == "" {
		return "", fmt.Errorf("kode bank rajabiller kosong")
	}
	if head, tail, ok := strings.Cut(kodeProduk, ":"); ok {
		if isDigitsOnly(head) {
			return strings.TrimSpace(head), nil
		}
		if isDigitsOnly(tail) {
			return strings.TrimSpace(tail), nil
		}
	}
	if isDigitsOnly(kodeProduk) {
		return kodeProduk, nil
	}
	return "", fmt.Errorf("kode bank rajabiller tidak ditemukan dari %s", kodeProduk)
}

func deriveBankCodeFromProviderRoute(providerName string, mode string, kodeProduk string) (string, error) {
	providerName = strings.ToLower(strings.TrimSpace(providerName))
	switch providerName {
	case "smb":
		return deriveBankCodeFromSMBRoute(mode, kodeProduk)
	case "rajabiller":
		return deriveBankCodeFromRajabillerRoute(kodeProduk)
	default:
		kodeProduk = strings.TrimSpace(kodeProduk)
		if isDigitsOnly(kodeProduk) {
			return kodeProduk, nil
		}
		return "", fmt.Errorf("kode bank %s tidak ditemukan dari %s", providerName, kodeProduk)
	}
}

func buildLoketBankDest(bankCode string, dest string) string {
	bankCode = strings.TrimSpace(bankCode)
	dest = strings.TrimSpace(dest)
	if bankCode == "" || dest == "" {
		return dest
	}
	if strings.HasPrefix(dest, bankCode) {
		return dest
	}
	return bankCode + dest
}

func currentSMBSpecial(attempt providerRouteAttempt) string {
	return strings.ToUpper(strings.TrimSpace(attempt.SpecialCode))
}

func normalizeBankPrimarySMBAttempt(attempt providerRouteAttempt) providerRouteAttempt {
	attempt.Mode = "DIRECT"
	if currentSMBSpecial(attempt) == "" {
		attempt.SpecialCode = "BIFASTOPEN"
	}
	return attempt
}

func expandBankProviderAttempts(attempts []providerRouteAttempt, loketEnabled bool) []providerRouteAttempt {
	if len(attempts) == 0 {
		return attempts
	}
	out := make([]providerRouteAttempt, 0, len(attempts))
	loket := make([]providerRouteAttempt, 0, 1)
	for _, attempt := range attempts {
		if strings.EqualFold(strings.TrimSpace(attempt.Name), "smb") {
			primary := normalizeBankPrimarySMBAttempt(attempt)
			out = append(out, primary)
			continue
		}
		if loketEnabled && strings.EqualFold(strings.TrimSpace(attempt.Name), "loketbayar") && isLoketBayarBankTransferProduct(currentSMBSpecial(attempt)) {
			loket = append(loket, attempt)
			continue
		}
		out = append(out, attempt)
	}
	out = append(out, loket...)
	return out
}

func latestBankCodeFromAttempts(rows []repository.ProviderAttemptRow) (string, error) {
	for _, row := range rows {
		providerName := strings.ToLower(strings.TrimSpace(row.Provider))
		if providerName != "smb" && providerName != "rajabiller" {
			continue
		}
		code, err := deriveBankCodeFromProviderRoute(providerName, "", row.KodeProduk)
		if err == nil && strings.TrimSpace(code) != "" {
			return code, nil
		}
	}
	return "", fmt.Errorf("attempt bank smb/rajabiller tidak ditemukan")
}
