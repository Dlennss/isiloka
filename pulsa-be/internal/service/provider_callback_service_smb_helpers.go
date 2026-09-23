package service

import (
	"context"
	"net/url"
	"strings"

	"pulsa2/internal/repository"
	"pulsa2/smb"
)

func normalizeSMBRouteCodeHint(code string) string {
	code = strings.ToUpper(strings.TrimSpace(code))
	if i := strings.Index(code, ":"); i > 0 {
		code = strings.TrimSpace(code[:i])
	}
	return code
}

func (s *ProviderCallbackService) getSMBCallbackRow(ctx context.Context, refid string, routeCodeHint string) (*repository.ProviderTrxRefRow, error) {
	routeCodeHint = normalizeSMBRouteCodeHint(routeCodeHint)
	if routeCodeHint != "" {
		if row, err := s.repo.GetLatestByRefIDProviderCodeHint(ctx, refid, "smb", routeCodeHint); err != nil {
			return nil, err
		} else if row != nil {
			return row, nil
		}
	}
	return s.repo.GetLatestByRefIDProvider(ctx, refid, "smb")
}

func smbCallbackStage(rawRefID string) string {
	rawRefID = strings.ToUpper(strings.TrimSpace(rawRefID))
	switch {
	case strings.HasPrefix(rawRefID, "CEK"):
		return "check"
	case strings.HasPrefix(rawRefID, "BYR"):
		return "pay"
	default:
		return ""
	}
}

func smbFailedAfterPriorSuccessMessage(msg string) bool {
	return strings.Contains(strings.ToUpper(strings.TrimSpace(msg)), "SEMPAT SUKSES")
}

func smbFailureAllowsDowngradeOrFallback(isBank bool, routeMode smb.Mode, routeCode, msg string) bool {
	if isBank {
		return true
	}
	if routeMode == smb.ModeDirect && strings.EqualFold(strings.TrimSpace(routeCode), "BIFASTOPEN") {
		return true
	}
	return smbFailedAfterPriorSuccessMessage(msg)
}

func smbCallbackValue(q url.Values, keys ...string) string {
	for _, key := range keys {
		if v := strings.TrimSpace(q.Get(key)); v != "" {
			return v
		}
	}
	return ""
}

func smbCheckCallbackAlreadyPromoted(row *repository.ProviderTrxRefRow) bool {
	if row == nil {
		return false
	}

	msg := ""
	if row.Pesan != nil {
		msg = strings.ToUpper(strings.TrimSpace(*row.Pesan))
	}

	if strings.Contains(msg, "PAYSUKSES") || strings.Contains(msg, "PAYBERHASIL") ||
		strings.Contains(msg, "UNDER PROSES") || strings.Contains(msg, "MENUNGGU") ||
		strings.HasPrefix(msg, "BYR") {
		return true
	}
	return false
}

func smbCheckCallbackAttemptClosed(row *repository.ProviderTrxRefRow) bool {
	if row == nil {
		return false
	}
	if strings.EqualFold(strings.TrimSpace(row.Status), "failed") {
		return true
	}
	msg := ""
	if row.Pesan != nil {
		msg = strings.ToUpper(strings.TrimSpace(*row.Pesan))
	}
	return strings.Contains(msg, "SMB GAGAL") ||
		strings.Contains(msg, "CALLBACK CEK TIDAK DITERIMA")
}

func smbCheckCallbackReadyToPay(msg string) bool {
	up := strings.ToUpper(strings.TrimSpace(msg))
	if up == "" {
		return false
	}
	return strings.Contains(up, "INQSUKSES") || strings.Contains(up, "CEK WITHDRAWAL BERHASIL")
}

func smbEffectiveWalletPrice(callbackPrice int64, row *repository.ProviderTrxRefRow) int64 {
	if callbackPrice > 0 {
		return callbackPrice
	}
	if row != nil && row.Harga != nil && *row.Harga > 0 {
		return *row.Harga
	}
	return 0
}

func smbPersistedMessage(_ string, msg string) string {
	return strings.TrimSpace(msg)
}

func (s *ProviderCallbackService) tryFallbackFromSMB(ctx context.Context, row *repository.ProviderTrxRefRow, trx *repository.CallbackTrxMemberFull, smbMsg string) (bool, string, int64, error) {
	return s.tryFallbackFromSMBBankChain(ctx, row, trx, smbMsg)
}

func smbPayAlreadyProcessed(row *repository.ProviderTrxRefRow) bool {
	if row == nil {
		return false
	}
	if strings.EqualFold(strings.TrimSpace(row.Status), "success") {
		return true
	}
	msg := ""
	if row.Pesan != nil {
		msg = strings.ToUpper(strings.TrimSpace(*row.Pesan))
	}
	if strings.Contains(msg, "TRANSAKSI WITHDRAWAL BERHASIL") ||
		strings.Contains(msg, "PAYSUKSES") ||
		strings.Contains(msg, "PAYBERHASIL") {
		return true
	}
	return false
}
