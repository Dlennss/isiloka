package controller

import (
	"pulsa2/internal/service"
)

type AuthController struct {
	svc *service.AuthService
}

func derefInt64(v *int64) int64 {
	if v == nil {
		return 0
	}
	return *v
}

func NewAuthController(svc *service.AuthService) *AuthController {
	return &AuthController{svc: svc}
}
