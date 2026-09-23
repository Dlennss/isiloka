package service

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
)

type MemberSelfService struct {
	repo       *repository.MemberSelfRepository
	mu         sync.Mutex
	statsCache map[int64]memberStatsCacheEntry
}

type memberStatsCacheEntry struct {
	expiresAt time.Time
	value     map[string]any
}

const memberStatsCacheTTL = 5 * time.Minute

func NewMemberSelfService(repo *repository.MemberSelfRepository) *MemberSelfService {
	return &MemberSelfService{repo: repo, statsCache: make(map[int64]memberStatsCacheEntry)}
}

func (s *MemberSelfService) Profile(ctx context.Context, memberID int64) (map[string]any, error) {
	p, err := s.repo.GetMemberProfile(ctx, memberID)
	if err != nil {
		return nil, err
	}
	keys, _ := s.repo.ListAPIKeys(ctx, memberID)
	ips, _ := s.repo.ListIPWhitelist(ctx, memberID)
	return map[string]any{"ok": true, "profile": p, "api_keys": keys, "ip_whitelist": ips}, nil
}

func (s *MemberSelfService) UpdateProfile(ctx context.Context, memberID int64, nama, phone, profilePhotoURL string) (map[string]any, error) {
	nama = strings.TrimSpace(nama)
	phone = normalizeProfilePhone(phone)
	profilePhotoURL = strings.TrimSpace(profilePhotoURL)
	if memberID <= 0 {
		return nil, errors.New("invalid member_id")
	}
	if nama == "" {
		return nil, errors.New("nama wajib diisi")
	}
	if phone != "" && (len(phone) < 10 || len(phone) > 14 || !strings.HasPrefix(phone, "08")) {
		return nil, errors.New("nomor handphone tidak valid")
	}
	if profilePhotoURL != "" {
		if !strings.HasPrefix(profilePhotoURL, "data:image/") {
			return nil, errors.New("format foto profil tidak valid")
		}
		if len(profilePhotoURL) > 800000 {
			return nil, errors.New("ukuran foto profil maksimal 600KB")
		}
	}
	p, err := s.repo.UpdateMemberProfile(ctx, memberID, nama, phone, profilePhotoURL)
	if err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "profile": p}, nil
}

func (s *MemberSelfService) Stats(ctx context.Context, memberID int64) (map[string]any, error) {
	now := time.Now()
	s.mu.Lock()
	if cached, ok := s.statsCache[memberID]; ok && now.Before(cached.expiresAt) {
		s.mu.Unlock()
		return cached.value, nil
	}
	s.mu.Unlock()

	rows, err := s.repo.GetMemberStats3Months(ctx, memberID, now)
	if err != nil {
		return nil, err
	}
	overall, err := s.repo.GetMemberOverallStats(ctx, memberID)
	if err != nil {
		return nil, err
	}
	out := map[string]any{"ok": true, "rows": rows, "overall": overall}
	s.mu.Lock()
	if s.statsCache == nil {
		s.statsCache = make(map[int64]memberStatsCacheEntry)
	}
	s.statsCache[memberID] = memberStatsCacheEntry{expiresAt: now.Add(memberStatsCacheTTL), value: out}
	s.mu.Unlock()
	return out, nil
}

func (s *MemberSelfService) ChangePassword(ctx context.Context, memberID int64, oldPassword, newPassword string) error {
	oldPassword = strings.TrimSpace(oldPassword)
	newPassword = strings.TrimSpace(newPassword)
	if oldPassword == "" || newPassword == "" {
		return errors.New("old_password/new_password required")
	}
	if len(newPassword) < 6 {
		return errors.New("new_password minimal 6 karakter")
	}
	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.repo.ChangePassword(ctx, memberID, oldPassword, string(newHash))
}

func (s *MemberSelfService) ChangePIN(ctx context.Context, memberID int64, oldPIN, newPIN string) error {
	oldPIN = strings.TrimSpace(oldPIN)
	newPIN = strings.TrimSpace(newPIN)
	if oldPIN == "" || newPIN == "" {
		return errors.New("old_pin/new_pin required")
	}
	if len(newPIN) < 4 {
		return errors.New("new_pin minimal 4 digit")
	}
	newHash, err := bcrypt.GenerateFromPassword([]byte(newPIN), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.repo.ChangePIN(ctx, memberID, oldPIN, string(newHash))
}

func (s *MemberSelfService) ResetAPIKey(ctx context.Context, memberID int64) (string, error) {
	apiKey := helper.RandHex(32)
	if apiKey == "" {
		return "", errors.New("failed to generate api key")
	}
	if err := s.repo.ResetAPIKey(ctx, memberID, apiKey, "default"); err != nil {
		return "", err
	}
	return apiKey, nil
}

func (s *MemberSelfService) ListIPWhitelist(ctx context.Context, memberID int64) ([]repository.MemberIPWhitelist, error) {
	return s.repo.ListIPWhitelist(ctx, memberID)
}

func (s *MemberSelfService) AddIPWhitelist(ctx context.Context, memberID int64, ip, label, webhookURL string) error {
	return s.repo.AddIPWhitelist(ctx, memberID, ip, label, webhookURL)
}

func (s *MemberSelfService) DeleteIPWhitelist(ctx context.Context, memberID int64, idStr string) error {
	return s.repo.DeleteIPWhitelist(ctx, memberID, idStr)
}

func (s *MemberSelfService) UpdateChargeReceiver(ctx context.Context, memberID int64, role string, chargeReceiver bool) error {
	if memberID <= 0 {
		return errors.New("invalid member_id")
	}
	if strings.TrimSpace(strings.ToLower(role)) != "member" {
		return errors.New("khusus role member")
	}

	profile, err := s.repo.GetMemberProfile(ctx, memberID)
	if err != nil {
		return err
	}
	if profile.ID <= 0 {
		return errors.New("member not found")
	}
	if profile.ChargeReceiver == chargeReceiver {
		return nil
	}

	return s.repo.UpdateChargeReceiver(ctx, memberID, chargeReceiver)
}

func normalizeProfilePhone(value string) string {
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
