package service

import (
	"strings"

	"pulsa2/internal/helper"
	"pulsa2/minions"
	"pulsa2/sagaramobile"
)

func hasProviderFailureSignal(msg string) bool {
	up := strings.ToUpper(strings.TrimSpace(msg))
	if up == "" {
		return false
	}
	return strings.Contains(up, "GAGAL") ||
		strings.Contains(up, "FAILED") ||
		strings.Contains(up, "ERROR") ||
		strings.Contains(up, "TIDAK DIPROSES") ||
		strings.Contains(up, "CUTOFF") ||
		strings.Contains(up, "DIBATALKAN") ||
		strings.Contains(up, "BATAL") ||
		strings.Contains(up, "NOMOR TUJUAN SALAH") ||
		strings.Contains(up, "STOK DIKEMBALIKAN") ||
		strings.Contains(up, "ALLOWED QTY") ||
		strings.Contains(up, "QTY TIDAK SESUAI") ||
		strings.Contains(up, "TIMEOUT") ||
		strings.Contains(up, "TIDAK VALID")
}

func hasProviderSuccessSignal(msg string) bool {
	up := strings.ToUpper(strings.TrimSpace(msg))
	return strings.Contains(up, "SUKSES") || strings.Contains(up, "BERHASIL") || strings.Contains(up, "LUNAS")
}

func hasProviderPendingSignal(msg string) bool {
	up := strings.ToUpper(strings.TrimSpace(msg))
	return strings.Contains(up, "PROSES") || strings.Contains(up, "MENUNGGU") || strings.Contains(up, "ANTRI")
}

// resolveCallbackFinalStatus: pesan duluan, RC hanya fallback terakhir
func resolveCallbackFinalStatus(msg string) string {
	if hasProviderSuccessSignal(msg) && !hasProviderFailureSignal(msg) {
		return "success"
	}
	if hasProviderFailureSignal(msg) {
		return "failed"
	}
	if hasProviderPendingSignal(msg) {
		return "pending"
	}
	return "pending" // pesan tidak dikenali dari callback = tunggu status/callback berikutnya
}

func resolveYuscomFinalStatus(_ int, _, msg string) string {
	return resolveCallbackFinalStatus(msg)
}

func resolveMultikomFinalStatus(_ int, _, msg string) string {
	if helper.LooksLikeMultikomSuccess(msg) {
		return "success"
	}
	return resolveCallbackFinalStatus(msg)
}

func resolveSagaraFinalStatus(_ int, _, msg string) string {
	if sagaramobile.LooksLikeSuccess(msg) {
		return "success"
	}
	if sagaramobile.LooksLikeImmediateReject(msg) || hasProviderFailureSignal(msg) {
		return "failed"
	}
	if sagaramobile.LooksLikePending(msg) {
		return "pending"
	}
	return "pending"
}

func resolveMinionsFinalStatus(_ int, _, msg string) string {
	if minions.LooksLikeSuccess(msg) {
		return "success"
	}
	if minions.LooksLikeImmediateReject(msg) || hasProviderFailureSignal(msg) {
		return "failed"
	}
	if minions.LooksLikePending(msg) {
		return "pending"
	}
	return "pending"
}

func resolveTrionikFinalStatus(_ int, _, msg string) string {
	return resolveCallbackFinalStatus(msg)
}

func resolveAJSFinalStatus(_ int, _, msg string) string {
	if helper.LooksLikeAJSSuccess(msg) {
		return "success"
	}
	if helper.LooksLikeAJSImmediateReject(msg) || hasProviderFailureSignal(msg) {
		return "failed"
	}
	if helper.LooksLikeAJSAccepted(msg) {
		return "pending"
	}
	return "pending"
}

func resolveGemilangFinalStatus(_ int, _, msg string) string {
	if helper.LooksLikeGemilangSuccess(msg) {
		return "success"
	}
	return resolveCallbackFinalStatus(msg)
}

func resolveLoketBayarFinalStatus(rcStr, msg string) string {
	switch helper.ProviderResponseStateOf("loketbayar", rcStr, msg) {
	case helper.ProviderResponseSuccess:
		return "success"
	case helper.ProviderResponseFailed:
		return "failed"
	default:
		return "pending"
	}
}

func fallbackAcceptedForProvider(provider, body string) bool {
	return helper.ProviderResponseAccepted(provider, body)
}

func fallbackImmediateRejectForProvider(provider, body string) bool {
	return helper.ProviderResponseImmediateReject(provider, body)
}

func fallbackFinalFailureMessage(baseMsg string, fallbackErr error) string {
	if fallbackErr == nil {
		return strings.TrimSpace(baseMsg)
	}
	msg := strings.TrimSpace(fallbackErr.Error())
	for _, prefix := range []string{
		"yuscom reject bisnis:", "talenta reject bisnis:", "multikom reject bisnis:",
		"minions reject bisnis:", "trionik reject bisnis:", "ajs reject bisnis:",
		"smb reject bisnis:", "gemilang reject bisnis:",
		"respons yuscom tidak dikenali:", "respons talenta tidak dikenali:",
		"respons multikom tidak dikenali:", "respons minions tidak dikenali:",
		"respons trionik tidak dikenali:", "respons ajs tidak dikenali:",
		"respons smb tidak dikenali:", "respons gemilang tidak dikenali:",
	} {
		if strings.HasPrefix(strings.ToLower(msg), prefix) {
			msg = strings.TrimSpace(msg[len(prefix):])
			break
		}
	}
	if msg != "" {
		return msg
	}
	return strings.TrimSpace(baseMsg)
}
