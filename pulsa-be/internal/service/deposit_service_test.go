package service

import (
	"testing"
	"time"

	"pulsa2/internal/repository"
)

func TestDepositTicketRequestOfflineAt(t *testing.T) {
	loc := depositTicketLocation()
	tests := []struct {
		name string
		at   time.Time
		want bool
	}{
		{name: "before offline window", at: time.Date(2026, 6, 4, 23, 29, 59, 0, loc), want: false},
		{name: "offline starts 2330", at: time.Date(2026, 6, 4, 23, 30, 0, 0, loc), want: true},
		{name: "offline before midnight", at: time.Date(2026, 6, 4, 23, 59, 59, 0, loc), want: true},
		{name: "offline after midnight", at: time.Date(2026, 6, 5, 0, 0, 0, 0, loc), want: true},
		{name: "offline through 0030", at: time.Date(2026, 6, 5, 0, 30, 59, 0, loc), want: true},
		{name: "available at 0031", at: time.Date(2026, 6, 5, 0, 31, 0, 0, loc), want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := depositTicketRequestOfflineAt(tt.at); got != tt.want {
				t.Fatalf("depositTicketRequestOfflineAt(%s) = %v, want %v", tt.at.Format(time.RFC3339), got, tt.want)
			}
		})
	}
}

func TestDepositQrisTopupHasNoAdminFee(t *testing.T) {
	amount := int64(500000)
	feeAdmin := calcDepositQrisFee()
	if feeAdmin != 0 {
		t.Fatalf("calcDepositQrisFee(%d) = %d, want 0", amount, feeAdmin)
	}

	payload := buildDepositMidtransChargeRequest("DQR-TEST", amount, feeAdmin, amount+feeAdmin)
	if payload.TransactionDetails.GrossAmt != amount {
		t.Fatalf("GrossAmt = %d, want %d", payload.TransactionDetails.GrossAmt, amount)
	}
	if payload.Items == nil {
		t.Fatalf("Items is nil")
	}
	if got := len(*payload.Items); got != 1 {
		t.Fatalf("item count = %d, want 1", got)
	}
	if got := (*payload.Items)[0].ID; got != "TOPUP-SALDO" {
		t.Fatalf("first item ID = %q, want TOPUP-SALDO", got)
	}
}

func TestDepositQrisPaidGrossAmountUsesMidtransPayload(t *testing.T) {
	payload := map[string]any{"gross_amount": "503500"}
	if got := depositQrisPaidGrossAmount(500000, payload); got != 503500 {
		t.Fatalf("depositQrisPaidGrossAmount = %d, want 503500", got)
	}

	if got := depositQrisPaidGrossAmount(500000, map[string]any{}); got != 500000 {
		t.Fatalf("depositQrisPaidGrossAmount fallback = %d, want 500000", got)
	}
}

func TestIsMemberDepositBankRejectsSystemAndInternalBanks(t *testing.T) {
	tests := []struct {
		name string
		row  repository.BankRow
		want bool
	}{
		{name: "normal active bank", row: repository.BankRow{Nama: "BCA"}, want: true},
		{name: "qris system bank", row: repository.BankRow{Nama: "QRIS"}, want: false},
		{name: "qrtp system bank", row: repository.BankRow{Nama: "QRTP"}, want: false},
		{name: "qrtp decorated name", row: repository.BankRow{Nama: "Bank QRTP Internal"}, want: false},
		{name: "admin staff bank", row: repository.BankRow{Nama: "BCA H2H", AdminStaffOnly: true}, want: false},
		{name: "blank name", row: repository.BankRow{Nama: "  "}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isMemberDepositBank(tt.row); got != tt.want {
				t.Fatalf("isMemberDepositBank(%+v) = %v, want %v", tt.row, got, tt.want)
			}
		})
	}
}
