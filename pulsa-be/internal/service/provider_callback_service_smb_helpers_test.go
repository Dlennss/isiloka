package service

import (
	"testing"

	"pulsa2/smb"
)

func TestSMBFailureAllowsDowngradeOrFallback(t *testing.T) {
	t.Run("bank remains eligible", func(t *testing.T) {
		if !smbFailureAllowsDowngradeOrFallback(true, "", "", "nomor salah") {
			t.Fatal("bank failure should remain eligible")
		}
	})

	t.Run("legacy bifastopen remains eligible", func(t *testing.T) {
		if !smbFailureAllowsDowngradeOrFallback(false, smb.ModeDirect, "BIFASTOPEN", "gagal") {
			t.Fatal("BIFASTOPEN direct failure should remain eligible")
		}
	})

	t.Run("ewallet failed after prior success is eligible", func(t *testing.T) {
		msg := "[SEMPAT SUKSES SN:123] REFF#abc DANA GAGAL, KET: REFUND/SALDO DIKEMBALIKAN"
		if !smbFailureAllowsDowngradeOrFallback(false, smb.ModeDirect, "DANA", msg) {
			t.Fatal("ewallet failed after prior success should be eligible")
		}
	})

	t.Run("ordinary ewallet failure stays unchanged", func(t *testing.T) {
		if smbFailureAllowsDowngradeOrFallback(false, smb.ModeDirect, "DANA", "REFF#abc DANA GAGAL") {
			t.Fatal("ordinary ewallet failure should not be newly eligible")
		}
	})
}
