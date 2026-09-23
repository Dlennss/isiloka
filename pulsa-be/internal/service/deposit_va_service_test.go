package service

import "testing"

func TestParseDepositVATicketAmount(t *testing.T) {
	got, err := parseDepositVATicketAmount("10002011")
	if err != nil {
		t.Fatalf("parseDepositVATicketAmount returned error: %v", err)
	}
	if got != 10002011 {
		t.Fatalf("amount = %d, want 10002011", got)
	}
}

func TestParseDepositVATicketAmountRejectsNonNumeric(t *testing.T) {
	if _, err := parseDepositVATicketAmount("ABC123"); err == nil {
		t.Fatalf("parseDepositVATicketAmount expected error")
	}
}
