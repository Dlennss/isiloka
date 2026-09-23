package helper

import (
	"net"
	"net/http"
	"regexp"
	"strings"
)

var reProviderStatusCode = regexp.MustCompile(`(?i)\b(?:status|rc|code|kode_response|kode_respon)\s*[:=]\s*([0-9]{1,3})\b`)
var reLoketBayarSaldoOnlyResponse = regexp.MustCompile(`(?i)^\s*(?:status\s*=\s*)?(?:keterangan\s*=\s*)?\.?\s*SALDO\s*:\s*[0-9][0-9.,]*\s*$`)

func NormalizeInternalProductCode(product string) string {
	p := strings.TrimSpace(strings.ToUpper(product))
	switch p {
	case "DNID":
		return "DANA"
	default:
		return p
	}
}

func GetClientIP(r *http.Request) string {
	if v := strings.TrimSpace(r.Header.Get("CF-Connecting-IP")); v != "" {
		return v
	}
	if v := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); v != "" {
		parts := strings.Split(v, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil && host != "" {
		return host
	}
	return strings.TrimSpace(r.RemoteAddr)
}

func IsFinalFailRC(rc string) bool {
	switch strings.TrimSpace(rc) {
	case "", "0", "1", "2", "3", "20":
		return false
	default:
		return true
	}
}

func IsPendingRC(rc string) bool {
	switch strings.TrimSpace(rc) {
	case "0", "1", "2", "3":
		return true
	default:
		return false
	}
}

func IsSuccessRC(rc string) bool { return strings.TrimSpace(rc) == "20" }

func LooksLikeYuscomImmediateReject(body string) bool {
	up := strings.ToUpper(strings.TrimSpace(body))
	if up == "" {
		return false
	}
	if strings.Contains(up, "GAGAL") {
		return true
	}
	if strings.Contains(up, "FAILED") || strings.Contains(up, "ERROR") {
		return true
	}
	if strings.Contains(up, "TIDAK DIPROSES") || strings.Contains(up, "CUTOFF") {
		return true
	}
	if strings.Contains(up, "DIBATALKAN") || strings.Contains(up, "BATAL") {
		return true
	}
	if strings.Contains(up, "SALDO") && strings.Contains(up, "TIDAK") && strings.Contains(up, "CUKUP") {
		return true
	}
	if strings.Contains(up, "LIMIT 0") {
		return true
	}
	if strings.Contains(up, "ALLOWED QTY") || strings.Contains(up, "QTY TIDAK SESUAI") {
		return true
	}
	if strings.Contains(up, "NOMOR TUJUAN SALAH") || strings.Contains(up, "STOK DIKEMBALIKAN") {
		return true
	}
	return false
}

func LooksLikeYuscomAccepted(body string) bool {
	up := strings.ToUpper(strings.TrimSpace(body))
	if up == "" {
		return false
	}
	return strings.Contains(up, "AKAN DIPROSES")
}

func LooksLikeAJSAccepted(body string) bool {
	up := strings.ToUpper(strings.TrimSpace(body))
	if up == "" {
		return false
	}
	return strings.Contains(up, "AKAN DIPROSES") ||
		strings.Contains(up, "MENUNGGU JAWABAN")
}

func LooksLikeAJSSuccess(body string) bool {
	up := strings.ToUpper(strings.TrimSpace(body))
	if up == "" {
		return false
	}
	return strings.Contains(up, "SUKSES") || strings.Contains(up, "BERHASIL") ||
		(strings.Contains(up, "SDH PERNAH") && strings.Contains(up, "STATUS SUKSES"))
}

func LooksLikeAJSImmediateReject(body string) bool {
	if LooksLikeYuscomImmediateReject(body) {
		return true
	}
	up := strings.ToUpper(strings.TrimSpace(body))
	if up == "" {
		return false
	}
	return strings.Contains(up, "NOMOR TIDAK DITEMUKAN") ||
		strings.Contains(up, "PRODUK SEDANG GANGGUAN") ||
		strings.Contains(up, "SUSPECT")
}

func LooksLikeTalentaAccepted(body string) bool {
	up := strings.ToUpper(strings.TrimSpace(body))
	if up == "" {
		return false
	}
	return strings.Contains(up, "AKAN DIPROSES") || strings.Contains(up, "SUKSES") || strings.Contains(up, "BERHASIL")
}

func LooksLikeTalentaSuccess(body string) bool {
	up := strings.ToUpper(strings.TrimSpace(body))
	if up == "" {
		return false
	}
	return strings.Contains(up, "SUKSES") || strings.Contains(up, "BERHASIL")
}

func LooksLikeTalentaImmediateReject(body string) bool {
	if LooksLikeYuscomImmediateReject(body) {
		return true
	}
	return IsRetryableTalentaReject(body)
}

func LooksLikeMultikomAccepted(body string) bool {
	return LooksLikeMultikomPending(body) || LooksLikeMultikomSuccess(body)
}

func LooksLikeYuscomSuccess(body string) bool {
	up := strings.ToUpper(strings.TrimSpace(body))
	if up == "" {
		return false
	}
	return strings.Contains(up, "SUKSES") || strings.Contains(up, "BERHASIL")
}

func LooksLikeMultikomImmediateReject(body string) bool {
	return LooksLikeMultikomFailed(body) || LooksLikeYuscomImmediateReject(body)
}

func LooksLikeLoketBayarSuccess(rc string, msg string) bool {
	rc = strings.TrimSpace(strings.ToUpper(rc))
	msg = strings.TrimSpace(strings.ToUpper(msg))
	return rc == "00" || strings.Contains(msg, "SUKSES") || strings.Contains(msg, "SUCCESS")
}

func LooksLikeLoketBayarPending(rc string, msg string) bool {
	rc = strings.TrimSpace(strings.ToUpper(rc))
	msg = strings.TrimSpace(strings.ToUpper(msg))
	if looksLikeLoketBayarExplicitFailure(rc, msg) {
		return false
	}
	if rc == "" && reLoketBayarSaldoOnlyResponse.MatchString(msg) {
		return true
	}
	return rc == "9" || rc == "68" || strings.Contains(msg, "DALAM PROSES") || strings.Contains(msg, "SEDANG DIPROSES") || strings.Contains(msg, "PENDING")
}

func LooksLikeLoketBayarImmediateReject(rc string, msg string) bool {
	rc = strings.TrimSpace(strings.ToUpper(rc))
	msg = strings.TrimSpace(strings.ToUpper(msg))
	if looksLikeLoketBayarExplicitFailure(rc, msg) {
		return true
	}
	if LooksLikeLoketBayarSuccess(rc, msg) || LooksLikeLoketBayarPending(rc, msg) {
		return false
	}
	if rc == "" && msg == "" {
		return false
	}
	return true
}

func looksLikeLoketBayarExplicitFailure(rc string, msg string) bool {
	rc = strings.TrimSpace(strings.ToUpper(rc))
	msg = strings.TrimSpace(strings.ToUpper(msg))
	return rc == "GAGAL" || rc == "FAILED" ||
		strings.Contains(msg, "STATUS GAGAL") ||
		strings.Contains(msg, "TRANSAKSI GAGAL DI PROVIDER")
}

func LooksLikeRajabillerSuccess(rc string, msg string) bool {
	rc = strings.TrimSpace(strings.ToUpper(rc))
	msg = strings.TrimSpace(strings.ToUpper(msg))
	if rc != "" {
		return rc == "00"
	}
	return rc == "00" || strings.Contains(msg, "SUKSES") || strings.Contains(msg, "SUCCESS") || strings.Contains(msg, "BERHASIL")
}

func LooksLikeRajabillerPending(rc string, msg string) bool {
	rc = strings.TrimSpace(strings.ToUpper(rc))
	msg = strings.TrimSpace(strings.ToUpper(msg))
	if rc != "" {
		if rc == "68" {
			return true
		}
		if rc == "00" || isRajabillerFinalFailRC(rc) {
			return false
		}
		return true
	}
	return rc == "68" || strings.Contains(msg, "SEDANG DIPROSES") || strings.Contains(msg, "PENDING") || strings.Contains(msg, "PROSES") || strings.Contains(msg, "PROCESS")
}

func LooksLikeRajabillerImmediateReject(rc string, msg string) bool {
	rc = strings.TrimSpace(strings.ToUpper(rc))
	if rc != "" {
		return isRajabillerFinalFailRC(rc)
	}
	msg = strings.TrimSpace(strings.ToUpper(msg))
	if msg == "" {
		return false
	}
	return strings.Contains(msg, "GAGAL") ||
		strings.Contains(msg, "FAILED") ||
		strings.Contains(msg, "SALAH") ||
		strings.Contains(msg, "TIDAK TERDAFTAR") ||
		strings.Contains(msg, "TIDAK MENCUKUPI") ||
		strings.Contains(msg, "DUPLIKAT") ||
		strings.Contains(msg, "SUDAH TERBAYAR") ||
		strings.Contains(msg, "DIBLOKIR") ||
		strings.Contains(msg, "GANGGUAN")
}

func isRajabillerFinalFailRC(rc string) bool {
	switch strings.TrimSpace(strings.ToUpper(rc)) {
	case "01", "03", "04", "05", "06", "08", "09", "16", "77", "87", "97", "401", "403", "429":
		return true
	default:
		return false
	}
}

func LooksLikeGemilangAccepted(body string) bool {
	up := strings.ToUpper(strings.TrimSpace(body))
	if up == "" {
		return false
	}
	return strings.Contains(up, "AKAN DIPROSES") || strings.Contains(up, "SEDANG DIPROSES") || strings.Contains(up, "PENDING") || strings.Contains(up, "SUKSES") || strings.Contains(up, "BERHASIL")
}

func LooksLikeGemilangSuccess(body string) bool {
	return LooksLikeYuscomSuccess(body)
}

func LooksLikeGemilangImmediateReject(body string) bool {
	if LooksLikeYuscomImmediateReject(body) {
		return true
	}
	up := strings.ToUpper(strings.TrimSpace(body))
	if up == "" {
		return false
	}
	return strings.Contains(up, "SALAH KODE ATAU NOMINAL") ||
		strings.Contains(up, "NOMOR TUJUAN SALAH") ||
		strings.Contains(up, "PRODUK SEDANG GANGGUAN")
}

func ExtractProviderStatusCode(body string) string {
	body = strings.TrimSpace(body)
	if body == "" {
		return ""
	}
	m := reProviderStatusCode.FindStringSubmatch(body)
	if len(m) != 2 {
		return ""
	}
	return strings.TrimSpace(m[1])
}

func IsRetryableFailRC(rc string) bool {
	switch strings.TrimSpace(rc) {
	case "47", "50", "55", "61":
		return true
	default:
		return false
	}
}

func IsMissingProviderRC(rc string) bool {
	return strings.TrimSpace(rc) == ""
}

func IsRetryableTalentaReject(body string) bool {
	up := strings.ToUpper(strings.TrimSpace(body))
	if up == "" {
		return false
	}
	if IsRetryableFailRC(ExtractProviderStatusCode(body)) {
		return true
	}
	return strings.Contains(up, "STATUS=61") || strings.Contains(up, "STATUS = 61")
}

func MapProductForProviderFallback(provider string, internalProduct string) string {
	p := NormalizeInternalProductCode(internalProduct)
	prov := strings.TrimSpace(strings.ToLower(provider))

	if prov == "yuscom" {
		return p
	}
	return p
}

func SummarizeJPResp(jpResp map[string]any) map[string]any {
	out := map[string]any{}
	if jpResp == nil {
		return out
	}
	if v, ok := jpResp["status"]; ok {
		out["status"] = v
	}
	if data, ok := jpResp["data"].(map[string]any); ok {
		out["data"] = map[string]any{
			"rc":           data["rc"],
			"message":      data["message"],
			"trxid":        data["trxid"],
			"noreff":       data["noreff"],
			"price":        data["price"],
			"last_balance": data["last_balance"],
		}
	}
	if errv, ok := jpResp["error"]; ok {
		out["error"] = errv
	}
	return out
}

func PtrString(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &s
}

func PtrI64(v int64) *int64 {
	if v <= 0 {
		return nil
	}
	return &v
}
