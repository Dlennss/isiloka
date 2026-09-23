package repository

import (
	"database/sql"
	"strings"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func rolesForScope(scope string) []string {
	switch strings.TrimSpace(strings.ToLower(scope)) {
	case "h2h":
		return []string{"member", "agent_member", "master_member"}
	case "retail":
		return []string{"user", "agent", "master", "marketing", "analis"}
	case "internal":
		return []string{"admin", "staff", "auditor", "operator_trx", "operator_wallet"}
	default:
		return nil
	}
}
func (r *UserRepository) ScopeHasRole(scope, role string) bool {
	scopeRoles := rolesForScope(scope)
	if len(scopeRoles) == 0 {
		return true
	}
	needle := strings.TrimSpace(strings.ToLower(role))
	for _, item := range scopeRoles {
		if item == needle {
			return true
		}
	}
	return false
}
