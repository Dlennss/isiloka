package service

import (
	"context"
	"pulsa2/internal/provider"
	"pulsa2/internal/repository"
	"testing"
)

func TestDisplayedPulsa24JamQuote(t *testing.T) {
	price := int64(1000)
	in := repository.AppOrderCreateInput{ExpectedBasePrice: &price, ExpectedPriceType: "OPEN_AMOUNT"}
	if validateDisplayedPulsa24JamQuote(in, "OPEN_AMOUNT", price) != nil {
		t.Fatal("rejected matching quote")
	}
	if validateDisplayedPulsa24JamQuote(in, "OPEN_AMOUNT", price+1) == nil {
		t.Fatal("accepted stale displayed price")
	}
	if validateDisplayedPulsa24JamQuote(in, "FIXED", price) == nil {
		t.Fatal("accepted changed price type")
	}
}

func TestPulsa24JamQuoteMustMatchCatalog(t *testing.T) {
	price := int64(1000)
	p := &repository.ProdukRow{SKU: "DANA", Nama: "DANA OPEN AMOUNT", TipeHarga: "OPEN_AMOUNT"}
	live := provider.Pulsa24JamProduct{SKU: p.SKU, Name: p.Nama, PriceType: p.TipeHarga, Active: true, AdditionalFee: &price}
	if err := validatePulsa24JamQuote(p, price, &live); err != nil {
		t.Fatal(err)
	}
	for _, change := range []func(*provider.Pulsa24JamProduct){
		func(p *provider.Pulsa24JamProduct) { p.SKU = "UDDND" },
		func(p *provider.Pulsa24JamProduct) { p.PriceType = "FIXED" },
		func(p *provider.Pulsa24JamProduct) { p.Active = false },
		func(p *provider.Pulsa24JamProduct) { p.AdditionalFee = nil },
		func(p *provider.Pulsa24JamProduct) { max := int64(50000); p.MaximumNominal = &max },
	} {
		other := live
		change(&other)
		if validatePulsa24JamQuote(p, price, &other) == nil {
			t.Fatalf("accepted %+v", other)
		}
	}
	if validatePulsa24JamQuote(p, price+1, &live) == nil {
		t.Fatal("accepted changed price")
	}
	if _, err := (&AppOrderService{}).validatePulsa24JamProduct(context.Background(), "DANA"); err == nil {
		t.Fatal("accepted unconfigured provider")
	}
}
