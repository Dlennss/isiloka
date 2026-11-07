package service

import (
	"testing"

	"pulsa2/internal/provider"
	"pulsa2/internal/repository"
)

func TestPulsa24JamEWalletAllowlist(t *testing.T) {
	items := []repository.Pulsa24JamCatalogItem{
		{SKU: "PPOBDANA", CategoryName: "E-Wallet", Price: 1000},
		{SKU: "GOPAY", CategoryName: "E-Wallet", Price: 1200},
		{SKU: "GPC4P", CategoryName: "E-Wallet", Price: 1200},
		{SKU: "PULSA10", CategoryName: "Pulsa", Price: 11000},
	}
	got := filterPulsa24JamEWalletCatalog(items, " PPOBDANA, GOPAY ")
	if len(got) != 3 || got[0] != items[0] || got[1] != items[1] || got[2] != items[3] {
		t.Fatalf("allowlist changed pricing or other categories: %+v", got)
	}
	if got := filterPulsa24JamEWalletCatalog(items, ""); len(got) != len(items) {
		t.Fatal("empty policy must preserve the catalogue")
	}
	if got := filterPulsa24JamEWalletCatalog(items, "UNKNOWN"); len(got) != 1 || got[0].SKU != "PULSA10" {
		t.Fatal("unknown allowed SKU must not enable other wallets")
	}
}

func TestPulsa24JamCatalogItemFromProductKeepsUpstreamCatalog(t *testing.T) {
	price := int64(12500)
	tests := []struct {
		name    string
		product provider.Pulsa24JamProduct
	}{
		{
			name:    "indosat remains available",
			product: provider.Pulsa24JamProduct{SKU: "YIT10", Name: "Indosat 10K", BrandName: "indosat", CategoryName: "Pulsa", Price: &price},
		},
		{
			name:    "formerly blocked prefix remains available",
			product: provider.Pulsa24JamProduct{SKU: "UDGD15", Name: "Gopay Driver 15K", BrandName: "gopay", CategoryName: "E-Money", Price: &price},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pulsa24JamCatalogItemFromProduct(tt.product)
			if got.SKU != tt.product.SKU || got.Name != tt.product.Name || got.Price != price {
				t.Fatalf("produk berubah saat dipetakan: %+v", got)
			}
		})
	}
}

func TestPulsa24JamCatalogItemUsesAppBasePrice(t *testing.T) {
	legacyPrice := int64(12500)
	appBasePrice := int64(13000)
	got := pulsa24JamCatalogItemFromProduct(provider.Pulsa24JamProduct{
		SKU: "ML10", Name: "Mobile Legend 10 Diamond", Price: &legacyPrice, AppBasePrice: &appBasePrice,
	})
	if got.Price != appBasePrice {
		t.Fatalf("harga katalog = %d, want %d", got.Price, appBasePrice)
	}
}
