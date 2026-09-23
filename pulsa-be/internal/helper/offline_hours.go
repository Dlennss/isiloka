package helper

import (
	"strings"
	"time"
)

var wib = time.FixedZone("WIB", 7*60*60)

func H2HProductUsesOfflineHours(product string) bool {
	p := strings.ToUpper(strings.TrimSpace(product))
	return strings.HasPrefix(p, "OVO")
}

func IsH2HProductAvailableForNow(product, jamBuka, jamTutup string) bool {
	if !H2HProductUsesOfflineHours(product) {
		return true
	}
	return IsProductAvailableNow(jamBuka, jamTutup)
}

// IsProductAvailableNow — cek apakah produk tersedia berdasarkan jam_buka dan jam_tutup.
// jamBuka dan jamTutup format "HH:MM:SS" dari database.
// Hanya dipakai untuk produk yang memang punya offline window.
// Jika kosong, default produk tersedia 00:31 - 23:29.
func IsProductAvailableNow(jamBuka, jamTutup string) bool {
	now := time.Now().In(wib)
	nowMinutes := now.Hour()*60 + now.Minute()

	buka := parseTimeToMinutes(jamBuka, 0*60+31)    // default 00:31
	tutup := parseTimeToMinutes(jamTutup, 23*60+29) // default 23:29

	if buka <= tutup {
		return nowMinutes >= buka && nowMinutes <= tutup
	}
	// crossing midnight (misal buka=22:00 tutup=06:00)
	return nowMinutes >= buka || nowMinutes <= tutup
}

// parseTimeToMinutes — parse "HH:MM:SS" atau "HH:MM" ke total menit sejak 00:00.
func parseTimeToMinutes(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	t, err := time.Parse("15:04:05", s)
	if err != nil {
		t, err = time.Parse("15:04", s)
		if err != nil {
			return defaultVal
		}
	}
	return t.Hour()*60 + t.Minute()
}
