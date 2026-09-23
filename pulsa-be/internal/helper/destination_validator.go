package helper

import (
	"strings"
	"unicode"
)

// IsValidDestinationFormat — validasi format tujuan untuk nominal bebas.
// Harus angka semua dan diawali 08 atau 628.
func IsValidDestinationFormat(dest string) bool {
	dest = strings.TrimSpace(dest)
	if dest == "" {
		return false
	}

	// Harus semua angka — tidak boleh ada huruf, spasi, +, -, dll
	for _, r := range dest {
		if !unicode.IsDigit(r) {
			return false
		}
	}

	if len(dest) < 4 {
		return false
	}

	return strings.HasPrefix(dest, "08") || strings.HasPrefix(dest, "628")
}

// IsValidBankAccountFormat — validasi nomor rekening bank.
// Harus angka semua, minimal 5 digit.
func IsValidBankAccountFormat(dest string) bool {
	dest = strings.TrimSpace(dest)
	if dest == "" {
		return false
	}
	for _, r := range dest {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return len(dest) >= 5
}
