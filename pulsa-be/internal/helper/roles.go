package helper

import "strings"

const (
	RoleAdmin           = "admin"
	RoleStaff           = "staff"
	RoleAuditor         = "auditor"
	RoleMember          = "member"
	RoleUser            = "user"
	RoleRetailAgent     = "agent"
	RoleRetailMaster    = "master"
	RoleRetailMarketing = "marketing"
	RoleRetailAnalyst   = "analis"
	RoleH2HAgent        = "agent_member"
	RoleH2HMaster       = "master_member"
	RoleOperatorTrx     = "operator_trx"
	RoleOperatorWallet  = "operator_wallet"

	DefaultRetailCommissionRp int64 = 100
)

func NormalizeRole(role string) string {
	role = strings.TrimSpace(strings.ToLower(role))
	switch role {
	case "agen-member", "agent-member":
		return RoleH2HAgent
	case "master-member", "mester-member":
		return RoleH2HMaster
	case "analyst", "analis kredit", "credit_analyst", "credit-analyst", "operator", "operator kredit", "operator_kredit", "operator_credit", "operator-credit":
		return RoleRetailAnalyst
	default:
		return role
	}
}

func IsAdminLikeRole(role string) bool {
	switch NormalizeRole(role) {
	case RoleAdmin, RoleStaff:
		return true
	default:
		return false
	}
}

func IsRetailRole(role string) bool {
	switch NormalizeRole(role) {
	case RoleUser, RoleRetailAgent, RoleRetailMaster, RoleRetailMarketing, RoleRetailAnalyst:
		return true
	default:
		return false
	}
}

func IsH2HRole(role string) bool {
	switch NormalizeRole(role) {
	case RoleMember, RoleH2HAgent, RoleH2HMaster:
		return true
	default:
		return false
	}
}

func ApplyRetailCommissionDefaults(role string, agentCommissionRp, masterCommissionRp int64) (int64, int64) {
	switch NormalizeRole(role) {
	case RoleRetailAgent, RoleRetailMaster:
		if agentCommissionRp == 0 {
			agentCommissionRp = DefaultRetailCommissionRp
		}
		if masterCommissionRp == 0 {
			masterCommissionRp = DefaultRetailCommissionRp
		}
	}
	return agentCommissionRp, masterCommissionRp
}
