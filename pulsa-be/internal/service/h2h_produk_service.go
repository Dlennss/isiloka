package service

import (
	"context"
	"errors"
	"strings"

	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
)

type H2HProdukService struct {
	authRepo *repository.MemberTrxMemberRepository
	repo     *repository.H2HProdukRepository
}

func NewH2HProdukService(authRepo *repository.MemberTrxMemberRepository, repo *repository.H2HProdukRepository) *H2HProdukService {
	return &H2HProdukService{authRepo: authRepo, repo: repo}
}

func (s *H2HProdukService) List(ctx context.Context, apiKey, q, kategoriName, brandName string) ([]repository.H2HProdukRow, error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil, &ServiceError{Kind: ErrUnauthorized, Message: "missing X-Api-Key"}
	}

	auth, err := s.authRepo.AuthByAPIKey(ctx, apiKey)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrUnauthorized):
			return nil, &ServiceError{Kind: ErrUnauthorized, Message: "unauthorized"}
		case errors.Is(err, repository.ErrForbidden):
			return nil, &ServiceError{Kind: ErrForbidden, Message: "forbidden"}
		default:
			return nil, err
		}
	}
	if !helper.IsH2HRole(auth.Role) {
		return nil, &ServiceError{Kind: ErrForbidden, Message: "h2h only"}
	}

	rows, err := s.repo.ListByMember(ctx, auth.MemberID, q, kategoriName, brandName)
	if err != nil {
		return nil, err
	}
	return rows, nil
}
