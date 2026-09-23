package service

import "strings"

const (
	loketBayarBankTransferProduct       = "DBALLBANK"
	loketBayarLegacyBankTransferProduct = "TRFBANK"
)

func isLoketBayarBankTransferProduct(product string) bool {
	switch strings.ToUpper(strings.TrimSpace(product)) {
	case loketBayarBankTransferProduct, loketBayarLegacyBankTransferProduct:
		return true
	default:
		return false
	}
}
