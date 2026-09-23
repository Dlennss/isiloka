package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
)

type H2HService struct {
	repo     *repository.H2HRepository
	authRepo *repository.AuthRepository
	bankRepo *repository.BankRepository
}

type H2HRegisterDownlineInput struct {
	Email       string
	Nama        string
	Password    string
	PIN         string
	Role        string
	InitialFees repository.H2HInitialFeeSetupInput
}

type H2HWithdrawCreateInput struct {
	Amount        int64
	BankName      string
	AccountName   string
	AccountNumber string
	Note          string
}

const (
	h2hWithdrawSourceAccountNumber = "8761518267"
	h2hWithdrawSourceLabel         = "BCA H2H LISA OKTARIA 8761518267"
)

func NewH2HService(repo *repository.H2HRepository, authRepo *repository.AuthRepository, bankRepo *repository.BankRepository) *H2HService {
	return &H2HService{repo: repo, authRepo: authRepo, bankRepo: bankRepo}
}

func (s *H2HService) ListDownlines(ctx context.Context, actorID int64) ([]repository.H2HDownlineRow, error) {
	actor, err := s.repo.GetMemberContext(ctx, actorID)
	if err != nil {
		return nil, err
	}
	if !helper.IsH2HRole(actor.Role) {
		return nil, errors.New("h2h only")
	}
	if actor.Role != helper.RoleH2HMaster && actor.Role != helper.RoleH2HAgent {
		return []repository.H2HDownlineRow{}, nil
	}
	return s.repo.ListDownlines(ctx, actor)
}

func (s *H2HService) RegisterDownline(ctx context.Context, actorID int64, in H2HRegisterDownlineInput) (int64, string, error) {
	actor, err := s.repo.GetMemberContext(ctx, actorID)
	if err != nil {
		return 0, "", err
	}
	if !helper.IsH2HRole(actor.Role) {
		return 0, "", errors.New("h2h only")
	}

	in.Email = strings.TrimSpace(strings.ToLower(in.Email))
	in.Nama = strings.TrimSpace(in.Nama)
	in.Password = strings.TrimSpace(in.Password)
	in.PIN = strings.TrimSpace(in.PIN)
	in.Role = helper.NormalizeRole(in.Role)

	if in.Email == "" || in.Nama == "" || len(in.Password) < 8 {
		return 0, "", errors.New("email, nama, dan password valid wajib diisi")
	}
	if len(in.PIN) < 4 || len(in.PIN) > 12 {
		return 0, "", errors.New("pin harus 4-12 karakter")
	}
	if in.Role != helper.RoleMember && in.Role != helper.RoleH2HAgent {
		return 0, "", errors.New("role bawahan H2H tidak valid")
	}
	switch actor.Role {
	case helper.RoleH2HMaster:
		// master_member boleh buat agent_member atau member
	case helper.RoleH2HAgent:
		if in.Role != helper.RoleMember {
			return 0, "", errors.New("agent member hanya boleh menambahkan member")
		}
	default:
		return 0, "", errors.New("role tidak boleh menambahkan downline")
	}

	passHash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, "", err
	}
	pinHash, err := bcrypt.GenerateFromPassword([]byte(in.PIN), bcrypt.DefaultCost)
	if err != nil {
		return 0, "", err
	}

	createIn := repository.UserCreateInput{
		Email:        in.Email,
		Nama:         in.Nama,
		PasswordHash: string(passHash),
		PinHash:      string(pinHash),
		Role:         in.Role,
		APIKey:       helper.RandHex(32),
		Label:        "default",
		Aktif:        false,
	}
	switch actor.Role {
	case helper.RoleH2HMaster:
		createIn.H2HMasterID = &actor.MemberID
	case helper.RoleH2HAgent:
		createIn.H2HAgentID = &actor.MemberID
		if actor.H2HMasterID != nil && *actor.H2HMasterID > 0 {
			createIn.H2HMasterID = actor.H2HMasterID
		}
	}

	return s.authRepo.CreateMemberWithKeyAndInitialFees(ctx, createIn, in.InitialFees)
}

func (s *H2HService) ListCommissions(ctx context.Context, actorID int64, limit, offset int) ([]repository.H2HCommissionLedgerRow, error) {
	actor, err := s.repo.GetMemberContext(ctx, actorID)
	if err != nil {
		return nil, err
	}
	if !helper.IsH2HRole(actor.Role) {
		return nil, errors.New("h2h only")
	}
	return s.repo.ListCommissionLedger(ctx, actorID, limit, offset)
}

func (s *H2HService) CommissionSummary(ctx context.Context, actorID int64) (*repository.H2HCommissionSummaryRow, error) {
	actor, err := s.repo.GetMemberContext(ctx, actorID)
	if err != nil {
		return nil, err
	}
	if !helper.IsH2HRole(actor.Role) {
		return nil, errors.New("h2h only")
	}
	return s.repo.GetCommissionSummary(ctx, actorID)
}

func (s *H2HService) CreateWithdrawRequest(ctx context.Context, actorID int64, in H2HWithdrawCreateInput) (*repository.H2HWithdrawRequestRow, error) {
	actor, err := s.repo.GetMemberContext(ctx, actorID)
	if err != nil {
		return nil, err
	}
	if !helper.IsH2HRole(actor.Role) {
		return nil, errors.New("h2h only")
	}

	in.BankName = strings.TrimSpace(in.BankName)
	in.AccountName = strings.TrimSpace(in.AccountName)
	in.AccountNumber = strings.TrimSpace(in.AccountNumber)
	in.Note = strings.TrimSpace(in.Note)
	if in.Amount <= 0 {
		return nil, errors.New("amount harus > 0")
	}
	if in.BankName == "" || in.AccountName == "" || in.AccountNumber == "" {
		return nil, errors.New("data rekening wajib lengkap")
	}
	return s.repo.CreateWithdrawRequest(ctx, actorID, in.Amount, in.BankName, in.AccountName, in.AccountNumber, s.repo.NewWithdrawRefID(), in.Note)
}

func (s *H2HService) ListOwnWithdrawRequests(ctx context.Context, actorID int64, limit, offset int) ([]repository.H2HWithdrawRequestRow, error) {
	actor, err := s.repo.GetMemberContext(ctx, actorID)
	if err != nil {
		return nil, err
	}
	if !helper.IsH2HRole(actor.Role) {
		return nil, errors.New("h2h only")
	}
	return s.repo.ListWithdrawRequestsByMember(ctx, actorID, limit, offset)
}

func (s *H2HService) AdminListWithdrawRequests(ctx context.Context, status, q string, limit, offset int) ([]repository.H2HWithdrawRequestRow, error) {
	return s.repo.AdminListWithdrawRequests(ctx, status, q, limit, offset)
}

func (s *H2HService) AdminApproveWithdrawRequest(ctx context.Context, reqID, actorID int64, _ string, _ int64, fee int64, note string) error {
	if reqID <= 0 || actorID <= 0 {
		return errors.New("request invalid")
	}
	if fee < 0 {
		return errors.New("fee tidak valid")
	}
	if s.bankRepo == nil {
		return errors.New("bank repository tidak tersedia")
	}
	sourceBank, err := s.bankRepo.GetActiveByAccountNumber(ctx, h2hWithdrawSourceAccountNumber, true)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("rekening sumber H2H " + h2hWithdrawSourceLabel + " tidak ditemukan atau belum aktif")
		}
		return err
	}
	err = s.repo.ApproveWithdrawRequest(ctx, reqID, actorID, sourceBank.ID, fee, strings.TrimSpace(note))
	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("withdraw request tidak ditemukan atau bukan pending")
	}
	return err
}

func (s *H2HService) AdminRejectWithdrawRequest(ctx context.Context, reqID, actorID int64, reason string) error {
	if reqID <= 0 || actorID <= 0 {
		return errors.New("request invalid")
	}
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return errors.New("alasan reject wajib diisi")
	}
	err := s.repo.RejectWithdrawRequest(ctx, reqID, actorID, reason)
	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("withdraw request tidak ditemukan")
	}
	return err
}
