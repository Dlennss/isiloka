package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"
	"strings"

	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
)

func (s *ProviderCallbackService) ProcessJavapayCallback(ctx context.Context, rawQuery string, payload map[string]any) (int, map[string]any) {
	data := helper.AsMap(payload["data"])
	trxObj := helper.AsMap(payload["trx"])

	msg := helper.PickStr(payload, data, "message")
	if msg == "" {
		msg = helper.PickStr(payload, nil, "pesan")
	}
	if msg == "" && trxObj != nil {
		msg = helper.PickStr(payload, trxObj, "status_desc")
	}

	refID := helper.PickStr(payload, data, "refid", "refId", "refID", "ref_id")
	if refID == "" {
		refID = helper.PickStr(payload, nil, "refid", "refId", "refID", "ref_id")
	}
	if refID == "" {
		refID = helper.ExtractRefIDFromPesan(msg)
	}

	var (
		trxid       string
		rcNum       int
		rcStr       string
		price       int64
		qty         int64
		dest        string
		sn          string
		noreff      string
		lastBalance int64
	)

	if trxObj != nil {
		trxid = helper.PickStr(payload, nil, "trxid")
		rcNum = int(helper.PickI64(payload, trxObj, "status"))
		rcStr = strconv.Itoa(rcNum)
		price = helper.PickI64(payload, trxObj, "price")
		qty = helper.PickI64(payload, trxObj, "qty", "amount")
		dest = helper.PickStr(payload, trxObj, "dest", "tujuan")
		sn = helper.PickStr(payload, trxObj, "sn")
		noreff = sn
	} else {
		trxid = helper.PickStr(payload, data, "trxid")
		rcNum = int(helper.PickI64(payload, data, "rc"))
		rcStr = helper.PickStr(payload, data, "rc")
		if rcStr == "" {
			rcStr = strconv.Itoa(rcNum)
		}
		price = helper.PickI64(payload, data, "price")
		qty = helper.PickI64(payload, data, "qty", "amount")
		dest = helper.PickStr(payload, data, "dest", "tujuan")
		noreff = helper.PickStr(payload, data, "noreff")
		lastBalance = helper.PickI64(payload, data, "last_balance")
		if strings.TrimSpace(noreff) != "" {
			sn = strings.TrimSpace(noreff)
		}
	}

	if refID == "" {
		var qtyPtr *int64
		if qty > 0 {
			qtyPtr = &qty
		}
		dest = strings.TrimSpace(dest)
		var destPtr *string
		if dest != "" {
			destPtr = &dest
		}
		if err := s.repo.InsertAnomali(ctx, repository.ProviderAnomasiIn{
			Provider:   "javapay",
			RefID:      "",
			RawQuery:   rawQuery,
			RawBody:    "",
			Payload:    payload,
			KodeRespon: &rcStr,
			Pesan:      &msg,
			Harga:      &price,
			Tujuan:     destPtr,
			Qty:        qtyPtr,
		}); err != nil {
			helper.AppendProviderServiceLog("provider_anomali.log", "javapay callback insert anomali failed missing refid err=%v", err)
		}
		saldoAnomali := lastBalance
		if saldoAnomali <= 0 {
			if bal, ok := helper.ExtractSaldoTerakhirFromMsg(msg); ok {
				saldoAnomali = bal
			}
		}
		s.insertAnomaliSnapshot(ctx, "javapay", refID, "callback_javapay_anomali", saldoAnomali, payload)
		return 200, map[string]any{"ok": true, "ignored": true, "error": "missing refid"}
	}

	row, err := s.repo.GetLatestByRefIDProvider(ctx, refID, "javapay")
	if err != nil {
		return 502, map[string]any{"ok": false, "error": err.Error()}
	}
	if row == nil {
		var qtyPtr *int64
		if qty > 0 {
			qtyPtr = &qty
		}
		dest = strings.TrimSpace(dest)
		var destPtr *string
		if dest != "" {
			destPtr = &dest
		}
		if err := s.repo.InsertAnomali(ctx, repository.ProviderAnomasiIn{
			Provider:   "javapay",
			RefID:      refID,
			RawQuery:   rawQuery,
			RawBody:    "",
			Payload:    payload,
			KodeRespon: &rcStr,
			Pesan:      &msg,
			Harga:      &price,
			Tujuan:     destPtr,
			Qty:        qtyPtr,
		}); err != nil {
			helper.AppendProviderServiceLog("provider_anomali.log", "javapay callback insert anomali failed unmatched refid=%s err=%v", refID, err)
		}
		saldoAnomali := lastBalance
		if saldoAnomali <= 0 {
			if bal, ok := helper.ExtractSaldoTerakhirFromMsg(msg); ok {
				saldoAnomali = bal
			}
		}
		s.insertAnomaliSnapshot(ctx, "javapay", refID, "callback_javapay_anomali", saldoAnomali, payload)
		return 200, map[string]any{"ok": true, "ignored": true, "refid": refID}
	}

	upd := repository.UpdateResult{ResponMentah: payload}
	httpStatus := 200
	upd.HTTPStatus = &httpStatus
	if trxid != "" {
		upd.TrxIDJavapay = &trxid
	}
	if rcStr != "" {
		upd.KodeRespon = &rcStr
	}
	if msg != "" {
		upd.Pesan = &msg
	}
	if noreff != "" {
		upd.NoReferensi = &noreff
	}
	upd.Harga = &price

	if lastBalance > 0 {
		upd.SaldoTerakhir = &lastBalance
	}
	if upd.SaldoTerakhir == nil {
		if bal, ok := helper.ExtractSaldoTerakhirFromMsg(msg); ok {
			upd.SaldoTerakhir = &bal
		}
	}

	if err := s.repo.UpdateResult(ctx, row.ID, upd); err != nil {
		helper.AppendProviderServiceLog("provider_callback_service.log", "javapay callback update result failed refid=%s rowID=%d err=%v", refID, row.ID, err)
	}

	if upd.SaldoTerakhir != nil && *upd.SaldoTerakhir > 0 {
		tmID := row.TransaksiMemberID
		tpID := row.ID
		rawJSON, _ := json.Marshal(payload)
		_ = s.repo.InsertProviderSnapshot(ctx, repository.ProviderSnapshotIn{
			Provider:            "javapay",
			SaldoProvider:       *upd.SaldoTerakhir,
			RefID:               refID,
			TransaksiMemberID:   &tmID,
			TransaksiProviderID: &tpID,
			Sumber:              "callback_javapay",
			RawJSON:             rawJSON,
		})
	}

	trx, err := s.repo.GetTransaksiMemberByID(ctx, row.TransaksiMemberID)
	if err != nil || trx == nil {
		return 200, map[string]any{"ok": true, "refid": refID}
	}
	finalStatus := string(helper.ClassifyJavapayResponseStatus(rcStr, msg))
	// Advisory lock: serialize concurrent callbacks for same transaction
	if locked, _ := s.repo.AcquireCallbackLock(ctx, trx.ID); locked {
		defer s.repo.ReleaseCallbackLock(ctx, trx.ID)
		if fresh, _ := s.repo.GetTransaksiMemberByID(ctx, trx.ID); fresh != nil {
			trx = fresh
		}
	}
	if strings.EqualFold(strings.TrimSpace(trx.Status), "success") && finalStatus == "failed" {
		if !helper.ShouldBlockProviderFallback(msg) {
			if fallbackStarted, provider, providerRowID, fbErr := s.tryFallbackFromJavapay(ctx, trx, msg); fallbackStarted {
				if fbErr != nil {
					helper.AppendProviderServiceLog("provider_callback_error.log", "javapay finalize fallback failed while member already success refid=%s trx_member_id=%d err=%v", refID, trx.ID, fbErr)
				} else {
					helper.AppendProviderServiceLog("provider_callback_service.log", "javapay provider-row failed after success; member kept success refid=%s trx_id=%d fallback_provider=%s provider_row_id=%d", refID, trx.ID, provider, providerRowID)
					return 200, map[string]any{"ok": true, "already_final": true, "refid": refID, "status": "success", "fallback_provider": provider, "provider_row_id": providerRowID}
				}
			}
		}
		helper.AppendAlreadyFinalLog("callback already_final provider=javapay refid=%s trx_member_id=%d status=success callback_status=failed", refID, trx.ID)
		return 200, map[string]any{"ok": true, "already_final": true, "refid": refID, "status": "success"}
	}

	if shouldKeepExistingMemberFinalStatus(trx.Status, finalStatus) {
		helper.AppendAlreadyFinalLog("callback already_final provider=javapay refid=%s trx_member_id=%d status=%s callback_status=%s", refID, trx.ID, trx.Status, finalStatus)
		return 200, map[string]any{"ok": true, "already_final": true, "refid": refID}
	}

	if finalStatus == "failed" && !helper.ShouldBlockProviderFallback(msg) {
		if fallbackStarted, provider, _, fbErr := s.tryFallbackFromJavapay(ctx, trx, msg); fallbackStarted {
			if fbErr != nil {
				helper.AppendProviderServiceLog("provider_callback_service.log", "javapay callback fallback error refid=%s target=%s err=%v", refID, provider, fbErr)
			} else {
				helper.AppendProviderServiceLog("provider_callback_service.log", "javapay callback fallback started refid=%s target=%s", refID, provider)
			}
			return 200, map[string]any{"ok": true, "refid": refID, "status": "pending", "fallback_provider": provider}
		}
	}

	ket, info := helper.SafeMemberKeterangan(finalStatus, msg)
	snOut := strings.TrimSpace(info.SN)
	if snOut == "" {
		snOut = strings.TrimSpace(sn)
	}
	if snOut == "" {
		snOut = strings.TrimSpace(noreff)
	}
	if snOut == "" {
		snOut = strings.TrimSpace(trxid)
	}

	var biayaAktual int64
	hargaMember := effectiveMemberSellingPrice(trx.HargaMember, trx.BiayaPerkiraan)
	switch finalStatus {
	case "success":
		if err := s.prepareMemberTrxSuccessTransition(ctx, "javapay", trx); err != nil {
			helper.AppendProviderServiceLog("provider_callback_error.log", "success promotion refund recovery failed provider=javapay refid=%s trx_member_id=%d err=%v", refID, trx.ID, err)
			return 200, map[string]any{"ok": false, "refid": refID, "error": err.Error(), "repair_needed": "member_refund_recovery"}
		}
		_ = s.syncJavapayWalletSuccess(ctx, refID, row, price, "auto debit by callback (success)")
		biayaAktual = trx.BiayaPerkiraan
		if err := s.repo.UpdateTransaksiMemberSettle(ctx, trx.ID, "success", snOut, biayaAktual, price, hargaMember); err != nil && err != sql.ErrNoRows {
			helper.AppendProviderServiceLog("provider_callback_error.log", "settle failed provider=javapay refid=%s trx_member_id=%d status=success err=%v", refID, trx.ID, err)
			return 200, map[string]any{"ok": false, "error": err.Error(), "repair_needed": "member_settle_failed"}
		}
		s.applyH2HCommission(ctx, trx.ID, refID)
	case "failed":
		ketDB := strings.TrimSpace(ket)
		if ketDB == "" || isWeakProviderValue(ketDB) {
			ketDB = strings.TrimSpace(snOut)
		}
		if ketDB == "" || isWeakProviderValue(ketDB) {
			ketDB = strings.TrimSpace(info.Reff)
		}
		if err := s.repo.UpdateTransaksiMemberSettle(ctx, trx.ID, "failed", ketDB, 0, price, hargaMember); err != nil && err != sql.ErrNoRows {
			helper.AppendProviderServiceLog("provider_callback_error.log", "settle failed provider=javapay refid=%s trx_member_id=%d status=failed err=%v", refID, trx.ID, err)
			return 200, map[string]any{"ok": false, "error": err.Error(), "repair_needed": "member_settle_failed"}
		}
		// Refund di-handle oleh DB trigger saat status di-update ke 'failed'.
	default:
		if err := s.repo.UpdateTransaksiMemberSettle(ctx, trx.ID, "pending", snOut, 0, price, hargaMember); err != nil && err != sql.ErrNoRows {
			helper.AppendProviderServiceLog("provider_callback_error.log", "settle failed provider=javapay refid=%s trx_member_id=%d status=pending err=%v", refID, trx.ID, err)
		}
	}

	if finalStatus == "pending" {
		return 200, map[string]any{"ok": true, "refid": refID, "status": "pending"}
	}

	webhookURL, err := s.repo.GetMemberWebhookURL(ctx, trx.MemberID)
	if err != nil || webhookURL == "" {
		return 200, map[string]any{"ok": true, "refid": refID}
	}

	memberSaldo := int64(0)
	if bal, err := s.repo.GetSaldo(ctx, trx.MemberID); err == nil {
		memberSaldo = bal
	}

	return s.sendDirectMemberWebhook(ctx, "javapay", trx, webhookURL, refID, finalStatus, ket, info.Reff, snOut, memberSaldo, func() int64 {
		if finalStatus == "success" {
			return biayaAktual
		}
		return 0
	}(), price)
}
