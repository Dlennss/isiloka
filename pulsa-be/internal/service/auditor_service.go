package service

import (
	"context"
	"strings"
	"time"

	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
)

type AuditorService struct {
	repo     *repository.AuditorRepository
	bankRepo *repository.BankRepository
}

func NewAuditorService(repo *repository.AuditorRepository, bankRepo *repository.BankRepository) *AuditorService {
	return &AuditorService{repo: repo, bankRepo: bankRepo}
}

func (s *AuditorService) CreateInternalFinanceEntry(ctx context.Context, actorID int64, actorRole string, in repository.InternalFinanceCreateInput) (*repository.InternalFinanceEntryRow, error) {
	in.EntryType = strings.TrimSpace(strings.ToLower(in.EntryType))
	in.Category = strings.TrimSpace(in.Category)
	in.Counterparty = strings.TrimSpace(in.Counterparty)
	in.Note = strings.TrimSpace(in.Note)
	if in.BankID > 0 {
		if _, err := s.bankRepo.GetVisible(ctx, in.BankID, helper.IsAdminLikeRole(actorRole)); err != nil {
			return nil, err
		}
	}
	return s.repo.CreateInternalFinanceEntry(ctx, actorID, in)
}

func (s *AuditorService) ListInternalFinanceEntries(ctx context.Context, role string, bankID int64, entryType, from, to string, limit, offset int) ([]repository.InternalFinanceEntryRow, int64, error) {
	return s.repo.ListInternalFinanceEntries(ctx, bankID, strings.TrimSpace(strings.ToLower(entryType)), strings.TrimSpace(from), strings.TrimSpace(to), limit, offset, helper.IsAdminLikeRole(role))
}

func (s *AuditorService) UpsertOpeningBalances(ctx context.Context, actorID int64, periodMonth time.Time, items []repository.OpeningBalanceUpsertInput) error {
	for i := range items {
		items[i].AccountCode = strings.TrimSpace(strings.ToUpper(items[i].AccountCode))
		items[i].AccountName = strings.TrimSpace(items[i].AccountName)
		items[i].AccountGroup = strings.TrimSpace(strings.ToLower(items[i].AccountGroup))
		items[i].Note = strings.TrimSpace(items[i].Note)
	}
	return s.repo.UpsertOpeningBalances(ctx, actorID, periodMonth, items)
}

func (s *AuditorService) ListOpeningBalances(ctx context.Context, periodMonth time.Time) ([]repository.OpeningBalanceRow, error) {
	return s.repo.ListOpeningBalances(ctx, periodMonth)
}

func (s *AuditorService) ListTradingSummary(ctx context.Context, in repository.AuditorTradingSummaryArgs) ([]repository.AuditorTradingSummaryRow, error) {
	in.Scope = strings.TrimSpace(strings.ToLower(in.Scope))
	in.Period = strings.TrimSpace(strings.ToLower(in.Period))
	return s.repo.ListTradingSummary(ctx, in)
}

func (s *AuditorService) ListTradingDetails(ctx context.Context, in repository.AuditorTradingDetailArgs) ([]repository.AuditorTradingDetailRow, int64, error) {
	in.Scope = strings.TrimSpace(strings.ToLower(in.Scope))
	in.RefID = strings.TrimSpace(in.RefID)
	return s.repo.ListTradingDetails(ctx, in)
}

func (s *AuditorService) ListFinance(ctx context.Context, role string, in repository.AuditorFinanceArgs) ([]repository.AuditorFinanceRow, int64, error) {
	in.Type = strings.TrimSpace(strings.ToLower(in.Type))
	return s.repo.ListFinance(ctx, in, helper.IsAdminLikeRole(role))
}

func (s *AuditorService) GetProfitLoss(ctx context.Context, role string, in repository.AuditorProfitLossArgs) (*repository.AuditorProfitLossReport, error) {
	return s.repo.GetProfitLoss(ctx, in, helper.IsAdminLikeRole(role))
}
