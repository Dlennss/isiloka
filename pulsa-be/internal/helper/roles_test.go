package helper

import "testing"

func TestNormalizeRoleCreditOperatorAliases(t *testing.T) {
	for _, role := range []string{"operator", "operator kredit", "operator_kredit", "operator_credit", "operator-credit", "analis", "analyst"} {
		if got := NormalizeRole(role); got != RoleRetailAnalyst {
			t.Fatalf("NormalizeRole(%q) = %q, want %q", role, got, RoleRetailAnalyst)
		}
	}
}

func TestApplyRetailCommissionDefaults(t *testing.T) {
	tests := []struct {
		name       string
		role       string
		agentIn    int64
		masterIn   int64
		agentWant  int64
		masterWant int64
	}{
		{name: "agent defaults both retail commissions", role: RoleRetailAgent, agentWant: DefaultRetailCommissionRp, masterWant: DefaultRetailCommissionRp},
		{name: "master defaults both retail commissions", role: RoleRetailMaster, agentWant: DefaultRetailCommissionRp, masterWant: DefaultRetailCommissionRp},
		{name: "agent keeps explicit commission", role: RoleRetailAgent, agentIn: 250, agentWant: 250, masterWant: DefaultRetailCommissionRp},
		{name: "master keeps explicit commission", role: RoleRetailMaster, masterIn: 300, agentWant: DefaultRetailCommissionRp, masterWant: 300},
		{name: "user unchanged", role: RoleUser},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agentGot, masterGot := ApplyRetailCommissionDefaults(tt.role, tt.agentIn, tt.masterIn)
			if agentGot != tt.agentWant || masterGot != tt.masterWant {
				t.Fatalf("got agent=%d master=%d, want agent=%d master=%d", agentGot, masterGot, tt.agentWant, tt.masterWant)
			}
		})
	}
}
