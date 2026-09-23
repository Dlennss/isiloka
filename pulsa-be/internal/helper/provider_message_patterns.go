package helper

import (
	"strings"
	"sync"
)

// MessagePatternStore — pattern untuk classify pesan provider.
// Default di-hardcode, tapi bisa di-override dari database via SetPatterns.
// Thread-safe. Setelah SetPatterns dipanggil, semua fungsi classify
// (ProviderResponseStateOf, ClassifyJavapayResponseStatus, dll) otomatis
// pakai pattern baru tanpa restart.

type messagePatterns struct {
	mu      sync.RWMutex
	failed  []string
	success []string
	pending []string
	loaded  bool
}

var patterns = &messagePatterns{
	failed: []string{
		"GAGAL", "FAILED", "BATAL", "DITOLAK", "TIDAK", "ERROR",
		"TIMEOUT", "TUJUAN SALAH", "DO NOT HONO", "NOMOR TIDAK DIKENAL",
		"DIABAIKAN", "FORMAT SALAH", "PRODUK SALAH",
	},
	success: []string{
		"SUKSES", "BERHASIL", "LUNAS",
	},
	pending: []string{
		"MENUNGGU JAWABAN", "SEDANG DIPROSES", "AKAN DIPROSES",
	},
}

// SetPatterns — dipanggil dari service layer setelah baca database.
// Dipanggil periodically (misal tiap 5 menit) untuk refresh.
func SetPatterns(failed, success, pending []string) {
	patterns.mu.Lock()
	defer patterns.mu.Unlock()
	if len(failed) > 0 {
		patterns.failed = failed
	}
	if len(success) > 0 {
		patterns.success = success
	}
	if len(pending) > 0 {
		patterns.pending = pending
	}
	patterns.loaded = true
}

// MatchesFailedPattern — cek apakah pesan mengandung pattern failed
func MatchesFailedPattern(upperMsg string) bool {
	patterns.mu.RLock()
	defer patterns.mu.RUnlock()
	for _, p := range patterns.failed {
		if strings.Contains(upperMsg, p) {
			return true
		}
	}
	return false
}

// MatchesSuccessPattern — cek apakah pesan mengandung pattern success
func MatchesSuccessPattern(upperMsg string) bool {
	patterns.mu.RLock()
	defer patterns.mu.RUnlock()
	for _, p := range patterns.success {
		if strings.Contains(upperMsg, p) {
			return true
		}
	}
	return false
}

// MatchesPendingPattern — cek apakah pesan mengandung pattern pending
func MatchesPendingPattern(upperMsg string) bool {
	patterns.mu.RLock()
	defer patterns.mu.RUnlock()
	for _, p := range patterns.pending {
		if strings.Contains(upperMsg, p) {
			return true
		}
	}
	return false
}

// PatternsLoaded — apakah sudah pernah di-load dari database
func PatternsLoaded() bool {
	patterns.mu.RLock()
	defer patterns.mu.RUnlock()
	return patterns.loaded
}
