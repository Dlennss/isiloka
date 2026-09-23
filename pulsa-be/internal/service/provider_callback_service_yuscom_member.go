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
)

func (s *ProviderCallbackService) ProcessYuscomCallback(ctx context.Context, rawQuery string, q url.Values) (int, map[string]any) {
	refid := strings.TrimSpace(q.Get("refid"))
	statusRaw := strings.TrimSpace(q.Get("status"))
	statusNum, _ := strconv.Atoi(statusRaw)
	price, _ := strconv.ParseInt(strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(q.Get("price")), ".", ""), ",", ""), 10, 64)
	msg := strings.TrimSpace(q.Get("message"))

	if refid == "" {
		rc := strconv.Itoa(statusNum)
		var pricePtr *int64
		if price > 0 {
			pricePtr = &price
		}
		anomaliPayload := map[string]any{
			"t":       q.Get("t"),
			"refid":   q.Get("refid"),
			"status":  statusNum,
			"price":   price,
			"message": msg,
		}
		_ = s.repo.InsertAnomali(ctx, repository.ProviderAnomasiIn{
			Provider:   "yuscom",
			RefID:      "",
			KodeRespon: &rc,
			Pesan:      &msg,
			Harga:      pricePtr,
			RawQuery:   rawQuery,
			RawBody:    "",
			Payload:    anomaliPayload,
		})
		saldoAnomali := int64(0)
		if bal, ok := helper.ParseSaldoAfterFromYuscomMsg(msg); ok {
			saldoAnomali = bal
		}
		s.insertAnomaliSnapshot(ctx, "yuscom", refid, "callback_yuscom_anomali", saldoAnomali, anomaliPayload)
		return 200, map[string]any{"ok": true, "ignored": true, "error": "missing refid"}
	}

	row, err := s.repo.GetLatestByRefIDProvider(ctx, refid, "yuscom")
	if err != nil {
		return 502, map[string]any{"ok": false, "error": err.Error()}
	}
	if row == nil {
		appRow, appErr := s.appProviderRepo.GetByRefID(ctx, refid, "yuscom")
		if appErr == nil && appRow != nil {
			helper.AppendProviderServiceLog("provider_callback_service.log", "yuscom callback matched app_order_provider_trx refid=%s app_provider_id=%d app_order_id=%d", refid, appRow.ID, appRow.AppOrderID)
			return s.processYuscomAppCallback(ctx, refid, statusNum, price, msg, q)
		}
		if appErr != nil && appErr != sql.ErrNoRows {
			helper.AppendProviderServiceLog("provider_callback_service.log", "yuscom callback app lookup error refid=%s err=%v", refid, appErr)
			return 502, map[string]any{"ok": false, "error": appErr.Error()}
		}
		checkRow, checkErr := s.billingCheckRepo.GetByRefID(ctx, refid)
		if checkErr == nil && checkRow != nil {
			helper.AppendProviderServiceLog("provider_callback_service.log", "yuscom callback matched app_billing_check refid=%s billing_check_id=%d", refid, checkRow.ID)
			return s.processYuscomBillingCheckCallback(ctx, refid, statusNum, price, msg, q)
		}
		if checkErr != nil && checkErr != sql.ErrNoRows {
			helper.AppendProviderServiceLog("provider_callback_service.log", "yuscom callback billing check lookup error refid=%s err=%v", refid, checkErr)
			return 502, map[string]any{"ok": false, "error": checkErr.Error()}
		}
		helper.AppendProviderServiceLog("provider_anomali.log", "yuscom callback unmatched refid=%s member_row=nil app_row=nil", refid)
		rc := strconv.Itoa(statusNum)
		var pricePtr *int64
		if price > 0 {
			pricePtr = &price
		}
		anomaliPayload := map[string]any{
			"t":       q.Get("t"),
			"refid":   refid,
			"status":  statusNum,
			"price":   price,
			"message": msg,
		}
		_ = s.repo.InsertAnomali(ctx, repository.ProviderAnomasiIn{
			Provider:   "yuscom",
			RefID:      refid,
			KodeRespon: &rc,
			Pesan:      &msg,
			Harga:      pricePtr,
			RawQuery:   rawQuery,
			RawBody:    "",
			Payload:    anomaliPayload,
		})
		saldoAnomali := int64(0)
		if bal, ok := helper.ParseSaldoAfterFromYuscomMsg(msg); ok {
			saldoAnomali = bal
		}
		s.insertAnomaliSnapshot(ctx, "yuscom", refid, "callback_yuscom_anomali", saldoAnomali, anomaliPayload)
		return 200, map[string]any{"ok": true, "ignored": true, "refid": refid}
	}

	httpStatus := 200
	rcStr := strings.TrimSpace(statusRaw)
	if rcStr == "" && helper.IsRetryableFailRC(helper.ExtractProviderStatusCode(msg)) {
		rcStr = helper.ExtractProviderStatusCode(msg)
	}
	if rcStr == "" && statusNum != 0 {
		rcStr = strconv.Itoa(statusNum)
	}
	var saldoTerakhir *int64
	if bal, ok := helper.ParseSaldoAfterFromYuscomMsg(msg); ok {
		saldoTerakhir = &bal
	}
	noreff := helper.ParseTicketFromYuscomMsg(msg)
	if strings.TrimSpace(noreff) == "" {
		pr, sn := providersn.ParseMultikomSNRefFromMsg(msg)
		if strings.TrimSpace(sn) != "" {
			noreff = strings.TrimSpace(sn)
		} else if strings.TrimSpace(pr) != "" {
			noreff = strings.TrimSpace(pr)
		}
	}
	var noreffPtr *string
	if strings.TrimSpace(noreff) != "" {
		noreffPtr = &noreff
	}
	upd := repository.UpdateResult{
		HTTPStatus:    &httpStatus,
		KodeRespon:    &rcStr,
		Pesan:         &msg,
		Harga:         &price,
		NoReferensi:   noreffPtr,
		SaldoTerakhir: saldoTerakhir,
		ResponMentah: map[string]any{
			"t":       q.Get("t"),
			"refid":   refid,
			"status":  statusNum,
			"price":   price,
			"message": msg,
		},
	}
	_ = s.repo.UpdateResult(ctx, row.ID, upd)

	if upd.SaldoTerakhir != nil && *upd.SaldoTerakhir > 0 {
		tmID := row.TransaksiMemberID
		tpID := row.ID
		rawJSON, _ := json.Marshal(upd.ResponMentah)
		_ = s.repo.InsertProviderSnapshot(ctx, repository.ProviderSnapshotIn{
			Provider:            "yuscom",
			SaldoProvider:       *upd.SaldoTerakhir,
			RefID:               refid,
			TransaksiMemberID:   &tmID,
			TransaksiProviderID: &tpID,
			Sumber:              "callback_yuscom",
			RawJSON:             rawJSON,
		})
	}

	trx, err := s.repo.GetTransaksiMemberByID(ctx, row.TransaksiMemberID)
	if err != nil || trx == nil {
		return 200, map[string]any{"ok": true, "refid": refid}
	}
	finalStatus := resolveYuscomFinalStatus(statusNum, rcStr, msg)
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
			did, provider, providerRowID, ferr := s.tryFallbackFromYuscom(fbCtx, trx, msg)
			if did {
				helper.AppendProviderServiceLog("provider_callback_service.log", "yuscom provider-row failed after success; member kept success refid=%s trx_id=%d fallback_provider=%s provider_row_id=%d", refid, trx.ID, provider, providerRowID)
				return 200, map[string]any{"ok": true, "already_final": true, "refid": refid, "status": "success", "fallback_provider": provider, "provider_row_id": providerRowID}
			}
			if ferr != nil {
				helper.AppendProviderServiceLog("provider_callback_error.log", "yuscom finalize fallback failed while member already success refid=%s trx_member_id=%d err=%v", refid, trx.ID, ferr)
			}
		}
		helper.AppendAlreadyFinalLog("callback already_final provider=yuscom refid=%s trx_member_id=%d status=success callback_status=failed", refid, trx.ID)
		return 200, map[string]any{"ok": true, "already_final": true, "refid": refid, "status": "success"}
	}

	if shouldKeepExistingMemberFinalStatus(trx.Status, finalStatus) {
		helper.AppendAlreadyFinalLog("callback already_final provider=yuscom refid=%s trx_member_id=%d status=%s callback_status=%s", refid, trx.ID, trx.Status, finalStatus)
		return 200, map[string]any{"ok": true, "already_final": true, "refid": refid}
	}

	_, snParsed := providersn.ParseYuscomSNRefFromMsg(msg)
	snParsed = strings.TrimSpace(snParsed)

	if finalStatus == "pending" {
		ket, _ := helper.SafeMemberKeterangan("pending", msg)
		ketDB := snParsed
		if ketDB == "" {
			ketDB = ket
		}
		_ = s.repo.UpdateTransaksiMemberSettle(ctx, trx.ID, "pending", ketDB, 0, price, trx.HargaMember)
		return 200, map[string]any{"ok": true, "refid": refid, "status": "pending"}
	}

	if finalStatus == "failed" && !helper.ShouldBlockProviderFallback(msg) {
		fbCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
		defer cancel()
		did, provider, providerRowID, ferr := s.tryFallbackFromYuscom(fbCtx, trx, msg)
		if did {
			return 200, map[string]any{
				"ok":              true,
				"refid":           refid,
				"status":          "pending",
				"fallback":        provider,
				"provider_row_id": providerRowID,
				"error": func() any {
					if ferr != nil {
						return ferr.Error()
					}
					return nil
				}(),
			}
		}
		if ferr != nil {
			msg = fallbackFinalFailureMessage(msg, ferr)
		}
	}

	ket, info := helper.SafeMemberKeterangan(finalStatus, msg)
	ketDB := snParsed
	if ketDB == "" {
		ketDB = ket
	}

	hargaMember := effectiveMemberSellingPrice(trx.HargaMember, trx.BiayaPerkiraan)
	switch finalStatus {
	case "success":
		if err := s.prepareMemberTrxSuccessTransition(ctx, "yuscom", trx); err != nil {
			helper.AppendProviderServiceLog("provider_callback_error.log", "success promotion refund recovery failed provider=yuscom refid=%s trx_member_id=%d err=%v", refid, trx.ID, err)
			return 200, map[string]any{"ok": false, "refid": refid, "error": err.Error(), "repair_needed": "member_refund_recovery"}
		}
		biayaAktual := trx.BiayaPerkiraan
		if err := s.repo.UpdateTransaksiMemberSettle(ctx, trx.ID, "success", ketDB, biayaAktual, price, hargaMember); err != nil && err != sql.ErrNoRows {
			helper.AppendProviderServiceLog("provider_callback_error.log", "settle failed provider=yuscom refid=%s trx_member_id=%d status=success err=%v", refid, trx.ID, err)
			return 200, map[string]any{"ok": false, "error": err.Error(), "repair_needed": "member_settle_failed"}
		}
		s.applyH2HCommission(ctx, trx.ID, refid)
		// Debit dompet provider di-handle oleh DB trigger.
	case "failed":
		if err := s.repo.UpdateTransaksiMemberSettle(ctx, trx.ID, "failed", ketDB, 0, price, hargaMember); err != nil && err != sql.ErrNoRows {
			helper.AppendProviderServiceLog("provider_callback_error.log", "settle failed provider=yuscom refid=%s trx_member_id=%d status=failed err=%v", refid, trx.ID, err)
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
	switch finalStatus {
	case "success":
		biayaAktualOut = trx.BiayaPerkiraan
	}
	providerRefOut := strings.TrimSpace(info.Reff)
	snOut := strings.TrimSpace(info.SN)
	if providerRefOut == "" || snOut == "" || isWeakProviderValue(providerRefOut) || isWeakProviderValue(snOut) {
		pr, sn := providersn.ParseYuscomSNRefFromMsg(msg)
		if providerRefOut == "" || isWeakProviderValue(providerRefOut) {
			providerRefOut = strings.TrimSpace(pr)
		}
		if snOut == "" || isWeakProviderValue(snOut) {
			snOut = strings.TrimSpace(sn)
		}
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

	return s.sendDirectMemberWebhook(ctx, "yuscom", trx, webhookURL, refid, finalStatus, ket, providerRefOut, snOut, memberSaldo, biayaAktualOut, price)
}
