package service

import (
	"context"
	"strings"

	"pulsa2/internal/repository"
)

type AdminBusinessReportService struct {
	repo *repository.AdminBusinessReportRepository
}

func NewAdminBusinessReportService(repo *repository.AdminBusinessReportRepository) *AdminBusinessReportService {
	return &AdminBusinessReportService{repo: repo}
}

func (s *AdminBusinessReportService) ListCommissionBySource(ctx context.Context, in repository.AdminCommissionBySourceArgs) ([]repository.AdminCommissionBySourceRow, error) {
	in.Scope = strings.ToLower(strings.TrimSpace(in.Scope))
	return s.repo.ListCommissionBySource(ctx, in)
}

func (s *AdminBusinessReportService) ListDailyBusiness(ctx context.Context, in repository.AdminDailyBusinessArgs) ([]repository.AdminDailyBusinessRow, error) {
	in.Scope = strings.ToLower(strings.TrimSpace(in.Scope))
	return s.repo.ListDailyBusiness(ctx, in)
}

func (s *AdminBusinessReportService) RefreshDailyBusinessCache(ctx context.Context, months, days int) (*repository.AdminDailyBusinessCacheRefreshResult, error) {
	if days <= 0 {
		if months <= 0 || months > 12 {
			months = 3
		}
		days = months * 31
	}
	if days <= 0 || days > 366 {
		days = 93
	}
	return s.repo.RefreshDailyBusinessCache(ctx, days)
}
