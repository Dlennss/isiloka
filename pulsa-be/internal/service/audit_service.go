package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"pulsa2/internal/helper"
	"pulsa2/internal/provider"
	"pulsa2/internal/repository"
	"pulsa2/smb"
)

type AuditService struct {
	repo         *repository.AuditRepository
	providerRepo *repository.ProviderCallbackRepository
	smClient     *smb.Client
}

func NewAuditService(repo *repository.AuditRepository, smClient *smb.Client) *AuditService {
	return &AuditService{repo: repo, providerRepo: repository.NewProviderCallbackRepository(repo.DB()), smClient: smClient}
}

func (s *AuditService) AdminListStatusMismatch(
	ctx context.Context,
	limit, offset int,
	provider, statusMember, mismatchType, fromStr, toStr string,
) ([]repository.AdminStatusMismatchRow, int64, error) {
	return s.repo.AdminListStatusMismatch(ctx, limit, offset, provider, statusMember, mismatchType, fromStr, toStr)
}

func (s *AuditService) AdminListGuestRefundMissing(
	ctx context.Context,
	limit, offset int,
	fromStr, toStr, invoiceID string,
) ([]repository.AdminGuestRefundMissingRow, int64, error) {
	return s.repo.AdminListGuestRefundMissing(ctx, limit, offset, fromStr, toStr, invoiceID)
}

func (s *AuditService) AdminListProviderEmptyResponse(
	ctx context.Context,
	limit, offset int,
	provider, refID, kodeProduk, tujuan, fromStr, toStr string,
) ([]repository.AdminProviderEmptyResponseRow, int64, error) {
	return s.repo.AdminListProviderEmptyResponse(ctx, limit, offset, provider, refID, kodeProduk, tujuan, fromStr, toStr)
}

func (s *AuditService) AdminListProviderWalletMissingDebit(
	ctx context.Context,
	limit, offset int,
	provider, refID, fromStr, toStr string,
) ([]repository.AdminProviderWalletMissingDebitRow, int64, error) {
	return s.repo.AdminListProviderWalletMissingDebit(ctx, limit, offset, provider, refID, fromStr, toStr)
}

func (s *AuditService) AdminListProviderSuccessSuspiciousMessage(
	ctx context.Context,
	limit, offset int,
	provider, refID, resolveStatus, fromStr, toStr string,
	includeTotal bool,
) ([]repository.AdminProviderSuccessSuspiciousMessageRow, int64, bool, error) {
	return s.repo.AdminListProviderSuccessSuspiciousMessage(ctx, limit, offset, provider, refID, resolveStatus, fromStr, toStr, includeTotal)
}

func (s *AuditService) ResolveProviderSuccessSuspiciousMessage(
	ctx context.Context,
	userID int64,
	transaksiProviderID int64,
	note string,
) (*repository.MarkingTrxResolveResult, error) {
	return s.repo.ResolveProviderSuccessSuspiciousMessage(ctx, userID, transaksiProviderID, note)
}

func (s *AuditService) SettleProviderSuccessSuspiciousMessageWithBankDebit(
	ctx context.Context,
	userID int64,
	transaksiProviderID int64,
	nominal int64,
	fee int64,
	note string,
) (*repository.ProviderSuspectBankSettlementResult, error) {
	return s.repo.SettleProviderSuccessSuspiciousMessageWithBankDebit(ctx, userID, transaksiProviderID, nominal, fee, note)
}

func (s *AuditService) ResolveProviderSuccessSuspiciousMessageBulk(
	ctx context.Context,
	userID int64,
	transaksiProviderIDs []int64,
	note string,
) (*repository.BulkMarkingTrxResolveResult, error) {
	out := &repository.BulkMarkingTrxResolveResult{
		ProcessedIDs: make([]int64, 0, len(transaksiProviderIDs)),
		Failed:       make([]map[string]any, 0),
	}
	seen := make(map[int64]bool, len(transaksiProviderIDs))
	for _, id := range transaksiProviderIDs {
		if id <= 0 || seen[id] {
			continue
		}
		seen[id] = true
		out.Processed++
		if _, err := s.repo.ResolveProviderSuccessSuspiciousMessage(ctx, userID, id, note); err != nil {
			out.FailedCount++
			out.Failed = append(out.Failed, map[string]any{"transaksi_provider_id": id, "error": err.Error()})
			continue
		}
		out.ResolvedCount++
		out.ProcessedIDs = append(out.ProcessedIDs, id)
	}
	if len(out.Failed) == 0 {
		out.Failed = nil
	}
	return out, nil
}

func (s *AuditService) ResolveProviderWalletMissingDebit(
	ctx context.Context,
	actorID int64,
	transaksiProviderID int64,
) (*repository.ResolveProviderWalletMissingDebitResult, error) {
	return s.repo.ResolveProviderWalletMissingDebit(ctx, actorID, transaksiProviderID)
}

func (s *AuditService) IgnoreProviderWalletMissingDebit(
	ctx context.Context,
	actorID int64,
	transaksiProviderID int64,
	note string,
) error {
	return s.repo.IgnoreProviderWalletMissingDebit(ctx, actorID, transaksiProviderID, note)
}

func (s *AuditService) ResolveProviderWalletMissingDebitBulk(
	ctx context.Context,
	actorID int64,
	transaksiProviderIDs []int64,
) (*repository.BulkResolveProviderWalletMissingDebitResult, error) {
	return s.repo.ResolveProviderWalletMissingDebitBulk(ctx, actorID, transaksiProviderIDs)
}

func (s *AuditService) IgnoreProviderWalletMissingDebitBulk(
	ctx context.Context,
	actorID int64,
	transaksiProviderIDs []int64,
	note string,
) (*repository.BulkIgnoreProviderWalletMissingDebitResult, error) {
	return s.repo.IgnoreProviderWalletMissingDebitBulk(ctx, actorID, transaksiProviderIDs, note)
}

func (s *AuditService) ResendProviderSuccessSuspiciousMessageToBIFASTOPEN2(
	ctx context.Context,
	userID int64,
	transaksiProviderID int64,
) (map[string]any, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("user_id invalid")
	}
	if s == nil || s.repo == nil || s.providerRepo == nil || s.smClient == nil {
		return nil, fmt.Errorf("service resend SMB tidak siap")
	}
	target, err := s.repo.GetProviderSuccessSuspiciousResendTarget(ctx, transaksiProviderID)
	if err != nil {
		return nil, err
	}
	backupMapID, backupCode, err := s.repo.FindBackupSMBMapForInternalSKU(ctx, target.ProdukMember, target.Qty)
	if err != nil {
		return nil, err
	}
	createIn := repository.ProviderTrxCreateIn{
		Provider:            "smb",
		TransaksiMemberID:   target.TransaksiMemberID,
		RefID:               target.RefID,
		Perintah:            "PAY",
		ProdukSKUSnapshot:   target.ProdukMember,
		ProdukProviderMapID: backupMapID,
		KodeProduk:          backupCode,
		Tujuan:              target.Tujuan,
		Qty:                 target.Qty,
	}
	if existing, err := s.providerRepo.FindProviderTrxByRoute(ctx, createIn); err != nil {
		return nil, err
	} else if existing != nil {
		return map[string]any{
			"already_exists":  true,
			"provider_row_id": existing.ID,
			"refid":           target.RefID,
			"product_sent":    backupCode,
		}, nil
	}
	row, err := s.providerRepo.CreateProviderTrx(ctx, createIn, map[string]any{
		"source":                  "admin_transaksi_suspect_resend_bifastopen2",
		"trigger_provider_trx_id": target.TransaksiProviderID,
		"trigger_message":         target.Pesan,
		"actor_user_id":           userID,
		"product_in":              target.ProdukMember,
		"product_sent":            backupCode,
		"mode":                    "DIRECT",
	})
	if err != nil {
		return nil, err
	}
	adapter := &provider.SMBAdapter{C: s.smClient}
	resp, callErr := adapter.Pay(ctx, provider.PayRequest{
		Product: backupCode,
		Mode:    "DIRECT",
		Dest:    target.Tujuan,
		Qty:     target.Qty,
		RefID:   target.RefID,
	})
	if resp == nil {
		hs := 0
		msg := "admin resend BIFASTOPEN2: respons SMB kosong"
		_ = s.providerRepo.UpdateResult(ctx, row.ID, repository.UpdateResult{HTTPStatus: &hs, Pesan: &msg, ResponMentah: map[string]any{"error": msg}})
		return nil, errors.New(msg)
	}
	upd := repository.UpdateResult{
		HTTPStatus:   &resp.HTTPStatus,
		Pesan:        helper.PtrString(resp.Message),
		Harga:        helper.PtrI64(resp.Price),
		NoReferensi:  helper.PtrString(resp.ProviderRef),
		ResponMentah: resp.Raw,
	}
	if resp.RC != "" {
		upd.KodeRespon = &resp.RC
	}
	if resp.Balance > 0 {
		upd.SaldoTerakhir = &resp.Balance
		_ = s.providerRepo.InsertProviderSnapshot(ctx, repository.ProviderSnapshotIn{
			Provider:            "smb",
			SaldoProvider:       resp.Balance,
			RefID:               target.RefID,
			TransaksiMemberID:   &target.TransaksiMemberID,
			TransaksiProviderID: &row.ID,
			Sumber:              "admin_transaksi_suspect_resend_bifastopen2",
		})
	}
	if (resp.HTTPStatus != 200 || callErr != nil) && strings.TrimSpace(resp.Message) == "" {
		msg := fmt.Sprintf("admin resend BIFASTOPEN2 gagal http=%d", resp.HTTPStatus)
		upd.Pesan = &msg
	}
	_ = s.providerRepo.UpdateResult(ctx, row.ID, upd)

	finalState := ""
	if resp.HTTPStatus == 200 {
		finalState = helper.ProviderResponseStatusString("smb", upd.KodeRespon, upd.Pesan)
	}
	memberFinalized := false
	if finalState == "failed" {
		if trx, err := s.providerRepo.GetTransaksiMemberByID(ctx, target.TransaksiMemberID); err == nil && trx != nil {
			if strings.EqualFold(strings.TrimSpace(trx.Status), "pending") {
				ketDB := strings.TrimSpace(resp.ProviderRef)
				if ketDB == "" {
					ketDB = strings.TrimSpace(resp.Message)
				}
				hargaMember := effectiveMemberSellingPrice(trx.HargaMember, trx.BiayaPerkiraan)
				if err := s.providerRepo.UpdateTransaksiMemberSettle(ctx, trx.ID, "failed", ketDB, 0, resp.Price, hargaMember); err == nil {
					memberFinalized = true
				}
			}
		}
	}
	return map[string]any{
		"provider_row_id":      row.ID,
		"refid":                target.RefID,
		"product_sent":         backupCode,
		"http_status":          resp.HTTPStatus,
		"accepted":             resp.HTTPStatus == 200 && helper.ProviderResponseAccepted("smb", resp.Body),
		"immediate_reject":     helper.ProviderResponseImmediateReject("smb", resp.Body),
		"message":              strings.TrimSpace(resp.Body),
		"final_provider_state": finalState,
		"member_finalized":     memberFinalized,
	}, callErr
}
