package service

import (
	"context"
	"strings"

	"pulsa2/internal/repository"
)

type ProviderReportingService struct {
	repo *repository.ProviderReportingRepository
}

func NewProviderReportingService(repo *repository.ProviderReportingRepository) *ProviderReportingService {
	return &ProviderReportingService{repo: repo}
}

func (s *ProviderReportingService) ListTransactions(ctx context.Context, in repository.ProviderTransactionListArgs) ([]repository.ProviderTransactionRecord, int64, error) {
	in.Provider = strings.TrimSpace(in.Provider)
	in.Q = strings.TrimSpace(in.Q)
	in.KodeProduk = strings.TrimSpace(in.KodeProduk)
	in.RefID = strings.TrimSpace(in.RefID)
	in.Tujuan = strings.TrimSpace(in.Tujuan)
	in.KodeRespon = strings.TrimSpace(in.KodeRespon)
	in.Status = strings.ToLower(strings.TrimSpace(in.Status))
	return s.repo.ListTransactions(ctx, in)
}

func (s *ProviderReportingService) Analytics(ctx context.Context, in repository.ProviderAnalyticsArgs) (repository.ProviderAnalyticsBundle, error) {
	return s.repo.Analytics(ctx, in)
}

func (s *ProviderReportingService) RefreshAnalyticsCache(ctx context.Context, months, days int) (*repository.ProviderAnalyticsCacheRefreshResult, error) {
	if days <= 0 {
		if months <= 0 || months > 12 {
			months = 3
		}
		days = months * 31
	}
	if days <= 0 || days > 366 {
		days = 93
	}
	return s.repo.RefreshAnalyticsCache(ctx, days)
}

func (s *ProviderReportingService) DailyProductSuccess(ctx context.Context, in repository.DailyProductSuccessArgs) (repository.DailyProductSuccessBundle, error) {
	in.Q = strings.TrimSpace(in.Q)
	return s.repo.DailyProductSuccess(ctx, in)
}

func (s *ProviderReportingService) ListAnomalies(ctx context.Context, in repository.ProviderAnomalyListArgs) ([]repository.ProviderAnomalyRecord, int64, error) {
	in.Provider = strings.TrimSpace(strings.ToLower(in.Provider))
	in.Q = strings.TrimSpace(in.Q)
	in.RefID = strings.TrimSpace(in.RefID)
	in.KodeRespon = strings.TrimSpace(in.KodeRespon)
	in.Status = strings.TrimSpace(strings.ToLower(in.Status))
	return s.repo.ListAnomalies(ctx, in)
}
