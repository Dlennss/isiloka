package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"pulsa2/internal/helper"
	"pulsa2/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

const dashboardTokenTTL = 7 * 24 * time.Hour

func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	password = strings.TrimSpace(password)
	if email == "" || password == "" {
		return "", errors.New("email/password required")
	}

	m, err := s.repo.GetForLogin(ctx, email)
	if err != nil {
		return "", err
	}
	if m == nil || !m.Aktif || strings.TrimSpace(m.PasswordHash) == "" {
		return "", errors.New("invalid credentials")
	}
	if bcrypt.CompareHashAndPassword([]byte(m.PasswordHash), []byte(password)) != nil {
		return "", errors.New("invalid credentials")
	}

	tok, err := helper.MakeJWT(s.jwtSecret, m.ID, m.Role, dashboardTokenTTL)
	if err != nil {
		return "", err
	}
	return tok, nil
}

func (s *AuthService) Refresh(ctx context.Context, token string) (string, string, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return "", "", errors.New("missing token")
	}
	claims, err := helper.ParseJWT(s.jwtSecret, token)
	if err != nil {
		return "", "", errors.New("invalid token")
	}
	me, err := s.repo.GetMe(ctx, claims.Sub)
	if err != nil {
		return "", "", err
	}
	if me == nil || !me.Aktif {
		return "", "", errors.New("member inactive")
	}
	role := helper.NormalizeRole(me.Role)
	if role == "" {
		role = helper.RoleMember
	}
	tok, err := helper.MakeJWT(s.jwtSecret, me.ID, role, dashboardTokenTTL)
	if err != nil {
		return "", "", err
	}
	return tok, role, nil
}

func (s *AuthService) Me(ctx context.Context, memberID int64) (*repository.AuthMeRow, error) {
	if memberID <= 0 {
		return nil, errors.New("unauthorized")
	}
	return s.repo.GetMe(ctx, memberID)
}

func (s *AuthService) Register(ctx context.Context, in repository.UserCreateInput, initialFees *repository.H2HInitialFeeSetupInput) (int64, string, error) {
	in.Email = strings.TrimSpace(strings.ToLower(in.Email))
	in.Nama = strings.TrimSpace(in.Nama)
	in.Phone = normalizeMemberPhone(in.Phone)
	in.Role = helper.NormalizeRole(in.Role)
	if in.Role == "" {
		in.Role = helper.RoleMember
	}
	if in.Email == "" || in.PasswordHash == "" {
		return 0, "", errors.New("email/password required")
	}
	if in.Role != helper.RoleAdmin &&
		in.Role != helper.RoleStaff &&
		in.Role != helper.RoleAuditor &&
		in.Role != helper.RoleMember &&
		in.Role != helper.RoleUser &&
		in.Role != helper.RoleRetailAgent &&
		in.Role != helper.RoleRetailMaster &&
		in.Role != helper.RoleRetailMarketing &&
		in.Role != helper.RoleRetailAnalyst &&
		in.Role != helper.RoleH2HAgent &&
		in.Role != helper.RoleH2HMaster &&
		in.Role != helper.RoleOperatorTrx &&
		in.Role != helper.RoleOperatorWallet {
		return 0, "", errors.New("role tidak valid")
	}
	if helper.IsH2HRole(in.Role) && in.PinHash == "" {
		return 0, "", errors.New("pin required for member")
	}
	if !in.Aktif {
		in.Aktif = true
	}
	in.RetailAgentCommissionRp, in.RetailMasterCommissionRp = helper.ApplyRetailCommissionDefaults(in.Role, in.RetailAgentCommissionRp, in.RetailMasterCommissionRp)
	if helper.IsH2HRole(in.Role) {
		if initialFees == nil {
			return 0, "", errors.New("fee kategori awal member wajib diisi")
		}
		if in.APIKey == "" {
			in.APIKey = helper.RandHex(32)
		}
		if in.Label == "" {
			in.Label = "default"
		}
		return s.repo.CreateMemberWithKeyAndInitialFees(ctx, in, *initialFees)
	}
	return s.repo.CreateInternalUser(ctx, in)
}

func (s *AuthService) RegisterPublic(ctx context.Context, email, nama, phone, passwordHash string, refundIn *RegisterPublicGuestRefundInput) (int64, *repository.AppOrderGuestRefundRow, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	nama = strings.TrimSpace(nama)
	phone = normalizeMemberPhone(phone)
	passwordHash = strings.TrimSpace(passwordHash)
	if email == "" || passwordHash == "" {
		return 0, nil, errors.New("email/password required")
	}
	if phone == "" {
		return 0, nil, errors.New("nomor telepon required")
	}
	if len(phone) < 10 || len(phone) > 14 || !strings.HasPrefix(phone, "08") {
		return 0, nil, errors.New("nomor telepon tidak valid")
	}

	memberID, err := s.repo.CreatePublicUser(ctx, email, nama, phone, passwordHash)
	if err != nil {
		return 0, nil, err
	}

	if refundIn == nil || s.orderRepo == nil {
		return memberID, nil, nil
	}

	invoiceID := strings.TrimSpace(refundIn.InvoiceID)
	guestEmail := normalizeGuestEmail(refundIn.GuestEmail)
	guestPhone := normalizeGuestPhone(refundIn.GuestPhone)
	if invoiceID == "" || guestEmail == "" || guestPhone == "" {
		return memberID, nil, nil
	}

	refundRow, refundErr := s.claimGuestRefundWithRetry(ctx, memberID, invoiceID, guestEmail, guestPhone)
	if refundErr != nil {
		return memberID, nil, nil
	}
	return memberID, refundRow, nil
}

func (s *AuthService) claimGuestRefundWithRetry(ctx context.Context, memberID int64, invoiceID, guestEmail, guestPhone string) (*repository.AppOrderGuestRefundRow, error) {
	if s.orderRepo == nil {
		return nil, errors.New("order repo not configured")
	}

	var lastErr error
	for attempt := 0; attempt < 8; attempt++ {
		row, err := s.orderRepo.ClaimGuestRefundToMember(ctx, memberID, invoiceID, guestEmail, guestPhone)
		if err == nil {
			return row, nil
		}
		lastErr = err

		lowerErr := strings.ToLower(strings.TrimSpace(err.Error()))
		retryable := strings.Contains(lowerErr, "tidak ditemukan") || strings.Contains(lowerErr, "status refund tidak valid")
		if !retryable || attempt == 7 {
			break
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(400 * time.Millisecond):
		}
	}
	return nil, lastErr
}

func normalizeMemberPhone(value string) string {
	phone := strings.TrimSpace(value)
	phone = strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		if r == '+' {
			return r
		}
		return -1
	}, phone)
	if strings.HasPrefix(phone, "+62") {
		return "0" + strings.TrimPrefix(phone, "+62")
	}
	if strings.HasPrefix(phone, "62") {
		return "0" + strings.TrimPrefix(phone, "62")
	}
	return phone
}
