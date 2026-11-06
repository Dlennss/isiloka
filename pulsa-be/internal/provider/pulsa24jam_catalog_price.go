package provider

import "fmt"

// CatalogPrice distinguishes an absent price from a legitimate zero-priced inquiry.
func (p Pulsa24JamProduct) CatalogPrice() (int64, error) {
	var price *int64
	switch p.PriceType {
	case "FIXED":
		price = firstInt64Pointer(p.AppBasePrice, p.Price)
	case "OPEN_AMOUNT":
		price = firstInt64Pointer(p.AppBasePrice, p.AdditionalFee)
	default:
		return 0, fmt.Errorf("tipe harga Pulsa24Jam tidak valid: %s", p.SKU)
	}
	if price == nil || *price < 0 {
		return 0, fmt.Errorf("harga Pulsa24Jam tidak tersedia atau tidak valid: %s", p.SKU)
	}
	return *price, nil
}
