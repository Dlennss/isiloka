package provider

import "testing"

func TestCatalogPrice(t *testing.T) {
	zero, fee, fixed, negative := int64(0), int64(1000), int64(10100), int64(-1)
	for _, tc := range []struct {
		name    string
		p       Pulsa24JamProduct
		want    int64
		invalid bool
	}{
		{"fixed", Pulsa24JamProduct{PriceType: "FIXED", Price: &fixed}, fixed, false},
		{"inquiry zero", Pulsa24JamProduct{PriceType: "FIXED", Price: &zero}, 0, false},
		{"open fee not face value", Pulsa24JamProduct{PriceType: "OPEN_AMOUNT", Price: &fixed, AdditionalFee: &fee}, fee, false},
		{"app price", Pulsa24JamProduct{PriceType: "FIXED", Price: &fixed, AppBasePrice: &fee}, fee, false},
		{"missing", Pulsa24JamProduct{PriceType: "FIXED"}, 0, true},
		{"missing open fee", Pulsa24JamProduct{PriceType: "OPEN_AMOUNT", Price: &fixed}, 0, true},
		{"negative", Pulsa24JamProduct{PriceType: "FIXED", Price: &negative}, 0, true},
		{"unknown type", Pulsa24JamProduct{Price: &fixed}, 0, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.p.CatalogPrice()
			if (err != nil) != tc.invalid || got != tc.want {
				t.Fatalf("got %d, %v", got, err)
			}
		})
	}
}
