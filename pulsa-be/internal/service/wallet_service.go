package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
)

type WalletService struct {
	repo     *repository.WalletRepository
	bankRepo *repository.BankRepository
}

func NewWalletService(repo *repository.WalletRepository, bankRepo *repository.BankRepository) *WalletService {
	return &WalletService{repo: repo, bankRepo: bankRepo}
}

func (s *WalletService) AdminAdjustMemberWallet(ctx context.Context, actorID, memberID, amount int64, direction, note, reason string) (string, error) {
	direction = strings.TrimSpace(strings.ToLower(direction))
	note = strings.TrimSpace(note)
	reason = strings.TrimSpace(reason)

	if memberID <= 0 || amount <= 0 {
		return "", errors.New("member_id/amount invalid")
	}
	if direction != "credit" && direction != "debit" {
		return "", errors.New("direction must be credit or debit")
	}
	if note == "" {
		return "", errors.New("note required")
	}
	if reason == "" {
		reason = "admin manual " + direction
	}

	refID := "ADM-" + time.Now().Format("20060102150405") + "-" + helper.RandHex(4)
	if err := s.repo.AdminAdjustMemberWallet(ctx, actorID, memberID, amount, direction, reason, note, refID); err != nil {
		return "", err
	}
	return refID, nil
}

type ProviderWalletRow struct {
	Provider      string     `json:"provider"`
	SaldoInternal int64      `json:"saldo_internal"`
	SaldoProvider int64      `json:"saldo_provider"`
	SnapshotAt    *time.Time `json:"snapshot_at,omitempty"`
	Selisih       int64      `json:"selisih"`
}

func (s *WalletService) ListProviderWallets(ctx context.Context) ([]ProviderWalletRow, error) {
	providers, err := s.repo.ListWalletProviderNames(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]ProviderWalletRow, 0, len(providers))

	for _, p := range providers {
		saldoInt, err := s.repo.GetProviderSaldo(ctx, p)
		if err != nil {
			return nil, err
		}
		snap, ok, err := s.repo.GetProviderLatestSnapshot(ctx, p)
		if err != nil {
			return nil, err
		}
		var tsPtr *time.Time
		var saldoProv int64
		selisih := int64(0)
		if ok {
			saldoProv = snap.Saldo
			t := snap.SnapshotAt
			tsPtr = &t
			selisih = saldoProv - saldoInt
		}
		out = append(out, ProviderWalletRow{
			Provider:      p,
			SaldoInternal: saldoInt,
			SaldoProvider: saldoProv,
			SnapshotAt:    tsPtr,
			Selisih:       selisih,
		})
	}
	return out, nil
}

func (s *WalletService) DepositProviderWallet(ctx context.Context, actorID int64, actorRole string, bankID int64, provider string, amount int64, adminFee int64, note string) (refID string, bankSaldo int64, saldoInternal int64, err error) {
	provider = strings.TrimSpace(strings.ToLower(provider))
	if bankID <= 0 || provider == "" || amount <= 0 {
		return "", 0, 0, errors.New("bank_id, provider & amount required")
	}
	if adminFee < 0 {
		return "", 0, 0, errors.New("admin fee invalid")
	}
	return "", 0, 0, errors.New("top up provider harus dari mutasi bank rekening; gunakan menu Mutasi Bank lalu assign debit bank ke provider")
}

type ProviderAdjustResult struct {
	Mode          string `json:"mode"`
	Arah          string `json:"arah"`
	AmountApplied int64  `json:"amount_applied"`
	RefID         string `json:"refid"`
	SaldoInternal int64  `json:"saldo_internal"`
	Unchanged     bool   `json:"unchanged,omitempty"`
}

func (s *WalletService) AdjustProviderWallet(ctx context.Context, actorID int64, provider string, amount int64, mode, note string) (*ProviderAdjustResult, error) {
	provider = strings.TrimSpace(strings.ToLower(provider))
	mode = strings.TrimSpace(strings.ToLower(mode))
	note = strings.TrimSpace(note)

	if provider == "" || amount < 0 {
		return nil, errors.New("provider & amount required")
	}
	if mode == "" {
		mode = "set"
	}
	if mode != "set" && mode != "credit" && mode != "debit" {
		return nil, errors.New("mode invalid")
	}
	if mode != "set" && amount == 0 {
		return nil, errors.New("amount must be > 0")
	}

	current, err := s.repo.GetProviderSaldo(ctx, provider)
	if err != nil {
		return nil, err
	}

	arah := mode
	jumlah := amount
	if mode == "set" {
		diff := amount - current
		if diff == 0 {
			return &ProviderAdjustResult{
				Mode:          "set",
				Unchanged:     true,
				SaldoInternal: current,
			}, nil
		}
		if diff > 0 {
			arah = "credit"
			jumlah = diff
		} else {
			arah = "debit"
			jumlah = -diff
		}
	}

	refID := "PADJ-" + time.Now().Format("20060102-150405")
	alasan := "PROVIDER_ADJUST_" + strings.ToUpper(mode)
	if mode == "set" {
		alasan = "PROVIDER_SET_BALANCE"
	}

	_, after, err := s.repo.ApplyProviderTx(ctx, repository.ProviderWalletTxIn{
		Provider:   provider,
		RefID:      refID,
		Arah:       arah,
		Jumlah:     jumlah,
		Alasan:     alasan,
		Catatan:    note,
		DiubahOleh: &actorID,
	})
	if err != nil {
		return nil, err
	}

	return &ProviderAdjustResult{
		Mode:          mode,
		Arah:          arah,
		AmountApplied: jumlah,
		RefID:         refID,
		SaldoInternal: after,
	}, nil
}

func (s *WalletService) ProviderHistory(ctx context.Context, provider, arah, refID, from, to string, limit, offset int) ([]repository.ProviderWalletMutasiRow, int64, error) {
	return s.repo.ListProviderMutasi(ctx, provider, arah, refID, from, to, limit, offset)
}

func (s *WalletService) MemberDepositHistory(ctx context.Context, memberID int64, arah, refID, from, to string, limit, offset int) ([]repository.MemberWalletMutasiRow, int64, error) {
	return s.repo.ListMemberDepositMutasi(ctx, memberID, arah, refID, from, to, limit, offset)
}

func (s *WalletService) ProviderDepositHistory(ctx context.Context, provider, refID, from, to string, limit, offset int) ([]repository.ProviderWalletMutasiRow, int64, error) {
	return s.repo.ListProviderDepositMutasi(ctx, provider, refID, from, to, limit, offset)
}

type WalletDashboardActivity struct {
	From    string                                 `json:"from"`
	To      string                                 `json:"to"`
	Summary *repository.WalletActorActivitySummary `json:"summary"`
	Recent  []repository.WalletActorActivityRow    `json:"recent"`
}

func (s *WalletService) ActorActivity(ctx context.Context, actorID int64, from, to string, limit int) (*WalletDashboardActivity, error) {
	if actorID <= 0 {
		return nil, errors.New("actor invalid")
	}
	loc, _ := time.LoadLocation("Asia/Jakarta")

	var fromTime time.Time
	var toTime time.Time
	if strings.TrimSpace(from) == "" {
		now := time.Now().In(loc)
		fromTime = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
	} else {
		t, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(from), loc)
		if err != nil {
			return nil, errors.New("from invalid")
		}
		fromTime = t
	}
	if strings.TrimSpace(to) == "" {
		now := time.Now().In(loc)
		toTime = now.AddDate(0, 0, 1)
	} else {
		t, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(to), loc)
		if err != nil {
			return nil, errors.New("to invalid")
		}
		toTime = t.AddDate(0, 0, 1)
	}
	if !toTime.After(fromTime) {
		return nil, errors.New("range invalid")
	}

	summary, err := s.repo.WalletActorActivitySummary(ctx, actorID, fromTime, toTime)
	if err != nil {
		return nil, err
	}
	recent, err := s.repo.WalletActorRecentActivity(ctx, actorID, fromTime, toTime, limit)
	if err != nil {
		return nil, err
	}

	return &WalletDashboardActivity{
		From:    fromTime.Format("2006-01-02"),
		To:      toTime.AddDate(0, 0, -1).Format("2006-01-02"),
		Summary: summary,
		Recent:  recent,
	}, nil
}
