package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
)

func (s *UserService) Create(ctx context.Context, email, nama, password, pin, role string, aktif bool) (int64, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	nama = strings.TrimSpace(nama)
	password = strings.TrimSpace(password)
	pin = strings.TrimSpace(pin)
	role = helper.NormalizeRole(role)
	if role == "" {
		role = helper.RoleMember
	}
	if role != helper.RoleAdmin &&
		role != helper.RoleStaff &&
		role != helper.RoleMember &&
		role != helper.RoleAuditor &&
		role != helper.RoleUser &&
		role != helper.RoleRetailAgent &&
		role != helper.RoleRetailMaster &&
		role != helper.RoleRetailMarketing &&
		role != helper.RoleRetailAnalyst &&
		role != helper.RoleH2HAgent &&
		role != helper.RoleH2HMaster &&
		role != helper.RoleOperatorTrx &&
		role != helper.RoleOperatorWallet {
		return 0, errors.New("role tidak valid")
	}
	if email == "" || password == "" {
		return 0, errors.New("email/password required")
	}
	if len(password) < 8 {
		return 0, errors.New("password min 8 chars")
	}
	if helper.IsH2HRole(role) && (len(pin) < 4 || len(pin) > 12) {
		return 0, errors.New("pin length 4-12")
	}

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}
	pinHash := ""
	if helper.IsH2HRole(role) {
		hash, hashErr := bcrypt.GenerateFromPassword([]byte(pin), bcrypt.DefaultCost)
		if hashErr != nil {
			return 0, hashErr
		}
		pinHash = string(hash)
	}

	retailAgentCommissionRp, retailMasterCommissionRp := helper.ApplyRetailCommissionDefaults(role, 0, 0)

	return s.repo.Create(ctx, repository.UserCreateInput{
		Email:                    email,
		Nama:                     nama,
		PasswordHash:             string(passHash),
		PinHash:                  pinHash,
		Role:                     role,
		APIKey:                   helper.RandHex(32),
		Label:                    "default",
		Aktif:                    aktif,
		RetailAgentCommissionRp:  retailAgentCommissionRp,
		RetailMasterCommissionRp: retailMasterCommissionRp,
	})
}

func (s *UserService) Update(ctx context.Context, in repository.UserUpdateInput, newPassword, newPin string) error {
	in.Email = strings.TrimSpace(strings.ToLower(in.Email))
	in.Nama = strings.TrimSpace(in.Nama)
	in.Phone = normalizeMemberPhone(in.Phone)
	in.Role = helper.NormalizeRole(in.Role)
	newPassword = strings.TrimSpace(newPassword)
	newPin = strings.TrimSpace(newPin)
	if in.ID <= 0 {
		return errors.New("id invalid")
	}
	if in.Email == "" {
		return errors.New("email required")
	}
	if in.Role == "" {
		in.Role = helper.RoleMember
	}
	if in.Role != helper.RoleAdmin &&
		in.Role != helper.RoleStaff &&
		in.Role != helper.RoleMember &&
		in.Role != helper.RoleAuditor &&
		in.Role != helper.RoleUser &&
		in.Role != helper.RoleRetailAgent &&
		in.Role != helper.RoleRetailMaster &&
		in.Role != helper.RoleRetailMarketing &&
		in.Role != helper.RoleRetailAnalyst &&
		in.Role != helper.RoleH2HAgent &&
		in.Role != helper.RoleH2HMaster &&
		in.Role != helper.RoleOperatorTrx &&
		in.Role != helper.RoleOperatorWallet {
		return errors.New("role tidak valid")
	}
	in.RetailAgentCommissionRp, in.RetailMasterCommissionRp = helper.ApplyRetailCommissionDefaults(in.Role, in.RetailAgentCommissionRp, in.RetailMasterCommissionRp)
	if in.FeeMemberRp < 0 || in.RetailAgentCommissionRp < 0 || in.RetailMasterCommissionRp < 0 || in.H2HAgentCommissionRp < 0 || in.H2HMasterCommissionRp < 0 {
		return errors.New("nilai fee/komisi tidak boleh negatif")
	}
	existing, err := s.repo.Get(ctx, in.ID)
	if err != nil {
		return err
	}
	if existing == nil {
		return sql.ErrNoRows
	}
	if helper.IsH2HRole(in.Role) && !helper.IsH2HRole(existing.Role) && newPin == "" {
		return errors.New("pin wajib diisi saat mengubah akun menjadi H2H")
	}

	if newPassword != "" {
		if len(newPassword) < 8 {
			return errors.New("password min 8 chars")
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		sv := string(hash)
		in.PasswordHash = &sv
	}
	if helper.IsH2HRole(in.Role) && newPin != "" {
		if len(newPin) < 4 || len(newPin) > 12 {
			return errors.New("pin length 4-12")
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(newPin), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		sv := string(hash)
		in.PinHash = &sv
	}
	return s.repo.Update(ctx, in)
}

func (s *UserService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.New("id invalid")
	}
	err := s.repo.Delete(ctx, id)
	if err == nil {
		return nil
	}
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23503" {
		return errors.New("akun tidak dapat dihapus karena masih memiliki riwayat transaksi, kredit, atau audit; nonaktifkan akun agar riwayat tetap aman")
	}
	return err
}

func (s *UserService) ListMarketingManagement(ctx context.Context) ([]repository.MarketingAccountRow, error) {
	accounts, err := s.repo.ListMarketingAccounts(ctx)
	if err != nil {
		return nil, err
	}
	return accounts, nil
}

func (s *UserService) UpdateMarketingAccount(ctx context.Context, id int64, nama, phone, newPassword string, aktif bool) error {
	row, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if row == nil || helper.NormalizeRole(row.Role) != helper.RoleRetailMarketing {
		return errors.New("akun marketing tidak ditemukan")
	}
	if strings.TrimSpace(nama) == "" {
		return errors.New("nama marketing wajib diisi")
	}
	if normalizeMemberPhone(phone) == "" {
		return errors.New("nomor telepon marketing wajib diisi")
	}
	return s.Update(ctx, repository.UserUpdateInput{
		ID: id, Email: row.Email, Nama: nama, Phone: phone,
		Role: helper.RoleRetailMarketing, Aktif: aktif,
	}, newPassword, "")
}

func (s *UserService) SetFee(ctx context.Context, userID int64, feeRp int64) error {
	if userID <= 0 {
		return errors.New("user_id invalid")
	}
	if feeRp < 0 {
		return errors.New("fee_member_rp must be >= 0")
	}
	row, err := s.repo.Get(ctx, userID)
	if err != nil {
		return err
	}
	if row == nil {
		return errors.New("user not found")
	}
	if helper.IsH2HRole(row.Role) {
		return errors.New("akun H2H wajib menggunakan fee kategori, fee flat dinonaktifkan")
	}
	return s.repo.SetFee(ctx, userID, feeRp)
}

func (s *UserService) SetPassword(ctx context.Context, userID int64, newPassword string) error {
	if userID <= 0 || len(strings.TrimSpace(newPassword)) < 8 {
		return errors.New("invalid payload")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(strings.TrimSpace(newPassword)), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.repo.SetPassword(ctx, userID, string(hash))
}

func (s *UserService) SetPIN(ctx context.Context, userID int64, newPIN string) error {
	p := strings.TrimSpace(newPIN)
	if userID <= 0 || len(p) < 4 {
		return errors.New("pin min 4 digit")
	}
	for _, c := range p {
		if c < '0' || c > '9' {
			return errors.New("pin harus angka")
		}
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.repo.SetPIN(ctx, userID, string(hash))
}
