package service

import (
	"testing"

	"pulsa2/internal/provider"
)

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
