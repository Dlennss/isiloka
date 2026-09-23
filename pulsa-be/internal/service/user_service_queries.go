package service

import (
	"context"
	"errors"

	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
)

func (s *UserService) List(ctx context.Context, q, role, scope string, limit, offset int) ([]repository.UserRow, error) {
	role = helper.NormalizeRole(role)
	scope = normalizeScope(scope)
	if scope == "__invalid__" {
		return nil, errors.New("scope tidak valid")
	}
	if role != "" && !isValidManageableRole(role) {
		return nil, errors.New("role tidak valid")
	}
	if role != "" && !s.repo.ScopeHasRole(scope, role) {
		return nil, errors.New("role tidak cocok dengan scope")
	}
	return s.repo.List(ctx, q, role, scope, limit, offset)
}

func (s *UserService) SumSaldo(ctx context.Context, q, role, scope string) (int64, error) {
	role = helper.NormalizeRole(role)
	scope = normalizeScope(scope)
	if scope == "__invalid__" {
		return 0, errors.New("scope tidak valid")
	}
	if role != "" && !isValidManageableRole(role) {
		return 0, errors.New("role tidak valid")
	}
	if role != "" && !s.repo.ScopeHasRole(scope, role) {
		return 0, errors.New("role tidak cocok dengan scope")
	}
	return s.repo.SumSaldo(ctx, q, role, scope)
}

func (s *UserService) Count(ctx context.Context, q, role, scope string) (int64, error) {
	role = helper.NormalizeRole(role)
	scope = normalizeScope(scope)
	if scope == "__invalid__" {
		return 0, errors.New("scope tidak valid")
	}
	if role != "" && !isValidManageableRole(role) {
		return 0, errors.New("role tidak valid")
	}
	if role != "" && !s.repo.ScopeHasRole(scope, role) {
		return 0, errors.New("role tidak cocok dengan scope")
	}
	return s.repo.Count(ctx, q, role, scope)
}

func (s *UserService) Get(ctx context.Context, id int64) (*repository.UserRow, error) {
	return s.repo.Get(ctx, id)
}
