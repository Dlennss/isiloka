package helper

import "strings"

func LooksLikeMultikomSuccess(msg string) bool {
	up := strings.ToUpper(strings.TrimSpace(msg))
	if up == "" {
		return false
	}
	return strings.Contains(up, " SUKSES ") ||
		strings.Contains(up, "STATUS SUKSES") ||
		strings.Contains(up, "STATUS=SUKSES")
}

func LooksLikeMultikomPending(msg string) bool {
	up := strings.ToUpper(strings.TrimSpace(msg))
	if up == "" {
		return false
	}
	return strings.Contains(up, "AKAN DIPROSES") ||
		strings.Contains(up, "SEDANG DIPROSES") ||
		strings.Contains(up, "PENDING")
}

func LooksLikeMultikomFailed(msg string) bool {
	up := strings.ToUpper(strings.TrimSpace(msg))
	if up == "" {
		return false
	}
	return strings.Contains(up, "GAGAL") ||
		strings.Contains(up, "STATUS TIMEOUT") ||
		strings.Contains(up, "STATUS TIMEOUT.") ||
		strings.Contains(up, "STATUS TIMEOUT") ||
		strings.Contains(up, "CUTOFF") ||
		strings.Contains(up, "TIDAK DIPROSES") ||
		strings.Contains(up, "DIBATALKAN") ||
		strings.Contains(up, "NOMOR TUJUAN SALAH") ||
		strings.Contains(up, "PRODUK SALAH") ||
		strings.Contains(up, "STOK TIDAK CUKUP") ||
		strings.Contains(up, "STOK SEDANG KOSONG/DITUTUP") ||
		strings.Contains(up, "PRODUK SEDANG GANGGUAN") ||
		strings.Contains(up, "STOK DIKEMBALIKAN")
}
