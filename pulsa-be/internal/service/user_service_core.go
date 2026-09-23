package service

import (
	"strings"

	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
)

type UserService struct {
	repo       *repository.UserRepository
	retailRepo *repository.RetailRepository
	h2hRepo    *repository.H2HRepository
}

func NewUserService(repo *repository.UserRepository, retailRepo *repository.RetailRepository, h2hRepo *repository.H2HRepository) *UserService {
	return &UserService{repo: repo, retailRepo: retailRepo, h2hRepo: h2hRepo}
}

func isValidManageableRole(role string) bool {
	return role == helper.RoleAdmin ||
		role == helper.RoleStaff ||
		role == helper.RoleAuditor ||
		role == helper.RoleMember ||
		role == helper.RoleUser ||
		role == helper.RoleRetailAgent ||
		role == helper.RoleRetailMaster ||
		role == helper.RoleRetailMarketing ||
		role == helper.RoleRetailAnalyst ||
		role == helper.RoleH2HAgent ||
		role == helper.RoleH2HMaster ||
		role == helper.RoleOperatorTrx ||
		role == helper.RoleOperatorWallet
}

func normalizeScope(scope string) string {
	scope = strings.TrimSpace(strings.ToLower(scope))
	switch scope {
	case "", "h2h", "retail", "internal":
		return scope
	default:
		return "__invalid__"
	}
}
