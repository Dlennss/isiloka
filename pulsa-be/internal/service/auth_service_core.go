package service

import (
	"pulsa2/internal/repository"
)

type AuthService struct {
	repo      *repository.AuthRepository
	orderRepo *repository.AppOrderRepository
	jwtSecret []byte
}

type RegisterPublicGuestRefundInput struct {
	InvoiceID  string
	GuestEmail string
	GuestPhone string
}

type googleTokenInfo struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified string `json:"email_verified"`
	Aud           string `json:"aud"`
}

func NewAuthService(repo *repository.AuthRepository, orderRepo *repository.AppOrderRepository, jwtSecret []byte) *AuthService {
	return &AuthService{repo: repo, orderRepo: orderRepo, jwtSecret: jwtSecret}
}
