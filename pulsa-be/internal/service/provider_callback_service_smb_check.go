package service

import (
	"context"
	"fmt"
	"strings"

	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
	"pulsa2/smb"
)

func (s *ProviderCallbackService) processSMBCheckStage(ctx context.Context, row *repository.ProviderTrxRefRow, trx *repository.CallbackTrxMemberFull, data *smbCallbackData) (bool, int, map[string]any) {
	if data.stage != "check" {
		return false, 0, nil
	}
	if smbCheckCallbackAlreadyPromoted(row) {
		return true, 200, map[string]any{"ok": true, "ignored": true, "refid": data.refid, "reason": "duplicate_check_after_pay_started"}
	}
	if smbCheckCallbackAttemptClosed(row) {
		return true, 200, map[string]any{"ok": true, "ignored": true, "refid": data.refid, "reason": "late_check_after_attempt_closed"}
	}
	if trx.Status == "success" || trx.Status == "failed" {
		helper.AppendAlreadyFinalLog("callback already_final provider=smb stage=check refid=%s trx_member_id=%d status=%s — jangan bayar", data.refid, trx.ID, trx.Status)
		return true, 200, map[string]any{"ok": true, "already_final": true, "refid": data.refid}
	}

	isBank := s.isBankH2HProduct(ctx, trx.KodeProduk)

	payable := smb.ParsePayableAmount(data.msg)
	if payable > 0 && data.hasLastBalance && data.lastBalance > 0 && data.lastBalance < payable {
		data.msg = fmt.Sprintf("SMB CHECK saldo provider kurang: saldo=%d jumlah=%d", data.lastBalance, payable)
		data.statusRaw = "0061"
		if err := s.updateSMBProviderResult(ctx, row, *data); err != nil {
			helper.AppendProviderServiceLog("provider_callback_error.log", "smb callback update result failed refid=%s rowID=%d stage=%s err=%v", data.refid, row.ID, data.stage, err)
			return true, 502, map[string]any{"ok": false, "error": err.Error(), "refid": data.refid}
		}
	}

	if smb.LooksLikeImmediateReject(data.msg) {
		ok, provider, _, ferr := s.tryFallbackFromSMB(ctx, row, trx, data.msg)
		if ok && ferr == nil {
			waitReason := "fallback_after_smb_check_failed"
			if strings.EqualFold(strings.TrimSpace(provider), "loketbayar") {
				waitReason = "wait_loketbayar_status"
			}
			ketPending, _ := helper.SafeMemberKeterangan("pending", waitReason)
			_ = s.repo.UpdateTransaksiMemberSettle(ctx, trx.ID, "pending", ketPending, 0, data.price, effectiveMemberSellingPrice(trx.HargaMember, trx.BiayaPerkiraan))
			return true, 200, map[string]any{"ok": true, "refid": data.refid, "status": "pending", "fallback_provider": provider}
		}
		if ferr != nil {
			helper.AppendProviderServiceLog("provider_callback_error.log", "smb check fallback failed refid=%s trx_id=%d err=%v", data.refid, trx.ID, ferr)
		}
		if isBank {
			return false, 0, nil
		}
	} else if !smbCheckCallbackReadyToPay(data.msg) {
		helper.AppendProviderServiceLog("provider_callback_service.log", "SMB CHECK stage ready_to_pay=%t refid=%s msg_len=%d", smbCheckCallbackReadyToPay(data.msg), data.refid, len(data.msg))
		ketPending, _ := helper.SafeMemberKeterangan("pending", "wait_smb_check_callback")
		_ = s.repo.UpdateTransaksiMemberSettle(ctx, trx.ID, "pending", ketPending, 0, data.price, effectiveMemberSellingPrice(trx.HargaMember, trx.BiayaPerkiraan))
		return true, 200, map[string]any{"ok": true, "refid": data.refid, "status": "pending", "reason": "wait_smb_check_callback"}
	} else {
		mode, code, _, parseErr := smb.ParseMappedCodeTargetWithMode(ptrString(row.RequestMode), row.KodeProduk)
		if parseErr != nil || s.smClient == nil {
			ketPending, _ := helper.SafeMemberKeterangan("pending", "wait_smb_pay_callback")
			_ = s.repo.UpdateTransaksiMemberSettle(ctx, trx.ID, "pending", ketPending, 0, data.price, effectiveMemberSellingPrice(trx.HargaMember, trx.BiayaPerkiraan))
			return true, 200, map[string]any{"ok": true, "refid": data.refid, "status": "pending", "reason": "wait_smb_pay_callback"}
		}

		reqQty := trx.QtyProvider
		if reqQty <= 0 {
			reqQty = trx.Qty
		}

		dest := trx.Tujuan
		latestRow, _ := s.getSMBCallbackRow(ctx, data.refid, data.routeCodeHint)
		if latestRow != nil && strings.EqualFold(strings.TrimSpace(latestRow.Status), "failed") {
			helper.AppendProviderServiceLog("provider_callback_service.log", "SMB DISPATCH PAY dibatalkan — provider row sudah failed refid=%s", data.refid)
			return true, 200, map[string]any{"ok": true, "ignored": true, "refid": data.refid, "reason": "provider_already_failed"}
		}
		latestTrx, _ := s.repo.GetTransaksiMemberByID(ctx, trx.ID)
		if latestTrx != nil && (latestTrx.Status == "success" || latestTrx.Status == "failed") {
			helper.AppendProviderServiceLog("provider_callback_service.log", "SMB DISPATCH PAY dibatalkan — member sudah %s refid=%s", latestTrx.Status, data.refid)
			return true, 200, map[string]any{"ok": true, "ignored": true, "refid": data.refid, "reason": "member_already_final"}
		}

		helper.AppendProviderServiceLog("provider_callback_service.log", "SMB DISPATCH PAY refid=%s mode=%s code=%s qty=%d dest=%s is_bank=%t", data.refid, mode, code, reqQty, dest, isBank)
		dispatch, callErr := s.smClient.DispatchPayOnly(ctx, mode, smb.Request{
			KodeProduk: code,
			Tujuan:     dest,
			Qty:        reqQty,
			RefID:      trx.RefID,
		})
		if dispatch == nil || callErr != nil {
			ketPending, _ := helper.SafeMemberKeterangan("pending", "wait_smb_pay_callback")
			_ = s.repo.UpdateTransaksiMemberSettle(ctx, trx.ID, "pending", ketPending, 0, data.price, effectiveMemberSellingPrice(trx.HargaMember, trx.BiayaPerkiraan))
			return true, 200, map[string]any{"ok": true, "refid": data.refid, "status": "pending", "reason": "wait_smb_pay_callback"}
		}

		data.msg = dispatch.Final.Body
		data.statusRaw = smb.ExtractStatusCode(data.msg)
		data.price = smb.ParsePrice(data.msg)
		data.providerRef = smb.ParseProviderRef(data.msg)
		data.payload["stage"] = "pay_after_check_callback"
		data.payload["dispatch"] = dispatch
		if bal, ok := smb.ParseLastBalance(data.msg); ok && bal > 0 {
			data.lastBalance = bal
			data.hasLastBalance = true
		} else if dispatch.LastBalance != nil && *dispatch.LastBalance > 0 {
			data.lastBalance = *dispatch.LastBalance
			data.hasLastBalance = true
		}
		if data.hasLastBalance && data.lastBalance > 0 {
			_ = s.repo.InsertProviderSnapshot(ctx, repository.ProviderSnapshotIn{
				Provider:            "smb",
				SaldoProvider:       data.lastBalance,
				RefID:               data.refid,
				TransaksiMemberID:   &row.TransaksiMemberID,
				TransaksiProviderID: &row.ID,
				Sumber:              "callback_smb_pay_after_check",
			})
		}
		if err := s.updateSMBProviderResult(ctx, row, *data); err != nil {
			helper.AppendProviderServiceLog("provider_callback_error.log", "smb callback update result failed refid=%s rowID=%d stage=%s err=%v", data.refid, row.ID, data.stage, err)
			return true, 502, map[string]any{"ok": false, "error": err.Error(), "refid": data.refid}
		}

		if isBank {
			helper.AppendProviderServiceLog("provider_callback_service.log",
				"SMB PAY bank dispatched — pending tunggu callback BYR refid=%s status_raw=%s", data.refid, data.statusRaw)
			ketPending, _ := helper.SafeMemberKeterangan("pending", "wait_smb_bank_byr_callback")
			_ = s.repo.UpdateTransaksiMemberSettle(ctx, trx.ID, "pending", ketPending, 0, data.price, effectiveMemberSellingPrice(trx.HargaMember, trx.BiayaPerkiraan))
			return true, 200, map[string]any{"ok": true, "refid": data.refid, "status": "pending", "reason": "wait_smb_bank_byr_callback"}
		}

		if dispatch.Final.HTTPStatus != 200 || smb.LooksLikeSystemIssue(data.msg) {
			ketPending, _ := helper.SafeMemberKeterangan("pending", "wait_smb_pay_callback")
			_ = s.repo.UpdateTransaksiMemberSettle(ctx, trx.ID, "pending", ketPending, 0, data.price, effectiveMemberSellingPrice(trx.HargaMember, trx.BiayaPerkiraan))
			return true, 200, map[string]any{"ok": true, "refid": data.refid, "status": "pending", "reason": "wait_smb_pay_callback"}
		}
	}

	return false, 0, nil
}
