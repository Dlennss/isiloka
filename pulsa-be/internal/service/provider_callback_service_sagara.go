package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/url"
	"strconv"
	"strings"
	"time"

	"pulsa2/internal/helper"
	"pulsa2/internal/helper/providersn"
	"pulsa2/internal/repository"
	"pulsa2/sagaramobile"
)

func (s *ProviderCallbackService) ProcessSagaraCallback(ctx context.Context, rawQuery string, q url.Values) (int, map[string]any) {
	msg := strings.TrimSpace(q.Get("message"))
	statusRaw := strings.TrimSpace(q.Get("status"))
	if statusRaw == "" {
		statusRaw = strings.TrimSpace(sagaramobile.ExtractStatusCode(msg))
	}
	statusNum, _ := strconv.Atoi(statusRaw)
	refid := strings.TrimSpace(q.Get("refid"))
	if refid == "" {
		refid = helper.ExtractRefIDFromPesan(msg)
	}
	if refid == "" {
		refid = sagaramobile.ParseRefIDFromMsg(msg)
	}
	price := sagaramobile.ParsePriceFromMsg(msg)
	providerRefParsed, snParsed := providersn.ParseSagaraSNRefFromMsg(msg)
	lastBalance, hasLastBalance := sagaramobile.ParseLastBalanceFromMsg(msg)
	if !hasLastBalance {
		if bal, ok := helper.ExtractSaldoTerakhirFromMsg(msg); ok {
			lastBalance = bal
			hasLastBalance = true
		}
	}

	payload := map[string]any{"status": statusNum, "message": msg}
	if refid == "" {
		rc := strings.TrimSpace(statusRaw)
		if rc == "" && statusNum != 0 {
			rc = strconv.Itoa(statusNum)
		}
		var pricePtr *int64
		if price > 0 {
			pricePtr = &price
		}
		_ = s.repo.InsertAnomali(ctx, repository.ProviderAnomasiIn{Provider: "sagaramobile", RefID: "", KodeRespon: helper.PtrString(rc), Pesan: helper.PtrString(msg), Harga: pricePtr, RawQuery: rawQuery, RawBody: "", Payload: payload})
		if hasLastBalance && lastBalance > 0 {
			s.insertAnomaliSnapshot(ctx, "sagaramobile", refid, "callback_sagaramobile_anomali", lastBalance, payload)
		}
		return 200, map[string]any{"ok": true, "ignored": true, "error": "missing refid"}
	}

	row, err := s.repo.GetLatestByRefIDProvider(ctx, refid, "sagaramobile")
	if err != nil {
		return 502, map[string]any{"ok": false, "error": err.Error()}
	}
	if row == nil {
		rc := strings.TrimSpace(statusRaw)
		if rc == "" && statusNum != 0 {
			rc = strconv.Itoa(statusNum)
		}
		var pricePtr *int64
		if price > 0 {
			pricePtr = &price
		}
		_ = s.repo.InsertAnomali(ctx, repository.ProviderAnomasiIn{Provider: "sagaramobile", RefID: refid, KodeRespon: helper.PtrString(rc), Pesan: helper.PtrString(msg), Harga: pricePtr, RawQuery: rawQuery, RawBody: "", Payload: payload})
		if hasLastBalance && lastBalance > 0 {
			s.insertAnomaliSnapshot(ctx, "sagaramobile", refid, "callback_sagaramobile_anomali", lastBalance, payload)
		}
		return 200, map[string]any{"ok": true, "ignored": true, "refid": refid}
	}

	httpStatus := 200
	rcStr := strings.TrimSpace(statusRaw)
	if rcStr == "" {
		rcStr = strconv.Itoa(statusNum)
	}
	var noreffPtr *string
	if strings.TrimSpace(snParsed) != "" {
		noreffPtr = helper.PtrString(snParsed)
	} else if strings.TrimSpace(providerRefParsed) != "" {
		noreffPtr = helper.PtrString(providerRefParsed)
	}
	upd := repository.UpdateResult{HTTPStatus: &httpStatus, KodeRespon: helper.PtrString(rcStr), Pesan: helper.PtrString(msg), Harga: helper.PtrI64(price), NoReferensi: noreffPtr, ResponMentah: payload}
	if hasLastBalance && lastBalance > 0 {
		upd.SaldoTerakhir = &lastBalance
	}
	_ = s.repo.UpdateResult(ctx, row.ID, upd)

	if upd.SaldoTerakhir != nil && *upd.SaldoTerakhir > 0 {
		tmID := row.TransaksiMemberID
		tpID := row.ID
		rawJSON, _ := json.Marshal(payload)
		_ = s.repo.InsertProviderSnapshot(ctx, repository.ProviderSnapshotIn{Provider: "sagaramobile", SaldoProvider: *upd.SaldoTerakhir, RefID: refid, TransaksiMemberID: &tmID, TransaksiProviderID: &tpID, Sumber: "callback_sagaramobile", RawJSON: rawJSON})
	}

	trx, err := s.repo.GetTransaksiMemberByID(ctx, row.TransaksiMemberID)
	if err != nil || trx == nil {
		return 200, map[string]any{"ok": true, "refid": refid}
	}
	finalStatus := resolveSagaraFinalStatus(statusNum, rcStr, msg)
	// Advisory lock: serialize concurrent callbacks for same transaction
	if locked, _ := s.repo.AcquireCallbackLock(ctx, trx.ID); locked {
		defer s.repo.ReleaseCallbackLock(ctx, trx.ID)
		if fresh, _ := s.repo.GetTransaksiMemberByID(ctx, trx.ID); fresh != nil {
			trx = fresh
		}
	}
	if strings.EqualFold(strings.TrimSpace(trx.Status), "success") && finalStatus == "failed" {
		if !helper.ShouldBlockProviderFallback(msg) {
			fbCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
			defer cancel()
			did, provider, providerRowID, ferr := s.tryFallbackFromSagara(fbCtx, trx, msg)
			if did {
				helper.AppendProviderServiceLog("provider_callback_service.log", "sagaramobile provider-row failed after success; member kept success refid=%s trx_id=%d fallback_provider=%s provider_row_id=%d", refid, trx.ID, provider, providerRowID)
				return 200, map[string]any{"ok": true, "already_final": true, "refid": refid, "status": "success", "fallback_provider": provider, "provider_row_id": providerRowID}
			}
			if ferr != nil {
				helper.AppendProviderServiceLog("provider_callback_error.log", "sagaramobile finalize fallback failed while member already success refid=%s trx_member_id=%d err=%v", refid, trx.ID, ferr)
			}
		}
		var pricePtr *int64
		if price > 0 {
			pricePtr = &price
		}
		_ = s.repo.InsertAnomali(ctx, repository.ProviderAnomasiIn{Provider: "sagaramobile", RefID: refid, KodeRespon: helper.PtrString(rcStr), Pesan: helper.PtrString(msg), Harga: pricePtr, RawQuery: rawQuery, RawBody: "", Payload: payload})
		helper.AppendProviderServiceLog("provider_callback_service.log", "sagaramobile late failed callback after member success fallback exhausted; stored as anomaly refid=%s trx_id=%d provider_row_id=%d rc=%s", refid, trx.ID, row.ID, rcStr)
		helper.AppendAlreadyFinalLog("callback already_final provider=sagaramobile refid=%s trx_member_id=%d status=success callback_status=failed suspect=1", refid, trx.ID)
		return 200, map[string]any{"ok": true, "already_final": true, "refid": refid, "status": "success", "suspect": true}
	}

	if shouldKeepExistingMemberFinalStatus(trx.Status, finalStatus) {
		helper.AppendAlreadyFinalLog("callback already_final provider=sagaramobile refid=%s trx_member_id=%d status=%s callback_status=%s", refid, trx.ID, trx.Status, finalStatus)
		return 200, map[string]any{"ok": true, "already_final": true, "refid": refid}
	}
	if finalStatus == "pending" {
		ket, _ := helper.SafeMemberKeterangan("pending", msg)
		ketDB := strings.TrimSpace(snParsed)
		if ketDB == "" {
			ketDB = ket
		}
		_ = s.repo.UpdateTransaksiMemberSettle(ctx, trx.ID, "pending", ketDB, 0, price, trx.HargaMember)
		return 200, map[string]any{"ok": true, "refid": refid, "status": "pending"}
	}

	if finalStatus == "failed" && !helper.ShouldBlockProviderFallback(msg) {
		fbCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
		defer cancel()
		did, provider, providerRowID, ferr := s.tryFallbackFromSagara(fbCtx, trx, msg)
		if did {
			return 200, map[string]any{"ok": true, "refid": refid, "status": "pending", "fallback": provider, "provider_row_id": providerRowID, "error": func() any {
				if ferr != nil {
					return ferr.Error()
				}
				return nil
			}()}
		}
		if ferr != nil {
			msg = fallbackFinalFailureMessage(msg, ferr)
		}
	}

	ket, info := helper.SafeMemberKeterangan(finalStatus, msg)
	ketDB := strings.TrimSpace(snParsed)
	if ketDB == "" {
		ketDB = ket
	}

	hargaMember := effectiveMemberSellingPrice(trx.HargaMember, trx.BiayaPerkiraan)
	switch finalStatus {
	case "success":
		if err := s.prepareMemberTrxSuccessTransition(ctx, "sagaramobile", trx); err != nil {
			helper.AppendProviderServiceLog("provider_callback_error.log", "success promotion refund recovery failed provider=sagaramobile refid=%s trx_member_id=%d err=%v", refid, trx.ID, err)
			return 200, map[string]any{"ok": false, "refid": refid, "error": err.Error(), "repair_needed": "member_refund_recovery"}
		}
		biayaAktual := trx.BiayaPerkiraan
		if err := s.repo.UpdateTransaksiMemberSettle(ctx, trx.ID, "success", ketDB, biayaAktual, price, hargaMember); err != nil && err != sql.ErrNoRows {
			helper.AppendProviderServiceLog("provider_callback_error.log", "settle failed provider=sagaramobile refid=%s trx_member_id=%d status=success err=%v", refid, trx.ID, err)
			return 200, map[string]any{"ok": false, "error": err.Error(), "repair_needed": "member_settle_failed"}
		}
		s.applyH2HCommission(ctx, trx.ID, refid)
		// Debit dompet provider di-handle oleh DB trigger.
	case "failed":
		ketFailed := strings.TrimSpace(ket)
		if ketFailed == "" || isWeakProviderValue(ketFailed) {
			ketFailed = ketDB
		}
		if err := s.repo.UpdateTransaksiMemberSettle(ctx, trx.ID, "failed", ketFailed, 0, price, hargaMember); err != nil && err != sql.ErrNoRows {
			helper.AppendProviderServiceLog("provider_callback_error.log", "settle failed provider=sagaramobile refid=%s trx_member_id=%d status=failed err=%v", refid, trx.ID, err)
			return 200, map[string]any{"ok": false, "error": err.Error(), "repair_needed": "member_settle_failed"}
		}
		// Refund di-handle oleh DB trigger saat status di-update ke 'failed'.
	}

	webhookURL, err := s.repo.GetMemberWebhookURL(ctx, trx.MemberID)
	if err != nil || strings.TrimSpace(webhookURL) == "" {
		return 200, map[string]any{"ok": true, "refid": refid}
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
	snOut := strings.TrimSpace(info.SN)
	if providerRefOut == "" || isWeakProviderValue(providerRefOut) {
		providerRefOut = strings.TrimSpace(providerRefParsed)
	}
	if snOut == "" || isWeakProviderValue(snOut) {
		snOut = strings.TrimSpace(snParsed)
	}
	if providerRefOut == "" || isWeakProviderValue(providerRefOut) {
		providerRefOut = snOut
	}
	if strings.EqualFold(finalStatus, "success") {
		if providerRefOut != "" && !isWeakProviderValue(providerRefOut) {
			ket = "Transaksi berhasil. Ref: " + providerRefOut
		} else {
			ket = "Transaksi berhasil"
		}
	}
	return s.sendDirectMemberWebhook(ctx, "sagaramobile", trx, webhookURL, refid, finalStatus, ket, providerRefOut, snOut, memberSaldo, biayaAktualOut, price)
}
