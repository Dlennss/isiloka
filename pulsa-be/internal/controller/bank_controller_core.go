package controller

import (
	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
	"pulsa2/internal/service"
)

type BankController struct {
	svc *service.BankService
}

func NewBankController(svc *service.BankService) *BankController {
	return &BankController{svc: svc}
}

func sanitizeBankRowsForOperatorTrx(items []repository.BankRow) []repository.BankRow {
	out := make([]repository.BankRow, 0, len(items))
	for _, item := range items {
		item.Saldo = 0
		out = append(out, item)
	}
	return out
}

func sanitizeBankRowForOperatorTrx(item *repository.BankRow) *repository.BankRow {
	if item == nil {
		return nil
	}
	cp := *item
	cp.Saldo = 0
	return &cp
}

func bankAuthRole(role string) string {
	role = helper.NormalizeRole(role)
	if helper.IsAdminLikeRole(role) {
		return helper.RoleAdmin
	}
	return role
}
