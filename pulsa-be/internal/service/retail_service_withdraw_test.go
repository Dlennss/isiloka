package service

import (
	"testing"

	"pulsa2/internal/provider"
)

func TestRetailWithdrawOpenWalletRequest(t *testing.T) {
	tests := []struct {
		name        string
		item        provider.Pulsa24JamProduct
		amount      int64
		wantProduct string
		wantQty     int64
		wantOK      bool
	}{
		{
			name:        "gopay regular uses open amount request",
			item:        provider.Pulsa24JamProduct{SKU: "UDGP10", Name: "Gopay 10.000"},
			amount:      10000,
			wantProduct: "GOPAY",
			wantQty:     10000,
			wantOK:      true,
		},
		{
			name:        "dana regular uses open amount request",
			item:        provider.Pulsa24JamProduct{SKU: "UDDND50", Name: "Dana 50.000"},
			amount:      50000,
			wantProduct: "DANA",
			wantQty:     50000,
			wantOK:      true,
		},
		{
			name:   "gopay driver keeps fixed sku",
			item:   provider.Pulsa24JamProduct{SKU: "UDGD10", Name: "Gopay Driver 10.000"},
			amount: 10000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product, qty, ok := retailWithdrawOpenWalletRequest(tt.item, tt.amount)
			if product != tt.wantProduct || qty != tt.wantQty || ok != tt.wantOK {
				t.Fatalf("got product=%q qty=%d ok=%v, want product=%q qty=%d ok=%v", product, qty, ok, tt.wantProduct, tt.wantQty, tt.wantOK)
			}
		})
	}
}

func TestRetailPulsa24JamNestedMemberStatus(t *testing.T) {
	successBody := `{"ok":true,"transaksi_member":{"ref_id":"RWD-1","status":2}}`
	if !retailPulsa24JamLooksSuccess(successBody) {
		t.Fatal("nested transaksi_member status 2 must be treated as successful")
	}
	if retailPulsa24JamLooksRejected(successBody) {
		t.Fatal("nested transaksi_member status 2 must not be treated as rejected")
	}

	rejectedBody := `{"ok":true,"transaksi_member":{"ref_id":"RWD-2","status":3}}`
	if !retailPulsa24JamLooksRejected(rejectedBody) {
		t.Fatal("nested transaksi_member status 3 must be treated as rejected")
	}
}
