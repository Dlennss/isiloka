package service

import (
	"testing"

	"pulsa2/internal/repository"
)

func TestDeriveBankCodeFromProviderRouteRajabiller(t *testing.T) {
	got, err := deriveBankCodeFromProviderRoute("rajabiller", "", "BLTRFAG:014")
	if err != nil {
		t.Fatalf("deriveBankCodeFromProviderRoute returned error: %v", err)
	}
	if got != "014" {
		t.Fatalf("bank code = %q, want 014", got)
	}
}

func TestDeriveBankCodeFromProviderRouteRajabillerLegacyLayout(t *testing.T) {
	got, err := deriveBankCodeFromProviderRoute("rajabiller", "", "014:BLTRFAG")
	if err != nil {
		t.Fatalf("deriveBankCodeFromProviderRoute returned error: %v", err)
	}
	if got != "014" {
		t.Fatalf("bank code = %q, want 014", got)
	}
}

func TestLatestBankCodeFromAttemptsSupportsRajabiller(t *testing.T) {
	got, err := latestBankCodeFromAttempts([]repository.ProviderAttemptRow{
		{Provider: "rajabiller", KodeProduk: "BLTRFAG:008"},
	})
	if err != nil {
		t.Fatalf("latestBankCodeFromAttempts returned error: %v", err)
	}
	if got != "008" {
		t.Fatalf("bank code = %q, want 008", got)
	}
}

func TestExpandBankProviderAttemptsKeepsLoketLast(t *testing.T) {
	attempts := []providerRouteAttempt{
		{Name: "loketbayar", KodeProduk: "014", SpecialCode: "DBALLBANK"},
		{Name: "rajabiller", KodeProduk: "BLTRFAG", SpecialCode: "014"},
		{Name: "smb", KodeProduk: "014"},
	}

	got := expandBankProviderAttempts(attempts, true)
	if len(got) != 3 {
		t.Fatalf("len = %d, want 3", len(got))
	}
	if got[0].Name != "rajabiller" || got[1].Name != "smb" || got[2].Name != "loketbayar" {
		t.Fatalf("order = %s,%s,%s; want rajabiller,smb,loketbayar", got[0].Name, got[1].Name, got[2].Name)
	}
	if got[1].Mode != "DIRECT" || got[1].SpecialCode != "BIFASTOPEN" {
		t.Fatalf("smb normalized mode=%q special=%q, want DIRECT/BIFASTOPEN", got[1].Mode, got[1].SpecialCode)
	}
}

func TestLoketBayarBankTransferProductAcceptsLegacyTRFBANK(t *testing.T) {
	if !isLoketBayarBankTransferProduct("DBALLBANK") {
		t.Fatal("DBALLBANK should be recognized as LoketBayar bank transfer product")
	}
	if !isLoketBayarBankTransferProduct("TRFBANK") {
		t.Fatal("legacy TRFBANK should still be recognized")
	}
}

func TestBankFallbackRankMainProviderBeforeLoket(t *testing.T) {
	if got := bankFallbackRank("smb", "rajabiller"); got != 0 {
		t.Fatalf("rank smb->rajabiller = %d, want 0", got)
	}
	if got := bankFallbackRank("smb", "loketbayar"); got != 1 {
		t.Fatalf("rank smb->loketbayar = %d, want 1", got)
	}
	if got := bankFallbackRank("rajabiller", "smb"); got != 0 {
		t.Fatalf("rank rajabiller->smb = %d, want 0", got)
	}
	if got := bankFallbackRank("rajabiller", "loketbayar"); got != 1 {
		t.Fatalf("rank rajabiller->loketbayar = %d, want 1", got)
	}
	if got := bankFallbackRank("smb", "chytron"); got >= 0 {
		t.Fatalf("rank smb->chytron = %d, want negative", got)
	}
}
