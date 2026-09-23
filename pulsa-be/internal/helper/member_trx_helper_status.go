package helper

import "strings"

func SafeMemberKeterangan(status string, providerMsg string) (ket string, info ProviderInfo) {
	info = extractProviderInfo(providerMsg)
	switch status {
	case "success":
		if info.Reff != "" {
			return "Transaksi berhasil. Ref: " + info.Reff, info
		}
		return "Transaksi berhasil", info
	case "failed":
		if failure := normalizeFailureMessage(providerMsg); failure != "" {
			return failure, info
		}
		return "Transaksi gagal", info
	default:
		return "Sedang diproses", info
	}
}

func normalizeFailureMessage(providerMsg string) string {
	msg := strings.TrimSpace(providerMsg)
	if msg == "" {
		return ""
	}

	compact := strings.Join(strings.Fields(msg), " ")
	if compact == "" {
		return ""
	}

	if ExtractProviderStatusCode(compact) == "52" {
		return "Nomor tujuan salah"
	}

	upper := strings.ToUpper(compact)
	replacements := []struct {
		match string
		out   string
	}{
		{"INVALID PIN OR PASSWORD", "Kredensial provider tidak valid"},
		{"INVALID MEMBERID", "Kredensial provider tidak valid"},
		{"INVALID SIGNATURE", "Signature provider tidak valid"},
		{"INVALID REQUEST", "Provider error"},
		{"NOMOR TUJUAN SALAH", "Nomor tujuan salah"},
		{"TUJUAN SALAH", "Nomor tujuan salah"},
		{"CEK KEMBALI NOMOR TUJUAN", "Nomor tujuan salah"},
		{"NOMOR TUJUAN DALAM BLACKLIST", "Nomor tujuan salah"},
		{"NOMOR /ID SALAH/TIDAK BISA DIPROSES", "Nomor tujuan salah"},
		{"NOMOR TUJUAN SALAH, MOHON DIPERIKSA KEMBALI", "Nomor tujuan salah"},
		{"BATAS MAKSIMAL PEMBELIAN", "Batas maksimal pembelian / Akun sdh Limit"},
		{"AKUN SDH LIMIT", "Batas maksimal pembelian / Akun sdh Limit"},
		{"SALDO TIDAK CUKUP", "Saldo tidak cukup"},
		{"STOK TIDAK CUKUP", "Stok tidak cukup"},
		{"QTY TIDAK SESUAI", "Qty tidak sesuai"},
		{"ALLOWED QTY", "Qty tidak sesuai"},
		{"LIMIT 0", "Stok provider habis"},
		{"STOK DIKEMBALIKAN", "Stok dikembalikan"},
		{"CUTOFF", "Provider cutoff"},
		{"TIMEOUT", "Provider timeout"},
		{"DIBATALKAN", "Transaksi dibatalkan"},
		{"PRODUK KEHABISAN STOK", "Produk kehabisan stok"},
		{"PRODUK TIDAK TERDAFTAR", "Produk tidak terdaftar"},
		{"PRODUK SEDANG GANGGUAN", "Produk sedang gangguan"},
		{"NOMOR TIDAK DAPAT DIPROSES", "Nomor tidak dapat diproses"},
		{"SOMETHING WRONG", "Provider error"},
		{"PROVIDER ERROR", "Provider error"},
	}
	for _, repl := range replacements {
		if strings.Contains(upper, repl.match) {
			return repl.out
		}
	}

	if IsProviderRawMemberMessage(compact) {
		return "Transaksi gagal"
	}
	if strings.Contains(upper, "GAGAL") {
		return "Transaksi gagal"
	}
	return "Transaksi gagal"
}

func IsProviderRawMemberMessage(providerMsg string) bool {
	compact := strings.ToUpper(strings.Join(strings.Fields(strings.TrimSpace(providerMsg)), " "))
	if compact == "" {
		return false
	}
	markers := []string{
		"REJECT BISNIS",
		" SALDO:",
		"SALDO ",
		" SAL:",
		" STOK:",
		"STOK ",
		" HARGA:",
		" HRG:",
		"STATUS=",
		"MESSAGE=",
		"HTTP=",
		" ERR=",
		"ERROR=",
		"SQLMESSAGE",
		"SQLSTATE",
		"TRANSAKSIPIPA",
		"RESPON_MENTAH",
		"REQUEST_MENTAH",
	}
	for _, marker := range markers {
		if strings.Contains(compact, marker) {
			return true
		}
	}
	return false
}

func ShouldBlockProviderFallback(providerMsg string) bool {
	compact := strings.Join(strings.Fields(strings.TrimSpace(providerMsg)), " ")
	if ExtractProviderStatusCode(compact) == "52" {
		return true
	}

	upper := strings.ToUpper(compact)
	if upper == "" {
		return false
	}
	// Terminal across providers: trying another provider will not fix these.
	return strings.Contains(upper, "NOMOR TUJUAN SALAH") ||
		strings.Contains(upper, "TUJUAN SALAH") ||
		strings.Contains(upper, "NOMOR /ID SALAH/TIDAK BISA DIPROSES") ||
		strings.Contains(upper, "NOMOR TIDAK DAPAT DIPROSES") ||
		strings.Contains(upper, "CUSTOMER ACCOUNT NOT FOUND") ||
		strings.Contains(upper, "ACCOUNT NOT FOUND") ||
		strings.Contains(upper, "NOMOR TIDAK DITEMUKAN") ||
		strings.Contains(upper, "BATAS MAKSIMAL PEMBELIAN") ||
		strings.Contains(upper, "AKUN SDH LIMIT")
}

func LooksLikeProviderOutOfBalance(providerMsg string) bool {
	compact := strings.Join(strings.Fields(strings.TrimSpace(providerMsg)), " ")
	if compact == "" {
		return false
	}
	upper := strings.ToUpper(compact)
	return strings.Contains(upper, "SALDO TIDAK MENCUKUPI") ||
		strings.Contains(upper, "SALDO TIDAK CUKUP") ||
		strings.Contains(upper, "SALDO PROVIDER TIDAK CUKUP") ||
		strings.Contains(upper, "STOK TIDAK CUKUP") ||
		strings.Contains(upper, "INSUFFICIENT BALANCE")
}

func IsJavapayStatusNotFoundMessage(msg string) bool {
	upperMsg := strings.ToUpper(strings.TrimSpace(msg))
	if upperMsg == "" {
		return false
	}
	return strings.Contains(upperMsg, "TIDAK DITEMUKAN") ||
		strings.Contains(upperMsg, "NOT FOUND") ||
		strings.Contains(upperMsg, "REFID NOT FOUND") ||
		strings.Contains(upperMsg, "TRX NOT FOUND")
}

func IsJavapayWaitingMessage(msg string) bool {
	upperMsg := strings.ToUpper(strings.Join(strings.Fields(strings.TrimSpace(msg)), " "))
	if upperMsg == "" {
		return false
	}
	return strings.Contains(upperMsg, "MENUNGGU JAWABAN") ||
		strings.Contains(upperMsg, "SEDANG DIPROSES")
}

func ClassifyJavapayResponseStatus(rc, msg string) ProviderResponseState {
	if strings.TrimSpace(rc) == "" && strings.TrimSpace(msg) == "" {
		return ProviderResponseUnknown
	}
	if IsJavapayStatusNotFoundMessage(msg) {
		return ProviderResponseFailed
	}
	if shouldRetryStaleJavapayStatus(rc, msg) {
		return ProviderResponsePending
	}
	upperMsg := strings.ToUpper(strings.TrimSpace(msg))
	switch {
	case MatchesSuccessPattern(upperMsg):
		return ProviderResponseSuccess
	case IsJavapayWaitingMessage(msg):
		return ProviderResponsePending
	case MatchesFailedPattern(upperMsg):
		return ProviderResponseFailed
	default:
		return ProviderResponsePending
	}
}

func shouldRetryStaleJavapayStatus(rc, msg string) bool {
	upperMsg := strings.ToUpper(strings.TrimSpace(msg))
	return strings.TrimSpace(rc) == "64" ||
		strings.Contains(upperMsg, "DIABAIKAN")
}
