package service

import (
	"strings"

	"pulsa2/internal/helper"
)

func isChargeReceiverEligibleProduct(product string, tipeHarga string, kategoriNama string, brandNama string) bool {
	if !strings.EqualFold(strings.TrimSpace(tipeHarga), "OPEN_AMOUNT") {
		return false
	}

	p := strings.ToUpper(strings.TrimSpace(product))
	if helper.ClassifyH2HFeeCategoryFromMetadata(p, kategoriNama, brandNama) == helper.H2HFeeCategoryBank {
		return true
	}
	return strings.HasPrefix(p, "DANA") ||
		strings.HasPrefix(p, "DNID") ||
		strings.HasPrefix(p, "OVO") ||
		strings.HasPrefix(p, "GOPAY") ||
		strings.HasPrefix(p, "SHOPEE") ||
		strings.HasPrefix(p, "LINKAJA")
}
