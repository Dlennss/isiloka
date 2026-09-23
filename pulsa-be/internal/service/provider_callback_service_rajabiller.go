package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/url"
	"strconv"
	"strings"

	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
)

func parseRajabillerCallbackPayload(raw string, q url.Values) (map[string]any, string, string, string, string, int64, *int64) {
	payload := map[string]any{}
	trimmed := strings.TrimSpace(raw)
	if strings.HasPrefix(trimmed, "{") {
		dec := json.NewDecoder(strings.NewReader(trimmed))
		dec.UseNumber()
		_ = dec.Decode(&payload)
	}

	getAny := func(keys ...string) string {
		for _, key := range keys {
			if v := strings.TrimSpace(q.Get(key)); v != "" {
				return v
			}
			if rawV, ok := payload[key]; ok {
				if out := rajabillerPayloadValueString(rawV); out != "" {
					return out
				}
			}
		}
		return ""
	}
	parseI64 := func(keys ...string) int64 {
		v := strings.TrimSpace(getAny(keys...))
		if v == "" {
			return 0
		}
		clean := strings.ReplaceAll(strings.ReplaceAll(v, ".", ""), ",", "")
		n, _ := strconv.ParseInt(clean, 10, 64)
		return n
	}

	refid := getAny("trxid", "ref1", "ref_id", "refid_client", "client_ref", "order_id")
	if refid == "" {
		refid = getAny("ref", "refid")
	}
	rc := getAny("rc", "response_code", "kode_respon", "kode_response", "code")
	msg := strings.TrimSpace(getAny("status", "keterangan", "message", "pesan", "msg", "description"))
	if msg == "" {
		msg = trimmed
	}
	providerRef := strings.TrimSpace(getAny("sn", "token", "ref", "refid", "no_ref", "noref", "provider_ref", "serial_number"))
	price := parseI64("harga", "price", "saldo_terpotong", "total_bayar", "total", "nominal", "amount")
	var saldoPtr *int64
	if saldo := parseI64("saldo_akhir", "sisa_saldo", "saldo", "balance", "last_balance"); saldo > 0 {
		saldoPtr = &saldo
	}
	if len(payload) == 0 {
		payload = map[string]any{
			"refid":        refid,
			"rc":           rc,
			"status":       msg,
			"provider_ref": providerRef,
			"harga":        price,
		}
		if tujuan := strings.TrimSpace(getAny("idpel", "tujuan", "dest", "target", "msisdn", "nomor")); tujuan != "" {
			payload["tujuan"] = tujuan
		}
		if produk := strings.TrimSpace(getAny("produk", "kode_produk", "kodeproduk")); produk != "" {
			payload["produk"] = produk
		}
		if saldoPtr != nil {
			payload["saldo_akhir"] = *saldoPtr
		}
	}
	return payload, refid, rc, msg, providerRef, price, saldoPtr
}

func rajabillerPayloadValueString(rawV any) string {
	switch vv := rawV.(type) {
	case string:
		return strings.TrimSpace(vv)
	case float64:
		return strconv.FormatInt(int64(vv), 10)
	case json.Number:
		return strings.TrimSpace(vv.String())
	case int64:
		return strconv.FormatInt(vv, 10)
	case int:
		return strconv.Itoa(vv)
	case bool:
		if vv {
			return "true"
		}
		return "false"
	default:
		return ""
	}
}

func rajabillerPayloadString(payload map[string]any, keys ...string) string {
	for _, key := range keys {
		rawV, ok := payload[key]
		if !ok {
			continue
		}
		if out := rajabillerPayloadValueString(rawV); out != "" {
			return out
		}
	}
	return ""
}

func rajabillerPayloadInt64(payload map[string]any, keys ...string) int64 {
	v := strings.TrimSpace(rajabillerPayloadString(payload, keys...))
	if v == "" {
		return 0
	}
	clean := strings.ReplaceAll(strings.ReplaceAll(v, ".", ""), ",", "")
	n, _ := strconv.ParseInt(clean, 10, 64)
	return n
}

func resolveRajabillerFinalStatus(rcStr, msg string) string {
	switch helper.ProviderResponseStateOf("rajabiller", rcStr, msg) {
	case helper.ProviderResponseSuccess:
		return "success"
	case helper.ProviderResponseFailed:
		return "failed"
	case helper.ProviderResponsePending:
		return "pending"
	default:
		return "pending"
	}
}

func (s *ProviderCallbackService) insertRajabillerCallbackAnomali(ctx context.Context, refid string, payload map[string]any, rawQuery, rcStr, msg string, price int64) {
	var pricePtr *int64
	if price > 0 {
		pricePtr = &price
	}
	anomaliIn := repository.ProviderAnomasiIn{
		Provider:   "rajabiller",
		RefID:      refid,
		KodeRespon: helper.PtrString(rcStr),
		Pesan:      helper.PtrString(msg),
		Harga:      pricePtr,
		RawQuery:   rawQuery,
		RawBody:    rawQuery,
		Payload:    payload,
	}
	if tujuan := strings.TrimSpace(rajabillerPayloadString(payload, "idpel", "tujuan", "dest", "target", "msisdn", "nomor")); tujuan != "" {
		anomaliIn.Tujuan = &tujuan
	}
	if qty := rajabillerPayloadInt64(payload, "qty", "nominal", "amount"); qty > 0 {
		anomaliIn.Qty = &qty
	}
	_ = s.repo.InsertAnomali(ctx, anomaliIn)
}

func (s *ProviderCallbackService) ProcessRajabillerCallback(ctx context.Context, rawQuery string, q url.Values) (int, map[string]any) {
	payload, refid, rcStr, msg, providerRef, price, saldoTerakhir := parseRajabillerCallbackPayload(rawQuery, q)
	finalStatus := resolveRajabillerFinalStatus(rcStr, msg)

	if refid == "" {
		s.insertRajabillerCallbackAnomali(ctx, "", payload, rawQuery, rcStr, msg, price)
		if saldoTerakhir != nil && *saldoTerakhir > 0 {
			s.insertAnomaliSnapshot(ctx, "rajabiller", "", "callback_rajabiller_anomali", *saldoTerakhir, payload)
		}
		return 200, map[string]any{"ok": true, "ignored": true, "error": "missing refid"}
	}

	row, err := s.repo.GetLatestByRefIDProvider(ctx, refid, "rajabiller")
	if err != nil {
		return 502, map[string]any{"ok": false, "error": err.Error()}
	}
	if row == nil {
		s.insertRajabillerCallbackAnomali(ctx, refid, payload, rawQuery, rcStr, msg, price)
		if saldoTerakhir != nil && *saldoTerakhir > 0 {
			s.insertAnomaliSnapshot(ctx, "rajabiller", refid, "callback_rajabiller_anomali", *saldoTerakhir, payload)
		}
		return 200, map[string]any{"ok": true, "ignored": true, "refid": refid}
	}

	httpStatus := 200
	providerRef = strings.TrimSpace(providerRef)
	var noreffPtr *string
	if providerRef != "" {
		noreffPtr = &providerRef
	}
	upd := repository.UpdateResult{HTTPStatus: &httpStatus, KodeRespon: helper.PtrString(rcStr), Pesan: helper.PtrString(msg), Harga: helper.PtrI64(price), NoReferensi: noreffPtr, ResponMentah: payload}
	if saldoTerakhir != nil && *saldoTerakhir > 0 {
		upd.SaldoTerakhir = saldoTerakhir
	}
	_ = s.repo.UpdateResult(ctx, row.ID, upd)

	if upd.SaldoTerakhir != nil && *upd.SaldoTerakhir > 0 {
		tmID := row.TransaksiMemberID
		tpID := row.ID
		rawJSON, _ := json.Marshal(payload)
		_ = s.repo.InsertProviderSnapshot(ctx, repository.ProviderSnapshotIn{Provider: "rajabiller", SaldoProvider: *upd.SaldoTerakhir, RefID: refid, TransaksiMemberID: &tmID, TransaksiProviderID: &tpID, Sumber: "callback_rajabiller", RawJSON: rawJSON})
	}

	trx, err := s.repo.GetTransaksiMemberByID(ctx, row.TransaksiMemberID)
	if err != nil || trx == nil {
		return 200, map[string]any{"ok": true, "refid": refid}
	}
	if locked, _ := s.repo.AcquireCallbackLock(ctx, trx.ID); locked {
		defer s.repo.ReleaseCallbackLock(ctx, trx.ID)
		if fresh, _ := s.repo.GetTransaksiMemberByID(ctx, trx.ID); fresh != nil {
			trx = fresh
		}
	}
	if strings.EqualFold(strings.TrimSpace(trx.Status), "success") && finalStatus == "failed" && s.isBankH2HProduct(ctx, trx.KodeProduk) {
		ok, providerName, _, ferr := s.tryFallbackFromProvider(ctx, trx, "rajabiller", "rajabiller_callback_fallback", msg)
		if ok && ferr == nil {
			helper.AppendProviderServiceLog("provider_callback_service.log",
				"Rajabiller provider-row failed after success; member kept success refid=%s trx_id=%d fallback_provider=%s", refid, trx.ID, providerName)
			return 200, map[string]any{"ok": true, "already_final": true, "refid": refid, "status": "success", "fallback_provider": providerName}
		}
		if ferr != nil {
			helper.AppendProviderServiceLog("provider_callback_error.log", "rajabiller fallback failed while member already success refid=%s trx_member_id=%d err=%v", refid, trx.ID, ferr)
		}
		helper.AppendAlreadyFinalLog("callback already_final provider=rajabiller refid=%s trx_member_id=%d status=success callback_status=failed", refid, trx.ID)
		return 200, map[string]any{"ok": true, "already_final": true, "refid": refid, "status": "success"}
	}
	if shouldKeepExistingMemberFinalStatus(trx.Status, finalStatus) {
		helper.AppendAlreadyFinalLog("callback already_final provider=rajabiller refid=%s trx_member_id=%d status=%s callback_status=%s", refid, trx.ID, trx.Status, finalStatus)
		return 200, map[string]any{"ok": true, "already_final": true, "refid": refid}
	}

	ket, info := helper.SafeMemberKeterangan(finalStatus, msg)
	ketDB := strings.TrimSpace(providerRef)
	if ketDB == "" {
		ketDB = ket
	}
	hargaMember := effectiveMemberSellingPrice(trx.HargaMember, trx.BiayaPerkiraan)
	switch finalStatus {
	case "pending":
		_ = s.repo.UpdateTransaksiMemberSettle(ctx, trx.ID, "pending", ketDB, 0, price, trx.HargaMember)
		return 200, map[string]any{"ok": true, "refid": refid, "status": "pending"}
	case "success":
		if err := s.prepareMemberTrxSuccessTransition(ctx, "rajabiller", trx); err != nil {
			helper.AppendProviderServiceLog("provider_callback_error.log", "success promotion refund recovery failed provider=rajabiller refid=%s trx_member_id=%d err=%v", refid, trx.ID, err)
			return 200, map[string]any{"ok": false, "refid": refid, "error": err.Error(), "repair_needed": "member_refund_recovery"}
		}
		biayaAktual := trx.BiayaPerkiraan
		if err := s.repo.UpdateTransaksiMemberSettle(ctx, trx.ID, "success", ketDB, biayaAktual, price, hargaMember); err != nil && err != sql.ErrNoRows {
			helper.AppendProviderServiceLog("provider_callback_error.log", "settle failed provider=rajabiller refid=%s trx_member_id=%d status=success err=%v", refid, trx.ID, err)
			return 200, map[string]any{"ok": false, "error": err.Error(), "repair_needed": "member_settle_failed"}
		}
		s.applyH2HCommission(ctx, trx.ID, refid)
	case "failed":
		if s.isBankH2HProduct(ctx, trx.KodeProduk) {
			ok, providerName, _, ferr := s.tryFallbackFromProvider(ctx, trx, "rajabiller", "rajabiller_callback_fallback", msg)
			if ok && ferr == nil {
				waitReason := "wait_provider_status"
				if strings.EqualFold(strings.TrimSpace(providerName), "loketbayar") {
					waitReason = "wait_loketbayar_status"
				}
				ketPending, _ := helper.SafeMemberKeterangan("pending", waitReason)
				pendErr := s.repo.UpdateTransaksiMemberSettle(ctx, trx.ID, "pending", ketPending, 0, price, hargaMember)
				if pendErr != nil && pendErr != sql.ErrNoRows {
					helper.AppendProviderServiceLog("provider_callback_error.log", "rajabiller reopen pending failed refid=%s trx_member_id=%d provider=%s err=%v", refid, trx.ID, providerName, pendErr)
					return 502, map[string]any{"ok": false, "refid": refid, "error": pendErr.Error(), "repair_needed": "member_reopen_pending"}
				}
				return 200, map[string]any{"ok": true, "refid": refid, "status": "pending", "fallback_provider": providerName}
			}
			if ferr != nil {
				helper.AppendProviderServiceLog("provider_callback_error.log", "rajabiller finalize fallback failed refid=%s trx_member_id=%d err=%v", refid, trx.ID, ferr)
			}
		}
		ketFailed := strings.TrimSpace(ket)
		if ketFailed == "" || isWeakProviderValue(ketFailed) {
			ketFailed = ketDB
		}
		if err := s.repo.UpdateTransaksiMemberSettle(ctx, trx.ID, "failed", ketFailed, 0, price, hargaMember); err != nil && err != sql.ErrNoRows {
			helper.AppendProviderServiceLog("provider_callback_error.log", "settle failed provider=rajabiller refid=%s trx_member_id=%d status=failed err=%v", refid, trx.ID, err)
			return 200, map[string]any{"ok": false, "error": err.Error(), "repair_needed": "member_settle_failed"}
		}
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
		providerRefOut = providerRef
	}
	if snOut == "" || isWeakProviderValue(snOut) {
		snOut = providerRef
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
	return s.sendDirectMemberWebhook(ctx, "rajabiller", trx, webhookURL, refid, finalStatus, ket, providerRefOut, snOut, memberSaldo, biayaAktualOut, price)
}
