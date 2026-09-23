package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"pulsa2/internal/helper"
	"pulsa2/internal/provider"
	"pulsa2/internal/repository"
)

const pulsa24JamDepositNoteMarker = "provider=pulsa24jam"

func (s *DepositService) CreateQrisRequest(ctx context.Context, memberID int64, role string, amount int64) (*DepositQrisCreateResult, error) {
	if memberID <= 0 {
		return nil, errors.New("unauthorized")
	}
	if !helper.IsRetailRole(role) {
		return nil, errors.New("topup qris hanya untuk retail")
	}
	if amount <= 0 {
		return nil, errors.New("amount harus > 0")
	}
	if s == nil || s.p24 == nil || !s.p24.Configured() {
		return nil, errors.New("topup QRIS Pulsa24Jam belum dikonfigurasi")
	}

	refID := depositQrisPrefix + time.Now().Format("20060102150405") + "-" + strings.ToUpper(helper.RandHex(4))
	upstream, err := s.p24.CreateDepositQRIS(ctx, refID, amount)
	if err != nil {
		return nil, err
	}
	if upstream.Amount > 0 && upstream.Amount != amount {
		return nil, errors.New("nominal QRIS Pulsa24Jam tidak sesuai")
	}
	note := pulsa24JamDepositNote(upstream)
	if err := s.repo.CreateQrisRequest(ctx, memberID, amount, depositQrisMethod, refID, note); err != nil {
		return nil, err
	}
	if err := s.repo.UpdateQrisPending(ctx, refID, upstream.QRURL, note); err != nil {
		return nil, err
	}
	return pulsa24JamDepositResult(refID, amount, upstream), nil
}

func (s *DepositService) GetQrisStatusByRefID(ctx context.Context, memberID int64, role, refID string, refresh bool) (*DepositQrisCreateResult, *repository.DepositRequestRow, error) {
	if memberID <= 0 {
		return nil, nil, errors.New("unauthorized")
	}
	if !helper.IsRetailRole(role) {
		return nil, nil, errors.New("topup qris hanya untuk retail")
	}
	refID = strings.TrimSpace(refID)
	if refID == "" {
		return nil, nil, errors.New("ref_id wajib diisi")
	}
	row, err := s.repo.GetByRefID(ctx, refID)
	if err != nil {
		return nil, nil, err
	}
	if row.MemberID != memberID {
		return nil, nil, errors.New("deposit qris tidak ditemukan")
	}
	if !strings.Contains(strings.ToLower(row.Note), pulsa24JamDepositNoteMarker) {
		return nil, nil, errors.New("deposit QRIS bukan milik Pulsa24Jam")
	}

	var upstream *provider.Pulsa24JamDepositQRISResponse
	if refresh && strings.EqualFold(strings.TrimSpace(row.Status), "pending") {
		upstream, err = s.syncPulsa24JamQrisStatus(ctx, row)
		if err != nil {
			return nil, nil, err
		}
		row, err = s.repo.GetByRefID(ctx, refID)
		if err != nil {
			return nil, nil, err
		}
	}
	result := buildDepositQrisResultFromRow(row)
	if upstream != nil {
		result = pulsa24JamDepositResult(refID, row.Amount, upstream)
		result.Status = row.Status
	}
	return result, row, nil
}

func (s *DepositService) syncPulsa24JamQrisStatus(ctx context.Context, row *repository.DepositRequestRow) (*provider.Pulsa24JamDepositQRISResponse, error) {
	if s == nil || s.p24 == nil || !s.p24.Configured() {
		return nil, errors.New("topup QRIS Pulsa24Jam belum dikonfigurasi")
	}
	if row == nil || strings.TrimSpace(row.RefID) == "" {
		return nil, errors.New("deposit QRIS tidak valid")
	}
	upstream, err := s.p24.DepositQRISStatus(ctx, row.RefID)
	if err != nil {
		return nil, err
	}
	note := pulsa24JamDepositNote(upstream)
	if err := s.repo.UpdateQrisPending(ctx, row.RefID, upstream.QRURL, note); err != nil {
		return nil, err
	}
	switch normalizePulsa24JamDepositStatus(upstream.Status) {
	case "approved":
		if err := s.repo.ApproveQrisByRefID(ctx, row.RefID, note); err != nil {
			return nil, err
		}
	case "rejected":
		if err := s.repo.RejectQrisByRefID(ctx, row.RefID, note); err != nil {
			return nil, err
		}
	}
	return upstream, nil
}

func (s *DepositService) ReconcilePulsa24JamQris(ctx context.Context, limit int) (int, error) {
	if s == nil || s.p24 == nil || !s.p24.Configured() {
		return 0, nil
	}
	rows, err := s.repo.AdminList(ctx, "pending", 0, "", "", limit, 0, "asc")
	if err != nil {
		return 0, err
	}
	processed := 0
	for i := range rows {
		row := &rows[i]
		if !strings.EqualFold(strings.TrimSpace(row.Metode), depositQrisMethod) || !strings.Contains(strings.ToLower(row.Note), pulsa24JamDepositNoteMarker) {
			continue
		}
		if _, err := s.syncPulsa24JamQrisStatus(ctx, row); err != nil {
			continue
		}
		processed++
	}
	return processed, nil
}

func pulsa24JamDepositResult(refID string, amount int64, upstream *provider.Pulsa24JamDepositQRISResponse) *DepositQrisCreateResult {
	return &DepositQrisCreateResult{
		RefID:         refID,
		Amount:        amount,
		FeeAdmin:      upstream.FeeAdmin,
		GrossAmount:   upstream.GrossAmount,
		Status:        normalizePulsa24JamDepositStatus(upstream.Status),
		PaymentType:   upstream.PaymentType,
		TransactionID: upstream.TransactionID,
		QRURL:         upstream.QRURL,
		ExpiredAt:     upstream.ExpiredAt,
		Actions:       upstream.Actions,
	}
}

func pulsa24JamDepositNote(upstream *provider.Pulsa24JamDepositQRISResponse) string {
	return fmt.Sprintf("%s provider_refid=%s status=%s", pulsa24JamDepositNoteMarker, strings.TrimSpace(upstream.ProviderRefID), normalizePulsa24JamDepositStatus(upstream.Status))
}

func normalizePulsa24JamDepositStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "approved", "success", "successful", "settlement", "settled", "paid":
		return "approved"
	case "rejected", "failed", "failure", "expired", "cancelled", "canceled", "deny":
		return "rejected"
	default:
		return "pending"
	}
}
