package repository

import "testing"

func TestValidatePulsa24JamCatalog(t *testing.T) {
	valid := Pulsa24JamCatalogItem{SKU: "DB COMMONWEALTH", Name: "Transfer Bank", CategoryName: "Bank Transfer", BrandName: "Commonwealth", PriceType: "OPEN_AMOUNT", Price: 1500}
	if err := ValidatePulsa24JamCatalog([]Pulsa24JamCatalogItem{valid}); err != nil {
		t.Fatal(err)
	}
	for _, items := range [][]Pulsa24JamCatalogItem{nil, {valid, valid}, {{SKU: "X"}}} {
		if ValidatePulsa24JamCatalog(items) == nil {
			t.Fatal("accepted incomplete or duplicate snapshot")
		}
	}
	for _, change := range []func(*Pulsa24JamCatalogItem){
		func(p *Pulsa24JamCatalogItem) { p.Price = -1 },
		func(p *Pulsa24JamCatalogItem) { p.PriceType = "unknown" },
		func(p *Pulsa24JamCatalogItem) { p.BrandName = "" },
		func(p *Pulsa24JamCatalogItem) { z := int64(0); p.MaximumNominal = &z },
	} {
		item := valid
		change(&item)
		if ValidatePulsa24JamCatalog([]Pulsa24JamCatalogItem{item}) == nil {
			t.Fatalf("accepted %+v", item)
		}
	}
}
