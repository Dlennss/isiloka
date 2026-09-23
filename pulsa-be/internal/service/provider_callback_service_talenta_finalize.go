package service

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"pulsa2/internal/helper"
	"pulsa2/internal/helper/providersn"
	"pulsa2/internal/repository"
)

func resolveTalentaFinalStatus(rcNum int, rcStr, msg string) string {
	finalStatus := "pending"
	upperMsg := strings.ToUpper(strings.TrimSpace(msg))
	parsedBodyStatus := helper.ExtractProviderStatusCode(msg)
	switch {
	case rcNum == 20 || rcStr == "20" || strings.Contains(upperMsg, "SUKSES"):
		finalStatus = "success"
	case helper.IsRetryableFailRC(rcStr) || helper.IsRetryableFailRC(parsedBodyStatus) || helper.IsMissingProviderRC(rcStr) || strings.Contains(upperMsg, "STATUS=61") || strings.Contains(upperMsg, "STATUS = 61"):
		finalStatus = "failed"
	case strings.Contains(upperMsg, "GAGAL") || strings.Contains(upperMsg, "FAILED") || strings.Contains(upperMsg, "ERROR") || strings.Contains(upperMsg, "DIBATALKAN") || strings.Contains(upperMsg, "BATAL"):
		finalStatus = "failed"
	case rcNum == 0 || rcNum == 1 || rcNum == 2 || rcNum == 3 || strings.Contains(upperMsg, "PENDING") || strings.Contains(upperMsg, "PROSES"):
		finalStatus = "pending"
	default:
		finalStatus = "pending"
	}
	return finalStatus
}

func buildTalentaFinalOutput(msg, noreff string, finalStatus string) (ket, providerRefOut, snOut, ketDB string) {
	ket, info := helper.SafeMemberKeterangan(finalStatus, msg)
	snOut = strings.TrimSpace(info.SN)
	if snOut == "" {
		snOut = strings.TrimSpace(noreff)
	}
	providerRefOut = strings.TrimSpace(info.Reff)
	if providerRefOut == "" || snOut == "" {
		pr, sn := providersn.ParseTalentaSNRefFromMsg(msg)
		if providerRefOut == "" {
			providerRefOut = strings.TrimSpace(pr)
		}
		if snOut == "" {
			snOut = strings.TrimSpace(sn)
		}
	}
	ketDB = strings.TrimSpace(snOut)
	if isWeakProviderValue(ketDB) {
		ketDB = ""
	}
	if ketDB == "" && providerRefOut != "" && !isWeakProviderValue(providerRefOut) {
		ketDB = providerRefOut
	}
	if ketDB == "" {
		ketDB = strings.TrimSpace(ket)
	}
	if ketDB == "" {
		ketDB = strings.TrimSpace(msg)
	}
	return
}

func (s *ProviderCallbackService) finalizeTalentaMemberTrx(ctx context.Context, _ *repository.ProviderTrxRefRow, trx *repository.CallbackTrxMemberFull, data talentaCallbackData) (int, map[string]any) {
	finalStatus := resolveTalentaFinalStatus(data.rcNum, data.rcStr, data.msg)
	// Advisory lock: serialize concurrent callbacks for same transaction
	if locked, _ := s.repo.AcquireCallbackLock(ctx, trx.ID); locked {
		defer s.repo.ReleaseCallbackLock(ctx, trx.ID)
		if fresh, _ := s.repo.GetTransaksiMemberByID(ctx, trx.ID); fresh != nil {
			trx = fresh
		}
	}
	if strings.EqualFold(strings.TrimSpace(trx.Status), "success") && finalStatus == "failed" {
		if !helper.ShouldBlockProviderFallback(data.msg) {
			fbCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
			defer cancel()
			did, provider, providerRowID, ferr := s.tryFallbackFromTalenta(fbCtx, trx, data.msg)
			if did {
				helper.AppendProviderServiceLog("provider_callback_service.log", "talentapay provider-row failed after success; member kept success refid=%s trx_id=%d fallback_provider=%s provider_row_id=%d", data.refid, trx.ID, provider, providerRowID)
				return 200, map[string]any{"ok": true, "already_final": true, "refid": data.refid, "status": "success", "fallback_provider": provider, "provider_row_id": providerRowID}
			}
			if ferr != nil {
				helper.AppendProviderServiceLog("provider_callback_error.log", "talentapay finalize fallback failed while member already success refid=%s trx_member_id=%d err=%v", data.refid, trx.ID, ferr)
			}
		}
		helper.AppendAlreadyFinalLog("callback already_final provider=talentapay refid=%s trx_member_id=%d status=success callback_status=failed", data.refid, trx.ID)
		return 200, map[string]any{"ok": true, "already_final": true, "refid": data.refid, "status": "success"}
	}

	if shouldKeepExistingMemberFinalStatus(trx.Status, finalStatus) {
		helper.AppendAlreadyFinalLog("callback already_final provider=talentapay refid=%s trx_member_id=%d status=%s callback_status=%s", data.refid, trx.ID, trx.Status, finalStatus)
		return 200, map[string]any{"ok": true, "already_final": true, "refid": data.refid}
	}

	ket, providerRefOut, snOut, ketDB := buildTalentaFinalOutput(data.msg, data.noreff, finalStatus)
	var biayaAktual int64
	hargaMember := effectiveMemberSellingPrice(trx.HargaMember, trx.BiayaPerkiraan)

	switch finalStatus {
	case "success":
		if err := s.prepareMemberTrxSuccessTransition(ctx, "talentapay", trx); err != nil {
			helper.AppendProviderServiceLog("provider_callback_error.log", "success promotion refund recovery failed provider=talentapay refid=%s trx_member_id=%d err=%v", data.refid, trx.ID, err)
			return 200, map[string]any{"ok": false, "refid": data.refid, "error": err.Error(), "repair_needed": "member_refund_recovery"}
		}
		biayaAktual = trx.BiayaPerkiraan
		if err := s.repo.UpdateTransaksiMemberSettle(ctx, trx.ID, "success", ketDB, biayaAktual, data.price, hargaMember); err != nil && err != sql.ErrNoRows {
			helper.AppendProviderServiceLog("provider_callback_error.log", "settle failed provider=talentapay refid=%s trx_member_id=%d status=success err=%v", data.refid, trx.ID, err)
			return 200, map[string]any{"ok": false, "error": err.Error(), "repair_needed": "member_settle_failed"}
		}
		s.applyH2HCommission(ctx, trx.ID, data.refid)
		// Debit dompet provider di-handle oleh DB trigger.
	case "failed":
		if !helper.ShouldBlockProviderFallback(data.msg) {
			fbCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
			defer cancel()
			did, provider, providerRowID, ferr := s.tryFallbackFromTalenta(fbCtx, trx, data.msg)
			if did {
				return 200, map[string]any{
					"ok":                true,
					"refid":             data.refid,
					"status":            "pending",
					"fallback_provider": provider,
					"provider_row_id":   providerRowID,
					"error": func() any {
						if ferr != nil {
							return ferr.Error()
						}
						return nil
					}(),
				}
			}
			if ferr != nil {
				data.msg = fallbackFinalFailureMessage(data.msg, ferr)
			}
		}
		if err := s.repo.UpdateTransaksiMemberSettle(ctx, trx.ID, "failed", ketDB, 0, data.price, hargaMember); err != nil && err != sql.ErrNoRows {
			helper.AppendProviderServiceLog("provider_callback_error.log", "settle failed provider=talentapay refid=%s trx_member_id=%d status=failed err=%v", data.refid, trx.ID, err)
			return 200, map[string]any{"ok": false, "error": err.Error(), "repair_needed": "member_settle_failed"}
		}
		// Refund di-handle oleh DB trigger saat status di-update ke 'failed'.
	default:
		if err := s.repo.UpdateTransaksiMemberSettle(ctx, trx.ID, "pending", ketDB, 0, data.price, hargaMember); err != nil && err != sql.ErrNoRows {
			helper.AppendProviderServiceLog("provider_callback_error.log", "settle failed provider=talentapay refid=%s trx_member_id=%d status=pending err=%v", data.refid, trx.ID, err)
		}
	}

	if finalStatus == "pending" {
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

	return s.sendDirectMemberWebhook(ctx, "talentapay", trx, webhookURL, data.refid, finalStatus, ket, providerRefOut, snOut, memberSaldo, biayaAktual, data.price)
}
