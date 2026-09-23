package service

import (
	"context"
	"strings"
	"testing"

	"pulsa2/internal/helper"
)

func TestSetLoanOperationalStatusRequiresSuperAdmin(t *testing.T) {
	svc := NewAgentCreditService(nil)
	_, err := svc.SetLoanOperationalStatus(context.Background(), helper.AuthInfo{MemberID: 10, Role: helper.RoleRetailAnalyst}, AgentCreditLoanStatusInput{
		ApplicationID: 1,
		Suspended:     true,
		Reason:        "Pemeriksaan risiko",
	})
	if err == nil || !strings.Contains(err.Error(), "super admin") {
		t.Fatalf("error = %v, want super admin only", err)
	}
}

func TestSetLoanOperationalStatusRequiresReason(t *testing.T) {
	svc := NewAgentCreditService(nil)
	_, err := svc.SetLoanOperationalStatus(context.Background(), helper.AuthInfo{MemberID: 10, Role: helper.RoleAdmin}, AgentCreditLoanStatusInput{
		ApplicationID: 1,
		Suspended:     true,
	})
	if err == nil || !strings.Contains(err.Error(), "alasan") {
		t.Fatalf("error = %v, want required reason", err)
	}
}

func TestManageableRolesIncludeMarketingAndCreditOperator(t *testing.T) {
	for _, role := range []string{helper.RoleRetailMarketing, helper.RoleRetailAnalyst} {
		if !isValidManageableRole(role) {
			t.Fatalf("role %q should be manageable by super admin", role)
		}
	}
}

func TestListTeamActivityRequiresSuperAdmin(t *testing.T) {
	svc := NewAgentCreditService(nil)
	_, err := svc.ListTeamActivity(context.Background(), helper.AuthInfo{MemberID: 10, Role: helper.RoleRetailAnalyst}, "", 50)
	if err == nil || !strings.Contains(err.Error(), "super admin") {
		t.Fatalf("error = %v, want super admin only", err)
	}
}
