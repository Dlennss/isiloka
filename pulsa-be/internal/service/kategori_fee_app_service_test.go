package service

import (
	"testing"

	"pulsa2/internal/repository"
)

func TestNormalizeKategoriFeeAppInputAllowsNegativeFees(t *testing.T) {
	in := repository.KategoriFeeAppUpsertInput{
		KategoriID: 1,
		FeeMaster:  -100,
		FeeAgent:   -200,
		FeeUser:    -300,
		FeeNonUser: -400,
		Aktif:      true,
	}

	got, err := normalizeKategoriFeeAppInput(in)
	if err != nil {
		t.Fatalf("normalizeKategoriFeeAppInput() error = %v", err)
	}
	if got != in {
		t.Fatalf("normalizeKategoriFeeAppInput() = %#v, want %#v", got, in)
	}
}

func TestNormalizeKategoriFeeAppInputRejectsInvalidKategori(t *testing.T) {
	_, err := normalizeKategoriFeeAppInput(repository.KategoriFeeAppUpsertInput{})
	if err == nil {
		t.Fatal("normalizeKategoriFeeAppInput() error = nil, want error")
	}
}
