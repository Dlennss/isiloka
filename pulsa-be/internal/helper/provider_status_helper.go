package helper

import (
	"encoding/json"
	"fmt"
	"strings"

	"pulsa2/minions"
	"pulsa2/sagaramobile"
	"pulsa2/smb"
)

type ProviderResponseState string

const (
	ProviderResponseUnknown ProviderResponseState = "unknown"
	ProviderResponseSuccess ProviderResponseState = "success"
	ProviderResponsePending ProviderResponseState = "pending"
	ProviderResponseFailed  ProviderResponseState = "failed"
)

func ExtractProviderStatusCodeFor(provider string, body string) string {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "sagaramobile":
		return strings.TrimSpace(sagaramobile.ExtractStatusCode(body))
	case "minions":
		return strings.TrimSpace(minions.ExtractStatusCode(body))
	case "smb", "chytron":
		return strings.TrimSpace(smb.ExtractStatusCode(body))
	default:
		return strings.TrimSpace(ExtractProviderStatusCode(body))
	}
}

func looksLikeSMBFailureAfterPriorSuccess(upper string) bool {
	upper = strings.ToUpper(strings.TrimSpace(upper))
	if !strings.Contains(upper, "SEMPAT SUKSES") {
		return false
	}
	return strings.Contains(upper, "GAGAL") ||
		strings.Contains(upper, "FAILED") ||
		strings.Contains(upper, "ERROR") ||
		strings.Contains(upper, "BATAL") ||
		strings.Contains(upper, "REFUND") ||
		strings.Contains(upper, "SALDO DIKEMBALIKAN") ||
		strings.Contains(upper, "DIKEMBALIKAN")
}

func pulsa24JamJSONField(body, key string) string {
	body = strings.TrimSpace(body)
	if body == "" || !strings.HasPrefix(body, "{") {
		return ""
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		return ""
	}
	for k, v := range payload {
		if strings.EqualFold(strings.TrimSpace(k), key) {
			return strings.TrimSpace(fmt.Sprint(v))
		}
	}
	return ""
}

func pulsa24JamResponseState(rc, msg string) ProviderResponseState {
	rc = strings.ToUpper(strings.TrimSpace(rc))
	msg = strings.TrimSpace(msg)
	upper := strings.ToUpper(msg)

	jsonStatus := strings.ToUpper(pulsa24JamJSONField(msg, "status"))
	jsonRC := strings.ToUpper(firstNonEmptyHelper(pulsa24JamJSONField(msg, "rc"), pulsa24JamJSONField(msg, "code")))
	jsonOK := strings.ToLower(pulsa24JamJSONField(msg, "ok"))
	jsonSuccess := strings.ToLower(pulsa24JamJSONField(msg, "success"))
	jsonMessage := strings.ToUpper(firstNonEmptyHelper(
		pulsa24JamJSONField(msg, "message"),
		pulsa24JamJSONField(msg, "msg"),
		pulsa24JamJSONField(msg, "keterangan"),
	))

	statusValue := firstNonEmptyHelper(jsonStatus, rc, jsonRC)
	messageValue := firstNonEmptyHelper(jsonMessage, upper)

	switch statusValue {
	case "2", "20", "00", "SUCCESS", "SUKSES":
		return ProviderResponseSuccess
	case "3", "52", "FAILED", "FAIL", "GAGAL", "ERROR":
		return ProviderResponseFailed
	case "1", "68", "0068", "PENDING", "PROCESS", "PROCESSING":
		return ProviderResponsePending
	}

	if jsonSuccess == "true" {
		return ProviderResponseSuccess
	}
	if jsonSuccess == "false" || jsonOK == "false" {
		return ProviderResponseFailed
	}

	switch {
	case strings.Contains(messageValue, "REFID TERLALU PANJANG") ||
		strings.Contains(messageValue, "NOMOR TUJUAN SALAH") ||
		strings.Contains(messageValue, "SALDO TIDAK CUKUP") ||
		strings.Contains(messageValue, "PRODUK TIDAK DITEMUKAN") ||
		strings.Contains(messageValue, "PRODUK KEHABISAN STOK") ||
		strings.Contains(messageValue, "GAGAL") ||
		strings.Contains(messageValue, "FAILED"):
		return ProviderResponseFailed
	case strings.Contains(messageValue, "SEDANG DIPROSES") ||
		strings.Contains(messageValue, "AKAN DIPROSES") ||
		strings.Contains(messageValue, "MENUNGGU") ||
		strings.Contains(messageValue, "PENDING"):
		return ProviderResponsePending
	case strings.Contains(messageValue, "TRANSAKSI BERHASIL") ||
		strings.Contains(messageValue, "SUKSES") ||
		strings.Contains(messageValue, "SUCCESS"):
		return ProviderResponseSuccess
	}

	// Pulsa24Jam/H2HR can return ok:true for an accepted request, not a final
	// delivery result. Keep it pending until status/SN/callback gives final proof.
	if jsonOK == "true" {
		return ProviderResponsePending
	}
	return ProviderResponsePending
}

func firstNonEmptyHelper(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

// ProviderResponseStateOf menentukan status transaksi berdasarkan PESAN,
// bukan RC. RC hanya dipakai sebagai fallback terakhir kalau pesan kosong.
// Prinsip: pesan dari provider adalah sumber kebenaran.
func ProviderResponseStateOf(provider, rc, msg string) ProviderResponseState {
	provider = strings.ToLower(strings.TrimSpace(provider))
	rc = strings.TrimSpace(rc)
	msg = strings.TrimSpace(msg)

	if rc == "" && msg == "" {
		return ProviderResponsePending
	}

	// Pesan internal sistem kita = sudah final
	upper := strings.ToUpper(msg)
	if strings.HasPrefix(upper, "PROVIDER ") || strings.HasPrefix(upper, "BUG ") ||
		strings.HasPrefix(upper, "HTTP ERROR") || strings.HasPrefix(upper, "SMB CEK CALLBACK") ||
		strings.HasPrefix(upper, "SMB GAGAL") {
		return ProviderResponseFailed
	}

	// Kalau RC kosong tapi pesan mengandung status=XX, extract RC dari pesan
	if rc == "" && msg != "" {
		if extracted := ExtractProviderStatusCodeFor(provider, msg); extracted != "" {
			rc = extracted
		}
	}

	switch provider {
	case "javapay":
		return ClassifyJavapayResponseStatus(rc, msg)
	case "talentapay":
		switch {
		case LooksLikeTalentaSuccess(msg) && !LooksLikeTalentaImmediateReject(msg):
			return ProviderResponseSuccess
		case LooksLikeTalentaImmediateReject(msg):
			return ProviderResponseFailed
		case LooksLikeTalentaAccepted(msg):
			return ProviderResponsePending
		}
	case "multikom":
		switch {
		case LooksLikeMultikomSuccess(msg) && !LooksLikeMultikomImmediateReject(msg):
			return ProviderResponseSuccess
		case LooksLikeMultikomImmediateReject(msg):
			return ProviderResponseFailed
		case LooksLikeMultikomAccepted(msg):
			return ProviderResponsePending
		}
	case "gemilang":
		switch {
		case LooksLikeGemilangSuccess(msg) && !LooksLikeGemilangImmediateReject(msg):
			return ProviderResponseSuccess
		case LooksLikeGemilangImmediateReject(msg):
			return ProviderResponseFailed
		case LooksLikeGemilangAccepted(msg):
			return ProviderResponsePending
		}
	case "yuscom":
		switch {
		case LooksLikeYuscomSuccess(msg) && !LooksLikeYuscomImmediateReject(msg):
			return ProviderResponseSuccess
		case LooksLikeYuscomImmediateReject(msg):
			return ProviderResponseFailed
		case LooksLikeYuscomAccepted(msg):
			return ProviderResponsePending
		}
	case "sagaramobile":
		switch {
		case sagaramobile.LooksLikeSuccess(msg) && !sagaramobile.LooksLikeImmediateReject(msg):
			return ProviderResponseSuccess
		case sagaramobile.LooksLikeImmediateReject(msg):
			return ProviderResponseFailed
		case sagaramobile.LooksLikePending(msg):
			return ProviderResponsePending
		}
	case "minions":
		switch {
		case minions.LooksLikeSuccess(msg) && !minions.LooksLikeImmediateReject(msg):
			return ProviderResponseSuccess
		case minions.LooksLikeImmediateReject(msg):
			return ProviderResponseFailed
		case minions.LooksLikePending(msg):
			return ProviderResponsePending
		}
	case "trionik":
		switch {
		case LooksLikeYuscomSuccess(msg) && !LooksLikeYuscomImmediateReject(msg):
			return ProviderResponseSuccess
		case LooksLikeYuscomImmediateReject(msg):
			return ProviderResponseFailed
		case LooksLikeYuscomAccepted(msg):
			return ProviderResponsePending
		}
	case "ajs":
		switch {
		case LooksLikeAJSSuccess(msg) && !LooksLikeAJSImmediateReject(msg):
			return ProviderResponseSuccess
		case LooksLikeAJSImmediateReject(msg):
			return ProviderResponseFailed
		case LooksLikeAJSAccepted(msg):
			return ProviderResponsePending
		}
	case "smb":
		switch {
		case looksLikeSMBFailureAfterPriorSuccess(upper):
			return ProviderResponseFailed
		case strings.Contains(upper, "INQSUKSES") || strings.Contains(upper, "CEK WITHDRAWAL"):
			return ProviderResponsePending
		case smb.LooksLikeSuccess(msg) && !smb.LooksLikeImmediateReject(msg):
			return ProviderResponseSuccess
		case smb.LooksLikeImmediateReject(msg):
			return ProviderResponseFailed
		case smb.LooksLikePending(msg):
			return ProviderResponsePending
		}
	case "chytron":
		switch {
		case smb.LooksLikeSuccess(msg) && !smb.LooksLikeImmediateReject(msg):
			return ProviderResponseSuccess
		case smb.LooksLikeImmediateReject(msg):
			return ProviderResponseFailed
		case smb.LooksLikePending(msg):
			return ProviderResponsePending
		}
	case "loketbayar":
		switch {
		case LooksLikeLoketBayarSuccess(rc, msg):
			return ProviderResponseSuccess
		case LooksLikeLoketBayarImmediateReject(rc, msg):
			return ProviderResponseFailed
		case LooksLikeLoketBayarPending(rc, msg):
			return ProviderResponsePending
		}
	case "rajabiller":
		switch {
		case LooksLikeRajabillerPending(rc, msg):
			return ProviderResponsePending
		case LooksLikeRajabillerSuccess(rc, msg):
			return ProviderResponseSuccess
		case LooksLikeRajabillerImmediateReject(rc, msg):
			return ProviderResponseFailed
		}
	case "pulsa24jam":
		return pulsa24JamResponseState(rc, msg)
	default:
		switch {
		case strings.Contains(upper, "ER_") || strings.Contains(upper, "SQLSTATE") || strings.Contains(upper, "SQLMESSAGE") || strings.Contains(upper, "DATA TOO LONG"):
			return ProviderResponseFailed
		case strings.Contains(upper, "SUKSES") || strings.Contains(upper, "BERHASIL"):
			return ProviderResponseSuccess
		case strings.Contains(upper, "GAGAL") || strings.Contains(upper, "FAILED") || strings.Contains(upper, "BATAL"):
			return ProviderResponseFailed
		}
	}

	if (provider == "smb" || provider == "chytron") && smb.LooksLikeAccepted(msg) {
		return ProviderResponsePending
	}

	// Fallback terakhir: classify dari pattern database (refresh otomatis)
	// Hati-hati: "INQSUKSES" bukan success (itu cek berhasil, bukan bayar)
	if MatchesSuccessPattern(upper) && !strings.Contains(upper, "INQSUKSES") && !strings.Contains(upper, "CEK WITHDRAWAL") {
		return ProviderResponseSuccess
	}
	if MatchesFailedPattern(upper) {
		return ProviderResponseFailed
	}
	// Pesan tidak dikenali → pending (tunggu callback)
	return ProviderResponsePending
}

func ProviderResponseStateFromBody(provider, body string) ProviderResponseState {
	return ProviderResponseStateOf(provider, ExtractProviderStatusCodeFor(provider, body), body)
}

func ProviderResponseStatusString(provider string, kodeRespon, pesan *string) string {
	rc := ""
	msg := ""
	if kodeRespon != nil {
		rc = strings.TrimSpace(*kodeRespon)
	}
	if pesan != nil {
		msg = strings.TrimSpace(*pesan)
	}
	return string(ProviderResponseStateOf(provider, rc, msg))
}

func ProviderResponseAccepted(provider, body string) bool {
	state := ProviderResponseStateFromBody(provider, body)
	return state == ProviderResponseSuccess || state == ProviderResponsePending
}

func ProviderResponseImmediateReject(provider, body string) bool {
	return ProviderResponseStateFromBody(provider, body) == ProviderResponseFailed
}
