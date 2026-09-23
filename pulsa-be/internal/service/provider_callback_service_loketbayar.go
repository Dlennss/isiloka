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

var (
	loketTextRefIDRe      = regexp.MustCompile(`#([A-Za-z0-9._-]+)`)
	loketTextStatusRe     = regexp.MustCompile(`(?i)\bstatus\s+([A-Za-z0-9_ -]+?)(?:[.\r\n|]|$)`)
	loketTextDestRe       = regexp.MustCompile(`(?i)(?:\bke|\bTUJUAN)\s+([0-9]+)`)
	loketTextProductRe    = regexp.MustCompile(`(?i)\bTRX\s+TOPUP\s+([A-Za-z0-9_:-]+)\s+ke\b`)
	loketTextReffRe       = regexp.MustCompile(`(?i)(?:^|[/\s])reff:([^/.\s]+)`)
	loketTextSNRe         = regexp.MustCompile(`(?i)\bSN(?:/REF)?:\s*(.+?)(?:\.\s*(?:HARGA|SALDO)\s*:|\.(?:HARGA|SALDO)\s*:|$)`)
	loketTextOrderRe      = regexp.MustCompile(`(?i)(?:^|[/\s])order:([^/.\s]+)`)
	loketTextHargaRe      = regexp.MustCompile(`(?i)\bHARGA:\s*([0-9.,]+)`)
	loketTextSaldoRe      = regexp.MustCompile(`(?i)\bSALDO:\s*([0-9.,]+)`)
	loketTextTotalRe      = regexp.MustCompile(`(?i)(?:^|[/\s])total:([0-9.,]+)`)
	loketTextNominalRe    = regexp.MustCompile(`(?i)(?:^|[/\s])nominal:([0-9.,]+)`)
	loketTextAdminRe      = regexp.MustCompile(`(?i)(?:^|[/\s])admin:([0-9.,]+)`)
	loketTextBankRe       = regexp.MustCompile(`(?i)(?:^|[/\s])bank:([^/]+)`)
	loketTextPenerimaRe   = regexp.MustCompile(`(?i)(?:^|[/\s])nama:([^/]+)`)
	loketTextPengirimRe   = regexp.MustCompile(`(?i)(?:^|[/\s])pengirim:([^/]+)`)
	loketTextCustomerName = regexp.MustCompile(`(?i)(?:^|[/\s])nama_penerima:([^/]+)`)
)

func parseLoketBayarCallbackPayload(raw string, q url.Values) (map[string]any, string, string, string, string, int64, *int64) {
	payload := map[string]any{}
	trimmed := strings.TrimSpace(raw)
	if strings.HasPrefix(trimmed, "{") {
		_ = json.Unmarshal([]byte(trimmed), &payload)
	}

	getAny := func(keys ...string) string {
		for _, key := range keys {
			if v := strings.TrimSpace(q.Get(key)); v != "" {
				return v
			}
			if rawV, ok := loketPayloadValue(payload, key); ok {
				if out := loketPayloadValueString(rawV); out != "" {
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

	if len(payload) == 0 {
		if textPayload, ok := parseLoketBayarOtomaxTextPayload(trimmed, q); ok {
			payload = textPayload
		}
	}

	refid := getAny("ref_id", "refid", "reference_id", "orderId")
	rc := getAny("status", "rc", "kode_respon", "status_code")
	msg := strings.TrimSpace(getAny("keterangan", "message", "pesan", "msg", "description"))
	if msg == "" {
		msg = strings.TrimSpace(raw)
	}
	providerRef := strings.TrimSpace(getAny("reff", "sn", "provider_ref", "noref", "no_ref", "reference"))
	price := parseI64("harga", "total", "nominal", "price")
	var saldoPtr *int64
	if saldo := parseI64("saldo", "balance", "saldo_akhir", "saldoTerakhir"); saldo > 0 {
		saldoPtr = &saldo
	}
	if len(payload) == 0 {
		payload = map[string]any{
			"refid":      refid,
			"status":     rc,
			"keterangan": msg,
			"reff":       providerRef,
			"harga":      price,
		}
		if saldoPtr != nil {
			payload["saldo"] = *saldoPtr
		}
	}
	return payload, refid, rc, msg, providerRef, price, saldoPtr
}

func parseLoketBayarOtomaxTextPayload(raw string, q url.Values) (map[string]any, bool) {
	text := strings.TrimSpace(q.Get("q"))
	if text == "" {
		if parsed, err := url.ParseQuery(raw); err == nil {
			text = strings.TrimSpace(parsed.Get("q"))
		}
	}
	if text == "" {
		text = strings.TrimSpace(raw)
		if strings.HasPrefix(text, "q=") {
			value := strings.TrimSpace(strings.TrimPrefix(text, "q="))
			if idx := strings.IndexByte(value, '&'); idx >= 0 {
				value = value[:idx]
			}
			if decoded, err := url.QueryUnescape(value); err == nil {
				text = strings.TrimSpace(decoded)
			} else {
				text = value
			}
		}
	}
	text = normalizeLoketBayarOtomaxText(text)
	if text == "" || !strings.Contains(text, "#") {
		return nil, false
	}

	refid := loketTextMatch(text, loketTextRefIDRe)
	if refid == "" {
		return nil, false
	}
	status := strings.ToUpper(strings.TrimSpace(loketTextMatch(text, loketTextStatusRe)))
	reff := strings.TrimSpace(loketTextMatch(text, loketTextReffRe))
	if reff == "" {
		reff = strings.Trim(strings.TrimSpace(loketTextMatch(text, loketTextSNRe)), ".")
	}
	payload := map[string]any{
		"refid":      refid,
		"ref_id":     refid,
		"status":     status,
		"keterangan": text,
		"reff":       reff,
		"format":     "otomax_text",
	}
	if product := strings.ToUpper(strings.TrimSpace(loketTextMatch(text, loketTextProductRe))); product != "" {
		payload["kodeProduk"] = product
	}
	if dest := strings.TrimSpace(loketTextMatch(text, loketTextDestRe)); dest != "" {
		payload["nomor_rekening"] = dest
	}
	if order := strings.TrimSpace(loketTextMatch(text, loketTextOrderRe)); order != "" {
		payload["provider_order_id"] = order
	}
	if harga := loketTextInt64(text, loketTextHargaRe); harga > 0 {
		payload["harga"] = harga
	}
	if total := loketTextInt64(text, loketTextTotalRe); total > 0 {
		payload["total"] = total
		if _, ok := payload["harga"]; !ok {
			payload["harga"] = total
		}
	}
	if nominal := loketTextInt64(text, loketTextNominalRe); nominal > 0 {
		payload["nominal"] = nominal
	}
	if admin := loketTextInt64(text, loketTextAdminRe); admin >= 0 && loketTextMatch(text, loketTextAdminRe) != "" {
		payload["admin"] = admin
	}
	if saldo := loketTextInt64(text, loketTextSaldoRe); saldo > 0 {
		payload["saldo"] = saldo
	}
	if bank := strings.TrimSpace(loketTextMatch(text, loketTextBankRe)); bank != "" {
		payload["bank"] = bank
	}
	if penerima := strings.TrimSpace(loketTextMatch(text, loketTextPenerimaRe)); penerima != "" {
		payload["nama_penerima"] = penerima
	} else if penerima := strings.TrimSpace(loketTextMatch(text, loketTextCustomerName)); penerima != "" {
		payload["nama_penerima"] = penerima
	}
	if pengirim := strings.TrimSpace(loketTextMatch(text, loketTextPengirimRe)); pengirim != "" {
		payload["pengirim"] = pengirim
	}
	return payload, true
}

func normalizeLoketBayarOtomaxText(text string) string {
	text = strings.Trim(strings.TrimSpace(text), "\"")
	for i := 0; i < 3; i++ {
		decoded, err := url.QueryUnescape(text)
		if err != nil {
			break
		}
		decoded = strings.Trim(strings.TrimSpace(decoded), "\"")
		if decoded == text {
			break
		}
		text = decoded
	}
	return strings.TrimSpace(text)
}

func loketTextMatch(text string, re *regexp.Regexp) string {
	match := re.FindStringSubmatch(text)
	if len(match) < 2 {
		return ""
	}
	return strings.TrimSpace(match[1])
}

func loketTextInt64(text string, re *regexp.Regexp) int64 {
	value := loketTextMatch(text, re)
	if value == "" {
		return 0
	}
	clean := strings.ReplaceAll(strings.ReplaceAll(value, ".", ""), ",", "")
	n, _ := strconv.ParseInt(clean, 10, 64)
	return n
}

func loketPayloadValue(payload map[string]any, key string) (any, bool) {
	if payload == nil {
		return nil, false
	}
	if rawV, ok := payload[key]; ok {
		return rawV, true
	}
	apiRaw, ok := payload["api"]
	if !ok {
		return nil, false
	}
	apiPayload, ok := apiRaw.(map[string]any)
	if !ok {
		return nil, false
	}
	rawV, ok := apiPayload[key]
	return rawV, ok
}

func loketPayloadValueString(rawV any) string {
	switch vv := rawV.(type) {
	case string:
		return strings.TrimSpace(vv)
	case float64:
		return strconv.FormatInt(int64(vv), 10)
	case json.Number:
		return vv.String()
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

func loketPayloadString(payload map[string]any, keys ...string) string {
	for _, key := range keys {
		rawV, ok := loketPayloadValue(payload, key)
		if !ok {
			continue
		}
		if out := loketPayloadValueString(rawV); out != "" {
			return out
		}
	}
	return ""
}

func loketPayloadInt64(payload map[string]any, keys ...string) int64 {
	v := strings.TrimSpace(loketPayloadString(payload, keys...))
	if v == "" {
		return 0
	}
	clean := strings.ReplaceAll(strings.ReplaceAll(v, ".", ""), ",", "")
	n, _ := strconv.ParseInt(clean, 10, 64)
	return n
}

func matchLoketCallbackDest(callbackDest, trxDest string) bool {
	callbackDest = strings.TrimSpace(callbackDest)
	trxDest = strings.TrimSpace(trxDest)
	if callbackDest == "" || trxDest == "" {
		return true
	}
	if callbackDest == trxDest {
		return true
	}
	return strings.HasSuffix(callbackDest, trxDest)
}

func deriveLoketCallbackKodeProduk(payload map[string]any, trx *repository.CallbackTrxMemberFull) string {
	productSent := strings.ToUpper(strings.TrimSpace(loketPayloadString(payload, "kodeProduk", "kode_produk", "produk")))
	if productSent == "" {
		return productSent
	}
	callbackDest := strings.TrimSpace(loketPayloadString(payload, "nomor_rekening", "idpel", "dest", "tujuan", "msisdn"))
	if trx == nil {
		return productSent
	}
	trxDest := strings.TrimSpace(trx.Tujuan)
	if callbackDest == "" || trxDest == "" || !strings.HasSuffix(callbackDest, trxDest) {
		return productSent
	}
	bankCode := strings.TrimSpace(strings.TrimSuffix(callbackDest, trxDest))
	if bankCode == "" {
		return productSent
	}
	return productSent + ":" + bankCode
}

func (s *ProviderCallbackService) ensureLoketBayarCallbackProviderRow(ctx context.Context, refid string, payload map[string]any) (*repository.ProviderTrxRefRow, *repository.CallbackTrxMemberFull, error) {
	if s == nil || s.repo == nil || strings.TrimSpace(refid) == "" {
		return nil, nil, nil
	}
	if row, err := s.repo.GetLatestByRefIDProvider(ctx, refid, "loketbayar"); err != nil || row != nil {
		return row, nil, err
	}
	trx, err := s.repo.GetLatestTransaksiMemberByRefID(ctx, refid)
	if err != nil || trx == nil {
		return nil, trx, err
	}
	productSent := strings.ToUpper(strings.TrimSpace(loketPayloadString(payload, "kodeProduk", "kode_produk", "produk")))
	if productSent != "" && !isLoketBayarBankTransferProduct(productSent) {
		return nil, trx, nil
	}
	callbackDest := strings.TrimSpace(loketPayloadString(payload, "nomor_rekening", "idpel", "dest", "tujuan", "msisdn"))
	if !matchLoketCallbackDest(callbackDest, trx.Tujuan) {
		return nil, trx, nil
	}
	nominal := loketPayloadInt64(payload, "nominal")
	if nominal > 0 && trx.Qty > 0 && nominal != trx.Qty {
		return nil, trx, nil
	}
	createIn := repository.ProviderTrxCreateIn{
		Provider:          "loketbayar",
		TransaksiMemberID: trx.ID,
		RefID:             trx.RefID,
		Perintah:          strings.ToUpper(strings.TrimSpace(trx.Perintah)),
		ProdukSKUSnapshot: strings.ToUpper(strings.TrimSpace(trx.KodeProduk)),
		KodeProduk:        deriveLoketCallbackKodeProduk(payload, trx),
		Tujuan:            callbackDest,
		Qty:               trx.Qty,
	}
	if createIn.KodeProduk == "" {
		createIn.KodeProduk = productSent
	}
	if createIn.Tujuan == "" {
		createIn.Tujuan = strings.TrimSpace(trx.Tujuan)
	}
	if nominal > 0 {
		createIn.Qty = nominal
	}
	created, err := s.repo.CreateProviderTrx(ctx, createIn, map[string]any{
		"source":  "callback_loketbayar_synthetic",
		"payload": payload,
	})
	if err != nil {
		return nil, trx, err
	}
	if created != nil {
		helper.AppendProviderServiceLog("provider_callback_service.log", "synthetic loketbayar provider row created refid=%s trx_member_id=%d provider_trx_id=%d kode_produk=%s tujuan=%s qty=%d", refid, trx.ID, created.ID, createIn.KodeProduk, createIn.Tujuan, createIn.Qty)
	}
	row, err := s.repo.GetLatestByRefIDProvider(ctx, refid, "loketbayar")
	return row, trx, err
}

func (s *ProviderCallbackService) loketHasOtherSuccessAttempt(ctx context.Context, refid string, currentProviderRowID int64) bool {
	if s == nil || s.repo == nil || strings.TrimSpace(refid) == "" {
		return false
	}
	rows, err := s.repo.ListAttemptsByRefID(ctx, refid)
	if err != nil {
		helper.AppendProviderServiceLog("provider_callback_error.log", "loketbayar check other success failed refid=%s err=%v", refid, err)
		return false
	}
	for _, attempt := range rows {
		if attempt.ID == currentProviderRowID {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(attempt.Status), "success") {
			return true
		}
	}
	return false
}

func (s *ProviderCallbackService) insertLoketBayarCallbackAnomali(ctx context.Context, refid string, payload map[string]any, rawQuery, rcStr, msg string, price int64) {
	var pricePtr *int64
	if price > 0 {
		pricePtr = &price
	}
	callbackDest := strings.TrimSpace(loketPayloadString(payload, "nomor_rekening", "idpel", "dest", "tujuan", "msisdn"))
	nominal := loketPayloadInt64(payload, "nominal")
	anomaliIn := repository.ProviderAnomasiIn{
		Provider:   "loketbayar",
		RefID:      refid,
		KodeRespon: helper.PtrString(rcStr),
		Pesan:      helper.PtrString(msg),
		Harga:      pricePtr,
		RawQuery:   rawQuery,
		RawBody:    rawQuery,
		Payload:    payload,
	}
	if callbackDest != "" {
		anomaliIn.Tujuan = &callbackDest
	}
	if nominal > 0 {
		anomaliIn.Qty = &nominal
	}
	_ = s.repo.InsertAnomali(ctx, anomaliIn)
}

func (s *ProviderCallbackService) tryApplyLoketBayarDepositVA(ctx context.Context, refid string, payload map[string]any, finalStatus, msg string, saldoTerakhir *int64) (bool, int, map[string]any) {
	if s == nil || s.depositRepo == nil || strings.TrimSpace(refid) == "" {
		return false, 0, nil
	}
	existing, err := s.depositRepo.GetByRefID(ctx, refid)
	if err == sql.ErrNoRows {
		return false, 0, nil
	}
	if err != nil {
		helper.AppendProviderServiceLog("provider_callback_error.log", "loketbayar va lookup failed refid=%s err=%v", refid, err)
		return true, 200, map[string]any{"ok": false, "refid": refid, "error": err.Error(), "repair_needed": "deposit_va_lookup_failed"}
	}
	if existing == nil || !strings.EqualFold(strings.TrimSpace(existing.Metode), "va") {
		return false, 0, nil
	}

	callbackDest := strings.TrimSpace(loketPayloadString(payload, "nomor_rekening", "tujuan", "dest", "idpel", "msisdn"))
	note := loketBayarDepositVACallbackNote(finalStatus, msg)
	row, err := s.depositRepo.ApplyVACallback(ctx, refid, finalStatus, callbackDest, note)
	if err != nil {
		helper.AppendProviderServiceLog("provider_callback_error.log", "loketbayar va apply failed refid=%s status=%s err=%v", refid, finalStatus, err)
		return true, 200, map[string]any{"ok": false, "refid": refid, "error": err.Error(), "repair_needed": "deposit_va_apply_failed"}
	}

	if saldoTerakhir != nil && *saldoTerakhir > 0 {
		rawJSON, _ := json.Marshal(payload)
		if err := s.repo.InsertProviderSnapshot(ctx, repository.ProviderSnapshotIn{
			Provider:      "loketbayar",
			SaldoProvider: *saldoTerakhir,
			RefID:         refid,
			Sumber:        "callback_loketbayar_va_deposit",
			RawJSON:       rawJSON,
		}); err != nil {
			helper.AppendProviderServiceLog("provider_callback_error.log", "loketbayar va snapshot failed refid=%s saldo=%d err=%v", refid, *saldoTerakhir, err)
		}
	}

	out := map[string]any{
		"ok":                 true,
		"refid":              refid,
		"deposit_va":         true,
		"status":             row.Status,
		"deposit_request_id": row.ID,
		"amount":             row.Amount,
	}
	if saldoTerakhir != nil {
		out["provider_saldo_after"] = *saldoTerakhir
	}
	return true, 200, out
}

func loketBayarDepositVACallbackNote(finalStatus, msg string) string {
	status := strings.TrimSpace(finalStatus)
	if status == "" {
		status = "pending"
	}
	msg = strings.Join(strings.Fields(strings.TrimSpace(msg)), " ")
	if len(msg) > 500 {
		msg = msg[:500]
	}
	if msg == "" {
		return "LoketBayar VA callback " + status
	}
	return "LoketBayar VA callback " + status + ": " + msg
}

func (s *ProviderCallbackService) tryApplyLoketBayarProviderTransfer(ctx context.Context, refid string, payload map[string]any, finalStatus, msg, providerRef string, price int64, saldoTerakhir *int64) (bool, int, map[string]any) {
	if s == nil || s.loketTransferRepo == nil || strings.TrimSpace(refid) == "" {
		return false, 0, nil
	}
	existing, err := s.loketTransferRepo.GetByRefID(ctx, refid)
	if err == sql.ErrNoRows {
		return false, 0, nil
	}
	if err != nil {
		helper.AppendProviderServiceLog("provider_callback_error.log", "loketbayar transfer lookup failed refid=%s err=%v", refid, err)
		return true, 200, map[string]any{"ok": false, "refid": refid, "error": err.Error(), "repair_needed": "loketbayar_transfer_lookup_failed"}
	}
	if existing == nil {
		return false, 0, nil
	}

	rawJSON, _ := json.Marshal(payload)
	row, err := s.loketTransferRepo.ApplyCallback(ctx, refid, finalStatus, providerRef, msg, price, saldoTerakhir, rawJSON)
	if err != nil {
		helper.AppendProviderServiceLog("provider_callback_error.log", "loketbayar transfer apply failed refid=%s status=%s err=%v", refid, finalStatus, err)
		return true, 200, map[string]any{"ok": false, "refid": refid, "error": err.Error(), "repair_needed": "loketbayar_transfer_apply_failed"}
	}

	if saldoTerakhir != nil && *saldoTerakhir > 0 {
		if err := s.repo.InsertProviderSnapshot(ctx, repository.ProviderSnapshotIn{
			Provider:      "loketbayar",
			SaldoProvider: *saldoTerakhir,
			RefID:         refid,
			Sumber:        "callback_loketbayar_provider_transfer",
			RawJSON:       rawJSON,
		}); err != nil {
			helper.AppendProviderServiceLog("provider_callback_error.log", "loketbayar transfer snapshot failed refid=%s saldo=%d err=%v", refid, *saldoTerakhir, err)
		}
	}

	out := map[string]any{
		"ok":                   true,
		"refid":                refid,
		"loketbayar_transfer":  true,
		"status":               row.Status,
		"transfer_id":          row.ID,
		"provider":             row.Provider,
		"amount":               row.Amount,
		"admin_fee":            row.AdminFee,
		"provider_saldo_after": row.ProviderSaldoAfter,
	}
	if saldoTerakhir != nil {
		out["source_saldo_after"] = *saldoTerakhir
	}
	return true, 200, out
}

func (s *ProviderCallbackService) ProcessLoketBayarCallback(ctx context.Context, rawQuery string, q url.Values) (int, map[string]any) {
	payload, refid, rcStr, msg, providerRef, price, saldoTerakhir := parseLoketBayarCallbackPayload(rawQuery, q)
	finalStatus := resolveLoketBayarFinalStatus(rcStr, msg)

	if refid == "" {
		var pricePtr *int64
		if price > 0 {
			pricePtr = &price
		}
		_ = s.repo.InsertAnomali(ctx, repository.ProviderAnomasiIn{Provider: "loketbayar", RefID: "", KodeRespon: helper.PtrString(rcStr), Pesan: helper.PtrString(msg), Harga: pricePtr, RawQuery: rawQuery, RawBody: rawQuery, Payload: payload})
		if saldoTerakhir != nil && *saldoTerakhir > 0 {
			s.insertAnomaliSnapshot(ctx, "loketbayar", "", "callback_loketbayar_anomali", *saldoTerakhir, payload)
		}
		return 200, map[string]any{"ok": true, "ignored": true, "error": "missing refid"}
	}

	if handled, status, out := s.tryApplyLoketBayarDepositVA(ctx, refid, payload, finalStatus, msg, saldoTerakhir); handled {
		return status, out
	}
	if handled, status, out := s.tryApplyLoketBayarProviderTransfer(ctx, refid, payload, finalStatus, msg, providerRef, price, saldoTerakhir); handled {
		return status, out
	}

	row, trxFallback, err := s.ensureLoketBayarCallbackProviderRow(ctx, refid, payload)
	if err != nil {
		return 502, map[string]any{"ok": false, "error": err.Error()}
	}
	if row == nil {
		s.insertLoketBayarCallbackAnomali(ctx, refid, payload, rawQuery, rcStr, msg, price)
		if saldoTerakhir != nil && *saldoTerakhir > 0 {
			s.insertAnomaliSnapshot(ctx, "loketbayar", refid, "callback_loketbayar_anomali", *saldoTerakhir, payload)
		}
		out := map[string]any{"ok": true, "ignored": true, "refid": refid}
		if trxFallback != nil {
			out["trx_member_id"] = trxFallback.ID
		}
		return 200, out
	}

	if strings.EqualFold(strings.TrimSpace(row.Status), "success") && finalStatus == "failed" && !s.loketHasOtherSuccessAttempt(ctx, refid, row.ID) {
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
		if err := s.repo.UpdateResult(ctx, row.ID, upd); err != nil {
			helper.AppendProviderServiceLog("provider_callback_error.log", "loketbayar already-final failed callback provider update failed refid=%s provider_row_id=%d err=%v", refid, row.ID, err)
			return 200, map[string]any{"ok": false, "refid": refid, "error": err.Error(), "repair_needed": "provider_failed_status_update"}
		}
		if upd.SaldoTerakhir != nil && *upd.SaldoTerakhir > 0 {
			tmID := row.TransaksiMemberID
			tpID := row.ID
			rawJSON, _ := json.Marshal(payload)
			_ = s.repo.InsertProviderSnapshot(ctx, repository.ProviderSnapshotIn{Provider: "loketbayar", SaldoProvider: *upd.SaldoTerakhir, RefID: refid, TransaksiMemberID: &tmID, TransaksiProviderID: &tpID, Sumber: "callback_loketbayar", RawJSON: rawJSON})
		}
		s.insertLoketBayarCallbackAnomali(ctx, refid, payload, rawQuery, rcStr, msg, price)
		helper.AppendAlreadyFinalLog("callback already_final provider=loketbayar refid=%s provider_row_id=%d provider_status=failed member_status=success callback_status=failed reason=no_other_provider_success", refid, row.ID)
		return 200, map[string]any{"ok": true, "already_final": true, "refid": refid, "status": "success", "provider_status": "failed", "reason": "no_other_provider_success"}
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
		_ = s.repo.InsertProviderSnapshot(ctx, repository.ProviderSnapshotIn{Provider: "loketbayar", SaldoProvider: *upd.SaldoTerakhir, RefID: refid, TransaksiMemberID: &tmID, TransaksiProviderID: &tpID, Sumber: "callback_loketbayar", RawJSON: rawJSON})
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
		helper.AppendAlreadyFinalLog("callback already_final provider=loketbayar refid=%s trx_member_id=%d status=%s callback_status=%s", refid, trx.ID, trx.Status, finalStatus)
		return 200, map[string]any{"ok": true, "already_final": true, "refid": refid}
	}

	if finalStatus == "failed" && helper.LooksLikeProviderOutOfBalance(msg) {
		s.disableProviderForOutOfBalance(ctx, "loketbayar", msg)
		if time.Since(row.DibuatPada) < 30*time.Second {
			helper.AppendProviderServiceLog("provider_callback_service.log", "loketbayar out_of_balance fresh callback ignored after auto-disable refid=%s row_id=%d age_ms=%d", refid, row.ID, time.Since(row.DibuatPada).Milliseconds())
			return 200, map[string]any{"ok": true, "refid": refid, "status": "pending", "reason": "provider_out_of_balance_auto_disabled_wait_primary_route"}
		}
	}

	if finalStatus == "failed" && !helper.ShouldBlockProviderFallback(msg) {
		fbCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
		defer cancel()
		did, provider, providerRowID, ferr := s.tryFallbackFromProvider(fbCtx, trx, "loketbayar", "loketbayar_callback_fallback", msg)
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
		if err := s.prepareMemberTrxSuccessTransition(ctx, "loketbayar", trx); err != nil {
			helper.AppendProviderServiceLog("provider_callback_error.log", "success promotion refund recovery failed provider=loketbayar refid=%s trx_member_id=%d err=%v", refid, trx.ID, err)
			return 200, map[string]any{"ok": false, "refid": refid, "error": err.Error(), "repair_needed": "member_refund_recovery"}
		}
		biayaAktual := trx.BiayaPerkiraan
		if err := s.repo.UpdateTransaksiMemberSettle(ctx, trx.ID, "success", ketDB, biayaAktual, price, hargaMember); err != nil && err != sql.ErrNoRows {
			helper.AppendProviderServiceLog("provider_callback_error.log", "settle failed provider=loketbayar refid=%s trx_member_id=%d status=success err=%v", refid, trx.ID, err)
			return 200, map[string]any{"ok": false, "error": err.Error(), "repair_needed": "member_settle_failed"}
		}
		s.applyH2HCommission(ctx, trx.ID, refid)
	case "failed":
		ketFailed := strings.TrimSpace(ket)
		if ketFailed == "" || isWeakProviderValue(ketFailed) {
			ketFailed = ketDB
		}
		if err := s.repo.UpdateTransaksiMemberSettle(ctx, trx.ID, "failed", ketFailed, 0, price, hargaMember); err != nil && err != sql.ErrNoRows {
			helper.AppendProviderServiceLog("provider_callback_error.log", "settle failed provider=loketbayar refid=%s trx_member_id=%d status=failed err=%v", refid, trx.ID, err)
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
	return s.sendDirectMemberWebhook(ctx, "loketbayar", trx, webhookURL, refid, finalStatus, ket, providerRefOut, snOut, memberSaldo, biayaAktualOut, price)
}
