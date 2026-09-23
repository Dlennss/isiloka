package service

import (
	"context"
	"database/sql"
	"strings"

	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
	"pulsa2/smb"
)

func (s *ProviderCallbackService) finalizeSMBCallback(ctx context.Context, row *repository.ProviderTrxRefRow, trx *repository.CallbackTrxMemberFull, data smbCallbackData) (int, map[string]any) {
	finalStatus := "pending"
	switch {
	case smb.LooksLikeImmediateReject(data.rawMsg):
		finalStatus = "failed"
	case smb.LooksLikeSuccess(data.rawMsg):
		finalStatus = "success"
	case smb.LooksLikePending(data.rawMsg):
		finalStatus = "pending"
	}

	ket, info := helper.SafeMemberKeterangan(finalStatus, data.msg)
	ketDB := strings.TrimSpace(data.providerRef)
	if ketDB == "" {
		ketDB = ket
	}
	hargaMember := effectiveMemberSellingPrice(trx.HargaMember, trx.BiayaPerkiraan)
	routeMode, routeCode, _, _ := smb.ParseMappedCodeTargetWithMode(ptrString(row.RequestMode), row.KodeProduk)
	isBank := s.isBankH2HProduct(ctx, trx.KodeProduk)
	fallbackEligible := smbFailureAllowsDowngradeOrFallback(isBank, routeMode, routeCode, data.msg)
	if locked, _ := s.repo.AcquireCallbackLock(ctx, trx.ID); locked {
		defer s.repo.ReleaseCallbackLock(ctx, trx.ID)
		if fresh, _ := s.repo.GetTransaksiMemberByID(ctx, trx.ID); fresh != nil {
			trx = fresh
		}
	}
	if strings.EqualFold(strings.TrimSpace(trx.Status), "success") && finalStatus == "failed" {
		if fallbackEligible {
			if lbRow, err := s.repo.GetLatestByRefIDProvider(ctx, data.refid, "loketbayar"); err == nil && lbRow != nil && strings.EqualFold(strings.TrimSpace(lbRow.Status), "success") {
				helper.AppendAlreadyFinalLog("callback already_final provider=smb refid=%s trx_member_id=%d status=success callback_status=failed route_code=%s reason=loketbayar_already_success", data.refid, trx.ID, routeCode)
				return 200, map[string]any{"ok": true, "already_final": true, "refid": data.refid, "status": "success", "reason": "loketbayar_already_success"}
			}
			ok, providerName, _, ferr := s.tryFallbackFromSMB(ctx, row, trx, data.msg)
			if ok && ferr == nil {
				helper.AppendProviderServiceLog("provider_callback_service.log",
					"SMB provider-row failed after success; member kept success refid=%s trx_id=%d route_code=%s is_bank=%t fallback_provider=%s", data.refid, trx.ID, routeCode, isBank, providerName)
				return 200, map[string]any{"ok": true, "already_final": true, "refid": data.refid, "status": "success", "fallback_provider": providerName}
			}
			if ferr != nil {
				helper.AppendProviderServiceLog("provider_callback_error.log", "smb finalize fallback failed while member already success refid=%s trx_member_id=%d err=%v", data.refid, trx.ID, ferr)
			}
		}
		helper.AppendAlreadyFinalLog("callback already_final provider=smb refid=%s trx_member_id=%d status=success callback_status=failed route_code=%s", data.refid, trx.ID, routeCode)
		return 200, map[string]any{"ok": true, "already_final": true, "refid": data.refid, "status": "success"}
	}

	if shouldKeepExistingMemberFinalStatus(trx.Status, finalStatus) {
		walletSynced := false
		if finalStatus == "success" {
			walletSynced = s.syncSMBWalletSuccess(ctx, data.refid, row, data.price, "auto debit by callback (success, already final)")
		}
		helper.AppendAlreadyFinalLog("callback already_final provider=smb refid=%s trx_member_id=%d status=%s callback_status=%s wallet_synced=%t", data.refid, trx.ID, trx.Status, finalStatus, walletSynced)
		return 200, map[string]any{"ok": true, "already_final": true, "refid": data.refid, "wallet_synced": walletSynced}
	}

	if finalStatus == "failed" && fallbackEligible {
		ok, providerName, _, ferr := s.tryFallbackFromSMB(ctx, row, trx, data.msg)
		if ok && ferr == nil {
			waitReason := "wait_provider_status"
			if strings.EqualFold(strings.TrimSpace(providerName), "loketbayar") {
				waitReason = "wait_loketbayar_status"
			}
			ketPending, _ := helper.SafeMemberKeterangan("pending", waitReason)
			pendErr := s.repo.UpdateTransaksiMemberSettle(ctx, trx.ID, "pending", ketPending, 0, data.price, hargaMember)
			if pendErr != nil && pendErr != sql.ErrNoRows {
				helper.AppendProviderServiceLog("provider_callback_error.log", "smb reopen pending failed refid=%s trx_member_id=%d provider=%s err=%v", data.refid, trx.ID, providerName, pendErr)
				return 502, map[string]any{"ok": false, "refid": data.refid, "error": pendErr.Error(), "repair_needed": "member_reopen_pending"}
			}
			return 200, map[string]any{"ok": true, "refid": data.refid, "status": "pending", "fallback_provider": providerName}
		}
		if ferr != nil {
			helper.AppendProviderServiceLog("provider_callback_error.log", "smb finalize fallback failed refid=%s trx_member_id=%d err=%v", data.refid, trx.ID, ferr)
		}
	}

	switch finalStatus {
	case "success":
		if err := s.prepareMemberTrxSuccessTransition(ctx, "smb", trx); err != nil {
			helper.AppendProviderServiceLog("provider_callback_error.log", "success promotion refund recovery failed provider=smb refid=%s trx_member_id=%d err=%v", data.refid, trx.ID, err)
			return 200, map[string]any{"ok": false, "refid": data.refid, "error": err.Error(), "repair_needed": "member_refund_recovery"}
		}
		biayaAktual := trx.BiayaPerkiraan
		if err := s.repo.UpdateTransaksiMemberSettle(ctx, trx.ID, "success", ketDB, biayaAktual, data.price, hargaMember); err != nil && err != sql.ErrNoRows {
			helper.AppendProviderServiceLog("provider_callback_error.log", "settle failed provider=smb refid=%s trx_member_id=%d status=success err=%v", data.refid, trx.ID, err)
			return 200, map[string]any{"ok": false, "error": err.Error(), "repair_needed": "member_settle_failed"}
		}
		s.applyH2HCommission(ctx, trx.ID, data.refid)
		_ = s.syncSMBWalletSuccess(ctx, data.refid, row, data.price, "auto debit by callback (success)")
	case "failed":
		if err := s.repo.UpdateTransaksiMemberSettle(ctx, trx.ID, "failed", ketDB, 0, data.price, hargaMember); err != nil && err != sql.ErrNoRows {
			helper.AppendProviderServiceLog("provider_callback_error.log", "settle failed provider=smb refid=%s trx_member_id=%d status=failed err=%v", data.refid, trx.ID, err)
			return 200, map[string]any{"ok": false, "error": err.Error(), "repair_needed": "member_settle_failed"}
		}
	default:
		_ = s.repo.UpdateTransaksiMemberSettle(ctx, trx.ID, "pending", ketDB, 0, data.price, hargaMember)
		return 200, map[string]any{"ok": true, "refid": data.refid, "status": "pending"}
	}

	webhookURL, err := s.repo.GetMemberWebhookURL(ctx, trx.MemberID)
	if err != nil || strings.TrimSpace(webhookURL) == "" {
		return 200, map[string]any{"ok": true, "refid": data.refid}
	}
	memberSaldo := int64(0)
	if bal, err := s.repo.GetSaldo(ctx, trx.MemberID); err == nil {
		memberSaldo = bal
	}
	biayaAktualOut := int64(0)
	if finalStatus == "success" {
		biayaAktualOut = trx.BiayaPerkiraan
	}
	providerRefOut := strings.TrimSpace(info.Reff)
	if providerRefOut == "" {
		providerRefOut = data.providerRef
	}
	snOut := strings.TrimSpace(info.SN)
	if snOut == "" {
		snOut = data.providerRef
	}
	return s.sendDirectMemberWebhook(ctx, "smb", trx, webhookURL, data.refid, finalStatus, ket, providerRefOut, snOut, memberSaldo, biayaAktualOut, data.price)
}
