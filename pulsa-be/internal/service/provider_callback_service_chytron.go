package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
)

var reChytronSaldoDebit = regexp.MustCompile(`(?i)\bsaldo\s*:?\s*[0-9][0-9.,]*\s*-\s*([0-9][0-9.,]*)\s*=`)

func parseChytronCallbackPayload(raw string, q url.Values) (map[string]any, string, string, string, string, int64, *int64) {
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
				switch vv := rawV.(type) {
				case string:
					if out := strings.TrimSpace(vv); out != "" {
						return out
					}
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

	refid := getAny("idtrx", "clientid", "client_id", "refid", "ref_id", "reffid", "trxid", "trx_id", "reference_id", "order_id")
	rc := getAny("rc", "status", "statuscode", "status_code", "kode_respon", "kode_response", "code")
	msg := strings.TrimSpace(getAny("message", "msg", "pesan", "keterangan", "description", "ket"))
	if msg == "" {
		msg = trimmed
	}
	providerRef := strings.TrimSpace(getAny("sn", "noref", "no_ref", "noreff", "reff", "ref", "provider_ref", "serial_number"))
	price := parseI64("harga", "hrg", "price", "nominal", "amount", "total")
	if price <= 0 {
		price = parseChytronDebitedAmount(msg)
	}
	var saldoPtr *int64
	if saldo := parseI64("saldo", "balance", "sisasaldo", "saldo_akhir", "last_balance"); saldo > 0 {
		saldoPtr = &saldo
	}
	if len(payload) == 0 {
		payload = map[string]any{
			"refid":        refid,
			"status":       rc,
			"message":      msg,
			"provider_ref": providerRef,
			"price":        price,
		}
		if tujuan := strings.TrimSpace(getAny("tujuan", "dest", "target", "msisdn", "nomor")); tujuan != "" {
			payload["tujuan"] = tujuan
		}
		if produk := strings.TrimSpace(getAny("kp", "kodeproduk", "kode_produk", "produk")); produk != "" {
			payload["kode_produk"] = produk
		}
		if saldoPtr != nil {
			payload["saldo"] = *saldoPtr
		}
	}
	return payload, refid, rc, msg, providerRef, price, saldoPtr
}

func parseChytronDebitedAmount(msg string) int64 {
	msg = strings.TrimSpace(msg)
	if msg == "" {
		return 0
	}
	m := reChytronSaldoDebit.FindStringSubmatch(msg)
	if len(m) != 2 {
		return 0
	}
	clean := strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(m[1]), ".", ""), ",", "")
	n, _ := strconv.ParseInt(clean, 10, 64)
	return n
}

func normalizeChytronRC(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	raw = strings.TrimLeft(raw, "0")
	if raw == "" {
		return "0"
	}
	return raw
}

func chytronLooksPending(msg string) bool {
	up := strings.ToUpper(strings.TrimSpace(msg))
	return strings.Contains(up, "PROSES") ||
		strings.Contains(up, "PROCESS") ||
		strings.Contains(up, "PENDING") ||
		strings.Contains(up, "ANTRI") ||
		strings.Contains(up, "MENUNGGU")
}

func chytronPayloadString(payload map[string]any, keys ...string) string {
	for _, key := range keys {
		rawV, ok := payload[key]
		if !ok {
			continue
		}
		switch vv := rawV.(type) {
		case string:
			if out := strings.TrimSpace(vv); out != "" {
				return out
			}
		case float64:
			return strconv.FormatInt(int64(vv), 10)
		case json.Number:
			return strings.TrimSpace(vv.String())
		case int64:
			return strconv.FormatInt(vv, 10)
		case int:
			return strconv.Itoa(vv)
		}
	}
	return ""
}

func chytronPayloadInt64(payload map[string]any, keys ...string) int64 {
	v := strings.TrimSpace(chytronPayloadString(payload, keys...))
	if v == "" {
		return 0
	}
	clean := strings.ReplaceAll(strings.ReplaceAll(v, ".", ""), ",", "")
	n, _ := strconv.ParseInt(clean, 10, 64)
	return n
}

func resolveChytronFinalStatus(rcStr, msg string) string {
	switch helper.ProviderResponseStateOf("chytron", rcStr, msg) {
	case helper.ProviderResponseSuccess:
		return "success"
	case helper.ProviderResponseFailed:
		return "failed"
	}

	rc := normalizeChytronRC(rcStr)
	if (rc == "1" || rc == "20") && !hasProviderFailureSignal(msg) && !chytronLooksPending(msg) {
		return "success"
	}
	if rc == "68" || rc == "0" || rc == "2" || rc == "3" || chytronLooksPending(msg) {
		return "pending"
	}
	if rc != "" && hasProviderFailureSignal(msg) {
		return "failed"
	}
	return "pending"
}

func (s *ProviderCallbackService) ProcessChytronCallback(ctx context.Context, rawQuery string, q url.Values) (int, map[string]any) {
	payload, refid, rcStr, msg, providerRef, price, saldoTerakhir := parseChytronCallbackPayload(rawQuery, q)
	finalStatus := resolveChytronFinalStatus(rcStr, msg)

	if refid == "" {
		var pricePtr *int64
		if price > 0 {
			pricePtr = &price
		}
		_ = s.repo.InsertAnomali(ctx, repository.ProviderAnomasiIn{Provider: "chytron", RefID: "", KodeRespon: helper.PtrString(rcStr), Pesan: helper.PtrString(msg), Harga: pricePtr, RawQuery: rawQuery, RawBody: rawQuery, Payload: payload})
		if saldoTerakhir != nil && *saldoTerakhir > 0 {
			s.insertAnomaliSnapshot(ctx, "chytron", "", "callback_chytron_anomali", *saldoTerakhir, payload)
		}
		return 200, map[string]any{"ok": true, "ignored": true, "error": "missing refid"}
	}

	row, err := s.repo.GetLatestByRefIDProvider(ctx, refid, "chytron")
	if err != nil {
		return 502, map[string]any{"ok": false, "error": err.Error()}
	}
	if row == nil {
		var pricePtr *int64
		if price > 0 {
			pricePtr = &price
		}
		anomaliIn := repository.ProviderAnomasiIn{Provider: "chytron", RefID: refid, KodeRespon: helper.PtrString(rcStr), Pesan: helper.PtrString(msg), Harga: pricePtr, RawQuery: rawQuery, RawBody: rawQuery, Payload: payload}
		if tujuan := strings.TrimSpace(chytronPayloadString(payload, "tujuan", "dest", "target", "msisdn", "nomor")); tujuan != "" {
			anomaliIn.Tujuan = &tujuan
			if strings.Contains(tujuan, "@") {
				parts := strings.Split(tujuan, "@")
				if len(parts) > 0 && strings.TrimSpace(parts[0]) != "" {
					baseTujuan := strings.TrimSpace(parts[0])
					anomaliIn.Tujuan = &baseTujuan
				}
				if len(parts) > 1 && strings.TrimSpace(parts[len(parts)-1]) != "" {
					clean := strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(parts[len(parts)-1]), ".", ""), ",", "")
					if parsedQty, err := strconv.ParseInt(clean, 10, 64); err == nil && parsedQty > 0 {
						anomaliIn.Qty = &parsedQty
					}
				}
			}
		}
		if qty := chytronPayloadInt64(payload, "qty", "nominal", "amount"); qty > 0 {
			anomaliIn.Qty = &qty
		}
		_ = s.repo.InsertAnomali(ctx, anomaliIn)
		if saldoTerakhir != nil && *saldoTerakhir > 0 {
			s.insertAnomaliSnapshot(ctx, "chytron", refid, "callback_chytron_anomali", *saldoTerakhir, payload)
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
		_ = s.repo.InsertProviderSnapshot(ctx, repository.ProviderSnapshotIn{Provider: "chytron", SaldoProvider: *upd.SaldoTerakhir, RefID: refid, TransaksiMemberID: &tmID, TransaksiProviderID: &tpID, Sumber: "callback_chytron", RawJSON: rawJSON})
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
	if shouldKeepExistingMemberFinalStatus(trx.Status, finalStatus) {
		helper.AppendAlreadyFinalLog("callback already_final provider=chytron refid=%s trx_member_id=%d status=%s callback_status=%s", refid, trx.ID, trx.Status, finalStatus)
		return 200, map[string]any{"ok": true, "already_final": true, "refid": refid}
	}

	if finalStatus == "failed" && !helper.ShouldBlockProviderFallback(msg) {
		fbCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
		defer cancel()
		did, provider, providerRowID, ferr := s.tryFallbackFromChytron(fbCtx, trx, msg)
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
		if err := s.prepareMemberTrxSuccessTransition(ctx, "chytron", trx); err != nil {
			helper.AppendProviderServiceLog("provider_callback_error.log", "success promotion refund recovery failed provider=chytron refid=%s trx_member_id=%d err=%v", refid, trx.ID, err)
			return 200, map[string]any{"ok": false, "refid": refid, "error": err.Error(), "repair_needed": "member_refund_recovery"}
		}
		biayaAktual := trx.BiayaPerkiraan
		if err := s.repo.UpdateTransaksiMemberSettle(ctx, trx.ID, "success", ketDB, biayaAktual, price, hargaMember); err != nil && err != sql.ErrNoRows {
			helper.AppendProviderServiceLog("provider_callback_error.log", "settle failed provider=chytron refid=%s trx_member_id=%d status=success err=%v", refid, trx.ID, err)
			return 200, map[string]any{"ok": false, "error": err.Error(), "repair_needed": "member_settle_failed"}
		}
		s.applyH2HCommission(ctx, trx.ID, refid)
	case "failed":
		ketFailed := strings.TrimSpace(ket)
		if ketFailed == "" || isWeakProviderValue(ketFailed) {
			ketFailed = ketDB
		}
		if err := s.repo.UpdateTransaksiMemberSettle(ctx, trx.ID, "failed", ketFailed, 0, price, hargaMember); err != nil && err != sql.ErrNoRows {
			helper.AppendProviderServiceLog("provider_callback_error.log", "settle failed provider=chytron refid=%s trx_member_id=%d status=failed err=%v", refid, trx.ID, err)
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
	return s.sendDirectMemberWebhook(ctx, "chytron", trx, webhookURL, refid, finalStatus, ket, providerRefOut, snOut, memberSaldo, biayaAktualOut, price)
}
