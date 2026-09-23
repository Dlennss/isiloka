package controller

import "pulsa2/internal/service"

type AuditController struct {
	svc *service.AuditService
}

func NewAuditController(svc *service.AuditService) *AuditController {
	return &AuditController{svc: svc}
}
