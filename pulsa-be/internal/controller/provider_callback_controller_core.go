package controller

import "pulsa2/internal/service"

type ProviderCallbackController struct {
	svc *service.ProviderCallbackService
}

func NewProviderCallbackController(svc *service.ProviderCallbackService) *ProviderCallbackController {
	return &ProviderCallbackController{svc: svc}
}
