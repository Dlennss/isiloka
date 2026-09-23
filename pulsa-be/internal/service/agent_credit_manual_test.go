package service

import (
	"context"
	"strings"
	"testing"

	"pulsa2/internal/helper"
)

func TestCreateManualApplicationRejectsUnauthorizedRole(t *testing.T) {
	svc := NewAgentCreditService(nil)
	_, err := svc.CreateManualApplication(context.Background(), helper.AuthInfo{MemberID: 1, Role: helper.RoleRetailAgent}, AgentCreditManualInput{
		MemberID:        2,
		RequestedAmount: minimumAgentCreditAmount,
	})
	if err == nil || !strings.Contains(err.Error(), "operator") {
		t.Fatalf("expected operator-only error, got %v", err)
	}
}

func TestCreateManualApplicationValidatesAmountBeforeRepository(t *testing.T) {
	svc := NewAgentCreditService(nil)
	_, err := svc.CreateManualApplication(context.Background(), helper.AuthInfo{MemberID: 1, Role: helper.RoleRetailAnalyst}, AgentCreditManualInput{
		MemberID:        2,
		RequestedAmount: minimumAgentCreditAmount - 1,
	})
	if err == nil || !strings.Contains(err.Error(), "Rp500.000") {
		t.Fatalf("expected amount validation error, got %v", err)
	}
}
